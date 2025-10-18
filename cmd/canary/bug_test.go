// Copyright (c) 2025 by Developer Network.
//
// For more details, see the LICENSE file in the root directory of this
// source code repository or contact Developer Network at info@devnw.com.

package main

import (
	"path/filepath"
	"regexp"
	"testing"

	"go.devnw.com/canary/internal/storage"
)

// Test bug ID generation
func TestGenerateBugID(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	// Test with no database (should start from 001)
	bugID, err := generateBugID("API", dbPath)
	if err != nil {
		t.Fatalf("Failed to generate bug ID: %v", err)
	}

	if bugID != "BUG-API-001" {
		t.Errorf("Expected BUG-API-001, got %s", bugID)
	}

	// Test with existing database
	if err := storage.MigrateDB(dbPath, "all"); err != nil {
		t.Fatalf("Failed to migrate database: %v", err)
	}

	db, err := storage.Open(dbPath)
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	// Add some existing bug tokens
	tokens := []*storage.Token{
		{ReqID: "BUG-API-001", Feature: "Bug 1", Aspect: "API", Status: "OPEN", FilePath: "test.go", LineNumber: 1, UpdatedAt: "2025-10-18"},
		{ReqID: "BUG-API-002", Feature: "Bug 2", Aspect: "API", Status: "OPEN", FilePath: "test.go", LineNumber: 2, UpdatedAt: "2025-10-18"},
		{ReqID: "BUG-API-005", Feature: "Bug 5", Aspect: "API", Status: "OPEN", FilePath: "test.go", LineNumber: 5, UpdatedAt: "2025-10-18"},
	}

	for _, token := range tokens {
		if err := db.UpsertToken(token); err != nil {
			t.Fatalf("Failed to insert token: %v", err)
		}
	}

	// Generate next ID (should be 006)
	bugID, err = generateBugID("API", dbPath)
	if err != nil {
		t.Fatalf("Failed to generate bug ID: %v", err)
	}

	if bugID != "BUG-API-006" {
		t.Errorf("Expected BUG-API-006, got %s", bugID)
	}

	// Test with different aspect
	bugID, err = generateBugID("CLI", dbPath)
	if err != nil {
		t.Fatalf("Failed to generate bug ID: %v", err)
	}

	if bugID != "BUG-CLI-001" {
		t.Errorf("Expected BUG-CLI-001, got %s", bugID)
	}
}

// Test bug metadata parsing
func TestParseBugMetadata(t *testing.T) {
	tests := []struct {
		name             string
		keywords         string
		expectedSeverity string
		expectedPriority string
	}{
		{
			name:             "full keywords",
			keywords:         "SEVERITY=S1;PRIORITY=P0",
			expectedSeverity: "S1",
			expectedPriority: "P0",
		},
		{
			name:             "severity only",
			keywords:         "SEVERITY=S2",
			expectedSeverity: "S2",
			expectedPriority: "P2", // default
		},
		{
			name:             "priority only",
			keywords:         "PRIORITY=P1",
			expectedSeverity: "S3", // default
			expectedPriority: "P1",
		},
		{
			name:             "empty keywords",
			keywords:         "",
			expectedSeverity: "S3", // default
			expectedPriority: "P2", // default
		},
		{
			name:             "with spaces",
			keywords:         "SEVERITY=S1 ; PRIORITY=P0",
			expectedSeverity: "S1",
			expectedPriority: "P0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			severity, priority := parseBugMetadata(tt.keywords)

			if severity != tt.expectedSeverity {
				t.Errorf("Expected severity %s, got %s", tt.expectedSeverity, severity)
			}

			if priority != tt.expectedPriority {
				t.Errorf("Expected priority %s, got %s", tt.expectedPriority, priority)
			}
		})
	}
}

// Test priority value parsing
func TestParsePriorityValue(t *testing.T) {
	tests := []struct {
		priority string
		expected int
	}{
		{"P0", 0},
		{"P1", 1},
		{"P2", 2},
		{"P3", 3},
		{"", 2},      // default
		{"invalid", 2}, // default
	}

	for _, tt := range tests {
		result := parsePriorityValue(tt.priority)
		if result != tt.expected {
			t.Errorf("Priority %s: expected %d, got %d", tt.priority, tt.expected, result)
		}
	}
}

// Test bug token filtering
func TestFilterBugTokens(t *testing.T) {
	tokens := []*storage.Token{
		{ReqID: "BUG-API-001", Feature: "Bug 1", Keywords: "SEVERITY=S1;PRIORITY=P0"},
		{ReqID: "BUG-API-002", Feature: "Bug 2", Keywords: "SEVERITY=S2;PRIORITY=P1"},
		{ReqID: "BUG-API-003", Feature: "Bug 3", Keywords: "SEVERITY=S3;PRIORITY=P2"},
		{ReqID: "BUG-API-004", Feature: "Bug 4", Keywords: "SEVERITY=S4;PRIORITY=P3"},
	}

	tests := []struct {
		name         string
		severity     string
		priority     string
		expectedCount int
	}{
		{
			name:         "filter by severity S1",
			severity:     "S1",
			priority:     "",
			expectedCount: 1,
		},
		{
			name:         "filter by priority P0",
			severity:     "",
			priority:     "P0",
			expectedCount: 1,
		},
		{
			name:         "filter by multiple severities",
			severity:     "S1,S2",
			priority:     "",
			expectedCount: 2,
		},
		{
			name:         "filter by multiple priorities",
			severity:     "",
			priority:     "P0,P1",
			expectedCount: 2,
		},
		{
			name:         "filter by both",
			severity:     "S1",
			priority:     "P0",
			expectedCount: 1,
		},
		{
			name:         "no filters",
			severity:     "",
			priority:     "",
			expectedCount: 4,
		},
		{
			name:         "non-matching filter",
			severity:     "S5",
			priority:     "",
			expectedCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filtered := filterBugTokens(tokens, tt.severity, tt.priority)

			if len(filtered) != tt.expectedCount {
				t.Errorf("Expected %d tokens, got %d", tt.expectedCount, len(filtered))
			}
		})
	}
}

