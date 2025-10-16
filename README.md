# canary

Scan, update, create, verify and manage **CANARY** tokens across
repositories, emit `status.json` / `status.csv`, and **verify** GAP claims.

**Now with full spec-kit integration!** Initialize projects with `.canary/` workflow structure including slash commands, constitutional principles, and AI agent integration.

## Build

```bash
go build -o ./bin/canary ./cmd/canary

# Install system-wide (optional)
sudo cp ./bin/canary /usr/local/bin/

# The binary is self-contained with embedded templates
# No additional files needed for installation
```

## CLI Commands

Canary provides spec-kit-inspired commands for managing CANARY tokens.

**Quick Reference:**
```bash
canary init <project>      # Initialize with full workflow
canary constitution        # Create/view project principles
canary specify "feature"   # Create requirement spec
canary plan CBIN-XXX       # Generate implementation plan
canary implement CBIN-XXX  # Show implementation locations (reduces context!)
canary create CBIN-XXX     # Generate CANARY token
canary scan               # Scan for tokens and generate reports

# Advanced: Structured Storage & Priority Management
canary index              # Build SQLite database from tokens
canary list --status STUB  # List tokens with filtering
canary search "keyword"    # Search by keywords
canary prioritize CBIN-XXX Feature 1  # Set priority (1=highest)
canary checkpoint "name"   # Create state snapshot
```

**Key Features:**
- `canary implement` shows exact file:line locations, reducing agent context by ~95%
- `canary index` + `list/search` enable priority-driven development with SQLite storage

See [CLI_COMMANDS.md](./CLI_COMMANDS.md) for complete agent reference documentation.

### Initialize a New Project

```bash
canary init <project-name>
# Creates:
# - .canary/ directory with full workflow structure
# - .canary/memory/constitution.md - Project governing principles
# - .canary/templates/commands/ - Slash commands for AI agents
# - .canary/templates/ - Spec and plan templates
# - .canary/scripts/ - Automation scripts
# - README_CANARY.md - Token format specification
# - GAP_ANALYSIS.md - Requirements tracking template
# - CLAUDE.md - AI agent integration guide
```

### Create a New Requirement Token

```bash
canary create CBIN-105 "FeatureName" --aspect API --status IMPL --owner team
# Outputs a properly formatted CANARY token ready to paste
```

### Scan for Tokens

```bash
canary scan --root . --out status.json --csv status.csv
canary scan --root . --verify GAP_ANALYSIS.md --strict
canary scan --root . --update-stale  # Auto-update stale TESTED/BENCHED tokens
```

### Exit Codes

- **Exit 0**: OK
- **Exit 2**: Verification/staleness failed
- **Exit 3**: Parse or IO error

### Legacy Usage

The standalone scanner is still available at `tools/canary`:

```bash
go run ./tools/canary --root . --out status.json
```

**Token format**

```text
Example template (replace with actual values):
CANARY: REQ=CBIN-101; FEATURE="MyFeature"; ASPECT=API; STATUS=IMPL; TEST=TestCANARY_CBIN_101_API_MyFeature; BENCH=BenchmarkCANARY_CBIN_101_API_MyFeature; OWNER=team; UPDATED=2025-10-15

Valid ASPECT values: API, CLI, Engine, Planner, Storage, Wire, Security, Docs, Encode, Decode, RoundTrip, Bench, FrontEnd, Dist
Valid STATUS values: MISSING, STUB, IMPL, TESTED, BENCHED, REMOVED
```

**Supported comment styles**: `//`, `#`, `--`, `<!--` (Python, Go, Bash, SQL, Markdown, etc.)

## Status Auto-Promotion

The scanner auto-promotes statuses based on evidence references:

| From        | Evidence Condition    | To      |
| ----------- | --------------------- | ------- |
| IMPL        | ≥1 test (TEST=)       | TESTED  |
| IMPL/TESTED | ≥1 benchmark (BENCH=) | BENCHED |

Notes:

- Promotion is applied in-memory; original source comments remain unchanged.
- BENCHED dominates TESTED in summary counts.
- `--strict` still validates staleness on TESTED/BENCHED after promotion.
- A future `--no-promote` flag may allow raw status reporting.

Example: if a feature is marked `STATUS=IMPL` and has a `TEST=TestCANARY_REQ_GQL_030_TxnCommit`, the report will show it as `TESTED`.

## Testing

```bash
cd tools/canary
go test -v
```

## CANARY at a glance

Policy excerpt (see `docs/CANARY_POLICY.md`). Example tokens:

`CANARY: REQ=CBIN-101; FEATURE="ScannerCore"; ASPECT=Engine; STATUS=TESTED; TEST=TestCANARY_CBIN_101_Engine_ScanBasic; BENCH=BenchmarkCANARY_CBIN_101_Engine_Scan; OWNER=canary; UPDATED=2025-09-20`

