// Copyright (c) 2025 by Developer Network.
//
// For more details, see the LICENSE file in the root directory of this
// source code repository or contact Developer Network at info@devnw.com.

// CANARY: REQ=CBIN-146; FEATURE="TokenNamespacing"; ASPECT=Storage; STATUS=IMPL; TEST=TestTokenNamespacing; UPDATED=2025-10-18
package storage

import (
	"testing"

	"go.devnw.com/canary/internal/storage/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTokenIsolationBetweenProjects(t *testing.T) {
	_, cleanup := testutil.TempHomeDir(t)
	defer cleanup()

	// Initialize database
	manager := NewDatabaseManager()
	err := manager.Initialize(GlobalMode)
	require.NoError(t, err)
	defer manager.Close()

	registry := NewProjectRegistry(manager)
	db := &DB{conn: manager.conn, path: manager.path}

	// Register two projects
	project1 := &Project{Name: "Project 1", Path: "/path/1"}
	err = registry.Register(project1)
	require.NoError(t, err)

	project2 := &Project{Name: "Project 2", Path: "/path/2"}
	err = registry.Register(project2)
	require.NoError(t, err)

	// Create same token in both projects (same req_id)
	token1 := &Token{
		ReqID:     "CBIN-100",
		Feature:   "TestFeature",
		Aspect:    "API",
		Status:    "IMPL",
		FilePath:  "/file1.go",
		LineNumber: 10,
		UpdatedAt: "2025-10-18",
		RawToken:  "// CANARY: REQ=CBIN-100; FEATURE=\"TestFeature\"; ASPECT=API; STATUS=IMPL",
		IndexedAt: "2025-10-18",
		ProjectID: project1.ID, // New field
	}

	token2 := &Token{
		ReqID:     "CBIN-100", // Same ID
		Feature:   "TestFeature",
		Aspect:    "API",
		Status:    "IMPL",
		FilePath:  "/file2.go",
		LineNumber: 20,
		UpdatedAt: "2025-10-18",
		RawToken:  "// CANARY: REQ=CBIN-100; FEATURE=\"TestFeature\"; ASPECT=API; STATUS=IMPL",
		IndexedAt: "2025-10-18",
		ProjectID: project2.ID, // Different project
	}

	// Both should succeed (different projects)
	err = db.UpsertToken(token1)
	require.NoError(t, err)

	err = db.UpsertToken(token2)
	require.NoError(t, err)

	// Verify tokens are isolated by project
	tokens1, err := db.GetTokensByProject(project1.ID)
	require.NoError(t, err)
	assert.Len(t, tokens1, 1)
	assert.Equal(t, project1.ID, tokens1[0].ProjectID)

	tokens2, err := db.GetTokensByProject(project2.ID)
	require.NoError(t, err)
	assert.Len(t, tokens2, 1)
	assert.Equal(t, project2.ID, tokens2[0].ProjectID)
}

func TestCrossProjectQuery(t *testing.T) {
	_, cleanup := testutil.TempHomeDir(t)
	defer cleanup()

	manager := NewDatabaseManager()
	err := manager.Initialize(GlobalMode)
	require.NoError(t, err)
	defer manager.Close()

	registry := NewProjectRegistry(manager)
	db := &DB{conn: manager.conn, path: manager.path}

	// Register projects
	project1 := &Project{Name: "Project 1", Path: "/path/1"}
	err = registry.Register(project1)
	require.NoError(t, err)

	project2 := &Project{Name: "Project 2", Path: "/path/2"}
	err = registry.Register(project2)
	require.NoError(t, err)

	// Create tokens in different projects
	token1 := &Token{
		ReqID:     "CBIN-101",
		Feature:   "Feature1",
		Aspect:    "API",
		Status:    "IMPL",
		FilePath:  "/file1.go",
		LineNumber: 10,
		UpdatedAt: "2025-10-18",
		RawToken:  "test",
		IndexedAt: "2025-10-18",
		ProjectID: project1.ID,
	}

	token2 := &Token{
		ReqID:     "CBIN-102",
		Feature:   "Feature2",
		Aspect:    "Storage",
		Status:    "TESTED",
		FilePath:  "/file2.go",
		LineNumber: 20,
		UpdatedAt: "2025-10-18",
		RawToken:  "test",
		IndexedAt: "2025-10-18",
		ProjectID: project2.ID,
	}

	err = db.UpsertToken(token1)
	require.NoError(t, err)

	err = db.UpsertToken(token2)
	require.NoError(t, err)

	// Query across all projects
	allTokens, err := db.GetAllTokens()
	require.NoError(t, err)
	assert.Len(t, allTokens, 2)

	// Verify each token has correct project ID
	projectIDs := make(map[string]bool)
	for _, token := range allTokens {
		projectIDs[token.ProjectID] = true
	}
	assert.True(t, projectIDs[project1.ID])
	assert.True(t, projectIDs[project2.ID])
}

