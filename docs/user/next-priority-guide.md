# Next Priority Command User Guide

**Requirement:** CBIN-132
**Status:** Complete
**Last Updated:** 2025-10-17

## Overview

The `canary next` command is your automated workflow assistant that identifies the highest priority unimplemented requirement and generates comprehensive implementation guidance. It eliminates manual prioritization decisions by intelligently selecting work based on priority scores, dependencies, status, and age.

**Key Benefits:**
- **Automated Prioritization**: No manual decision-making - system selects optimal next task
- **Comprehensive Guidance**: Generated prompts include spec, constitution, tests, and examples
- **Dependency-Aware**: Automatically resolves DEPENDS_ON relationships
- **Constitutional Adherence**: Every prompt includes relevant project principles
- **Fast Selection**: <100ms query time even in large codebases

## Getting Started

### Prerequisites

- CANARY CLI installed
- (Recommended) Database indexed with `canary index`
- (Optional) `.canary/memory/constitution.md` for project principles

### Quick Start

Get next priority task:

```bash
canary next
```

Generate full implementation prompt:

```bash
canary next --prompt
```

## Usage

### Basic Usage: Summary Mode

Display next priority requirement without full prompt:

```bash
canary next
```

**Output:**
```
📌 Next Priority: CBIN-138 - MultilineTokens

Priority: 3 | Status: STUB | Aspect: Engine
Specification: .canary/specs/CBIN-138-multiline-tokens/spec.md

Dependencies:
  ✅ CBIN-101 - ScannerCore (TESTED)

Ready to implement! Run with --prompt flag for full guidance.
```

### Prompt Generation Mode

Generate comprehensive implementation prompt:

```bash
canary next --prompt
```

**Output includes:**
1. Requirement specification
2. Implementation plan (if exists)
3. Constitutional principles
4. Test-first workflow guidance
5. Token placement examples
6. Success criteria checklist
7. Dependency information

### Filtering by Status

Select next STUB requirement only:

```bash
canary next --status STUB
```

Select next IMPL requirement needing tests:

```bash
canary next --status IMPL
```

### Filtering by Aspect

Select next Engine work:

```bash
canary next --aspect Engine
```

Select next CLI work:

```bash
canary next --aspect CLI
```

### Dry Run Mode

Preview what would be selected without generating prompt:

```bash
canary next --dry-run
```

### JSON Output

Get machine-readable output for automation:

```bash
canary next --json
```

## Common Workflows

### Workflow 1: Agent Autonomous Work

**Scenario:** AI agent completing task after task automatically

```bash
# 1. Agent completes current task
# 2. Agent runs next command
canary next --prompt

# 3. Agent reads generated prompt
# 4. Agent implements requirement following guidance
# 5. Agent places CANARY tokens
# 6. Agent updates status to TESTED
# 7. Repeat from step 2
```

This creates a continuous implementation loop where the agent always knows what to work on next.

### Workflow 2: Human Developer Daily Planning

**Scenario:** Developer starting work day

```bash
# 1. Check what's next
canary next

# 2. Review the requirement details
cat .canary/specs/CBIN-138-multiline-tokens/spec.md

# 3. Decide to implement
canary next --prompt > implementation-guidance.md

# 4. Follow guidance in implementation-guidance.md
# 5. Mark complete and get next task
canary next
```

### Workflow 3: Filtered Sprint Planning

**Scenario:** Sprint focused on API aspects

```bash
# 1. See all API work available
canary list --aspect API --status STUB

# 2. Get next API priority
canary next --aspect API --prompt

# 3. Implement API feature
# 4. Repeat for sprint duration
```

### Workflow 4: Dependency-Driven Development

**Scenario:** System automatically resolves dependencies

```bash
# Current state:
#   CBIN-105 (PRIORITY=1, DEPENDS_ON=CBIN-104)
#   CBIN-104 (PRIORITY=3, STATUS=STUB)

# System selects CBIN-104 first (dependency)
canary next
# Output: CBIN-104 - Must be completed before CBIN-105

# After CBIN-104 is TESTED, system selects CBIN-105
canary next
# Output: CBIN-105 - Now unblocked!
```

## Examples

### Example 1: First Implementation

**Scenario:** Starting fresh after `canary init`

```bash
$ canary next --prompt
```

**Output:**

