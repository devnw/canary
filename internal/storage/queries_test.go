// Copyright (c) 2025 by Developer Network.
//
// For more details, see the LICENSE file in the root directory of this
// source code repository or contact Developer Network at info@devnw.com.

package storage

import (
	"path/filepath"
	"testing"
)

// CANARY: REQ=CBIN-CLI-001; FEATURE="QueryAbstractionTests"; ASPECT=Storage; STATUS=TESTED; TEST=TestCANARY_CBIN_CLI_001_Storage_GetFilesByReqID; UPDATED=2025-10-16
func TestCANARY_CBIN_CLI_001_Storage_GetFilesByReqID(t *testing.T) {
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

	// Insert test tokens
	testTokens := []*Token{
		{
			ReqID:    "CBIN-QUERY-001",
			Feature:  "ImplFeature",
			Aspect:   "API",
			Status:   "IMPL",
			FilePath: "internal/api/handler.go",
		},
		{
			ReqID:    "CBIN-QUERY-001",
			Feature:  "ImplFeature2",
			Aspect:   "API",
			Status:   "IMPL",
			FilePath: "internal/api/handler.go", // Same file
		},
		{
			ReqID:    "CBIN-QUERY-001",
			Feature:  "SpecFeature",
			Aspect:   "Docs",
			Status:   "STUB",
			FilePath: ".canary/specs/CBIN-QUERY-001/spec.md",
		},
		{
			ReqID:    "CBIN-QUERY-001",
			Feature:  "PlanFeature",
			Aspect:   "Docs",
			Status:   "STUB",
			FilePath: ".canary/specs/CBIN-QUERY-001/plan.md",
		},
		{
			ReqID:    "CBIN-QUERY-001",
			Feature:  "TemplateFeature",
			Aspect:   "Docs",
			Status:   "STUB",
			FilePath: ".canary/templates/test-template.md",
		},
	}

	for _, token := range testTokens {
		if err := db.UpsertToken(token); err != nil {
			t.Fatalf("Failed to upsert token: %v", err)
		}
	}

	tests := []struct {
		name         string
		reqID        string
		excludeSpecs bool
		wantFiles    int
		wantTokens   int
	}{
		{
			name:         "exclude specs and templates",
			reqID:        "CBIN-QUERY-001",
			excludeSpecs: true,
			wantFiles:    1, // Only implementation file
			wantTokens:   2, // Two tokens in the implementation file
		},
		{
			name:         "include all files",
			reqID:        "CBIN-QUERY-001",
			excludeSpecs: false,
			wantFiles:    4, // All files
			wantTokens:   5, // All tokens
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fileGroups, err := db.GetFilesByReqID(tt.reqID, tt.excludeSpecs)
			if err != nil {
				t.Fatalf("GetFilesByReqID failed: %v", err)
			}

			if len(fileGroups) != tt.wantFiles {
				t.Errorf("got %d files, want %d", len(fileGroups), tt.wantFiles)
			}

			// Count total tokens
			totalTokens := 0
			for _, tokens := range fileGroups {
				totalTokens += len(tokens)
			}

			if totalTokens != tt.wantTokens {
				t.Errorf("got %d tokens, want %d", totalTokens, tt.wantTokens)
			}
		})
	}
}

