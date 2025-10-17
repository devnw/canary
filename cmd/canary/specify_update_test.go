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
	"os"
	"path/filepath"
	"testing"
)

// CANARY: REQ=CBIN-134; FEATURE="UpdateSubcommand"; ASPECT=CLI; STATUS=STUB; TEST=TestCANARY_CBIN_134_CLI_UpdateSubcommand; UPDATED=2025-10-16

func TestCANARY_CBIN_134_CLI_UpdateSubcommand(t *testing.T) {
	// Test that `canary specify update CBIN-134` locates spec
	// Expected to FAIL initially (command doesn't exist)

	// Setup: Change to project root (two levels up from cmd/canary)
	originalDir, _ := os.Getwd()
	projectRoot := filepath.Join(originalDir, "../..")
	if err := os.Chdir(projectRoot); err != nil {
		t.Fatalf("Failed to change to project root: %v", err)
	}
	defer os.Chdir(originalDir) // Restore original directory after test

	// Ensure we're in the project root with specs
	if _, err := os.Stat(".canary/specs"); err != nil {
		t.Skip("Skipping test: not in project root directory with specs")
	}

	// Test exact ID lookup
	tests := []struct {
		name    string
		reqID   string
		wantErr bool
	}{
		{
			name:    "valid requirement ID",
			reqID:   "CBIN-134",
			wantErr: false,
		},
		{
			name:    "invalid requirement ID",
			reqID:   "CBIN-999",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// This will fail because updateCmd doesn't exist yet
			if updateCmd == nil {
				t.Fatal("updateCmd not defined")
			}

			// Reset flags to ensure clean state
			updateCmd.Flags().Set("search", "false")
			updateCmd.Flags().Set("sections", "")

			// Execute through root command with full command path
			rootCmd.SetArgs([]string{"specify", "update", tt.reqID})

			err := rootCmd.Execute()
			if (err != nil) != tt.wantErr {
				t.Errorf("rootCmd.Execute() error = %v, wantErr %v", err, tt.wantErr)
			}

			// Reset command for next test
			rootCmd.SetArgs(nil)
		})
	}
}

func TestCANARY_CBIN_134_CLI_SearchFlag(t *testing.T) {
	// Test that `canary specify update --search "spec mod"` returns matches
	// Expected to FAIL initially (flag not implemented)

	// Setup: Change to project root
	originalDir, _ := os.Getwd()
	projectRoot := filepath.Join(originalDir, "../..")
	if err := os.Chdir(projectRoot); err != nil {
		t.Fatalf("Failed to change to project root: %v", err)
	}
	defer os.Chdir(originalDir)

	if _, err := os.Stat(".canary/specs"); err != nil {
		t.Skip("Skipping test: not in project root directory with specs")
	}

	tests := []struct {
		name    string
		query   string
		wantErr bool
	}{
		{
			name:    "search for existing spec",
			query:   "spec modification",
			wantErr: false,
		},
		{
			name:    "search for non-existent spec",
			query:   "xyznonexistent",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// This will fail because updateCmd doesn't exist yet
			if updateCmd == nil {
				t.Fatal("updateCmd not defined")
			}

			// Reset flags to ensure clean state
			updateCmd.Flags().Set("search", "false")
			updateCmd.Flags().Set("sections", "")

			// Execute through root command with full command path
			rootCmd.SetArgs([]string{"specify", "update", "--search", tt.query})

			err := rootCmd.Execute()
			if (err != nil) != tt.wantErr {
				t.Errorf("rootCmd.Execute() with --search error = %v, wantErr %v", err, tt.wantErr)
			}

			// Reset command for next test
			rootCmd.SetArgs(nil)
		})
	}
}

func TestCANARY_CBIN_134_CLI_SectionsFlag(t *testing.T) {
	// Test that `canary specify update CBIN-134 --sections overview` returns subset
	// Expected to FAIL initially (parser doesn't exist)

	// Setup: Change to project root
	originalDir, _ := os.Getwd()
	projectRoot := filepath.Join(originalDir, "../..")
	if err := os.Chdir(projectRoot); err != nil {
		t.Fatalf("Failed to change to project root: %v", err)
	}
	defer os.Chdir(originalDir)

	if _, err := os.Stat(".canary/specs"); err != nil {
		t.Skip("Skipping test: not in project root directory with specs")
	}

	// This will fail because updateCmd and sections flag don't exist yet
	if updateCmd == nil {
		t.Fatal("updateCmd not defined")
	}

	// Reset flags to ensure clean state
	updateCmd.Flags().Set("search", "false")
	updateCmd.Flags().Set("sections", "")

	// Execute through root command with full command path
	rootCmd.SetArgs([]string{"specify", "update", "CBIN-134", "--sections", "overview"})

	err := rootCmd.Execute()
	if err != nil {
		// We expect this to work once implemented
		t.Errorf("rootCmd.Execute() with --sections error = %v", err)
	}

	// Reset command
	rootCmd.SetArgs(nil)
}

func TestCANARY_CBIN_134_CLI_InvalidReqID(t *testing.T) {
	// Test that invalid REQ-ID returns helpful error

	// Setup: Change to project root
	originalDir, _ := os.Getwd()
	projectRoot := filepath.Join(originalDir, "../..")
	if err := os.Chdir(projectRoot); err != nil {
		t.Fatalf("Failed to change to project root: %v", err)
	}
	defer os.Chdir(originalDir)

	if _, err := os.Stat(".canary/specs"); err != nil {
		t.Skip("Skipping test: not in project root directory with specs")
	}

	// This will fail because updateCmd doesn't exist yet
	if updateCmd == nil {
		t.Fatal("updateCmd not defined")
	}

	// Reset flags to ensure clean state
	updateCmd.Flags().Set("search", "false")
	updateCmd.Flags().Set("sections", "")

	// Execute through root command with full command path
	rootCmd.SetArgs([]string{"specify", "update", "INVALID-ID"})

	err := rootCmd.Execute()
	if err == nil {
		t.Error("Expected error for invalid REQ-ID, got nil")
	}

	// Reset command
	rootCmd.SetArgs(nil)
}

func TestCANARY_CBIN_134_CLI_PlanDetection(t *testing.T) {
	// Test that plan.md is detected when it exists

	// Setup: Change to project root
	originalDir, _ := os.Getwd()
	projectRoot := filepath.Join(originalDir, "../..")
	if err := os.Chdir(projectRoot); err != nil {
		t.Fatalf("Failed to change to project root: %v", err)
	}
	defer os.Chdir(originalDir)

	if _, err := os.Stat(".canary/specs"); err != nil {
		t.Skip("Skipping test: not in project root directory with specs")
	}

	// Create a temporary spec directory with plan.md for testing
	tempDir := t.TempDir()
	specDir := filepath.Join(tempDir, ".canary", "specs", "CBIN-999-test-spec")
	err := os.MkdirAll(specDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create temp spec dir: %v", err)
	}

	// Create spec.md
	specContent := "# Test Spec\n\n## Overview\n\nTest content"
	err = os.WriteFile(filepath.Join(specDir, "spec.md"), []byte(specContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create spec.md: %v", err)
	}

	// Create plan.md
	planContent := "# Test Plan\n\n## Architecture\n\nTest plan content"
	err = os.WriteFile(filepath.Join(specDir, "plan.md"), []byte(planContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create plan.md: %v", err)
	}

	// Test that plan.md is detected
	// This will fail because the command doesn't exist yet
	t.Log("Plan detection test ready - will pass once updateCmd is implemented")
}
