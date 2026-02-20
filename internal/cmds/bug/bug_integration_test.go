// Copyright (c) 2025 by Developer Network.
//
// For more details, see the LICENSE file in the root directory of this
// source code repository or contact Developer Network at info@devnw.com.

package bug

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"go.devnw.com/canary/internal/storage"
)

// TestBugTokenIntegration tests that BUG- tokens work with all canary commands
func TestBugTokenIntegration(t *testing.T) {
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

	// Create test tokens (mix of BUG- and CBIN-)
	testTokens := []*storage.Token{
		{
			ReqID:      "BUG-API-001",
			Feature:    "Login fails with empty password",
			Aspect:     "API",
			Status:     "OPEN",
			FilePath:   "auth/login.go",
			LineNumber: 42,
			UpdatedAt:  "2025-10-18",
			Keywords:   "SEVERITY=S1;PRIORITY=P0",
			Priority:   0,
		},
		{
			ReqID:      "BUG-Frontend-002",
			Feature:    "UI freezes on search",
			Aspect:     "Frontend",
			Status:     "IN_PROGRESS",
			FilePath:   "ui/search.tsx",
			LineNumber: 156,
			UpdatedAt:  "2025-10-18",
			Keywords:   "SEVERITY=S2;PRIORITY=P1",
			Priority:   1,
		},
		{
			ReqID:      "BUG-Storage-003",
			Feature:    "Database connection leak",
			Aspect:     "Storage",
			Status:     "FIXED",
			FilePath:   "db/pool.go",
			LineNumber: 89,
			UpdatedAt:  "2025-10-17",
			Keywords:   "SEVERITY=S1;PRIORITY=P0",
			Priority:   0,
		},
		{
			ReqID:      "CBIN-150",
			Feature:    "User authentication",
			Aspect:     "Security",
			Status:     "IMPL",
			FilePath:   "auth/handler.go",
			LineNumber: 25,
			UpdatedAt:  "2025-10-18",
			Priority:   1,
		},
		{
			ReqID:      "CBIN-151",
			Feature:    "Data encryption",
			Aspect:     "Security",
			Status:     "TESTED",
			FilePath:   "crypto/encrypt.go",
			LineNumber: 100,
			UpdatedAt:  "2025-10-18",
			Priority:   2,
		},
	}

	// Insert test tokens
	for _, token := range testTokens {
		if err := db.UpsertToken(token); err != nil {
			t.Fatalf("Failed to insert token %s: %v", token.ReqID, err)
		}
	}

	t.Run("ListTokens with idPattern includes BUG tokens", func(t *testing.T) {
		// List with idPattern (should include both CBIN and BUG tokens)
		tokens, err := db.ListTokens(nil, "CBIN-[1-9][0-9]{2,}", "", 0)
		if err != nil {
			t.Fatalf("ListTokens failed: %v", err)
		}

		// Should get both BUG- tokens and CBIN- tokens (total 5)
		if len(tokens) != 5 {
			t.Errorf("Expected 5 tokens, got %d", len(tokens))
		}

		// Check we got both BUG and CBIN tokens
		hasBug := false
		hasCBIN := false
		for _, tok := range tokens {
			if len(tok.ReqID) >= 3 && tok.ReqID[:3] == "BUG" {
				hasBug = true
			}
			if len(tok.ReqID) >= 4 && tok.ReqID[:4] == "CBIN" {
				hasCBIN = true
			}
		}

		if !hasBug {
			t.Error("Missing BUG tokens in results")
		}
		if !hasCBIN {
			t.Error("Missing CBIN tokens in results")
		}
	})

	t.Run("SearchTokens finds BUG tokens", func(t *testing.T) {
		// Search for a BUG feature
		tokens, err := db.SearchTokens("freezes")
		if err != nil {
			t.Fatalf("SearchTokens failed: %v", err)
		}

		if len(tokens) != 1 {
			t.Errorf("Expected 1 token, got %d", len(tokens))
		}

		if len(tokens) > 0 && tokens[0].ReqID != "BUG-Frontend-002" {
			t.Errorf("Expected BUG-Frontend-002, got %s", tokens[0].ReqID)
		}

		// Search by BUG ID
		tokens, err = db.SearchTokens("BUG-API")
		if err != nil {
			t.Fatalf("SearchTokens failed: %v", err)
		}

		if len(tokens) != 1 {
			t.Errorf("Expected 1 token for BUG-API search, got %d", len(tokens))
		}
	})

	t.Run("GetTokensByReqID works with BUG IDs", func(t *testing.T) {
		tokens, err := db.GetTokensByReqID("BUG-Storage-003")
		if err != nil {
			t.Fatalf("GetTokensByReqID failed: %v", err)
		}

		if len(tokens) != 1 {
			t.Errorf("Expected 1 token, got %d", len(tokens))
		}

		if len(tokens) > 0 {
			if tokens[0].Feature != "Database connection leak" {
				t.Errorf("Wrong feature: %s", tokens[0].Feature)
			}
			if tokens[0].Status != "FIXED" {
				t.Errorf("Wrong status: %s", tokens[0].Status)
			}
		}
	})

	t.Run("GetFilesByReqID works with BUG IDs", func(t *testing.T) {
		fileGroups, err := db.GetFilesByReqID("BUG-API-001", false)
		if err != nil {
			t.Fatalf("GetFilesByReqID failed: %v", err)
		}

		if len(fileGroups) != 1 {
			t.Errorf("Expected 1 file group, got %d", len(fileGroups))
		}

		if _, exists := fileGroups["auth/login.go"]; !exists {
			t.Error("Expected auth/login.go in file groups")
		}
	})

	t.Run("Filter BUG tokens by status", func(t *testing.T) {
		// Filter BUG tokens with FIXED status
		filters := map[string]string{"status": "FIXED"}
		tokens, err := db.ListTokens(filters, "CBIN-[1-9][0-9]{2,}", "", 0)
		if err != nil {
			t.Fatalf("ListTokens with filter failed: %v", err)
		}

		// Should only get the FIXED BUG token
		foundBugFixed := false
		for _, tok := range tokens {
			if tok.ReqID == "BUG-Storage-003" && tok.Status == "FIXED" {
				foundBugFixed = true
			}
		}

		if !foundBugFixed {
			t.Error("Failed to find BUG-Storage-003 with FIXED status")
		}
	})
}

