# Canary CLI — Next Slices

## Completed (This Evaluation)

### ✅ Evidence-Based Gap Analysis (2025-10-15)
- **Scope:** Built canary binary, ran 4 acceptance tests (ALL PASS), scanned `tools/canary` for tokens, verified self-canary dogfood.
- **Artifacts:**
  - `GAP_ANALYSIS.md` — Updated with tested truth, 10 cross-cutting gaps identified
  - `CHECKLIST.md` — Created with evidence links for CBIN-101, CBIN-102, CBIN-103
  - `tools-canary-status.json` — 3 BENCHED requirements via auto-promotion
- **Acceptance:**
  ```bash
  go test ./tools/canary/internal -run TestAcceptance -v
  # Output: 4/4 PASS
  # - TestAcceptance_FixtureSummary: {"summary":{"by_status":{"IMPL":1,"STUB":1}}}
  # - TestAcceptance_Overclaim: ACCEPT Overclaim Exit=2
  # - TestAcceptance_Stale: ACCEPT Stale Exit=2
  # - TestAcceptance_SelfCanary: ACCEPT SelfCanary OK ids=[CBIN-101,CBIN-102]
  ```
- **Status:** TESTED (acceptance validation complete)

---

## Up Next (Small, Verifiable Slices)

### Slice 1: Create TestCANARY_CBIN_101_Engine_ScanBasic
**Scope:** Implement canonical test function matching CBIN-101 token reference.
**File:** `tools/canary/main_test.go` (new file)
**Acceptance:**
```bash
go test ./tools/canary -run TestCANARY_CBIN_101 -v
# Expected stdout:
# === RUN   TestCANARY_CBIN_101_Engine_ScanBasic
# --- PASS: TestCANARY_CBIN_101_Engine_ScanBasic (0.XXs)
# PASS
```
**Test Body:**
- Create temp dir with 3 token fixtures (CBIN-200, CBIN-201, CBIN-202)
- Call `scan(root, skipRe)` function from `main.go`
- Assert `rep.Summary.ByStatus["STUB"] == 1`, `rep.Summary.ByStatus["IMPL"] == 2`
- Assert `len(rep.Requirements) == 3`
**CANARY:** Update `tools/canary/main.go:3` to `STATUS=TESTED` once test passes.

---

### Slice 2: Create TestCANARY_CBIN_102_CLI_Verify
**Scope:** Implement canonical test function matching CBIN-102 token reference.
**File:** `tools/canary/verify_test.go` (new file)
**Acceptance:**
```bash
go test ./tools/canary -run TestCANARY_CBIN_102 -v
# Expected stdout:
# === RUN   TestCANARY_CBIN_102_CLI_Verify
# --- PASS: TestCANARY_CBIN_102_CLI_Verify (0.XXs)
# PASS
```
**Test Body:**
- Create temp GAP file with `✅ CBIN-999` claim
- Create temp repo with `CBIN-888` token (no CBIN-999)
- Call `verifyClaims(rep, gapPath)` from `verify.go`
- Assert diagnostics contain `CANARY_VERIFY_FAIL REQ=CBIN-999`
**CANARY:** Update `tools/canary/verify.go:3` to `STATUS=TESTED` once test passes.

---

### Slice 3: Create TestCANARY_CBIN_103_API_StatusSchema
**Scope:** Implement canonical test function matching CBIN-103 token reference.
**File:** `tools/canary/status_test.go` (new file)
**Acceptance:**
```bash
go test ./tools/canary -run TestCANARY_CBIN_103 -v
# Expected stdout:
# === RUN   TestCANARY_CBIN_103_API_StatusSchema
# --- PASS: TestCANARY_CBIN_103_API_StatusSchema (0.XXs)
# PASS
```
**Test Body:**
- Create `Report` struct with 2 requirements (CBIN-101, CBIN-102)
- Marshal to JSON via `writeJSON` or direct `json.Marshal`
- Assert JSON contains keys: `generated_at`, `requirements`, `summary.by_status`, `summary.by_aspect`
- Assert sorted key order: `by_aspect` before `by_status` in `summary`
**CANARY:** Update `tools/canary/status.go:3` to `STATUS=TESTED` once test passes.

