# Canary build file.
#
# Every target is fail-closed: a non-zero exit from any command fails the
# target. No recipe swallows a non-zero exit, and every downloaded tool is
# pinned to an exact version and fetched reproducibly via `go run
# <module>@<version>` so a CI run and a developer run resolve the same binary.

GO ?= go

# Pinned tool versions. golangci-lint v2.13+ is built with go1.27; older
# releases refuse to lint go1.27 modules (mirrors .gitlab-ci.yml's go-lint
# golangci_version). gosec and govulncheck track their latest go1.27-compatible
# releases.
GOLANGCI_LINT_VERSION := v2.13.2
# gosec < v2.29 aborts under go1.27 with "package fmt without types imported
# from command-line-arguments"; v2.29.0 type-checks go1.27 modules cleanly.
GOSEC_VERSION := v2.29.0
GOVULNCHECK_VERSION := v1.1.4

# Project key from .canary/project.yaml (project.key). Evidence and verify are
# scoped to it so records bind to this project, not a default.
CANARY_PROJECT := CBIN

BIN := bin/canary

.PHONY: all build test test-race lint security verify bench fuzz clean

# Default target: compile everything and run the unit tests -- the two checks
# that need no network and gate every change.
all: build test

build:
	$(GO) build -trimpath -o $(BIN) ./cmd/canary
	$(GO) build ./...

test:
	$(GO) test -count=1 ./...

test-race:
	CGO_ENABLED=1 $(GO) test -count=1 -race ./...

lint:
	$(GO) vet ./...
	$(GO) run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@$(GOLANGCI_LINT_VERSION) run ./...

security:
	$(GO) run github.com/securego/gosec/v2/cmd/gosec@$(GOSEC_VERSION) ./...
	$(GO) run golang.org/x/vuln/cmd/govulncheck@$(GOVULNCHECK_VERSION) ./...

# verify proves the gap-analysis claims with evidence rather than trusting
# STATUS= declarations: build the CLI, scan, turn this run's passing tests into
# evidence bound to HEAD, ingest it, then verify. Any failing step fails the
# target.
verify: build
	./$(BIN) scan --root . --out status.json
	COMMIT=$$(git rev-parse HEAD); $(GO) test -count=1 -json ./... > /tmp/canary-gotest.json && \
	./$(BIN) evidence from-go-test --project $(CANARY_PROJECT) --commit $$COMMIT < /tmp/canary-gotest.json > /tmp/canary-evidence.json && \
	./$(BIN) evidence ingest --in /tmp/canary-evidence.json --out .canary/evidence.json && \
	./$(BIN) verify --project $(CANARY_PROJECT) --format json

bench:
	$(GO) test -run '^$$' -bench . -benchmem ./pkg/canaryscan/... ./pkg/specs/... ./pkg/storage/...

fuzz:
	$(GO) test -run '^$$' -fuzz FuzzSerializeRoundTrip -fuzztime 15s ./pkg/canaryscan

clean:
	rm -rf bin dist out status.json status.csv
