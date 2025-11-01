# Status Command Prompt

## Purpose
Show implementation progress for a requirement.

## Task
Implement `canary status <REQ-ID>` to show completion statistics.

## Expected Behavior
```bash
canary status CBIN-001
```

## Output Format
```
Requirement: CBIN-001
Progress: 75% (3/4 implementations)

By Status:
  TESTED:  2 (50%)
  IMPL:    1 (25%)
  STUB:    1 (25%)

By Aspect:
  API:     TESTED
  CLI:     IMPL
  Storage: TESTED
  Docs:    STUB

Last Updated: 2025-10-15
```

## Standards
- Calculate percentage based on STATUS progression
- BENCHED=100%, TESTED=75%, IMPL=50%, STUB=25%
- Show breakdown by status and aspect
- Color-coded progress bar
