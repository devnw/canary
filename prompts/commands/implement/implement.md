# Implement Command Prompt

## Purpose
Get implementation guidance for a requirement.

## Task
Implement `canary implement <REQ-ID>` to show next implementation steps.

## Expected Behavior
```bash
canary implement CBIN-001
```

## Output Format
```
Requirement: CBIN-001 - User Authentication
Current Phase: PHASE-2 (Implementation)

? Completed:
  - Specification created
  - Implementation plan generated
  - Scaffolding in place

?? Current Focus:
  - Implement password hashing (src/auth/hash.go)
  - Add login endpoint (src/api/auth.go)

?? Next Steps:
  1. Review plan: .canary/specs/CBIN-001-user-auth/plan.md
  2. Add CANARY tokens to implementation files
  3. Write tests: TestCANARY_CBIN_001_*
  4. Update token STATUS to TESTED

Files to work on:
  - src/auth/hash.go (create)
  - src/api/auth.go (modify)
  - tests/auth_test.go (create)
```

## Implementation Phases
- PHASE-1: Specification and planning
- PHASE-2: Core implementation
- PHASE-3: Testing and benchmarking
- PHASE-4: Documentation and completion

## Standards
- Check for spec.md and plan.md files
- Scan existing tokens to determine phase
- Provide actionable next steps
- Link to relevant files and documentation
