<!-- CANARY: REQ=CBIN-117; FEATURE="AgentContextDoc"; ASPECT=Docs; STATUS=IMPL; OWNER=canary; UPDATED=2025-10-16 -->
# CANARY Agent Context

**Last Updated:** 2025-10-16
**Version:** 1.0

## Project Overview

This project uses CANARY requirement tracking with spec-kit-inspired workflows.

## Available Commands

### Requirement Management
- `/canary.constitution` - Create/update project principles
- `/canary.specify` - Create new requirement specification
- `/canary.plan` - Generate implementation plan
- `/canary.scan` - Scan for CANARY tokens
- `/canary.verify` - Verify GAP_ANALYSIS.md claims
- `/canary.update-stale` - Update stale tokens

### Development Workflow

1. **Establish Principles**: `/canary.constitution`
2. **Define Requirements**: `/canary.specify [feature description]`
3. **Plan Implementation**: `/canary.plan {{.ReqID}}-SECURITY_REVIEW-XXX [tech stack]`
4. **Scan & Verify**: `/canary.scan` and `/canary.verify`
5. **Update Stale**: `/canary.update-stale` (as needed)

## CANARY Token Format

```
// CANARY: REQ={{.ReqID}}-SECURITY_REVIEW-###; FEATURE="Name"; ASPECT=API; STATUS=IMPL; [TEST=TestName]; [BENCH=BenchName]; [OWNER=team]; UPDATED=YYYY-MM-DD
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

**Scan for tokens:**
```bash
canary scan --root . --out status.json --csv status.csv
```

**Verify claims:**
```bash
canary scan --root . --verify GAP_ANALYSIS.md --strict
```

**Update stale:**
```bash
canary scan --root . --update-stale
```

**Create token:**
```bash
canary create {{.ProjectKey}}-105 "FeatureName" --aspect API --status IMPL
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
    └── {{.ReqID}}-SECURITY_REVIEW-XXX-feature-name/   # Individual requirement specs
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
