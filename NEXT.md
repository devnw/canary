# Canary CLI — Next Slices

## Completed

### ✅ Evidence-Based Gap Analysis (2025-10-15 Initial)
- **Scope:** Built canary binary, ran acceptance tests, scanned `tools/canary` for tokens, verified self-canary dogfood.
- **Artifacts:** `GAP_ANALYSIS.md`, `CHECKLIST.md`, `tools-canary-status.json`
- **Status:** TESTED (acceptance validation complete)

### ✅ Phase 1: TestCANARY_* Functions (2025-10-15)
- **Scope:** Implemented 3 unit tests matching token references exactly
- **Deliverables:**
  - `tools/canary/main_test.go` — TestCANARY_CBIN_101_Engine_ScanBasic
  - `tools/canary/verify_test.go` — TestCANARY_CBIN_102_CLI_Verify
  - `tools/canary/status_test.go` — TestCANARY_CBIN_103_API_StatusSchema
- **Validation:** `go test -run TestCANARY_CBIN -v` → 3/3 PASS
- **Status:** COMPLETED (all tests pass, names match token refs)

### ✅ Phase 2: BenchmarkCANARY_* Functions (2025-10-15)
- **Scope:** Implemented 3 performance benchmarks with baseline measurements
- **Deliverables:**
  - BenchmarkCANARY_CBIN_101_Engine_Scan → 5.7ms/100 files, 1.1MB, 11357 allocs
  - BenchmarkCANARY_CBIN_102_CLI_Verify → 55µs/50 claims, 5.2KB, 13 allocs
  - BenchmarkCANARY_CBIN_103_API_Emit → 1.3ms/300 tokens, 36KB, 2119 allocs
- **Token Updates:** All 3 tokens → STATUS=BENCHED, UPDATED=2025-10-15
- **Validation:** `go test -bench BenchmarkCANARY -run ^$ -benchmem` → 3/3 RUN
- **Status:** COMPLETED (baselines established, extrapolated 50k perf: ~2.85s < 10s target)

### ✅ Phase 3: Documentation Updates (2025-10-15)
- **Scope:** Updated GAP_ANALYSIS.md, CHECKLIST.md, NEXT.md with Phase 1 & 2 results
- **Changes:**
  - Marked gaps #1 and #2 as RESOLVED in all docs
  - Added benchmark baselines to GAP_ANALYSIS.md
  - Updated evidence collection commands
  - Moved Slices 1-6 from "Up Next" to "Completed"
- **Status:** COMPLETED

---

## Up Next (Prioritized Slices)

### Slice 7: Fix CRUSH.md Placeholder
**Scope:** Remove invalid CANARY token causing parse errors
**File:** `CRUSH.md:27`
**Issue:** Line contains `ASPECT=<ASPECT>` placeholder causing `CANARY_PARSE_ERROR`
**Fix Options:**
1. Replace with valid ASPECT enum value (e.g., `ASPECT=Docs`)
2. Remove the example token entirely
3. Comment out the token line with additional `#` prefix
**Acceptance:**
```bash
./bin/canary --root . --out status.json --skip '(^|/)(.git|.direnv|node_modules)($|/)'
# Expected: EXIT=0 (no parse error)
```
**Estimated Time:** 5 minutes

---

### Slice 8: Add CI Workflow
**Scope:** Create GitHub Actions workflow for canary validation
**File:** `.github/workflows/canary.yml` (NEW)
**Jobs:**
1. **build** — `go build -o ./bin/canary ./tools/canary`
2. **test-unit** — `go test ./tools/canary -v`
3. **test-acceptance** — `go test ./tools/canary/internal -v`
4. **benchmark** — `go test ./tools/canary -bench BenchmarkCANARY -run ^$ -benchmem`
5. **verify-self** — `./bin/canary --root tools/canary --verify GAP_SELF.md --strict`
**Triggers:** `push` to `main`, `pull_request` to `main`
**Acceptance:**
```bash
# After push to branch:
# GitHub Actions UI shows workflow run
# All 5 jobs pass (green checkmarks)
```
**Estimated Time:** 1 hour

---

### Slice 9: CSV Row Order Test
**Scope:** Validate deterministic CSV row ordering
**File:** `tools/canary/csv_test.go` (NEW)
**Test:** `TestAcceptance_CSVOrder`
**Logic:**
- Create fixture with 10 requirements in random order (CBIN-050, CBIN-010, CBIN-030, ...)
- Scan and generate CSV
- Read CSV rows
- Assert rows sorted by REQ ID ascending (CBIN-010, CBIN-030, CBIN-050, ...)
**Acceptance:**
```bash
go test ./tools/canary -run TestAcceptance_CSVOrder -v
# Expected: PASS
```
**Estimated Time:** 1 hour

---

### Slice 10: Large-Scale Performance Benchmark (50k files)
**Scope:** Definitive validation of <10s requirement
**File:** `tools/canary/perf_test.go` (NEW)
**Benchmark:** `BenchmarkCANARY_CBIN_101_Perf50k`
**Setup:**
- Generate 50,000 files with CANARY tokens in temp dir
- Mix of file types (.go, .py, .js, .md, etc.)
- Realistic directory tree (depth 3-5, 100-500 files per dir)
**Measure:**
- Scan time (target: <10s)
- RSS memory (target: ≤512 MiB)
- Allocations per file
**Acceptance:**
```bash
go test ./tools/canary -bench BenchmarkCANARY_CBIN_101_Perf50k -run ^$ -benchmem -timeout 30s
# Expected:
# BenchmarkCANARY_CBIN_101_Perf50k-X    1    YYYY ns/op (<10s)    ZZZ MB/op (≤512 MiB)
# PASS
```
**Notes:**
- May need to adjust buffer sizes or enable parallel scanning
- First run will establish baseline for future regression detection
**Estimated Time:** 2-3 hours

---

## Prioritization Rationale

**Slices 7-10** address the remaining open gaps:
1. **Slice 7 (CRUSH.md fix)** — Quick win (5 min), unblocks full-repo scanning
2. **Slice 8 (CI)** — High value, enables automated validation on every PR
3. **Slice 9 (CSV order)** — Closes determinism gap, enables reproducible builds
4. **Slice 10 (50k perf)** — Definitive validation of performance requirement

**Estimated Total Time:** 4-5 hours (can parallelize Slices 7 & 9)

## Dependencies & Sequencing

- **Slice 7** — Independent, can be done first
- **Slice 8** — Should be done after Slice 7 (CI runs on clean repo)
- **Slice 9** — Independent, can be done in parallel with Slice 7
- **Slice 10** — Can be done anytime, but may inform optimizations

**Recommendation:** Slice 7 → Slice 8 → (Slice 9 + Slice 10 in parallel)

## Success Metrics

After completing Slices 7-10:
- ✅ All 5 critical gaps resolved
- ✅ CI enforces test/bench/verify on every PR
- ✅ CSV output deterministic and reproducible
- ✅ Performance validated at scale (50k files)
- ✅ Full repo scannable without parse errors
- ✅ CHECKLIST.md shows ✅ for all CBIN-101/102/103 capabilities
