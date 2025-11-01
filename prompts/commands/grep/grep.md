# Grep Command Prompt

## Purpose
Search tokens by pattern in specific fields.

## Task
Implement `canary grep` for pattern matching in token fields.

## Expected Behavior
```bash
# Search in feature names (regex)
canary grep --feature "Auth.*"

# Search in owner field
canary grep --owner "backend"

# Search in file paths
canary grep --file "src/api"

# Combine patterns
canary grep --feature "User" --status IMPL
```

## Search Fields
- `--feature`: Feature name pattern
- `--owner`: Owner pattern
- `--file`: File path pattern
- `--test`: Test name pattern
- `--bench`: Benchmark name pattern
- `--status`: Exact status match
- `--aspect`: Exact aspect match

## Output Format
```
Matches for --feature "Auth.*":

CBIN-001: UserAuthentication
  Feature: UserAuthentication
  Files: src/auth/api.go, src/auth/session.go
  Status: TESTED
  Owner: backend

CBIN-015: OAuth2Integration  
  Feature: OAuth2Integration
  Files: src/auth/oauth.go
  Status: IMPL
  Owner: backend

Total: 2 requirements matched
```

## Standards
- Support regex patterns (Go regexp syntax)
- Case-insensitive by default
- Highlight matching portions
- Combine multiple filters with AND logic
- Show full token details for matches
