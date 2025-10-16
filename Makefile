all: build tidy lint fmt test

#-------------------------------------------------------------------------
# Variables
# ------------------------------------------------------------------------
env=CGO_ENABLED=1
SHELL := $(shell which bash)

# Allow passing interface via environment (e.g., IFACE=eth0 make sanity)
IFACE ?= lo
CORPUS_DIR ?= ./datasets/captures/$(IFACE)_small
CORPUS_MANIFEST ?= ./datasets/$(IFACE)_small.json

# ICS repo test controls (can be overridden via environment)
ICS_REPO_TESTS ?= 0
ICS_MAX_FILES ?= 50
ICS_TSHARK_TIMEOUT ?= 30s
ICS_MAX_BYTES ?= 10485760

GOOS ?= $(shell go env GOOS)
GOARCH ?= $(shell go env GOARCH)

deps:
	mkdir -p out

pre-commit: deps upgrade tidy fmt lint build test

test:
	CGO_ENABLED=1 go test -v -cover -failfast -race ./...

# Run live-capture tests with zcap (requires Linux, cgo, and zcap headers/libs).
# Uses the path of `which zcap` and points CGO to ../include and ../lib.
UNAME_S := $(shell uname -s)

test-live:
	@bash scripts/test_live.sh

fuzz:
	fuzz

bench:
	bench

bench-json:
	go test -json -bench=. ./... | tee ./out/bench.json

perf-rss:
	@echo "[perf-rss] running 1GiB synthetic stream RSS gate"
	@mkdir -p ./out
	@PERFRSS_SIZE=$$((1<<30)) go run ./cmd/perfrss > ./out/perfrss.json || { echo "[perf-rss] FAILED"; cat ./out/perfrss.json || true; exit 1; }
	@echo "[perf-rss] result:"; cat ./out/perfrss.json

test-all: test fuzz

# ------------------------------------------------------------------------
# QEMU cross-arch smoke tests
# ------------------------------------------------------------------------
QEMU_AARCH64 := $(shell command -v qemu-aarch64 2>/dev/null)
HOST_ARCH := $(shell uname -m)


.PHONY: test-qemu-arm64 test-qemu
test-qemu-arm64:
	@bash scripts/test_qemu_arm64.sh

# Aggregate qemu tests (extendable for more archs later)
test-qemu: test-qemu-arm64

# Integration tests against tshark and tcpdump (optional)
.PHONY: test-integration

test-integration:
	@if ! command -v tshark >/dev/null 2>&1; then echo "tshark not found; skipping integration tests"; exit 0; fi
	@if ! command -v tcpdump >/dev/null 2>&1; then echo "tcpdump not found; skipping integration tests"; exit 0; fi
	CGO_ENABLED=1 go test -v ./integration

# ICS-pcap repository tests
.PHONY: test-ics test-ics-cross test-ics-all

# Ensure submodule is initialized before running ICS tests
ics-submodule:
	@git submodule status testdata/ICS-pcap >/dev/null 2>&1 || \
	  git submodule update --init --recursive testdata/ICS-pcap

# Normalize all files (no external tools required)
test-ics: ics-submodule
	@echo "[ics] Normalize all .pcap/.pcapng in testdata/ICS-pcap"
	ICS_REPO_TESTS=1 \
	go test ./integration -run TestICSRepo_NormalizePCAP -v

# Cross-tool sample using tshark/tcpdump with bounds
test-ics-cross: ics-submodule
	@if ! command -v tshark >/dev/null 2>&1; then echo "tshark not found; skipping"; exit 0; fi
	@if ! command -v tcpdump >/dev/null 2>&1; then echo "tcpdump not found; skipping"; exit 0; fi
	@echo "[ics] Cross-tool sample: max=$(ICS_MAX_FILES) timeout=$(ICS_TSHARK_TIMEOUT) max_bytes=$(ICS_MAX_BYTES)"
	ICS_REPO_TESTS=1 ICS_MAX_FILES=$(ICS_MAX_FILES) ICS_TSHARK_TIMEOUT=$(ICS_TSHARK_TIMEOUT) ICS_MAX_BYTES=$(ICS_MAX_BYTES) \
	go test ./integration -run TestICSRepo_CrossTool -v

# Run both ICS tests
test-ics-all: test-ics test-ics-cross

fmt:
	@echo "Formatting Go files (excluding testdata)..."
	@find . -type f -name "*.go" ! -path "*/testdata/*" -exec goimports -w {} +
	@find . -type f -name "*.go" ! -path "*/testdata/*" -exec gofmt -s -w {} +

