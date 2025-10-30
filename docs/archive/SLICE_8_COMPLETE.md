# Slice 8 Complete: Add CI Workflow ✅

**Date:** 2025-10-15
**Duration:** ~20 minutes
**Status:** COMPLETED

## Summary

Created comprehensive GitHub Actions CI workflow for canary validation with 5 jobs: build, unit tests, acceptance tests, benchmarks, and self-verification gate. All jobs validated locally and ready for CI execution.

## Problem

No automated CI validation existed for canary implementation. Manual testing was required to validate changes, increasing risk of regressions and missing test coverage.

## Solution

Created `.github/workflows/canary.yml` with 5 parallel/sequential jobs matching the Slice 8 specification:

### Job 1: Build Canary Binary
**Command:** `go build -o ./bin/canary ./tools/canary`
**Artifacts:** Uploads canary binary for use by verify-self job
**Validation:** ✅ EXIT=0

### Job 2: Run Unit Tests
**Command:** `go test ./tools/canary -v -run TestCANARY`
**Coverage:** 3 TestCANARY_* functions (CBIN-101, CBIN-102, CBIN-103)
**Validation:** ✅ 3/3 PASS

### Job 3: Run Acceptance Tests
**Command:** `go test ./tools/canary/internal -v -run TestAcceptance`
**Coverage:** 4 acceptance tests (FixtureSummary, Overclaim, Stale, SelfCanary)
**Validation:** ✅ 4/4 PASS

### Job 4: Run Benchmarks
**Command:** `go test ./tools/canary -bench BenchmarkCANARY -run ^$ -benchmem`
**Coverage:** 3 BenchmarkCANARY_* functions (Engine_Scan, CLI_Verify, API_Emit)
**Validation:** ✅ 3/3 RUN

### Job 5: Self-Verification Gate
**Command:** `./bin/canary --root tools/canary --verify GAP_SELF.md --strict --skip '(^|/)(.git|.direnv|node_modules|vendor|bin|dist|build|zig-out|.zig-cache|.crush)(/|$)'`
**Dependencies:** Requires job 1 (build) to complete first
**Artifacts:** Downloads canary binary from job 1
**Validation:** ✅ EXIT=0

## File Changes

### .github/workflows/canary.yml
**Status:** UPDATED (existing file replaced with new implementation)

**Before:**
- Single "scan" job with incorrect paths (`./main.go` instead of `./tools/canary`)
- Single "test" job with wrong test paths (`./internal/acceptance/...`)
- Go 1.23 (not 1.25.0)
- No benchmarks
- No proper verify gate

**After:**
- 5 separate jobs with clear responsibilities
- Correct build path: `./tools/canary`
- Correct test paths: `./tools/canary`, `./tools/canary/internal`
- Go 1.25.0 as required by CRUSH.md
- Dedicated benchmark job
- Proper verify gate with GAP_SELF.md and skip pattern

## Key Fixes

### Fix 1: Skip Pattern for .crush Directory
**Issue:** Verify gate was failing with `CANARY_PARSE_ERROR err=".crush/crush.db: invalid ASPECT SECURITY_REVIEW"`
**Root Cause:** `.crush/` directory contains database file with invalid token placeholders
**Solution:** Added `.crush` to skip pattern regex
**Before:** `'(^|/)(.git|.direnv|node_modules|vendor|bin|dist|build|zig-out|.zig-cache)(/|$)'`
**After:** `'(^|/)(.git|.direnv|node_modules|vendor|bin|dist|build|zig-out|.zig-cache|.crush)(/|$)'`

### Fix 2: Verify Syntax
**Issue:** Initial workflow used incorrect `verify` subcommand syntax
**Root Cause:** Misunderstanding of canary CLI flags vs subcommands
**Solution:** Changed to `--verify` flag syntax
**Before:** `./bin/canary verify --root tools/canary --gap GAP_ANALYSIS.md --strict`
**After:** `./bin/canary --root tools/canary --verify GAP_SELF.md --strict --skip '...'`

### Fix 3: GAP File Reference
**Issue:** Workflow referenced wrong GAP file (GAP_ANALYSIS.md)
**Root Cause:** Slice 8 spec requires GAP_SELF.md for self-verification
**Solution:** Changed to GAP_SELF.md
**GAP_SELF.md Contents:**
```markdown
# GAP
✅ CBIN-101
✅ CBIN-102
```

## Validation Results

All 5 workflow jobs validated locally:

