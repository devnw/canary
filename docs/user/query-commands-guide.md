# Query Commands User Guide

**Requirement:** CBIN-CLI-001
**Status:** Complete
**Last Updated:** 2025-10-17

## Overview

CANARY provides four powerful query commands that let you inspect, analyze, and navigate your requirement tokens without resorting to grep or SQL. These commands replace low-level shell operations with high-level, formatted output designed for both human developers and AI agents.

**The Query Commands:**
- `canary show` - Display all tokens for a requirement
- `canary files` - List implementation files for a requirement
- `canary status` - Check implementation progress for a requirement
- `canary grep` - Search tokens by pattern across all fields

**Key Benefits:**
- **Fast**: Database-backed queries complete in <100ms
- **Formatted**: Human-readable tables, not raw grep output
- **Filtered**: Automatically excludes templates, specs, and test files (configurable)
- **Informative**: Shows exactly what you need without extra noise
- **Fallback**: Works even without a database (slower filesystem search)

## Getting Started

### Prerequisites

- CANARY CLI installed
- (Recommended) Database built with `canary index`
- CANARY tokens in your codebase

### Quick Start

See all tokens for a requirement:

```bash
canary show CBIN-133
```

List files implementing a requirement:

```bash
canary files CBIN-133
```

Check implementation progress:

```bash
canary status CBIN-133
```

Search tokens by keyword:

```bash
canary grep FuzzyMatcher
```

## Commands

### `canary show` - Display Tokens

Shows all CANARY tokens for a specific requirement ID with detailed information.

**Usage:**
```bash
canary show <REQ-ID> [flags]
```

**Flags:**
- `--group-by <field>` - Group output by aspect, status, or file (default: aspect)
- `--json` - Output as JSON for machine processing
- `--no-color` - Disable colored output

**Example 1: Basic usage**
```bash
$ canary show CBIN-133
```

**Output:**
```
Tokens for CBIN-133:

## CLI

📌 CBIN-133 - ImplementCmd
   Status: TESTED | Aspect: CLI | Priority: 5
   Location: ./cmd/canary/main.go:861
   Test: TestCANARY_CBIN_133_CLI_ExactMatch
   Owner: canary

📌 CBIN-133 - ImplementCommand
   Status: TESTED | Aspect: CLI | Priority: 5
   Location: ./cmd/canary/implement.go:45
   Test: TestCANARY_CBIN_133_CLI_ImplementCommand

## API

📌 CBIN-133 - RequirementLookup
   Status: TESTED | Aspect: API | Priority: 5
   Location: ./cmd/canary/implement.go:112
   Test: TestCANARY_CBIN_133_API_RequirementLookup

## Engine

📌 CBIN-133 - FuzzyMatcher
   Status: TESTED | Aspect: Engine | Priority: 5
   Location: ./internal/matcher/fuzzy.go:23
   Test: TestCANARY_CBIN_133_Engine_FuzzyMatch

Total: 4 tokens (4 TESTED, 0 IMPL, 0 STUB)
```

**Example 2: Group by status**
```bash
$ canary show CBIN-134 --group-by status
```

**Output:**
```
Tokens for CBIN-134:

## TESTED

📌 CBIN-134 - UpdateSubcommand (CLI)
   Location: ./cmd/canary/specify_update.go:25

📌 CBIN-134 - ExactIDLookup (Engine)
   Location: ./internal/specs/lookup.go:15

## IMPL

📌 CBIN-134 - AdvancedFiltering (API)
   Location: ./cmd/canary/specify_update.go:89
   Status: Needs tests
```

**Example 3: JSON output for scripts**
```bash
$ canary show CBIN-133 --json
```

**Output:**
```json
[
  {
    "req_id": "CBIN-133",
    "feature": "ImplementCmd",
    "aspect": "CLI",
    "status": "TESTED",
    "file_path": "./cmd/canary/main.go",
    "line_number": 861,
    "test": "TestCANARY_CBIN_133_CLI_ExactMatch",
    "owner": "canary",
    "priority": 5
  },
  ...
]
```

### `canary files` - List Implementation Files

Lists all files containing implementation tokens for a requirement, excluding specs and templates.

**Usage:**
```bash
canary files <REQ-ID> [flags]
```

