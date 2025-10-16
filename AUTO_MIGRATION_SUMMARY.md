# Automatic Database Migration - Implementation Summary

## Overview

Enhanced CANARY's migration system to automatically detect and upgrade database schema versions when running any database command. Users never need to manually run `canary migrate` - it happens automatically!

## Problem Solved

**Before:**
- Users had to remember to run `canary migrate all` after upgrading
- Database commands would fail if schema was out of date
- Manual migration management required
- Risk of running commands on old schema

**After:**
- Automatic detection of database version on startup
- Auto-upgrade when new canary binary has newer migrations
- Seamless user experience - migrations "just work"
- User-friendly progress messages during migration

## Implementation

### 1. Version Constant (CBIN-130)

**File:** `internal/storage/db.go`

Added `LatestVersion` constant:
```go
const (
    DBDriver         = "sqlite"
    DBMigrationPath  = "migrations"
    DBSourceName     = "iofs"
    DBURLProtocol    = "sqlite://"
    MigrateAll       = "all"
    LatestVersion    = 1  // Update this when adding new migrations
)
```

**Purpose:** Single source of truth for expected schema version.

### 2. Version Detection (CBIN-130)

**File:** `internal/storage/db.go`

**Function:** `NeedsMigration(dbPath string) (bool, int, error)`

```go
func NeedsMigration(dbPath string) (bool, int, error) {
    // 1. Check if database file exists
    if _, err := os.Stat(dbPath); os.IsNotExist(err) {
        return false, 0, nil // DB doesn't exist yet
    }

    // 2. Open database and check schema_migrations table
    db, err := sqlx.Open(DBDriver, dbPath)
    defer db.Close()

    var tableExists bool
    err = db.Get(&tableExists, "SELECT EXISTS(...)")
    if !tableExists {
        return true, 0, nil // DB exists but not migrated
    }

    // 3. Get current version
    var currentVersion int
    err = db.Get(&currentVersion, "SELECT COALESCE(MAX(version), 0) FROM schema_migrations WHERE dirty = 0")

    // 4. Compare to latest version
    if currentVersion < LatestVersion {
        return true, currentVersion, nil
    }

    return false, currentVersion, nil
}
```

**Returns:**
- `bool` - Whether migration is needed
- `int` - Current database version
- `error` - Any error encountered

### 3. Auto-Migration Logic (CBIN-130)

**File:** `internal/storage/db.go`

**Function:** `AutoMigrate(dbPath string) error`

```go
func AutoMigrate(dbPath string) error {
    // Check if database file exists
    _, err := os.Stat(dbPath)
    dbExists := err == nil

    if dbExists {
        needsMigration, currentVersion, err := NeedsMigration(dbPath)
        if err != nil {
            return fmt.Errorf("failed to check migration status: %w", err)
        }

        if !needsMigration {
            slog.Debug("Database is up to date", "version", currentVersion)
            return nil
        }

        slog.Info("Database migration needed", "currentVersion", currentVersion, "targetVersion", LatestVersion)
        fmt.Printf("🔄 Migrating database from version %d to %d...\n", currentVersion, LatestVersion)
    } else {
        slog.Info("Database does not exist, will create with migrations", "path", dbPath)
        fmt.Printf("🔄 Creating database with schema version %d...\n", LatestVersion)
    }

    if err := MigrateDB(dbPath, MigrateAll); err != nil {
        return fmt.Errorf("auto-migration failed: %w", err)
    }

    if dbExists {
        fmt.Printf("✅ Database migrated to version %d\n", LatestVersion)
    } else {
        fmt.Printf("✅ Database created at version %d\n", LatestVersion)
    }
    return nil
}
```

**Features:**
- Detects if database exists or is being created
- Shows user-friendly progress messages
- Silent when database is already up to date
- Uses emoji indicators for visual feedback

### 4. PersistentPreRunE Hook (CBIN-130)

