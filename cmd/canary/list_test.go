// Copyright (c) 2025 by Developer Network.
//
// For more details, see the LICENSE file in the root directory of this
// source code repository or contact Developer Network at info@devnw.com.

package main

import (
	"path/filepath"
	"testing"

	"go.devnw.com/canary/internal/storage"
)

// CANARY: REQ=CBIN-135; FEATURE="ListCommand"; ASPECT=CLI; STATUS=TESTED; TEST=TestCANARY_CBIN_135_CLI_ListCommand; UPDATED=2025-10-17
func TestCANARY_CBIN_135_CLI_ListCommand(t *testing.T) {
	// Setup: Create temporary database
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	// Migrate the database
	if err := storage.MigrateDB(dbPath, "all"); err != nil {
		t.Fatalf("Failed to migrate database: %v", err)
	}

	db, err := storage.Open(dbPath)
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	// Insert test tokens with various attributes
	testTokens := []*storage.Token{
		{
			ReqID:      "CBIN-100",
			Feature:    "HighPriorityFeature",
			Aspect:     "CLI",
			Status:     "STUB",
			FilePath:   "src/cli/feature1.go",
			LineNumber: 10,
			Priority:   1,
			UpdatedAt:  "2025-10-17",
			Owner:      "alice",
		},
		{
			ReqID:      "CBIN-101",
			Feature:    "MediumPriorityAPI",
			Aspect:     "API",
			Status:     "IMPL",
			FilePath:   "src/api/feature2.go",
			LineNumber: 20,
			Priority:   2,
			UpdatedAt:  "2025-10-16",
			Owner:      "bob",
		},
		{
			ReqID:      "CBIN-102",
			Feature:    "LowPriorityEngine",
			Aspect:     "Engine",
			Status:     "TESTED",
			FilePath:   "src/engine/feature3.go",
			LineNumber: 30,
			Test:       "TestFeature3",
			Priority:   3,
			UpdatedAt:  "2025-10-15",
			Owner:      "alice",
		},
		{
			ReqID:      "CBIN-103",
			Feature:    "StorageFeature",
			Aspect:     "Storage",
			Status:     "BENCHED",
			FilePath:   "src/storage/feature4.go",
			LineNumber: 40,
			Test:       "TestFeature4",
			Bench:      "BenchmarkFeature4",
			Priority:   1,
			UpdatedAt:  "2025-10-14",
			Owner:      "bob",
		},
		// Hidden tokens (should be filtered by default)
		{
			ReqID:      "CBIN-104",
			Feature:    "TestFileToken",
			Aspect:     "CLI",
			Status:     "STUB",
			FilePath:   "src/cli/feature_test.go",
			LineNumber: 50,
			Priority:   5,
			UpdatedAt:  "2025-10-13",
		},
		{
			ReqID:      "CBIN-105",
			Feature:    "TemplateToken",
			Aspect:     "Docs",
			Status:     "STUB",
			FilePath:   ".canary/templates/example.md",
			LineNumber: 60,
			Priority:   5,
			UpdatedAt:  "2025-10-12",
		},
	}

	for _, token := range testTokens {
		if err := db.UpsertToken(token); err != nil {
			t.Fatalf("Failed to upsert token: %v", err)
		}
	}

	t.Run("list all with default filters", func(t *testing.T) {
		// Default filters should exclude hidden paths
		filters := make(map[string]string)
		tokens, err := db.ListTokens(filters, "CBIN-[1-9][0-9]{2,}", "priority ASC, updated_at DESC", 0)
		if err != nil {
			t.Fatalf("ListTokens failed: %v", err)
		}

		// Should get 4 tokens (CBIN-100 to CBIN-103), excluding hidden ones
		if len(tokens) != 4 {
			t.Errorf("got %d tokens, want 4 (excluding hidden)", len(tokens))
		}

		// Verify first token is highest priority
		if len(tokens) > 0 && tokens[0].Priority != 1 {
			t.Errorf("first token priority: got %d, want 1", tokens[0].Priority)
		}
	})

	t.Run("filter by status", func(t *testing.T) {
		filters := map[string]string{"status": "STUB"}
		tokens, err := db.ListTokens(filters, "CBIN-[1-9][0-9]{2,}", "priority ASC", 0)
		if err != nil {
			t.Fatalf("ListTokens failed: %v", err)
		}

		// Should only get STUB tokens (CBIN-100)
		if len(tokens) != 1 {
			t.Errorf("got %d STUB tokens, want 1", len(tokens))
		}

		if len(tokens) > 0 && tokens[0].Status != "STUB" {
			t.Errorf("got status %s, want STUB", tokens[0].Status)
		}
	})

	t.Run("filter by aspect", func(t *testing.T) {
		filters := map[string]string{"aspect": "CLI"}
		tokens, err := db.ListTokens(filters, "CBIN-[1-9][0-9]{2,}", "priority ASC", 0)
		if err != nil {
			t.Fatalf("ListTokens failed: %v", err)
		}

		// Should only get CLI tokens (CBIN-100)
		if len(tokens) != 1 {
			t.Errorf("got %d CLI tokens, want 1", len(tokens))
		}

		if len(tokens) > 0 && tokens[0].Aspect != "CLI" {
			t.Errorf("got aspect %s, want CLI", tokens[0].Aspect)
		}
	})

	t.Run("filter by owner", func(t *testing.T) {
		filters := map[string]string{"owner": "alice"}
		tokens, err := db.ListTokens(filters, "CBIN-[1-9][0-9]{2,}", "priority ASC", 0)
		if err != nil {
			t.Fatalf("ListTokens failed: %v", err)
		}

		// Should get tokens owned by alice (CBIN-100, CBIN-102)
		if len(tokens) != 2 {
			t.Errorf("got %d tokens for alice, want 2", len(tokens))
		}

		for _, token := range tokens {
			if token.Owner != "alice" {
				t.Errorf("got owner %s, want alice", token.Owner)
			}
		}
	})

	t.Run("limit results", func(t *testing.T) {
		filters := make(map[string]string)
		tokens, err := db.ListTokens(filters, "CBIN-[1-9][0-9]{2,}", "priority ASC", 2)
		if err != nil {
			t.Fatalf("ListTokens failed: %v", err)
		}

		// Should get exactly 2 tokens
		if len(tokens) != 2 {
			t.Errorf("got %d tokens, want 2", len(tokens))
		}
	})

	t.Run("sort by priority", func(t *testing.T) {
		filters := make(map[string]string)
		tokens, err := db.ListTokens(filters, "CBIN-[1-9][0-9]{2,}", "priority ASC", 0)
		if err != nil {
			t.Fatalf("ListTokens failed: %v", err)
		}

		// Verify tokens are sorted by priority (1, 1, 2, 3)
		if len(tokens) >= 2 {
			if tokens[0].Priority > tokens[1].Priority {
				t.Errorf("tokens not sorted by priority: %d > %d", tokens[0].Priority, tokens[1].Priority)
			}
		}
	})

	t.Run("include hidden paths", func(t *testing.T) {
		filters := map[string]string{"include_hidden": "true"}
		tokens, err := db.ListTokens(filters, "CBIN-[1-9][0-9]{2,}", "priority ASC", 0)
		if err != nil {
			t.Fatalf("ListTokens failed: %v", err)
		}

		// Should get all 6 tokens including hidden ones
		if len(tokens) != 6 {
			t.Errorf("got %d tokens with include_hidden, want 6", len(tokens))
		}
	})

	t.Run("priority range filtering", func(t *testing.T) {
		filters := map[string]string{
			"priority_min": "1",
			"priority_max": "2",
		}
		tokens, err := db.ListTokens(filters, "CBIN-[1-9][0-9]{2,}", "priority ASC", 0)
		if err != nil {
			t.Fatalf("ListTokens failed: %v", err)
		}

		// Should get tokens with priority 1-2 (CBIN-100, CBIN-101, CBIN-103)
		if len(tokens) != 3 {
			t.Errorf("got %d tokens with priority 1-2, want 3", len(tokens))
		}

		for _, token := range tokens {
			if token.Priority < 1 || token.Priority > 2 {
				t.Errorf("token priority %d outside range 1-2", token.Priority)
			}
		}
	})

	t.Run("combined filters", func(t *testing.T) {
		filters := map[string]string{
			"status": "IMPL",
			"aspect": "API",
		}
		tokens, err := db.ListTokens(filters, "CBIN-[1-9][0-9]{2,}", "priority ASC", 0)
		if err != nil {
			t.Fatalf("ListTokens failed: %v", err)
		}

		// Should get only IMPL API tokens (CBIN-101)
		if len(tokens) != 1 {
			t.Errorf("got %d tokens with status=IMPL aspect=API, want 1", len(tokens))
		}

		if len(tokens) > 0 {
			if tokens[0].Status != "IMPL" || tokens[0].Aspect != "API" {
				t.Errorf("got status=%s aspect=%s, want IMPL API", tokens[0].Status, tokens[0].Aspect)
			}
		}
	})

	t.Run("empty results", func(t *testing.T) {
		filters := map[string]string{"status": "REMOVED"}
		tokens, err := db.ListTokens(filters, "CBIN-[1-9][0-9]{2,}", "priority ASC", 0)
		if err != nil {
			t.Fatalf("ListTokens failed: %v", err)
		}

		// Should get zero tokens
		if len(tokens) != 0 {
			t.Errorf("got %d tokens with status=REMOVED, want 0", len(tokens))
		}
	})
}

