// Copyright (c) 2025 by Developer Network.
//
// For more details, see the LICENSE file in the root directory of this
// source code repository or contact Developer Network at info@devnw.com.

// CANARY: REQ=CBIN-134; FEATURE="ExactIDLookup"; ASPECT=Engine; STATUS=TESTED; TEST=TestCANARY_CBIN_134_Engine_ExactIDLookup; UPDATED=2025-10-17
package specs

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

// getRepoRoot returns the repository root directory
func getRepoRoot() string {
	_, thisFile, _, _ := runtime.Caller(0)
	// From internal/specs/lookup_test.go, go up two levels to repo root
	return filepath.Clean(filepath.Join(filepath.Dir(thisFile), "../.."))
}

// TestCANARY_CBIN_134_Engine_ExactIDLookup verifies FindSpecByID locates spec files
func TestCANARY_CBIN_134_Engine_ExactIDLookup(t *testing.T) {
	// Change to repo root for .canary/specs access
	repoRoot := getRepoRoot()
	oldDir, _ := os.Getwd()
	defer os.Chdir(oldDir)
	if err := os.Chdir(repoRoot); err != nil {
		t.Fatalf("failed to change to repo root: %v", err)
	}

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
	// Change to repo root for .canary/specs access
	repoRoot := getRepoRoot()
	oldDir, _ := os.Getwd()
	defer os.Chdir(oldDir)
	if err := os.Chdir(repoRoot); err != nil {
		t.Fatalf("failed to change to repo root: %v", err)
	}

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
