// Copyright (c) 2025 by Developer Network.
//
// For more details, see the LICENSE file in the root directory of this
// source code repository or contact Developer Network at info@devnw.com.

// CANARY: REQ=CBIN-146; FEATURE="ContextManagement"; ASPECT=Engine; STATUS=IMPL; TEST=TestContextManagement; UPDATED=2025-10-18
package storage

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.devnw.com/canary/internal/storage/testutil"
)

func TestDetectProjectFromPath(t *testing.T) {
	_, cleanup := testutil.TempHomeDir(t)
	defer cleanup()

	// Initialize global database
	manager := NewDatabaseManager()
	err := manager.Initialize(GlobalMode)
	require.NoError(t, err)
	defer manager.Close()

	registry := NewProjectRegistry(manager)

	// Create project directories
	projectDir, cleanupProject := testutil.TempDir(t)
	defer cleanupProject()

	// Register project
	project := &Project{
		Name: "Test Project",
		Path: projectDir,
	}
	err = registry.Register(project)
	require.NoError(t, err)

	// Create context manager
	ctx := NewContextManager(manager)

	// Change to project directory
	restoreDir := testutil.Chdir(t, projectDir)
	defer restoreDir()

	// Detect project
	detected, err := ctx.DetectProject()
	require.NoError(t, err)
	assert.Equal(t, project.ID, detected.ID)
	assert.Equal(t, project.Path, detected.Path)
}

func TestDetectProjectFromSubdirectory(t *testing.T) {
	_, cleanup := testutil.TempHomeDir(t)
	defer cleanup()

	manager := NewDatabaseManager()
	err := manager.Initialize(GlobalMode)
	require.NoError(t, err)
	defer manager.Close()

	registry := NewProjectRegistry(manager)

	// Create project directory with subdirectory
	projectDir, cleanupProject := testutil.TempDir(t)
	defer cleanupProject()

	subDir := filepath.Join(projectDir, "src", "internal")
	err = testutil.SetupProjectDir(t, subDir)
	require.NoError(t, err)

	// Register project
	project := &Project{
		Name: "Test Project",
		Path: projectDir,
	}
	err = registry.Register(project)
	require.NoError(t, err)

	ctx := NewContextManager(manager)

	// Change to subdirectory
	restoreDir := testutil.Chdir(t, subDir)
	defer restoreDir()

	// Should still detect parent project
	detected, err := ctx.DetectProject()
	require.NoError(t, err)
	assert.Equal(t, project.ID, detected.ID)
}

func TestDetectProjectNoMatch(t *testing.T) {
	_, cleanup := testutil.TempHomeDir(t)
	defer cleanup()

	manager := NewDatabaseManager()
	err := manager.Initialize(GlobalMode)
	require.NoError(t, err)
	defer manager.Close()

	ctx := NewContextManager(manager)

	// Change to directory with no registered project
	tmpDir, cleanupTmp := testutil.TempDir(t)
	defer cleanupTmp()

	restoreDir := testutil.Chdir(t, tmpDir)
	defer restoreDir()

	// Should return error
	_, err = ctx.DetectProject()
	assert.Error(t, err)
}

func TestSwitchProjectContext(t *testing.T) {
	_, cleanup := testutil.TempHomeDir(t)
	defer cleanup()

	manager := NewDatabaseManager()
	err := manager.Initialize(GlobalMode)
	require.NoError(t, err)
	defer manager.Close()

	registry := NewProjectRegistry(manager)
	ctx := NewContextManager(manager)

	// Register two projects
	project1 := &Project{Name: "Project 1", Path: "/path/1"}
	err = registry.Register(project1)
	require.NoError(t, err)

	project2 := &Project{Name: "Project 2", Path: "/path/2"}
	err = registry.Register(project2)
	require.NoError(t, err)

	// Switch to project 1
	err = ctx.SwitchTo(project1.ID)
	require.NoError(t, err)

	// Verify current project
	current, err := ctx.GetCurrent()
	require.NoError(t, err)
	assert.Equal(t, project1.ID, current.ID)

	// Switch to project 2
	err = ctx.SwitchTo(project2.ID)
	require.NoError(t, err)

	// Verify current project changed
	current, err = ctx.GetCurrent()
	require.NoError(t, err)
	assert.Equal(t, project2.ID, current.ID)
}

func TestSwitchToNonexistentProject(t *testing.T) {
	_, cleanup := testutil.TempHomeDir(t)
	defer cleanup()

	manager := NewDatabaseManager()
	err := manager.Initialize(GlobalMode)
	require.NoError(t, err)
	defer manager.Close()

	ctx := NewContextManager(manager)

	// Try to switch to non-existent project
	err = ctx.SwitchTo("nonexistent-project")
	assert.Error(t, err)
}

