// Copyright (c) 2025 by Developer Network.
//
// For more details, see the LICENSE file in the root directory of this
// source code repository or contact Developer Network at info@devnw.com.

// CANARY: REQ=ENG-4312; FEATURE="DatabaseMigrations"; ASPECT=Storage; STATUS=IMPL; OWNER=canary; UPDATED=2025-10-16
package storage

import (
	"embed"
	"errors"
	"fmt"
	"io/fs"
	"log/slog"
	"net/url"
	"os"
	"path/filepath"
	"strconv"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/sqlite"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/jmoiron/sqlx"
	_ "modernc.org/sqlite" // Pure Go SQLite implementation (no CGO)
)

//go:embed migrations/*.sql
var migrationFiles embed.FS

const (
	DBDriver        = "sqlite"
	DBMigrationPath = "migrations"
	DBSourceName    = "iofs"
	DBURLProtocol   = "sqlite://"
	MigrateAll      = "all"
	LatestVersion   = 7 // Update this when adding new migrations
)

var ErrDatabaseNotPopulated = errors.New("database not migrated")

// InitDB initializes the database connection
func InitDB(dbPath string) (*sqlx.DB, error) {
	slog.Info("Initializing database", "path", dbPath)
	dir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dir, 0750); err != nil {
		return nil, fmt.Errorf("failed to create database directory at %s: %w", dir, err)
	}
	db, err := sqlx.Open(DBDriver, dbPath)
	if err != nil {
		return nil, fmt.Errorf("error opening database at %s: %w", dbPath, err)
	}
	slog.Info("Database connection initialized")
	return db, nil
}

// MigrateDB applies the database migrations stored in migrations/*.sql
// It takes a single argument which is either "all" to migrate to the latest version
// or an integer to migrate by that many steps.
func MigrateDB(dbPath string, steps string) error {
	slog.Info("Migrating database", "path", dbPath, "steps", steps)

	// Ensure the database directory exists before migrating
	if err := os.MkdirAll(filepath.Dir(dbPath), 0750); err != nil {
		return fmt.Errorf("failed to create database directory at %s: %w", filepath.Dir(dbPath), err)
	}

	driver, err := iofs.New(migrationFiles, DBMigrationPath)
	if err != nil {
		return fmt.Errorf("failed to create migration source: %w", err)
	}

	m, err := migrate.NewWithSourceInstance(DBSourceName, driver, DBURLProtocol+dbPath)
	if err != nil {
		return fmt.Errorf("error creating migration instance for database at %s: %w", dbPath, err)
	}

	defer func() { _, _ = m.Close() }()

	switch {
	case steps == MigrateAll:
		slog.Info("Migrating database to latest version")
		if err = m.Up(); err != nil && err != migrate.ErrNoChange {
			return fmt.Errorf("failed to migrate database: %w", err)
		}
		if err == migrate.ErrNoChange {
			slog.Info("Database already at latest version")
		}
	case isInt(steps):
		slog.Info("Migrating database by steps", "steps", steps)
		stepCount, err := strconv.Atoi(steps)
		if err != nil {
			return fmt.Errorf("invalid number of migration steps: %s: %w", steps, err)
		}
		if stepCount == 0 {
			return errors.New("migration steps cannot be zero, please specify a positive integer or 'all'")
		}
		if err = m.Steps(stepCount); err != nil && err != migrate.ErrNoChange {
			return fmt.Errorf("failed to migrate database by %d steps: %w", stepCount, err)
		}
		if err == migrate.ErrNoChange {
			slog.Info("No migration changes to apply")
		}
	default:
		return fmt.Errorf("invalid argument for migration steps: %s, expected 'all' or an integer", steps)
	}

	slog.Info("Database migrated successfully")
	return nil
}

// TeardownDB is the negative inverse of MigrateDB, rolling back migrations
// It takes a single argument which is either "all" to roll back all migrations
// or an integer to roll back by that many steps.
func TeardownDB(dbPath string, steps string) error {
	slog.Debug("Tearing down database", "path", dbPath, "steps", steps)

	driver, err := iofs.New(migrationFiles, DBMigrationPath)
	if err != nil {
		return fmt.Errorf("failed to create migration source: %w", err)
	}

	m, err := migrate.NewWithSourceInstance(DBSourceName, driver, DBURLProtocol+dbPath)
	if err != nil {
		return fmt.Errorf("error creating migration instance: %w", err)
	}

	defer func() { _, _ = m.Close() }()

	switch {
	case steps == MigrateAll:
		slog.Info("Rolling back all migrations")
		if err = m.Down(); err != nil && err != migrate.ErrNoChange {
			return fmt.Errorf("failed to roll back all migrations: %w", err)
		}
	case isInt(steps):
		slog.Info("Rolling back database by steps", "steps", steps)
		stepCount, err := strconv.Atoi(steps)
		if err != nil {
			return fmt.Errorf("invalid number of migration steps: %s: %w", steps, err)
		}
		if stepCount == 0 {
			return errors.New("migration steps cannot be zero, please specify a positive integer or 'all'")
		}
		if err = m.Steps(-stepCount); err != nil && err != migrate.ErrNoChange {
			return fmt.Errorf("failed to roll back database by %d steps: %w", stepCount, err)
		}
	default:
		return fmt.Errorf("invalid argument for migration steps: %s, expected 'all' or an integer", steps)
	}

	slog.Info("Database teardown completed")
	return nil
}