### Local Validation Summary
```bash
=== Build ===
✅ Build: PASS

=== Unit Tests ===
=== RUN   TestCANARY_CBIN_101_Engine_ScanBasic
--- PASS: TestCANARY_CBIN_101_Engine_ScanBasic (0.00s)
=== RUN   TestCANARY_CBIN_103_API_StatusSchema
--- PASS: TestCANARY_CBIN_103_API_StatusSchema (0.00s)
=== RUN   TestCANARY_CBIN_102_CLI_Verify
--- PASS: TestCANARY_CBIN_102_CLI_Verify (0.00s)
PASS

=== Acceptance ===
=== RUN   TestAcceptance_FixtureSummary
--- PASS: TestAcceptance_FixtureSummary (0.08s)
=== RUN   TestAcceptance_Overclaim
--- PASS: TestAcceptance_Overclaim (0.07s)
=== RUN   TestAcceptance_Stale
--- PASS: TestAcceptance_Stale (0.07s)
=== RUN   TestAcceptance_SelfCanary
--- PASS: TestAcceptance_SelfCanary (0.08s)
PASS

=== Benchmarks ===
BenchmarkCANARY_CBIN_101_Engine_Scan-32      342    3364273 ns/op    1110184 B/op    11353 allocs/op
BenchmarkCANARY_CBIN_103_API_Emit-32        1393     883177 ns/op      36067 B/op     2119 allocs/op
BenchmarkCANARY_CBIN_102_CLI_Verify-32     34852      35059 ns/op       5180 B/op        13 allocs/op
PASS

=== Verify ===
✅ Verify: PASS (EXIT=0)
```

**Total:** 7 tests PASS, 3 benchmarks RUN, verify gate PASS

## Workflow Configuration

### Triggers
- `push` to `main` branch
- `pull_request` to `main` branch

### Permissions
- `contents: read` (minimal required permissions)

### Job Dependencies
- Jobs 1-4 run in parallel (independent)
- Job 5 (verify-self) runs after job 1 (build) completes (uses binary artifact)

### Go Version
- `1.25.0` (as required by CRUSH.md and NEXT.md)

### Actions Used
- `actions/checkout@v4` - Check out repository code
- `actions/setup-go@v5` - Install Go 1.25.0
- `actions/upload-artifact@v4` - Share binary between jobs
- `actions/download-artifact@v4` - Retrieve binary in verify job

## Success Criteria: ✅ ALL MET

- [x] `.github/workflows/canary.yml` created with 5 jobs
- [x] Job 1: Build canary binary from `./tools/canary`
- [x] Job 2: Run TestCANARY_* unit tests
- [x] Job 3: Run acceptance tests
- [x] Job 4: Run BenchmarkCANARY_* benchmarks
- [x] Job 5: Run verify gate with GAP_SELF.md
- [x] All jobs validated locally (PASS/EXIT=0)
- [x] Workflow uses Go 1.25.0
- [x] Workflow triggers on push/PR to main
- [x] Skip pattern includes `.crush` to avoid parse errors
- [x] Verify gate uses correct syntax and GAP file

**Slice 8 Status: COMPLETE** 🎉

## Impact

**Before Slice 8:**
- Manual testing required for every change
- No automated regression detection
- Risk of breaking changes in production
- Old workflow had incorrect paths and versions

**After Slice 8:**
- Automated validation on every PR/push
- 5 comprehensive test gates
- Regression detection via benchmarks
- Self-verification ensures claimed features have evidence
- Correct Go version (1.25.0) enforced

## Files Modified

- **`.github/workflows/canary.yml`** — Complete rewrite with 5 jobs (105 lines)

**Total:** 1 file modified

## Next Steps

**Slice 9:** CSV Row Order Test (1 hour)
- Create `TestAcceptance_CSVOrder` in `tools/canary/internal/acceptance_test.go`
- Validate CSV output has deterministic row ordering
- Test with fixtures containing multiple requirements and features
- Ensure consistent output across multiple runs

**Slice 10:** Large-Scale 50k File Performance Benchmark (2-3 hours)
- Create `BenchmarkCANARY_CBIN_101_Engine_Scan50k` with 50,000 file fixture
- Validate <10s target from CRUSH.md
- Establish baseline for regression detection
- Update Perf50k column in CHECKLIST.md from ◐ PARTIAL to ✅ DONE

**Estimated Time Remaining:** 3-4 hours for Slices 9-10

## Appendix: Workflow YAML

Full workflow available at: `.github/workflows/canary.yml`

Key sections:
```yaml
name: Canary CI

on:
  push:
    branches: [main]
  pull_request:
    branches: [main]

jobs:
  build:
    name: Build Canary Binary
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: '1.25.0'
      - run: go build -o ./bin/canary ./tools/canary
      - uses: actions/upload-artifact@v4
        with:
          name: canary-binary
          path: ./bin/canary

  verify-self:
    name: Self-Verification Gate
    runs-on: ubuntu-latest
    needs: build
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: '1.25.0'
      - uses: actions/download-artifact@v4
        with:
          name: canary-binary
          path: ./bin
      - run: chmod +x ./bin/canary
      - run: |
          ./bin/canary --root tools/canary --verify GAP_SELF.md --strict --skip '(^|/)(.git|.direnv|node_modules|vendor|bin|dist|build|zig-out|.zig-cache|.crush)(/|$)'
```

---

**Slice 8 Complete** — Ready for Slice 9 (CSV Row Order Test)