func TestGetTokensByReqIDAndProject(t *testing.T) {
	_, cleanup := testutil.TempHomeDir(t)
	defer cleanup()

	manager := NewDatabaseManager()
	err := manager.Initialize(GlobalMode)
	require.NoError(t, err)
	defer manager.Close()

	registry := NewProjectRegistry(manager)
	db := &DB{conn: manager.conn, path: manager.path}

	// Register projects
	project1 := &Project{Name: "Project 1", Path: "/path/1"}
	err = registry.Register(project1)
	require.NoError(t, err)

	project2 := &Project{Name: "Project 2", Path: "/path/2"}
	err = registry.Register(project2)
	require.NoError(t, err)

	// Create same req_id in both projects
	token1 := &Token{
		ReqID:     "CBIN-200",
		Feature:   "SharedFeature",
		Aspect:    "API",
		Status:    "IMPL",
		FilePath:  "/file1.go",
		LineNumber: 10,
		UpdatedAt: "2025-10-18",
		RawToken:  "test",
		IndexedAt: "2025-10-18",
		ProjectID: project1.ID,
	}

	token2 := &Token{
		ReqID:     "CBIN-200", // Same req_id
		Feature:   "SharedFeature",
		Aspect:    "API",
		Status:    "TESTED",
		FilePath:  "/file2.go",
		LineNumber: 20,
		UpdatedAt: "2025-10-18",
		RawToken:  "test",
		IndexedAt: "2025-10-18",
		ProjectID: project2.ID,
	}

	err = db.UpsertToken(token1)
	require.NoError(t, err)

	err = db.UpsertToken(token2)
	require.NoError(t, err)

	// Get tokens by req_id for specific project
	tokens1, err := db.GetTokensByReqIDAndProject("CBIN-200", project1.ID)
	require.NoError(t, err)
	assert.Len(t, tokens1, 1)
	assert.Equal(t, "IMPL", tokens1[0].Status)

	tokens2, err := db.GetTokensByReqIDAndProject("CBIN-200", project2.ID)
	require.NoError(t, err)
	assert.Len(t, tokens2, 1)
	assert.Equal(t, "TESTED", tokens2[0].Status)
}

func TestTokenUniqueConstraintWithProjects(t *testing.T) {
	_, cleanup := testutil.TempHomeDir(t)
	defer cleanup()

	manager := NewDatabaseManager()
	err := manager.Initialize(GlobalMode)
	require.NoError(t, err)
	defer manager.Close()

	registry := NewProjectRegistry(manager)
	db := &DB{conn: manager.conn, path: manager.path}

	project := &Project{Name: "Test Project", Path: "/path/test"}
	err = registry.Register(project)
	require.NoError(t, err)

	// Create a token
	token1 := &Token{
		ReqID:     "CBIN-300",
		Feature:   "UniqueTest",
		Aspect:    "API",
		Status:    "IMPL",
		FilePath:  "/file.go",
		LineNumber: 100,
		UpdatedAt: "2025-10-18",
		RawToken:  "test",
		IndexedAt: "2025-10-18",
		ProjectID: project.ID,
	}

	err = db.UpsertToken(token1)
	require.NoError(t, err)

	// Try to insert duplicate (same project, req_id, feature, file, line)
	token2 := &Token{
		ReqID:     "CBIN-300",
		Feature:   "UniqueTest",
		Aspect:    "API",
		Status:    "TESTED", // Different status
		FilePath:  "/file.go",
		LineNumber: 100,
		UpdatedAt: "2025-10-18",
		RawToken:  "test updated",
		IndexedAt: "2025-10-18",
		ProjectID: project.ID,
	}

	// Should update, not fail
	err = db.UpsertToken(token2)
	require.NoError(t, err)

	// Verify it was updated
	tokens, err := db.GetTokensByProject(project.ID)
	require.NoError(t, err)
	assert.Len(t, tokens, 1)
	assert.Equal(t, "TESTED", tokens[0].Status)
}

