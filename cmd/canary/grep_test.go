// Copyright (c) 2024 by Developer Network.
//
// For more details, see the LICENSE file in the root directory of this
// source code repository or contact Developer Network at info@devnw.com.


package main

import (
	"path/filepath"
	"testing"

	"go.spyder.org/canary/internal/storage"
)

// CANARY: REQ=CBIN-CLI-001; FEATURE="GrepCmdTests"; ASPECT=CLI; STATUS=TESTED; TEST=TestCANARY_CBIN_CLI_001_CLI_GrepCmd; UPDATED=2025-10-16
func TestCANARY_CBIN_CLI_001_CLI_GrepCmd(t *testing.T) {
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

	// Insert test tokens
	testTokens := []*storage.Token{
		{
			ReqID:    "CBIN-TEST-001",
			Feature:  "UserAuth",
			Aspect:   "API",
			Status:   "IMPL",
			FilePath: "internal/auth/handler.go",
			Test:     "TestUserAuth",
		},
		{
			ReqID:    "CBIN-TEST-002",
			Feature:  "UserProfile",
			Aspect:   "API",
			Status:   "TESTED",
			FilePath: "internal/user/profile.go",
			Test:     "TestUserProfile",
		},
		{
			ReqID:    "CBIN-TEST-003",
			Feature:  "DataValidation",
			Aspect:   "Engine",
			Status:   "IMPL",
			FilePath: "internal/validation/validator.go",
		},
	}

	for _, token := range testTokens {
		if err := db.UpsertToken(token); err != nil {
			t.Fatalf("Failed to upsert token: %v", err)
		}
	}

	tests := []struct {
		name        string
		pattern     string
		wantMatches int
		wantReqIDs  []string
	}{
		{
			name:        "search by feature name",
			pattern:     "User",
			wantMatches: 2,
			wantReqIDs:  []string{"CBIN-TEST-001", "CBIN-TEST-002"},
		},
		{
			name:        "search by file path",
			pattern:     "internal/auth",
			wantMatches: 1,
			wantReqIDs:  []string{"CBIN-TEST-001"},
		},
		{
			name:        "search by test name",
			pattern:     "TestUser",
			wantMatches: 2,
			wantReqIDs:  []string{"CBIN-TEST-001", "CBIN-TEST-002"},
		},
		{
			name:        "no matches",
			pattern:     "NonExistent",
			wantMatches: 0,
			wantReqIDs:  []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tokens, err := grepTokens(db, tt.pattern)
			if err != nil {
				t.Fatalf("grepTokens failed: %v", err)
			}

			if len(tokens) != tt.wantMatches {
				t.Errorf("got %d matches, want %d", len(tokens), tt.wantMatches)
			}

			// Verify we got the expected requirement IDs
			gotReqIDs := make(map[string]bool)
			for _, token := range tokens {
				gotReqIDs[token.ReqID] = true
			}

			for _, wantID := range tt.wantReqIDs {
				if !gotReqIDs[wantID] {
					t.Errorf("expected to find %s in results", wantID)
				}
			}
		})
	}
}

// CANARY: REQ=CBIN-CLI-001; FEATURE="GrepCmdTests"; ASPECT=CLI; STATUS=TESTED; TEST=TestCANARY_CBIN_CLI_001_CLI_GrepCmd; UPDATED=2025-10-16
func TestCANARY_CBIN_CLI_001_CLI_GrepCmd_CaseInsensitive(t *testing.T) {
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

	// Insert test token
	token := &storage.Token{
		ReqID:    "CBIN-TEST",
		Feature:  "UserAuthentication",
		Aspect:   "API",
		Status:   "IMPL",
		FilePath: "auth.go",
	}

	if err := db.UpsertToken(token); err != nil {
		t.Fatalf("Failed to upsert token: %v", err)
	}

	// Test case-insensitive search
	tests := []struct {
		pattern string
		want    bool
	}{
		{"user", true},
		{"USER", true},
		{"User", true},
		{"authentication", true},
		{"AUTHENTICATION", true},
	}

	for _, tt := range tests {
		t.Run(tt.pattern, func(t *testing.T) {
			tokens, err := grepTokens(db, tt.pattern)
			if err != nil {
				t.Fatalf("grepTokens failed: %v", err)
			}

			found := len(tokens) > 0
			if found != tt.want {
				t.Errorf("pattern %q: got found=%v, want %v", tt.pattern, found, tt.want)
			}
		})
	}
}

// CANARY: REQ=CBIN-CLI-001; FEATURE="GrepCmdTests"; ASPECT=CLI; STATUS=TESTED; TEST=TestCANARY_CBIN_CLI_001_CLI_GrepCmd; UPDATED=2025-10-16
func TestCANARY_CBIN_CLI_001_CLI_GrepCmd_EmptyPattern(t *testing.T) {
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

	// Empty pattern should return no results
	tokens, err := grepTokens(db, "")
	if err != nil {
		t.Fatalf("grepTokens failed: %v", err)
	}

	if len(tokens) != 0 {
		t.Errorf("empty pattern should return 0 results, got %d", len(tokens))
	}
}
