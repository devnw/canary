# Database Migrations - Implementation Guide

## Overview

CANARY uses `golang-migrate/migrate` for database schema versioning with a pure Go SQLite driver (`modernc.org/sqlite`) to avoid CGO dependencies.

## Architecture

### Components

1. **Migration Files** - `internal/storage/migrations/*.sql`
   - Versioned SQL files (e.g., `000001_initial_schema.up.sql`, `000001_initial_schema.down.sql`)
   - Embedded in binary via `//go:embed`
   - Named with format: `NNNNNN_description.{up|down}.sql`

2. **Database Layer** - `internal/storage/db.go`
   - `InitDB()` - Initialize database connection
   - `MigrateDB()` - Apply migrations forward
   - `TeardownDB()` - Roll back migrations
   - `DatabasePopulated()` - Check migration status

3. **Storage Layer** - `internal/storage/storage.go`
   - `Open()` - Opens DB and automatically runs migrations
   - Wraps `*sqlx.DB` for enhanced query capabilities
   - All storage operations use migrated schema

4. **CLI Commands** - `cmd/canary/main.go`
   - `canary migrate <steps>` - Run migrations
   - `canary rollback <steps>` - Roll back migrations
   - All storage commands auto-migrate on open

## Migration System

### Pure Go SQLite (modernc.org/sqlite)

**Why Pure Go?**
- No CGO dependency - easier cross-compilation
- Works on all platforms (Linux, macOS, Windows, ARM)
- Single binary distribution
- No external C libraries required
- Slightly larger binary (~11MB vs ~8MB) but more portable

**Drivers:**
```go
import (
    "github.com/jmoiron/sqlx"
    _ "modernc.org/sqlite"  // Pure Go driver
)
```

### Migration Flow

```
canary index
    ↓
storage.Open(dbPath)
    ↓
storage.InitDB(dbPath)  → Creates/opens SQLite file
    ↓
storage.MigrateDB(dbPath, "all")  → Runs migrations
    ↓
Apply PRAGMA foreign_keys = ON
    ↓
Return *DB with migrated schema
```

### Schema Versioning

Migrations create a `schema_migrations` table:

```sql
CREATE TABLE schema_migrations (
    version bigint NOT NULL PRIMARY KEY,
    dirty boolean NOT NULL
);
```

**Version tracking:**
- Each migration has a version number (e.g., `000001`, `000002`)
- `migrate` library tracks current version in `schema_migrations`
- `DatabasePopulated()` checks version against target

## Creating Migrations

### File Naming Convention

```
NNNNNN_description.up.sql    -- Apply migration
NNNNNN_description.down.sql  -- Rollback migration
```

**Example:**
```
000001_initial_schema.up.sql
000001_initial_schema.down.sql
000002_add_tags_column.up.sql
000002_add_tags_column.down.sql
```

### Migration Template

**Up Migration:**
```sql
-- CANARY: REQ=CBIN-XXX; FEATURE="MigrationName"; ASPECT=Storage; STATUS=IMPL; OWNER=canary; UPDATED=YYYY-MM-DD
-- Description of what this migration does

CREATE TABLE IF NOT EXISTS my_table (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_my_table_name ON my_table(name);
```

**Down Migration:**
```sql
-- CANARY: REQ=CBIN-XXX; FEATURE="MigrationName"; ASPECT=Storage; STATUS=IMPL; OWNER=canary; UPDATED=YYYY-MM-DD
-- Rollback description

DROP INDEX IF EXISTS idx_my_table_name;
DROP TABLE IF EXISTS my_table;
```

### Testing Migrations

```bash
# Apply migration
canary migrate 1
# Verify schema
sqlite3 .canary/canary.db ".schema"

# Roll back
canary rollback 1
# Verify tables removed
sqlite3 .canary/canary.db ".schema"

# Migrate all
canary migrate all
# Check version
sqlite3 .canary/canary.db "SELECT version FROM schema_migrations"
```

## Automatic Migration

