# Migration Guide: spec-kit & Legacy CANARY → Unified CANARY

## Overview

This guide helps you migrate existing spec-kit or legacy CANARY projects to the new unified CANARY system with minimal overhead.

## Quick Start

```bash
# 1. Detect your system type
canary detect

# 2. Preview migration (dry run)
canary migrate-from spec-kit --dry-run  # or legacy-canary

# 3. Execute migration
canary migrate-from spec-kit

# 4. Initialize database and index tokens
canary index

# 5. Run scanner
canary scan --root . --out status.json
```

## System Types

### spec-kit

A specification-driven development system with:
- `memory/constitution.md` - Project principles
- `templates/spec-template.md` - Specification templates
- `templates/commands/` - Slash commands for AI agents
- Focus on requirements and planning

**Detection indicators:**
- `memory/constitution.md`
- `templates/spec-template.md`
- `templates/plan-template.md`
- `templates/commands/specify.md`
- `templates/commands/plan.md`

### legacy-canary

A CANARY token-based tracking system with:
- `tools/canary/` - Standalone scanner
- `status.json` - Scanner output
- `GAP_ANALYSIS.md` - Requirement tracking
- CANARY tokens in source code
- No `.canary/` directory (pre-unified system)

**Detection indicators (needs 2/4):**
- `tools/canary/`
- `tools/canary/main.go`
- `status.json`
- `GAP_ANALYSIS.md`

**Example:** This repository itself was a legacy CANARY system before commit `846a6de` (Sept 2025), when it only had the standalone scanner in `tools/canary/` and no modern `.canary/` structure.

### migrated

A system that has already been migrated to the unified CANARY system:
- Has `.canary/` directory with modern structure
- May have `.canary/canary.db` (SQLite database)
- May have `.canary/templates/` (spec-kit templates)
- May still have legacy files (`tools/canary/`, `status.json`)

**Detection indicators:**
- `.canary/canary.db` exists, OR
- `.canary/templates/` exists

**Note:** If a system is detected as "migrated", the migration commands will refuse to run and suggest using the modern commands instead (`canary index`, `canary list`, etc.).

## Commands

### canary detect

Analyzes a directory to identify the system type.

```bash
# Detect in current directory
canary detect

# Detect in specific directory
canary detect /path/to/project
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

### canary migrate-from

Migrates an existing project to the unified CANARY system.

**Syntax:**
```bash
canary migrate-from <system-type> [directory] [flags]
```

**System types:**
- `spec-kit` - Migrate from spec-kit
- `legacy-canary` - Migrate from legacy CANARY

**Flags:**
- `--dry-run` - Preview changes without applying them
- `--force` - Force migration even if detection doesn't match

**Examples:**
```bash
# Dry run in current directory
canary migrate-from spec-kit --dry-run

# Migrate spec-kit project
canary migrate-from spec-kit

# Migrate legacy canary project
canary migrate-from legacy-canary

# Migrate specific directory
canary migrate-from spec-kit /path/to/project

