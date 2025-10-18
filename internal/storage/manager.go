// Copyright (c) 2025 by Developer Network.
//
// For more details, see the LICENSE file in the root directory of this
// source code repository or contact Developer Network at info@devnw.com.

// CANARY: REQ=CBIN-146; FEATURE="DatabaseModes"; ASPECT=Storage; STATUS=IMPL; UPDATED=2025-10-18
package storage

import (
	"database/sql"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/jmoiron/sqlx"
	_ "modernc.org/sqlite"
)

// DatabaseMode represents the initialization mode for the database
type DatabaseMode int

const (
	GlobalMode DatabaseMode = iota
	LocalMode
)

// String returns the string representation of DatabaseMode
func (dm DatabaseMode) String() string {
	switch dm {
	case GlobalMode:
		return "global"
	case LocalMode:
		return "local"
	default:
		return "unknown"
	}
}

// DatabaseManager manages both global and local database connections
type DatabaseManager struct {
	conn *sqlx.DB
	path string
	mode DatabaseMode
}

// NewDatabaseManager creates a new database manager
func NewDatabaseManager() *DatabaseManager {
	return &DatabaseManager{}
}

// Initialize initializes the database in the specified mode
func (dm *DatabaseManager) Initialize(mode DatabaseMode) error {
	var dbPath string

	switch mode {
	case GlobalMode:
		// Global database location: ~/.canary/canary.db
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("get home directory: %w", err)
		}
		dbPath = filepath.Join(homeDir, ".canary", "canary.db")

	case LocalMode:
		// Local database location: ./.canary/canary.db
		cwd, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("get working directory: %w", err)
		}
		dbPath = filepath.Join(cwd, ".canary", "canary.db")

	default:
		return fmt.Errorf("invalid database mode: %v", mode)
	}

	// Create directory if needed
	dbDir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dbDir, 0755); err != nil {
		return fmt.Errorf("create database directory: %w", err)
	}

	// Open database connection
	conn, err := InitDB(dbPath)
	if err != nil {
		return fmt.Errorf("initialize database: %w", err)
	}

	// Enable foreign keys
	if _, err := conn.Exec("PRAGMA foreign_keys = ON"); err != nil {
		conn.Close()
		return fmt.Errorf("enable foreign keys: %w", err)
	}

	dm.conn = conn
	dm.path = dbPath
	dm.mode = mode

	return nil
}

// Discover attempts to find an existing database, with local taking precedence over global
func (dm *DatabaseManager) Discover() error {
	// Check for local database first (./.canary/canary.db)
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("get working directory: %w", err)
	}

	localPath := filepath.Join(cwd, ".canary", "canary.db")
	if _, err := os.Stat(localPath); err == nil {
		// Local database exists - use it
		return dm.open(localPath, LocalMode)
	}

	// Check for global database (~/canary/canary.db)
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("get home directory: %w", err)
	}

	globalPath := filepath.Join(homeDir, ".canary", "canary.db")
	if _, err := os.Stat(globalPath); err == nil {
		// Global database exists - use it
		return dm.open(globalPath, GlobalMode)
	}

	// No database found
	return errors.New("no database found: run 'canary init' or 'canary init --global' to initialize")
}

// open opens an existing database at the specified path
func (dm *DatabaseManager) open(dbPath string, mode DatabaseMode) error {
	conn, err := InitDB(dbPath)
	if err != nil {
		return fmt.Errorf("open database: %w", err)
	}

	// Enable foreign keys
	if _, err := conn.Exec("PRAGMA foreign_keys = ON"); err != nil {
		conn.Close()
		return fmt.Errorf("enable foreign keys: %w", err)
	}

	dm.conn = conn
	dm.path = dbPath
	dm.mode = mode

	return nil
}

// Mode returns the current database mode
func (dm *DatabaseManager) Mode() DatabaseMode {
	return dm.mode
}

// Location returns the database file path
func (dm *DatabaseManager) Location() string {
	return dm.path
}

// DB returns the underlying database connection
func (dm *DatabaseManager) DB() *sql.DB {
	if dm.conn == nil {
		return nil
	}
	return dm.conn.DB
}

// Close closes the database connection
func (dm *DatabaseManager) Close() error {
	if dm.conn == nil {
		return nil
	}
	return dm.conn.Close()
}
