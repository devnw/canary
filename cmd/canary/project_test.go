// Copyright (c) 2025 by Developer Network.
//
// For more details, see the LICENSE file in the root directory of this
// source code repository or contact Developer Network at info@devnw.com.

// CANARY: REQ=CBIN-146; FEATURE="ProjectCLI"; ASPECT=CLI; STATUS=IMPL; TEST=TestProjectCLI; UPDATED=2025-10-18
package main

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.devnw.com/canary/internal/storage"
)

func TestDBInitGlobal(t *testing.T) {
	// Setup temp home directory
	originalHome := os.Getenv("HOME")
	tmpHome, err := os.MkdirTemp("", "canary-test-home-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpHome)
	os.Setenv("HOME", tmpHome)
	defer os.Setenv("HOME", originalHome)

	// Execute db init --global
	cmd := newRootCmd()
	cmd.SetArgs([]string{"db", "init", "--global"})

	var stdout bytes.Buffer
	cmd.SetOut(&stdout)

	err = cmd.Execute()
	require.NoError(t, err)

	// Verify global database was created
	globalDBPath := filepath.Join(tmpHome, ".canary", "canary.db")
	_, err = os.Stat(globalDBPath)
	assert.NoError(t, err, "global database should exist")

	// Verify output message
	output := stdout.String()
	assert.Contains(t, output, "global")
	assert.Contains(t, output, globalDBPath)
}

func TestDBInitLocal(t *testing.T) {
	// Setup temp directory
	tmpDir, err := os.MkdirTemp("", "canary-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	// Change to temp directory
	originalDir, err := os.Getwd()
	require.NoError(t, err)
	err = os.Chdir(tmpDir)
	require.NoError(t, err)
	defer os.Chdir(originalDir)

	// Execute db init --local
	cmd := newRootCmd()
	cmd.SetArgs([]string{"db", "init", "--local"})

	var stdout bytes.Buffer
	cmd.SetOut(&stdout)

	err = cmd.Execute()
	require.NoError(t, err)

	// Verify local database was created
	localDBPath := filepath.Join(tmpDir, ".canary", "canary.db")
	_, err = os.Stat(localDBPath)
	assert.NoError(t, err, "local database should exist")

	// Verify output message
	output := stdout.String()
	assert.Contains(t, output, "local")
	assert.Contains(t, output, ".canary/canary.db")
}

func TestDBInitDefault(t *testing.T) {
	// Setup temp home directory
	originalHome := os.Getenv("HOME")
	tmpHome, err := os.MkdirTemp("", "canary-test-home-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpHome)
	os.Setenv("HOME", tmpHome)
	defer os.Setenv("HOME", originalHome)

	// Execute db init (no flags - should default to global)
	cmd := newRootCmd()
	cmd.SetArgs([]string{"db", "init"})

	err = cmd.Execute()
	require.NoError(t, err)

	// Verify global database was created (default behavior)
	globalDBPath := filepath.Join(tmpHome, ".canary", "canary.db")
	_, err = os.Stat(globalDBPath)
	assert.NoError(t, err, "should create global database by default")
}

func TestProjectRegister(t *testing.T) {
	// Setup temp home and database
	originalHome := os.Getenv("HOME")
	tmpHome, err := os.MkdirTemp("", "canary-test-home-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpHome)
	os.Setenv("HOME", tmpHome)
	defer os.Setenv("HOME", originalHome)

	// Change to temp directory to avoid finding local .canary database
	originalDir, err := os.Getwd()
	require.NoError(t, err)
	err = os.Chdir(tmpHome)
	require.NoError(t, err)
	defer os.Chdir(originalDir)

	// Initialize global database
	manager := storage.NewDatabaseManager()
	err = manager.Initialize(storage.GlobalMode)
	require.NoError(t, err)
	manager.Close() // Close before running command

	// Execute project register with unique path
	uniquePath := filepath.Join(tmpHome, "test-proj-reg")
	cmd := newRootCmd()
	cmd.SetArgs([]string{"project", "register", "Test Project", uniquePath})

	var stdout bytes.Buffer
	cmd.SetOut(&stdout)

	err = cmd.Execute()
	require.NoError(t, err)

	// Verify project was registered (reopen database)
	manager2 := storage.NewDatabaseManager()
	err = manager2.Discover()
	require.NoError(t, err)
	defer manager2.Close()

	registry := storage.NewProjectRegistry(manager2)
	projects, err := registry.List()
	require.NoError(t, err)
	assert.Len(t, projects, 1)
	assert.Equal(t, "Test Project", projects[0].Name)
	assert.Equal(t, uniquePath, projects[0].Path)

	// Verify output
	output := stdout.String()
	assert.Contains(t, output, "Test Project")
	assert.Contains(t, output, "test-project") // slug
}

func TestProjectList(t *testing.T) {
	// Setup temp home and database
	originalHome := os.Getenv("HOME")
	tmpHome, err := os.MkdirTemp("", "canary-test-home-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpHome)
	os.Setenv("HOME", tmpHome)
	defer os.Setenv("HOME", originalHome)

	// Change to temp directory to avoid finding local .canary database
	originalDir, err := os.Getwd()
	require.NoError(t, err)
	err = os.Chdir(tmpHome)
	require.NoError(t, err)
	defer os.Chdir(originalDir)

	// Initialize and populate database
	manager := storage.NewDatabaseManager()
	err = manager.Initialize(storage.GlobalMode)
	require.NoError(t, err)

	registry := storage.NewProjectRegistry(manager)
	project1 := &storage.Project{Name: "Project 1", Path: filepath.Join(tmpHome, "proj1")}
	project2 := &storage.Project{Name: "Project 2", Path: filepath.Join(tmpHome, "proj2")}
	err = registry.Register(project1)
	require.NoError(t, err)
	err = registry.Register(project2)
	require.NoError(t, err)
	manager.Close() // Close before running command

	// Execute project list
	cmd := newRootCmd()
	cmd.SetArgs([]string{"project", "list"})

	var stdout bytes.Buffer
	cmd.SetOut(&stdout)

	err = cmd.Execute()
	require.NoError(t, err)

	// Verify output contains both projects
	output := stdout.String()
	assert.Contains(t, output, "Project 1")
	assert.Contains(t, output, "Project 2")
	assert.Contains(t, output, "proj1")
	assert.Contains(t, output, "proj2")
}

func TestProjectRemove(t *testing.T) {
	// Setup temp home and database
	originalHome := os.Getenv("HOME")
	tmpHome, err := os.MkdirTemp("", "canary-test-home-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpHome)
	os.Setenv("HOME", tmpHome)
	defer os.Setenv("HOME", originalHome)

	// Change to temp directory to avoid finding local .canary database
	originalDir, err := os.Getwd()
	require.NoError(t, err)
	err = os.Chdir(tmpHome)
	require.NoError(t, err)
	defer os.Chdir(originalDir)

	// Initialize and populate database
	manager := storage.NewDatabaseManager()
	err = manager.Initialize(storage.GlobalMode)
	require.NoError(t, err)

	registry := storage.NewProjectRegistry(manager)
	project := &storage.Project{Name: "Test Project", Path: filepath.Join(tmpHome, "test-rm")}
	err = registry.Register(project)
	require.NoError(t, err)
	projectID := project.ID
	manager.Close() // Close before running command

	// Execute project remove
	cmd := newRootCmd()
	cmd.SetArgs([]string{"project", "remove", projectID})

	var stdout bytes.Buffer
	cmd.SetOut(&stdout)

	err = cmd.Execute()
	require.NoError(t, err)

	// Verify project was removed (reopen database)
	manager2 := storage.NewDatabaseManager()
	err = manager2.Discover()
	require.NoError(t, err)
	defer manager2.Close()

	registry2 := storage.NewProjectRegistry(manager2)
	projects, err := registry2.List()
	require.NoError(t, err)
	assert.Len(t, projects, 0)

	// Verify output
	output := stdout.String()
	assert.Contains(t, output, "removed")
}

func TestProjectSwitch(t *testing.T) {
	// Setup temp home and database
	originalHome := os.Getenv("HOME")
	tmpHome, err := os.MkdirTemp("", "canary-test-home-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpHome)
	os.Setenv("HOME", tmpHome)
	defer os.Setenv("HOME", originalHome)

	// Change to temp directory to avoid finding local .canary database
	originalDir, err := os.Getwd()
	require.NoError(t, err)
	err = os.Chdir(tmpHome)
	require.NoError(t, err)
	defer os.Chdir(originalDir)

	// Initialize and populate database
	manager := storage.NewDatabaseManager()
	err = manager.Initialize(storage.GlobalMode)
	require.NoError(t, err)

	registry := storage.NewProjectRegistry(manager)
	project1 := &storage.Project{Name: "Project 1", Path: filepath.Join(tmpHome, "sw-proj1")}
	project2 := &storage.Project{Name: "Project 2", Path: filepath.Join(tmpHome, "sw-proj2")}
	err = registry.Register(project1)
	require.NoError(t, err)
	err = registry.Register(project2)
	require.NoError(t, err)
	project2ID := project2.ID
	manager.Close() // Close before running command

	// Execute project switch
	cmd := newRootCmd()
	cmd.SetArgs([]string{"project", "switch", project2ID})

	var stdout bytes.Buffer
	cmd.SetOut(&stdout)

	err = cmd.Execute()
	require.NoError(t, err)

	// Verify project 2 is now active (reopen database)
	manager2 := storage.NewDatabaseManager()
	err = manager2.Discover()
	require.NoError(t, err)
	defer manager2.Close()

	ctx := storage.NewContextManager(manager2)
	current, err := ctx.GetCurrent()
	require.NoError(t, err)
	assert.Equal(t, project2ID, current.ID)

	// Verify output
	output := stdout.String()
	assert.Contains(t, output, "Project 2")
}

func TestProjectCurrent(t *testing.T) {
	// Setup temp home and database
	originalHome := os.Getenv("HOME")
	tmpHome, err := os.MkdirTemp("", "canary-test-home-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpHome)
	os.Setenv("HOME", tmpHome)
	defer os.Setenv("HOME", originalHome)

	// Change to temp directory to avoid finding local .canary database
	originalDir, err := os.Getwd()
	require.NoError(t, err)
	err = os.Chdir(tmpHome)
	require.NoError(t, err)
	defer os.Chdir(originalDir)

	// Initialize and populate database
	manager := storage.NewDatabaseManager()
	err = manager.Initialize(storage.GlobalMode)
	require.NoError(t, err)

	registry := storage.NewProjectRegistry(manager)
	project := &storage.Project{Name: "Active Project", Path: filepath.Join(tmpHome, "active")}
	err = registry.Register(project)
	require.NoError(t, err)

	ctx := storage.NewContextManager(manager)
	err = ctx.SwitchTo(project.ID)
	require.NoError(t, err)
	projectID := project.ID
	manager.Close() // Close before running command

	// Execute project current
	cmd := newRootCmd()
	cmd.SetArgs([]string{"project", "current"})

	var stdout bytes.Buffer
	cmd.SetOut(&stdout)

	err = cmd.Execute()
	require.NoError(t, err)

	// Verify output shows active project
	output := stdout.String()
	assert.Contains(t, output, "Active Project")
	assert.Contains(t, output, "active")
	assert.Contains(t, output, projectID)
}

func TestProjectCurrentNoContext(t *testing.T) {
	// Setup temp home and database
	originalHome := os.Getenv("HOME")
	tmpHome, err := os.MkdirTemp("", "canary-test-home-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpHome)
	os.Setenv("HOME", tmpHome)
	defer os.Setenv("HOME", originalHome)

	// Initialize database with no active project
	manager := storage.NewDatabaseManager()
	err = manager.Initialize(storage.GlobalMode)
	require.NoError(t, err)
	defer manager.Close()

	// Execute project current
	cmd := newRootCmd()
	cmd.SetArgs([]string{"project", "current"})

	var stdout bytes.Buffer
	cmd.SetOut(&stdout)

	err = cmd.Execute()
	// Command should succeed but show "No active project"
	require.NoError(t, err)
	output := stdout.String()
	assert.Contains(t, output, "No active project")
}

// Helper to create a new root command for testing
func newRootCmd() *cobra.Command {
	// Return a fresh instance of rootCmd for each test
	// This prevents test interference
	root := &cobra.Command{
		Use:   "canary",
		Short: "Track requirements via CANARY tokens",
	}

	// Add the commands we're testing (will be defined in GREEN phase)
	root.AddCommand(newDBCmd())
	root.AddCommand(newProjectCmd())

	return root
}

// Command constructors are now in project.go
