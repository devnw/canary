# Deps Command Prompt

## Purpose
Show and manage requirement dependencies.

## Task
Implement `canary deps` to track and visualize requirement dependencies.

## Expected Behavior
```bash
# Show dependencies for a requirement
canary deps show CBIN-001

# List all dependencies
canary deps list

# Add dependency
canary deps add CBIN-042 --depends-on CBIN-001

# Check for circular dependencies
canary deps check
```

## Output Format
```
Dependencies for CBIN-001 (User Authentication):

Depends On:
  ? CBIN-005 (Database Schema)     [TESTED]
  ? CBIN-010 (Encryption Library)  [BENCHED]

Required By:
  ? CBIN-015 (OAuth2 Integration)  [IMPL]
  ? CBIN-020 (Session Management)  [STUB]

Dependency Status: ? All dependencies satisfied

Implementation Order:
  1. CBIN-001 (ready - dependencies met)
  2. CBIN-015 (blocked - waiting for CBIN-001)
  3. CBIN-020 (blocked - waiting for CBIN-001)
```

## Dependency Graph
- Track "depends-on" relationships
- Detect circular dependencies
- Calculate implementation order
- Show blocking/blocked requirements
- Visualize dependency tree

## Standards
- Store dependencies in database
- Validate no circular dependencies
- Consider in `canary next` command
- Show status of dependencies
- Warn about implementing without dependencies
