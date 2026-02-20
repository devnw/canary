# Bug Command Prompt

## Purpose
Manage bug tracking with CANARY tokens.

## Task
Implement `canary bug` subcommands for bug lifecycle management.

## Subcommands

### bug list
```bash
# List all bugs
canary bug list

# Filter by status
canary bug list --status OPEN

# Filter by severity
canary bug list --severity HIGH
```

### bug create
```bash
canary bug create "Memory leak in auth module" --severity HIGH --owner backend
```

### bug show
```bash
canary bug show BUG-001
```

### bug update
```bash
canary bug update BUG-001 --status FIXED --resolution "Fixed memory leak"
```

## Bug Token Format
```
// BUG: ID=BUG-###; TITLE="Description"; SEVERITY=HIGH; STATUS=OPEN; OWNER=team; CREATED=YYYY-MM-DD; UPDATED=YYYY-MM-DD
```

## Bug Status Progression
- **OPEN**: New bug reported
- **IN_PROGRESS**: Being worked on
- **FIXED**: Fix implemented
- **VERIFIED**: Fix tested and verified
- **CLOSED**: Completed

## Severity Levels
- **CRITICAL**: System down, data loss
- **HIGH**: Major functionality broken
- **MEDIUM**: Feature partially working
- **LOW**: Minor issue, cosmetic

## Standards
- Separate BUG-XXX numbering from REQ-XXX
- Track bug lifecycle with status updates
- Link bugs to requirements when relevant
- Store in same database as requirements
