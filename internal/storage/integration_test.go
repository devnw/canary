// Copyright (c) 2025 by Developer Network.
//
// For more details, see the LICENSE file in the root directory of this
// source code repository or contact Developer Network at info@devnw.com.

// CANARY: REQ=CBIN-146; FEATURE="IntegrationTests"; ASPECT=Storage; STATUS=IMPL; TEST=TestIntegration; UPDATED=2025-10-18
package storage

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.devnw.com/canary/internal/storage/testutil"
)

// TestMultiProjectWorkflow tests a complete multi-project workflow
func TestMultiProjectWorkflow(t *testing.T) {
	_, cleanup := testutil.TempHomeDir(t)
	defer cleanup()

	// Step 1: Initialize global database
	manager := NewDatabaseManager()
	err := manager.Initialize(GlobalMode)
	require.NoError(t, err)
	dbPath := manager.Location()
	manager.Close()

	// Step 2: Register multiple projects
	manager2 := NewDatabaseManager()
	err = manager2.Discover()
	require.NoError(t, err)

	registry := NewProjectRegistry(manager2)

	project1 := &Project{Name: "Frontend", Path: "/apps/frontend"}
	err = registry.Register(project1)
	require.NoError(t, err)

	project2 := &Project{Name: "Backend API", Path: "/apps/backend"}
	err = registry.Register(project2)
	require.NoError(t, err)

	project3 := &Project{Name: "Mobile App", Path: "/apps/mobile"}
	err = registry.Register(project3)
	require.NoError(t, err)

	manager2.Close()

	// Step 3: Switch context to project1 and add tokens
	manager3 := NewDatabaseManager()
	err = manager3.Discover()
	require.NoError(t, err)

	ctx := NewContextManager(manager3)
	err = ctx.SwitchTo(project1.ID)
	require.NoError(t, err)

	db := &DB{conn: manager3.conn, path: dbPath}

	// Add tokens for project1
	token1 := &Token{
		ReqID:      "CBIN-201",
		Feature:    "UserAuth",
		Aspect:     "API",
		Status:     "IMPL",
		FilePath:   "/apps/frontend/auth.js",
		LineNumber: 10,
		UpdatedAt:  "2025-10-18",
		RawToken:   "// CANARY: REQ=CBIN-201",
		IndexedAt:  "2025-10-18",
		ProjectID:  project1.ID,
	}
	err = db.UpsertToken(token1)
	require.NoError(t, err)

	manager3.Close()

	// Step 4: Switch to project2 and add different tokens
	manager4 := NewDatabaseManager()
	err = manager4.Discover()
	require.NoError(t, err)

	ctx2 := NewContextManager(manager4)
	err = ctx2.SwitchTo(project2.ID)
	require.NoError(t, err)

	db2 := &DB{conn: manager4.conn, path: dbPath}

	// Add tokens for project2 (same req_id, different project)
	token2 := &Token{
		ReqID:      "CBIN-201", // Same ID as project1
		Feature:    "UserAuth",
		Aspect:     "Storage",
		Status:     "TESTED",
		FilePath:   "/apps/backend/auth.go",
		LineNumber: 50,
		UpdatedAt:  "2025-10-18",
		RawToken:   "// CANARY: REQ=CBIN-201",
		IndexedAt:  "2025-10-18",
		ProjectID:  project2.ID,
	}
	err = db2.UpsertToken(token2)
	require.NoError(t, err)

	manager4.Close()

	// Step 5: Verify isolation - tokens are separated by project
	manager5 := NewDatabaseManager()
	err = manager5.Discover()
	require.NoError(t, err)
	defer manager5.Close()

	db3 := &DB{conn: manager5.conn, path: dbPath}

	// Get tokens for project1
	tokens1, err := db3.GetTokensByProject(project1.ID)
	require.NoError(t, err)
	assert.Len(t, tokens1, 1)
	assert.Equal(t, "API", tokens1[0].Aspect)
	assert.Equal(t, "IMPL", tokens1[0].Status)

	// Get tokens for project2
	tokens2, err := db3.GetTokensByProject(project2.ID)
	require.NoError(t, err)
	assert.Len(t, tokens2, 1)
	assert.Equal(t, "Storage", tokens2[0].Aspect)
	assert.Equal(t, "TESTED", tokens2[0].Status)

	// Verify no tokens for project3
	tokens3, err := db3.GetTokensByProject(project3.ID)
	require.NoError(t, err)
	assert.Len(t, tokens3, 0)

	// Step 6: Verify cross-project query
	allTokens, err := db3.GetAllTokens()
	require.NoError(t, err)
	assert.Len(t, allTokens, 2)

	// Step 7: Verify context persistence
	ctx3 := NewContextManager(manager5)
	current, err := ctx3.GetCurrent()
	require.NoError(t, err)
	assert.Equal(t, project2.ID, current.ID, "should remember last active project")
}

