# Canary CLI — Implementation Plan

**Generated:** 2025-10-15
**Based on:** GAP_ANALYSIS.md, CHECKLIST.md, NEXT.md

## Executive Summary

**Current State:**
- ✅ Working scanner in `tools/canary/` (builds successfully)
- ✅ 4/4 acceptance tests PASS with exact expected outputs
- ✅ Core capabilities proven: token parsing, enum validation, JSON/CSV export, verify gate, staleness checks
- ❌ **Critical Gap:** NO TestCANARY_* or BenchmarkCANARY_* functions exist (tokens reference phantom tests)
- ❌ NO CI integration
- ❌ NO performance benchmarks for 50k file requirement

**Goal:** Close the critical gaps in 6 small, verifiable slices to achieve full parity between token claims and actual test evidence.

**Timeline:** 6 slices × ~1 hour each = **~6 hours total** (can parallelize Slices 1-3)

---

## Phase 1: Create TestCANARY_* Functions (Slices 1-3)
**Duration:** 2-3 hours (parallel execution possible)
**Objective:** Align token references with actual test functions

### Slice 1: TestCANARY_CBIN_101_Engine_ScanBasic

**File:** `tools/canary/main_test.go` (NEW)

**Steps:**
1. Create new test file with package declaration:
   ```go
   package main

   import (
       "os"
       "path/filepath"
       "testing"
   )
   ```

2. Implement test function:
   ```go
   func TestCANARY_CBIN_101_Engine_ScanBasic(t *testing.T) {
       // Setup: temp dir with 3 fixture files
       dir := t.TempDir()
       fixtures := map[string]string{
           "file1.go": `package p
   // CANARY: REQ=CBIN-200; FEATURE="Alpha"; ASPECT=API; STATUS=STUB; UPDATED=2025-09-20
   `,
           "file2.go": `package p
   // CANARY: REQ=CBIN-201; FEATURE="Bravo"; ASPECT=CLI; STATUS=IMPL; UPDATED=2025-09-20
   `,
           "file3.go": `package p
   // CANARY: REQ=CBIN-202; FEATURE="Charlie"; ASPECT=Engine; STATUS=IMPL; UPDATED=2025-09-20
   `,
       }
       for name, content := range fixtures {
           if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0o644); err != nil {
               t.Fatal(err)
           }
       }

       // Execute: scan directory
       rep, err := scan(dir, skipDefault)
       if err != nil {
           t.Fatalf("scan failed: %v", err)
       }

       // Verify: status counts
       if rep.Summary.ByStatus["STUB"] != 1 {
           t.Errorf("expected STUB=1, got %d", rep.Summary.ByStatus["STUB"])
       }
       if rep.Summary.ByStatus["IMPL"] != 2 {
           t.Errorf("expected IMPL=2, got %d", rep.Summary.ByStatus["IMPL"])
       }

       // Verify: requirement count
       if len(rep.Requirements) != 3 {
           t.Errorf("expected 3 requirements, got %d", len(rep.Requirements))
       }

       // Verify: aspect diversity
       if rep.Summary.ByAspect["API"] != 1 {
           t.Errorf("expected API=1, got %d", rep.Summary.ByAspect["API"])
       }
       if rep.Summary.ByAspect["CLI"] != 1 {
           t.Errorf("expected CLI=1, got %d", rep.Summary.ByAspect["CLI"])
       }
       if rep.Summary.ByAspect["Engine"] != 1 {
           t.Errorf("expected Engine=1, got %d", rep.Summary.ByAspect["Engine"])
       }
   }
   ```

3. Run test:
   ```bash
   cd tools/canary
   go test -run TestCANARY_CBIN_101 -v
   ```

4. **Acceptance Criteria:**
   - Test compiles without errors
   - Test passes (exit 0)
   - Output shows: `--- PASS: TestCANARY_CBIN_101_Engine_ScanBasic`

5. **Update Token:** Change `tools/canary/main.go:3` from `STATUS=TESTED` to `STATUS=TESTED` (already TESTED, confirm evidence exists)

**Blockers:** None (scanner functions already exist)

---

### Slice 2: TestCANARY_CBIN_102_CLI_Verify

**File:** `tools/canary/verify_test.go` (NEW)

