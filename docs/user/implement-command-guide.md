# Implement Command User Guide

**Requirement:** CBIN-133
**Status:** Complete
**Last Updated:** 2025-10-17

## Overview

The `canary implement` command is your AI agent's gateway to implementing CANARY-tracked requirements. It intelligently locates requirements using fuzzy matching, loads complete specifications and plans, and generates comprehensive implementation prompts with all necessary context for safe, correct implementation.

**Key Benefits:**
- **Complete Context**: Loads spec, plan, and constitutional principles in one command
- **Fuzzy Search**: Find requirements by partial name or keywords
- **Test-First Guidance**: Automatically includes TDD workflow instructions
- **Token Placement**: Clear guidance on where and how to place CANARY tokens
- **100% Accuracy**: 42/42 features tested and working perfectly

## Getting Started

### Prerequisites

- CANARY CLI installed
- Project with `.canary/specs/` directory
- (Optional) `.canary/memory/constitution.md` for project principles

### Quick Start

Implement a requirement by ID:

```bash
canary implement CBIN-105
```

Or search by feature name:

```bash
canary implement "user auth"
```

## Usage

### Basic Usage: Implement by Exact ID

When you know the requirement ID:

```bash
canary implement CBIN-136
```

**Output includes:**
- Full specification (what to build)
- Implementation plan (how to build it)
- Constitutional principles (governing rules)
- Token placement examples
- Test-first workflow steps
- Success criteria checklist

### Fuzzy Search: Find by Name

When you remember the feature name but not the ID:

```bash
canary implement "documentation tracking"
```

Returns ranked matches:
```
Found 3 matching requirements:

  1. CBIN-136 - DocumentationTracking (92%)
  2. CBIN-115 - DocTemplate (45%)
  3. CBIN-107 - Constitution (12%)

Auto-selected: CBIN-136 (score > 80%)
```

The command automatically selects if there's a clear winner (>80% score).

### List Available Work

See what needs implementation:

```bash
canary implement --list
```

Shows all STUB and IMPL status requirements that need work.

## Common Workflows

### Workflow 1: Agent Implements Next Priority Feature

```bash
# 1. Find next priority requirement
canary list --status STUB --limit 5

# 2. Get implementation guidance
canary implement CBIN-138

# 3. Follow the generated prompt
#    - Read specification
#    - Review implementation plan
#    - Write tests first
#    - Implement features
#    - Place CANARY tokens
#    - Verify success criteria
```

### Workflow 2: Human Developer Picks Feature

```bash
# 1. List available work
canary implement --list

# 2. Search for interesting feature
canary implement "multiline tokens"

# 3. Review the generated prompt
# 4. Implement following test-first approach
```

### Workflow 3: Continue Partial Implementation

```bash
# Check current status
canary status CBIN-136

# Get implementation context again
canary implement CBIN-136

# Review what's done, continue with remaining tokens
```

## Examples

### Example 1: Implementing a New Feature

You want to implement CBIN-138 (Multiline Tokens):

```bash
canary implement CBIN-138
```

**What you get:**

1. **Specification Context:**
   - Purpose: Support multiline CANARY tokens
   - User stories and acceptance criteria
   - Success metrics

2. **Implementation Plan:**
   - Test file structure
   - Implementation phases (RED → GREEN → REFACTOR)
   - CANARY token placement locations
   - File paths and function signatures

3. **Constitutional Guidance:**
   - Article IV: Test-First Imperative
   - Article V: Simplicity principles
   - Token format requirements

4. **Actionable Checklist:**
   - [ ] Create test files
   - [ ] Write failing tests (RED)
   - [ ] Implement to pass tests (GREEN)
   - [ ] Place CANARY tokens
   - [ ] Update status to TESTED

### Example 2: Fuzzy Search When Unsure

You remember something about "authentication":

```bash
canary implement "auth"
```

Returns:
```
Found 2 requirements:
  1. CBIN-107 - UserAuthentication (95%)
  2. CBIN-109 - TokenAuth (78%)

Multiple matches found. Please specify:
  canary implement CBIN-107  # for User Authentication
  canary implement CBIN-109  # for Token Authentication
```

You select the correct one:

```bash
canary implement CBIN-107
```

### Example 3: Checking Available Work

```bash
canary implement --list
```

Shows:
```
Available Requirements:

STUB (not started):
  CBIN-138 - MultilineTokens (Engine)
  CBIN-140 - ValidationRules (Engine)

IMPL (needs tests):
  CBIN-105 - UserAuth (API)
  CBIN-200 - RestAPI (API)

Total: 4 requirements available for implementation
```

## Best Practices

### For AI Agents

1. **Always use `implement` before coding** - Get full context first
2. **Follow the test-first approach** - Write tests before implementation
3. **Place tokens as instructed** - Use exact format from guidance
4. **Verify success criteria** - Check all requirements before marking complete
5. **Update token status** - Change STATUS=STUB → IMPL → TESTED

### For Human Developers

1. **Use `--list` to explore** - See what's available before committing
2. **Search with keywords** - Fuzzy matching finds related requirements
3. **Read the full prompt** - Don't skip spec or constitutional guidance
4. **Ask for clarification** - Use [NEEDS CLARIFICATION] markers in specs
5. **Verify with `canary status`** - Check progress after implementation

## Troubleshooting

### Problem: "No requirements found"

**Symptoms:**
```
Error: no requirements found matching: "feature name"
```

**Solutions:**
1. Check specs exist: `ls .canary/specs/`
2. Try broader search: `canary implement "feature"`
3. List all: `canary list`
4. Check database: `canary index` (rebuild)

### Problem: Multiple Weak Matches

**Symptoms:**
```
Found 5 requirements all with scores 50-60%
```

**Solutions:**
1. Use more specific keywords: `"multiline token parser"` instead of `"token"`
2. Check exact ID: `canary list | grep CBIN`
3. Use exact ID: `canary implement CBIN-XXX`

## FAQ

**Q: Do I need a plan before implementing?**

A: Recommended but not required. The command works with spec-only. However, plans provide file structure guidance, test-first workflow steps, and token placement locations.

**Q: Can I implement multiple requirements at once?**

A: No. The command focuses on single-requirement implementation for clarity and traceability.

**Q: Does this update the database or modify files?**

A: No. This is a read-only command. It loads and displays context but doesn't modify anything.

## Related Documentation

- [canary list](./list-command-guide.md) - Finding requirements
- [canary status](./status-command-guide.md) - Checking implementation progress
- [CANARY Token Format](../../.canary/docs/token-format.md) - Token syntax

---

*Last verified: 2025-10-17 with canary v0.1.0*
*Implementation status: 42/42 features tested (100%)*