```markdown
# Implementation Guidance: CBIN-101 - ScannerCore

## Priority Information
- Requirement ID: CBIN-101
- Priority: 1 (Highest)
- Status: STUB
- Aspect: Engine

## Specification

[Full spec.md content loaded here...]

## Constitutional Principles

From .canary/memory/constitution.md:

**Article IV: Test-First Imperative**
All features SHALL be implemented using test-first development...

**Article V: Simplicity and Anti-Abstraction**
Prefer simple, direct solutions over complex abstractions...

## Implementation Guidance

### Step 1: Write Tests (RED phase)
Create test file at: internal/scanner/scanner_test.go

```go
// CANARY: REQ=CBIN-101; FEATURE="ScannerCore"; ASPECT=Engine; STATUS=STUB; TEST=TestCANARY_CBIN_101_Engine_BasicScan; UPDATED=2025-10-17
func TestCANARY_CBIN_101_Engine_BasicScan(t *testing.T) {
    // Test implementation...
}
```

[... continued with full implementation guidance ...]
```

### Example 2: No Work Available

**Scenario:** All requirements completed

```bash
$ canary next
```

**Output:**
```
🎉 All requirements completed! No work available.

Suggestions:
  • Run: canary scan --verify GAP_ANALYSIS.md
  • Review completed requirements
  • Consider creating new specifications

Congratulations on completing the project roadmap!
```

### Example 3: Dependency Blocking

**Scenario:** High priority requirement is blocked

```bash
$ canary next --dry-run
```

**Output:**
```
Next priority (dry run): CBIN-104 - TokenParser
Priority: 3 | Status: STUB | Aspect: Engine
Location: .canary/specs/CBIN-104-token-parser/spec.md

Note: CBIN-105 (PRIORITY=1) is blocked by this requirement.
Completing CBIN-104 will unblock CBIN-105.
```

### Example 4: Database Fallback

**Scenario:** Database not yet created

```bash
$ canary next
```

**Output:**
```
ℹ️  Database not found, scanning filesystem...

📌 Next Priority: CBIN-101 - ScannerCore

[... rest of output ...]

💡 Tip: Run 'canary index' to improve performance
```

## Best Practices

### For AI Agents

1. **Run after every completion** - Use `canary next` after marking each requirement TESTED
2. **Always use --prompt** - Get full context for correct implementation
3. **Follow test-first guidance** - Respect RED → GREEN → REFACTOR workflow
4. **Update tokens immediately** - Change STATUS as you progress through phases
5. **Verify dependencies** - Check that DEPENDS_ON requirements are TESTED

### For Human Developers

1. **Morning routine** - Start day with `canary next` to see priorities
2. **Review before commit** - Use dry-run mode to preview next work
3. **Filter by skill** - Use `--aspect` to match your expertise
4. **Track progress** - Watch priority numbers decrease as work completes
5. **Update priorities** - Use `canary prioritize` to adjust as needs change

### For CI/CD Systems

1. **Use --json mode** - Machine-readable output for automation
2. **Check exit codes** - 0 = success/no work, non-zero = error
3. **Batch processing** - Implement multiple requirements in sequence
4. **Checkpoint tracking** - Create checkpoints after each completion
5. **Fail gracefully** - Handle "no work available" as success

## Priority Determination

The system uses a multi-factor algorithm to determine priority:

### Factor 1: Explicit PRIORITY Field

```
PRIORITY=1  (Highest priority)
PRIORITY=2
PRIORITY=3
PRIORITY=4
PRIORITY=5  (Default)
...
PRIORITY=10 (Lowest priority)
```

Lower numbers selected first.

### Factor 2: STATUS Value

Among requirements with same PRIORITY:

1. **STUB** - Not yet implemented (highest urgency)
2. **IMPL** - Implemented but needs tests
3. **TESTED** - Skip (already complete)
4. **BENCHED** - Skip (already complete)

### Factor 3: Dependencies (DEPENDS_ON)

Requirements with unmet dependencies are automatically skipped. The system selects blocking dependencies first.

Example:
```
CBIN-105: DEPENDS_ON=CBIN-104
CBIN-104: STATUS=STUB
```

Result: CBIN-104 selected (must complete before CBIN-105)

### Factor 4: Age (UPDATED Field)

