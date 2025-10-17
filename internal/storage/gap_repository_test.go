// Copyright (c) 2024 by CodePros.
//
// This software is proprietary information of CodePros.
// Unauthorized use, copying, modification, distribution, and/or
// disclosure is strictly prohibited, except as provided under the terms
// of the commercial license agreement you have entered into with
// CodePros.
//
// For more details, see the LICENSE file in the root directory of this
// source code repository or contact CodePros at info@codepros.org.

package storage

import (
	"path/filepath"
	"testing"
	"time"
)

// CANARY: REQ=CBIN-140; FEATURE="GapRepositoryTests"; ASPECT=Storage; STATUS=IMPL; TEST=TestGapRepository_CreateEntry; UPDATED=2025-10-17
func TestGapRepository_CreateEntry(t *testing.T) {
	// Setup: Create temporary database
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	// Migrate the database
	if err := MigrateDB(dbPath, "all"); err != nil {
		t.Fatalf("Failed to migrate database: %v", err)
	}

	db, err := Open(dbPath)
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	repo := NewGapRepository(db)

	// Test creating a gap entry
	entry := &GapEntry{
		GapID:            "GAP-CBIN-140-001",
		ReqID:            "CBIN-140",
		Feature:          "GapTracking",
		Aspect:           "CLI",
		Category:         "logic_error",
		Description:      "Incorrect query ordering in GetFilesByReqID",
		CorrectiveAction: "Added ORDER BY clause to sort by file path",
		CreatedBy:        "test-agent",
	}

	err = repo.CreateEntry(entry)
	if err != nil {
		t.Fatalf("Failed to create gap entry: %v", err)
	}

	// Verify the entry was created
	retrieved, err := repo.GetEntryByGapID("GAP-CBIN-140-001")
	if err != nil {
		t.Fatalf("Failed to retrieve gap entry: %v", err)
	}

	if retrieved.ReqID != entry.ReqID {
		t.Errorf("ReqID = %s, want %s", retrieved.ReqID, entry.ReqID)
	}
	if retrieved.Feature != entry.Feature {
		t.Errorf("Feature = %s, want %s", retrieved.Feature, entry.Feature)
	}
	if retrieved.Description != entry.Description {
		t.Errorf("Description = %s, want %s", retrieved.Description, entry.Description)
	}
}

// CANARY: REQ=CBIN-140; FEATURE="GapRepositoryTests"; ASPECT=Storage; STATUS=IMPL; TEST=TestGapRepository_GetEntriesByReqID; UPDATED=2025-10-17
func TestGapRepository_GetEntriesByReqID(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	if err := MigrateDB(dbPath, "all"); err != nil {
		t.Fatalf("Failed to migrate database: %v", err)
	}

	db, err := Open(dbPath)
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	repo := NewGapRepository(db)

	// Create multiple entries for same requirement
	entries := []*GapEntry{
		{
			GapID:       "GAP-CBIN-140-001",
			ReqID:       "CBIN-140",
			Feature:     "GapTracking",
			Category:    "logic_error",
			Description: "First gap",
		},
		{
			GapID:       "GAP-CBIN-140-002",
			ReqID:       "CBIN-140",
			Feature:     "GapQuery",
			Category:    "test_failure",
			Description: "Second gap",
		},
		{
			GapID:       "GAP-CBIN-141-001",
			ReqID:       "CBIN-141",
			Feature:     "PromptFlag",
			Category:    "edge_case",
			Description: "Different requirement",
		},
	}

	for _, entry := range entries {
		if err := repo.CreateEntry(entry); err != nil {
			t.Fatalf("Failed to create entry: %v", err)
		}
	}

	// Query entries for CBIN-140
	results, err := repo.GetEntriesByReqID("CBIN-140")
	if err != nil {
		t.Fatalf("Failed to get entries by ReqID: %v", err)
	}

	if len(results) != 2 {
		t.Fatalf("Got %d entries, want 2", len(results))
	}

	// Verify all entries have correct ReqID
	for _, entry := range results {
		if entry.ReqID != "CBIN-140" {
			t.Errorf("Entry has ReqID %s, want CBIN-140", entry.ReqID)
		}
	}
}

