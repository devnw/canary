// Copyright (c) 2024 by Developer Network.
//
// For more details, see the LICENSE file in the root directory of this
// source code repository or contact Developer Network at info@devnw.com.


// CANARY: REQ=CBIN-139; FEATURE="AspectValidator"; ASPECT=Engine; STATUS=IMPL; UPDATED=2025-10-16
package reqid

import (
	"fmt"
	"strings"

	"go.spyder.org/canary/internal/matcher"
)

// validAspects defines all valid aspect values with their canonical casing
var validAspects = []string{
	"API",
	"CLI",
	"Engine",
	"Storage",
	"Security",
	"Docs",
	"Wire",
	"Planner",
	"Decode",
	"Encode",
	"RoundTrip",
	"Bench",
	"FrontEnd",
	"Dist",
}

// ValidateAspect checks if the given aspect is valid
// Accepts: exact canonical form, all-lowercase, or all-uppercase
// Rejects: partial capitalizations (e.g., "Frontend" when it should be "FrontEnd")
func ValidateAspect(aspect string) error {
	if aspect == "" {
		return fmt.Errorf("aspect cannot be empty")
	}

	// Check for valid match
	for _, valid := range validAspects {
		// Exact match (canonical form)
		if aspect == valid {
			return nil
		}

		// All lowercase match
		if aspect == strings.ToLower(valid) {
			return nil
		}

		// All uppercase match
		if aspect == strings.ToUpper(valid) {
			return nil
		}
	}

	// Not valid, provide suggestion
	suggestion := SuggestAspect(aspect)
	if suggestion != "" {
		return fmt.Errorf("invalid aspect %q, did you mean: %s", aspect, suggestion)
	}

	return fmt.Errorf("invalid aspect %q", aspect)
}

// SuggestAspect returns fuzzy match suggestions for invalid aspects
func SuggestAspect(typo string) string {
	if typo == "" {
		return ""
	}

	bestScore := 0
	bestMatch := ""

	// Find best fuzzy match
	for _, valid := range validAspects {
		score := matcher.ScoreMatch(typo, valid)
		if score > bestScore {
			bestScore = score
			bestMatch = valid
		}
	}

	// Only suggest if score is reasonable (60+ is a decent match)
	if bestScore >= 60 {
		return bestMatch
	}

	return ""
}

// NormalizeAspect normalizes aspect casing to the canonical form
func NormalizeAspect(input string) string {
	if input == "" {
		return input
	}

	// Find case-insensitive match and return canonical form
	for _, valid := range validAspects {
		if strings.EqualFold(input, valid) {
			return valid
		}
	}

	// Not found, return as-is
	return input
}