// DatabasePopulated checks if the database is fully migrated and populated
// We only return an error here if we're getting database issues. Bool return should
// reflect the state of the database.
func DatabasePopulated(db *sqlx.DB, targetVersion int) (bool, error) {
	slog.Debug("Checking if database is fully migrated and populated")

	var populated bool
	err := db.Get(&populated, "SELECT EXISTS(SELECT 1 FROM schema_migrations)")
	if err != nil {
		return false, fmt.Errorf("failed to check if database is populated: %w", err)
	}

	if !populated {
		slog.Warn("Database is not populated", "targetVersion", targetVersion)
		return false, nil
	}

	// If no specific target version is provided, consider population sufficient.
	if targetVersion <= 0 {
		return true, nil
	}

	var version int
	err = db.Get(&version, "SELECT version FROM schema_migrations ORDER BY version DESC LIMIT 1")
	if err != nil {
		return false, fmt.Errorf("failed to retrieve current database version: %w", err)
	}

	slog.Debug("Current database version", "version", version)

	if version < targetVersion {
		slog.Warn("Database is not fully migrated", "currentVersion", version, "targetVersion", targetVersion)
		return false, nil
	}

	slog.Debug("Database version is up to date or ahead", "version", version, "targetVersion", targetVersion)
	return true, nil
}

// isInt checks if a string is a valid integer
func isInt(s string) bool {
	_, err := strconv.Atoi(s)
	return err == nil
}

// NeedsMigration checks if the database exists and needs migration
func NeedsMigration(dbPath string) (bool, int, error) {
	// Check if database file exists
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		return false, 0, nil // Database doesn't exist yet
	}

	// Open database to check version
	db, err := sqlx.Open(DBDriver, dbPath)
	if err != nil {
		return false, 0, fmt.Errorf("failed to open database: %w", err)
	}

	defer func() { _ = db.Close() }()

	// Check if schema_migrations table exists
	var tableExists bool
	err = db.Get(&tableExists, "SELECT EXISTS(SELECT 1 FROM sqlite_master WHERE type='table' AND name='schema_migrations')")
	if err != nil {
		return false, 0, fmt.Errorf("failed to check schema_migrations table: %w", err)
	}

	if !tableExists {
		return true, 0, nil // Database exists but not migrated
	}

	// Get current version
	var currentVersion int
	err = db.Get(&currentVersion, "SELECT COALESCE(MAX(version), 0) FROM schema_migrations WHERE dirty = 0")
	if err != nil {
		return false, 0, fmt.Errorf("failed to get current version: %w", err)
	}

	// Check if migration needed
	if currentVersion < LatestVersion {
		return true, currentVersion, nil
	}

	return false, currentVersion, nil
}

// AutoMigrate automatically migrates the database if needed.
//
// Progress banners go to stderr, never stdout: several commands emit a
// machine-readable line on stdout, and a schema notice interleaved with it
// would corrupt output the caller is parsing. AutoMigrate creates the
// database when it is missing, so only a writer may call it.
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
		fmt.Fprintf(os.Stderr, "🔄 Migrating database from version %d to %d...\n", currentVersion, LatestVersion)
	} else {
		slog.Info("Database does not exist, will create with migrations", "path", dbPath)
		fmt.Fprintf(os.Stderr, "🔄 Creating database with schema version %d...\n", LatestVersion)
	}

	if err := MigrateDB(dbPath, MigrateAll); err != nil {
		return fmt.Errorf("auto-migration failed: %w", err)
	}

	if dbExists {
		fmt.Fprintf(os.Stderr, "✅ Database migrated to version %d\n", LatestVersion)
	} else {
		fmt.Fprintf(os.Stderr, "✅ Database created at version %d\n", LatestVersion)
	}
	return nil
}