**Flags:**
- `--all` - Include spec and template files (normally filtered out)
- `--aspect <aspect>` - Filter to specific aspect (CLI, API, Engine, etc.)

**Example 1: Basic usage**
```bash
$ canary files CBIN-133
```

**Output:**
```
Implementation Files for CBIN-133:

CLI (2 files):
  • cmd/canary/main.go (1 token)
  • cmd/canary/implement.go (1 token)

API (1 file):
  • cmd/canary/implement.go (2 tokens)

Engine (1 file):
  • internal/matcher/fuzzy.go (1 token)

Total: 3 unique files, 5 tokens
```

**Example 2: Include specs and plans**
```bash
$ canary files CBIN-133 --all
```

**Output:**
```
All Files for CBIN-133:

CLI (4 files):
  • cmd/canary/main.go (1 token)
  • cmd/canary/implement.go (1 token)
  • .canary/specs/CBIN-133-implement/spec.md (3 tokens)
  • .canary/specs/CBIN-133-implement/plan.md (7 tokens)

[... rest of output ...]

Total: 5 unique files, 15 tokens (including 10 in specs/plans)
```

**Example 3: Filter by aspect**
```bash
$ canary files CBIN-133 --aspect Engine
```

**Output:**
```
Engine Files for CBIN-133:

  • internal/matcher/fuzzy.go (1 token)

Total: 1 file
```

### `canary status` - Check Implementation Progress

Shows implementation progress summary including completion percentage and remaining work.

**Usage:**
```bash
canary status <REQ-ID>
```

**Example 1: Fully completed requirement**
```bash
$ canary status CBIN-133
```

**Output:**
```
Implementation Status for CBIN-133:

Progress: [====================================] 100%

Total:     4 tokens
Completed: 4 (100%)
Status Breakdown:
  • TESTED: 4

✅ All features completed!
```

**Example 2: In-progress requirement**
```bash
$ canary status CBIN-134
```

**Output:**
```
Implementation Status for CBIN-134:

Progress: [===========>                            ] 29%

Total:     27 tokens
Completed: 8 (29%)
In Progress:
  • IMPL:   11
  • STUB:   8

Status Breakdown:
  • TESTED: 8
  • IMPL:   11
  • STUB:   8

Incomplete Work:
  IMPL UpdateSubcommand - ./cmd/canary/specify_update.go
  IMPL FuzzySearch - ./internal/specs/search.go
  STUB AdvancedFilters - ./cmd/canary/specify_update.go
  STUB SectionParser - ./internal/specs/parser.go
  STUB ValidationRules - ./internal/specs/validate.go
  [... 3 more ...]

Next Steps:
  • Complete 11 IMPL features (add tests)
  • Implement 8 STUB features
```

**Example 3: Requirement not started**
```bash
$ canary status CBIN-999
```

**Output:**
```
Error: No tokens found for CBIN-999

Suggestions:
  • Check requirement ID spelling
  • Run: canary list to see all requirements
  • Run: canary grep "keywords" to search
```

### `canary grep` - Search Tokens by Pattern

Search for tokens matching a pattern across requirement ID, feature name, aspect, owner, and keywords.

**Usage:**
```bash
canary grep <pattern> [flags]
```

**Flags:**
- `--case-sensitive` - Enable case-sensitive search (default: case-insensitive)
- `--regex` - Treat pattern as regular expression
- `--field <field>` - Search only specific field (req_id, feature, aspect, owner, keywords)

**Example 1: Basic keyword search**
```bash
$ canary grep Fuzzy
```

**Output:**
```
Search results for 'Fuzzy' (3 tokens):

📌 CBIN-133 - FuzzyMatcher
   Status: TESTED | Priority: 5 | ./internal/matcher/fuzzy.go:23
   Match: Feature name "FuzzyMatcher"

📌 CBIN-133 - FuzzySearch
   Status: IMPL | Priority: 5 | ./internal/specs/search.go:45
   Match: Feature name "FuzzySearch"

📌 CBIN-134 - FuzzySpecSearch
   Status: TESTED | Priority: 5 | ./internal/specs/lookup.go:89
   Match: Feature name "FuzzySpecSearch"

Total: 3 matches
```

**Example 2: Search by aspect**
```bash
$ canary grep --field aspect Engine
```