// CANARY: REQ=CBIN-140; FEATURE="GapRepositoryTests"; ASPECT=Storage; STATUS=IMPL; TEST=TestGapRepository_UpdateHelpfulness; UPDATED=2025-10-17
func TestGapRepository_UpdateHelpfulness(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	if err := MigrateDB(dbPath, "all"); err != nil {
		t.Fatalf("Failed to migrate database: %v", err)
	}

	db, err := Open(dbPath)
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	repo := NewGapRepository(db)

	entry := &GapEntry{
		GapID:       "GAP-CBIN-140-001",
		ReqID:       "CBIN-140",
		Feature:     "GapTracking",
		Category:    "logic_error",
		Description: "Test gap",
	}

	if err := repo.CreateEntry(entry); err != nil {
		t.Fatalf("Failed to create entry: %v", err)
	}

	// Mark as helpful
	if err := repo.MarkHelpful("GAP-CBIN-140-001"); err != nil {
		t.Fatalf("Failed to mark helpful: %v", err)
	}

	// Verify helpful count increased
	retrieved, err := repo.GetEntryByGapID("GAP-CBIN-140-001")
	if err != nil {
		t.Fatalf("Failed to retrieve entry: %v", err)
	}

	if retrieved.HelpfulCount != 1 {
		t.Errorf("HelpfulCount = %d, want 1", retrieved.HelpfulCount)
	}

	// Mark as unhelpful
	if err := repo.MarkUnhelpful("GAP-CBIN-140-001"); err != nil {
		t.Fatalf("Failed to mark unhelpful: %v", err)
	}

	retrieved, err = repo.GetEntryByGapID("GAP-CBIN-140-001")
	if err != nil {
		t.Fatalf("Failed to retrieve entry: %v", err)
	}

	if retrieved.UnhelpfulCount != 1 {
		t.Errorf("UnhelpfulCount = %d, want 1", retrieved.UnhelpfulCount)
	}
}

// CANARY: REQ=CBIN-140; FEATURE="GapRepositoryTests"; ASPECT=Storage; STATUS=IMPL; TEST=TestGapRepository_QueryWithFilters; UPDATED=2025-10-17
func TestGapRepository_QueryWithFilters(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	if err := MigrateDB(dbPath, "all"); err != nil {
		t.Fatalf("Failed to migrate database: %v", err)
	}

	db, err := Open(dbPath)
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	repo := NewGapRepository(db)

	// Create entries with different categories and features
	entries := []*GapEntry{
		{
			GapID:       "GAP-001",
			ReqID:       "CBIN-140",
			Feature:     "FeatureA",
			Category:    "logic_error",
			Description: "Logic error gap",
		},
		{
			GapID:       "GAP-002",
			ReqID:       "CBIN-140",
			Feature:     "FeatureB",
			Category:    "test_failure",
			Description: "Test failure gap",
		},
		{
			GapID:       "GAP-003",
			ReqID:       "CBIN-141",
			Feature:     "FeatureA",
			Category:    "logic_error",
			Description: "Another logic error",
		},
	}

	for _, entry := range entries {
		if err := repo.CreateEntry(entry); err != nil {
			t.Fatalf("Failed to create entry: %v", err)
		}
	}

	tests := []struct {
		name       string
		filter     GapQueryFilter
		wantCount  int
		wantGapIDs []string
	}{
		{
			name:       "filter by category",
			filter:     GapQueryFilter{Category: "logic_error"},
			wantCount:  2,
			wantGapIDs: []string{"GAP-001", "GAP-003"},
		},
		{
			name:       "filter by feature",
			filter:     GapQueryFilter{Feature: "FeatureA"},
			wantCount:  2,
			wantGapIDs: []string{"GAP-001", "GAP-003"},
		},
		{
			name:       "filter by reqID",
			filter:     GapQueryFilter{ReqID: "CBIN-140"},
			wantCount:  2,
			wantGapIDs: []string{"GAP-001", "GAP-002"},
		},
		{
			name:       "filter by reqID and category",
			filter:     GapQueryFilter{ReqID: "CBIN-140", Category: "logic_error"},
			wantCount:  1,
			wantGapIDs: []string{"GAP-001"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			results, err := repo.QueryEntries(tt.filter)
			if err != nil {
				t.Fatalf("QueryEntries failed: %v", err)
			}

			if len(results) != tt.wantCount {
				t.Errorf("Got %d results, want %d", len(results), tt.wantCount)
			}

			// Verify correct entries returned
			gotGapIDs := make(map[string]bool)
			for _, entry := range results {
				gotGapIDs[entry.GapID] = true
			}

			for _, wantID := range tt.wantGapIDs {
				if !gotGapIDs[wantID] {
					t.Errorf("Expected to find gap %s in results", wantID)
				}
			}
		})
	}
}

