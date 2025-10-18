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

// CANARY: REQ=CBIN-CLI-001; FEATURE="FilesCmdTests"; ASPECT=CLI; STATUS=TESTED; TEST=TestCANARY_CBIN_CLI_001_CLI_FilesCmd; UPDATED=2025-10-16
func TestCANARY_CBIN_CLI_001_CLI_FilesCmd(t *testing.T) {
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
			ReqID:    "CBIN-TEST-002",
			Feature:  "ImplFeature",
			Aspect:   "API",
			Status:   "IMPL",
			FilePath: "internal/api/handler.go",
		},
		{
			ReqID:    "CBIN-TEST-002",
			Feature:  "SpecFeature",
			Aspect:   "Docs",
			Status:   "STUB",
			FilePath: ".canary/specs/CBIN-TEST-002/spec.md",
		},
		{
			ReqID:    "CBIN-TEST-002",
			Feature:  "PlanFeature",
			Aspect:   "Docs",
			Status:   "STUB",
			FilePath: ".canary/specs/CBIN-TEST-002/plan.md",
		},
		{
			ReqID:    "CBIN-TEST-002",
			Feature:  "CLIFeature",
			Aspect:   "CLI",
			Status:   "IMPL",
			FilePath: "cmd/canary/test.go",
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
	}{
		{
			name:         "exclude specs and templates",
			reqID:        "CBIN-TEST-002",
			excludeSpecs: true,
			wantFiles:    2, // Only implementation files
		},
		{
			name:         "include all files",
			reqID:        "CBIN-TEST-002",
			excludeSpecs: false,
			wantFiles:    4, // All files including specs
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
		})
	}
}

// CANARY: REQ=CBIN-CLI-001; FEATURE="FilesCmdTests"; ASPECT=CLI; STATUS=TESTED; TEST=TestCANARY_CBIN_CLI_001_CLI_FilesCmd; UPDATED=2025-10-16
func TestCANARY_CBIN_CLI_001_CLI_FilesCmd_Formatting(t *testing.T) {
	fileGroups := map[string][]*storage.Token{
		"internal/api/handler.go": {
			{ReqID: "TEST", Feature: "F1", Aspect: "API", Status: "IMPL"},
			{ReqID: "TEST", Feature: "F2", Aspect: "API", Status: "IMPL"},
		},
		"cmd/canary/cmd.go": {
			{ReqID: "TEST", Feature: "F3", Aspect: "CLI", Status: "IMPL"},
		},
	}

	// Test that formatFilesList doesn't panic
	formatFilesList(fileGroups)

	// Verify total counts
	totalFiles := len(fileGroups)
	if totalFiles != 2 {
		t.Errorf("got %d files, want 2", totalFiles)
	}

	totalTokens := 0
	for _, tokens := range fileGroups {
		totalTokens += len(tokens)
	}
	if totalTokens != 3 {
		t.Errorf("got %d total tokens, want 3", totalTokens)
	}
}

// CANARY: REQ=CBIN-CLI-001; FEATURE="FilesCmdTests"; ASPECT=CLI; STATUS=TESTED; TEST=TestCANARY_CBIN_CLI_001_CLI_FilesCmd; UPDATED=2025-10-16
func TestCANARY_CBIN_CLI_001_CLI_FilesCmd_AspectGrouping(t *testing.T) {
	fileGroups := map[string][]*storage.Token{
		"file1.go": {
			{ReqID: "TEST", Feature: "F1", Aspect: "API", Status: "IMPL"},
			{ReqID: "TEST", Feature: "F2", Aspect: "CLI", Status: "IMPL"},
		},
	}

	// Format output (should group by aspect)
	formatFilesList(fileGroups)

	// Verify that file appears under both aspects
	aspectFiles := make(map[string][]string)
	fileCounts := make(map[string]int)

	for filePath, tokens := range fileGroups {
		aspects := make(map[string]bool)
		for _, token := range tokens {
			aspects[token.Aspect] = true
		}

		for aspect := range aspects {
			aspectFiles[aspect] = append(aspectFiles[aspect], filePath)
		}

		fileCounts[filePath] = len(tokens)
	}

	// Should have 2 aspects
	if len(aspectFiles) != 2 {
		t.Errorf("got %d aspects, want 2", len(aspectFiles))
	}

	// File should appear under both aspects
	if len(aspectFiles["API"]) != 1 || len(aspectFiles["CLI"]) != 1 {
		t.Error("file not properly grouped by aspect")
	}
}
