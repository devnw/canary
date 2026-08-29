# Index Command Prompt

## Purpose
Index codebase tokens into the database.

## Task
Implement `canary index` to populate/update the SQLite database with scanned tokens.

## Expected Behavior
```bash
# Index all tokens
canary index

# Index specific directory
canary index --root src/

# Force reindex
canary index --force
```

## Process
1. Scan codebase for CANARY tokens (similar to `scan`)
2. Parse and validate token fields
3. Insert/update tokens in `.canary/canary.db`
4. Update summary statistics
5. Report indexing results

## Output Format
```
Indexing codebase...
  Scanned: 1,234 files
  Found: 89 CANARY tokens
  Inserted: 12 new tokens
  Updated: 3 existing tokens
  Skipped: 74 unchanged tokens

Summary:
  Total requirements: 42
  By status: STUB=15, IMPL=18, TESTED=8, BENCHED=1
```

## Standards
- Use `pkg/storage` for database operations
- Perform upsert (insert or update)
- Track last indexed timestamp
- Show progress for large codebases
- Atomic transaction for consistency