`CANARY: REQ=CBIN-102; FEATURE="VerifyGate"; ASPECT=CLI; STATUS=TESTED; TEST=TestCANARY_CBIN_102_CLI_Verify; BENCH=BenchmarkCANARY_CBIN_102_CLI_Verify; OWNER=canary; UPDATED=2025-09-20`

## Structured Storage & Priority Management

Canary now includes SQLite-based structured storage for advanced token management:

### Index and Query Tokens

```bash
# Build/rebuild database from codebase
canary index --root . --db .canary/canary.db

# List tokens with filtering and priority ordering
canary list --status STUB --limit 10
canary list --phase Phase1 --owner backend
canary list --order-by "priority ASC, updated_at DESC"

# Search by keywords
canary search "authentication"
canary search "oauth jwt"

# Update priorities (1=highest, 10=lowest)
canary prioritize CBIN-001 JWTGeneration 1
```

### Extended Metadata

Tokens can now include:
- **PRIORITY**: 1-10 (affects ordering)
- **PHASE**: Phase0, Phase1, Phase2, Phase3
- **KEYWORDS**: Comma-separated tags for search
- **SPEC_STATUS**: draft, approved, in-progress, completed, archived
- **DEPENDS_ON**: Comma-separated requirement IDs
- **BLOCKS**: Requirement IDs this blocks
- **RELATED_TO**: Related requirement IDs
- **Git Integration**: Automatic commit hash and branch tracking

### Checkpoints

Create state snapshots for tracking progress:

```bash
canary checkpoint "phase1-complete" "All Phase 1 features implemented"
canary checkpoint "v1.0.0" "Release 1.0.0 snapshot"
```

Checkpoints capture:
- Token counts by status (STUB, IMPL, TESTED, BENCHED)
- Git commit hash and timestamp
- Full JSON snapshot of all tokens

## Spec-Kit Integration

Canary includes a full spec-kit-inspired workflow for requirement-driven development:

### AI Agent Integration

After running `canary init`, AI agents can use slash commands:

- `/canary.constitution` - Create/update project governing principles
- `/canary.specify` - Create requirement specification from feature description
- `/canary.plan` - Generate technical implementation plan for a requirement
- `/canary.scan` - Scan codebase for CANARY tokens and generate reports
- `/canary.verify` - Verify GAP_ANALYSIS.md claims against actual implementation
- `/canary.update-stale` - Auto-update UPDATED field for stale tokens (>30 days)

### Constitutional Governance

Projects initialized with `canary init` include a constitution (`.canary/memory/constitution.md`) with 9 articles:

1. **Requirement-First Development** - Every feature starts with a CANARY token
2. **Specification Discipline** - Focus on WHAT before HOW
3. **Token-Driven Planning** - Trackable, verifiable units of work
4. **Test-First Imperative** - Non-negotiable TDD approach
5. **Simplicity and Anti-Abstraction** - Minimal complexity, prefer standard library
6. **Integration-First Testing** - Real environments over mocks
7. **Documentation Currency** - CANARY tokens ARE the documentation
8. **Continuous Improvement** - Metrics-driven development
9. **Amendment Process** - How to evolve the constitution

### Workflow Example

```bash
# Initialize project with full workflow
canary init my-project
cd my-project

# Use AI agent slash commands (in Claude Code, Cursor, etc.)
# /canary.constitution
# /canary.specify "Add user authentication with OAuth2"
# /canary.plan CBIN-001 "Use Go standard library"

# Find implementation points (reduces context!)
canary implement CBIN-001 --status STUB
# Shows exact locations of unimplemented features

# Get context for specific feature
canary implement CBIN-001 --feature JWTGeneration --context
# Shows file:line with surrounding code

# Track progress
canary implement CBIN-001
# Shows: Progress: 67% (2/3)

# Scan and verify
canary scan --root . --out status.json --csv status.csv
canary scan --root . --verify GAP_ANALYSIS.md --strict
```

### Project Structure

```
.canary/
├── memory/
│   └── constitution.md          # Project principles
├── scripts/
│   └── create-new-requirement.sh # Automation
├── templates/
│   ├── commands/                # Slash command definitions
│   ├── spec-template.md         # Requirement template
│   └── plan-template.md         # Implementation plan template
└── specs/
    └── CBIN-XXX-feature/        # Individual requirements
        ├── spec.md
        └── plan.md

GAP_ANALYSIS.md                   # Requirement tracking
CLAUDE.md                         # AI agent integration guide
README_CANARY.md                  # Token specification
status.json                       # Scanner output
```