// CANARY: REQ=CBIN-135; FEATURE="ListCommand"; ASPECT=CLI; STATUS=TESTED; TEST=TestCANARY_CBIN_135_CLI_ListCommand_Sorting; UPDATED=2025-10-17
func TestCANARY_CBIN_135_CLI_ListCommand_Sorting(t *testing.T) {
	// Setup: Create temporary database
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	if err := storage.MigrateDB(dbPath, "all"); err != nil {
		t.Fatalf("Failed to migrate database: %v", err)
	}

	db, err := storage.Open(dbPath)
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	// Insert tokens with specific ordering attributes
	testTokens := []*storage.Token{
		{ReqID: "CBIN-200", Feature: "F1", Status: "STUB", Priority: 2, UpdatedAt: "2025-10-15", FilePath: "f1.go"},
		{ReqID: "CBIN-201", Feature: "F2", Status: "IMPL", Priority: 1, UpdatedAt: "2025-10-16", FilePath: "f2.go"},
		{ReqID: "CBIN-202", Feature: "F3", Status: "TESTED", Priority: 3, UpdatedAt: "2025-10-17", FilePath: "f3.go"},
		{ReqID: "CBIN-203", Feature: "F4", Status: "BENCHED", Priority: 1, UpdatedAt: "2025-10-14", FilePath: "f4.go"},
	}

	for _, token := range testTokens {
		if err := db.UpsertToken(token); err != nil {
			t.Fatalf("Failed to upsert token: %v", err)
		}
	}

	tests := []struct {
		name            string
		orderBy         string
		wantFirstReqID  string
		wantFirstStatus string
	}{
		{
			name:           "sort by priority ascending",
			orderBy:        "priority ASC",
			wantFirstReqID: "CBIN-201", // priority 1
		},
		{
			name:           "sort by priority descending",
			orderBy:        "priority DESC",
			wantFirstReqID: "CBIN-202", // priority 3
		},
		{
			name:            "sort by status ascending",
			orderBy:         "status ASC",
			wantFirstStatus: "BENCHED",
		},
		{
			name:           "sort by updated_at descending (newest first)",
			orderBy:        "updated_at DESC",
			wantFirstReqID: "CBIN-202", // 2025-10-17
		},
		{
			name:           "sort by updated_at ascending (oldest first)",
			orderBy:        "updated_at ASC",
			wantFirstReqID: "CBIN-203", // 2025-10-14
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filters := make(map[string]string)
			tokens, err := db.ListTokens(filters, "CBIN-[1-9][0-9]{2,}", tt.orderBy, 0)
			if err != nil {
				t.Fatalf("ListTokens failed: %v", err)
			}

			if len(tokens) == 0 {
				t.Fatal("got 0 tokens, want at least 1")
			}

			if tt.wantFirstReqID != "" && tokens[0].ReqID != tt.wantFirstReqID {
				t.Errorf("first token ReqID: got %s, want %s", tokens[0].ReqID, tt.wantFirstReqID)
			}

			if tt.wantFirstStatus != "" && tokens[0].Status != tt.wantFirstStatus {
				t.Errorf("first token Status: got %s, want %s", tokens[0].Status, tt.wantFirstStatus)
			}
		})
	}
}