guards:
	@[ -d ./tools/importguard ] && { echo "[guard] import allowlist"; go run ./tools/importguard; } || echo "[guard] import allowlist (skipped)"
	@[ -d ./tools/doccheck ] && { echo "[guard] doc comments"; go run ./tools/doccheck; } || echo "[guard] doc comments (skipped)"

lint: fmt guards
	lint || true

build: test
	$(env) go build ./...

# Build pcapdump CLI
.PHONY: pcapdump
pcapdump:
	@./scripts/build_pcapdump.sh

# Build demo application (livecapdemo) using same zcap discovery mechanism.
.PHONY: livecapdemo
livecapdemo: 
	@./scripts/build_livecapdemo.sh

# Build minimal static-leaning (external) example for downstream packages: prints zcap stats via StreamStats.
.PHONY: example-static
example-static:
	@echo "Building example static consumer (no code provided yet; reuses pcapdump for demonstration)"
	$(MAKE) pcapdump

# Build fully static binary using MUSL via zig cc (Linux only)
.PHONY: pcapdump-static
pcapdump-static:
	@bash scripts/build_pcapdump_static.sh

# Cross-compile fully static arm64 MUSL binary (requires arm64 static libzcap)
.PHONY: pcapdump-static-arm64
pcapdump-static-arm64:
	@bash scripts/build_pcapdump_static_arm64.sh

release-dev:
	$(env) goreleaser release --clean --snapshot

upgrade:
	upgrade

tidy: fmt
	tidy

release: 
	if [ -z "$(tag)" ]; then echo "tag is required"; exit 1; fi
	git tag -a ${tag} -m "${tag}"
	git push origin ${tag}

# Canary targets
canary:
	go build -o ./bin/canary ./main.go && ./bin/canary --root . --out status.json --csv status.csv

canary-verify:
	./bin/canary --root . --verify GAP_ANALYSIS.md --strict

canary-build:
	go build -ldflags="-s -w" -o ./bin/canary ./main.go

canary-install:
	go install ./main.go

acceptance:
	go test ./tools/canary/... -run TestAcceptance -v
	go test ./internal/acceptance/... -v

# GoReleaser targets
release-snapshot:
	goreleaser release --snapshot --clean

release-check:
	goreleaser check

release-local:
	goreleaser release --snapshot --skip=publish --clean

clean: 
	rm -rf dist
	rm -rf out

#-------------------------------------------------------------------------
# CI targets
#-------------------------------------------------------------------------
build-ci: lint
	$(env) go build ./...

test-ci: deps build-ci
	CGO_ENABLED=1 go test \
		-cover \
		-covermode=atomic \
		-coverprofile=./out/coverage.txt \
		-failfast \
		-race ./...
	@# Downstream import compile checks (dynamic + static preference)
	@if [ "$(UNAME_S)" = "Linux" ]; then \
	  echo "[ci] downstream import checks"; \
	  CGO_ENABLED=1 go test -c -tags zcap ./internal/captest || exit 1; \
	  CGO_ENABLED=1 go test -c -tags "zcap zcapstatic" ./internal/captest || exit 1; \
	else \
	  echo "[ci] skipping downstream import checks (non-Linux host)"; \
	fi
	@if [ "$(UNAME_S)" = "Linux" ] && [ "$(HOST_ARCH)" = "x86_64" ]; then \
		$(MAKE) test-qemu-arm64 || exit 1; \
	else \
		echo "[ci] skipping qemu arm64 test (not x86_64 Linux host)"; \
	fi
	make fuzz FUZZ_TIME=10


bench-ci: deps test-ci
	go test -bench=. ./... | tee ./out/bench-output.txt


release-ci: bench-ci
	$(env) goreleaser release --clean

sanity:
	# External-only collection to work without zcap; then manifest and compare
	sudo -E go run ./cmd/pcapcollect -iface $(IFACE) -outdir $(CORPUS_DIR) -count 3 -duration 2s -formats "" -external tcpdump,tshark
	go run ./cmd/pcapmanifest -root $(CORPUS_DIR) > $(CORPUS_MANIFEST)
	go run ./cmd/pcapcompare -root $(CORPUS_DIR) -tool tshark

sanity-long:
	# Longer external-only collection
	sudo -E go run ./cmd/pcapcollect -iface $(IFACE) -outdir $(CORPUS_DIR) -count 50 -duration 2s -formats "" -external tcpdump,tshark
	go run ./cmd/pcapmanifest -root $(CORPUS_DIR) > $(CORPUS_MANIFEST)
	go run ./cmd/pcapcompare -root $(CORPUS_DIR) -tool tshark

