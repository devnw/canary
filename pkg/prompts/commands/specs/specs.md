# Specs Command Prompt

## Purpose
Manage requirement specifications.

## Task
Implement `canary specs` to list and manage specification files.

## Expected Behavior
```bash
# List all specifications
canary specs list

# Show specification status
canary specs status

# Find spec for requirement
canary specs show CBIN-001

# Validate specifications
canary specs validate
```

## Output Format
```
Requirement Specifications:

- CBIN-001: User Authentication
   Spec: .canary/specs/CBIN-001-user-auth/spec.md
   Plan: .canary/specs/CBIN-001-user-auth/plan.md
   Status: TESTED (4 tokens)

- CBIN-015: OAuth2 Integration
   Spec: .canary/specs/CBIN-015-oauth2/spec.md
   Plan: Missing
   Status: IMPL (2 tokens)

- CBIN-042: Data Validation
   Spec: Missing
   Plan: Missing
   Status: STUB (1 token)

Summary:
  Total: 42 requirements
  With specs: 38 (90%)
  With plans: 32 (76%)
  Complete (spec+plan): 30 (71%)
```

## Validation
- Check spec.md exists for all requirements
- Verify spec.md follows template structure
- Warn about requirements without specs
- Check plan.md exists for non-STUB requirements
- Link specs to actual implementations

## Standards
- Scan `.canary/specs/` directory
- Match specs to database requirements
- Validate markdown structure
- Report missing or incomplete specs
- Suggest which specs need attention