// CANARY: REQ=CBIN-CLI-001; FEATURE="QueryAbstractionTests"; ASPECT=Storage; STATUS=TESTED; TEST=TestCANARY_CBIN_CLI_001_Storage_GetFilesByReqID; UPDATED=2025-10-16
func TestCANARY_CBIN_CLI_001_Storage_GetFilesByReqID_TokenGrouping(t *testing.T) {
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

	// Multiple tokens in the same file
	testTokens := []*Token{
		{ReqID: "TEST", Feature: "F1", FilePath: "file.go"},
		{ReqID: "TEST", Feature: "F2", FilePath: "file.go"},
		{ReqID: "TEST", Feature: "F3", FilePath: "file.go"},
	}

	for _, token := range testTokens {
		if err := db.UpsertToken(token); err != nil {
			t.Fatalf("Failed to upsert token: %v", err)
		}
	}

	fileGroups, err := db.GetFilesByReqID("TEST", false)
	if err != nil {
		t.Fatalf("GetFilesByReqID failed: %v", err)
	}

	// Should have 1 file with 3 tokens
	if len(fileGroups) != 1 {
		t.Fatalf("got %d files, want 1", len(fileGroups))
	}

	tokens := fileGroups["file.go"]
	if len(tokens) != 3 {
		t.Errorf("got %d tokens in file.go, want 3", len(tokens))
	}
}

// CANARY: REQ=CBIN-CLI-001; FEATURE="QueryAbstractionTests"; ASPECT=Storage; STATUS=TESTED; TEST=TestCANARY_CBIN_CLI_001_Storage_GetFilesByReqID; UPDATED=2025-10-16
func TestCANARY_CBIN_CLI_001_Storage_ShouldExcludeFile(t *testing.T) {
	tests := []struct {
		name        string
		path        string
		wantExclude bool
	}{
		{
			name:        "spec file",
			path:        ".canary/specs/CBIN-001/spec.md",
			wantExclude: true,
		},
		{
			name:        "plan file",
			path:        ".canary/specs/CBIN-001/plan.md",
			wantExclude: true,
		},
		{
			name:        "template file",
			path:        ".canary/templates/spec-template.md",
			wantExclude: true,
		},
		{
			name:        "base file",
			path:        "base/something.go",
			wantExclude: true,
		},
		{
			name:        "implementation file",
			path:        "internal/api/handler.go",
			wantExclude: false,
		},
		{
			name:        "cmd file",
			path:        "cmd/canary/main.go",
			wantExclude: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := shouldExcludeFile(tt.path)
			if got != tt.wantExclude {
				t.Errorf("shouldExcludeFile(%q) = %v, want %v", tt.path, got, tt.wantExclude)
			}
		})
	}
}

// CANARY: REQ=CBIN-CLI-001; FEATURE="QueryAbstractionTests"; ASPECT=Storage; STATUS=TESTED; TEST=TestCANARY_CBIN_CLI_001_Storage_GetFilesByReqID; UPDATED=2025-10-16
func TestCANARY_CBIN_CLI_001_Storage_GetTokensByReqID(t *testing.T) {
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

	// Insert tokens for multiple requirements
	testTokens := []*Token{
		{ReqID: "CBIN-001", Feature: "F1", Status: "IMPL"},
		{ReqID: "CBIN-001", Feature: "F2", Status: "TESTED"},
		{ReqID: "CBIN-002", Feature: "F3", Status: "STUB"},
	}

	for _, token := range testTokens {
		if err := db.UpsertToken(token); err != nil {
			t.Fatalf("Failed to upsert token: %v", err)
		}
	}

	tests := []struct {
		name       string
		reqID      string
		wantTokens int
	}{
		{
			name:       "requirement with 2 tokens",
			reqID:      "CBIN-001",
			wantTokens: 2,
		},
		{
			name:       "requirement with 1 token",
			reqID:      "CBIN-002",
			wantTokens: 1,
		},
		{
			name:       "non-existent requirement",
			reqID:      "CBIN-999",
			wantTokens: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tokens, err := db.GetTokensByReqID(tt.reqID)
			if err != nil {
				t.Fatalf("GetTokensByReqID failed: %v", err)
			}

			if len(tokens) != tt.wantTokens {
				t.Errorf("got %d tokens, want %d", len(tokens), tt.wantTokens)
			}

			// Verify all tokens have the correct ReqID
			for _, token := range tokens {
				if token.ReqID != tt.reqID {
					t.Errorf("token has ReqID %s, want %s", token.ReqID, tt.reqID)
				}
			}
		})
	}
}
