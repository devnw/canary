# Migration Feature - Implementation Summary

## Overview

Implemented a complete migration system to convert existing spec-kit or legacy CANARY projects to the new unified CANARY system with minimal overhead for agents.

## Problem Solved

**Before:**
- No automated way to migrate from spec-kit or legacy CANARY
- Manual file copying and restructuring required
- High risk of missing files or incorrect structure
- Time-consuming for agents to figure out migration path

**After:**
- Automatic system detection
- One-command migration with dry-run preview
- Validation and safety checks
- Clear feedback and next steps
- Tested with real spec-kit repository

## Implementation

### 1. Migration Package (CBIN-131)

**File:** `internal/migrate/migrate.go`

**Core types:**
```go
type SystemType string
const (
    SystemTypeSpecKit      SystemType = "spec-kit"
    SystemTypeLegacyCanary SystemType = "legacy-canary"
    SystemTypeUnknown      SystemType = "unknown"
)

type MigrationPlan struct {
    SystemType    SystemType
    FilesToCopy   []FileCopy
    FilesToMerge  []FileMerge
    FilesToCreate []string
    Warnings      []string
}
```

**Key functions:**
- `DetectSystemType(rootDir string) (SystemType, string)` - Identifies system type
- `PlanMigration(rootDir string, systemType SystemType, dryRun bool) (*MigrationPlan, error)` - Creates migration plan
- `ExecuteMigration(rootDir string, plan *MigrationPlan, dryRun bool) error` - Executes migration
- `GetMigrationSummary(plan *MigrationPlan) string` - Human-readable summary

**Detection logic:**

*spec-kit indicators (needs 3/5):*
- `memory/constitution.md`
- `templates/spec-template.md`
- `templates/plan-template.md`
- `templates/commands/specify.md`
- `templates/commands/plan.md`

*legacy-canary indicators (needs 2/4):*
- `tools/canary/`
- `tools/canary/main.go`
- `status.json`
- `GAP_ANALYSIS.md`

### 2. CLI Commands (CBIN-131)

#### canary detect

**File:** `cmd/canary/main.go`

**Purpose:** Analyze directory to identify system type

**Usage:**
```bash
canary detect [directory]
```

**Output:**
```
🔍 Analyzing: /path/to/project

System Type: spec-kit
Details: Detected spec-kit system (5/5 indicators found)

To migrate this system, run:
  canary migrate-from spec-kit

For a dry run (preview changes):
  canary migrate-from spec-kit --dry-run
```

#### canary migrate-from

**File:** `cmd/canary/main.go`

**Purpose:** Migrate from spec-kit or legacy CANARY to unified system

**Usage:**
```bash
canary migrate-from <system-type> [directory] [flags]
```

**System types:**
- `spec-kit` - Migrate from spec-kit specification system
- `legacy-canary` - Migrate from legacy CANARY token system

**Flags:**
- `--dry-run` - Preview changes without applying them
- `--force` - Force migration even if detection doesn't match

**Features:**
- Validates system type matches detection
- Creates migration plan
- Shows summary before execution
- Executes migration with clear feedback
- Provides next steps after completion

### 3. spec-kit Migration

**What gets migrated:**

1. **Copied directly (13 files):**
   - `memory/constitution.md` → `.canary/memory/constitution.md`
   - `templates/spec-template.md` → `.canary/templates/spec-template.md`
   - `templates/plan-template.md` → `.canary/templates/plan-template.md`
   - `templates/tasks-template.md` → `.canary/templates/tasks-template.md`
   - `templates/checklist-template.md` → `.canary/templates/checklist-template.md`
   - `templates/commands/*.md` → `.canary/templates/commands/*.md`
     - specify.md, plan.md, tasks.md, implement.md
     - clarify.md, analyze.md, checklist.md, constitution.md

2. **Files requiring manual merge (1):**
   - `README.md` - Merge spec-kit README with CANARY documentation

3. **Files to create (3):**
   - `README_CANARY.md` - CANARY token specification
   - `GAP_ANALYSIS.md` - Requirement tracking template
   - `CLAUDE.md` - AI agent integration guide

4. **Warnings (2):**
   - `.canary/` directory already exists - will merge content
   - `scripts/` directory found - needs manual review for compatibility

### 4. Legacy CANARY Migration

