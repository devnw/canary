# List Command Prompt

## Purpose
List CANARY tokens with optional filtering.

## Task
Implement `canary list` to query and display tokens from the database.

## Expected Behavior
```bash
# List all tokens
canary list

# Filter by status
canary list --status IMPL

# Filter by aspect
canary list --aspect API

# Filter by owner
canary list --owner backend

# Limit results
canary list --limit 10

# Combine filters
canary list --status TESTED --aspect CLI --limit 5
```

## Output Format
- Tabular display with columns: REQ, FEATURE, ASPECT, STATUS, OWNER, UPDATED
- Color-coded status (green=TESTED/BENCHED, yellow=IMPL, red=STUB)
- Total count at bottom

## Standards
- Query `internal/storage` database
- Support multiple filters simultaneously
- Sort by REQ ID ascending
- Handle empty results gracefully
