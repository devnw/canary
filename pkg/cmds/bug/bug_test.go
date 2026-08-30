// Copyright (c) 2025 by Developer Network.
//
// For more details, see the LICENSE file in the root directory of this
// source code repository or contact Developer Network at info@devnw.com.

package bug

import (
	"os"
	"path/filepath"
	"regexp"
	"testing"

	"devnw.dev/canary/pkg/canaryscan"
	"devnw.dev/canary/pkg/storage"
)

// TestReserveBugID covers bug id allocation, which is now a transactional
// reservation in storage rather than a read-max-then-increment in this
// package. The old generator could hand the same id to two callers racing on
// the same series; the reservation table's primary key makes that impossible.
func TestReserveBugID(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "test.db")
	db, err := storage.OpenRW(dbPath)
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	// An empty index starts the series at 001.
	bugID, err := db.ReserveID("default", "BUG-API")
	if err != nil {
		t.Fatalf("Failed to reserve bug ID: %v", err)
	}
	if bugID != "BUG-API-001" {
		t.Errorf("Expected BUG-API-001, got %s", bugID)
	}

	// Ids already carried by indexed tokens are never re-issued, even ones
	// that entered the index from source rather than through a reservation.
	tokens := []*storage.Token{
		{ReqID: "BUG-API-002", Feature: "Bug 2", Aspect: "API", Status: "OPEN", FilePath: "main.go", LineNumber: 2, UpdatedAt: "2025-10-18", ProjectID: "default"},
		{ReqID: "BUG-API-005", Feature: "Bug 5", Aspect: "API", Status: "OPEN", FilePath: "main.go", LineNumber: 5, UpdatedAt: "2025-10-18", ProjectID: "default"},
	}
	for _, token := range tokens {
		if err := db.UpsertToken(token); err != nil {
			t.Fatalf("Failed to insert token: %v", err)
		}
	}

	bugID, err = db.ReserveID("default", "BUG-API")
	if err != nil {
		t.Fatalf("Failed to reserve bug ID: %v", err)
	}
	if bugID != "BUG-API-006" {
		t.Errorf("Expected BUG-API-006, got %s", bugID)
	}

	// A different aspect is a different series.
	bugID, err = db.ReserveID("default", "BUG-CLI")
	if err != nil {
		t.Fatalf("Failed to reserve bug ID: %v", err)
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
		{"", 2},        // default
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
		name          string
		severity      string
		priority      string
		expectedCount int
	}{
		{
			name:          "filter by severity S1",
			severity:      "S1",
			priority:      "",
			expectedCount: 1,
		},
		{
			name:          "filter by priority P0",
			severity:      "",
			priority:      "P0",
			expectedCount: 1,
		},
		{
			name:          "filter by multiple severities",
			severity:      "S1,S2",
			priority:      "",
			expectedCount: 2,
		},
		{
			name:          "filter by multiple priorities",
			severity:      "",
			priority:      "P0,P1",
			expectedCount: 2,
		},
		{
			name:          "filter by both",
			severity:      "S1",
			priority:      "P0",
			expectedCount: 1,
		},
		{
			name:          "no filters",
			severity:      "",
			priority:      "",
			expectedCount: 4,
		},
		{
			name:          "non-matching filter",
			severity:      "S5",
			priority:      "",
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
		{"BUG-api-001", true},   // lowercase aspect
		{"CBIN-001", false},     // wrong prefix
		{"BUG-API-1", false},    // not 3 digits
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

	db, err := storage.OpenRW(dbPath)
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
				FilePath:   "main.go",
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
		// List all tokens (empty pattern) and filter BUG tokens manually
		allTokens, err := db.ListTokens("", nil, "", "", 0)
		if err != nil {
			t.Fatalf("Failed to list bugs: %v", err)
		}

		// Filter for BUG tokens
		var tokens []*storage.Token
		bugPattern := regexp.MustCompile(`^BUG-[A-Za-z]+-[0-9]{3}$`)
		for _, tok := range allTokens {
			if bugPattern.MatchString(tok.ReqID) {
				tokens = append(tokens, tok)
			}
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
		filters := map[string]any{"status": "OPEN"}
		allTokens, err := db.ListTokens("", filters, "", "", 0)
		if err != nil {
			t.Fatalf("Failed to filter bugs: %v", err)
		}

		// Filter for BUG tokens
		var tokens []*storage.Token
		bugPattern := regexp.MustCompile(`^BUG-[A-Za-z]+-[0-9]{3}$`)
		for _, tok := range allTokens {
			if bugPattern.MatchString(tok.ReqID) {
				tokens = append(tokens, tok)
			}
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
		tokens, err := db.GetTokensByReqID("", "BUG-API-001")
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
		updated, err := db.GetTokensByReqID("", "BUG-API-001")
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

// TestCANARY_CBIN_302_BugTokenSingleLineParseable verifies that bug tokens
// are emitted as single-line tokens that the scanner can parse correctly.
// CANARY: REQ=CBIN-302; FEATURE="Bug tokens are single-line parseable"; ASPECT=Encode; STATUS=TESTED; UPDATED=2026-08-29
func TestCANARY_CBIN_302_BugTokenSingleLineParseable(t *testing.T) {
	// Create a temp directory to hold the test file
	tmpRoot := t.TempDir()

	// Generate bug token using the helper function
	bugID := "BUG-API-999"
	title := "Test bug for scanner"
	aspect := "API"
	severity := "S1"
	priority := "P0"
	updatedDate := "2025-08-29"

	// Build the token string via the same code path
	token := buildBugToken(bugID, title, title, aspect, "OPEN", severity, priority, updatedDate)

	// Write to a temp file
	testFile := filepath.Join(tmpRoot, "test.go")
	err := os.WriteFile(testFile, []byte(token), 0644)
	if err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	// Scan the temp root
	rep, err := canaryscan.Scan(tmpRoot, nil, nil, nil, nil)
	if err != nil {
		t.Fatalf("Scanner failed (this is expected if token is multi-line): %v", err)
	}

	// Check that the BUG id is in requirements
	found := false
	for _, req := range rep.Requirements {
		if req.ID == bugID {
			found = true
			// Verify it has STATUS=OPEN
			if len(req.Features) == 0 {
				t.Fatalf("Bug requirement %s has no features", bugID)
			}
			if req.Features[0].Status != "OPEN" {
				t.Errorf("Expected STATUS=OPEN for %s, got %s", bugID, req.Features[0].Status)
			}
			break
		}
	}

	if !found {
		t.Errorf("Bug token %s not found in scan results. Requirements found: %v", bugID, func() []string {
			var ids []string
			for _, r := range rep.Requirements {
				ids = append(ids, r.ID)
			}
			return ids
		}())
	}
}
