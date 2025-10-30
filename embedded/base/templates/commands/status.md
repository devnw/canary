---
description: Show implementation progress summary for a requirement
---

<!-- CANARY: REQ=CBIN-CLI-001; FEATURE="StatusCmd"; ASPECT=Docs; STATUS=IMPL; OWNER=canary; UPDATED=2025-10-16 -->

## User Input

```text
$ARGUMENTS
```

## Outline

Display implementation progress for a requirement with completion percentage and breakdown.

1. **Parse requirement ID**:
   - Extract REQ-ID from arguments (e.g., {{.ReqID}}-SECURITY_REVIEW-102)
   - Validate format

2. **Run canary status command**:
   ```bash
   canary status <REQ-ID>
   ```

   Available flags:
   - `--no-color`: Disable colored output
   - `--db <path>`: Custom database path (default: `.canary/canary.db`)

3. **Display progress**:
   - Show progress bar with completion percentage
   - Total token count
   - Breakdown by status (BENCHED, TESTED, IMPL, STUB)
   - List of incomplete work items
   - Completion status indicator

4. **Calculate metrics**:
   - Completion % = (BENCHED + TESTED) / Total
   - In-Progress % = IMPL / Total
   - Not-Started % = STUB / Total

5. **Provide next steps**:
   - If 100% complete: "All features completed! ✅"
   - If IMPL tokens exist: "Add tests for IMPL tokens"
   - If TESTED tokens exist: "Consider adding benchmarks"
   - If STUB tokens exist: "Use `/canary.plan <REQ-ID>` to plan implementation"

## Example Output

```markdown
## Implementation Status for {{.ReqID}}-API-102

Progress: [========================--------] 75%

**Total:** 20 tokens
**Completed:** 15 (75%)

**In Progress:**
- IMPL: 3 (15%)
- STUB: 2 (10%)

**Status Breakdown:**
- BENCHED: 5 (25%)
- TESTED: 10 (50%)
- IMPL: 3 (15%)
- STUB: 2 (10%)

**Incomplete Work:**

IMPL (needs tests):
- UserList (src/api/users.go:45)
- DataFilter (internal/filter/engine.go:102)
- QueryBuilder (internal/db/queries.go:78)

STUB (not started):
- AdvancedFilters (internal/filter/advanced.go:23)
- CacheLayer (internal/cache/layer.go:15)

**Next Steps:**
1. Add tests for 3 IMPL features
2. Plan implementation for 2 STUB features using `/canary.plan {{.ReqID}}-API-102`
3. Consider adding benchmarks for performance-critical features
```

## Guidelines

- **Visual Progress**: Use progress bar for quick assessment
- **Actionable**: List specific files and line numbers for incomplete work
- **Priority**: Focus on STUB and IMPL items first
- **Clear Metrics**: Show both counts and percentages
- **Next Steps**: Provide specific commands and actions
- **Success Recognition**: Celebrate 100% completion
- **Database Required**: Suggest `canary index` if database missing