**CANARY automatically detects and runs migrations when needed!**

### How It Works

Before any database command executes, CANARY:
1. Checks if database file exists
2. If exists, checks current schema version
3. Compares to latest version (defined in `storage.LatestVersion`)
4. Auto-migrates if needed
5. Shows user-friendly progress messages

### Example Scenarios

**First time using database:**
```bash
$ canary index
🔄 Creating database with schema version 1...
✅ Database created at version 1
✅ Indexed 288 CANARY tokens
```

**Database already up to date:**
```bash
$ canary list
Found 10 tokens:
...
# No migration message - already at latest version
```

**After upgrading canary binary:**
```bash
$ canary search "auth"
🔄 Migrating database from version 1 to 2...
✅ Database migrated to version 2
Search results for 'auth' (5 tokens):
...
```

### Commands That Auto-Migrate

✅ **Database commands (auto-migrate before running):**
- `canary index` - Build token database
- `canary list` - List tokens
- `canary search` - Search tokens
- `canary prioritize` - Update priorities
- `canary checkpoint` - Create snapshots

❌ **Non-database commands (skip migration check):**
- `canary init` - Initialize project structure
- `canary create` - Generate token templates
- `canary implement` - Scan source files
- `canary scan` - Legacy scanner
- `canary migrate` - Manual migration management
- `canary rollback` - Manual rollback

### Implementation Details

**PersistentPreRunE hook:**
```go
rootCmd.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
    // Skip non-database commands
    if skipCommands[cmd.Name()] {
        return nil
    }

    // Auto-migrate if needed
    return storage.AutoMigrate(dbPath)
}
```

**Version detection:**
- Checks `schema_migrations` table
- Compares `MAX(version)` to `storage.LatestVersion`
- Runs migrations if current < latest

**User experience:**
- Silent when up to date
- Friendly progress messages when migrating
- Emoji indicators (🔄 for in-progress, ✅ for complete)
- No interruption to workflow

## CLI Commands

### canary migrate

```bash
# Migrate to latest version
canary migrate all

# Migrate forward by N steps
canary migrate 1
canary migrate 3

# Custom database path
canary migrate all --db /path/to/db
```

**Output:**
```
Running migrations on: .canary/canary.db
✅ Migrations completed successfully
```

### canary rollback

```bash
# Roll back all migrations
canary rollback all

# Roll back by N steps
canary rollback 1
canary rollback 2

# Custom database path
canary rollback 1 --db /path/to/db
```

**Output:**
```
Rolling back migrations on: .canary/canary.db
✅ Rollback completed successfully
```

## Automatic Migration

**All storage commands automatically run migrations:**

```bash
# These commands call storage.Open(), which auto-migrates
canary index
canary list
canary search "keyword"
canary prioritize CBIN-001 Feature 1
canary checkpoint "name"
```

**No need to manually migrate** unless:
- Testing migrations specifically
- Rolling back for debugging
- Managing schema versions manually

## Error Handling

### Common Errors

**1. Database locked:**
```
Error: database is locked
```
**Solution:** Close other connections, check for running processes

**2. Migration already applied:**
```
Error: no change
```
**Solution:** Normal - migration already at target version

**3. Dirty migration:**
```
Error: Dirty database version 1. Fix and force version.
```
**Solution:**
```bash
# Force version (advanced - be careful)
sqlite3 .canary/canary.db "UPDATE schema_migrations SET dirty=false WHERE version=1"
```

**4. Migration file not found:**
```
Error: failed to create migration source
```
**Solution:** Rebuild binary to re-embed migration files

## Development Workflow

### Adding a New Migration

1. **Create migration files:**
```bash
# Create next version number (e.g., 000002)
touch internal/storage/migrations/000002_add_tags.up.sql
touch internal/storage/migrations/000002_add_tags.down.sql
```

2. **Write up migration:**
```sql
-- 000002_add_tags.up.sql
ALTER TABLE tokens ADD COLUMN tags TEXT;
CREATE INDEX idx_tokens_tags ON tokens(tags);
```

