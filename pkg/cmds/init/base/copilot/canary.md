<!-- CANARY: REQ=CBIN-148; FEATURE="InstructionTemplates"; ASPECT=Docs; STATUS=BENCHED; TEST=TestCopilotInstructionTemplateValidity; BENCH=BenchmarkCreateCopilotInstructions; UPDATED=2025-10-19 -->

# CANARY Directory Guidelines

You are working in the `.canary/` directory - the heart of the CANARY requirement tracking system.

## Directory Structure

```
.canary/
├── commands/                # Slash command definitions (READ THESE!)
│   ├── specify.md
│   ├── plan.md
│   ├── implement.md
│   ├── scan.md
│   └── ... (all commands)
├── specs/                  # Requirement specifications (WHAT/WHY)
│   └── {{.ProjectKey}}-XXX-feature/
│       ├── spec.md        # Requirement specification
│       └── plan.md        # Technical implementation plan (HOW)
├── templates/              # Templates for specs, plans, commands
├── memory/                 # Project context and principles
│   └── constitution.md    # Governing principles
└── scripts/                # Automation scripts
```

## Key Files

### constitution.md
Project governing principles. **Read this before implementing any feature.**

**Core Principles:**
- Article I: Requirement-First Development
- Article IV: Test-First Imperative (non-negotiable)
- Article V: Simplicity and Anti-Abstraction
- Article VI: Integration-First Testing

### commands/*.md
**Complete workflow definitions for all slash commands.** These files contain:
- Step-by-step execution workflows
- Example inputs and outputs
- Quality validation criteria
- Constitutional compliance checks

**Always read the command file before executing that command.**

### specs/{{.ProjectKey}}-XXX-feature/
Each requirement has its own directory containing:
- **spec.md** - WHAT users need and WHY (technology-agnostic)
- **plan.md** - HOW to implement (technical details)

## Working with CANARY

### Creating New Requirements

**Read `.canary/commands/specify.md` first**, then:

```bash
# Use slash command
/canary.specify "feature description"

# Or use CLI
canary create {{.ProjectKey}}-XXX "FeatureName"
```

### Planning Implementation

**Read `.canary/commands/plan.md` first**, then:

```bash
# Use slash command
/canary.plan {{.ProjectKey}}-XXX

# Creates plan.md with architecture and TDD phases
```

### Implementing Features

**Read `.canary/commands/implement.md` first**, then:

```bash
# Use slash command
/canary.implement {{.ProjectKey}}-XXX

# Follow test-first approach from plan
```

### Scanning Progress

**Read `.canary/commands/scan.md` first**, then:

```bash
# Use slash command
/canary.scan

# Or use CLI
canary scan --root . --out status.json
```

## Token Management

CANARY tokens track requirement status directly in source code.

**Token Format:**
```
// CANARY: REQ={{.ProjectKey}}-###; FEATURE="Name"; ASPECT=API; STATUS=TESTED; TEST=TestName; UPDATED=YYYY-MM-DD
```

**Status Progression:**
- STUB → IMPL → TESTED → BENCHED

**Evidence Required:**
- TESTED: Must have TEST=TestName field
- BENCHED: Must have BENCH=BenchName field

See `.canary/commands/specify.md` for complete token format details.

## Available Commands

**All commands are documented in `.canary/commands/` - read those files for workflows:**

- `/canary.specify` → `commands/specify.md` - Create new requirement specification
- `/canary.plan` → `commands/plan.md` - Generate implementation plan
- `/canary.implement` → `commands/implement.md` - Get implementation guidance
- `/canary.scan` → `commands/scan.md` - Scan for tokens and generate reports
- `/canary.verify` → `commands/verify.md` - Verify GAP_ANALYSIS.md claims
- `/canary.show` → `commands/show.md` - Display all tokens for a requirement
- `/canary.status` → `commands/status.md` - Show implementation progress
- `/canary.files` → `commands/files.md` - List files containing tokens
- `/canary.grep` → `commands/grep.md` - Search tokens by keyword

**For complete workflows, examples, and validation criteria, read the command file.**

## Related Files

- `commands/*.md` - Complete command workflows (PRIMARY REFERENCE)
- `memory/constitution.md` - Governing principles
- `templates/spec-template.md` - Specification template
- `templates/plan-template.md` - Implementation plan template
- `AGENT_CONTEXT.md` - Complete agent reference
