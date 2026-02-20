# Prioritize Command Prompt

## Purpose
Set priority level for a requirement.

## Task
Implement `canary prioritize <REQ-ID> <priority>` to manage requirement priorities.

## Expected Behavior
```bash
# Set priority
canary prioritize CBIN-001 HIGH

# Valid priorities: LOW, MEDIUM, HIGH, CRITICAL
canary prioritize CBIN-042 CRITICAL
```

## Output Format
```
Updated priority for CBIN-001
  Old: MEDIUM
  New: HIGH

Affected:
  - 4 CANARY tokens updated
  - Priority queue reordered

Current high-priority requirements:
  1. CBIN-042 (CRITICAL) - Security fix
  2. CBIN-001 (HIGH) - User authentication
  3. CBIN-015 (HIGH) - Data validation
```

## Priority Levels
- **CRITICAL**: Security, data loss, blocking issues
- **HIGH**: Important features, significant bugs
- **MEDIUM**: Standard features, improvements
- **LOW**: Nice-to-haves, minor enhancements

## Standards
- Update all tokens for the requirement
- Store priority in database
- Validate priority enum
- Show impact on priority queue
- Affect `canary next` command output
