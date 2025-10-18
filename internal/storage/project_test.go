// Copyright (c) 2025 by Developer Network.
//
// For more details, see the LICENSE file in the root directory of this
// source code repository or contact Developer Network at info@devnw.com.

// CANARY: REQ=CBIN-146; FEATURE="ProjectRegistry"; ASPECT=Storage; STATUS=IMPL; TEST=TestProjectRegistry; UPDATED=2025-10-18
package storage

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.devnw.com/canary/internal/storage/testutil"
)

func TestRegisterProject(t *testing.T) {
	_, cleanup := testutil.TempHomeDir(t)
	defer cleanup()

	// Initialize database
	manager := NewDatabaseManager()
	err := manager.Initialize(GlobalMode)
	require.NoError(t, err)
	defer manager.Close()

	// Create project registry
	registry := NewProjectRegistry(manager)

	// Register a project
	project := &Project{
		Name: "Test Project",
		Path: "/path/to/project",
	}

	err = registry.Register(project)
	require.NoError(t, err)

	// Verify project was created with ID
	assert.NotEmpty(t, project.ID)
	assert.Equal(t, "test-project", project.ID) // slug from name
	assert.NotEmpty(t, project.CreatedAt)
}

func TestRegisterProjectDuplicate(t *testing.T) {
	_, cleanup := testutil.TempHomeDir(t)
	defer cleanup()

	manager := NewDatabaseManager()
	err := manager.Initialize(GlobalMode)
	require.NoError(t, err)
	defer manager.Close()

	registry := NewProjectRegistry(manager)

	// Register first project
	project1 := &Project{
		Name: "My Project",
		Path: "/path/to/project1",
	}
	err = registry.Register(project1)
	require.NoError(t, err)

	// Try to register project with same path - should fail
	project2 := &Project{
		Name: "Another Project",
		Path: "/path/to/project1", // Same path
	}
	err = registry.Register(project2)
	assert.Error(t, err)
}

func TestListProjects(t *testing.T) {
	_, cleanup := testutil.TempHomeDir(t)
	defer cleanup()

	manager := NewDatabaseManager()
	err := manager.Initialize(GlobalMode)
	require.NoError(t, err)
	defer manager.Close()

	registry := NewProjectRegistry(manager)

	// Register multiple projects
	projects := []*Project{
		{Name: "Project A", Path: "/path/a"},
		{Name: "Project B", Path: "/path/b"},
		{Name: "Project C", Path: "/path/c"},
	}

	for _, p := range projects {
		err = registry.Register(p)
		require.NoError(t, err)
	}

	// List all projects
	listed, err := registry.List()
	require.NoError(t, err)
	assert.Len(t, listed, 3)

	// Verify all projects are present
	ids := make(map[string]bool)
	for _, p := range listed {
		ids[p.ID] = true
	}
	assert.True(t, ids["project-a"])
	assert.True(t, ids["project-b"])
	assert.True(t, ids["project-c"])
}

func TestRemoveProject(t *testing.T) {
	_, cleanup := testutil.TempHomeDir(t)
	defer cleanup()

	manager := NewDatabaseManager()
	err := manager.Initialize(GlobalMode)
	require.NoError(t, err)
	defer manager.Close()

	registry := NewProjectRegistry(manager)

	// Register a project
	project := &Project{
		Name: "To Remove",
		Path: "/path/to/remove",
	}
	err = registry.Register(project)
	require.NoError(t, err)

	// Verify it exists
	listed, err := registry.List()
	require.NoError(t, err)
	assert.Len(t, listed, 1)

	// Remove the project
	err = registry.Remove(project.ID)
	require.NoError(t, err)

	// Verify it's gone
	listed, err = registry.List()
	require.NoError(t, err)
	assert.Len(t, listed, 0)
}

func TestRemoveNonexistentProject(t *testing.T) {
	_, cleanup := testutil.TempHomeDir(t)
	defer cleanup()

	manager := NewDatabaseManager()
	err := manager.Initialize(GlobalMode)
	require.NoError(t, err)
	defer manager.Close()

	registry := NewProjectRegistry(manager)

	// Try to remove non-existent project
	err = registry.Remove("nonexistent-id")
	assert.Error(t, err)
}