// CANARY: REQ=CBIN-140; FEATURE="GapRepositoryTests"; ASPECT=Storage; STATUS=IMPL; TEST=TestGapRepository_GetTopGaps; UPDATED=2025-10-17
func TestGapRepository_GetTopGaps(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	if err := MigrateDB(dbPath, "all"); err != nil {
		t.Fatalf("Failed to migrate database: %v", err)
	}

	db, err := Open(dbPath)
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	repo := NewGapRepository(db)

	// Create entries with different helpful counts
	entries := []*GapEntry{
		{
			GapID:       "GAP-001",
			ReqID:       "CBIN-140",
			Feature:     "Feature1",
			Category:    "logic_error",
			Description: "Gap 1",
		},
		{
			GapID:       "GAP-002",
			ReqID:       "CBIN-140",
			Feature:     "Feature2",
			Category:    "logic_error",
			Description: "Gap 2",
		},
		{
			GapID:       "GAP-003",
			ReqID:       "CBIN-140",
			Feature:     "Feature3",
			Category:    "logic_error",
			Description: "Gap 3",
		},
	}

	for _, entry := range entries {
		if err := repo.CreateEntry(entry); err != nil {
			t.Fatalf("Failed to create entry: %v", err)
		}
	}

	// Mark helpfulness for different entries
	// GAP-003: 3 helpful
	for i := 0; i < 3; i++ {
		repo.MarkHelpful("GAP-003")
	}
	// GAP-001: 2 helpful
	for i := 0; i < 2; i++ {
		repo.MarkHelpful("GAP-001")
	}
	// GAP-002: 0 helpful

	// Query top 2 gaps
	config := &GapConfig{
		MaxGapInjection:     2,
		MinHelpfulThreshold: 0,
		RankingStrategy:     "helpful_desc",
	}

	results, err := repo.GetTopGaps("CBIN-140", config)
	if err != nil {
		t.Fatalf("GetTopGaps failed: %v", err)
	}

	// Should return top 2 by helpful count
	if len(results) != 2 {
		t.Fatalf("Got %d results, want 2", len(results))
	}

	// Verify ordering (most helpful first)
	if results[0].GapID != "GAP-003" {
		t.Errorf("First result GapID = %s, want GAP-003", results[0].GapID)
	}
	if results[1].GapID != "GAP-001" {
		t.Errorf("Second result GapID = %s, want GAP-001", results[1].GapID)
	}
}

// CANARY: REQ=CBIN-140; FEATURE="GapRepositoryTests"; ASPECT=Storage; STATUS=IMPL; TEST=TestGapRepository_GetCategories; UPDATED=2025-10-17
func TestGapRepository_GetCategories(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	if err := MigrateDB(dbPath, "all"); err != nil {
		t.Fatalf("Failed to migrate database: %v", err)
	}

	db, err := Open(dbPath)
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	repo := NewGapRepository(db)

	categories, err := repo.GetCategories()
	if err != nil {
		t.Fatalf("GetCategories failed: %v", err)
	}

	// Should have default categories from migration
	expectedCategories := []string{
		"logic_error",
		"test_failure",
		"performance",
		"security",
		"edge_case",
		"integration",
		"documentation",
		"other",
	}

	if len(categories) != len(expectedCategories) {
		t.Errorf("Got %d categories, want %d", len(categories), len(expectedCategories))
	}

	// Verify all expected categories exist
	categoryMap := make(map[string]bool)
	for _, cat := range categories {
		categoryMap[cat.Name] = true
	}

	for _, expected := range expectedCategories {
		if !categoryMap[expected] {
			t.Errorf("Expected category %s not found", expected)
		}
	}
}

