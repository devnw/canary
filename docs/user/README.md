# CANARY User Documentation

**Last Updated:** 2025-10-17

## Overview

This directory contains comprehensive user guides for the CANARY requirement tracking system. These guides cover installation, core concepts, command usage, and best practices for both AI agents and human developers.

## Documentation Index

### Getting Started

**[Getting Started Guide](./getting-started.md)** - Start here!
- Installation and initialization
- Core concepts (tokens, status progression, aspects)
- Complete workflow walkthrough
- Best practices and common pitfalls
- Quick reference for all commands

**Recommended for:** New users, onboarding, tutorial walkthroughs

---

### Core Workflow Commands

**[Query Commands Guide](./query-commands-guide.md)**
- `canary show` - Display all tokens for a requirement
- `canary files` - List implementation files
- `canary status` - Check progress summary
- `canary grep` - Search tokens by pattern

**Recommended for:** Daily development, progress inspection, token discovery

---

**[Next Priority Guide](./next-priority-guide.md)**
- `canary next` - Automated workflow assistant
- Priority determination algorithm
- Dependency-aware selection
- Prompt generation for AI agents

**Recommended for:** AI agents, automated workflows, sprint planning

---

**[Implement Command Guide](./implement-command-guide.md)**
- `canary implement` - Get implementation guidance
- Fuzzy matching for requirement selection
- Constitution-aware prompts
- Test-first workflow integration

**Recommended for:** Feature implementation, AI agent guidance

---

**[Specification Modification Guide](./spec-modification-guide.md)**
- `canary specify` - Create new requirements
- `canary specify update` - Modify existing specs
- Interactive specification wizard
- Fuzzy matching for finding requirements

**Recommended for:** Requirement creation, specification updates

---

### Maintenance and Tracking

**[Documentation Tracking Guide](./documentation-tracking-guide.md)**
- `canary doc status` - Check documentation currency
- `canary doc update` - Update documentation hashes
- `canary doc report` - Generate coverage reports
- DOC= and DOC_HASH= field usage

**Recommended for:** Documentation maintenance, ensuring docs stay current

---

## Quick Command Reference

### Essential Commands

```bash
# Initialize new project
canary init my-project

# Build token database
canary index

# Get next priority work
canary next --prompt

# Inspect requirement
canary show CBIN-105
canary files CBIN-105
canary status CBIN-105

# Search for features
canary grep Authentication
canary list --status STUB --aspect API

# Get implementation guidance
canary implement CBIN-105
canary implement fuzzy  # Fuzzy match

# Create/modify specifications
canary specify
canary specify update CBIN-105

# Verify implementation
canary scan --verify GAP_ANALYSIS.md

# Check documentation
canary doc report --show-undocumented
canary doc status CBIN-105 FeatureName
```

## Documentation by User Type

### For AI Agents

**Primary Guides:**
1. [Getting Started Guide](./getting-started.md) - Core concepts
2. [Next Priority Guide](./next-priority-guide.md) - Automated workflow
3. [Implement Command Guide](./implement-command-guide.md) - Implementation guidance

**Key Slash Commands:**
- `/canary.next` - Get next priority requirement
- `/canary.show` - Display requirement tokens
- `/canary.status` - Check implementation progress

**Workflow:**
```
/canary.next → Implement with test-first → Update tokens → /canary.next
```

---

### For Human Developers

**Primary Guides:**
1. [Getting Started Guide](./getting-started.md) - Installation and basics
2. [Query Commands Guide](./query-commands-guide.md) - Daily inspection commands
3. [Specification Modification Guide](./spec-modification-guide.md) - Creating requirements

**Daily Commands:**
```bash
canary next              # Morning routine: see what's next
canary show CBIN-XXX     # Inspect requirement details
canary files CBIN-XXX    # Find implementation files
canary status CBIN-XXX   # Check progress
canary scan --verify     # Verify claims before commit
```

---

### For Project Maintainers

**Primary Guides:**
1. [Documentation Tracking Guide](./documentation-tracking-guide.md) - Maintaining docs
2. [Next Priority Guide](./next-priority-guide.md) - Priority management
3. [Getting Started Guide](./getting-started.md) - Onboarding reference

**Maintenance Commands:**
```bash
canary doc report --show-undocumented  # Find undocumented features
canary doc update --all                # Update all doc hashes
canary scan --update-stale             # Update stale UPDATED fields
canary list --status TESTED            # Find completed work
```

---

## Documentation Coverage

As of 2025-10-17:

- **Total Requirements**: 91
- **Documented Requirements**: 8
- **Coverage**: 8.8%

**Documented Features:**
1. CBIN-136 - Documentation Tracking (DOC_CURRENT)
2. CBIN-133 - Implement Command (DOC_CURRENT)
3. CBIN-132 - Next Priority Command (DOC_CURRENT)
4. CBIN-CLI-001 - Query Commands (DOC_CURRENT)
5. CBIN-134 - Specification Modification (DOC_CURRENT)
6. CBIN-131 - Fuzzy List Filtering (DOC_CURRENT)
7. CBIN-125 - List Command (DOC_CURRENT)
8. CBIN-124 - Index Command (DOC_CURRENT)

**Documentation Types:**
- User Guides: 7 (including this index)
- Architecture Docs: 1

## Best Practices

### Writing Documentation

1. **Start with overview** - Explain purpose and benefits
2. **Provide quick start** - Show simplest usage first
3. **Include examples** - Real-world scenarios with output
4. **Cover edge cases** - Troubleshooting section
5. **Reference related docs** - Link to other guides
6. **Add CANARY tokens** - Track documentation with DOC= field

### Maintaining Documentation

1. **Update with code changes** - Change docs when behavior changes
2. **Use documentation tracking** - Add DOC= and DOC_HASH= fields
3. **Check currency regularly** - Run `canary doc report`
4. **Update hashes after edits** - Run `canary doc update`
5. **Review quarterly** - Ensure examples still work

### For AI Agents

1. **Reference constitution** - Check `.canary/memory/constitution.md` first
2. **Follow test-first** - Write tests before implementation
3. **Update tokens immediately** - Change STATUS as you progress
4. **Add documentation** - Document all TESTED features
5. **Verify regularly** - Run `/canary.scan` to check token placement

## Contributing

When adding new user guides:

1. Follow the structure of existing guides (Overview → Usage → Examples → Troubleshooting)
2. Add the guide to this README index
3. Update the "Documentation Coverage" section
4. Add DOC= field to relevant CANARY tokens
5. Run `canary doc update` to calculate hash
6. Verify with `canary doc status`

## Related Documentation

- [Project README](../../README.md) - Project overview
- [CLAUDE.md](../../CLAUDE.md) - AI agent context
- [AGENT_CONTEXT.md](../../.canary/AGENT_CONTEXT.md) - Complete agent reference
- [Constitution](../../.canary/memory/constitution.md) - Project principles
- [GAP_ANALYSIS.md](../../GAP_ANALYSIS.md) - Requirement tracking

---

**Questions or Issues?**

- Check command help: `canary <command> --help`
- Review troubleshooting sections in individual guides
- Check [GitHub Issues](https://github.com/yourusername/canary/issues)

---

*Last verified: 2025-10-17 with canary v0.1.0*
