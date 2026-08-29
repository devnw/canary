# DB Command Prompt

## Purpose
Direct database management operations.

## Task
Implement `canary db` for low-level database operations.

## Expected Behavior
```bash
# Show database statistics
canary db stats

# Compact/vacuum database
canary db vacuum

# Export database to JSON
canary db export --format json > canary-export.json

# Import from JSON
canary db import canary-export.json

# Run SQL query (read-only)
canary db query "SELECT * FROM tokens WHERE status='STUB'"
```

## Output Format
```
Database Statistics:

Location: .canary/canary.db
Size: 2.4 MB
Created: 2025-09-01
Last Modified: 2025-11-01

Tables:
  tokens: 156 rows
  requirements: 42 rows
  checkpoints: 8 rows
  migrations: 5 rows

Indexes: 12 total
Fragmentation: 8.2%
```

## Operations
- **stats**: Database statistics
- **vacuum**: Compact and optimize
- **export**: Dump to JSON/CSV
- **import**: Load from JSON/CSV
- **query**: Execute read-only SQL
- **backup**: Create database backup
- **restore**: Restore from backup

## Standards
- Use `pkg/storage` database layer
- Read-only queries by default
- Validate imports before applying
- Atomic operations (backup before modify)
- Clear warnings for destructive operations