**Steps:**
1. Create new test file:
   ```go
   package main

   import (
       "os"
       "path/filepath"
       "strings"
       "testing"
   )
   ```

2. Implement test function:
   ```go
   func TestCANARY_CBIN_102_CLI_Verify(t *testing.T) {
       // Setup: temp GAP file with overclaim
       dir := t.TempDir()
       gapFile := filepath.Join(dir, "GAP.md")
       gapContent := `# Gap Analysis

   ## Implemented
   ✅ CBIN-999
   ✅ CBIN-888
   `
       if err := os.WriteFile(gapFile, []byte(gapContent), 0o644); err != nil {
           t.Fatal(err)
       }

       // Setup: repo with only CBIN-888 token (CBIN-999 missing)
       repoDir := t.TempDir()
       repoFile := filepath.Join(repoDir, "code.go")
       repoContent := `package p
   // CANARY: REQ=CBIN-888; FEATURE="Present"; ASPECT=API; STATUS=IMPL; UPDATED=2025-09-20
   `
       if err := os.WriteFile(repoFile, []byte(repoContent), 0o644); err != nil {
           t.Fatal(err)
       }

       // Execute: scan repo
       rep, err := scan(repoDir, skipDefault)
       if err != nil {
           t.Fatalf("scan failed: %v", err)
       }

       // Execute: verify claims
       diags := verifyClaims(rep, gapFile)

       // Verify: overclaim detected
       if len(diags) == 0 {
           t.Fatal("expected verification failures, got none")
       }

       found := false
       for _, d := range diags {
           if strings.Contains(d, "CANARY_VERIFY_FAIL") && strings.Contains(d, "CBIN-999") {
               found = true
               break
           }
       }
       if !found {
           t.Errorf("expected CANARY_VERIFY_FAIL for CBIN-999, got: %v", diags)
       }
   }
   ```

3. Run test:
   ```bash
   cd tools/canary
   go test -run TestCANARY_CBIN_102 -v
   ```

4. **Acceptance Criteria:**
   - Test compiles without errors
   - Test passes (exit 0)
   - Output shows: `--- PASS: TestCANARY_CBIN_102_CLI_Verify`

5. **Update Token:** Confirm `tools/canary/verify.go:3` STATUS remains `TESTED`

**Blockers:** None (verifyClaims function already exists)

---

### Slice 3: TestCANARY_CBIN_103_API_StatusSchema

**File:** `tools/canary/status_test.go` (NEW)

**Steps:**
1. Create new test file:
   ```go
   package main

   import (
       "encoding/json"
       "os"
       "path/filepath"
       "testing"
   )
   ```

