# CANARY Development - AI Agent Guide
<!-- CANARY: REQ=CBIN-CLI-001; FEATURE="AgentInstructions"; ASPECT=Docs; STATUS=TESTED; UPDATED=2025-10-16 -->

**Context File for AI Coding Agents**

This project uses CANARY requirement tracking with spec-kit-inspired workflows.

## Available Slash Commands

See [.canary/AGENT_CONTEXT.md](./.canary/AGENT_CONTEXT.md) for detailed information.

### Workflow Commands

- **/canary.constitution** - Create or update project governing principles
- **/canary.specify** - Create a new requirement specification from feature description
- **/canary.plan** - Generate technical implementation plan for a requirement
- **/canary.scan** - Scan codebase for CANARY tokens and generate reports
- **/canary.verify** - Verify GAP_ANALYSIS.md claims against actual implementation
- **/canary.update-stale** - Auto-update UPDATED field for stale tokens (>30 days)

### Command Definitions

All slash commands are defined in:
- `.canary/templates/commands/constitution.md`
- `.canary/templates/commands/specify.md`
- `.canary/templates/commands/plan.md`
- `.canary/templates/commands/scan.md`
- `.canary/templates/commands/verify.md`
- `.canary/templates/commands/update-stale.md`

## Quick Start Workflow

1. **Establish Principles**: `/canary.constitution Create principles for code quality and testing`
2. **Define Requirement**: `/canary.specify Add user authentication with OAuth2 support`
3. **Create Plan**: `/canary.plan CBIN-001 Use Go standard library with bcrypt`
4. **Scan & Verify**: `/canary.scan` then `/canary.verify`
5. **Update Stale**: `/canary.update-stale` (as needed)

## CANARY Token Format

```
// CANARY: REQ=CBIN-###; FEATURE="Name"; ASPECT=API; STATUS=IMPL; UPDATED=YYYY-MM-DD
```

**Status Progression:**
- STUB → IMPL → TESTED → BENCHED

**Valid Aspects:**
API, CLI, Engine, Storage, Security, Docs, Wire, Planner, Decode, Encode, RoundTrip, Bench, FrontEnd, Dist

## Constitutional Principles

See [.canary/memory/constitution.md](./.canary/memory/constitution.md) for full details.

**Core Principles:**
1. **Requirement-First**: Every feature starts with a CANARY token
2. **Test-First**: Tests written before implementation (Article IV)
3. **Evidence-Based**: Status promoted based on TEST=/BENCH= fields
4. **Simplicity**: Minimal complexity, prefer standard library
5. **Documentation Currency**: Keep tokens current with UPDATED field

## CLI Commands

```bash
# Query tokens for a requirement (USE THESE INSTEAD OF GREP/SQLITE3!)
canary show CBIN-133              # Display all tokens
canary files CBIN-133             # List implementation files
canary status CBIN-133            # Show progress summary

# Initialize new project
canary init my-project

# Create requirement token
canary create CBIN-105 "FeatureName" --aspect API --status IMPL

# Scan for tokens
canary scan --root . --out status.json --csv status.csv

# Verify claims
canary scan --root . --verify GAP_ANALYSIS.md --strict

# Update stale tokens
canary scan --root . --update-stale
```

## Project Structure

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
status.json                       # Scanner output
```

## For AI Agents

**Before implementing:**
1. Reference `.canary/memory/constitution.md`
2. Use `/canary.specify` to create structured requirements
3. Follow test-first approach (Article IV)

**During implementation:**
1. Update CANARY tokens as code evolves
2. **Use `canary show/files/status` commands to query tokens** (NOT grep or sqlite3!)
   - ❌ DON'T: `grep -rn "CANARY.*CBIN-133" --include="*.go" .`
   - ✅ DO: `canary show CBIN-133`
   - ❌ DON'T: `sqlite3 .canary/canary.db "SELECT * FROM tokens..."`
   - ✅ DO: `canary files CBIN-133` or `canary status CBIN-133`

**After implementing:**
1. Run `/canary.scan` to verify status
2. Run `/canary.verify` to confirm claims

**Key Files:**
- [.canary/AGENT_CONTEXT.md](./.canary/AGENT_CONTEXT.md) - Complete context for AI agents
- [.canary/memory/constitution.md](./.canary/memory/constitution.md) - Constitutional principles
- [GAP_ANALYSIS.md](./GAP_ANALYSIS.md) - Requirement tracking
