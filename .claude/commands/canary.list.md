# CANARY: REQ=CBIN-135; FEATURE="AgentSlashCommand"; ASPECT=Docs; STATUS=IMPL; UPDATED=2025-10-16

List top priority CANARY requirements with filtering and sorting.

## Usage

```bash
canary list [flags]
```

## Common Patterns for AI Agents

**View top priorities (default: 10 items):**
```bash
canary list
```

**Find new work to implement:**
```bash
canary list --status STUB --limit 5
```

**Find requirements needing tests:**
```bash
canary list --status IMPL --limit 10
```

**Focus on specific aspect:**
```bash
canary list --aspect CLI --status STUB
```

**View all CLI work:**
```bash
canary list --aspect CLI --limit 0
```

**Sort by last updated (find stale requirements):**
```bash
canary list --order-by "updated_at ASC" --limit 20
```

## Available Flags

- `--limit N`: Maximum results (default: unlimited, use 10 for typical agent queries)
- `--status VALUE`: Filter by status (STUB, IMPL, TESTED, BENCHED, REMOVED)
- `--aspect VALUE`: Filter by aspect (CLI, API, Engine, Storage, Security, Docs)
- `--owner NAME`: Filter by owner
- `--phase VALUE`: Filter by phase (Phase0, Phase1, Phase2, Phase3)
- `--order-by CLAUSE`: Custom SQL ORDER BY (default: "priority ASC, updated_at DESC")
- `--json`: Output as JSON for parsing

## Output Format

Default table format shows:
- **ID**: Requirement identifier (CBIN-XXX)
- **Feature**: Feature name
- **Status**: Current status
- **Aspect**: Category
- **Priority**: Priority level (1=highest)
- **Location**: File path and line number

Example output:
```
Found 5 tokens:

📌 CBIN-105 - InitWorkflow
   Status: IMPL | Aspect: CLI | Priority: 5
   Location: cmd/canary/main.go:356

📌 CBIN-134 - SpecModification
   Status: STUB | Aspect: CLI | Priority: 1
   Location: .canary/specs/CBIN-134-spec-modification/spec.md:1
```

## Context Usage Tips

- Default output with `--limit 10` uses approximately 1500-2000 tokens
- Use `--limit 5` for minimal context (~800 tokens)
- Use `--limit 3` for ultra-minimal context (~500 tokens)
- JSON output (`--json`) is more compact for programmatic parsing

## Integration with Workflow

**Before starting work:**
```bash
canary list --status STUB --aspect CLI --limit 5
```

**Finding next priority:**
```bash
canary next
```
(The `canary next` command automatically selects highest priority unblocked work)

**After implementation:**
```bash
canary list --status IMPL --limit 10
```
(Find items needing tests)

## Requirements

- Database must exist: run `canary index` first to build `.canary/canary.db`
- If database doesn't exist, command will return error suggesting to run `canary index`