**File:** `cmd/canary/main.go`

**Added to rootCmd:**
```go
var rootCmd = &cobra.Command{
    Use:   "canary",
    Short: "Track requirements via CANARY tokens in source code",
    // ... other fields ...
    PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
        // Skip auto-migration for commands that don't use the database
        skipCommands := map[string]bool{
            "init":         true,  // Project initialization
            "create":       true,  // Token template generation
            "constitution": true,  // Constitution management
            "specify":      true,  // Spec creation
            "plan":         true,  // Plan creation
            "implement":    true,  // Source file scanning
            "scan":         true,  // Legacy scanner
            "help":         true,  // Help display
            "completion":   true,  // Shell completion
            "migrate":      true,  // Manages migrations itself
            "rollback":     true,  // Manages migrations itself
        }

        if skipCommands[cmd.Name()] {
            return nil
        }

        // Get database path from flags
        dbPath := ".canary/canary.db" // default
        if cmd.Flags().Lookup("db") != nil {
            dbPath, _ = cmd.Flags().GetString("db")
        }

        // Auto-migrate if needed
        if err := storage.AutoMigrate(dbPath); err != nil {
            return fmt.Errorf("auto-migration failed: %w", err)
        }

        return nil
    },
}
```

**How it works:**
1. Runs before every command via `PersistentPreRunE`
2. Skips non-database commands for performance
3. Extracts database path from command flags
4. Calls `AutoMigrate()` which handles version detection
5. Returns error if migration fails (prevents command execution)

### 5. Removed Redundant Migration (CBIN-130)

**File:** `internal/storage/storage.go`

**Before:**
```go
func Open(dbPath string) (*DB, error) {
    conn := InitDB(dbPath)
    MigrateDB(dbPath, MigrateAll)  // ❌ Redundant - now in PersistentPreRunE
    // ...
}
```

**After:**
```go
func Open(dbPath string) (*DB, error) {
    // Note: Migrations handled by CLI's PersistentPreRunE
    conn := InitDB(dbPath)
    // Enable foreign keys
    conn.Exec("PRAGMA foreign_keys = ON")
    return &DB{conn: conn, path: dbPath}, nil
}
```

**Reason:** Migration now happens once before command execution, not every time `Open()` is called.

## Testing

### Test 1: Fresh Database Creation
```bash
$ rm -f .canary/canary.db
$ ./bin/canary index

🔄 Creating database with schema version 1...
✅ Database created at version 1
✅ Indexed 288 CANARY tokens
```
**Result:** ✅ Database created with migrations

### Test 2: Database Already Up to Date
```bash
$ ./bin/canary list --limit 3

Found 3 tokens:
...
```
**Result:** ✅ No migration message (silent when up to date)

### Test 3: Upgrade from Old Version
```bash
$ ./bin/canary rollback 1
✅ Rollback completed successfully

$ ./bin/canary list

🔄 Migrating database from version 0 to 1...
✅ Database migrated to version 1
Found 10 tokens:
...
```
**Result:** ✅ Auto-detected old version and upgraded

