<!-- CANARY: REQ=CBIN-117; FEATURE="AgentContextDoc"; ASPECT=Docs; STATUS=IMPL; OWNER=canary; UPDATED=2025-10-16 -->
# CANARY Agent Context

**Last Updated:** 2025-10-16
**Version:** 1.0

## Project Overview

This project uses CANARY requirement tracking with spec-kit-inspired workflows.

## Available Commands

### Core CLI Commands

#### Requirement Management
```bash
# Create or update constitutional principles
canary constitution [description]

# Create new requirement specification
canary specify <feature-description> --aspect <aspect>
canary specify update <REQ-ID> [--sections <sections>]
canary specify update --search "query"

# Generate implementation plan
canary plan <REQ-ID> [tech-stack]

# Create CANARY token template
canary create <req-id> <feature-name> --aspect <aspect> --status <status>
```

#### Scanning & Analysis
```bash
# Scan codebase for CANARY tokens
canary scan --root . --out status.json --csv status.csv
canary scan --verify GAP_ANALYSIS.md --strict
canary scan --update-stale --project-only

# Show tokens for a requirement
canary show <REQ-ID> [--group-by aspect|status] [--json]

# List all tokens with filtering
canary list [--status STUB|IMPL|TESTED|BENCHED]
canary list [--aspect API|CLI|Engine|Storage|Security|Docs]
canary list [--phase Phase0|Phase1|Phase2|Phase3]
canary list [--include-hidden] [--limit N]

# Search and grep tokens
canary search <keywords> [--json]
canary grep <pattern> [--group-by none|requirement]

# Show implementation status
canary status <REQ-ID> [--no-color]

# List files for a requirement
canary files <REQ-ID> [--all]
```

#### Implementation Workflow
```bash
# Get next highest priority requirement
canary next [--aspect <aspect>] [--status <status>] [--prompt]

# Generate implementation guidance
canary implement <query> [--list] [--prompt]

# Update token priority
canary prioritize <REQ-ID> <feature> <priority>
```

#### Database Management
```bash
# Index/rebuild token database
canary index [--root .] [--db .canary/canary.db]

# Run migrations
canary migrate <steps|all>
canary rollback <steps|all>

# Create checkpoint snapshot
canary checkpoint <name> [description]
```

#### Documentation Management
```bash
# Create documentation from template
canary doc create <REQ-ID> --type user|technical|feature|api|architecture --output <path>

# Check documentation status
canary doc status [REQ-ID] [--all] [--stale-only]

# Update documentation hashes
canary doc update [REQ-ID] [--all] [--stale-only]

# Generate documentation report
canary doc report [--format text|json] [--show-undocumented]
```

#### Gap Analysis & Learning
```bash
# Mark implementation gap
canary gap mark <req-id> <feature> --category <category> --description <desc> [--action <action>]

# Query gaps
canary gap query [--req-id <id>] [--category <cat>] [--feature <name>]

# Generate gap report
canary gap report <req-id>

# Mark gap helpful/unhelpful
canary gap helpful <gap-id>
canary gap unhelpful <gap-id>

# View/update gap configuration
canary gap config [--max-gaps N] [--min-helpful N] [--ranking strategy]

# List gap categories
canary gap categories
```

#### Project Initialization
```bash
# Initialize new project
canary init [project-name] --key <PREFIX>
canary init --agents claude,cursor,windsurf
canary init --all-agents

# Migrate from existing system
canary migrate-from spec-kit|legacy-canary [directory] [--dry-run] [--force]

# Detect system type
canary detect [directory]
```

### Slash Command Workflow

1. **Establish Principles**: `/canary.constitution [description]`
2. **Define Requirements**: `/canary.specify [feature description]`
3. **Plan Implementation**: `/canary.plan <REQ-ID> [tech stack]`
4. **Implement**: `/canary.implement <query>`
5. **Next Task**: `/canary.next`
6. **Scan & Verify**: `/canary.scan` and `/canary.verify`
7. **Track Progress**: `/canary.status <REQ-ID>` and `/canary.show <REQ-ID>`
8. **Update Stale**: `/canary.update-stale`

## CANARY Token Format

```
// CANARY: REQ={{.ReqID}}-<ASPECT>-###; FEATURE="Name"; ASPECT=API; STATUS=IMPL; [TEST=TestName]; [BENCH=BenchName]; [OWNER=team]; UPDATED=YYYY-MM-DD
```

## Status Progression

- **STUB**: Planned but not implemented
- **IMPL**: Implemented (token placed in code)
- **TESTED**: Implemented with tests (auto-promoted when TEST= field added)
- **BENCHED**: Tested with benchmarks (auto-promoted when BENCH= field added)

## Valid Aspects

API, CLI, Engine, Storage, Security, Docs, Wire, Planner, Decode, Encode, RoundTrip, Bench, FrontEnd, Dist

## Constitutional Principles

1. **Requirement-First**: Every feature starts with a CANARY token
2. **Test-First**: Tests written before implementation
3. **Evidence-Based**: Status promoted based on TEST=/BENCH= fields
4. **Simplicity**: Minimal complexity, standard library preferred
5. **Documentation Currency**: Tokens kept current with UPDATED field

## Quick Reference

### Common Workflows

