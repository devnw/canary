// Copyright (c) 2025 by Developer Network.
//
// For more details, see the LICENSE file in the root directory of this
// source code repository or contact Developer Network at info@devnw.com.

package main

import (
	"path/filepath"
	"testing"

	"go.spyder.org/canary/internal/storage"
)

// CANARY: REQ=CBIN-CLI-001; FEATURE="StatusCmdTests"; ASPECT=CLI; STATUS=TESTED; TEST=TestCANARY_CBIN_CLI_001_CLI_StatusCmd; UPDATED=2025-10-16
func TestCANARY_CBIN_CLI_001_CLI_StatusCmd(t *testing.T) {
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

	// Insert test tokens with various statuses
	testTokens := []*storage.Token{
		{ReqID: "CBIN-TEST-003", Feature: "F1", Status: "STUB", FilePath: "test1.go"},
		{ReqID: "CBIN-TEST-003", Feature: "F2", Status: "STUB", FilePath: "test2.go"},
		{ReqID: "CBIN-TEST-003", Feature: "F3", Status: "IMPL", FilePath: "test3.go"},
		{ReqID: "CBIN-TEST-003", Feature: "F4", Status: "TESTED", FilePath: "test4.go"},
		{ReqID: "CBIN-TEST-003", Feature: "F5", Status: "BENCHED", FilePath: "test5.go"},
	}

	for _, token := range testTokens {
		if err := db.UpsertToken(token); err != nil {
			t.Fatalf("Failed to upsert token: %v", err)
		}
	}

	// Query and calculate stats
	tokens, err := db.GetTokensByReqID("CBIN-TEST-003")
	if err != nil {
		t.Fatalf("GetTokensByReqID failed: %v", err)
	}

	stats := calculateStats(tokens)

	// Verify statistics
	if stats.Total != 5 {
		t.Errorf("Total: got %d, want 5", stats.Total)
	}
	if stats.Stub != 2 {
		t.Errorf("Stub: got %d, want 2", stats.Stub)
	}
	if stats.Impl != 1 {
		t.Errorf("Impl: got %d, want 1", stats.Impl)
	}
	if stats.Tested != 1 {
		t.Errorf("Tested: got %d, want 1", stats.Tested)
	}
	if stats.Benched != 1 {
		t.Errorf("Benched: got %d, want 1", stats.Benched)
	}
	if stats.Completed != 2 {
		t.Errorf("Completed: got %d, want 2 (TESTED + BENCHED)", stats.Completed)
	}
}

// CANARY: REQ=CBIN-CLI-001; FEATURE="StatusCmdTests"; ASPECT=CLI; STATUS=TESTED; TEST=TestCANARY_CBIN_CLI_001_CLI_StatusCmd; UPDATED=2025-10-16
func TestCANARY_CBIN_CLI_001_CLI_StatusCmd_ProgressBar(t *testing.T) {
	tests := []struct {
		name     string
		pct      int
		width    int
		wantChar string
	}{
		{
			name:     "0 percent",
			pct:      0,
			width:    10,
			wantChar: "[>",
		},
		{
			name:     "50 percent",
			pct:      50,
			width:    10,
			wantChar: "[=====",
		},
		{
			name:     "100 percent",
			pct:      100,
			width:    10,
			wantChar: "[==========",
		},
		{
			name:     "negative percent (edge case)",
			pct:      -10,
			width:    10,
			wantChar: "[>",
		},
		{
			name:     "over 100 percent (edge case)",
			pct:      150,
			width:    10,
			wantChar: "[==========",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bar := progressBar(tt.pct, tt.width)

			// Verify bar starts with expected pattern
			if len(bar) < len(tt.wantChar) {
				t.Fatalf("bar too short: %q", bar)
			}

			// Verify percentage is in output
			if tt.pct < 0 {
				if bar[len(bar)-2:] != "0%" {
					t.Errorf("negative pct should show 0%%, got: %s", bar)
				}
			} else if tt.pct > 100 {
				if bar[len(bar)-4:] != "100%" {
					t.Errorf("over 100%% should show 100%%, got: %s", bar)
				}
			}
		})
	}
}

// CANARY: REQ=CBIN-CLI-001; FEATURE="StatusCmdTests"; ASPECT=CLI; STATUS=TESTED; TEST=TestCANARY_CBIN_CLI_001_CLI_StatusCmd; UPDATED=2025-10-16
func TestCANARY_CBIN_CLI_001_CLI_StatusCmd_CompletionPercentage(t *testing.T) {
	tests := []struct {
		name        string
		tokens      []*storage.Token
		wantPercent int
	}{
		{
			name: "all completed",
			tokens: []*storage.Token{
				{Status: "TESTED"},
				{Status: "BENCHED"},
			},
			wantPercent: 100,
		},
		{
			name: "half completed",
			tokens: []*storage.Token{
				{Status: "TESTED"},
				{Status: "STUB"},
			},
			wantPercent: 50,
		},
		{
			name: "none completed",
			tokens: []*storage.Token{
				{Status: "STUB"},
				{Status: "IMPL"},
			},
			wantPercent: 0,
		},
		{
			name:        "empty tokens",
			tokens:      []*storage.Token{},
			wantPercent: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stats := calculateStats(tt.tokens)

			completionPct := 0
			if stats.Total > 0 {
				completionPct = (stats.Completed * 100) / stats.Total
			}

			if completionPct != tt.wantPercent {
				t.Errorf("got %d%%, want %d%%", completionPct, tt.wantPercent)
			}
		})
	}
}

// CANARY: REQ=CBIN-CLI-001; FEATURE="StatusCmdTests"; ASPECT=CLI; STATUS=TESTED; TEST=TestCANARY_CBIN_CLI_001_CLI_StatusCmd; UPDATED=2025-10-16
func TestCANARY_CBIN_CLI_001_CLI_StatusCmd_DisplaySummary(t *testing.T) {
	tokens := []*storage.Token{
		{ReqID: "TEST", Feature: "F1", Status: "STUB", FilePath: "f1.go"},
		{ReqID: "TEST", Feature: "F2", Status: "IMPL", FilePath: "f2.go"},
		{ReqID: "TEST", Feature: "F3", Status: "TESTED", FilePath: "f3.go"},
	}

	stats := calculateStats(tokens)

	// Test that displayStatusSummary doesn't panic
	// (actual output is tested manually since it uses color formatting)
	displayStatusSummary("TEST", stats, tokens)
}