---

### Slice 4: Create BenchmarkCANARY_CBIN_101_Engine_Scan
**Scope:** Implement benchmark matching CBIN-101 token reference.
**File:** `tools/canary/main_test.go`
**Acceptance:**
```bash
go test ./tools/canary -bench BenchmarkCANARY_CBIN_101 -run ^$
# Expected output:
# BenchmarkCANARY_CBIN_101_Engine_Scan-X    NNNN   YYYY ns/op   Z allocs/op
# PASS
```
**Benchmark Body:**
```go
func BenchmarkCANARY_CBIN_101_Engine_Scan(b *testing.B) {
    dir := setupFixture(b, 100) // 100 files with CANARY tokens
    skip := skipDefault
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        _, err := scan(dir, skip)
        if err != nil {
            b.Fatal(err)
        }
    }
}
```
**Regression Guard:** `allocs/op ≤ 10` (baseline to be established)
**CANARY:** Update `tools/canary/main.go:3` to `STATUS=BENCHED` once bench runs.

---

### Slice 5: Create BenchmarkCANARY_CBIN_102_CLI_Verify
**Scope:** Implement benchmark matching CBIN-102 token reference.
**File:** `tools/canary/verify_test.go`
**Acceptance:**
```bash
go test ./tools/canary -bench BenchmarkCANARY_CBIN_102 -run ^$
# Expected output:
# BenchmarkCANARY_CBIN_102_CLI_Verify-X    NNNN   YYYY ns/op   Z allocs/op
# PASS
```
**Benchmark Body:**
- Setup: GAP file with 50 claims, repo scan result with 50 matching requirements
- Measure: `verifyClaims(rep, gapPath)` execution time
**Regression Guard:** `allocs/op ≤ 5` (baseline to be established)
**CANARY:** Update `tools/canary/verify.go:3` to `STATUS=BENCHED` once bench runs.

---

### Slice 6: Create BenchmarkCANARY_CBIN_103_API_Emit
**Scope:** Implement benchmark matching CBIN-103 token reference.
**File:** `tools/canary/status_test.go`
**Acceptance:**
```bash
go test ./tools/canary -bench BenchmarkCANARY_CBIN_103 -run ^$
# Expected output:
# BenchmarkCANARY_CBIN_103_API_Emit-X    NNNN   YYYY ns/op   Z allocs/op
# PASS
```
**Benchmark Body:**
- Setup: `Report` with 100 requirements, 300 features
- Measure: `writeJSON` and `writeCSV` execution time
**Regression Guard:** `allocs/op ≤ 15` (baseline to be established)
**CANARY:** Update `tools/canary/status.go:3` to `STATUS=BENCHED` once bench runs.

---

## Prioritization Rationale

1. **Slices 1-3 (TestCANARY_*)** close the highest-priority gap (token references without actual tests). These enable proper evidence validation and self-consistency.
2. **Slices 4-6 (BenchmarkCANARY_*)** establish performance baselines and enable future regression detection. Required before claiming BENCHED status.
3. All slices are **independently testable** with exact acceptance commands.
4. Each slice is **<100 LOC** and **<1 hour** of implementation time.
5. Completing Slices 1-6 moves CBIN-101, CBIN-102, CBIN-103 from "auto-promoted BENCHED" to "actually BENCHED with evidence."

## Dependencies & Sequencing

- Slices 1-3 can be done in **parallel** (independent test files).
- Slices 4-6 **depend on** Slices 1-3 (benchmarks reference same fixtures/functions as tests).
- After Slice 6: Update `GAP_ANALYSIS.md` and `CHECKLIST.md` to reflect new test/bench evidence.

## Post-Slices 1-6: Next Priorities

1. **Fix CRUSH.md placeholder** — Remove `ASPECT=<ASPECT>` to allow full-repo scanning.
2. **Add CI workflow** — `.github/workflows/canary.yml` to run acceptance tests + verify gate on PR.
3. **CSV row order test** — Validate deterministic sort (by REQ ID, then feature name).
4. **Performance benchmark (50k files)** — Add `BenchmarkCANARY_CBIN_101_Perf50k` with large fixture.
