// Copyright (c) 2025 by Developer Network.
//
// For more details, see the LICENSE file in the root directory of this
// source code repository or contact Developer Network at info@devnw.com.

// CANARY: REQ=CBIN-146; FEATURE="PerformanceBenchmarks"; ASPECT=Storage; STATUS=IMPL; BENCH=BenchmarkMultiProject; UPDATED=2025-10-18
package storage

import (
	"fmt"
	"testing"

	"go.devnw.com/canary/internal/storage/testutil"
)

// BenchmarkProjectRegistration measures project registration performance
func BenchmarkProjectRegistration(b *testing.B) {
	_, cleanup := testutil.TempHomeDirB(b)
	defer cleanup()

	manager := NewDatabaseManager()
	err := manager.Initialize(GlobalMode)
	if err != nil {
		b.Fatal(err)
	}
	defer manager.Close()

	registry := NewProjectRegistry(manager)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		project := &Project{
			Name: fmt.Sprintf("Benchmark Project %d", i),
			Path: fmt.Sprintf("/bench/path/%d", i),
		}
		if err := registry.Register(project); err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkProjectListing measures performance of listing projects
func BenchmarkProjectListing(b *testing.B) {
	_, cleanup := testutil.TempHomeDirB(b)
	defer cleanup()

	manager := NewDatabaseManager()
	err := manager.Initialize(GlobalMode)
	if err != nil {
		b.Fatal(err)
	}
	defer manager.Close()

	registry := NewProjectRegistry(manager)

	// Pre-populate with projects
	for i := 0; i < 100; i++ {
		project := &Project{
			Name: fmt.Sprintf("Project %d", i),
			Path: fmt.Sprintf("/path/%d", i),
		}
		if err := registry.Register(project); err != nil {
			b.Fatal(err)
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := registry.List()
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkContextSwitching measures performance of switching between projects
func BenchmarkContextSwitching(b *testing.B) {
	_, cleanup := testutil.TempHomeDirB(b)
	defer cleanup()

	manager := NewDatabaseManager()
	err := manager.Initialize(GlobalMode)
	if err != nil {
		b.Fatal(err)
	}
	defer manager.Close()

	registry := NewProjectRegistry(manager)
	ctx := NewContextManager(manager)

	// Create two projects
	project1 := &Project{Name: "Project 1", Path: "/path/1"}
	if err := registry.Register(project1); err != nil {
		b.Fatal(err)
	}

	project2 := &Project{Name: "Project 2", Path: "/path/2"}
	if err := registry.Register(project2); err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if i%2 == 0 {
			if err := ctx.SwitchTo(project1.ID); err != nil {
				b.Fatal(err)
			}
		} else {
			if err := ctx.SwitchTo(project2.ID); err != nil {
				b.Fatal(err)
			}
		}
	}
}

// BenchmarkTokenInsertionSingleProject measures token insertion for a single project
func BenchmarkTokenInsertionSingleProject(b *testing.B) {
	_, cleanup := testutil.TempHomeDirB(b)
	defer cleanup()

	manager := NewDatabaseManager()
	err := manager.Initialize(GlobalMode)
	if err != nil {
		b.Fatal(err)
	}
	defer manager.Close()

	registry := NewProjectRegistry(manager)
	project := &Project{Name: "Benchmark Project", Path: "/bench"}
	if err := registry.Register(project); err != nil {
		b.Fatal(err)
	}

	db := &DB{conn: manager.conn, path: manager.path}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		token := &Token{
			ReqID:      fmt.Sprintf("CBIN-%d", i),
			Feature:    "BenchFeature",
			Aspect:     "API",
			Status:     "IMPL",
			FilePath:   fmt.Sprintf("/file%d.go", i),
			LineNumber: i,
			UpdatedAt:  "2025-10-18",
			RawToken:   "// CANARY: bench",
			IndexedAt:  "2025-10-18",
			ProjectID:  project.ID,
		}
		if err := db.UpsertToken(token); err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkTokenInsertionMultiProject measures token insertion across multiple projects
func BenchmarkTokenInsertionMultiProject(b *testing.B) {
	_, cleanup := testutil.TempHomeDirB(b)
	defer cleanup()

	manager := NewDatabaseManager()
	err := manager.Initialize(GlobalMode)
	if err != nil {
		b.Fatal(err)
	}
	defer manager.Close()

	registry := NewProjectRegistry(manager)

	// Create 10 projects
	projects := make([]*Project, 10)
	for i := 0; i < 10; i++ {
		projects[i] = &Project{
			Name: fmt.Sprintf("Project %d", i),
			Path: fmt.Sprintf("/path/%d", i),
		}
		if err := registry.Register(projects[i]); err != nil {
			b.Fatal(err)
		}
	}

	db := &DB{conn: manager.conn, path: manager.path}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		projectIndex := i % 10
		token := &Token{
			ReqID:      "CBIN-1000", // Same req_id across projects
			Feature:    "SharedFeature",
			Aspect:     "API",
			Status:     "IMPL",
			FilePath:   fmt.Sprintf("/file%d.go", i),
			LineNumber: i,
			UpdatedAt:  "2025-10-18",
			RawToken:   "// CANARY: bench",
			IndexedAt:  "2025-10-18",
			ProjectID:  projects[projectIndex].ID,
		}
		if err := db.UpsertToken(token); err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkGetTokensByProject measures retrieval of tokens for a specific project
func BenchmarkGetTokensByProject(b *testing.B) {
	_, cleanup := testutil.TempHomeDirB(b)
	defer cleanup()

	manager := NewDatabaseManager()
	err := manager.Initialize(GlobalMode)
	if err != nil {
		b.Fatal(err)
	}
	defer manager.Close()

	registry := NewProjectRegistry(manager)
	project := &Project{Name: "Benchmark Project", Path: "/bench"}
	if err := registry.Register(project); err != nil {
		b.Fatal(err)
	}

	db := &DB{conn: manager.conn, path: manager.path}

	// Pre-populate with 1000 tokens
	for i := 0; i < 1000; i++ {
		token := &Token{
			ReqID:      fmt.Sprintf("CBIN-%d", i),
			Feature:    "Feature",
			Aspect:     "API",
			Status:     "IMPL",
			FilePath:   fmt.Sprintf("/file%d.go", i),
			LineNumber: i,
			UpdatedAt:  "2025-10-18",
			RawToken:   "// CANARY",
			IndexedAt:  "2025-10-18",
			ProjectID:  project.ID,
		}
		if err := db.UpsertToken(token); err != nil {
			b.Fatal(err)
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := db.GetTokensByProject(project.ID)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkGetTokensByReqIDAndProject measures project-scoped requirement queries
func BenchmarkGetTokensByReqIDAndProject(b *testing.B) {
	_, cleanup := testutil.TempHomeDirB(b)
	defer cleanup()

	manager := NewDatabaseManager()
	err := manager.Initialize(GlobalMode)
	if err != nil {
		b.Fatal(err)
	}
	defer manager.Close()

	registry := NewProjectRegistry(manager)

	// Create 5 projects
	projects := make([]*Project, 5)
	for i := 0; i < 5; i++ {
		projects[i] = &Project{
			Name: fmt.Sprintf("Project %d", i),
			Path: fmt.Sprintf("/path/%d", i),
		}
		if err := registry.Register(projects[i]); err != nil {
			b.Fatal(err)
		}
	}

	db := &DB{conn: manager.conn, path: manager.path}

	// Add same req_id across all projects
	for _, project := range projects {
		for i := 0; i < 100; i++ {
			token := &Token{
				ReqID:      "CBIN-5000",
				Feature:    "SharedFeature",
				Aspect:     "API",
				Status:     "IMPL",
				FilePath:   fmt.Sprintf("/file%d.go", i),
				LineNumber: i,
				UpdatedAt:  "2025-10-18",
				RawToken:   "// CANARY",
				IndexedAt:  "2025-10-18",
				ProjectID:  project.ID,
			}
			if err := db.UpsertToken(token); err != nil {
				b.Fatal(err)
			}
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		projectIndex := i % 5
		_, err := db.GetTokensByReqIDAndProject("CBIN-5000", projects[projectIndex].ID)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkGetAllTokens measures cross-project token retrieval
func BenchmarkGetAllTokens(b *testing.B) {
	_, cleanup := testutil.TempHomeDirB(b)
	defer cleanup()

	manager := NewDatabaseManager()
	err := manager.Initialize(GlobalMode)
	if err != nil {
		b.Fatal(err)
	}
	defer manager.Close()

	registry := NewProjectRegistry(manager)

	// Create 3 projects
	for i := 0; i < 3; i++ {
		project := &Project{
			Name: fmt.Sprintf("Project %d", i),
			Path: fmt.Sprintf("/path/%d", i),
		}
		if err := registry.Register(project); err != nil {
			b.Fatal(err)
		}

		db := &DB{conn: manager.conn, path: manager.path}

		// Add 100 tokens per project
		for j := 0; j < 100; j++ {
			token := &Token{
				ReqID:      fmt.Sprintf("CBIN-%d", j),
				Feature:    "Feature",
				Aspect:     "API",
				Status:     "IMPL",
				FilePath:   fmt.Sprintf("/file%d.go", j),
				LineNumber: j,
				UpdatedAt:  "2025-10-18",
				RawToken:   "// CANARY",
				IndexedAt:  "2025-10-18",
				ProjectID:  project.ID,
			}
			if err := db.UpsertToken(token); err != nil {
				b.Fatal(err)
			}
		}
	}

	db := &DB{conn: manager.conn, path: manager.path}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := db.GetAllTokens()
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkDatabaseInitialization measures database initialization time
func BenchmarkDatabaseInitialization(b *testing.B) {
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		_, cleanup := testutil.TempHomeDirB(b)
		b.StartTimer()

		manager := NewDatabaseManager()
		err := manager.Initialize(GlobalMode)
		if err != nil {
			b.Fatal(err)
		}
		manager.Close()

		b.StopTimer()
		cleanup()
		b.StartTimer()
	}
}

// BenchmarkDatabaseDiscovery measures database discovery performance
func BenchmarkDatabaseDiscovery(b *testing.B) {
	_, cleanup := testutil.TempHomeDirB(b)
	defer cleanup()

	// Create a database first
	manager := NewDatabaseManager()
	err := manager.Initialize(GlobalMode)
	if err != nil {
		b.Fatal(err)
	}
	manager.Close()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		manager := NewDatabaseManager()
		err := manager.Discover()
		if err != nil {
			b.Fatal(err)
		}
		manager.Close()
	}
}

// BenchmarkProjectScaling measures performance with increasing number of projects
func BenchmarkProjectScaling(b *testing.B) {
	projectCounts := []int{10, 50, 100, 500}

	for _, count := range projectCounts {
		b.Run(fmt.Sprintf("Projects_%d", count), func(b *testing.B) {
			_, cleanup := testutil.TempHomeDirB(b)
			defer cleanup()

			manager := NewDatabaseManager()
			err := manager.Initialize(GlobalMode)
			if err != nil {
				b.Fatal(err)
			}
			defer manager.Close()

			registry := NewProjectRegistry(manager)

			// Create projects
			for i := 0; i < count; i++ {
				project := &Project{
					Name: fmt.Sprintf("Project %d", i),
					Path: fmt.Sprintf("/path/%d", i),
				}
				if err := registry.Register(project); err != nil {
					b.Fatal(err)
				}
			}

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_, err := registry.List()
				if err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

// BenchmarkTokenScaling measures performance with increasing token counts per project
func BenchmarkTokenScaling(b *testing.B) {
	tokenCounts := []int{100, 500, 1000, 5000}

	for _, count := range tokenCounts {
		b.Run(fmt.Sprintf("Tokens_%d", count), func(b *testing.B) {
			_, cleanup := testutil.TempHomeDirB(b)
			defer cleanup()

			manager := NewDatabaseManager()
			err := manager.Initialize(GlobalMode)
			if err != nil {
				b.Fatal(err)
			}
			defer manager.Close()

			registry := NewProjectRegistry(manager)
			project := &Project{Name: "Test Project", Path: "/test"}
			if err := registry.Register(project); err != nil {
				b.Fatal(err)
			}

			db := &DB{conn: manager.conn, path: manager.path}

			// Pre-populate tokens
			for i := 0; i < count; i++ {
				token := &Token{
					ReqID:      fmt.Sprintf("CBIN-%d", i),
					Feature:    "Feature",
					Aspect:     "API",
					Status:     "IMPL",
					FilePath:   fmt.Sprintf("/file%d.go", i),
					LineNumber: i,
					UpdatedAt:  "2025-10-18",
					RawToken:   "// CANARY",
					IndexedAt:  "2025-10-18",
					ProjectID:  project.ID,
				}
				if err := db.UpsertToken(token); err != nil {
					b.Fatal(err)
				}
			}

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_, err := db.GetTokensByProject(project.ID)
				if err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}
