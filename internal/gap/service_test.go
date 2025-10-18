// Copyright (c) 2025 by Developer Network.
//
// For more details, see the LICENSE file in the root directory of this
// source code repository or contact Developer Network at info@devnw.com.

package gap

import (
	"path/filepath"
	"strings"
	"testing"

	"go.devnw.com/canary/internal/storage"
)

// CANARY: REQ=CBIN-140; FEATURE="GapServiceTests"; ASPECT=Engine; STATUS=IMPL; TEST=TestService_MarkGap; UPDATED=2025-10-17
func TestService_MarkGap(t *testing.T) {
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

	repo := storage.NewGapRepository(db)
	service := NewService(repo)

	// Test marking a gap
	gapID, err := service.MarkGap(
		"CBIN-140",
		"GapTracking",
		"CLI",
		"logic_error",
		"Incorrect query ordering",
		"Added ORDER BY clause",
		"test-agent",
	)

	if err != nil {
		t.Fatalf("MarkGap failed: %v", err)
	}

	if gapID == "" {
		t.Error("Expected non-empty gap ID")
	}

	// Verify gap was created
	entry, err := repo.GetEntryByGapID(gapID)
	if err != nil {
		t.Fatalf("Failed to retrieve gap: %v", err)
	}

	if entry.ReqID != "CBIN-140" {
		t.Errorf("ReqID = %s, want CBIN-140", entry.ReqID)
	}
}