2. Implement test function:
   ```go
   func TestCANARY_CBIN_103_API_StatusSchema(t *testing.T) {
       // Setup: create report with known data
       rep := Report{
           GeneratedAt: "2025-10-15T00:00:00Z",
           Requirements: []Requirement{
               {
                   ID: "CBIN-101",
                   Features: []Feature{
                       {
                           Feature: "ScannerCore",
                           Aspect:  "Engine",
                           Status:  "TESTED",
                           Files:   []string{"main.go"},
                           Tests:   []string{"TestCANARY_CBIN_101_Engine_ScanBasic"},
                           Benches: []string{},
                           Owner:   "canary",
                           Updated: "2025-09-20",
                       },
                   },
               },
               {
                   ID: "CBIN-102",
                   Features: []Feature{
                       {
                           Feature: "VerifyGate",
                           Aspect:  "CLI",
                           Status:  "TESTED",
                           Files:   []string{"verify.go"},
                           Tests:   []string{"TestCANARY_CBIN_102_CLI_Verify"},
                           Benches: []string{},
                           Owner:   "canary",
                           Updated: "2025-09-20",
                       },
                   },
               },
           },
           Summary: Summary{
               ByStatus:           StatusCounts{"TESTED": 2},
               ByAspect:           AspectCounts{"Engine": 1, "CLI": 1},
               TotalTokens:        2,
               UniqueRequirements: 2,
           },
       }

       // Execute: marshal to JSON
       b, err := json.Marshal(rep)
       if err != nil {
           t.Fatalf("marshal failed: %v", err)
       }

       // Verify: valid JSON
       var parsed map[string]interface{}
       if err := json.Unmarshal(b, &parsed); err != nil {
           t.Fatalf("unmarshal failed: %v", err)
       }

       // Verify: top-level keys present
       requiredKeys := []string{"generated_at", "requirements", "summary"}
       for _, key := range requiredKeys {
           if _, ok := parsed[key]; !ok {
               t.Errorf("missing required key: %s", key)
           }
       }

       // Verify: summary structure
       summary, ok := parsed["summary"].(map[string]interface{})
       if !ok {
           t.Fatal("summary is not a map")
       }
       summaryKeys := []string{"by_status", "by_aspect", "total_tokens", "unique_requirements"}
       for _, key := range summaryKeys {
           if _, ok := summary[key]; !ok {
               t.Errorf("summary missing key: %s", key)
           }
       }

       // Verify: key ordering (by_aspect before by_status alphabetically)
       jsonStr := string(b)
       aspectIdx := -1
       statusIdx := -1
       // Simple check: "by_aspect" appears before "by_status" in JSON string
       for i := 0; i < len(jsonStr)-10; i++ {
           if aspectIdx < 0 && jsonStr[i:i+10] == `"by_aspect` {
               aspectIdx = i
           }
           if statusIdx < 0 && jsonStr[i:i+10] == `"by_status` {
               statusIdx = i
           }
       }
       if aspectIdx < 0 || statusIdx < 0 {
           t.Error("could not find by_aspect or by_status in JSON")
       } else if aspectIdx > statusIdx {
           t.Error("expected by_aspect before by_status (sorted key order)")
       }
   }
   ```

3. Run test:
   ```bash
   cd tools/canary
   go test -run TestCANARY_CBIN_103 -v
   ```

4. **Acceptance Criteria:**
   - Test compiles without errors
   - Test passes (exit 0)
   - Output shows: `--- PASS: TestCANARY_CBIN_103_API_StatusSchema`

5. **Update Token:** Confirm `tools/canary/status.go:3` STATUS remains `TESTED`

**Blockers:** None (Report struct and marshalers already exist)

---

### Phase 1 Validation

**After completing Slices 1-3:**
```bash
cd tools/canary
go test -run TestCANARY_CBIN -v
# Expected: 3/3 PASS
```

**Update Evidence:**
- Run: `grep -r "^func.*TestCANARY" . --include="*_test.go"`
- Confirm: 3 functions found
- Update `CHECKLIST.md`: Change "TestCANARY_* functions missing" from gap to ✅

---

## Phase 2: Create BenchmarkCANARY_* Functions (Slices 4-6)
**Duration:** 2-3 hours (depends on Phase 1 completion)
**Objective:** Establish performance baselines and enable regression detection

### Slice 4: BenchmarkCANARY_CBIN_101_Engine_Scan

**File:** `tools/canary/main_test.go` (APPEND)

**Steps:**
1. Add helper function to create large fixture:
   ```go
   func setupFixture(tb testing.TB, numFiles int) string {
       tb.Helper()
       dir := tb.TempDir()
       for i := 0; i < numFiles; i++ {
           content := fmt.Sprintf(`package p
   // CANARY: REQ=CBIN-%03d; FEATURE="Feature%d"; ASPECT=API; STATUS=IMPL; UPDATED=2025-09-20
   `, i, i)
           if err := os.WriteFile(filepath.Join(dir, fmt.Sprintf("file%d.go", i)), []byte(content), 0o644); err != nil {
               tb.Fatal(err)
           }
       }
       return dir
   }
   ```

2. Implement benchmark:
   ```go
   func BenchmarkCANARY_CBIN_101_Engine_Scan(b *testing.B) {
       dir := setupFixture(b, 100)
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

3. Run benchmark:
   ```bash
   cd tools/canary
   go test -bench BenchmarkCANARY_CBIN_101 -run ^$ -benchmem
   ```

4. **Acceptance Criteria:**
   - Benchmark compiles without errors
   - Output shows: `BenchmarkCANARY_CBIN_101_Engine_Scan-X    NNNN   YYYY ns/op   Z allocs/op`
   - Baseline established: Record `ns/op` and `allocs/op` values

5. **Update Token:** Change `tools/canary/main.go:3` from `STATUS=TESTED` to `STATUS=BENCHED`

**Regression Guard:** After establishing baseline, add check: `allocs/op ≤ baseline + 20%`

---

### Slice 5: BenchmarkCANARY_CBIN_102_CLI_Verify

**File:** `tools/canary/verify_test.go` (APPEND)

**Steps:**
1. Add helper to create large GAP file:
   ```go
   func setupGAPFixture(tb testing.TB, numClaims int) (string, *Report) {
       tb.Helper()
       dir := tb.TempDir()

       // Create GAP file with N claims
       gapFile := filepath.Join(dir, "GAP.md")
       var gapContent strings.Builder
       gapContent.WriteString("# Gap Analysis\n\n")
       for i := 0; i < numClaims; i++ {
           fmt.Fprintf(&gapContent, "✅ CBIN-%03d\n", i)
       }
       if err := os.WriteFile(gapFile, []byte(gapContent.String()), 0o644); err != nil {
           tb.Fatal(err)
       }

       // Create matching report with N requirements
       rep := &Report{
           GeneratedAt:  "2025-10-15T00:00:00Z",
           Requirements: make([]Requirement, numClaims),
           Summary: Summary{
               ByStatus:           StatusCounts{"IMPL": numClaims},
               ByAspect:           AspectCounts{"API": numClaims},
               TotalTokens:        numClaims,
               UniqueRequirements: numClaims,
           },
       }
       for i := 0; i < numClaims; i++ {
           rep.Requirements[i] = Requirement{
               ID: fmt.Sprintf("CBIN-%03d", i),
               Features: []Feature{
                   {Feature: fmt.Sprintf("Feature%d", i), Aspect: "API", Status: "IMPL", Updated: "2025-09-20"},
               },
           }
       }

       return gapFile, rep
   }
   ```

2. Implement benchmark:
   ```go
   func BenchmarkCANARY_CBIN_102_CLI_Verify(b *testing.B) {
       gapFile, rep := setupGAPFixture(b, 50)
       b.ResetTimer()
       for i := 0; i < b.N; i++ {
           _ = verifyClaims(rep, gapFile)
       }
   }
   ```

3. Run benchmark:
   ```bash
   cd tools/canary
   go test -bench BenchmarkCANARY_CBIN_102 -run ^$ -benchmem
   ```

4. **Acceptance Criteria:**
   - Benchmark compiles without errors
   - Output shows: `BenchmarkCANARY_CBIN_102_CLI_Verify-X    NNNN   YYYY ns/op   Z allocs/op`
   - Baseline established

5. **Update Token:** Change `tools/canary/verify.go:3` from `STATUS=TESTED` to `STATUS=BENCHED`

**Regression Guard:** `allocs/op ≤ 10` (verify should be low-allocation)

---

### Slice 6: BenchmarkCANARY_CBIN_103_API_Emit

**File:** `tools/canary/status_test.go` (APPEND)

**Steps:**
1. Add helper to create large report:
   ```go
   func setupLargeReport(tb testing.TB, numReqs int, featuresPerReq int) *Report {
       tb.Helper()
       rep := &Report{
           GeneratedAt:  "2025-10-15T00:00:00Z",
           Requirements: make([]Requirement, numReqs),
           Summary: Summary{
               ByStatus:           StatusCounts{},
               ByAspect:           AspectCounts{},
               TotalTokens:        numReqs * featuresPerReq,
               UniqueRequirements: numReqs,
           },
       }
       for i := 0; i < numReqs; i++ {
           features := make([]Feature, featuresPerReq)
           for j := 0; j < featuresPerReq; j++ {
               features[j] = Feature{
                   Feature: fmt.Sprintf("Feature%d_%d", i, j),
                   Aspect:  "API",
                   Status:  "IMPL",
                   Files:   []string{fmt.Sprintf("file%d_%d.go", i, j)},
                   Tests:   []string{},
                   Benches: []string{},
                   Owner:   "team",
                   Updated: "2025-09-20",
               }
               rep.Summary.ByStatus["IMPL"]++
               rep.Summary.ByAspect["API"]++
           }
           rep.Requirements[i] = Requirement{
               ID:       fmt.Sprintf("CBIN-%03d", i),
               Features: features,
           }
       }
       return rep
   }
   ```

2. Implement benchmark:
   ```go
   func BenchmarkCANARY_CBIN_103_API_Emit(b *testing.B) {
       rep := setupLargeReport(b, 100, 3) // 100 reqs × 3 features = 300 features
       dir := b.TempDir()
       jsonPath := filepath.Join(dir, "status.json")
       csvPath := filepath.Join(dir, "status.csv")

       b.ResetTimer()
       for i := 0; i < b.N; i++ {
           if err := writeJSON(jsonPath, rep); err != nil {
               b.Fatal(err)
           }
           if err := writeCSV(csvPath, rep); err != nil {
               b.Fatal(err)
           }
       }
   }
   ```

3. Run benchmark:
   ```bash
   cd tools/canary
   go test -bench BenchmarkCANARY_CBIN_103 -run ^$ -benchmem
   ```

4. **Acceptance Criteria:**
   - Benchmark compiles without errors
   - Output shows: `BenchmarkCANARY_CBIN_103_API_Emit-X    NNNN   YYYY ns/op   Z allocs/op`
   - Baseline established

5. **Update Token:** Change `tools/canary/status.go:3` from `STATUS=TESTED` to `STATUS=BENCHED`

**Regression Guard:** `allocs/op ≤ 20` (JSON/CSV emission should be reasonably efficient)

---

### Phase 2 Validation

**After completing Slices 4-6:**
```bash
cd tools/canary
go test -bench BenchmarkCANARY -run ^$ -benchmem
# Expected: 3 benchmarks run successfully
# Output format:
# BenchmarkCANARY_CBIN_101_Engine_Scan-X      ...
# BenchmarkCANARY_CBIN_102_CLI_Verify-X       ...
# BenchmarkCANARY_CBIN_103_API_Emit-X         ...
```

**Update Evidence:**
- Run: `grep -r "^func.*BenchmarkCANARY" . --include="*_test.go"`
- Confirm: 3 functions found
- Update `CHECKLIST.md`: Change "BenchmarkCANARY_* functions missing" from gap to ✅

---

## Phase 3: Update Documentation & Evidence (Post-Slices 1-6)

**Duration:** 30 minutes

### Step 1: Re-run Evidence Collection

```bash
# From repo root
cd /home/benji/src/spyder/canary

# Build latest
cd tools/canary
go build -o ../../bin/canary .
cd ../..

# Scan tools/canary with updated tests/benches
./bin/canary --root tools/canary --out tools-canary-status-v2.json --csv tools-canary-status-v2.csv

# Verify self-canary still passes
./bin/canary --root tools/canary --verify GAP_SELF.md --strict
echo "Exit code: $?"  # Should be 0

# Run all tests
go test ./tools/canary -v

# Run all benchmarks
go test ./tools/canary -bench . -run ^$ -benchmem
```

### Step 2: Update CHECKLIST.md

Replace "Critical Gaps Summary" section:
```markdown
## Critical Gaps Summary

1. ~~**TestCANARY_* functions missing**~~ ✅ RESOLVED — 3 functions implemented
2. ~~**BenchmarkCANARY_* functions missing**~~ ✅ RESOLVED — 3 functions implemented
3. **CI workflow missing** (no `.github/workflows/canary.yml`) — OPEN
4. **Performance benchmarks missing** (no `BenchmarkCANARY_CBIN_101_Perf50k` for 50k file test) — OPEN
5. **CSV row order not validated** (deterministic sort untested) — OPEN
```

### Step 3: Update GAP_ANALYSIS.md

Update "Cross-Cutting Gaps" section:
```markdown
## Cross-Cutting Gaps

1. ~~**TestCANARY_* functions missing**~~ ✅ RESOLVED (2025-10-15) — Implemented TestCANARY_CBIN_101_Engine_ScanBasic, TestCANARY_CBIN_102_CLI_Verify, TestCANARY_CBIN_103_API_StatusSchema

2. ~~**BenchmarkCANARY_* functions missing**~~ ✅ RESOLVED (2025-10-15) — Implemented BenchmarkCANARY_CBIN_101_Engine_Scan, BenchmarkCANARY_CBIN_102_CLI_Verify, BenchmarkCANARY_CBIN_103_API_Emit. Baselines: [record actual values]

3. **cmd/canary build failure** — [unchanged]
...
```

### Step 4: Update NEXT.md

Mark Slices 1-6 as completed:
```markdown
## Completed (Recent Slices)

### ✅ TestCANARY_* Functions (Slices 1-3) — 2025-10-15
- TestCANARY_CBIN_101_Engine_ScanBasic → `tools/canary/main_test.go`
- TestCANARY_CBIN_102_CLI_Verify → `tools/canary/verify_test.go`
- TestCANARY_CBIN_103_API_StatusSchema → `tools/canary/status_test.go`
- **Validation:** `go test -run TestCANARY_CBIN -v` → 3/3 PASS

### ✅ BenchmarkCANARY_* Functions (Slices 4-6) — 2025-10-15
- BenchmarkCANARY_CBIN_101_Engine_Scan → baseline: XXX ns/op, YYY allocs/op
- BenchmarkCANARY_CBIN_102_CLI_Verify → baseline: XXX ns/op, YYY allocs/op
- BenchmarkCANARY_CBIN_103_API_Emit → baseline: XXX ns/op, YYY allocs/op
- **Validation:** `go test -bench BenchmarkCANARY -run ^$` → 3/3 run

## Up Next (Post-Phase 2)

### Slice 7: Fix CRUSH.md Placeholder
[move from previous "Up Next"]

### Slice 8: Add CI Workflow
[move from previous plan]
...
```

---

## Phase 4: Next Priorities (Post-Phases 1-3)

**Duration:** 4-6 hours

### Priority 1: Fix CRUSH.md Placeholder (30 min)
- **Issue:** Line 27 has `ASPECT=<ASPECT>` causing parse errors
- **Fix:** Replace with valid example or remove token entirely
- **Validation:** `./bin/canary --root . --out status.json` exits 0

### Priority 2: Add CI Workflow (1 hour)
- **File:** `.github/workflows/canary.yml`
- **Jobs:**
  1. Build canary binary
  2. Run acceptance tests (`go test ./tools/canary/internal -v`)
  3. Run unit tests (`go test ./tools/canary -v`)
  4. Run verify gate (`./bin/canary --root tools/canary --verify GAP_SELF.md --strict`)
- **Trigger:** PR to main, push to main
- **Validation:** Push branch, verify workflow runs green

### Priority 3: CSV Row Order Test (1 hour)
- **File:** `tools/canary/csv_test.go` (NEW)
- **Test:** `TestAcceptance_CSVOrder`
- **Logic:** Create fixture with 10 requirements (mixed order), assert CSV rows sorted by REQ ID
- **Validation:** Test passes, CSV determinism proven

### Priority 4: Performance Benchmark (50k files) (2-3 hours)
- **File:** `tools/canary/perf_test.go` (NEW)
- **Benchmark:** `BenchmarkCANARY_CBIN_101_Perf50k`
- **Setup:** Generate 50k files with CANARY tokens in temp dir
- **Measure:** Scan time, RSS, allocs/op
- **Acceptance:** <10s scan time, ≤512 MiB RSS
- **Notes:** May require adjustments to buffer sizes, parallel scanning

---

## Risk Mitigation

### Risk 1: Tests depend on internal functions
**Mitigation:** Tests are in `package main` (same package as scanner), can access unexported functions like `scan()`, `verifyClaims()`, `writeJSON()`.

### Risk 2: Benchmark baselines unknown
**Mitigation:** First run establishes baseline. Document values in NEXT.md. Future runs compare against baseline with ±20% tolerance.

### Risk 3: Large fixture generation slow
**Mitigation:** Use `b.TempDir()` for auto-cleanup. For 50k file benchmark, cache fixture in global var with `sync.Once` initialization.

### Risk 4: CI workflow permissions
**Mitigation:** Use standard `actions/setup-go@v5`, no special permissions needed. Verify gate runs read-only (no push/tag).

---

## Success Criteria (End of Phase 2)

**Functional:**
- [x] 3 TestCANARY_* functions exist and pass
- [x] 3 BenchmarkCANARY_* functions exist and run
- [x] Token references align with actual test/bench names
- [x] All acceptance tests still pass (4/4)
- [x] Self-canary verification still passes (exit 0)

**Evidence:**
- [x] `grep -r "^func.*TestCANARY"` returns 3 functions
- [x] `grep -r "^func.*BenchmarkCANARY"` returns 3 functions
- [x] `go test ./tools/canary -v` → all tests green
- [x] `go test ./tools/canary -bench . -run ^$` → 3 benchmarks with ns/op, allocs/op

**Documentation:**
- [x] CHECKLIST.md updated (gaps 1-2 resolved)
- [x] GAP_ANALYSIS.md updated (cross-cutting gaps 1-2 resolved)
- [x] NEXT.md updated (slices 1-6 marked completed, new priorities listed)

**Measurement:**
- Baseline performance recorded for 100-file scans
- Regression guards established (allocs/op thresholds)
- Evidence trail: token → test function → test output

---

## Execution Checklist

Use this checklist while implementing:

### Phase 1: Tests
- [ ] Create `tools/canary/main_test.go` with TestCANARY_CBIN_101_Engine_ScanBasic
- [ ] Run `go test -run TestCANARY_CBIN_101 -v` → PASS
- [ ] Create `tools/canary/verify_test.go` with TestCANARY_CBIN_102_CLI_Verify
- [ ] Run `go test -run TestCANARY_CBIN_102 -v` → PASS
- [ ] Create `tools/canary/status_test.go` with TestCANARY_CBIN_103_API_StatusSchema
- [ ] Run `go test -run TestCANARY_CBIN_103 -v` → PASS
- [ ] Run `go test ./tools/canary -run TestCANARY_CBIN -v` → 3/3 PASS

### Phase 2: Benchmarks
- [ ] Add `setupFixture()` helper to `tools/canary/main_test.go`
- [ ] Add BenchmarkCANARY_CBIN_101_Engine_Scan to `tools/canary/main_test.go`
- [ ] Run `go test -bench BenchmarkCANARY_CBIN_101 -run ^$` → record baseline
- [ ] Add `setupGAPFixture()` helper to `tools/canary/verify_test.go`
- [ ] Add BenchmarkCANARY_CBIN_102_CLI_Verify to `tools/canary/verify_test.go`
- [ ] Run `go test -bench BenchmarkCANARY_CBIN_102 -run ^$` → record baseline
- [ ] Add `setupLargeReport()` helper to `tools/canary/status_test.go`
- [ ] Add BenchmarkCANARY_CBIN_103_API_Emit to `tools/canary/status_test.go`
- [ ] Run `go test -bench BenchmarkCANARY_CBIN_103 -run ^$` → record baseline
- [ ] Run `go test ./tools/canary -bench BenchmarkCANARY -run ^$` → 3/3 run

### Phase 3: Documentation
- [ ] Re-scan: `./bin/canary --root tools/canary --out tools-canary-status-v2.json`
- [ ] Verify: `./bin/canary --root tools/canary --verify GAP_SELF.md --strict` → exit 0
- [ ] Update CHECKLIST.md (mark gaps 1-2 resolved)
- [ ] Update GAP_ANALYSIS.md (mark cross-cutting gaps 1-2 resolved, record baselines)
- [ ] Update NEXT.md (move slices 1-6 to "Completed", add new priorities)
- [ ] Git commit with message: "feat: implement TestCANARY and BenchmarkCANARY functions for CBIN-101/102/103"

### Phase 4: Next Priorities
- [ ] Fix CRUSH.md placeholder (remove `ASPECT=<ASPECT>`)
- [ ] Create `.github/workflows/canary.yml`
- [ ] Implement TestAcceptance_CSVOrder
- [ ] Implement BenchmarkCANARY_CBIN_101_Perf50k

---

## Timeline Summary

| Phase | Duration | Parallel? | Deliverable |
|:------|:--------:|:---------:|:------------|
| Phase 1 (Slices 1-3) | 2-3 hours | ✅ Yes | 3 TestCANARY_* functions |
| Phase 2 (Slices 4-6) | 2-3 hours | ⚠️ Partial | 3 BenchmarkCANARY_* functions |
| Phase 3 (Docs) | 30 min | ❌ No | Updated GAP_ANALYSIS, CHECKLIST, NEXT |
| Phase 4 (Next priorities) | 4-6 hours | ⚠️ Partial | CI, CSV test, perf bench |
| **Total** | **9-13 hours** | | **Full parity achieved** |

**Recommendation:** Execute Phases 1-3 as single sprint (~6 hours), then schedule Phase 4 separately.
