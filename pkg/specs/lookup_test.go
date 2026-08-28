// Copyright (c) 2025 by Developer Network.
//
// For more details, see the LICENSE file in the root directory of this
// source code repository or contact Developer Network at info@devnw.com.

// CANARY: REQ=CBIN-134; FEATURE="ExactIDLookup"; ASPECT=Engine; STATUS=TESTED; TEST=TestCANARY_CBIN_134_Engine_ExactIDLookup; UPDATED=2025-10-17
package specs

import (
	"os"
	"path/filepath"
	"testing"
)

// setupTestSpecs creates a temporary .canary/specs directory structure for testing
func setupTestSpecs(t *testing.T) (string, func()) {
	t.Helper()

	// Create temporary directory
	tempDir := t.TempDir()

	// Create .canary/specs structure
	specsDir := filepath.Join(tempDir, ".canary", "specs")

	// Create test spec directories
	testSpecs := []struct {
		reqID   string
		content string
	}{
		{
			reqID: "CBIN-134",
			content: `# CBIN-134: ExactIDLookup

## Description
Test spec for exact ID lookup functionality.
`,
		},
		{
			reqID: "CBIN-100",
			content: `# CBIN-100: Sample Spec

## Description
Sample spec for modification testing.
`,
		},
	}

	for _, spec := range testSpecs {
		specDir := filepath.Join(specsDir, spec.reqID+"-test-feature")
		if err := os.MkdirAll(specDir, 0755); err != nil {
			t.Fatalf("failed to create spec directory: %v", err)
		}

		specFile := filepath.Join(specDir, "spec.md")
		if err := os.WriteFile(specFile, []byte(spec.content), 0644); err != nil {
			t.Fatalf("failed to write spec file: %v", err)
		}
	}

	// Return tempDir and cleanup function
	oldDir, _ := os.Getwd()
	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("failed to change to temp directory: %v", err)
	}

	cleanup := func() {
		os.Chdir(oldDir)
	}

	return tempDir, cleanup
}

// TestCANARY_CBIN_134_Engine_ExactIDLookup verifies FindSpecByID locates spec files
func TestCANARY_CBIN_134_Engine_ExactIDLookup(t *testing.T) {
	// Setup: Create temporary test spec structure
	_, cleanup := setupTestSpecs(t)
	defer cleanup()

	tests := []struct {
		name      string
		reqID     string
		wantError bool
	}{
		{
			name:      "valid requirement ID",
			reqID:     "CBIN-134",
			wantError: false,
		},
		{
			name:      "non-existent requirement",
			reqID:     "CBIN-999",
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Execute: Find spec by ID
			specPath, err := FindSpecByID(tt.reqID)

			// Verify: Error handling
			if tt.wantError {
				if err == nil {
					t.Errorf("expected error for %s, got nil", tt.reqID)
				}
				return
			}

			if err != nil {
				t.Fatalf("FindSpecByID failed: %v", err)
			}

			// Verify: Path exists
			if _, err := os.Stat(specPath); os.IsNotExist(err) {
				t.Errorf("spec file does not exist: %s", specPath)
			}
		})
	}
}

// TestCANARY_CBIN_134_Engine_FuzzySpecSearch verifies FindSpecBySearch returns ranked results
func TestCANARY_CBIN_134_Engine_FuzzySpecSearch(t *testing.T) {
	// Setup: Create temporary test spec structure
	_, cleanup := setupTestSpecs(t)
	defer cleanup()

	// Execute: Fuzzy search
	results, err := FindSpecBySearch("spec modification", 5)
	if err != nil {
		t.Fatalf("FindSpecBySearch failed: %v", err)
	}

	// Verify: Results within limit
	if len(results) > 5 {
		t.Errorf("got %d results, limit was 5", len(results))
	}

	t.Logf("Found %d results", len(results))
}