// CANARY: REQ=CBIN-140; FEATURE="GapRepositoryTests"; ASPECT=Storage; STATUS=IMPL; TEST=TestGapRepository_GetConfig; UPDATED=2025-10-17
func TestGapRepository_GetConfig(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	if err := MigrateDB(dbPath, "all"); err != nil {
		t.Fatalf("Failed to migrate database: %v", err)
	}

	db, err := Open(dbPath)
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	repo := NewGapRepository(db)

	config, err := repo.GetConfig()
	if err != nil {
		t.Fatalf("GetConfig failed: %v", err)
	}

	// Verify default configuration from migration
	if config.MaxGapInjection != 10 {
		t.Errorf("MaxGapInjection = %d, want 10", config.MaxGapInjection)
	}
	if config.MinHelpfulThreshold != 1 {
		t.Errorf("MinHelpfulThreshold = %d, want 1", config.MinHelpfulThreshold)
	}
	if config.RankingStrategy != "helpful_desc" {
		t.Errorf("RankingStrategy = %s, want helpful_desc", config.RankingStrategy)
	}
}

// CANARY: REQ=CBIN-140; FEATURE="GapRepositoryTests"; ASPECT=Storage; STATUS=IMPL; TEST=TestGapRepository_UpdateConfig; UPDATED=2025-10-17
func TestGapRepository_UpdateConfig(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	if err := MigrateDB(dbPath, "all"); err != nil {
		t.Fatalf("Failed to migrate database: %v", err)
	}

	db, err := Open(dbPath)
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	repo := NewGapRepository(db)

	// Update configuration
	newConfig := &GapConfig{
		MaxGapInjection:     20,
		MinHelpfulThreshold: 2,
		RankingStrategy:     "recency_desc",
		UpdatedAt:           time.Now(),
	}

	if err := repo.UpdateConfig(newConfig); err != nil {
		t.Fatalf("UpdateConfig failed: %v", err)
	}

	// Retrieve and verify
	retrieved, err := repo.GetConfig()
	if err != nil {
		t.Fatalf("GetConfig failed: %v", err)
	}

	if retrieved.MaxGapInjection != 20 {
		t.Errorf("MaxGapInjection = %d, want 20", retrieved.MaxGapInjection)
	}
	if retrieved.MinHelpfulThreshold != 2 {
		t.Errorf("MinHelpfulThreshold = %d, want 2", retrieved.MinHelpfulThreshold)
	}
	if retrieved.RankingStrategy != "recency_desc" {
		t.Errorf("RankingStrategy = %s, want recency_desc", retrieved.RankingStrategy)
	}
}

// CANARY: REQ=CBIN-140; FEATURE="GapRepositoryTests"; ASPECT=Storage; STATUS=IMPL; TEST=TestGapRepository_PerformanceWith1000Entries; UPDATED=2025-10-17
func TestGapRepository_PerformanceWith1000Entries(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping performance test in short mode")
	}

	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	if err := MigrateDB(dbPath, "all"); err != nil {
		t.Fatalf("Failed to migrate database: %v", err)
	}

	db, err := Open(dbPath)
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	repo := NewGapRepository(db)

	// Create 1000 entries
	for i := 0; i < 1000; i++ {
		entry := &GapEntry{
			GapID:       "GAP-" + string(rune('A'+i/26)) + string(rune('A'+i%26)),
			ReqID:       "CBIN-" + string(rune('1'+i/100)),
			Feature:     "Feature" + string(rune('A'+i%10)),
			Category:    "logic_error",
			Description: "Performance test entry",
		}
		if err := repo.CreateEntry(entry); err != nil {
			t.Fatalf("Failed to create entry %d: %v", i, err)
		}
	}

	// Measure query performance
	start := time.Now()
	filter := GapQueryFilter{Category: "logic_error"}
	results, err := repo.QueryEntries(filter)
	elapsed := time.Since(start)

	if err != nil {
		t.Fatalf("QueryEntries failed: %v", err)
	}

	if len(results) != 1000 {
		t.Errorf("Got %d results, want 1000", len(results))
	}

	// Verify query completes in under 2 seconds (requirement from plan)
	if elapsed > 2*time.Second {
		t.Errorf("Query took %v, want < 2s", elapsed)
	} else {
		t.Logf("Query performance: %v for 1000 entries", elapsed)
	}
}