### Test 4: Non-Database Commands Skip Migration
```bash
$ ./bin/canary create CBIN-999 TestFeature

// CANARY: REQ=CBIN-999; FEATURE="TestFeature"; ASPECT=API; STATUS=IMPL; UPDATED=2025-10-16
```
**Result:** ✅ No migration check (command doesn't use database)

### Test 5: Manual Migration Still Works
```bash
$ ./bin/canary migrate all

Running migrations on: .canary/canary.db
✅ Migrations completed successfully
```
**Result:** ✅ Manual migration commands still work

## User Experience

### Before Auto-Migration

```bash
# User upgrades canary binary
$ canary list

Error: SQL logic error: no such table: new_table (1)

# User has to remember to migrate
$ canary migrate all
✅ Migrations completed successfully

$ canary list
Found 10 tokens:
...
```

### After Auto-Migration

```bash
# User upgrades canary binary
$ canary list

🔄 Migrating database from version 1 to 2...
✅ Database migrated to version 2
Found 10 tokens:
...
```

**Difference:** No manual step, seamless experience, automatic handling.

## Benefits

### 1. Zero Manual Intervention
- Users never need to run `canary migrate`
- Upgrades happen transparently
- "It just works" philosophy

### 2. Safe Upgrades
- Version detection prevents running old schema
- Automatic migration before any database operation
- Error handling prevents corruption

### 3. Clear Communication
- Progress messages during migration
- Emoji indicators (🔄, ✅)
- Silent when up to date

### 4. Performance Optimized
- Skip commands that don't use database
- Check version only once per command
- Minimal overhead when up to date

### 5. Developer Friendly
- Single constant (`LatestVersion`) to update
- Clear separation of concerns
- Easy to test migration logic

## Developer Workflow

### Adding a New Migration

1. **Create migration files:**
```bash
touch internal/storage/migrations/000002_add_tags.up.sql
touch internal/storage/migrations/000002_add_tags.down.sql
```

2. **Write migration SQL:**
```sql
-- 000002_add_tags.up.sql
ALTER TABLE tokens ADD COLUMN tags TEXT;
CREATE INDEX idx_tokens_tags ON tokens(tags);
```

3. **Update version constant:**
```go
// internal/storage/db.go
const LatestVersion = 2  // Was 1, now 2
```

4. **Rebuild and test:**
```bash
go build -o ./bin/canary ./cmd/canary
./bin/canary list  # Auto-migrates to version 2
```

**That's it!** Users automatically get the new schema.

## Commands Affected

### Auto-Migrate Before Running
- ✅ `canary index`
- ✅ `canary list`
- ✅ `canary search`
- ✅ `canary prioritize`
- ✅ `canary checkpoint`

### Skip Auto-Migration
- ❌ `canary init`
- ❌ `canary create`
- ❌ `canary constitution`
- ❌ `canary specify`
- ❌ `canary plan`
- ❌ `canary implement`
- ❌ `canary scan`
- ❌ `canary migrate` (manages migrations itself)
- ❌ `canary rollback` (manages migrations itself)

## Implementation Details

### Flow Diagram

```
User runs: canary list
    ↓
rootCmd.PersistentPreRunE()
    ↓
Check if command uses database
    ↓ Yes
storage.AutoMigrate(dbPath)
    ↓
Check if DB exists
    ↓ Yes
NeedsMigration(dbPath)
    ↓
Get current version from schema_migrations
    ↓
Compare to LatestVersion
    ↓ currentVersion < LatestVersion
Show progress message
    ↓
MigrateDB(dbPath, "all")
    ↓
Show completion message
    ↓
Continue with command
    ↓
storage.Open(dbPath)
    ↓
Execute list logic
    ↓
Display results
```

### Error Handling

**If migration fails:**
```bash
$ canary list

🔄 Migrating database from version 1 to 2...
Error: auto-migration failed: migration failed: <error details>
```

**Command does not execute** - prevents corruption from running on wrong schema.

## CANARY Token

**Added:**
- CBIN-130: AutoMigration (Storage, IMPL)
  - `NeedsMigration()` function
  - `AutoMigrate()` function
  - PersistentPreRunE hook in rootCmd

## Summary

✅ **Implemented:**
- Automatic database version detection
- Auto-migration before database commands
- User-friendly progress messages
- Skip non-database commands for performance
- Single `LatestVersion` constant to maintain

✅ **Benefits:**
- Zero manual migration steps
- Seamless binary upgrades
- Clear user communication
- Safe schema handling
- Developer-friendly workflow

✅ **Testing:**
- Fresh database creation: ✅
- Database up to date: ✅
- Upgrade from old version: ✅
- Non-database commands: ✅
- Manual migration: ✅

**Auto-migration is production-ready and user-friendly!** 🎉