func TestProjectSlugGeneration(t *testing.T) {
	_, cleanup := testutil.TempHomeDir(t)
	defer cleanup()

	manager := NewDatabaseManager()
	err := manager.Initialize(GlobalMode)
	require.NoError(t, err)
	defer manager.Close()

	registry := NewProjectRegistry(manager)

	tests := []struct {
		name         string
		projectName  string
		expectedSlug string
	}{
		{
			name:         "simple name",
			projectName:  "MyProject",
			expectedSlug: "myproject",
		},
		{
			name:         "name with spaces",
			projectName:  "My Cool Project",
			expectedSlug: "my-cool-project",
		},
		{
			name:         "name with special chars",
			projectName:  "Project: V2.0",
			expectedSlug: "project-v2-0",
		},
		{
			name:         "name with underscores",
			projectName:  "my_awesome_project",
			expectedSlug: "my-awesome-project",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			project := &Project{
				Name: tt.projectName,
				Path: "/path/" + tt.projectName,
			}
			err := registry.Register(project)
			require.NoError(t, err)
			assert.Equal(t, tt.expectedSlug, project.ID)
		})
	}
}

func TestProjectSlugCollision(t *testing.T) {
	_, cleanup := testutil.TempHomeDir(t)
	defer cleanup()

	manager := NewDatabaseManager()
	err := manager.Initialize(GlobalMode)
	require.NoError(t, err)
	defer manager.Close()

	registry := NewProjectRegistry(manager)

	// Register first project
	project1 := &Project{
		Name: "My Project",
		Path: "/path/1",
	}
	err = registry.Register(project1)
	require.NoError(t, err)
	assert.Equal(t, "my-project", project1.ID)

	// Register second project with same name (different path)
	project2 := &Project{
		Name: "My Project", // Same name
		Path: "/path/2",    // Different path
	}
	err = registry.Register(project2)
	require.NoError(t, err)
	assert.Equal(t, "my-project-2", project2.ID) // Should have counter appended

	// Register third project with same name
	project3 := &Project{
		Name: "My Project",
		Path: "/path/3",
	}
	err = registry.Register(project3)
	require.NoError(t, err)
	assert.Equal(t, "my-project-3", project3.ID)
}

func TestGetProjectByID(t *testing.T) {
	_, cleanup := testutil.TempHomeDir(t)
	defer cleanup()

	manager := NewDatabaseManager()
	err := manager.Initialize(GlobalMode)
	require.NoError(t, err)
	defer manager.Close()

	registry := NewProjectRegistry(manager)

	// Register a project
	project := &Project{
		Name: "Test Project",
		Path: "/test/path",
	}
	err = registry.Register(project)
	require.NoError(t, err)

	// Get by ID
	retrieved, err := registry.GetByID(project.ID)
	require.NoError(t, err)
	assert.Equal(t, project.ID, retrieved.ID)
	assert.Equal(t, project.Name, retrieved.Name)
	assert.Equal(t, project.Path, retrieved.Path)
}

func TestGetProjectByPath(t *testing.T) {
	_, cleanup := testutil.TempHomeDir(t)
	defer cleanup()

	manager := NewDatabaseManager()
	err := manager.Initialize(GlobalMode)
	require.NoError(t, err)
	defer manager.Close()

	registry := NewProjectRegistry(manager)

	// Register a project
	projectPath := "/unique/test/path"
	project := &Project{
		Name: "Test Project",
		Path: projectPath,
	}
	err = registry.Register(project)
	require.NoError(t, err)

	// Get by path
	retrieved, err := registry.GetByPath(projectPath)
	require.NoError(t, err)
	assert.Equal(t, project.ID, retrieved.ID)
	assert.Equal(t, project.Name, retrieved.Name)
	assert.Equal(t, project.Path, retrieved.Path)
}

func TestProjectTimestamps(t *testing.T) {
	_, cleanup := testutil.TempHomeDir(t)
	defer cleanup()

	manager := NewDatabaseManager()
	err := manager.Initialize(GlobalMode)
	require.NoError(t, err)
	defer manager.Close()

	registry := NewProjectRegistry(manager)

	// Register a project
	beforeCreate := time.Now().UTC()
	project := &Project{
		Name: "Time Test",
		Path: "/time/test",
	}
	err = registry.Register(project)
	require.NoError(t, err)
	afterCreate := time.Now().UTC()

	// Verify timestamp is within reasonable range
	assert.NotEmpty(t, project.CreatedAt)
	createdTime, err := time.Parse(time.RFC3339, project.CreatedAt)
	require.NoError(t, err)

	// Use truncation to second precision for comparison
	beforeTrunc := beforeCreate.Truncate(time.Second)
	afterTrunc := afterCreate.Add(time.Second).Truncate(time.Second)
	createdTrunc := createdTime.Truncate(time.Second)

	assert.True(t, createdTrunc.After(beforeTrunc) || createdTrunc.Equal(beforeTrunc))
	assert.True(t, createdTrunc.Before(afterTrunc) || createdTrunc.Equal(afterTrunc))
}
