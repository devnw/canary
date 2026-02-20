# Specify Command Prompt

## Purpose
Create a requirement specification from a feature description.

## Task
Implement `canary specify` to generate structured requirement documents.

## Expected Behavior
```bash
canary specify "Add user authentication with OAuth2 support"
```

## Output
- Creates `.canary/specs/CBIN-XXX-user-auth/spec.md`
- Uses template from `.canary/templates/spec-template.md`
- Auto-assigns next available requirement ID

## Specification Template Structure
```markdown
# CBIN-XXX: User Authentication

## Summary
Brief description of the requirement

## Motivation
Why this feature is needed

## Requirements
- Functional requirements (must-haves)
- Non-functional requirements (performance, security)

## Acceptance Criteria
- Testable conditions for completion

## Dependencies
- Other requirements this depends on

## Implementation Notes
- Technical considerations
- Suggested approach
```

## Standards
- Interactive prompts for specification fields
- Save to `.canary/specs/` directory
- Create directory structure automatically
- Update GAP_ANALYSIS.md with new requirement
