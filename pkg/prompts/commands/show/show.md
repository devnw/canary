# Show Command Prompt

## Purpose
Display all CANARY tokens for a specific requirement ID.

## Task
Implement `canary show <REQ-ID>` to display detailed information about a requirement.

## Expected Behavior
```bash
canary show CBIN-001
```

## Output Format
```
Requirement: CBIN-001
Feature: UserAuthentication
Status: TESTED
Owner: backend

Implementations:
  API (TESTED)
    File: src/auth/api.go
    Test: TestCANARY_CBIN_001_UserAuth
    Updated: 2025-10-15

  CLI (IMPL)
    File: cli/auth/login.go
    Updated: 2025-10-10

Summary: 2 implementations (1 TESTED, 1 IMPL)
```

## Standards
- Query `pkg/storage` by requirement ID
- Group by feature name
- Show files, tests, benchmarks
- Clear hierarchy and formatting
