# CANARY: REQ=CBIN-117; FEATURE="AgentContextDoc"; ASPECT=Docs; STATUS=IMPL; OWNER=canary; UPDATED=2025-10-16
# CANARY Agent Context

**Last Updated:** 2025-10-16
**Version:** 1.0

## Project Overview

This project uses CANARY requirement tracking with spec-kit-inspired workflows.

## Available Commands

### Requirement Management
- `/canary.constitution` - Create/update project principles
- `/canary.specify` - Create new requirement specification
- `/canary.plan` - Generate implementation plan (includes gap analysis)
- `/canary.scan` - Scan for CANARY tokens
- `/canary.verify` - Verify GAP_ANALYSIS.md claims
- `/canary.update-stale` - Update stale tokens

### Gap Analysis & Learning
- `canary gap mark <req-id> <feature>` - Record implementation mistakes
- `canary gap query` - Search gap entries with filters
- `canary gap report <req-id>` - Generate gap analysis report
- `canary gap helpful <gap-id>` - Mark gap as useful
- `canary gap config` - View/update gap settings

### Development Workflow

1. **Establish Principles**: `/canary.constitution`
2. **Define Requirements**: `/canary.specify [feature description]`
3. **Plan Implementation**: `/canary.plan CBIN-XXX [tech stack]`
4. **Scan & Verify**: `/canary.scan` and `/canary.verify`
5. **Update Stale**: `/canary.update-stale` (as needed)

## CANARY Token Format

```
// CANARY: REQ=CBIN-###; FEATURE="Name"; ASPECT=API; STATUS=IMPL; [TEST=TestName]; [BENCH=BenchName]; [OWNER=team]; UPDATED=YYYY-MM-DD
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
canary create CBIN-105 "FeatureName" --aspect API --status IMPL
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
    └── CBIN-XXX-feature-name/   # Individual requirement specs
        ├── spec.md
        └── plan.md

GAP_ANALYSIS.md                   # Requirement tracking
status.json                       # Scanner output
status.csv                        # Scanner output (CSV)
.canary/canary.db                 # Gap analysis database
```

## Gap Analysis System (CBIN-140)

The gap analysis system helps agents learn from implementation mistakes.

### When to Record Gaps

Record a gap entry when:
- ❌ Implementation logic was incorrect
- ❌ Tests were inadequate or missing critical cases
- ❌ Performance issues were discovered
- ❌ Security vulnerabilities were found
- ❌ Edge cases were not handled
- ❌ Integration issues occurred

### Recording a Gap

```bash
canary gap mark <REQ-ID> <FEATURE> \
  --category <category> \
  --description "What went wrong" \
  --action "How it was fixed" \
  --aspect <ASPECT>
```

**Categories:**
- `logic_error` - Incorrect business logic or algorithm
- `test_failure` - Tests incorrectly written or missing cases
- `performance` - Performance issues or inefficient implementation
- `security` - Security vulnerabilities or insecure practices
- `edge_case` - Unhandled edge cases or boundary conditions
- `integration` - Integration issues with existing systems
- `documentation` - Incorrect or misleading documentation
- `other` - Other types of implementation gaps

**Example:**
```bash
canary gap mark CBIN-140 GapTracking \
  --category logic_error \
  --description "Query was missing ORDER BY clause causing non-deterministic results" \
  --action "Added ORDER BY helpful_count DESC, created_at DESC" \
  --aspect Storage
```

### Querying Gaps

```bash
# Query all gaps for a requirement
canary gap query --req-id CBIN-140

# Query by category
canary gap query --category security

# Query by aspect
canary gap query --aspect API

# Limit results
canary gap query --req-id CBIN-140 --limit 5
```

### Marking Gaps as Helpful

When a gap warning prevents a future mistake:
```bash
canary gap helpful GAP-CBIN-140-001
```

This increases the gap's priority in future planning prompts.

### Gap Injection into Planning

When running `/canary.plan CBIN-XXX`, the system automatically:
1. Queries top gaps for that requirement
2. Includes them in the "Past Implementation Gaps" section
3. Formats them with lessons learned

This ensures agents see historical mistakes before implementing.

### Gap Analysis Reports

```bash
# Generate report for a requirement
canary gap report CBIN-140

# View all gap categories
canary gap categories

# View/update configuration
canary gap config
canary gap config --max-gaps 20 --ranking weighted
```

### Integration with Workflow

**Standard development flow with gaps:**
1. Plan: `/canary.plan CBIN-XXX` (sees past gaps automatically)
2. Implement: Follow plan, avoiding recorded mistakes
3. Test: Verify implementation
4. Review: If issues found, record with `canary gap mark`
5. Rate: Mark helpful gaps with `canary gap helpful`

## Notes for AI Agents

- Reference `.canary/memory/constitution.md` before planning
- Use `/canary.specify` to create structured requirements
- Follow test-first approach (Article IV of constitution)
- Update CANARY tokens as implementation progresses
- Run `/canary.scan` after implementation to verify status
