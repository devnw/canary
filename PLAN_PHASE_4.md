# Phase 4 Plan: Remaining Gaps & Polish

**Date:** 2025-10-15
**Status:** PLANNING

## Current State Summary

### ✅ Completed (Phases 1-3 + Slices 7-10)
- **Phase 1:** 3 TestCANARY_* functions (all PASS)
- **Phase 2:** 4 BenchmarkCANARY_* functions (all RUN)
- **Phase 3:** Documentation synchronized (GAP_ANALYSIS.md, CHECKLIST.md, NEXT.md)
- **Slice 7:** CRUSH.md placeholder fixed (CBIN-101 parseable)
- **Slice 8:** CI workflow created (.github/workflows/canary.yml, 5 jobs)
- **Slice 9:** CSV row order test (TestAcceptance_CSVOrder)
- **Slice 10:** 50k file benchmark (1.85s, 81.5% under target)

### ✅ Test Status
- **Unit Tests:** 3/3 PASS
- **Acceptance Tests:** 5/5 PASS
- **Benchmarks:** 4/4 RUN
- **CI Jobs:** 5/5 validated locally

### 📊 Gap Resolution Progress
- **Resolved:** 7/10 gaps (#1, #2, #4, #5, #7, #8, plus docs)
- **Remaining:** 3 gaps (#3, #6, #9, #10)

## Remaining Gaps Analysis

### Gap #3: cmd/canary Build Failure
**Priority:** LOW (tools/canary works, cmd/canary is legacy)
**Status:** ❌ OPEN
**Issue:** `cmd/canary/init.go` references non-existent packages
**Effort:** 2-4 hours (requires refactoring or removal)
**Decision:** DEFER (not blocking production use)

### Gap #6: Regex Portability (--skip)
**Priority:** MEDIUM
**Status:** ❌ OPEN
**Issue:** Edge cases not tested (symlinks, Unicode paths, nested dotfiles)
**Effort:** 1-2 hours
**Value:** Improves robustness for diverse environments

### Gap #9: JSON Determinism
**Priority:** HIGH
**Status:** ❌ OPEN
**Issue:** No explicit byte-exact comparison across runs
**Effort:** 30-45 minutes
**Value:** Ensures reproducible builds, stable diffs in version control

### Gap #10: Stale Token Remediation UX
**Priority:** MEDIUM
**Status:** ❌ OPEN
**Issue:** No automation for updating UPDATED field
**Effort:** 2-3 hours
**Value:** Improves developer experience, reduces manual toil

## Phase 4 Plan: Slices 11-14

### Slice 11: JSON Determinism Test ⭐ START HERE
**Priority:** HIGH (quick win, high value)
**Estimated Time:** 30-45 minutes

**Objective:** Validate that JSON output is byte-for-byte identical across multiple runs with same inputs.

**Implementation:**
- **File:** `tools/canary/status_test.go`
- **Test:** `TestCANARY_CBIN_103_API_JSONDeterminism`
- **Logic:**
  1. Create fixture directory with 20 CANARY tokens
  2. Run scanner 5 times
  3. Compare JSON outputs byte-for-byte (use SHA256 hash)
  4. Assert all hashes identical
  5. Verify sorted key order (by_status, by_aspect, etc.)

**Acceptance Criteria:**
- Test PASS: All 5 JSON outputs produce identical SHA256 hash
- Test validates sorted keys in JSON structure
- Gap #9 marked RESOLVED in GAP_ANALYSIS.md
- CHECKLIST.md updated

**Risks:** Low (JSON marshaling already uses custom marshalers)

---

### Slice 12: Update NEXT.md with Slices 7-10 Complete
**Priority:** HIGH (documentation hygiene)
**Estimated Time:** 15 minutes

**Objective:** Update NEXT.md to reflect completion of Slices 7-10 and add new Slices 11-14.

**Changes:**
1. Move Slices 7-10 from "Up Next" to "Completed" section
2. Add completion dates, validation results, evidence links
3. Add new "Up Next (Slices 11-14)" section
4. Update success metrics and prioritization rationale

**Acceptance Criteria:**
- NEXT.md accurately reflects current state
- New slices clearly documented
- Estimated times and dependencies listed

---

### Slice 13: Regex Portability Tests
**Priority:** MEDIUM
**Estimated Time:** 1-2 hours

**Objective:** Validate --skip regex handles edge cases correctly.

**Implementation:**
- **File:** `tools/canary/internal/acceptance_test.go`
- **Test:** `TestAcceptance_SkipEdgeCases`
- **Fixtures:**
  1. Symlink to directory (should skip if matches pattern)
  2. Unicode filename (e.g., `测试.go`)
  3. Nested dotfiles (`.config/.env`)
  4. Mixed separators (Windows-style paths on Linux)
  5. Very long paths (>256 chars)

**Logic:**
- Create fixture with edge case files
- Run scanner with various --skip patterns
- Verify expected files scanned vs. skipped
- Assert no panics or errors

**Acceptance Criteria:**
- Test PASS: All edge cases handled correctly
- Gap #6 marked RESOLVED
- Documentation updated

---

### Slice 14: Stale Token Auto-Update (Optional)
**Priority:** LOW-MEDIUM
**Estimated Time:** 2-3 hours

**Objective:** Add `--update-stale` flag to automatically rewrite UPDATED field.

**Implementation:**
- **File:** `tools/canary/update.go` (NEW)
- **CLI Flag:** `--update-stale`
- **Logic:**
  1. Scan for TESTED/BENCHED tokens with UPDATED > 30 days
  2. Parse source file line-by-line
  3. Rewrite UPDATED=YYYY-MM-DD to current date
  4. Preserve exact formatting (spacing, comment style)
  5. Write modified file back to disk

**Acceptance Criteria:**
- Flag implemented and documented
- Test validates UPDATED field rewritten correctly
- Verify gate passes after update
- Gap #10 marked RESOLVED

**Risks:** MEDIUM (file rewriting can be error-prone)

---

## Recommended Sequencing

### Phase 4A: Quick Wins (1 hour)
1. **Slice 11** — JSON Determinism Test (30-45 min)
2. **Slice 12** — Update NEXT.md (15 min)

### Phase 4B: Optional Enhancements (2-3 hours)
3. **Slice 13** — Regex Portability Tests (1-2 hours)
4. **Slice 14** — Stale Token Auto-Update (2-3 hours) — DEFER if time-constrained

## Success Metrics

**After Slice 11-12 (Minimum Viable):**
- ✅ 8/10 gaps resolved (or 9/10 if Slice 13 done)
- ✅ All core functionality tested and validated
- ✅ CI enforces quality on every PR
- ✅ JSON/CSV outputs deterministic and reproducible
- ✅ Performance validated at scale (50k files < 2s)

**After All Slices (Complete):**
- ✅ 9/10 gaps resolved (only cmd/canary refactor remains)
- ✅ Edge cases handled robustly
- ✅ Developer UX improved (auto-update stale tokens)
- ✅ Production-ready for any environment

## Decision: Start with Slice 11

**Rationale:**
- Quick win (30-45 min)
- High value (reproducible builds)
- Low risk (builds on existing code)
- Closes gap #9 immediately

**Next:** After Slice 11, reassess and decide whether to continue with Slice 12-14 or stop.
