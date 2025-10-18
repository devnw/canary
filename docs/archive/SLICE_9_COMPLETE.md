# Slice 9 Complete: CSV Row Order Test ✅

**Date:** 2025-10-15
**Duration:** ~15 minutes
**Status:** COMPLETED

## Summary

Created `TestAcceptance_CSVOrder` acceptance test to validate CSV output has deterministic row ordering. Test verifies that multiple scan runs produce identical CSV output with rows sorted by requirement ID.

## Problem

CSV output ordering was not explicitly tested. Without deterministic ordering, CSV files could change between runs even when input data is identical, making it difficult to:
- Track changes in version control
- Compare results across runs
- Use CSV for reliable reporting

## Solution

Created comprehensive acceptance test `TestAcceptance_CSVOrder` that:
1. Creates fixtures with multiple requirements and features in non-alphabetical filename order
2. Runs scanner twice on the same fixtures
3. Verifies CSV outputs are byte-for-byte identical (deterministic)
4. Validates rows are sorted by REQ ID

### Test Implementation

**File:** `tools/canary/internal/acceptance_test.go`
**Function:** `TestAcceptance_CSVOrder`
**Lines Added:** 72 (including test logic and fixtures)

### Test Fixtures

Created 5 test files with deliberately non-alphabetical filenames:
- `z.txt` → CBIN-999 (Zulu)
- `a.txt` → CBIN-101 (Alpha)
- `m.txt` → CBIN-500 (Mike)
- `b.txt` → CBIN-102 (Bravo)
- `c.txt` → CBIN-102 (Charlie) — Same REQ as Bravo to test secondary sort

### Test Validations

1. **Determinism Check:**
   - Run scanner twice with identical inputs
   - Compare CSV outputs byte-for-byte
   - Fail if outputs differ

2. **Sort Order Check:**
   - Parse CSV rows
   - Verify REQ IDs are in ascending order
   - Fail if any row has a REQ ID less than previous row

### Expected CSV Output

```csv
req,feature,aspect,status,file,test,bench,owner,updated
CBIN-101,Alpha,CLI,TESTED,/tmp/csvorder/a.txt,TestAlpha,,,2025-10-15
CBIN-102,Bravo,API,BENCHED,/tmp/csvorder/b.txt,TestBravo,BenchBravo,,2025-10-15
CBIN-102,Charlie,CLI,IMPL,/tmp/csvorder/c.txt,,,,2025-10-15
CBIN-500,Mike,Engine,STUB,/tmp/csvorder/m.txt,,,,2025-10-15
CBIN-999,Zulu,API,IMPL,/tmp/csvorder/z.txt,,,,2025-10-15
```

**Sort Key:** Primary = REQ ID (ascending), Secondary = Feature name (for same REQ)

## Validation Results

### Test Execution
```bash
$ go test ./tools/canary/internal -v -run TestAcceptance_CSVOrder
=== RUN   TestAcceptance_CSVOrder
ACCEPT CSVOrder deterministic and sorted
--- PASS: TestAcceptance_CSVOrder (0.08s)
PASS
ok  	go.spyder.org/canary/tools/canary/internal	0.081s
```
✅ **PASS** — CSV output is deterministic and sorted

### All Acceptance Tests
```bash
$ go test ./tools/canary/internal -v -run TestAcceptance
=== RUN   TestAcceptance_FixtureSummary
{"summary":{"by_status":{"IMPL":1,"STUB":1}}}
--- PASS: TestAcceptance_FixtureSummary (0.07s)
=== RUN   TestAcceptance_Overclaim
ACCEPT Overclaim Exit=2
--- PASS: TestAcceptance_Overclaim (0.07s)
=== RUN   TestAcceptance_Stale
ACCEPT Stale Exit=2
--- PASS: TestAcceptance_Stale (0.07s)
=== RUN   TestAcceptance_SelfCanary
ACCEPT SelfCanary OK ids=[CBIN-101,CBIN-102]
--- PASS: TestAcceptance_SelfCanary (0.08s)
=== RUN   TestAcceptance_CSVOrder
ACCEPT CSVOrder deterministic and sorted
--- PASS: TestAcceptance_CSVOrder (0.08s)
PASS
ok  	go.spyder.org/canary/tools/canary/internal	0.380s
```
✅ **5/5 PASS** (was 4/5, now includes CSVOrder)

### Manual Verification
```bash
$ ./bin/canary --root /tmp/csvorder --out status.json --csv status.csv
$ cat status.csv
req,feature,aspect,status,file,test,bench,owner,updated
CBIN-101,Alpha,CLI,TESTED,/tmp/csvorder/a.txt,TestAlpha,,,2025-10-15
CBIN-102,Bravo,API,BENCHED,/tmp/csvorder/b.txt,TestBravo,BenchBravo,,2025-10-15
CBIN-102,Charlie,CLI,IMPL,/tmp/csvorder/c.txt,,,,2025-10-15
CBIN-500,Mike,Engine,STUB,/tmp/csvorder/m.txt,,,,2025-10-15
CBIN-999,Zulu,API,IMPL,/tmp/csvorder/z.txt,,,,2025-10-15
```
✅ **Verified** — Rows sorted by CBIN-101 → CBIN-102 → CBIN-500 → CBIN-999