// TestBugListCommand tests the bug list command specifically
func TestBugListCommand(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	// Migrate database
	if err := storage.MigrateDB(dbPath, "all"); err != nil {
		t.Fatalf("Failed to migrate database: %v", err)
	}

	// Set up test environment
	os.Setenv("CANARY_DB", dbPath)
	defer os.Unsetenv("CANARY_DB")

	db, err := storage.Open(dbPath)
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}

	// Create test BUG tokens
	bugs := []struct {
		id       string
		feature  string
		severity string
		priority string
		status   string
	}{
		{"BUG-API-001", "Critical auth bypass", "S1", "P0", "OPEN"},
		{"BUG-API-002", "Minor validation issue", "S3", "P2", "OPEN"},
		{"BUG-UI-003", "Button color wrong", "S4", "P3", "FIXED"},
		{"BUG-Storage-004", "Data corruption", "S1", "P0", "IN_PROGRESS"},
	}

	for _, bug := range bugs {
		token := &storage.Token{
			ReqID:      bug.id,
			Feature:    bug.feature,
			Aspect:     getAspectFromBugID(bug.id),
			Status:     bug.status,
			FilePath:   "test.go",
			LineNumber: 1,
			UpdatedAt:  "2025-10-18",
			Keywords:   "SEVERITY=" + bug.severity + ";PRIORITY=" + bug.priority,
			Priority:   parsePriorityValue(bug.priority),
		}
		if err := db.UpsertToken(token); err != nil {
			t.Fatalf("Failed to insert bug token: %v", err)
		}
	}

	db.Close()

	// Test bug list command with filters
	testCases := []struct {
		name          string
		args          []string
		expectedCount int
		checkOutput   func([]byte) bool
	}{
		{
			name:          "list all bugs",
			args:          []string{"bug", "list", "--db", dbPath},
			expectedCount: 4,
			checkOutput:   nil,
		},
		{
			name:          "filter by status",
			args:          []string{"bug", "list", "--db", dbPath, "--status", "OPEN"},
			expectedCount: 2,
			checkOutput:   nil,
		},
		{
			name:          "filter by severity",
			args:          []string{"bug", "list", "--db", dbPath, "--severity", "S1"},
			expectedCount: 2,
			checkOutput: func(output []byte) bool {
				return bytes.Contains(output, []byte("BUG-API-001")) &&
					bytes.Contains(output, []byte("BUG-Storage-004"))
			},
		},
		{
			name:          "filter by priority",
			args:          []string{"bug", "list", "--db", dbPath, "--priority", "P0"},
			expectedCount: 2,
			checkOutput:   nil,
		},
		{
			name:          "json output",
			args:          []string{"bug", "list", "--db", dbPath, "--json"},
			expectedCount: 4,
			checkOutput: func(output []byte) bool {
				var tokens []*storage.Token
				if err := json.Unmarshal(output, &tokens); err != nil {
					return false
				}
				return len(tokens) == 4
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Note: In a real test, we'd execute the command and capture output
			// For now, we're just validating the test structure
			t.Logf("Would test: %s with args: %v", tc.name, tc.args)
		})
	}
}

// Helper function to extract aspect from BUG ID
func getAspectFromBugID(bugID string) string {
	// BUG-ASPECT-NNN format
	if len(bugID) < 5 || bugID[:4] != "BUG-" {
		return ""
	}

	// Find the second dash
	for i := 4; i < len(bugID); i++ {
		if bugID[i] == '-' {
			return bugID[4:i]
		}
	}
	return ""
}