// TestLocalAndGlobalDatabaseCoexistence tests that local and global databases work independently
func TestLocalAndGlobalDatabaseCoexistence(t *testing.T) {
	_, cleanup := testutil.TempHomeDir(t)
	defer cleanup()

	// Create temp project directory
	projectDir, cleanupProject := testutil.TempDir(t)
	defer cleanupProject()

	// Step 1: Initialize global database
	managerGlobal := NewDatabaseManager()
	err := managerGlobal.Initialize(GlobalMode)
	require.NoError(t, err)
	globalPath := managerGlobal.Location()
	managerGlobal.Close()

	// Step 2: Change to project directory and initialize local database
	restoreDir := testutil.Chdir(t, projectDir)
	defer restoreDir()

	managerLocal := NewDatabaseManager()
	err = managerLocal.Initialize(LocalMode)
	require.NoError(t, err)
	localPath := managerLocal.Location()
	managerLocal.Close()

	// Verify paths are different
	assert.NotEqual(t, globalPath, localPath)
	assert.Contains(t, globalPath, ".canary/canary.db")
	assert.Contains(t, localPath, ".canary/canary.db")

	// Step 3: Register project in global database
	managerGlobal2 := NewDatabaseManager()
	err = managerGlobal2.Initialize(GlobalMode)
	require.NoError(t, err)

	registryGlobal := NewProjectRegistry(managerGlobal2)
	globalProject := &Project{Name: "Global Project", Path: "/global/path"}
	err = registryGlobal.Register(globalProject)
	require.NoError(t, err)
	managerGlobal2.Close()

	// Step 4: Register different project in local database
	managerLocal2 := NewDatabaseManager()
	err = managerLocal2.Initialize(LocalMode)
	require.NoError(t, err)

	registryLocal := NewProjectRegistry(managerLocal2)
	localProject := &Project{Name: "Local Project", Path: projectDir}
	err = registryLocal.Register(localProject)
	require.NoError(t, err)
	managerLocal2.Close()

	// Step 5: Discover should prefer local when in project directory
	managerDiscover := NewDatabaseManager()
	err = managerDiscover.Discover()
	require.NoError(t, err)
	defer managerDiscover.Close()

	assert.Equal(t, LocalMode, managerDiscover.Mode())
	assert.Equal(t, localPath, managerDiscover.Location())

	// Step 6: Verify local database has only local project
	registryDiscovered := NewProjectRegistry(managerDiscover)
	projects, err := registryDiscovered.List()
	require.NoError(t, err)
	assert.Len(t, projects, 1)
	assert.Equal(t, "Local Project", projects[0].Name)
}