// Test bug ID validation
func TestBugIDValidation(t *testing.T) {
	validPattern := regexp.MustCompile(`^BUG-[A-Za-z]+-[0-9]{3}$`)

	tests := []struct {
		bugID   string
		isValid bool
	}{
		{"BUG-API-001", true},
		{"BUG-CLI-123", true},
		{"BUG-Storage-999", true},
		{"BUG-api-001", true}, // lowercase aspect
		{"CBIN-001", false},   // wrong prefix
		{"BUG-API-1", false},  // not 3 digits
		{"BUG-API-1234", false}, // too many digits
		{"BUG-API", false},      // no number
		{"BUG--001", false},     // empty aspect
		{"", false},             // empty
	}

	for _, tt := range tests {
		result := validPattern.MatchString(tt.bugID)
		if result != tt.isValid {
			t.Errorf("Bug ID %s: expected valid=%v, got %v", tt.bugID, tt.isValid, result)
		}
	}
}

// Integration test for bug commands
func TestBugCommandIntegration(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	// Migrate database
	if err := storage.MigrateDB(dbPath, "all"); err != nil {
		t.Fatalf("Failed to migrate database: %v", err)
	}

	db, err := storage.Open(dbPath)
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	t.Run("create bug tokens", func(t *testing.T) {
		// Create several bug tokens
		bugs := []struct {
			title    string
			aspect   string
			severity string
			priority string
			status   string
		}{
			{"Login fails", "API", "S1", "P0", "OPEN"},
			{"UI freezes", "Frontend", "S2", "P1", "OPEN"},
			{"Memory leak", "Engine", "S1", "P0", "IN_PROGRESS"},
			{"Typo in docs", "Docs", "S4", "P3", "FIXED"},
		}

		for i, bug := range bugs {
			token := &storage.Token{
				ReqID:      generateTestBugID(bug.aspect, i+1),
				Feature:    bug.title,
				Aspect:     bug.aspect,
				Status:     bug.status,
				FilePath:   "test.go",
				LineNumber: i + 1,
				UpdatedAt:  "2025-10-18",
				Priority:   parsePriorityValue(bug.priority),
				Keywords:   "SEVERITY=" + bug.severity + ";PRIORITY=" + bug.priority,
			}

			if err := db.UpsertToken(token); err != nil {
				t.Fatalf("Failed to create bug token: %v", err)
			}
		}
	})

	t.Run("list bug tokens", func(t *testing.T) {
		// List all BUG tokens
		tokens, err := db.ListTokens(nil, "BUG-[A-Za-z]+-[0-9]{3}", "priority ASC", 0)
		if err != nil {
			t.Fatalf("Failed to list bugs: %v", err)
		}

		if len(tokens) != 4 {
			t.Errorf("Expected 4 bugs, got %d", len(tokens))
		}

		// Verify ordering by priority
		if len(tokens) > 1 && tokens[0].Priority > tokens[1].Priority {
			t.Error("Bugs not sorted by priority")
		}
	})

	t.Run("filter by status", func(t *testing.T) {
		filters := map[string]string{"status": "OPEN"}
		tokens, err := db.ListTokens(filters, "BUG-[A-Za-z]+-[0-9]{3}", "priority ASC", 0)
		if err != nil {
			t.Fatalf("Failed to filter bugs: %v", err)
		}

		if len(tokens) != 2 {
			t.Errorf("Expected 2 OPEN bugs, got %d", len(tokens))
		}

		for _, token := range tokens {
			if token.Status != "OPEN" {
				t.Errorf("Expected status OPEN, got %s", token.Status)
			}
		}
	})

	t.Run("update bug status", func(t *testing.T) {
		// Get a bug to update
		tokens, err := db.GetTokensByReqID("BUG-API-001")
		if err != nil || len(tokens) == 0 {
			t.Fatalf("Failed to find bug to update")
		}

		token := tokens[0]
		token.Status = "FIXED"
		token.UpdatedAt = "2025-10-19"

		if err := db.UpsertToken(token); err != nil {
			t.Fatalf("Failed to update bug: %v", err)
		}

		// Verify update
		updated, err := db.GetTokensByReqID("BUG-API-001")
		if err != nil || len(updated) == 0 {
			t.Fatalf("Failed to verify bug update")
		}

		if updated[0].Status != "FIXED" {
			t.Errorf("Expected status FIXED, got %s", updated[0].Status)
		}
	})
}

// Helper function for tests
func generateTestBugID(aspect string, num int) string {
	return "BUG-" + aspect + "-" + padNumber(num, 3)
}

func padNumber(num int, width int) string {
	s := ""
	for i := 0; i < width; i++ {
		s = "0" + s
	}
	numStr := s + string(rune('0'+num%10))
	if num >= 10 {
		numStr = s[:len(s)-1] + string(rune('0'+num/10)) + string(rune('0'+num%10))
	}
	if num >= 100 {
		numStr = string(rune('0'+num/100)) + string(rune('0'+(num%100)/10)) + string(rune('0'+num%10))
	}
	return numStr[len(numStr)-width:]
}