func TestDefaultProjectForBackwardCompatibility(t *testing.T) {
	_, cleanup := testutil.TempHomeDir(t)
	defer cleanup()

	manager := NewDatabaseManager()
	err := manager.Initialize(GlobalMode)
	require.NoError(t, err)
	defer manager.Close()

	db := &DB{conn: manager.conn, path: manager.path}

	// Create token without project_id (backward compatibility)
	token := &Token{
		ReqID:     "CBIN-400",
		Feature:   "BackwardCompat",
		Aspect:    "API",
		Status:    "IMPL",
		FilePath:  "/file.go",
		LineNumber: 50,
		UpdatedAt: "2025-10-18",
		RawToken:  "test",
		IndexedAt: "2025-10-18",
		ProjectID: "", // Empty project ID
	}

	// Should use default project or handle gracefully
	err = db.UpsertToken(token)
	require.NoError(t, err)

	// Verify token was stored
	allTokens, err := db.GetAllTokens()
	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(allTokens), 1)
}

func TestProjectScopedTokenOperations(t *testing.T) {
	_, cleanup := testutil.TempHomeDir(t)
	defer cleanup()

	manager := NewDatabaseManager()
	err := manager.Initialize(GlobalMode)
	require.NoError(t, err)
	defer manager.Close()

	registry := NewProjectRegistry(manager)
	db := &DB{conn: manager.conn, path: manager.path}

	// Create projects
	project1 := &Project{Name: "Project 1", Path: "/p1"}
	err = registry.Register(project1)
	require.NoError(t, err)

	project2 := &Project{Name: "Project 2", Path: "/p2"}
	err = registry.Register(project2)
	require.NoError(t, err)

	// Add multiple tokens to each project
	for i := 0; i < 3; i++ {
		token := &Token{
			ReqID:      "CBIN-" + string(rune(500+i)),
			Feature:    "Feature" + string(rune(i)),
			Aspect:     "API",
			Status:     "IMPL",
			FilePath:   "/file.go",
			LineNumber: i * 10,
			UpdatedAt:  "2025-10-18",
			RawToken:   "test",
			IndexedAt:  "2025-10-18",
			ProjectID:  project1.ID,
		}
		err = db.UpsertToken(token)
		require.NoError(t, err)
	}

	for i := 0; i < 2; i++ {
		token := &Token{
			ReqID:      "CBIN-" + string(rune(600+i)),
			Feature:    "Feature" + string(rune(i)),
			Aspect:     "Storage",
			Status:     "TESTED",
			FilePath:   "/storage.go",
			LineNumber: i * 10,
			UpdatedAt:  "2025-10-18",
			RawToken:   "test",
			IndexedAt:  "2025-10-18",
			ProjectID:  project2.ID,
		}
		err = db.UpsertToken(token)
		require.NoError(t, err)
	}

	// Verify project 1 has 3 tokens
	tokens1, err := db.GetTokensByProject(project1.ID)
	require.NoError(t, err)
	assert.Len(t, tokens1, 3)

	// Verify project 2 has 2 tokens
	tokens2, err := db.GetTokensByProject(project2.ID)
	require.NoError(t, err)
	assert.Len(t, tokens2, 2)

	// Verify all tokens have correct project IDs
	for _, token := range tokens1 {
		assert.Equal(t, project1.ID, token.ProjectID)
	}

	for _, token := range tokens2 {
		assert.Equal(t, project2.ID, token.ProjectID)
	}
}