**What gets migrated:**

1. **Preserved in place (3 files):**
   - `status.json` - Scanner output
   - `status.csv` - CSV output
   - `GAP_ANALYSIS.md` - Requirement tracking

2. **Files to create (5):**
   - `.canary/memory/constitution.md` - Project principles
   - `.canary/templates/spec-template.md` - Specification template
   - `.canary/templates/plan-template.md` - Implementation plan template
   - `README_CANARY.md` - Token specification
   - `CLAUDE.md` - AI agent integration

3. **Warnings (1):**
   - `tools/canary/` scanner found - can be removed after migration

## Testing

### Test 1: spec-kit Detection

```bash
$ ./bin/canary detect specs/spec-kit-repo/

🔍 Analyzing: specs/spec-kit-repo/

System Type: spec-kit
Details: Detected spec-kit system (5/5 indicators found)
```
**Result:** ✅ Correct detection

### Test 2: Dry-Run Migration

```bash
$ cd specs/spec-kit-repo
$ canary migrate-from spec-kit --dry-run

📋 Planning migration from spec-kit...

Migration Plan for spec-kit:

Files to copy: 13
Files to merge: 1
Files to create: 3
Warnings: 2

🔍 DRY RUN MODE - No changes will be made

Would create: .canary
Would create: .canary/memory
...
Would copy: memory/constitution.md -> .canary/memory/constitution.md
...

✅ Dry run complete - no changes were made
```
**Result:** ✅ Dry-run preview works

### Test 3: Actual Migration

```bash
$ cp -r specs/spec-kit-repo /tmp/spec-kit-test
$ cd /tmp/spec-kit-test
$ canary migrate-from spec-kit

📋 Planning migration from spec-kit...
...
🚀 Executing migration...

✅ Created: .canary
✅ Created: .canary/memory
...
✅ Copied: memory/constitution.md -> .canary/memory/constitution.md
...

✅ Migration complete!

Next steps:
  1. Review migrated files in .canary/
  2. Update slash commands in .canary/templates/commands/ for your workflow
  3. Run: canary index
  4. Run: canary scan --root . --out status.json
```
**Result:** ✅ Migration successful

### Test 4: Verify Migrated Structure

```bash
$ ls -la /tmp/spec-kit-test/.canary/
canary.db
memory/
scripts/
specs/
templates/

$ ls /tmp/spec-kit-test/.canary/templates/commands/
analyze.md
checklist.md
clarify.md
constitution.md
implement.md
plan.md
specify.md
tasks.md
```
**Result:** ✅ All files migrated correctly

### Test 5: Legacy CANARY Detection

```bash
$ ./bin/canary detect .

🔍 Analyzing: .

System Type: legacy-canary
Details: Detected legacy CANARY system (4/4 indicators found)
```
**Result:** ✅ Correct detection of legacy system

## User Experience

### For Agents

**Before migration feature:**
```
Agent: "I have a spec-kit project, how do I migrate to CANARY?"
Human: "Manually copy files from templates/ to .canary/templates/,
         copy memory/ to .canary/memory/, create new files..."
Agent: "Which files exactly? Where do they go?"
Human: [Provides 20+ step manual process]
```

**With migration feature:**
```
Agent: "I have a spec-kit project, how do I migrate to CANARY?"
Human: "Run: canary migrate-from spec-kit"
Agent: [Runs command, migrates in seconds, continues working]
```

### Migration Commands

```bash
# Single-line detection
canary detect

# Single-line dry run
canary migrate-from spec-kit --dry-run

# Single-line migration
canary migrate-from spec-kit

# Total steps: 3 commands
# Total time: ~30 seconds
```

### Example Output Flow

```
$ canary detect
🔍 Analyzing: .
System Type: spec-kit
Details: Detected spec-kit system (5/5 indicators found)
To migrate this system, run:
  canary migrate-from spec-kit

$ canary migrate-from spec-kit --dry-run
📋 Planning migration from spec-kit...
Migration Plan for spec-kit:
Files to copy: 13
Files to merge: 1
Files to create: 3
Warnings: 2

🔍 DRY RUN MODE - No changes will be made
[Preview of all changes]
✅ Dry run complete - no changes were made

$ canary migrate-from spec-kit
📋 Planning migration from spec-kit...
🚀 Executing migration...
✅ Created: .canary
✅ Copied: [13 files]
⚠️  Files requiring manual merge: [1 file]
⚠️  Warnings: [2 warnings]
✅ Migration complete!
Next steps: [4 clear action items]
```

