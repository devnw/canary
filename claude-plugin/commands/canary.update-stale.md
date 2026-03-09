---
description: Automatically update UPDATED field for stale CANARY tokens
---


## User Input

```text
$ARGUMENTS
```

## Outline

Automatically update stale CANARY tokens (TESTED/BENCHED with UPDATED > 30 days old).

1. **Read project configuration**:
   - Load `.canary/project.yaml` to determine the project key

2. **Run staleness scan**:
   ```bash
   canary scan --root . --strict
   ```

   This identifies all TESTED and BENCHED tokens older than 30 days.

3. **Parse stale tokens**:
   - Extract requirement IDs from scanner output
   - List affected files and line numbers
   - Show age of each stale token

4. **Confirm update** (if arguments don't include --force):
   ```markdown
   ## Stale Tokens Found: N

   The following CANARY tokens need updating:

   - <PROJECT_KEY>-<NNN>: FeatureName (N days old)
     - File: src/api/auth.go:10
     - Current: UPDATED=YYYY-MM-DD
     - New: UPDATED=YYYY-MM-DD

   Run `canary scan --update-stale` to update these tokens automatically.

   Proceed with update? (y/n)
   ```

5. **Execute update**:
   ```bash
   canary scan --root . --update-stale
   ```

   This automatically:
   - Finds all stale TESTED/BENCHED tokens
   - Updates UPDATED field to current date
   - Preserves all other token fields
   - Updates files in-place

6. **Verify updates**:
   - Re-run scan to confirm no more stale tokens
   - Show before/after comparison
   - List all updated files

7. **Generate update report**:
   ```markdown
   ## Stale Token Update Results

   **Update Date:** YYYY-MM-DD
   **Tokens Updated:** N

   ### Updated Tokens
   - <PROJECT_KEY>-<NNN>: src/api/auth.go (old-date -> new-date)
   - <PROJECT_KEY>-<NNN>: internal/cache/cache.go (old-date -> new-date)

   ### Files Modified
   - src/api/auth.go
   - internal/cache/cache.go

   ### Verification
   No stale tokens remaining. All TESTED/BENCHED requirements are current.

   **Next Steps:**
   - Commit updated tokens
   - Run `canary scan --verify GAP_ANALYSIS.md --strict` to verify claims
   ```

8. **Suggest git commit**:
   ```bash
   git add <modified files>
   git commit -m "chore: update stale CANARY tokens

   Updated UPDATED field for N stale tokens.
   All TESTED/BENCHED requirements now current (within 30 days).
   "
   ```

## Example Output

```markdown
## Stale Token Update Results

**Update Date:** YYYY-MM-DD
**Tokens Updated:** 2

### Before Update
```go
// CANARY: REQ=<PROJECT_KEY>-<NNN>; FEATURE="UserAuth"; ASPECT=API; STATUS=TESTED; TEST=TestUserAuth; UPDATED=2024-01-01
```

### After Update
```go
// CANARY: REQ=<PROJECT_KEY>-<NNN>; FEATURE="UserAuth"; ASPECT=API; STATUS=TESTED; TEST=TestUserAuth; UPDATED=YYYY-MM-DD
```

### Files Modified
- src/api/auth.go (1 token updated)
- internal/cache/cache.go (1 token updated)

### Verification
No stale tokens remaining.
All TESTED/BENCHED requirements current.
Ready to commit changes.

**Git Command:**
```bash
git add src/api/auth.go internal/cache/cache.go
git commit -m "chore: update stale CANARY tokens"
```
```

## Guidelines

- **Safety First**: Only update TESTED/BENCHED tokens (not STUB/IMPL)
- **Preserve Fields**: Keep all token fields except UPDATED
- **Batch Updates**: Update all stale tokens in one operation
- **Verification**: Re-scan after update to confirm success
- **Git Integration**: Provide ready-to-use commit command
- **Transparency**: Show exactly what changed (before/after)
