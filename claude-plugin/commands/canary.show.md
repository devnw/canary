---
description: Display all CANARY tokens for a specific requirement ID
---


## User Input

```text
$ARGUMENTS
```

## Outline

Display all CANARY tokens for a specific requirement, organized by aspect or status.

1. **Read project configuration**:
   - Load `.canary/project.yaml` to determine the project key

2. **Parse requirement ID**:
   - Extract REQ-ID from arguments (e.g., `<PROJECT_KEY>-<NNN>`)
   - Validate format (should match the project's ID pattern)

3. **Run canary show command**:
   ```bash
   canary show <REQ-ID>
   ```

   Available flags:
   - `--group-by aspect`: Group by aspect (CLI, API, Engine, etc.) [default]
   - `--group-by status`: Group by status (STUB, IMPL, TESTED, BENCHED)
   - `--json`: Output in JSON format for parsing
   - `--no-color`: Disable colored output
   - `--db <path>`: Custom database path (default: `.canary/canary.db`)

4. **Display results**:
   - Show feature name, aspect, status
   - Include file location and line number
   - List test and benchmark references
   - Show owner and priority information
   - Group by aspect (default) or status

5. **Analyze output**:
   - Count total tokens for requirement
   - Identify missing tests (STATUS=IMPL without TEST=)
   - Identify missing benchmarks (STATUS=TESTED without BENCH=)
   - Note file locations for implementation

6. **Provide recommendations**:
   - If STATUS=STUB: "Use `/canary.plan <REQ-ID>` to create implementation plan"
   - If STATUS=IMPL without tests: "Add TEST= field and create test functions"
   - If STATUS=TESTED without benchmarks: "Add BENCH= field for performance testing"

## Example Output

```markdown
## Tokens for <PROJECT_KEY>-<NNN>

### API Aspect

<PROJECT_KEY>-<NNN> - UserAuthentication
   Status: TESTED | Aspect: API | Priority: 1
   Location: src/api/auth.go:25
   Test: TestCANARY_<PROJECT_KEY>_<NNN>_API_UserAuthentication
   Owner: api-team

<PROJECT_KEY>-<NNN> - ValidationMiddleware
   Status: IMPL | Aspect: API | Priority: 2
   Location: src/api/middleware.go:45

### Storage Aspect

<PROJECT_KEY>-<NNN> - SessionStore
   Status: BENCHED | Aspect: Storage | Priority: 1
   Location: internal/storage/session.go:67
   Test: TestCANARY_<PROJECT_KEY>_Storage_<NNN>_SessionStore
   Bench: BenchCANARY_<PROJECT_KEY>_Storage_<NNN>_SessionStore
   Owner: backend-team

**Summary:**
- Total: 3 tokens
- BENCHED: 1 (33%)
- TESTED: 1 (33%)
- IMPL: 1 (33%)

**Recommendations:**
- <PROJECT_KEY>-<NNN>/ValidationMiddleware: Add tests for IMPL status token
```

## Guidelines

- **Automatic Execution**: Run command without prompting unless REQ-ID is missing
- **Visual Grouping**: Use clear headers for aspect/status grouping
- **Actionable**: Provide specific next steps based on token status
- **Database Required**: If database doesn't exist, suggest running `canary index`
- **Context**: Include summary statistics and recommendations