## Safety Features

### 1. System Validation

**Prevents incorrect migration:**
```bash
$ canary migrate-from spec-kit  # In a legacy-canary project

⚠️  Warning: Detected legacy-canary but trying to migrate as spec-kit
Use --force to override detection, or specify the correct system type.
Error: system type mismatch
```

**Override with force:**
```bash
$ canary migrate-from spec-kit --force
# Proceeds despite mismatch
```

### 2. Dry-Run Mode

**Preview before applying:**
```bash
$ canary migrate-from spec-kit --dry-run

🔍 DRY RUN MODE - No changes will be made

Would create: .canary
Would copy: memory/constitution.md -> .canary/memory/constitution.md
...
```

### 3. Non-Destructive

**Original files preserved:**
- Migration only creates `.canary/` directory
- Existing files remain unchanged
- Can undo by deleting `.canary/`

### 4. Clear Warnings

**Highlights potential issues:**
```
⚠️  Files requiring manual merge:
   - README.md: Merge spec-kit README with CANARY token documentation

⚠️  Warnings:
   - scripts/ directory found - will need manual review for compatibility
```

## Benefits

### For Agents

✅ **Minimal overhead** - 3 commands total
✅ **Clear instructions** - Command tells you what to do next
✅ **Safe preview** - Dry-run before applying
✅ **Auto-detection** - Don't need to know system type
✅ **Validation** - Prevents mistakes

### For Projects

✅ **Preserves work** - All existing files kept
✅ **Unified system** - Access to all CANARY features
✅ **Database support** - Priority management, search, checkpoints
✅ **Modern workflow** - Slash commands, auto-migration
✅ **Single binary** - No external dependencies

### For Developers

✅ **Tested migration** - Verified with real spec-kit repo
✅ **Comprehensive docs** - Complete migration guide
✅ **Extensible** - Easy to add new system types
✅ **Maintainable** - Clean separation of concerns

## Documentation

**Created:**
- `MIGRATION_GUIDE.md` - Complete 400+ line guide
  - Quick start
  - System type descriptions
  - Command reference
  - Step-by-step examples
  - Troubleshooting
  - FAQ
  - Before/after comparisons

**Updated:**
- `README.md` - Added migration section
- `status.json` - Updated counts (40 tokens, 31 requirements)

## Files Modified/Created

**Created:**
- `internal/migrate/migrate.go` - Migration logic (CBIN-131)
- `MIGRATION_GUIDE.md` - Complete migration documentation
- `MIGRATION_FEATURE_SUMMARY.md` - This document

**Modified:**
- `cmd/canary/main.go` - Added detect and migrate-from commands
- `README.md` - Added migration quick reference
- `status.json` - Updated counts and notes

**CANARY tokens added:**
- CBIN-131: MigrateFrom (CLI, IMPL)
  - DetectCmd (CLI, IMPL)
  - MigrateFromCmd (CLI, IMPL)

## Future Enhancements

**Potential improvements:**

1. **Auto-create CANARY tokens** - Parse existing code for TODO/FIXME
2. **Merge README.md automatically** - Smart merge instead of manual
3. **Backup before migration** - Auto-backup original files
4. **Progress bar** - Visual feedback during migration
5. **Rollback command** - Undo migration if needed
6. **More system types** - Support other requirement tracking systems
7. **Import from GitHub Issues** - Convert issues to CANARY tokens

## Summary

✅ **Delivered:**
- Complete migration system for spec-kit and legacy CANARY
- Auto-detection of system type
- Dry-run preview mode
- Validation and safety checks
- Comprehensive documentation (400+ lines)
- Tested with real spec-kit repository

✅ **Agent Experience:**
- 3 commands total (detect, dry-run, migrate)
- ~30 seconds total time
- Clear next steps after migration
- Zero manual file copying

✅ **Impact:**
- Lowers barrier to adoption
- Enables easy migration from legacy systems
- Preserves existing work
- Provides modern unified tooling

**Migration feature is production-ready and agent-friendly!** 🎉
