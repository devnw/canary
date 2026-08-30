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
│   └── <PROJECT_KEY>-XXX-feature/
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

### specs/<PROJECT_KEY>-XXX-feature/
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
canary create <PROJECT_KEY>-XXX "FeatureName"
```

### Planning Implementation

**Read `.canary/commands/plan.md` first**, then:

```bash
# Use slash command
/canary.plan <PROJECT_KEY>-XXX

# Creates plan.md with architecture and TDD phases
```

### Implementing Features

**Read `.canary/commands/implement.md` first**, then:

```bash
# Use slash command
/canary.implement <PROJECT_KEY>-XXX

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
// CANARY: REQ=<PROJECT_KEY>-###; FEATURE="Name"; ASPECT=API; STATUS=TESTED; TEST=TestName; UPDATED=YYYY-MM-DD
```

**Status Progression:**
- STUB → IMPL → TESTED → BENCHED

**Evidence Required:**
- TESTED: Must have TEST=TestName field
- BENCHED: Must have BENCH=BenchName field

See `.canary/commands/specify.md` for complete token format details.

## Available Commands

**All commands are documented in `.canary/commands/` - read those files for workflows:**

- `/canary.view` → `commands/view.md` - Full picture of a requirement in one call (status, files, tests, deps, spec, ticket); e.g. `canary view <REQ-ID>`
- `/canary.deps` → `commands/deps.md` - Check/graph/reverse/validate requirement dependencies (`canary deps graph <REQ-ID> --format mermaid`)
- `/canary.specify` → `commands/specify.md` - Create new requirement specification
- `/canary.plan` → `commands/plan.md` - Generate implementation plan
- `/canary.implement` → `commands/implement.md` - Get implementation guidance
- `/canary.scan` → `commands/scan.md` - Scan for tokens and generate reports
- `/canary.verify` → `commands/verify.md` - Verify GAP_ANALYSIS.md claims
- `/canary.show` → `commands/show.md` - Display all tokens for a requirement
- `/canary.status` → `commands/status.md` - Show implementation progress
- `/canary.files` → `commands/files.md` - List files containing tokens
- `/canary.grep` → `commands/grep.md` - Search tokens by keyword
- `/canary.search` → `commands/search.md` - Search tokens by keyword across all fields
- `/canary.gap` → `commands/gap.md` - Record and query implementation-gap learnings
- `/canary.index` → `commands/index.md` - Rebuild the token database
- `/canary.onboard` → `commands/onboard.md` - Fresh-codebase adoption analysis (languages, entry points, MIGRATE notes)
- `/canary.upgrade` → `commands/upgrade.md` - Rewrite legacy on-disk token shapes (dry run by default)
- `/canary.drift` → `commands/drift.md` - Detect code-vs-token drift (`canary drift --strict` for CI)
- `/canary.ticket` → `commands/ticket.md` - Reconcile tokens against configured ticket sources (e.g. JIRA)

**For complete workflows, examples, and validation criteria, read the command file.**

## Ticket Sources

External ticket sources (JIRA, etc.) are configured in `.canary/project.yaml`'s `sources:` list. `canary ticket sync` computes a reconciliation plan against them, and without credentials it stays plan-only and never touches the network; `canary ticket status [--refresh]` reports or refreshes the cached remote status on its own. A `destination: true` source (or the first non-flatfile source by default) is where new issues are created. External dependencies (ticket-sourced or `peers:`-owned) block `canary next` unless they resolve *satisfied*: an unsatisfied one blocks, and so does one whose state cannot be resolved from disk at all. Pass `canary next --allow-unknown-external` to accept an unresolvable dependency deliberately. `canary deps validate` treats both as informational until `--strict-external`.

## Related Files

- `commands/*.md` - Complete command workflows (PRIMARY REFERENCE)
- `memory/constitution.md` - Governing principles
- `templates/spec-template.md` - Specification template
- `templates/plan-template.md` - Implementation plan template
- `AGENT_CONTEXT.md` - Complete agent reference