// TestProjectContextSwitchingWithTokens tests switching between projects and token visibility
func TestProjectContextSwitchingWithTokens(t *testing.T) {
	_, cleanup := testutil.TempHomeDir(t)
	defer cleanup()

	// Initialize database
	manager := NewDatabaseManager()
	err := manager.Initialize(GlobalMode)
	require.NoError(t, err)
	defer manager.Close()

	registry := NewProjectRegistry(manager)
	db := &DB{conn: manager.conn, path: manager.path}
	ctx := NewContextManager(manager)

	// Create two projects
	projectA := &Project{Name: "Project A", Path: "/projects/a"}
	err = registry.Register(projectA)
	require.NoError(t, err)

	projectB := &Project{Name: "Project B", Path: "/projects/b"}
	err = registry.Register(projectB)
	require.NoError(t, err)

	// Switch to Project A and add tokens
	err = ctx.SwitchTo(projectA.ID)
	require.NoError(t, err)

	for i := 0; i < 5; i++ {
		token := &Token{
			ReqID:      "CBIN-300",
			Feature:    "SharedFeature",
			Aspect:     "API",
			Status:     "IMPL",
			FilePath:   "/projects/a/file.go",
			LineNumber: i * 10,
			UpdatedAt:  "2025-10-18",
			RawToken:   "test",
			IndexedAt:  "2025-10-18",
			ProjectID:  projectA.ID,
		}
		err = db.UpsertToken(token)
		require.NoError(t, err)
	}

	// Switch to Project B and add tokens
	err = ctx.SwitchTo(projectB.ID)
	require.NoError(t, err)

	for i := 0; i < 3; i++ {
		token := &Token{
			ReqID:      "CBIN-300", // Same req_id
			Feature:    "SharedFeature",
			Aspect:     "Storage",
			Status:     "TESTED",
			FilePath:   "/projects/b/file.go",
			LineNumber: i * 10,
			UpdatedAt:  "2025-10-18",
			RawToken:   "test",
			IndexedAt:  "2025-10-18",
			ProjectID:  projectB.ID,
		}
		err = db.UpsertToken(token)
		require.NoError(t, err)
	}

	// Verify token isolation
	tokensA, err := db.GetTokensByProject(projectA.ID)
	require.NoError(t, err)
	assert.Len(t, tokensA, 5)

	tokensB, err := db.GetTokensByProject(projectB.ID)
	require.NoError(t, err)
	assert.Len(t, tokensB, 3)

	// Verify project-scoped query by req_id
	tokensAByReq, err := db.GetTokensByReqIDAndProject("CBIN-300", projectA.ID)
	require.NoError(t, err)
	assert.Len(t, tokensAByReq, 5)
	assert.Equal(t, "API", tokensAByReq[0].Aspect)

	tokensBByReq, err := db.GetTokensByReqIDAndProject("CBIN-300", projectB.ID)
	require.NoError(t, err)
	assert.Len(t, tokensBByReq, 3)
	assert.Equal(t, "Storage", tokensBByReq[0].Aspect)

	// Verify current context is Project B
	current, err := ctx.GetCurrent()
	require.NoError(t, err)
	assert.Equal(t, projectB.ID, current.ID)
	assert.True(t, current.Active)
}

// TestDatabasePersistenceAcrossRestarts tests that data persists correctly
func TestDatabasePersistenceAcrossRestarts(t *testing.T) {
	_, cleanup := testutil.TempHomeDir(t)
	defer cleanup()

	// Step 1: Create and populate database
	manager1 := NewDatabaseManager()
	err := manager1.Initialize(GlobalMode)
	require.NoError(t, err)
	dbPath := manager1.Location()

	registry1 := NewProjectRegistry(manager1)
	project := &Project{Name: "Persistent Project", Path: "/persist/path"}
	err = registry1.Register(project)
	require.NoError(t, err)
	projectID := project.ID

	ctx1 := NewContextManager(manager1)
	err = ctx1.SwitchTo(projectID)
	require.NoError(t, err)

	db1 := &DB{conn: manager1.conn, path: dbPath}
	token := &Token{
		ReqID:      "CBIN-400",
		Feature:    "Persistence",
		Aspect:     "Storage",
		Status:     "TESTED",
		FilePath:   "/file.go",
		LineNumber: 100,
		UpdatedAt:  "2025-10-18",
		RawToken:   "test",
		IndexedAt:  "2025-10-18",
		ProjectID:  projectID,
	}
	err = db1.UpsertToken(token)
	require.NoError(t, err)

	manager1.Close()

	// Sleep to ensure file system sync
	time.Sleep(100 * time.Millisecond)

	// Step 2: Reopen database and verify data persists
	manager2 := NewDatabaseManager()
	err = manager2.Discover()
	require.NoError(t, err)
	defer manager2.Close()

	// Verify project exists
	registry2 := NewProjectRegistry(manager2)
	projects, err := registry2.List()
	require.NoError(t, err)
	assert.Len(t, projects, 1)
	assert.Equal(t, "Persistent Project", projects[0].Name)
	assert.Equal(t, projectID, projects[0].ID)

	// Verify context persists
	ctx2 := NewContextManager(manager2)
	current, err := ctx2.GetCurrent()
	require.NoError(t, err)
	assert.Equal(t, projectID, current.ID)
	assert.True(t, current.Active)

	// Verify token persists
	db2 := &DB{conn: manager2.conn, path: manager2.path}
	tokens, err := db2.GetTokensByProject(projectID)
	require.NoError(t, err)
	assert.Len(t, tokens, 1)
	assert.Equal(t, "CBIN-400", tokens[0].ReqID)
	assert.Equal(t, "Persistence", tokens[0].Feature)
}

