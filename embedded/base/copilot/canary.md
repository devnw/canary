<!-- CANARY: REQ=CBIN-148; FEATURE="InstructionTemplates"; ASPECT=Docs; STATUS=TESTED; TEST=TestCopilotInstructionTemplateValidity; UPDATED=2025-10-19 -->

# CANARY Directory Guidelines

You are working in the `.canary/` directory - the heart of the CANARY requirement tracking system.

## Directory Structure

```
.canary/
├── specs/                  # Requirement specifications (WHAT/WHY)
│   └── {{.ProjectKey}}-XXX-feature/
│       ├── spec.md        # Requirement specification
│       └── plan.md        # Technical implementation plan (HOW)
├── templates/              # Templates for specs, plans, commands
├── memory/                 # Project context and principles
│   └── constitution.md    # Governing principles
├── scripts/                # Automation scripts
└── agents/                 # AI agent configurations
```

## Key Files

### constitution.md
Project governing principles. Review before implementing features.

**Core Principles:**
- Article I: Requirement-First Development
- Article IV: Test-First Imperative (non-negotiable)
- Article V: Simplicity and Anti-Abstraction
- Article VI: Integration-First Testing

### specs/{{.ProjectKey}}-XXX-feature/
Each requirement has its own directory containing:
- **spec.md** - WHAT users need and WHY (technology-agnostic)
- **plan.md** - HOW to implement (technical details)

## Working with CANARY

### Creating New Requirements

```bash
# Use slash command
/canary.specify "feature description"

# Or use CLI
canary create {{.ProjectKey}}-XXX "FeatureName"
```

### Planning Implementation

```bash
# Use slash command
/canary.plan {{.ProjectKey}}-XXX

# Creates plan.md with architecture and TDD phases
```

### Implementing Features

```bash
# Use slash command
/canary.implement {{.ProjectKey}}-XXX

# Follow test-first approach from plan
```

### Scanning Progress

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

## Related Commands

- `/canary.specify` - Create new requirement specification
- `/canary.plan` - Generate implementation plan
- `/canary.implement` - Get implementation guidance
- `/canary.scan` - Scan for tokens and generate reports
- `/canary.verify` - Verify GAP_ANALYSIS.md claims