// dsn builds a driver DSN for dbPath with the given query parameters. The
// path is carried as a file: URI so a path containing '?' or '#' cannot be
// mistaken for the start of the parameter list.
func dsn(dbPath string, params url.Values) (string, error) {
	abs, err := filepath.Abs(dbPath)
	if err != nil {
		return "", fmt.Errorf("resolve database path %s: %w", dbPath, err)
	}
	u := url.URL{Scheme: "file", Path: filepath.ToSlash(abs), RawQuery: params.Encode()}
	return u.String(), nil
}

// OpenRW opens dbPath for reading and writing, creating and migrating it when
// necessary. Only a command that may legitimately modify the index calls
// this: it is the one entry point allowed to bring a database into existence.
//
// The connection is configured with a busy timeout (so a concurrent writer
// yields a wait rather than an immediate SQLITE_BUSY), WAL journalling (so a
// reader is never blocked by the indexer), and immediate transactions (so
// `canary index` takes its write lock at BEGIN instead of discovering a
// conflict halfway through the rebuild).
func OpenRW(dbPath string) (*DB, error) {
	if err := os.MkdirAll(filepath.Dir(dbPath), 0o750); err != nil {
		return nil, fmt.Errorf("create database directory: %w", err)
	}
	if err := AutoMigrate(dbPath); err != nil {
		return nil, err
	}

	name, err := dsn(dbPath, url.Values{
		"_pragma": {"busy_timeout(5000)", "journal_mode(WAL)", "foreign_keys(ON)"},
		"_txlock": {"immediate"},
	})
	if err != nil {
		return nil, err
	}

	conn, err := sqlx.Open(DBDriver, name)
	if err != nil {
		return nil, fmt.Errorf("open database at %s: %w", dbPath, err)
	}
	if err := conn.Ping(); err != nil {
		_ = conn.Close()
		return nil, fmt.Errorf("open database at %s: %w", dbPath, err)
	}
	return &DB{conn: conn, path: dbPath}, nil
}

// CANARY: REQ=CBIN-309; FEATURE="StaleSchemaGuard"; ASPECT=Storage; STATUS=TESTED; TEST=TestOpenRORefusesStaleSchema,TestStaleSchemaRefusesRead; UPDATED=2026-08-30

// ErrSchemaOutOfDate is what a read-only open returns when the database on
// disk predates the current schema. It is a state problem with one fix, and
// it is deliberately NOT solved by migrating: a read command that rewrote the
// schema behind the caller would be the very side effect OpenRO exists to
// prevent, and it would do so while another process may be reading. The
// message is the whole remedy, so commands surface it verbatim.
var ErrSchemaOutOfDate = errors.New("index schema is out of date; run 'canary index'")

// OpenRO opens dbPath read-only. A missing database is reported as
// fs.ErrNotExist and NOTHING is created -- not the file, not its parent
// directory. Read commands use this so running `canary list` in a repository
// that was never indexed leaves the repository exactly as it found it.
//
// A database that exists but predates the current schema is refused with
// ErrSchemaOutOfDate. This is the single choke point for that check because
// every read command reaches the index through here: without it a v6
// database answered `canary list` with a raw "no such column: content_hash"
// from deep inside the query layer, which names neither the problem nor the
// fix.
//
// The connection is opened with SQLite's own mode=ro and query_only, so a
// write attempted through it fails at the engine rather than relying on every
// caller to behave.
func OpenRO(dbPath string) (*DB, error) {
	if _, err := os.Stat(dbPath); err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return nil, fmt.Errorf("no index at %s: %w", dbPath, fs.ErrNotExist)
		}
		return nil, fmt.Errorf("stat database at %s: %w", dbPath, err)
	}

	stale, currentVersion, err := NeedsMigration(dbPath)
	if err != nil {
		return nil, fmt.Errorf("check index schema at %s: %w", dbPath, err)
	}
	if stale {
		slog.Debug("Refusing read against stale index schema",
			"path", dbPath, "currentVersion", currentVersion, "targetVersion", LatestVersion)
		return nil, ErrSchemaOutOfDate
	}

	name, err := dsn(dbPath, url.Values{
		"mode":    {"ro"},
		"_pragma": {"busy_timeout(5000)", "query_only(ON)"},
	})
	if err != nil {
		return nil, err
	}

	conn, err := sqlx.Open(DBDriver, name)
	if err != nil {
		return nil, fmt.Errorf("open database at %s: %w", dbPath, err)
	}
	if err := conn.Ping(); err != nil {
		_ = conn.Close()
		return nil, fmt.Errorf("open database at %s: %w", dbPath, err)
	}
	return &DB{conn: conn, path: dbPath, readOnly: true}, nil
}