// CANARY: REQ=CBIN-140; FEATURE="GapServiceTests"; ASPECT=Engine; STATUS=IMPL; TEST=TestService_MarkGap_Validation; UPDATED=2025-10-17
func TestService_MarkGap_Validation(t *testing.T) {
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

	repo := storage.NewGapRepository(db)
	service := NewService(repo)

	tests := []struct {
		name        string
		reqID       string
		feature     string
		category    string
		description string
		wantErr     bool
	}{
		{
			name:        "missing req_id",
			reqID:       "",
			feature:     "Feature",
			category:    "logic_error",
			description: "Description",
			wantErr:     true,
		},
		{
			name:        "missing feature",
			reqID:       "CBIN-140",
			feature:     "",
			category:    "logic_error",
			description: "Description",
			wantErr:     true,
		},
		{
			name:        "invalid category",
			reqID:       "CBIN-140",
			feature:     "Feature",
			category:    "invalid_category",
			description: "Description",
			wantErr:     true,
		},
		{
			name:        "missing description",
			reqID:       "CBIN-140",
			feature:     "Feature",
			category:    "logic_error",
			description: "",
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := service.MarkGap(tt.reqID, tt.feature, "", tt.category, tt.description, "", "test")
			if (err != nil) != tt.wantErr {
				t.Errorf("MarkGap() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// CANARY: REQ=CBIN-140; FEATURE="GapServiceTests"; ASPECT=Engine; STATUS=IMPL; TEST=TestService_QueryGaps; UPDATED=2025-10-17
func TestService_QueryGaps(t *testing.T) {
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

	repo := storage.NewGapRepository(db)
	service := NewService(repo)

	// Create test gaps
	service.MarkGap("CBIN-140", "Feature1", "CLI", "logic_error", "Description 1", "", "test")
	service.MarkGap("CBIN-140", "Feature2", "API", "test_failure", "Description 2", "", "test")
	service.MarkGap("CBIN-141", "Feature3", "CLI", "logic_error", "Description 3", "", "test")

	// Query by reqID
	gaps, err := service.QueryGaps("CBIN-140", "", "", "", 0)
	if err != nil {
		t.Fatalf("QueryGaps failed: %v", err)
	}

	if len(gaps) != 2 {
		t.Errorf("Got %d gaps, want 2", len(gaps))
	}

	// Query by category
	gaps, err = service.QueryGaps("", "", "", "logic_error", 0)
	if err != nil {
		t.Fatalf("QueryGaps failed: %v", err)
	}

	if len(gaps) != 2 {
		t.Errorf("Got %d gaps with logic_error category, want 2", len(gaps))
	}
}

// CANARY: REQ=CBIN-140; FEATURE="GapServiceTests"; ASPECT=Engine; STATUS=IMPL; TEST=TestService_GenerateReport; UPDATED=2025-10-17
func TestService_GenerateReport(t *testing.T) {
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

	repo := storage.NewGapRepository(db)
	service := NewService(repo)

	// Create test gaps
	service.MarkGap("CBIN-140", "Feature1", "CLI", "logic_error", "Description 1", "Action 1", "test")
	service.MarkGap("CBIN-140", "Feature2", "API", "test_failure", "Description 2", "Action 2", "test")

	// Generate report
	report, err := service.GenerateReport("CBIN-140")
	if err != nil {
		t.Fatalf("GenerateReport failed: %v", err)
	}

	if report == "" {
		t.Error("Expected non-empty report")
	}

	// Verify report contains key information
	if !strings.Contains(report, "Gap Analysis Report for CBIN-140") {
		t.Error("Report missing title")
	}
	if !strings.Contains(report, "Total Gaps: 2") {
		t.Error("Report missing gap count")
	}
}

// CANARY: REQ=CBIN-140; FEATURE="GapServiceTests"; ASPECT=Engine; STATUS=IMPL; TEST=TestService_MarkHelpful; UPDATED=2025-10-17
func TestService_MarkHelpful(t *testing.T) {
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

	repo := storage.NewGapRepository(db)
	service := NewService(repo)

	// Create a gap
	gapID, _ := service.MarkGap("CBIN-140", "Feature1", "CLI", "logic_error", "Description", "", "test")

	// Mark as helpful
	if err := service.MarkHelpful(gapID); err != nil {
		t.Fatalf("MarkHelpful failed: %v", err)
	}

	// Verify count increased
	entry, _ := repo.GetEntryByGapID(gapID)
	if entry.HelpfulCount != 1 {
		t.Errorf("HelpfulCount = %d, want 1", entry.HelpfulCount)
	}

	// Test invalid gap ID
	if err := service.MarkHelpful("INVALID"); err == nil {
		t.Error("Expected error for invalid gap ID")
	}
}

// CANARY: REQ=CBIN-140; FEATURE="GapServiceTests"; ASPECT=Engine; STATUS=IMPL; TEST=TestService_FormatGapsForInjection; UPDATED=2025-10-17
func TestService_FormatGapsForInjection(t *testing.T) {
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

	repo := storage.NewGapRepository(db)
	service := NewService(repo)

	// Create gaps and mark helpful
	gapID1, _ := service.MarkGap("CBIN-140", "Feature1", "CLI", "logic_error", "Problem 1", "Solution 1", "test")
	gapID2, _ := service.MarkGap("CBIN-140", "Feature2", "API", "test_failure", "Problem 2", "Solution 2", "test")

	repo.MarkHelpful(gapID1)
	repo.MarkHelpful(gapID2)

	// Format for injection
	output, err := service.FormatGapsForInjection("CBIN-140")
	if err != nil {
		t.Fatalf("FormatGapsForInjection failed: %v", err)
	}

	if output == "" {
		t.Error("Expected non-empty output")
	}

	// Verify output contains key information
	if !strings.Contains(output, "Past Implementation Gaps") {
		t.Error("Output missing header")
	}
	if !strings.Contains(output, "Problem 1") {
		t.Error("Output missing problem description")
	}
	if !strings.Contains(output, "Solution 1") {
		t.Error("Output missing solution")
	}
}

// CANARY: REQ=CBIN-140; FEATURE="GapServiceTests"; ASPECT=Engine; STATUS=IMPL; TEST=TestService_UpdateConfig; UPDATED=2025-10-17
func TestService_UpdateConfig(t *testing.T) {
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

	repo := storage.NewGapRepository(db)
	service := NewService(repo)

	// Update config
	err = service.UpdateConfig(20, 2, "weighted")
	if err != nil {
		t.Fatalf("UpdateConfig failed: %v", err)
	}

	// Retrieve and verify
	config, _ := service.GetConfig()
	if config.MaxGapInjection != 20 {
		t.Errorf("MaxGapInjection = %d, want 20", config.MaxGapInjection)
	}
	if config.RankingStrategy != "weighted" {
		t.Errorf("RankingStrategy = %s, want weighted", config.RankingStrategy)
	}

	// Test invalid strategy
	err = service.UpdateConfig(10, 1, "invalid")
	if err == nil {
		t.Error("Expected error for invalid ranking strategy")
	}
}

// CANARY: REQ=CBIN-140; FEATURE="GapServiceTests"; ASPECT=Engine; STATUS=IMPL; TEST=TestService_GenerateGapID; UPDATED=2025-10-17
func TestService_GenerateGapID(t *testing.T) {
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

	repo := storage.NewGapRepository(db)
	service := NewService(repo)

	// Create first gap
	gapID1, _ := service.MarkGap("CBIN-140", "Feature1", "CLI", "logic_error", "Description 1", "", "test")
	if gapID1 != "GAP-CBIN-140-001" {
		t.Errorf("First gap ID = %s, want GAP-CBIN-140-001", gapID1)
	}

	// Create second gap
	gapID2, _ := service.MarkGap("CBIN-140", "Feature2", "API", "test_failure", "Description 2", "", "test")
	if gapID2 != "GAP-CBIN-140-002" {
		t.Errorf("Second gap ID = %s, want GAP-CBIN-140-002", gapID2)
	}
}