**Complete Token Lifecycle:**
```bash
# 1. Create specification
canary specify "User authentication system" --aspect Security

# 2. Generate implementation plan
canary plan CBIN-Security-001 "Go stdlib + JWT"

# 3. Get implementation guidance
canary implement CBIN-Security-001 --prompt

# 4. During implementation: create tokens
canary create CBIN-Security-001 "AuthHandler" --aspect API --status IMPL --test TestAuthHandler

# 5. Check progress
canary status CBIN-Security-001
canary show CBIN-Security-001

# 6. Scan and verify
canary scan --root . --out status.json
canary scan --verify GAP_ANALYSIS.md --strict

# 7. Update stale tokens
canary scan --update-stale

# 8. Track gaps (if issues found)
canary gap mark CBIN-Security-001 AuthHandler \
  --category logic_error \
  --description "Missing token expiry validation" \
  --action "Added expiry check in middleware"
```

**Database Operations:**
```bash
# Index tokens into database
canary index --root . --db .canary/canary.db

# Create progress checkpoint
canary checkpoint "post-auth-implementation" "Completed auth module"

# Run database migrations
canary migrate all
```

**Finding Work:**
```bash
# Get next priority item
canary next --prompt

# List all unimplemented requirements
canary list --status STUB --limit 10

# Search for specific features
canary search "authentication user"
canary grep "Auth" --group-by requirement

# Find requirements needing tests
canary list --status IMPL
```

**Documentation Tracking:**
```bash
# Create documentation
canary doc create CBIN-Security-001 --type technical --output docs/auth.md

# Check for stale docs
canary doc status --all --stale-only

# Update doc hashes after editing
canary doc update CBIN-Security-001

# Generate documentation report
canary doc report --show-undocumented
```

**Gap Analysis & Learning:**
```bash
# Mark gaps during development
canary gap mark <req-id> <feature> \
  --category logic_error|test_failure|performance|security|edge_case \
  --description "what went wrong" \
  --action "how it was fixed"

# Query gaps for learning
canary gap query --category logic_error --limit 10

# View gaps for a requirement
canary gap report CBIN-Security-001

# Configure gap injection
canary gap config --max-gaps 15 --min-helpful 2 --ranking weighted
```

## Project Structure

```
.canary/
├── memory/
│   └── constitution.md          # Project principles
├── scripts/
│   └── create-new-requirement.sh # Automation scripts
├── templates/
│   ├── commands/                # Slash command definitions
│   ├── spec-template.md         # Requirement spec template
│   └── plan-template.md         # Implementation plan template
└── specs/
    └── {{.ReqID}}-<ASPECT>-XXX-feature-name/   # Individual requirement specs
        ├── spec.md
        └── plan.md

GAP_ANALYSIS.md                   # Requirement tracking
status.json                       # Scanner output
status.csv                        # Scanner output (CSV)
```

## Notes for AI Agents

- Reference `.canary/memory/constitution.md` before planning
- Use `/canary.specify` to create structured requirements
- Follow test-first approach (Article IV of constitution)
- Update CANARY tokens as implementation progresses
- Run `/canary.scan` after implementation to verify status
- Use `canary gap mark` to record mistakes for future learning
- Check `canary next --prompt` for highest priority work
- Use `canary doc` commands to track documentation currency

## Detailed CLI Reference

### Gap Categories
- `logic_error` - Incorrect business logic or algorithm
- `test_failure` - Tests incorrectly written or missing cases
- `performance` - Performance issues or inefficient implementation
- `security` - Security vulnerabilities or insecure practices
- `edge_case` - Unhandled edge cases or boundary conditions
- `integration` - Integration issues with existing systems
- `documentation` - Incorrect or misleading documentation
- `other` - Other types of implementation gaps

### Documentation Types
- `user` - User-facing documentation
- `technical` - Technical design documentation
- `feature` - Feature specification documentation
- `api` - API reference documentation
- `architecture` - Architecture decision records (ADR)

### Status Values
- `STUB` - Planned but not implemented
- `IMPL` - Implemented (code exists)
- `TESTED` - Implemented with tests (TEST= field present)
- `BENCHED` - Tested with benchmarks (BENCH= field present)

### Priority Levels
- `1` - Highest priority (critical/blocking)
- `2-3` - High priority (important features)
- `4-6` - Medium priority (normal features)
- `7-9` - Low priority (nice-to-have)
- `10` - Lowest priority (future/deferred)

### Common Flags Reference

**Database flags (most commands):**
- `--db <path>` - Database file path (default: `.canary/canary.db`)

**Filtering flags:**
- `--aspect <aspect>` - Filter by aspect
- `--status <status>` - Filter by status
- `--phase <phase>` - Filter by phase
- `--owner <owner>` - Filter by owner
- `--spec-status <status>` - Filter by spec status

**Output flags:**
- `--json` - Output as JSON
- `--no-color` - Disable colored output
- `--format <format>` - Output format (text|json)

**Scan flags:**
- `--root <dir>` - Root directory (default: `.`)
- `--out <file>` - Output JSON path (default: `status.json`)
- `--csv <file>` - Output CSV path
- `--strict` - Enforce 30-day staleness check
- `--update-stale` - Update UPDATED field for stale tokens
- `--verify <file>` - Verify GAP_ANALYSIS.md claims
- `--skip <regex>` - Skip paths matching regex
- `--project-only` - Filter by project requirement ID pattern

**List flags:**
- `--limit <n>` - Maximum number of results
- `--order-by <clause>` - Custom ORDER BY clause
- `--include-hidden` - Include test files, templates, examples

**Create flags:**
- `--aspect <aspect>` - Requirement aspect/category
- `--status <status>` - Implementation status
- `--test <name>` - Test function name
- `--bench <name>` - Benchmark function name
- `--owner <owner>` - Team/person responsible