func TestContextPersistence(t *testing.T) {
	_, cleanup := testutil.TempHomeDir(t)
	defer cleanup()

	// Initialize database and set context
	manager1 := NewDatabaseManager()
	err := manager1.Initialize(GlobalMode)
	require.NoError(t, err)

	registry := NewProjectRegistry(manager1)
	project := &Project{Name: "Persistent Project", Path: "/persist/path"}
	err = registry.Register(project)
	require.NoError(t, err)

	ctx1 := NewContextManager(manager1)
	err = ctx1.SwitchTo(project.ID)
	require.NoError(t, err)

	// Close first manager
	manager1.Close()

	// Open new manager and context
	manager2 := NewDatabaseManager()
	err = manager2.Initialize(GlobalMode)
	require.NoError(t, err)
	defer manager2.Close()

	ctx2 := NewContextManager(manager2)

	// Context should be persisted
	current, err := ctx2.GetCurrent()
	require.NoError(t, err)
	assert.Equal(t, project.ID, current.ID)
}

func TestGetCurrentNoContext(t *testing.T) {
	_, cleanup := testutil.TempHomeDir(t)
	defer cleanup()

	manager := NewDatabaseManager()
	err := manager.Initialize(GlobalMode)
	require.NoError(t, err)
	defer manager.Close()

	ctx := NewContextManager(manager)

	// No context set yet
	_, err = ctx.GetCurrent()
	assert.Error(t, err)
}

func TestSetActiveFlag(t *testing.T) {
	_, cleanup := testutil.TempHomeDir(t)
	defer cleanup()

	manager := NewDatabaseManager()
	err := manager.Initialize(GlobalMode)
	require.NoError(t, err)
	defer manager.Close()

	registry := NewProjectRegistry(manager)
	ctx := NewContextManager(manager)

	// Register projects
	project1 := &Project{Name: "Project 1", Path: "/path/1"}
	err = registry.Register(project1)
	require.NoError(t, err)

	project2 := &Project{Name: "Project 2", Path: "/path/2"}
	err = registry.Register(project2)
	require.NoError(t, err)

	// Switch to project 1
	err = ctx.SwitchTo(project1.ID)
	require.NoError(t, err)

	// Verify only project 1 is active
	p1, err := registry.GetByID(project1.ID)
	require.NoError(t, err)
	assert.True(t, p1.Active)

	p2, err := registry.GetByID(project2.ID)
	require.NoError(t, err)
	assert.False(t, p2.Active)

	// Switch to project 2
	err = ctx.SwitchTo(project2.ID)
	require.NoError(t, err)

	// Verify only project 2 is active now
	p1, err = registry.GetByID(project1.ID)
	require.NoError(t, err)
	assert.False(t, p1.Active)

	p2, err = registry.GetByID(project2.ID)
	require.NoError(t, err)
	assert.True(t, p2.Active)
}

func TestDetectProjectWithLocalDatabase(t *testing.T) {
	_, cleanup := testutil.TempHomeDir(t)
	defer cleanup()

	// Create project directory
	projectDir, cleanupProject := testutil.TempDir(t)
	defer cleanupProject()

	// Change to project directory
	restoreDir := testutil.Chdir(t, projectDir)
	defer restoreDir()

	// Initialize local database
	manager := NewDatabaseManager()
	err := manager.Initialize(LocalMode)
	require.NoError(t, err)
	defer manager.Close()

	registry := NewProjectRegistry(manager)

	// Register project in local database
	project := &Project{
		Name: "Local Project",
		Path: projectDir,
	}
	err = registry.Register(project)
	require.NoError(t, err)

	ctx := NewContextManager(manager)

	// Should detect from local database
	detected, err := ctx.DetectProject()
	require.NoError(t, err)
	assert.Equal(t, project.ID, detected.ID)
}

func TestGetCurrentProject(t *testing.T) {
	_, cleanup := testutil.TempHomeDir(t)
	defer cleanup()

	manager := NewDatabaseManager()
	err := manager.Initialize(GlobalMode)
	require.NoError(t, err)
	defer manager.Close()

	registry := NewProjectRegistry(manager)
	ctx := NewContextManager(manager)

	// Register and activate project
	project := &Project{Name: "Current Project", Path: "/current/path"}
	err = registry.Register(project)
	require.NoError(t, err)

	err = ctx.SwitchTo(project.ID)
	require.NoError(t, err)

	// Get current project
	current, err := ctx.GetCurrent()
	require.NoError(t, err)
	assert.Equal(t, project.ID, current.ID)
	assert.Equal(t, project.Name, current.Name)
	assert.Equal(t, project.Path, current.Path)
	assert.True(t, current.Active)
}
