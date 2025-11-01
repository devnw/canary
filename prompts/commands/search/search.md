# Search Command Prompt

## Purpose
Search CANARY tokens by keywords.

## Task
Implement `canary search <query>` for full-text search across tokens.

## Expected Behavior
```bash
# Search in feature names
canary search "authentication"

# Search in all fields
canary search "OAuth2"
```

## Search Fields
- Feature names
- Requirement IDs
- File paths
- Owner names
- Test names

## Output Format
- List matching tokens with highlighted search terms
- Show relevance score
- Sort by relevance descending

## Standards
- Case-insensitive search
- Support partial matches
- Highlight matching terms in output
- Limit to top 50 results by default
