# Next Command Prompt

## Purpose
Identify next highest priority unimplemented requirement.

## Task
Implement `canary next` to find the next task to work on.

## Expected Behavior
```bash
canary next
```

## Output Format
```
Next Priority: CBIN-042

Requirement: CBIN-042
Feature: DataValidation
Status: STUB
Priority: HIGH
Owner: backend

Files to implement:
  - src/validation/rules.go (API)
  - cli/validate/cmd.go (CLI)

Suggested workflow:
  1. canary specify CBIN-042
  2. canary plan CBIN-042
  3. Implement and add CANARY tokens
```

## Priority Rules
1. High priority requirements first
2. STUB status before IMPL
3. Requirements with specs/plans ready
4. Dependencies satisfied first

## Standards
- Check priority field in tokens
- Consider dependencies between requirements
- Show actionable next steps
- Link to spec/plan files if they exist
