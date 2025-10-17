# CANARY: REQ=CBIN-132; FEATURE="NextCmd"; ASPECT=CLI; STATUS=STUB; OWNER=canary; UPDATED=2025-10-16

Identify and implement the next highest priority CANARY requirement.

## Workflow

1. **Query next priority:**
   ```bash
   canary next --prompt
   ```

2. **Review generated guidance:**
   - Read specification details
   - Review constitutional principles
   - Check dependencies
   - Understand test requirements

3. **Implement following test-first approach:**
   - Write tests first (Article IV)
   - Place CANARY token
   - Implement functionality
   - Verify success criteria
   - Update token STATUS

4. **Verify completion:**
   ```bash
   canary scan --root . --project-only
   canary scan --root . --verify GAP_ANALYSIS.md --project-only
   ```

## Command Options

- `canary next` - Show next priority requirement summary
- `canary next --prompt` - Generate full implementation guidance
- `canary next --json` - Machine-readable output
- `canary next --status STUB` - Filter by status
- `canary next --aspect API` - Filter by aspect

## Priority Factors

The system determines priority based on:
1. **PRIORITY field** (1=highest, 10=lowest)
2. **STATUS** (STUB > IMPL > TESTED)
3. **Dependencies** (DEPENDS_ON must be satisfied)
4. **Age** (older UPDATED dates get priority boost)

## Constitutional Principles

Every implementation must follow:
- **Article I**: Requirement-First Development
- **Article IV**: Test-First Imperative
- **Article V**: Simplicity and Anti-Abstraction
- **Article VII**: Documentation Currency

See `.canary/memory/constitution.md` for complete principles.

## Example Usage

```bash
# AI agent automatically gets next task
/canary.next

# The system will:
# 1. Query database (or scan filesystem)
# 2. Identify highest priority STUB/IMPL requirement
# 3. Load specification from .canary/specs/
# 4. Load constitutional principles
# 5. Resolve dependencies
# 6. Generate comprehensive implementation prompt
# 7. Provide test examples and token placement guidance
```

## Expected Output

The command generates a comprehensive prompt including:
- Requirement specification details
- Constitutional guidance for this implementation
- Dependency verification
- Test-first implementation steps
- CANARY token examples
- Success criteria checklist
- Suggested file locations
- Verification steps

## Notes

- If no requirements are available, displays completion message
- If dependencies are blocking, selects next unblocked requirement
- Falls back to filesystem scan if database unavailable
- Prompts are 2,000-5,000 words for comprehensive guidance