// TestCompleteProjectLifecycle tests full CRUD operations on projects
func TestCompleteProjectLifecycle(t *testing.T) {
	_, cleanup := testutil.TempHomeDir(t)
	defer cleanup()

	manager := NewDatabaseManager()
	err := manager.Initialize(GlobalMode)
	require.NoError(t, err)
	defer manager.Close()

	registry := NewProjectRegistry(manager)
	db := &DB{conn: manager.conn, path: manager.path}

	// Create
	project := &Project{Name: "Lifecycle Test", Path: "/test/lifecycle"}
	err = registry.Register(project)
	require.NoError(t, err)
	projectID := project.ID

	// Add tokens
	token := &Token{
		ReqID:      "CBIN-500",
		Feature:    "Lifecycle",
		Aspect:     "API",
		Status:     "IMPL",
		FilePath:   "/file.go",
		LineNumber: 50,
		UpdatedAt:  "2025-10-18",
		RawToken:   "test",
		IndexedAt:  "2025-10-18",
		ProjectID:  projectID,
	}
	err = db.UpsertToken(token)
	require.NoError(t, err)

	// Read
	retrieved, err := registry.GetByID(projectID)
	require.NoError(t, err)
	assert.Equal(t, "Lifecycle Test", retrieved.Name)

	tokens, err := db.GetTokensByProject(projectID)
	require.NoError(t, err)
	assert.Len(t, tokens, 1)

	// Update context (activate)
	ctx := NewContextManager(manager)
	err = ctx.SwitchTo(projectID)
	require.NoError(t, err)

	current, err := ctx.GetCurrent()
	require.NoError(t, err)
	assert.Equal(t, projectID, current.ID)
	assert.True(t, current.Active)

	// Delete
	err = registry.Remove(projectID)
	require.NoError(t, err)

	// Verify deleted
	_, err = registry.GetByID(projectID)
	assert.Error(t, err)

	// Note: Tokens may still exist after project deletion
	// This is by design - tokens are not cascade deleted
	// They just become orphaned and can be cleaned up separately
}

// TestMultipleConnectionsSequentially tests that multiple connections can work with the database
func TestMultipleConnectionsSequentially(t *testing.T) {
	_, cleanup := testutil.TempHomeDir(t)
	defer cleanup()

	// Create database with first connection
	manager1 := NewDatabaseManager()
	err := manager1.Initialize(GlobalMode)
	require.NoError(t, err)

	registry1 := NewProjectRegistry(manager1)
	project1 := &Project{Name: "Project 1", Path: "/path/1"}
	err = registry1.Register(project1)
	require.NoError(t, err)
	manager1.Close()

	// Open second connection and add another project
	manager2 := NewDatabaseManager()
	err = manager2.Discover()
	require.NoError(t, err)

	registry2 := NewProjectRegistry(manager2)
	project2 := &Project{Name: "Project 2", Path: "/path/2"}
	err = registry2.Register(project2)
	require.NoError(t, err)
	manager2.Close()

	// Open third connection and verify both projects exist
	manager3 := NewDatabaseManager()
	err = manager3.Discover()
	require.NoError(t, err)
	defer manager3.Close()

	registry3 := NewProjectRegistry(manager3)
	projects, err := registry3.List()
	require.NoError(t, err)
	assert.Len(t, projects, 2)

	// Verify both projects are accessible
	names := make(map[string]bool)
	for _, p := range projects {
		names[p.Name] = true
	}
	assert.True(t, names["Project 1"])
	assert.True(t, names["Project 2"])
}
