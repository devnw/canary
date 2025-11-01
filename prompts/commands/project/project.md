# Project Command Prompt

## Purpose
Manage project-level CANARY configuration.

## Task
Implement `canary project` to configure project settings.

## Expected Behavior
```bash
# Initialize project configuration
canary project init

# Show project configuration
canary project show

# Update project settings
canary project set key "CBIN"
canary project set aspect-enums "API,CLI,Storage,Security"
```

## Configuration File
Stored in `.canary/project.yaml`:

```yaml
name: myproject
description: My project with CANARY tracking
key: CBIN  # Requirement ID prefix

metadata:
  owner: backend-team
  repository: github.com/org/repo

enums:
  aspects:
    - API
    - CLI
    - Storage
    - Security
  statuses:
    - STUB
    - IMPL
    - TESTED
    - BENCHED

settings:
  stale_days: 30
  default_owner: backend
  require_tests: true
```

## Commands
- `init`: Create project configuration
- `show`: Display current configuration
- `set`: Update configuration value
- `validate`: Check configuration validity

## Standards
- Store in `.canary/project.yaml`
- Validate enum values against config
- Use in other commands for defaults
- Support project-specific customization
- Document all configuration options
