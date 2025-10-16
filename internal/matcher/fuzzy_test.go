// CANARY: REQ=CBIN-133; FEATURE="FuzzyMatcherTests"; ASPECT=Engine; STATUS=TESTED; TEST=TestCANARY_CBIN_133_Engine_Levenshtein; OWNER=canary; UPDATED=2025-10-16
package matcher_test

import (
	"os"
	"testing"

	"go.spyder.org/canary/internal/matcher"
)

func TestCANARY_CBIN_133_Engine_Levenshtein(t *testing.T) {
	tests := []struct {
		name     string
		s1       string
		s2       string
		expected int
	}{
		{"empty strings", "", "", 0},
		{"identical strings", "hello", "hello", 0},
		{"one char difference", "hello", "hallo", 1},
		{"kitten to sitting", "kitten", "sitting", 3},
		{"hyphen difference", "CBIN105", "CBIN-105", 1},
		{"case insensitive", "Hello", "hello", 0},
		{"completely different", "abc", "xyz", 3},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := matcher.CalculateLevenshtein(tc.s1, tc.s2)
			if result != tc.expected {
				t.Errorf("CalculateLevenshtein(%q, %q) = %d; want %d", tc.s1, tc.s2, result, tc.expected)
			}
		})
	}
}

func TestCANARY_CBIN_133_Engine_FuzzyScoring(t *testing.T) {
	tests := []struct {
		name      string
		query     string
		candidate string
		minScore  int // Score should be >= this
		maxScore  int // Score should be <= this
	}{
		{"exact match", "hello", "hello", 100, 100},
		{"substring match", "auth", "UserAuthentication", 80, 100},
		{"fuzzy match", "user auth", "UserAuthentication", 60, 90},
		{"CBIN ID match", "CBIN105", "CBIN-105", 90, 100},
		{"abbreviation match", "ua", "UserAuthentication", 70, 80},
		{"poor match", "xyz", "UserAuthentication", 0, 40},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			score := matcher.ScoreMatch(tc.query, tc.candidate)
			if score < tc.minScore || score > tc.maxScore {
				t.Errorf("ScoreMatch(%q, %q) = %d; want between %d and %d",
					tc.query, tc.candidate, score, tc.minScore, tc.maxScore)
			}
		})
	}
}

func TestCANARY_CBIN_133_Engine_FindBestMatches(t *testing.T) {
	// Setup: Create temp directory with test spec directories
	tmpDir := t.TempDir()
	testSpecs := []string{
		"CBIN-105-UserAuthentication",
		"CBIN-110-OAuthIntegration",
		"CBIN-112-DataValidation",
		"CBIN-115-EmailNotifications",
	}

	for _, spec := range testSpecs {
		specDir := tmpDir + "/" + spec
		if err := mkdir(specDir); err != nil {
			t.Fatalf("failed to create spec dir: %v", err)
		}
	}

	tests := []struct {
		name          string
		query         string
		limit         int
		expectedFirst string // Expected first match ReqID
		minMatches    int
	}{
		{"exact ID match", "CBIN-105", 5, "CBIN-105", 1},
		{"fuzzy feature match", "user auth", 5, "CBIN-105", 1},
		{"partial match", "auth", 5, "CBIN-110", 2}, // Matches OAuthIntegration and UserAuthentication
		{"list all", "", 5, "", 4},                  // Empty query returns all
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			matches, err := matcher.FindBestMatches(tc.query, tmpDir, tc.limit)
			if err != nil {
				t.Fatalf("FindBestMatches failed: %v", err)
			}

			if len(matches) < tc.minMatches {
				t.Errorf("Expected at least %d matches, got %d", tc.minMatches, len(matches))
			}

			if tc.expectedFirst != "" && len(matches) > 0 {
				if matches[0].ReqID != tc.expectedFirst {
					t.Errorf("Expected first match %s, got %s", tc.expectedFirst, matches[0].ReqID)
				}
			}
		})
	}
}

// Helper function for test setup
func mkdir(path string) error {
	return os.MkdirAll(path, 0755)
}
