// Copyright (c) 2025 by Developer Network.
//
// For more details, see the LICENSE file in the root directory of this
// source code repository or contact Developer Network at info@devnw.com.

// CANARY: REQ=CBIN-129; FEATURE="DatabaseMigrations"; ASPECT=Storage; STATUS=IMPL; OWNER=canary; UPDATED=2025-10-16
package storage

import (
	"embed"
	"errors"
	"fmt"
	"log/slog"
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
	LatestVersion   = 5 // Update this when adding new migrations
)

var ErrDatabaseNotPopulated = errors.New("database not migrated")

// InitDB initializes the database connection
func InitDB(dbPath string) (*sqlx.DB, error) {
	slog.Info("Initializing database", "path", dbPath)
	dir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
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
	if err := os.MkdirAll(filepath.Dir(dbPath), 0755); err != nil {
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

	defer m.Close()

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

	defer m.Close()

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

	defer db.Close()

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

// AutoMigrate automatically migrates the database if needed
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
