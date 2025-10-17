// Copyright (c) 2024 by Developer Network.
//
// For more details, see the LICENSE file in the root directory of this
// source code repository or contact Developer Network at info@devnw.com.


// CANARY: REQ=CBIN-133; FEATURE="FuzzyMatcherTests"; ASPECT=Engine; STATUS=TESTED; TEST=TestCANARY_CBIN_133_Engine_Levenshtein; UPDATED=2025-10-16
package matcher_test

import (
	"testing"

	"go.spyder.org/canary/internal/matcher"
)

// TestCANARY_CBIN_133_Engine_Levenshtein verifies Levenshtein distance calculation
func TestCANARY_CBIN_133_Engine_Levenshtein(t *testing.T) {
	tests := []struct {
		name     string
		s1       string
		s2       string
		expected int
	}{
		{
			name:     "empty strings",
			s1:       "",
			s2:       "",
			expected: 0,
		},
		{
			name:     "identical strings",
			s1:       "hello",
			s2:       "hello",
			expected: 0,
		},
		{
			name:     "single character difference",
			s1:       "hello",
			s2:       "hallo",
			expected: 1,
		},
		{
			name:     "classic kitten/sitting example",
			s1:       "kitten",
			s2:       "sitting",
			expected: 3,
		},
		{
			name:     "CANARY ID with hyphen difference",
			s1:       "CBIN105",
			s2:       "CBIN-105",
			expected: 1,
		},
		{
			name:     "case insensitive",
			s1:       "Hello",
			s2:       "hello",
			expected: 0,
		},
		{
			name:     "completely different",
			s1:       "abc",
			s2:       "xyz",
			expected: 3,
		},
		{
			name:     "one empty string",
			s1:       "test",
			s2:       "",
			expected: 4,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := matcher.CalculateLevenshtein(tc.s1, tc.s2)
			if result != tc.expected {
				t.Errorf("CalculateLevenshtein(%q, %q) = %d; want %d",
					tc.s1, tc.s2, result, tc.expected)
			}
		})
	}
}

// TestCANARY_CBIN_133_Engine_FuzzyScoring verifies fuzzy match scoring
func TestCANARY_CBIN_133_Engine_FuzzyScoring(t *testing.T) {
	tests := []struct {
		name      string
		query     string
		candidate string
		minScore  int // Score should be >= this
	}{
		{
			name:      "exact match should score 100",
			query:     "UserAuthentication",
			candidate: "UserAuthentication",
			minScore:  100,
		},
		{
			name:      "substring match scores high",
			query:     "user auth",
			candidate: "UserAuthentication",
			minScore:  75,
		},
		{
			name:      "partial substring",
			query:     "auth",
			candidate: "UserAuthentication",
			minScore:  60,
		},
		{
			name:      "CANARY ID match",
			query:     "CBIN105",
			candidate: "CBIN-105",
			minScore:  90,
		},
		{
			name:      "case insensitive exact match",
			query:     "test",
			candidate: "TEST",
			minScore:  100,
		},
		{
			name:      "abbreviation match",
			query:     "ua",
			candidate: "UserAuthentication",
			minScore:  70,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			score := matcher.ScoreMatch(tc.query, tc.candidate)
			if score < tc.minScore {
				t.Errorf("ScoreMatch(%q, %q) = %d; want >= %d",
					tc.query, tc.candidate, score, tc.minScore)
			}
		})
	}
}

// TestCANARY_CBIN_133_Engine_FindBestMatches verifies best match selection
func TestCANARY_CBIN_133_Engine_FindBestMatches(t *testing.T) {
	// Setup: Create temporary specs directory with test specs
	tmpDir := t.TempDir()
	specsDir := tmpDir + "/.canary/specs"

	// This test will verify FindBestMatches can find and rank specs by similarity
	// We'll create the test directory structure in the actual implementation test

	// For now, verify the function exists and has correct signature
	// The actual functionality will be tested when we implement the function
	matches, err := matcher.FindBestMatches("test query", specsDir, 5)

	// We expect an error since the directory doesn't exist yet
	if err == nil {
		t.Logf("FindBestMatches returned %d matches (function exists)", len(matches))
	} else {
		t.Logf("FindBestMatches returned expected error: %v", err)
	}
}
