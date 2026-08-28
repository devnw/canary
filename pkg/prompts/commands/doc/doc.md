# Doc Command Prompt

## Purpose
Generate documentation from CANARY tokens.

## Task
Implement `canary doc` to auto-generate requirement documentation.

## Expected Behavior
```bash
# Generate full documentation
canary doc generate

# Generate for specific requirement
canary doc generate CBIN-001

# Update existing documentation
canary doc update
```

## Output Formats

### Markdown (default)
```markdown
# Requirements Documentation

## CBIN-001: User Authentication
**Status**: TESTED  
**Owner**: backend  
**Last Updated**: 2025-10-15

### Implementation
- API authentication (TESTED)
- CLI login command (IMPL)

### Tests
- TestCANARY_CBIN_001_Login
- TestCANARY_CBIN_001_Session

### Files
- src/auth/api.go
- cli/auth/login.go
```

### HTML
Generate browsable HTML documentation with:
- Navigation sidebar
- Status indicators
- Search functionality
- Dependency graphs

## Generated Files
- `docs/requirements.md` - Overview
- `docs/requirements/CBIN-XXX.md` - Individual requirement pages
- `docs/requirements.html` - HTML version
- `docs/status-report.md` - Current status report

## Standards
- Auto-generate from database
- Link to source files
- Include status and progress
- Show test coverage
- Update automatically in CI
