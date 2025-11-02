# Plan Command Prompt

## Purpose
Generate technical implementation plan for a requirement.

## Task
Implement `canary plan <REQ-ID>` to create detailed implementation plans.

## Expected Behavior
```bash
canary plan CBIN-001 "Use Go standard library with bcrypt for password hashing"
```

## Output
- Creates `.canary/specs/CBIN-001-feature/plan.md`
- Uses template from `.canary/templates/plan-template.md`
- Links to specification file

## Plan Template Structure
```markdown
# Implementation Plan: CBIN-001

## Overview
Technical approach summary

## Architecture
- Components to create/modify
- Data structures
- API design

## Implementation Phases
1. Phase 1: Setup and scaffolding
2. Phase 2: Core implementation
3. Phase 3: Testing and validation
4. Phase 4: Documentation

## Testing Strategy
- Unit tests
- Integration tests
- Performance benchmarks

## Risks & Mitigations
- Potential issues
- How to address them

## Timeline
- Estimated effort
- Milestones
```

## Standards
- Read existing spec.md for context
- Generate phase breakdown
- Suggest test strategy
- Consider constitutional principles from `.canary/memory/constitution.md`
