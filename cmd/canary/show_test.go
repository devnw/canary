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

package main

import (
	"bytes"
	"encoding/json"
	"path/filepath"
	"testing"

	"go.spyder.org/canary/internal/storage"
)

// CANARY: REQ=CBIN-CLI-001; FEATURE="ShowCmdTests"; ASPECT=CLI; STATUS=TESTED; TEST=TestCANARY_CBIN_CLI_001_CLI_ShowCmd; UPDATED=2025-10-16
func TestCANARY_CBIN_CLI_001_CLI_ShowCmd(t *testing.T) {
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
			ReqID:      "CBIN-TEST-001",
			Feature:    "TestFeature1",
			Aspect:     "API",
			Status:     "IMPL",
			FilePath:   "test/file1.go",
			LineNumber: 10,
			Priority:   5,
		},
		{
			ReqID:      "CBIN-TEST-001",
			Feature:    "TestFeature2",
			Aspect:     "CLI",
			Status:     "TESTED",
			FilePath:   "test/file2.go",
			LineNumber: 20,
			Test:       "TestFeature2",
			Priority:   5,
		},
	}

	for _, token := range testTokens {
		if err := db.UpsertToken(token); err != nil {
			t.Fatalf("Failed to upsert token: %v", err)
		}
	}

	tests := []struct {
		name       string
		reqID      string
		groupBy    string
		wantTokens int
		wantErr    bool
	}{
		{
			name:       "valid requirement with tokens",
			reqID:      "CBIN-TEST-001",
			groupBy:    "aspect",
			wantTokens: 2,
			wantErr:    false,
		},
		{
			name:       "valid requirement grouped by status",
			reqID:      "CBIN-TEST-001",
			groupBy:    "status",
			wantTokens: 2,
			wantErr:    false,
		},
		{
			name:       "non-existent requirement",
			reqID:      "CBIN-INVALID",
			groupBy:    "aspect",
			wantTokens: 0,
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Query tokens
			tokens, err := db.GetTokensByReqID(tt.reqID)
			if err != nil {
				t.Fatalf("GetTokensByReqID failed: %v", err)
			}

			if len(tokens) != tt.wantTokens {
				t.Errorf("got %d tokens, want %d", len(tokens), tt.wantTokens)
			}

			if tt.wantErr && len(tokens) > 0 {
				t.Error("expected error for invalid requirement")
			}

			// Test grouping
			if len(tokens) > 0 {
				groups := groupTokens(tokens, tt.groupBy)
				if len(groups) == 0 {
					t.Error("grouping returned no groups")
				}
			}
		})
	}
}

// CANARY: REQ=CBIN-CLI-001; FEATURE="ShowCmdTests"; ASPECT=CLI; STATUS=TESTED; TEST=TestCANARY_CBIN_CLI_001_CLI_ShowCmd; UPDATED=2025-10-16
func TestCANARY_CBIN_CLI_001_CLI_ShowCmd_JSONOutput(t *testing.T) {
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

	// Insert test token
	testToken := &storage.Token{
		ReqID:      "CBIN-TEST-JSON",
		Feature:    "JSONFeature",
		Aspect:     "API",
		Status:     "IMPL",
		FilePath:   "test/json.go",
		LineNumber: 30,
		Priority:   5,
	}

	if err := db.UpsertToken(testToken); err != nil {
		t.Fatalf("Failed to upsert token: %v", err)
	}

	// Test JSON output
	tokens, err := db.GetTokensByReqID("CBIN-TEST-JSON")
	if err != nil {
		t.Fatalf("GetTokensByReqID failed: %v", err)
	}

	// Capture JSON output
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetIndent("", "  ")
	if err := enc.Encode(tokens); err != nil {
		t.Fatalf("JSON encoding failed: %v", err)
	}

	// Verify JSON is valid
	var decoded []*storage.Token
	if err := json.Unmarshal(buf.Bytes(), &decoded); err != nil {
		t.Fatalf("JSON decoding failed: %v", err)
	}

	if len(decoded) != 1 {
		t.Errorf("got %d decoded tokens, want 1", len(decoded))
	}

	if decoded[0].ReqID != "CBIN-TEST-JSON" {
		t.Errorf("got ReqID %s, want CBIN-TEST-JSON", decoded[0].ReqID)
	}
}

// CANARY: REQ=CBIN-CLI-001; FEATURE="ShowCmdTests"; ASPECT=CLI; STATUS=TESTED; TEST=TestCANARY_CBIN_CLI_001_CLI_ShowCmd; UPDATED=2025-10-16
func TestCANARY_CBIN_CLI_001_CLI_ShowCmd_Grouping(t *testing.T) {
	tokens := []*storage.Token{
		{ReqID: "TEST", Feature: "F1", Aspect: "API", Status: "IMPL"},
		{ReqID: "TEST", Feature: "F2", Aspect: "API", Status: "TESTED"},
		{ReqID: "TEST", Feature: "F3", Aspect: "CLI", Status: "IMPL"},
		{ReqID: "TEST", Feature: "F4", Aspect: "CLI", Status: "STUB"},
	}

	tests := []struct {
		name       string
		groupBy    string
		wantGroups int
	}{
		{
			name:       "group by aspect",
			groupBy:    "aspect",
			wantGroups: 2, // API, CLI
		},
		{
			name:       "group by status",
			groupBy:    "status",
			wantGroups: 3, // IMPL, TESTED, STUB
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			groups := groupTokens(tokens, tt.groupBy)
			if len(groups) != tt.wantGroups {
				t.Errorf("got %d groups, want %d", len(groups), tt.wantGroups)
			}
		})
	}
}