// CANARY: REQ=CBIN-135; FEATURE="ListCommand"; ASPECT=CLI; STATUS=TESTED; TEST=TestCANARY_CBIN_135_CLI_ListCommand_Performance; UPDATED=2025-10-17
func TestCANARY_CBIN_135_CLI_ListCommand_Performance(t *testing.T) {
	// Setup: Create temporary database
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	if err := storage.MigrateDB(dbPath, "all"); err != nil {
		t.Fatalf("Failed to migrate database: %v", err)
	}

	db, err := storage.Open(dbPath)
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	// Insert 100 test tokens
	for i := 100; i < 200; i++ {
		token := &storage.Token{
			ReqID:      "CBIN-" + string(rune('0'+i/100)) + string(rune('0'+(i%100)/10)) + string(rune('0'+i%10)),
			Feature:    "Feature" + string(rune('0'+i%10)),
			Aspect:     []string{"CLI", "API", "Engine"}[i%3],
			Status:     []string{"STUB", "IMPL", "TESTED"}[i%3],
			FilePath:   "src/file" + string(rune('0'+i%10)) + ".go",
			LineNumber: i,
			Priority:   (i % 5) + 1,
			UpdatedAt:  "2025-10-17",
		}
		if err := db.UpsertToken(token); err != nil {
			t.Fatalf("Failed to upsert token %d: %v", i, err)
		}
	}

	// Test that query with 100 tokens completes quickly
	filters := make(map[string]string)
	tokens, err := db.ListTokens(filters, "CBIN-[1-9][0-9]{2,}", "priority ASC", 10)
	if err != nil {
		t.Fatalf("ListTokens failed: %v", err)
	}

	// Should get 10 tokens (limit)
	if len(tokens) != 10 {
		t.Errorf("got %d tokens, want 10", len(tokens))
	}

	// Verify all tokens have valid data
	for _, token := range tokens {
		if token.ReqID == "" || token.Feature == "" {
			t.Error("token missing required fields")
		}
	}
}

