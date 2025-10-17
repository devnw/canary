# Specification Modification User Guide

**Requirement:** CBIN-134
**Status:** Complete
**Last Updated:** 2025-10-17

## Overview

The `canary specify update` command allows you to efficiently locate and modify existing CANARY requirement specifications without loading excessive context. This is especially useful for AI agents and developers who need to make targeted changes to specifications without scanning through all spec files.

**Key Benefits:**
- Find specs by exact ID in under 1 second
- Filter to specific sections to reduce context usage by 50-80%
- Automatic detection of related plan.md files
- Fuzzy search support for when you don't know the exact ID

## Getting Started

### Prerequisites

- CANARY CLI installed and in your PATH
- A project with `.canary/specs/` directory containing requirement specifications
- (Optional) `.canary/canary.db` database for faster lookups

### Installation / Setup

No additional setup required! The command is available immediately:

```bash
canary specify update --help
```

## Usage

### Basic Usage

Find and display a specification by its exact requirement ID:

```bash
canary specify update CBIN-134
```

Expected output:
```
✅ Found specification: .canary/specs/CBIN-134-spec-modification/spec.md
📋 Plan exists: .canary/specs/CBIN-134-spec-modification/plan.md

--- Spec Content ---

# Feature Specification: Specification Modification Command
...
```

### Common Tasks

####  Task 1: Load Only Specific Sections

Reduce context usage by loading only the sections you need:

```bash
canary specify update CBIN-134 --sections overview,requirements
```

This returns only the Overview and Requirements sections, saving 50-80% of tokens.

#### Task 2: Find Spec by Keyword Search

When you don't know the exact ID, use fuzzy search:

```bash
canary specify update --search "authentication"
```

Returns ranked matches:
```
Found 3 matching specs:

  1. CBIN-107 - UserAuthentication (95%)
  2. CBIN-109 - TokenAuth (78%)
  3. CBIN-112 - OAuth Integration (65%)
```

#### Task 3: Quickly Check if Spec Exists

```bash
canary specify update CBIN-999 2>&1 | head -5
```

If the spec doesn't exist, you'll get a helpful error suggesting the search command.

## Examples

### Example 1: Updating Success Criteria

You need to update the success criteria for CBIN-105 without loading the entire 500-line spec:

```bash
canary specify update CBIN-105 --sections "success criteria"
```

**Explanation:**
- Loads only the success criteria section
- Saves ~400 lines of context
- Returns the exact section you need to modify
- Preserves other sections unchanged

### Example 2: Finding the Right Spec to Modify

You remember the spec is about "fuzzy matching" but don't know the ID:

```bash
canary specify update --search "fuzzy match"
```

Returns:
```
Found 2 matching specs:

  1. CBIN-133 - FuzzyMatching (92%)
  2. CBIN-134 - SpecModification (45%)

Auto-selected: CBIN-133
```

The command automatically selects CBIN-133 since it has >90% confidence.

### Example 3: Checking for Related Plans

```bash
canary specify update CBIN-136
```

Output shows both spec and plan:
```
✅ Found specification: .canary/specs/CBIN-136-documentation-tracking/spec.md
📋 Plan exists: .canary/specs/CBIN-136-documentation-tracking/plan.md
```

This reminds you to update the plan if your spec changes affect implementation.

## Configuration

### Command Flags

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--search` | boolean | false | Use fuzzy search instead of exact ID lookup |
| `--sections` | string[] | [] | Comma-separated list of sections to load |

### Section Names

Valid section names (case-insensitive, partial matching):
- overview
- requirements / functional requirements
- success criteria
- user stories
- assumptions
- constraints
- dependencies
- risks

## Best Practices

- **Use exact IDs when you know them** - Faster than search (< 1 second vs ~2 seconds)
- **Load specific sections for large specs** - Use `--sections` to reduce context by 50-80%
- **Check for plan.md** - If shown, remember to update it when spec changes affect implementation
- **Use search for exploration** - When you're not sure which spec to modify, start with `--search`

## Troubleshooting

### Problem: "Spec not found" Error

**Symptoms:**
```
Error: spec not found for CBIN-999
Try: canary specify update --search "CBIN-999"
```

**Solution:**
1. Verify the requirement ID exists: `canary list | grep CBIN-999`
2. Check if specs are in `.canary/specs/` directory: `ls .canary/specs/`
3. Try fuzzy search: `canary specify update --search "feature keywords"`

### Problem: Multiple Matches from Search

**Symptoms:**
```
Found 5 matching specs:
  1. CBIN-101 - Feature (85%)
  2. CBIN-102 - Similar Feature (84%)
  ...

Error: multiple matches - use exact REQ-ID instead
```

**Solution:**
1. Review the list and identify the correct requirement ID
2. Run again with exact ID: `canary specify update CBIN-101`

### Problem: Section Not Found

**Symptoms:**
```
Error: no matching sections found for: [incorret-name]
```

**Solution:**
1. Check available sections: Look at the section headers (## headings) in the spec
2. Use partial matching: `--sections overview` matches "## Overview"
3. Case doesn't matter: `--sections REQUIREMENTS` works fine

## FAQ

**Q: Can I modify multiple specs at once?**

A: No, the command is designed for focused, single-spec modifications to minimize context usage. For bulk changes, use a script that calls the command multiple times.

**Q: Does this update the database?**

A: No, this is a read-only command. It locates and displays specs but doesn't modify files or the database. Use your editor or other commands to make actual changes.

**Q: What's the difference between this and `/canary.specify`?**

A: `/canary.specify` creates new specifications from scratch. `canary specify update` is for modifying existing ones. Use update when you know a spec exists and want to change it.

**Q: Can I use wildcards in section names?**

A: Not explicitly, but the command uses fuzzy matching. `--sections req` will match "Requirements", "Functional Requirements", etc.

## Related Documentation

- [canary specify create](../commands/specify.md) - Creating new specifications
- [canary list](../commands/list.md) - Finding requirements to update
- [CANARY Specification Template](../../.canary/templates/spec-template.md) - Spec structure reference
- [Implementation Plans](./plan-guide.md) - When to update plan.md vs spec.md

## Support

If you encounter issues:
1. Check the troubleshooting section above
2. Verify your `.canary/specs/` directory structure follows conventions
3. Try rebuilding the database: `canary index`
4. Report issues at https://github.com/spyder/canary/issues

---

*Last verified: 2025-10-17 with canary v0.1.0*