Among requirements with same PRIORITY and STATUS, older tokens (older UPDATED dates) get slight priority boost to prevent stale work from accumulating.

## Troubleshooting

### Problem: "No tokens found"

**Symptoms:**
```
Error: no tokens found in database or filesystem
```

**Solutions:**
1. Run `canary index` to build database
2. Check that CANARY tokens exist: `grep -r "CANARY:" .`
3. Verify `.canaryignore` isn't excluding all files
4. Check you're in project root directory

### Problem: "All requirements completed" but work remains

**Symptoms:**
```
🎉 All requirements completed!
```
But you know there are STUB requirements.

**Solutions:**
1. Check if hidden requirements exist: `canary list --include-hidden --status STUB`
2. Verify status filters: `canary list --status STUB` (see all STUB items)
3. Check if requirements are in ignored directories (templates, tests, etc.)
4. Rebuild database: `canary index`

### Problem: Dependency Loop Detected

**Symptoms:**
```
Error: circular dependency detected: CBIN-105 → CBIN-106 → CBIN-105
```

**Solutions:**
1. Review DEPENDS_ON fields in both requirements
2. Remove circular dependency (one requirement must not depend on the other)
3. Update CANARY tokens to reflect correct dependency order
4. Run `canary index` to refresh dependency graph

### Problem: Prompt Generation Fails

**Symptoms:**
```
Error: failed to generate prompt: template execution failed at line 42
```

**Solutions:**
1. Check that specification file exists at referenced path
2. Verify constitution.md exists at `.canary/memory/constitution.md`
3. Ensure spec.md is valid markdown (no corrupted characters)
4. Run without --prompt flag to test selection logic independently

## FAQ

**Q: Can I manually override priority selection?**

A: Not directly. Use `canary prioritize <REQ-ID> <feature> <new-priority>` to adjust priorities, then run `canary next` again.

**Q: Does the command modify any files?**

A: No. `canary next` is read-only. It identifies and generates prompts but doesn't modify code or tokens.

**Q: How do I implement multiple requirements in parallel?**

A: Run `canary next --json` multiple times with different filters (e.g., by aspect) to get independent work items for parallel development.

**Q: What if I disagree with the selected priority?**

A: You can:
1. Adjust PRIORITY fields in CANARY tokens
2. Use `--status` or `--aspect` filters to narrow selection
3. Use `canary implement <specific-req-id>` to manually select a requirement

**Q: Does this work without a database?**

A: Yes! The command falls back to filesystem scanning if `.canary/canary.db` doesn't exist. Performance is slightly slower but functionality is identical.

**Q: How does this integrate with slash commands?**

A: The `/canary.next` slash command in AI agents automatically runs `canary next --prompt` and feeds the result directly to the agent as implementation guidance.

## Related Documentation

- [canary list](./list-command-guide.md) - Viewing all requirements
- [canary implement](./implement-command-guide.md) - Manual requirement selection
- [canary prioritize](./prioritize-command-guide.md) - Adjusting priorities
- [Constitutional Principles](../../.canary/memory/constitution.md) - Project governance

## Integration Examples

### Integration 1: GitHub Actions Workflow

```yaml
name: Auto-implement next requirement

on:
  schedule:
    - cron: '0 */6 * * *'  # Every 6 hours

jobs:
  auto-implement:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - name: Install canary
        run: go install ./cmd/canary
      - name: Get next requirement
        id: next
        run: |
          canary next --json > next.json
          echo "req_id=$(jq -r '.ReqID' next.json)" >> $GITHUB_OUTPUT
      - name: Implement with AI agent
        run: |
          # Call your AI agent API here
          # Pass requirement from steps.next.outputs.req_id
```

### Integration 2: Pre-commit Hook

```bash
#!/bin/bash
# .git/hooks/pre-commit

# Show next priority after commit
canary next --dry-run
```

### Integration 3: Claude Code Workflow

In Claude Code IDE:

```
User: /canary.next
Claude: [Reads generated prompt]
Claude: I'll implement CBIN-138 (MultilineTokens) following the specification...
[Claude implements, tests, and updates tokens]
User: continue
Claude: [Automatically runs /canary.next again for next task]
```

---

*Last verified: 2025-10-17 with canary v0.1.0*
*Implementation status: BENCHED (fully tested and benchmarked)*