3. **Write down migration:**
```sql
-- 000002_add_tags.down.sql
DROP INDEX idx_tokens_tags;
ALTER TABLE tokens DROP COLUMN tags;
```

4. **Test locally:**
```bash
go build -o ./bin/canary ./cmd/canary
./bin/canary rollback all  # Start fresh
./bin/canary migrate all   # Apply all including new
./bin/canary rollback 1    # Test rollback
./bin/canary migrate 1     # Test forward again
```

5. **Verify schema:**
```bash
sqlite3 .canary/canary.db ".schema tokens"
```

### Testing Checklist

- [ ] Up migration creates tables/columns/indexes
- [ ] Down migration removes them cleanly
- [ ] Can migrate forward and backward multiple times
- [ ] No orphaned tables/indexes after rollback
- [ ] `schema_migrations` version updates correctly
- [ ] All storage commands work after migration

## Advanced Usage

### Check Migration Version

```bash
sqlite3 .canary/canary.db "SELECT version FROM schema_migrations"
```

### Manual Migration Control

```go
import "go.spyder.org/canary/internal/storage"

// Initialize without auto-migrate
db, err := storage.InitDB("/path/to/db")

// Run specific migration
err = storage.MigrateDB("/path/to/db", "1")

// Check if populated
populated, err := storage.DatabasePopulated(db, 1)
```

### Inspect Migration Files

```bash
# List embedded migrations
ls -la internal/storage/migrations/

# View migration content
cat internal/storage/migrations/000001_initial_schema.up.sql
```

## Production Considerations

### Backup Before Migration

```bash
# Backup database
cp .canary/canary.db .canary/canary.db.backup

# Run migration
canary migrate all

# If issues, restore
mv .canary/canary.db.backup .canary/canary.db
```

### Zero-Downtime Migrations

For production systems:

1. **Backward-compatible migrations first:**
   - Add new columns as nullable
   - Don't drop columns immediately
   - Use multi-step migrations

2. **Deploy code:**
   - New code works with old + new schema

3. **Clean up in later migration:**
   - Drop old columns
   - Add constraints

### Migration Best Practices

1. **Always test rollback** - Down migrations should work
2. **Keep migrations small** - One logical change per migration
3. **Use IF EXISTS** - Makes migrations idempotent
4. **Version control** - Commit migration files with code
5. **Document breaking changes** - Use comments in SQL
6. **Test with production data** - Use realistic test data

## Troubleshooting

### Migration won't apply

```bash
# Check current version
sqlite3 .canary/canary.db "SELECT * FROM schema_migrations"

# Check dirty flag
sqlite3 .canary/canary.db "SELECT dirty FROM schema_migrations WHERE version=1"

# If dirty, investigate last migration and clean up
```

### Database corrupted

```bash
# Check integrity
sqlite3 .canary/canary.db "PRAGMA integrity_check"

# If corrupted, rebuild
rm .canary/canary.db
canary migrate all
canary index  # Re-index all tokens
```

### Migrations out of sync

```bash
# Nuclear option - rebuild from scratch
canary rollback all
canary migrate all
canary index
```

## References

- **golang-migrate:** https://github.com/golang-migrate/migrate
- **modernc.org/sqlite:** https://gitlab.com/cznic/sqlite
- **sqlx:** https://github.com/jmoiron/sqlx
- **SQLite:** https://www.sqlite.org/

## Summary

✅ **Migration System:**
- Pure Go SQLite driver (no CGO)
- Embedded migration files
- Versioned schema management
- Automatic migrations on storage.Open()
- CLI commands for manual control

✅ **Commands:**
- `canary migrate all` - Apply all migrations
- `canary migrate N` - Apply N migrations
- `canary rollback N` - Roll back N migrations
- All storage commands auto-migrate

✅ **Benefits:**
- Cross-platform compatibility
- Single binary distribution
- Version-controlled schema
- Safe rollback capability
- No external dependencies