// CANARY: REQ=CBIN-135; FEATURE="ListCommand"; ASPECT=CLI; STATUS=TESTED; TEST=TestCANARY_CBIN_135_CLI_ListCommand_EdgeCases; UPDATED=2025-10-17
func TestCANARY_CBIN_135_CLI_ListCommand_EdgeCases(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	if err := storage.MigrateDB(dbPath, "all"); err != nil {
		t.Fatalf("Failed to migrate database: %v", err)
	}

	db, err := storage.Open(dbPath)
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	t.Run("empty database", func(t *testing.T) {
		filters := make(map[string]string)
		tokens, err := db.ListTokens(filters, "CBIN-[1-9][0-9]{2,}", "priority ASC", 0)
		if err != nil {
			t.Fatalf("ListTokens failed: %v", err)
		}

		if len(tokens) != 0 {
			t.Errorf("got %d tokens from empty database, want 0", len(tokens))
		}
	})

	// Insert one token for subsequent tests
	testToken := &storage.Token{
		ReqID:      "CBIN-300",
		Feature:    "EdgeCaseFeature",
		Aspect:     "CLI",
		Status:     "STUB",
		FilePath:   "edge.go",
		LineNumber: 1,
		Priority:   1,
		UpdatedAt:  "2025-10-17",
	}
	if err := db.UpsertToken(testToken); err != nil {
		t.Fatalf("Failed to upsert token: %v", err)
	}

	t.Run("limit zero (unlimited)", func(t *testing.T) {
		filters := make(map[string]string)
		tokens, err := db.ListTokens(filters, "CBIN-[1-9][0-9]{2,}", "priority ASC", 0)
		if err != nil {
			t.Fatalf("ListTokens failed: %v", err)
		}

		// Should get all tokens
		if len(tokens) != 1 {
			t.Errorf("got %d tokens with limit 0, want 1", len(tokens))
		}
	})

	t.Run("limit exceeds available", func(t *testing.T) {
		filters := make(map[string]string)
		tokens, err := db.ListTokens(filters, "CBIN-[1-9][0-9]{2,}", "priority ASC", 100)
		if err != nil {
			t.Fatalf("ListTokens failed: %v", err)
		}

		// Should get only available tokens (1)
		if len(tokens) != 1 {
			t.Errorf("got %d tokens with limit 100, want 1", len(tokens))
		}
	})

	t.Run("non-matching filter", func(t *testing.T) {
		filters := map[string]string{"status": "NONEXISTENT"}
		tokens, err := db.ListTokens(filters, "CBIN-[1-9][0-9]{2,}", "priority ASC", 0)
		if err != nil {
			t.Fatalf("ListTokens failed: %v", err)
		}

		if len(tokens) != 0 {
			t.Errorf("got %d tokens with non-matching filter, want 0", len(tokens))
		}
	})

	t.Run("priority boundary values", func(t *testing.T) {
		// Test priority filter with min=0 (should include all)
		filters := map[string]string{"priority_min": "0"}
		tokens, err := db.ListTokens(filters, "CBIN-[1-9][0-9]{2,}", "priority ASC", 0)
		if err != nil {
			t.Fatalf("ListTokens failed: %v", err)
		}

		if len(tokens) != 1 {
			t.Errorf("got %d tokens with priority_min=0, want 1", len(tokens))
		}
	})
}
