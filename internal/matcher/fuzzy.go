// Copyright (c) 2024 by Developer Network.
//
// For more details, see the LICENSE file in the root directory of this
// source code repository or contact Developer Network at info@devnw.com.


// CANARY: REQ=CBIN-133; FEATURE="FuzzyMatcher"; ASPECT=Engine; STATUS=TESTED; TEST=TestCANARY_CBIN_133_Engine_Levenshtein; OWNER=canary; UPDATED=2025-10-16
package matcher

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
	"unicode"
)

// CalculateLevenshtein computes edit distance between two strings
func CalculateLevenshtein(s1, s2 string) int {
	s1 = strings.ToLower(s1)
	s2 = strings.ToLower(s2)

	if s1 == s2 {
		return 0
	}

	// Create matrix
	d := make([][]int, len(s1)+1)
	for i := range d {
		d[i] = make([]int, len(s2)+1)
		d[i][0] = i
	}
	for j := range d[0] {
		d[0][j] = j
	}

	// Fill matrix
	for i := 1; i <= len(s1); i++ {
		for j := 1; j <= len(s2); j++ {
			cost := 0
			if s1[i-1] != s2[j-1] {
				cost = 1
			}

			d[i][j] = min(
				d[i-1][j]+1,      // deletion
				d[i][j-1]+1,      // insertion
				d[i-1][j-1]+cost, // substitution
			)
		}
	}

	return d[len(s1)][len(s2)]
}

// ScoreMatch calculates similarity score (0-100) between query and candidate
func ScoreMatch(query, candidate string) int {
	// Save original candidate for abbreviation matching (needs capitals)
	origCandidate := candidate

	query = strings.ToLower(query)
	candidate = strings.ToLower(candidate)

	// Exact match
	if query == candidate {
		return 100
	}

	// Try removing hyphens for ID matching (e.g., "cbin105" vs "cbin-105")
	queryNoHyphen := strings.ReplaceAll(query, "-", "")
	candidateNoHyphen := strings.ReplaceAll(candidate, "-", "")
	if queryNoHyphen == candidateNoHyphen {
		return 95 // Very close match, just formatting difference
	}

	// Substring match gets high score
	if strings.Contains(candidate, query) {
		ratio := float64(len(query)) / float64(len(candidate))
		return int(80 + (ratio * 20)) // 80-100 range
	}

	// Multi-word fuzzy match (e.g., "user auth" vs "UserAuthentication")
	if strings.Contains(query, " ") {
		words := strings.Fields(query)
		allMatch := true
		for _, word := range words {
			if !strings.Contains(candidate, word) {
				allMatch = false
				break
			}
		}
		if allMatch {
			// All words found as substrings
			ratio := float64(len(query)-len(words)+1) / float64(len(candidate))
			return int(70 + (ratio * 20)) // 70-90 range for multi-word matches
		}
	}

	// Abbreviation match (e.g., "ua" matches "UserAuthentication")
	// Pass original candidate (not lowercased) to preserve capitals
	abbrevScore := abbreviationScore(query, origCandidate)
	if abbrevScore >= 75 {
		return abbrevScore
	}

	// Levenshtein distance
	distance := CalculateLevenshtein(query, candidate)
	maxLen := max(len(query), len(candidate))

	if maxLen == 0 {
		return 0
	}

	// Convert distance to similarity percentage
	similarity := float64(maxLen-distance) / float64(maxLen)
	score := int(similarity * 100)

	// Use abbreviation score if it's better
	if abbrevScore > score {
		return abbrevScore
	}

	if score < 0 {
		return 0
	}
	return score
}

// abbreviationScore calculates score based on abbreviation matching
func abbreviationScore(query, candidate string) int {
	// Extract abbreviation from candidate (capital letters)
	var abbrev strings.Builder
	for i, ch := range candidate {
		if i == 0 || unicode.IsUpper(ch) {
			abbrev.WriteRune(unicode.ToLower(ch))
		}
	}

	abbrevStr := abbrev.String()

	// Exact abbreviation match
	if query == abbrevStr {
		return 75
	}

	// Query is substring of abbreviation
	if strings.Contains(abbrevStr, query) {
		ratio := float64(len(query)) / float64(len(abbrevStr))
		return int(70 + (ratio * 20)) // 70-90 range
	}

	// Levenshtein distance for abbreviation
	if len(abbrevStr) > 0 {
		distance := CalculateLevenshtein(query, abbrevStr)
		maxLen := max(len(query), len(abbrevStr))
		if distance <= 1 && maxLen > 0 {
			similarity := float64(maxLen-distance) / float64(maxLen)
			score := int(similarity * 80) // Max 80 for close abbreviation match
			if score >= 70 {
				return score
			}
		}
	}

	return 0
}

// matchesAbbreviation checks if query matches first letters of words in candidate
func matchesAbbreviation(query, candidate string) bool {
	var abbrev strings.Builder
	for i, ch := range candidate {
		if i == 0 || unicode.IsUpper(ch) {
			abbrev.WriteRune(unicode.ToLower(ch))
		}
	}
	return strings.Contains(abbrev.String(), query)
}

func min(a, b, c int) int {
	if a < b {
		if a < c {
			return a
		}
		return c
	}
	if b < c {
		return b
	}
	return c
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// Match represents a fuzzy match result
type Match struct {
	ReqID       string
	FeatureName string
	Score       int
	SpecPath    string
}

// FindBestMatches returns top N matches for query
func FindBestMatches(query string, specsDir string, limit int) ([]Match, error) {
	entries, err := os.ReadDir(specsDir)
	if err != nil {
		return nil, err
	}

	var matches []Match
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		// Parse directory name: CBIN-XXX-feature-name
		// Need to extract CBIN-XXX as reqID, rest as feature name
		parts := strings.Split(entry.Name(), "-")
		if len(parts) < 3 {
			continue
		}

		// ReqID is first two parts joined: "CBIN" + "-" + "XXX"
		reqID := parts[0] + "-" + parts[1]
		// Feature name is everything after that
		featureName := strings.Join(parts[2:], "-")

		// Empty query returns all matches
		if query == "" {
			matches = append(matches, Match{
				ReqID:       reqID,
				FeatureName: featureName,
				Score:       100, // All items have equal score
				SpecPath:    filepath.Join(specsDir, entry.Name()),
			})
			continue
		}

		// Score against both ID and feature name
		idScore := ScoreMatch(query, reqID)
		nameScore := ScoreMatch(query, featureName)
		score := max(idScore, nameScore)

		if score >= 60 { // Minimum threshold
			matches = append(matches, Match{
				ReqID:       reqID,
				FeatureName: featureName,
				Score:       score,
				SpecPath:    filepath.Join(specsDir, entry.Name()),
			})
		}
	}

	// Sort by score (highest first)
	sort.Slice(matches, func(i, j int) bool {
		return matches[i].Score > matches[j].Score
	})

	// Return top N
	if limit > 0 && len(matches) > limit {
		return matches[:limit], nil
	}
	return matches, nil
}