## Files Modified

### tools/canary/internal/acceptance_test.go
**Lines Added:** 72
**Location:** After `TestAcceptance_SelfCanary`, before `TestMetadata`

**Changes:**
- Added `TestAcceptance_CSVOrder` function
- Creates 5 test fixtures with diverse REQ IDs and features
- Runs scanner twice and compares outputs
- Validates CSV row ordering

## Success Criteria: ✅ ALL MET

- [x] `TestAcceptance_CSVOrder` created in `tools/canary/internal/acceptance_test.go`
- [x] Test creates fixtures with multiple requirements and features
- [x] Test verifies CSV output is byte-for-byte identical across runs
- [x] Test validates CSV rows are sorted by REQ ID
- [x] Test passes: ✅ PASS
- [x] All acceptance tests still pass: ✅ 5/5 PASS
- [x] CSV output manually verified to be sorted correctly

**Slice 9 Status: COMPLETE** 🎉

## Impact

**Before Slice 9:**
- CSV ordering not explicitly tested
- No guarantee of deterministic output
- Potential for unstable CSV diffs in version control
- Gap #4 (CSV row order untested) OPEN

**After Slice 9:**
- CSV ordering validated by acceptance test
- Deterministic output verified
- Stable CSV diffs for version control
- Gap #4 (CSV row order untested) ✅ RESOLVED

## Gap Status Update

**Gap #4: CSV row order untested**
- **Status Before:** ❌ OPEN
- **Status After:** ✅ RESOLVED (2025-10-15 Slice 9)
- **Evidence:** `tools/canary/internal/acceptance_test.go:136` (TestAcceptance_CSVOrder)
- **Test Output:** 5/5 acceptance tests PASS

## Next Steps

**Slice 10:** Large-Scale 50k File Performance Benchmark (2-3 hours)
- Create `BenchmarkCANARY_CBIN_101_Engine_Scan50k` with 50,000 file fixture
- Validate <10s target from CRUSH.md
- Establish baseline for regression detection
- Update Perf50k column in CHECKLIST.md from ◐ PARTIAL to ✅ DONE
- Resolve Gap #7 (50k file perf untested) from ◐ PARTIAL to ✅ RESOLVED

**Estimated Time:** 2-3 hours for Slice 10

## Test Code Reference

**Function:** `TestAcceptance_CSVOrder`
**File:** `tools/canary/internal/acceptance_test.go:136`
**Lines:** 136-208 (72 lines)

### Key Test Logic

```go
// Create fixture with multiple requirements and features in non-alphabetical order
fixtures := map[string]string{
    "z.txt": "CANARY: REQ=CBIN-999; FEATURE=\"Zulu\"; ASPECT=API; STATUS=IMPL; UPDATED=2025-10-15\n",
    "a.txt": "CANARY: REQ=CBIN-101; FEATURE=\"Alpha\"; ASPECT=CLI; STATUS=TESTED; TEST=TestAlpha; UPDATED=2025-10-15\n",
    "m.txt": "CANARY: REQ=CBIN-500; FEATURE=\"Mike\"; ASPECT=Engine; STATUS=STUB; UPDATED=2025-10-15\n",
    "b.txt": "CANARY: REQ=CBIN-102; FEATURE=\"Bravo\"; ASPECT=API; STATUS=BENCHED; TEST=TestBravo; BENCH=BenchBravo; UPDATED=2025-10-15\n",
    "c.txt": "CANARY: REQ=CBIN-102; FEATURE=\"Charlie\"; ASPECT=CLI; STATUS=IMPL; UPDATED=2025-10-15\n", // Same REQ as Bravo
}

// Run scanner twice and compare outputs
csv1, _ := os.ReadFile(csv1Path)
csv2, _ := os.ReadFile(csv2Path)

if string(csv1) != string(csv2) {
    t.Fatalf("CSV output not deterministic")
}

// Verify rows are sorted by REQ ID
var prevReq string
for _, line := range lines[1:] { // Skip header
    req := fields[0]
    if prevReq != "" && req < prevReq {
        t.Errorf("CSV not sorted by REQ: %s after %s", req, prevReq)
    }
    prevReq = req
}
```

## Acceptance Test Summary

| Test | Status | Purpose |
|------|--------|---------|
| TestAcceptance_FixtureSummary | ✅ PASS | Verify JSON summary counts |
| TestAcceptance_Overclaim | ✅ PASS | Verify overclaim detection (exit 2) |
| TestAcceptance_Stale | ✅ PASS | Verify staleness detection (exit 2) |
| TestAcceptance_SelfCanary | ✅ PASS | Verify self-verification with GAP file |
| **TestAcceptance_CSVOrder** | ✅ PASS | **Verify CSV deterministic ordering** |

**Total:** 5/5 acceptance tests PASS

---

**Slice 9 Complete** — Ready for Slice 10 (50k File Performance Benchmark)