# Force migration despite detection mismatch
canary migrate-from spec-kit --force
```

## Migration Process

### spec-kit Migration

**What gets migrated:**

1. **Copied directly:**
   - `memory/constitution.md` → `.canary/memory/constitution.md`
   - `templates/spec-template.md` → `.canary/templates/spec-template.md`
   - `templates/plan-template.md` → `.canary/templates/plan-template.md`
   - `templates/tasks-template.md` → `.canary/templates/tasks-template.md`
   - `templates/checklist-template.md` → `.canary/templates/checklist-template.md`
   - `templates/commands/*.md` → `.canary/templates/commands/*.md`

2. **Created:**
   - `README_CANARY.md` - CANARY token specification
   - `GAP_ANALYSIS.md` - Requirement tracking
   - `CLAUDE.md` - AI agent integration guide

3. **Manual merge required:**
   - `README.md` - Merge spec-kit README with CANARY documentation

4. **Warnings:**
   - `scripts/` directory - Review for compatibility
   - `.canary/` already exists - Will merge content

**Example output:**
```
📋 Planning migration from spec-kit...

Migration Plan for spec-kit:

Files to copy: 13
Files to merge: 1
Files to create: 3
Warnings: 2

🚀 Executing migration...

✅ Created: .canary
✅ Created: .canary/memory
✅ Created: .canary/templates
✅ Created: .canary/templates/commands
✅ Created: .canary/scripts
✅ Created: .canary/specs
✅ Copied: memory/constitution.md -> .canary/memory/constitution.md
✅ Copied: templates/spec-template.md -> .canary/templates/spec-template.md
...

⚠️  Files requiring manual merge:
   - README.md: Merge spec-kit README with CANARY token documentation

⚠️  Warnings:
   - scripts/ directory found - will need manual review for compatibility

✅ Migration complete!
```

### legacy-canary Migration

**What gets migrated:**

1. **Preserved in place:**
   - `status.json` - Scanner output
   - `status.csv` - CSV output
   - `GAP_ANALYSIS.md` - Requirement tracking

2. **Created:**
   - `.canary/memory/constitution.md` - Project principles
   - `.canary/templates/spec-template.md` - Specification template
   - `.canary/templates/plan-template.md` - Implementation plan template
   - `README_CANARY.md` - Token specification
   - `CLAUDE.md` - AI agent integration

3. **Warnings:**
   - `tools/canary/` - Legacy scanner (can be removed after migration)

## Post-Migration Steps

### 1. Review Migrated Files

```bash
# Check .canary directory structure
ls -R .canary/

# Review templates
cat .canary/templates/spec-template.md

# Review slash commands
ls .canary/templates/commands/
```

### 2. Update Slash Commands

Edit slash commands in `.canary/templates/commands/` to match your workflow:

```bash
# Example: Update specify command
vim .canary/templates/commands/specify.md
```

### 3. Initialize Database

```bash
# Create database and run migrations
canary index

# Verify database
sqlite3 .canary/canary.db "SELECT COUNT(*) FROM tokens"
```

### 4. Scan for Tokens

```bash
# Scan codebase
canary scan --root . --out status.json

# View results
cat status.json
```

### 5. Test Commands

```bash
# List tokens
canary list --limit 10

# Search tokens
canary search "authentication"

# View implementation points
canary implement CBIN-001
```

## Troubleshooting

### Detection Not Working

**Problem:** `canary detect` returns "unknown"

**Solution:**
- Ensure you're in the project root directory
- Check if indicator files exist (see Detection indicators above)
- Use `--force` flag with `migrate-from` to override detection

### .canary Already Exists

**Problem:** Migration warns that `.canary/` exists

**Solution:**
- Migration will merge content safely
- Existing files won't be overwritten
- Review warnings after migration

### System Type Mismatch

**Problem:** Detection shows different type than expected

**Solution:**
```bash
# Force migration to desired type
canary migrate-from spec-kit --force
```

### Missing Files After Migration

**Problem:** Some templates or commands are missing

**Solution:**
```bash
# Re-initialize to get default templates
canary init

# Copy missing templates from base/.canary/
```

## Comparison: Before vs After

### Before Migration (spec-kit)

```
my-project/
├── memory/
│   └── constitution.md
├── templates/
│   ├── spec-template.md
│   ├── plan-template.md
│   └── commands/
│       ├── specify.md
│       └── plan.md
├── scripts/
└── README.md
```

### After Migration (unified CANARY)

```
my-project/
├── .canary/
│   ├── memory/
│   │   └── constitution.md
│   ├── templates/
│   │   ├── spec-template.md
│   │   ├── plan-template.md
│   │   └── commands/
│   │       ├── specify.md
│   │       ├── plan.md
│   │       ├── tasks.md
│   │       └── ...
│   ├── scripts/
│   ├── specs/
│   └── canary.db
├── memory/              # Original preserved
├── templates/           # Original preserved
├── scripts/             # Original preserved
├── README.md
├── README_CANARY.md     # New
├── GAP_ANALYSIS.md      # New
├── CLAUDE.md            # New
└── status.json          # After canary scan
```

## Benefits of Migration

### Unified System

✅ **Single tool** - One binary for all workflows
✅ **Embedded templates** - No external dependencies
✅ **Database storage** - Advanced querying and priority management
✅ **Auto-migration** - Database schema upgrades automatically
✅ **Cross-platform** - Pure Go, works everywhere

### Enhanced Features

✅ **Priority management** - Order work by importance
✅ **Keyword search** - Find related features quickly
✅ **Checkpoints** - Track progress over time
✅ **Git integration** - Automatic commit/branch tracking
✅ **Phase tracking** - Organize by development stages

### Better Agent Integration

✅ **Context reduction** - `canary implement` shows exact locations
✅ **Slash commands** - AI-friendly workflow commands
✅ **Auto-migration** - No manual database management
✅ **Comprehensive docs** - CLAUDE.md, README_CANARY.md

## Example Workflow

### Complete Migration Example

```bash
# 1. Clone or navigate to existing project
cd my-spec-kit-project

# 2. Detect system type
canary detect
# Output: System Type: spec-kit

# 3. Preview migration
canary migrate-from spec-kit --dry-run
# Review: Files to copy: 13, Files to merge: 1, Warnings: 2

# 4. Execute migration
canary migrate-from spec-kit
# ✅ Migration complete!

# 5. Initialize database
canary index
# ✅ Indexed 45 CANARY tokens

# 6. List high-priority work
canary list --status STUB --order-by "priority ASC" --limit 5

# 7. Find specific feature
canary implement CBIN-001 --feature Authentication --context

# 8. Create checkpoint
canary checkpoint "post-migration" "Migrated from spec-kit"

# 9. Run legacy scanner for comparison
canary scan --root . --out status.json

# 10. Verify everything works
canary search "authentication"
canary list --phase Phase1
```

## FAQ

### Q: Will my existing files be modified?

**A:** No. The migration only copies files to `.canary/` and creates new files. Your original templates and documents remain unchanged.

### Q: Can I undo the migration?

**A:** Yes. Simply delete the `.canary/` directory and the new files (`README_CANARY.md`, `CLAUDE.md`, etc.). Your original files are preserved.

### Q: What if I have custom templates?

**A:** Custom templates in `templates/` are copied to `.canary/templates/`. You can continue using them and modify as needed.

### Q: Do I need to remove `tools/canary/`?

**A:** No, but it's recommended. The new binary includes the scanner functionality. You can safely remove `tools/canary/` after verifying the migration works.

### Q: Will this break my existing workflow?

**A:** No. The migration is additive. You can continue using your existing tools alongside the new CANARY system until you're ready to switch completely.

### Q: Can I migrate multiple projects?

**A:** Yes. Run the migration in each project directory:
```bash
canary migrate-from spec-kit /path/to/project1
canary migrate-from spec-kit /path/to/project2
```

### Q: What about Python dependencies (spec-kit)?

**A:** The unified CANARY system doesn't require Python. All functionality is in the single Go binary. You can remove `pyproject.toml` and Python dependencies after migration if you're no longer using the spec-kit Python CLI.

## Support

If you encounter issues:

1. **Check detection:** Run `canary detect` to verify system type
2. **Use dry-run:** Preview with `--dry-run` before applying
3. **Review warnings:** Address warnings shown during migration
4. **Consult docs:** See `README.md`, `MIGRATIONS.md`, `CLI_COMMANDS.md`

## Summary

The migration tool makes it easy to adopt the unified CANARY system:

✅ **Automatic detection** - Identifies your system type
✅ **Dry-run mode** - Preview changes safely
✅ **Non-destructive** - Preserves original files
✅ **Clear feedback** - Shows exactly what will change
✅ **Post-migration guide** - Clear next steps

**Total migration time:** ~2-5 minutes depending on project size

**Recommended approach:** Always start with `--dry-run` to understand the changes before applying them.
