// Copyright (c) 2025 by Developer Network.
//
// For more details, see the LICENSE file in the root directory of this
// source code repository or contact Developer Network at info@devnw.com.

// CANARY: REQ=CBIN-146; FEATURE="DatabaseModes"; ASPECT=Storage; STATUS=IMPL; TEST=TestDatabaseModeInitialization; UPDATED=2025-10-18
package storage

import (
	"path/filepath"
	"testing"

	"go.devnw.com/canary/internal/storage/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGlobalDatabaseInitialization(t *testing.T) {
	// Setup temp home directory
	tmpHome, cleanup := testutil.TempHomeDir(t)
	defer cleanup()

	// Initialize global database
	manager := NewDatabaseManager()
	err := manager.Initialize(GlobalMode)
	require.NoError(t, err)

	// Verify database exists at ~/.canary/canary.db
	dbPath := filepath.Join(tmpHome, ".canary", "canary.db")
	assert.True(t, testutil.FileExists(dbPath), "database file should exist at %s", dbPath)

	// Verify mode is set correctly
	assert.Equal(t, GlobalMode, manager.Mode())
	assert.Equal(t, dbPath, manager.Location())

	// Verify database is functional
	assert.NotNil(t, manager.DB())

	// Cleanup
	require.NoError(t, manager.Close())
}

func TestLocalDatabaseInitialization(t *testing.T) {
	// Setup temp project directory
	tmpDir, cleanup := testutil.TempDir(t)
	defer cleanup()

	// Change to project directory
	restoreDir := testutil.Chdir(t, tmpDir)
	defer restoreDir()

	// Initialize local database
	manager := NewDatabaseManager()
	err := manager.Initialize(LocalMode)
	require.NoError(t, err)

	// Verify database exists at ./.canary/canary.db
	dbPath := filepath.Join(tmpDir, ".canary", "canary.db")
	assert.True(t, testutil.FileExists(dbPath), "database file should exist at %s", dbPath)

	// Verify mode is set correctly
	assert.Equal(t, LocalMode, manager.Mode())
	assert.Equal(t, dbPath, manager.Location())

	// Verify database is functional
	assert.NotNil(t, manager.DB())

	// Cleanup
	require.NoError(t, manager.Close())
}

func TestDatabasePrecedence(t *testing.T) {
	// Setup: Create both global and local databases
	_, cleanupHome := testutil.TempHomeDir(t)
	defer cleanupHome()

	tmpProject, cleanupProject := testutil.TempDir(t)
	defer cleanupProject()

	// Initialize global database
	globalManager := NewDatabaseManager()
	err := globalManager.Initialize(GlobalMode)
	require.NoError(t, err)
	globalPath := globalManager.Location()
	globalManager.Close()

	// Change to project directory and create local database
	restoreDir := testutil.Chdir(t, tmpProject)
	defer restoreDir()

	localManager := NewDatabaseManager()
	err = localManager.Initialize(LocalMode)
	require.NoError(t, err)
	localPath := localManager.Location()
	localManager.Close()

	// Test: Discovery should prefer local over global
	discoveredManager := NewDatabaseManager()
	err = discoveredManager.Discover()
	require.NoError(t, err)

	// Verify local database is used (precedence)
	assert.Equal(t, LocalMode, discoveredManager.Mode())
	assert.Equal(t, localPath, discoveredManager.Location())
	assert.NotEqual(t, globalPath, discoveredManager.Location())

	discoveredManager.Close()
}

func TestDatabaseDiscovery(t *testing.T) {
	tests := []struct {
		name           string
		setupGlobal    bool
		setupLocal     bool
		expectedMode   DatabaseMode
		expectError    bool
	}{
		{
			name:         "local exists - use local",
			setupGlobal:  true,
			setupLocal:   true,
			expectedMode: LocalMode,
			expectError:  false,
		},
		{
			name:         "only global exists - use global",
			setupGlobal:  true,
			setupLocal:   false,
			expectedMode: GlobalMode,
			expectError:  false,
		},
		{
			name:         "only local exists - use local",
			setupGlobal:  false,
			setupLocal:   true,
			expectedMode: LocalMode,
			expectError:  false,
		},
		{
			name:         "neither exists - error",
			setupGlobal:  false,
			setupLocal:   false,
			expectedMode: GlobalMode, // doesn't matter, test expects error
			expectError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup temp home
			_, cleanupHome := testutil.TempHomeDir(t)
			defer cleanupHome()

			// Setup temp project
			tmpProject, cleanupProject := testutil.TempDir(t)
			defer cleanupProject()

			restoreDir := testutil.Chdir(t, tmpProject)
			defer restoreDir()

			// Setup global if needed
			if tt.setupGlobal {
				globalMgr := NewDatabaseManager()
				err := globalMgr.Initialize(GlobalMode)
				require.NoError(t, err)
				globalMgr.Close()
			}

			// Setup local if needed
			if tt.setupLocal {
				localMgr := NewDatabaseManager()
				err := localMgr.Initialize(LocalMode)
				require.NoError(t, err)
				localMgr.Close()
			}

			// Test discovery
			manager := NewDatabaseManager()
			err := manager.Discover()

			if tt.expectError {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expectedMode, manager.Mode())
				manager.Close()
			}
		})
	}
}

func TestGlobalDatabaseLocation(t *testing.T) {
	// Test with HOME environment variable
	tmpHome, cleanup := testutil.TempHomeDir(t)
	defer cleanup()

	manager := NewDatabaseManager()
	err := manager.Initialize(GlobalMode)
	require.NoError(t, err)
	defer manager.Close()

	expectedPath := filepath.Join(tmpHome, ".canary", "canary.db")
	assert.Equal(t, expectedPath, manager.Location())
}

func TestLocalDatabaseLocation(t *testing.T) {
	tmpDir, cleanup := testutil.TempDir(t)
	defer cleanup()

	restoreDir := testutil.Chdir(t, tmpDir)
	defer restoreDir()

	manager := NewDatabaseManager()
	err := manager.Initialize(LocalMode)
	require.NoError(t, err)
	defer manager.Close()

	expectedPath := filepath.Join(tmpDir, ".canary", "canary.db")
	assert.Equal(t, expectedPath, manager.Location())
}

func TestDatabaseManagerClose(t *testing.T) {
	_, cleanup := testutil.TempHomeDir(t)
	defer cleanup()

	manager := NewDatabaseManager()
	err := manager.Initialize(GlobalMode)
	require.NoError(t, err)

	// Close should not return error
	err = manager.Close()
	assert.NoError(t, err)

	// Closing twice should be safe
	err = manager.Close()
	assert.NoError(t, err)
}
