---
description: Display all CANARY tokens for a specific requirement ID
---

<!-- CANARY: REQ=CBIN-CLI-001; FEATURE="ShowCmd"; ASPECT=Docs; STATUS=IMPL; OWNER=canary; UPDATED=2025-10-16 -->

## User Input

```text
$ARGUMENTS
```

## Outline

Display all CANARY tokens for a specific requirement, organized by aspect or status.

1. **Parse requirement ID**:
   - Extract REQ-ID from arguments (e.g., {{.ReqID}}-<ASPECT>-133)
   - Validate format (should match XXXX-NNN pattern)

2. **Run canary show command**:
   ```bash
   canary show <REQ-ID>
   ```

   Available flags:
   - `--group-by aspect`: Group by aspect (CLI, API, Engine, etc.) [default]
   - `--group-by status`: Group by status (STUB, IMPL, TESTED, BENCHED)
   - `--json`: Output in JSON format for parsing
   - `--no-color`: Disable colored output
   - `--db <path>`: Custom database path (default: `.canary/canary.db`)

3. **Display results**:
   - Show feature name, aspect, status
   - Include file location and line number
   - List test and benchmark references
   - Show owner and priority information
   - Group by aspect (default) or status

4. **Analyze output**:
   - Count total tokens for requirement
   - Identify missing tests (STATUS=IMPL without TEST=)
   - Identify missing benchmarks (STATUS=TESTED without BENCH=)
   - Note file locations for implementation

5. **Provide recommendations**:
   - If STATUS=STUB: "Use `/canary.plan <REQ-ID>` to create implementation plan"
   - If STATUS=IMPL without tests: "Add TEST= field and create test functions"
   - If STATUS=TESTED without benchmarks: "Add BENCH= field for performance testing"

## Example Output

```markdown
## Tokens for {{.ReqID}}-<ASPECT>-133

### CLI Aspect

📌 {{.ReqID}}-<ASPECT>-133 - CreateCommand
   Status: TESTED | Aspect: CLI | Priority: 1
   Location: cmd/canary/create.go:25
   Test: TestCANARY_{{.ReqID}}-<ASPECT>_133_CLI_CreateCommand
   Owner: canary-team

📌 {{.ReqID}}-<ASPECT>-133 - CreateCommandHelp
   Status: IMPL | Aspect: CLI | Priority: 2
   Location: cmd/canary/create.go:112

### Engine Aspect

📌 {{.ReqID}}-<ASPECT>-133 - TokenGeneration
   Status: BENCHED | Aspect: Engine | Priority: 1
   Location: internal/engine/generate.go:45
   Test: TestCANARY_{{.ReqID}}-<ASPECT>_133_Engine_TokenGeneration
   Bench: BenchCANARY_{{.ReqID}}-<ASPECT>_133_Engine_TokenGeneration
   Owner: core-team

**Summary:**
- Total: 3 tokens
- BENCHED: 1 (33%)
- TESTED: 1 (33%)
- IMPL: 1 (33%)

**Recommendations:**
- {{.ReqID}}-<ASPECT>-133/CreateCommandHelp: Add tests for IMPL status token
```

## Guidelines

- **Automatic Execution**: Run command without prompting unless REQ-ID is missing
- **Visual Grouping**: Use clear headers and emoji indicators
- **Actionable**: Provide specific next steps based on token status
- **Database Required**: If database doesn't exist, suggest running `canary index`
- **Context**: Include summary statistics and recommendations
