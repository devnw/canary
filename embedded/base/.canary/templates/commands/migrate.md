---
description: Migrate orphaned CANARY tokens to formal specifications
---


## User Input

```text
$ARGUMENTS
```

## Outline

Detect and migrate orphaned requirements (CANARY tokens without specifications) to formal spec and plan files.

1. **Determine operation mode**:
   - Default: Detect and report orphaned requirements
   - If REQ-ID provided: Migrate specific requirement
   - If `--all` in arguments: Migrate all orphaned requirements
   - If `--dry-run` in arguments: Preview without creating files

2. **Detect orphaned requirements**:
   ```bash
   canary orphan detect --show-features
   ```

   This identifies requirements that have:
   - CANARY tokens in the codebase
   - No corresponding `.canary/specs/CBIN-XXX-*/spec.md` file
   - Tokens NOT in documentation directories (automatically filtered)

3. **Review orphan details**:
   For each orphaned requirement, show:
   - Requirement ID (e.g., CBIN-105)
   - Confidence level (HIGH/MEDIUM/LOW)
   - Feature count and aspects
   - File locations

4. **Execute migration** (if user confirms):
   ```bash
   # Single requirement
   canary orphan run CBIN-105

   # All orphaned requirements
   canary orphan run --all

   # Preview first
   canary orphan run --all --dry-run
   ```

5. **Generate migration artifacts**:
   For each requirement, creates:
   - `.canary/specs/CBIN-XXX-featurename/spec.md` - Auto-generated specification
   - `.canary/specs/CBIN-XXX-featurename/plan.md` - Implementation plan reflecting current state

6. **Post-migration tasks**:
   - Review generated specifications for accuracy
   - Update spec.md with detailed requirements (replace auto-generated content)
   - Update plan.md with implementation approach
   - Run `canary scan` to reindex database

## Example Output

```markdown
## Orphaned Requirements Detection

Found **3 orphaned requirements** with tokens but no specifications:

### 🟢 CBIN-200 (Confidence: HIGH)
- **Features:** 5
- **Aspects:** API (2), Engine (2), Storage (1)
- **Status:** 3 TESTED, 2 IMPL
- **Tests:** TestCANARY_CBIN_200_API_Handler, TestCANARY_CBIN_200_Engine_Core
- **Files:** pkg/api/handler.go, pkg/engine/core.go, pkg/storage/db.go

### 🟡 CBIN-301 (Confidence: MEDIUM)
- **Features:** 3
- **Aspects:** CLI (2), Docs (1)
- **Status:** 2 IMPL, 1 STUB
- **Files:** cmd/cli/commands.go, docs/README.md

### 🔴 CBIN-999 (Confidence: LOW)
- **Features:** 1
- **Aspects:** API (1)
- **Status:** 1 STUB
- **Files:** test.go
- ⚠️ **Warning:** Low confidence - may need manual review

---

## Migration Preview (Dry Run)

Would create the following:

1. **CBIN-200-handler** ✅
   - `.canary/specs/CBIN-200-handler/spec.md` (Confidence: HIGH)
   - `.canary/specs/CBIN-200-handler/plan.md`

2. **CBIN-301-clicommands** ⚠️
   - `.canary/specs/CBIN-301-clicommands/spec.md` (Confidence: MEDIUM)
   - `.canary/specs/CBIN-301-clicommands/plan.md`

3. **CBIN-999-singlefeature** 🔴
   - `.canary/specs/CBIN-999-singlefeature/spec.md` (Confidence: LOW - NEEDS REVIEW)
   - `.canary/specs/CBIN-999-singlefeature/plan.md`

Proceed with migration? [yes/no]
```

## Migration Results

After migration:

```markdown
## Migration Complete

✅ Successfully migrated **3 requirements**

### Created Specifications:
1. **CBIN-200** → `.canary/specs/CBIN-200-handler/`
   - spec.md (HIGH confidence)
   - plan.md

2. **CBIN-301** → `.canary/specs/CBIN-301-clicommands/`
   - spec.md (MEDIUM confidence - review recommended)
   - plan.md

3. **CBIN-999** → `.canary/specs/CBIN-999-singlefeature/`
   - spec.md (LOW confidence - **manual review required**)
   - plan.md

### Next Steps:

1. **Review Generated Specifications**:
   ```bash
   # Review high-priority specs
   cat .canary/specs/CBIN-200-handler/spec.md
   ```

2. **Update Specifications** (replace auto-generated content):
   - Add detailed Overview and Purpose
   - Define comprehensive User Stories
   - Specify Functional Requirements
   - Add success criteria and test scenarios

3. **Reindex Database**:
   ```bash
   canary scan --root . --out status.json
   ```

4. **Verify Migration**:
   ```bash
   canary orphan detect  # Should show 0 orphaned requirements
   ```

### ⚠️ Low Confidence Migrations

The following specs need manual review due to low confidence:
- **CBIN-999**: Only 1 feature, no tests - verify this is a complete requirement

### 📊 Migration Summary

- Total orphaned: 3
- Successfully migrated: 3
- High confidence: 1 (33%)
- Medium confidence: 1 (33%)
- Low confidence (needs review): 1 (33%)
```

## Guidelines

- **Automatic Detection**: Run `canary orphan detect` without prompting
- **Show Confidence Levels**: Use colors/emojis to indicate quality
  - 🟢 HIGH: 5+ features OR tests + benchmarks
  - 🟡 MEDIUM: 3-4 features OR has tests
  - 🔴 LOW: 1-2 features, no tests
- **Ask Before Migration**: Confirm with user before creating files
- **Dry Run First**: Suggest `--dry-run` for preview
- **Post-Migration Guidance**: Provide specific next steps
- **Quality Warnings**: Flag low-confidence migrations for manual review

## Common Use Cases

### Use Case 1: Legacy Codebase Migration
**Scenario**: Migrating existing project with CANARY tokens but no formal specs

```bash
canary orphan detect                    # Identify all orphans
canary orphan run --all --dry-run       # Preview migration
canary orphan run --all                 # Execute migration
```

### Use Case 2: Single Requirement Cleanup
**Scenario**: Found one orphaned requirement (e.g., CBIN-105)

```bash
canary orphan detect --show-features    # Review details
canary orphan run CBIN-105              # Migrate just this one
```

### Use Case 3: Documentation Examples
**Scenario**: Tokens in `/docs/` directories (automatically filtered)

These are excluded automatically - orphan detection skips:
- `/docs/` - Documentation examples
- `/.claude/` - AI agent configurations
- `/.cursor/` - IDE configurations
- `/.canary/specs/` - Existing specifications

## Error Handling

If migration fails:
1. Show clear error message
2. Suggest checking file permissions
3. Verify `.canary/specs/` directory exists
4. Check database connectivity

If low confidence detected:
1. Warn user during detection phase
2. Mark in generated spec with migration notice
3. Recommend manual review in summary