**Output:**
```
Search results for 'Engine' in aspect field (47 tokens):

📌 CBIN-133 - FuzzyMatcher
   Status: TESTED | Aspect: Engine

📌 CBIN-134 - ExactIDLookup
   Status: TESTED | Aspect: Engine

[... 45 more ...]

Total: 47 matches
```

**Example 3: Regex search**
```bash
$ canary grep --regex "^CBIN-13[0-9]$"
```

**Output:**
```
Search results for '^CBIN-13[0-9]$' (regex) (27 tokens):

📌 CBIN-133 - ImplementCmd
   Status: TESTED | Aspect: CLI

📌 CBIN-134 - UpdateSubcommand
   Status: TESTED | Aspect: CLI

📌 CBIN-135 - PlanCmd
   Status: IMPL | Aspect: CLI

[... 24 more ...]

Total: 27 matches in 3 requirements
```

## Common Workflows

### Workflow 1: Agent Inspects Requirement Before Implementation

**Scenario:** AI agent needs to understand what's already implemented

```bash
# 1. Check overall status
canary status CBIN-138

# 2. See all existing tokens
canary show CBIN-138

# 3. Find which files need editing
canary files CBIN-138

# 4. Begin implementation with full context
```

### Workflow 2: Developer Finds Related Code

**Scenario:** Developer needs to modify all authentication-related code

```bash
# 1. Search for authentication tokens
canary grep authentication

# 2. Filter to specific aspect
canary grep --field aspect API authentication

# 3. Get files for specific requirement
canary files CBIN-107

# 4. Edit identified files
```

### Workflow 3: Project Manager Checks Progress

**Scenario:** Track implementation status for sprint planning

```bash
# 1. Check status of all sprint requirements
canary status CBIN-133
canary status CBIN-134
canary status CBIN-135

# 2. Search for incomplete work
canary grep --field status IMPL

# 3. Generate report (JSON output for processing)
canary show CBIN-133 --json > sprint-progress.json
```

### Workflow 4: Code Review Verification

**Scenario:** Reviewer verifies CANARY tokens are correct

```bash
# 1. Show all tokens for reviewed requirement
canary show CBIN-140

# 2. Verify test coverage
canary show CBIN-140 --group-by status

# 3. Check that all implementation files are included
canary files CBIN-140

# 4. Ensure tokens match actual code
```

## Best Practices

### For AI Agents

1. **Always use `canary show` instead of grep** - Formatted output is easier to parse
2. **Check status before implementing** - Understand what's already done
3. **Use `files` to locate code** - Don't grep the entire codebase
4. **Verify with `status` after changes** - Ensure tokens are updated correctly
5. **Search with `grep` for discovery** - Find related implementations

### For Human Developers

1. **Use `status` for daily standup** - Quick progress check
2. **Use `files` when starting new work** - Find existing implementations
3. **Use `grep` for refactoring** - Find all instances of a pattern
4. **Use `--json` for scripting** - Automate reporting and analysis
5. **Check multiple requirements at once** - Script loops for batch checking

### For Project Managers

1. **Track completion with `status`** - Monitor progress over time
2. **Export JSON for dashboards** - Integrate with project tools
3. **Search by owner** - See individual contributor progress
4. **Filter by aspect** - Track backend vs frontend separately
5. **Combine with checkpoints** - Historical trend analysis

## Troubleshooting

### Problem: "No tokens found" but tokens exist

**Symptoms:**
```
$ canary show CBIN-133
No tokens found for CBIN-133
```
But you know tokens exist in the code.

**Solutions:**
1. **Rebuild database**: `canary index`
2. **Check requirement ID spelling**: IDs are case-sensitive
3. **Verify database exists**: `ls -la .canary/canary.db`
4. **Check for typos**: `canary list | grep 133`

### Problem: "Database not found, using filesystem search (slower)"

**Symptoms:**
```
⚠️  Database not found, using filesystem search (slower)
```

**Solutions:**
1. **Build database**: `canary index`
2. **Check database location**: Ensure you're in project root
3. **Verify .canary directory exists**: `ls -la .canary/`

**Note:** Filesystem fallback still works, just slower.

### Problem: Query is very slow (>1 second)

**Symptoms:**
Commands take multiple seconds to complete

**Solutions:**
1. **Use database**: Run `canary index` for 100x speed improvement
2. **Limit results**: Use `--aspect` or `--field` filters
3. **Check database size**: Very large codebases may need indexing
4. **Verify SSD performance**: Database queries are disk-bound

