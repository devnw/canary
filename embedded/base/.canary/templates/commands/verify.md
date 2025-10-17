---
description: Verify GAP_ANALYSIS.md claims against actual CANARY token status
---

## User Input

```text
$ARGUMENTS
```

## Outline

Verify that claims in GAP_ANALYSIS.md match actual requirement implementation status.

1. **Locate GAP_ANALYSIS.md**:
   - Default: `./GAP_ANALYSIS.md`
   - If arguments provided: Parse for custom path
   - Error if file not found

2. **Parse claimed requirements**:
   - Find lines starting with `✅` followed by requirement ID
   - Extract all claimed CBIN-XXX identifiers
   - Example: `✅ CBIN-001 - UserAuth fully implemented`

3. **Run verification scan**:
   ```bash
   canary scan --root . --verify GAP_ANALYSIS.md --strict
   ```

   This checks:
   - All claimed requirements exist in codebase
   - Claimed requirements have STATUS=TESTED or STATUS=BENCHED
   - No overclaims (claiming STUB or IMPL as complete)
   - No stale tokens (TESTED/BENCHED > 30 days old)

4. **Analyze verification results**:
   - Exit code 0: All claims verified ✅
   - Exit code 2: Verification failed (overclaim or staleness) ❌
   - Parse stderr for specific failures

5. **Generate verification report**:
   ```markdown
   ## GAP Analysis Verification Results

   **Verification Date:** YYYY-MM-DD
   **Status:** PASS / FAIL

   ### Claimed Requirements: N
   - ✅ CBIN-001: UserAuth (BENCHED, verified)
   - ✅ CBIN-002: DataValidation (TESTED, verified)
   - ❌ CBIN-003: ReportGen (IMPL only, overclaim)
   - ⚠️ CBIN-004: Cache (TESTED but stale, 288 days old)

   ### Verification Summary
   - Valid Claims: X
   - Overclaims: Y
   - Stale Claims: Z

   ### Action Required
   [If failures exist, list specific actions needed]
   ```

6. **Provide remediation steps** (if verification failed):
   - For overclaims: "Add TEST= field and create test function for CBIN-XXX"
   - For stale tokens: "Run `canary scan --update-stale` to refresh UPDATED fields"
   - For missing tokens: "Remove ✅ CBIN-XXX from GAP_ANALYSIS.md (not found in code)"

## Verification Rules

**Valid Claim**: Requirement with STATUS=TESTED or STATUS=BENCHED

**Overclaim**: Requirement claimed with ✅ but:
- STATUS=STUB (not implemented)
- STATUS=IMPL (implemented but not tested)
- STATUS=MISSING (placeholder only)

**Stale Claim**: Valid claim but UPDATED > 30 days ago (with `--strict`)

## Example Flow

**GAP_ANALYSIS.md contents:**
```markdown
# Requirements Gap Analysis

## Verified Requirements
✅ CBIN-001 - User authentication fully tested
✅ CBIN-002 - Data validation with comprehensive tests
✅ CBIN-003 - Report generation functional
```

**Verification scan finds:**
```
✅ CBIN-001: STATUS=BENCHED, UPDATED=2025-10-15 → Valid
✅ CBIN-002: STATUS=TESTED, UPDATED=2025-10-14 → Valid
❌ CBIN-003: STATUS=IMPL, missing TEST= field → Overclaim!
```

**Report:**
```markdown
## GAP Analysis Verification Results

**Status:** ❌ FAILED

### Verification Details
- ✅ CBIN-001: UserAuth (BENCHED, verified)
- ✅ CBIN-002: DataValidation (TESTED, verified)
- ❌ CBIN-003: ReportGen (IMPL only, overclaim detected)

### Action Required
1. Add tests for CBIN-003:
   - Create test function: `TestReportGeneration`
   - Add to token: `TEST=TestReportGeneration`
   - Ensure test passes
   - Re-run verification

OR

2. Update GAP_ANALYSIS.md:
   - Remove `✅` marker from CBIN-003
   - Move to "In Progress" section
```

## Guidelines

- **Strict Enforcement**: Use `--strict` flag to catch stale claims
- **Clear Failures**: Explicitly show which claims failed and why
- **Actionable Steps**: Provide exact commands to fix issues
- **CI Integration**: Verification can run in CI to prevent merging overclaims
- **Exit Codes**: Respect scanner exit codes (0=pass, 2=fail)