#-------------------------------------------------------------------------
# Comparison targets
#-------------------------------------------------------------------------
.PHONY: compare-testdata compare-testdata-tshark compare-testdata-tcpdump \
	compare-lo-small compare-lo-small-tshark compare-lo-small-tcpdump \
	compare-all

compare-testdata-tshark:
	@echo "[compare] testdata/pcap vs tshark"
	go run ./cmd/pcapcompare -root ./testdata/pcap -tool tshark

compare-testdata-tcpdump:
	@echo "[compare] testdata/pcap vs tcpdump"
	go run ./cmd/pcapcompare -root ./testdata/pcap -tool tcpdump

compare-testdata: compare-testdata-tshark compare-testdata-tcpdump

compare-lo-small-tshark:
	@if [ ! -d "$(CORPUS_DIR)" ]; then \
		echo "corpus not found at $(CORPUS_DIR); run 'IFACE=$(IFACE) make sanity' first"; exit 0; \
	fi
	@echo "[compare] $(CORPUS_DIR) vs tshark"
	go run ./cmd/pcapcompare -root $(CORPUS_DIR) -tool tshark

compare-lo-small-tcpdump:
	@if [ ! -d "$(CORPUS_DIR)" ]; then \
		echo "corpus not found at $(CORPUS_DIR); run 'IFACE=$(IFACE) make sanity' first"; exit 0; \
	fi
	@echo "[compare] $(CORPUS_DIR) vs tcpdump"
	go run ./cmd/pcapcompare -root $(CORPUS_DIR) -tool tcpdump

compare-lo-small: compare-lo-small-tshark compare-lo-small-tcpdump

compare-all: compare-testdata compare-lo-small

# ------------------------------------------------------------------------
# Example live capture builds (dynamic and static variants)
# ------------------------------------------------------------------------
.PHONY: example-livecap example-livecap-static
example-livecap:
	@if [ "$(UNAME_S)" != "Linux" ]; then echo "example-livecap: requires Linux"; exit 1; fi
	@if [ ! -d ./.zcap/linux/amd64/include ] && [ ! -d ./.zcap/linux/arm64/include ]; then \
		echo "[auto] fetching zcap artifacts"; \
		bash scripts/fetch_zcap.sh || { echo "fetch_zcap failed"; exit 1; }; \
	fi
	@if [ -z "$(ZCAP_PREFIX)" ]; then echo "[warn] ZCAP_PREFIX not set; using vendored .zcap if present"; fi
	CGO_ENABLED=1 go build -tags zcap -o ./out/livecap ./examples/livecap
	@echo "Built ./out/livecap (dynamic zcap)"

example-livecap-static:
	@if [ "$(UNAME_S)" != "Linux" ]; then echo "example-livecap-static: requires Linux"; exit 1; fi
	@if [ ! -d ./.zcap/linux/amd64/include ] && [ ! -d ./.zcap/linux/arm64/include ]; then \
		echo "[auto] fetching zcap artifacts"; \
		bash scripts/fetch_zcap.sh || { echo "fetch_zcap failed"; exit 1; }; \
	fi
	@if [ -z "$(ZCAP_PREFIX)" ]; then echo "[warn] ZCAP_PREFIX not set; using vendored .zcap if present"; fi
	CGO_ENABLED=1 go build -tags "zcap zcapstatic" -ldflags "-linkmode=external" -o ./out/livecap-static ./examples/livecap || { echo "static build attempt failed (ensure static libzcap available)"; exit 1; }
	@echo "Built ./out/livecap-static (static attempt)"

# ------------------------------------------------------------------------
# Downstream import test (ensures CGO directives usable by dependents)
# ------------------------------------------------------------------------
.PHONY: test-live-import
test-live-import:
	@if [ "$(UNAME_S)" != "Linux" ]; then echo "test-live-import: requires Linux"; exit 0; fi
	@if [ ! -d ./.zcap/linux/amd64/include ] && [ ! -d ./.zcap/linux/arm64/include ]; then \
		echo "[auto] fetching zcap artifacts"; \
		bash scripts/fetch_zcap.sh || { echo "fetch_zcap failed"; exit 1; }; \
	fi
	CGO_ENABLED=1 go test -c -tags zcap ./internal/captest
	CGO_ENABLED=1 go test -c -tags "zcap zcapstatic" ./internal/captest

#-------------------------------------------------------------------------
# Force targets
#-------------------------------------------------------------------------

FORCE: 

#-------------------------------------------------------------------------
# Phony targets
#-------------------------------------------------------------------------

.PHONY: build test test-live test-integration pcapdump pcapdump-static lint fuzz all clean guards FORCE
