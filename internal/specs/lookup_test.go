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

package specs

import (
	"os"
	"path/filepath"
	"testing"
)

// CANARY: REQ=CBIN-134; FEATURE="ExactIDLookup"; ASPECT=Engine; STATUS=STUB; TEST=TestCANARY_CBIN_134_Engine_ExactIDLookup; UPDATED=2025-10-16

func TestCANARY_CBIN_134_Engine_ExactIDLookup(t *testing.T) {
	// Test FindSpecByID with valid and invalid IDs
	// Expected to FAIL initially (function doesn't exist)

	// Setup: Change to project root if needed
	if _, err := os.Stat(".canary"); err != nil {
		t.Skip("Skipping test: not in project root directory")
	}

	tests := []struct {
		name    string
		reqID   string
		wantErr bool
	}{
		{
			name:    "valid CBIN-134",
			reqID:   "CBIN-134",
			wantErr: false,
		},
		{
			name:    "valid CBIN-135",
			reqID:   "CBIN-135",
			wantErr: false,
		},
		{
			name:    "invalid CBIN-999",
			reqID:   "CBIN-999",
			wantErr: true,
		},
		{
			name:    "empty string",
			reqID:   "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// This will fail because FindSpecByID doesn't exist yet
			specPath, err := FindSpecByID(tt.reqID)

			if (err != nil) != tt.wantErr {
				t.Errorf("FindSpecByID() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				// Verify the returned path exists
				if _, err := os.Stat(specPath); err != nil {
					t.Errorf("FindSpecByID() returned non-existent path: %s", specPath)
				}

				// Verify path matches expected pattern
				expectedPattern := filepath.Join(".canary", "specs", tt.reqID+"-*", "spec.md")
				matched, _ := filepath.Match(expectedPattern, specPath)
				if !matched {
					// More flexible check - just verify it contains the req ID
					if !stringContains(specPath, tt.reqID) {
						t.Errorf("FindSpecByID() path %s doesn't contain req ID %s", specPath, tt.reqID)
					}
				}
			}
		})
	}
}

func TestCANARY_CBIN_134_Engine_FuzzySpecSearch(t *testing.T) {
	// Test FindSpecBySearch returns ranked results
	// Expected to FAIL initially (function doesn't exist)

	if _, err := os.Stat(".canary"); err != nil {
		t.Skip("Skipping test: not in project root directory")
	}

	tests := []struct {
		name           string
		query          string
		limit          int
		wantMinMatches int
		wantErr        bool
	}{
		{
			name:           "search for 'specification'",
			query:          "specification",
			limit:          5,
			wantMinMatches: 1,
			wantErr:        false,
		},
		{
			name:           "search for 'priority list'",
			query:          "priority list",
			limit:          5,
			wantMinMatches: 1,
			wantErr:        false,
		},
		{
			name:           "search for non-existent",
			query:          "xyznonexistent123",
			limit:          5,
			wantMinMatches: 0,
			wantErr:        false, // No error, just empty results
		},
		{
			name:           "limit to 3 results",
			query:          "canary",
			limit:          3,
			wantMinMatches: 0,
			wantErr:        false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// This will fail because FindSpecBySearch doesn't exist yet
			matches, err := FindSpecBySearch(tt.query, tt.limit)

			if (err != nil) != tt.wantErr {
				t.Errorf("FindSpecBySearch() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if len(matches) < tt.wantMinMatches {
					t.Errorf("FindSpecBySearch() got %d matches, want at least %d", len(matches), tt.wantMinMatches)
				}

				// Verify results are limited
				if len(matches) > tt.limit {
					t.Errorf("FindSpecBySearch() got %d matches, limit was %d", len(matches), tt.limit)
				}

				// Verify matches are sorted by score (descending)
				for i := 1; i < len(matches); i++ {
					if matches[i-1].Score < matches[i].Score {
						t.Errorf("FindSpecBySearch() results not sorted by score: %d < %d", matches[i-1].Score, matches[i].Score)
					}
				}
			}
		})
	}
}

func TestCANARY_CBIN_134_Engine_DatabaseLookup(t *testing.T) {
	// Test FindSpecInDB with temporary database
	// Expected to FAIL initially (function doesn't exist)

	t.Skip("Database lookup test - implementation pending")

	// This test will be implemented once we have the function
	// It should:
	// 1. Create a temporary SQLite database
	// 2. Insert test fixture tokens
	// 3. Query for specs by REQ-ID
	// 4. Verify fallback to filesystem
}

// Helper function to check if string contains substring
func stringContains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && stringContainsHelper(s, substr))
}

func stringContainsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
