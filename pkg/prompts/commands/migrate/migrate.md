# Migrate Command Prompt

## Purpose
Migrate between CANARY database schema versions.

## Task
Implement `canary migrate` to handle database schema evolution.

## Expected Behavior
```bash
# Check current schema version
canary migrate version

# Migrate to latest schema
canary migrate up

# Rollback one migration
canary migrate down

# Show migration status
canary migrate status
```

## Output Format
```
Current schema: v3
Latest schema: v5

Pending migrations:
  [ ] v4: Add priority field to requirements
  [ ] v5: Add dependency tracking

Run: canary migrate up

---

Running migrations...
  ? v4: Add priority field (0.1s)
  ? v5: Add dependency tracking (0.2s)

Schema updated: v3 ? v5
```

## Migration Management
- Track applied migrations in database
- Support up (apply) and down (rollback)
- Validate schema integrity
- Backup database before migrations
- Atomic migrations (all-or-nothing)

## Standards
- Store migrations in `internal/storage/migrations/`
- Use SQL migration files
- Version format: `YYYYMMDD_description.sql`
- Test rollbacks
- Document schema changes
