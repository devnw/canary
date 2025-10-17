---
description: List all implementation files containing tokens for a requirement
---

<!-- CANARY: REQ=CBIN-CLI-001; FEATURE="FilesCmd"; ASPECT=Docs; STATUS=IMPL; OWNER=canary; UPDATED=2025-10-16 -->

## User Input

```text
$ARGUMENTS
```

## Outline

List all implementation files for a requirement, grouped by aspect with token counts.

1. **Parse requirement ID**:
   - Extract REQ-ID from arguments (e.g., {{.ReqID}}-<ASPECT>-XXX)
   - Validate format

2. **Run canary files command**:
   ```bash
   canary files <REQ-ID>
   ```

   Available flags:
   - `--all`: Include spec and template files (by default excluded)
   - `--db <path>`: Custom database path (default: `.canary/canary.db`)

3. **Display results**:
   - Group files by aspect (CLI, API, Engine, Storage, etc.)
   - Show token count per file
   - Exclude spec/template files by default
   - Show total file count and total token count

4. **Analyze file distribution**:
   - Identify which aspects have implementation
   - Note missing aspects that should have files
   - Check for scattered tokens (many files with 1 token each)

5. **Provide navigation**:
   - List actual file paths for easy navigation
   - Suggest specific files to open based on user's intent

## Example Output

```markdown
## Implementation Files for {{.ReqID}}-<ASPECT>-105

### CLI
- cmd/canary/init.go (4 tokens)
- cmd/canary/init_test.go (3 tokens)

### Engine
- internal/engine/init.go (2 tokens)
- internal/engine/templates.go (1 token)

### Storage
- internal/storage/setup.go (1 token)

### Docs
- .canary/AGENT_CONTEXT.md (1 token)
- CLAUDE.md (1 token)

**Total: 7 files, 13 tokens**

**Analysis:**
- Primary implementation: cmd/canary/init.go:356 (CLI)
- Engine support: internal/engine/init.go, templates.go
- Storage: Basic setup in internal/storage/setup.go

**Navigation:**
- To view CLI command: `cmd/canary/init.go:356`
- To view tests: `cmd/canary/init_test.go`
- To view engine logic: `internal/engine/init.go`
```

## Guidelines

- **Automatic Execution**: Run command immediately if REQ-ID is provided
- **Focus on Implementation**: Exclude spec/template files by default
- **Grouping**: Group by aspect for clarity
- **Navigation**: Provide file:line references for easy IDE navigation
- **Analysis**: Note primary implementation files vs support files
- **Database Required**: Suggest `canary index` if database missing
