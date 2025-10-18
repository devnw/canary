# Slice 14: Stale Token Auto-Update — COMPLETE

**Date:** 2025-10-16
**Duration:** 2 hours
**Status:** ✅ COMPLETED
**Gap Resolved:** Gap #10 (--update-stale flag)

## Objective
Implement `--update-stale` flag to automatically rewrite UPDATED field for stale TESTED/BENCHED tokens.

## Implementation

### 1. Added --update-stale Flag
- **File:** tools/canary/main.go:75
- Added flag: `flag.Bool("update-stale", false, "rewrite UPDATED field for stale TESTED/BENCHED tokens")`

### 2. Implemented updateStaleTokens Function
- **File:** tools/canary/main.go:290-389
- Parses stale diagnostics to extract REQ IDs
- Walks directory tree respecting skip patterns
- Reads files and searches for CANARY tokens
- Updates UPDATED field for stale TESTED/BENCHED tokens only
- Writes updated content back to files
- Returns map of modified files

### 3. Modified Main Workflow
- **File:** tools/canary/main.go:86-104
- When --update-stale is set:
  1. Run initial scan to detect stale tokens
  2. If stale tokens found, call updateStaleTokens()
  3. Re-scan after updates to get fresh status
  4. Print update summary to stderr

### 4. Added Acceptance Test
- **File:** tools/canary/internal/acceptance_test.go:336-418
- TestAcceptance_UpdateStale validates:
  - CBIN-001 (TESTED, stale) → updated
  - CBIN-002 (TESTED, fresh) → unchanged
  - CBIN-003 (IMPL, stale) → unchanged (only TESTED/BENCHED updated)
  - CBIN-004 (BENCHED, stale) → updated

### 5. Fixed TestAcceptance_SelfCanary
- **File:** tools/canary/internal/acceptance_test.go:123
- Updated skip pattern to exclude `internal/` directory
- Prevents scanning test fixture strings in acceptance_test.go
- Pattern: `(^|/)(.git|.direnv|node_modules|vendor|bin|dist|build|zig-out|.zig-cache|testdata|internal)(/|$)`

## Test Results

### Unit Tests (4/4 passing)
```
TestCANARY_CBIN_101_Engine_ScanBasic        PASS
TestCANARY_CBIN_103_API_StatusSchema        PASS
TestCANARY_CBIN_103_API_JSONDeterminism     PASS
TestCANARY_CBIN_102_CLI_Verify              PASS
```

### Acceptance Tests (7/7 passing)
```
TestAcceptance_FixtureSummary               PASS
TestAcceptance_Overclaim                    PASS
TestAcceptance_Stale                        PASS
TestAcceptance_SelfCanary                   PASS
TestAcceptance_CSVOrder                     PASS
TestAcceptance_SkipEdgeCases                PASS
TestAcceptance_UpdateStale                  PASS
```

## Manual Validation

```bash
$ ./bin/canary --root /tmp/test-stale2 --update-stale --out status.json
Updated 2 stale tokens in 1 files

# Before: CBIN-001 UPDATED=2024-01-01
# After:  CBIN-001 UPDATED=2025-10-16

# Before: CBIN-004 UPDATED=2024-01-01
# After:  CBIN-004 UPDATED=2025-10-16

# CBIN-002 (fresh) remained unchanged
# CBIN-003 (IMPL) remained unchanged
```

## Issues Encountered

### Issue 1: Assignment Mismatch
**Error:** `assignment mismatch: 1 variable but parseKV returns 2 values`
**Fix:** Changed `attrs := parseKV(match[1])` to `attrs, err := parseKV(match[1])` with error handling

### Issue 2: TestAcceptance_UpdateStale Validation
**Error:** Test failed with "CBIN-001 should have UPDATED field changed"
**Fix:** Rewrote validation to parse lines individually using findLine helper function

### Issue 3: TestAcceptance_SelfCanary Stale Tokens
**Error:** `CANARY_STALE REQ=CBIN-001 updated=2024-01-01`
**Root Cause:** Test fixture strings in acceptance_test.go were being scanned
**Fix:** Added `internal` to skip pattern to exclude test files

## Gap Resolution

**Gap #10: --update-stale flag** → ✅ RESOLVED
- Automated staleness remediation implemented
- Selective updates for TESTED/BENCHED only
- Safe file rewriting with error handling
- Full test coverage with acceptance test

## Files Modified

1. tools/canary/main.go (flag + updateStaleTokens + workflow)
2. tools/canary/internal/acceptance_test.go (new test + skip pattern fix)

## Next Steps

- Create PHASE_4_COMPLETE.md summary
- Update GAP_ANALYSIS.md (Gaps #6, #9, #10 resolved)
- Update NEXT.md with Phase 4 completion
- Final verification of all 10 gaps addressed