### Problem: Output doesn't fit in terminal

**Symptoms:**
Output is too wide or too long to read comfortably

**Solutions:**
1. **Use --json for scripts**: Pipe to jq for filtering
2. **Group differently**: Try `--group-by file` instead of aspect
3. **Filter by aspect**: `canary files CBIN-133 --aspect CLI`
4. **Pipe to less**: `canary show CBIN-133 | less`
5. **Increase terminal width**: Resize terminal window

### Problem: Colors don't display correctly

**Symptoms:**
See escape codes like `[0;32m` instead of colors

**Solutions:**
1. **Use --no-color flag**: `canary show CBIN-133 --no-color`
2. **Check terminal support**: Ensure TERM environment variable is set
3. **Update terminal**: Use modern terminal emulator
4. **Pipe output**: Colors are auto-disabled when piping

## FAQ

**Q: Which command should I use to find implementation files?**

A: Use `canary files <REQ-ID>`. It automatically filters out specs and templates, showing only real implementation files.

**Q: Can I use these commands in scripts?**

A: Yes! Use `--json` flag for machine-readable output. Example:
```bash
canary show CBIN-133 --json | jq '.[] | select(.status=="TESTED")'
```

**Q: Do I need a database for these commands to work?**

A: No. Commands work without a database (using filesystem search), but database queries are 100x faster. Run `canary index` for best performance.

**Q: How do I search across all requirements?**

A: Use `canary grep <pattern>`. It searches all tokens in the database/filesystem.

**Q: Can I filter `show` output to specific file types?**

A: Not directly, but you can combine with other tools:
```bash
canary show CBIN-133 --json | jq '.[] | select(.file_path | endswith(".go"))'
```

**Q: What's the difference between `canary show` and `canary status`?**

A: `show` displays all individual tokens with details. `status` provides a summary with progress percentage and completion statistics.

**Q: How do I see only STUB or IMPL tokens (work remaining)?**

A: Use `canary status <REQ-ID>` - it automatically lists incomplete work, or filter with:
```bash
canary show CBIN-134 --group-by status | grep -A5 "STUB\|IMPL"
```

**Q: Can I search by owner or team?**

A: Yes! Use:
```bash
canary grep --field owner backend
```

## Related Documentation

- [canary list](./list-command-guide.md) - Listing and filtering all requirements
- [canary implement](./implement-command-guide.md) - Getting implementation guidance
- [canary next](./next-priority-guide.md) - Finding next priority work
- [CANARY Token Format](../../.canary/docs/token-format.md) - Token syntax reference

## Advanced Usage

### Integration with jq

Process JSON output for custom reporting:

```bash
# Count tokens by status
canary show CBIN-133 --json | jq 'group_by(.status) | map({status: .[0].status, count: length})'

# Extract file paths
canary files CBIN-133 --json | jq -r '.[] | .file_path'

# Find tokens without tests
canary show CBIN-134 --json | jq '.[] | select(.test == "" and .status != "STUB")'
```

### Scripting Workflows

Check multiple requirements:

```bash
#!/bin/bash
for req in CBIN-133 CBIN-134 CBIN-135; do
  echo "=== $req ==="
  canary status $req
  echo ""
done
```

Generate coverage report:

```bash
#!/bin/bash
echo "Requirement,Total,TESTED,Completion%"
canary list --json | jq -r '.[].req_id' | sort -u | while read req; do
  canary show $req --json | jq -r "
    \"$req,\(length),\([.[] | select(.status==\"TESTED\")] | length),\(([.[] | select(.status==\"TESTED\")] | length) / length * 100 | floor)\"
  "
done
```

### Database Queries

For advanced queries beyond what commands provide, query the database directly:

```bash
# Direct SQLite query (when commands aren't sufficient)
sqlite3 .canary/canary.db "
  SELECT req_id, COUNT(*) as total,
         SUM(CASE WHEN status='TESTED' THEN 1 ELSE 0 END) as tested
  FROM tokens
  GROUP BY req_id
  HAVING tested < total
"
```

**Note:** Prefer using `canary` commands over direct SQL when possible for consistency.

---

*Last verified: 2025-10-17 with canary v0.1.0*
*Implementation status: 100% TESTED (all 4 commands)*
