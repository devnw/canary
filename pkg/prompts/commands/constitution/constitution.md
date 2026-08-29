# Constitution Command Prompt

## Purpose
Create or update project governing principles.

## Task
Implement `canary constitution` to manage project development principles.

## Expected Behavior
```bash
# Create/update constitution interactively
canary constitution

# Create with template
canary constitution --template strict

# Show current constitution
canary constitution show
```

## Constitution Structure
Stored in `.canary/memory/constitution.md`:

```markdown
# Project Constitution

## Article I: Requirement-First Development
Every feature must start with a CANARY requirement token.

## Article II: Status Progression
Status must progress: STUB -> IMPL -> TESTED -> BENCHED

## Article III: Test-First Development  
Tests must be written before promoting STATUS to TESTED.

## Article IV: Evidence-Based Promotion
Status promotion requires evidence:
- IMPL: Implementation exists
- TESTED: TEST= field present
- BENCHED: BENCH= field present

## Article V: Documentation Currency
UPDATED field must be current (within 30 days for TESTED/BENCHED).
```

## Templates
- **strict**: Test-first, evidence-required
- **balanced**: Pragmatic balance of speed and quality
- **flexible**: Minimal constraints

## Standards
- Interactive prompts for principles
- Store in `.canary/memory/constitution.md`
- Reference in other commands
- Enforce via scan --strict mode
- Show in help text and documentation
