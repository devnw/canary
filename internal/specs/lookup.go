// Copyright (c) 2024 by Developer Network.
//
// For more details, see the LICENSE file in the root directory of this
// source code repository or contact Developer Network at info@devnw.com.

// CANARY: REQ=CBIN-134; FEATURE="ExactIDLookup"; ASPECT=Engine; STATUS=IMPL; UPDATED=2025-10-16
package specs

import (
	"fmt"
	"os"
	"path/filepath"

	"go.spyder.org/canary/internal/matcher"
	"go.spyder.org/canary/internal/storage"
)

// FindSpecByID locates spec.md file by exact requirement ID
// Uses glob pattern: .canary/specs/CBIN-XXX-*/spec.md
func FindSpecByID(reqID string) (string, error) {
	if reqID == "" {
		return "", fmt.Errorf("requirement ID cannot be empty")
	}

	specsDir := ".canary/specs"
	pattern := filepath.Join(specsDir, reqID+"-*", "spec.md")

	matches, err := filepath.Glob(pattern)
	if err != nil {
		return "", fmt.Errorf("glob pattern error: %w", err)
	}

	if len(matches) == 0 {
		return "", fmt.Errorf("spec not found for %s", reqID)
	}

	if len(matches) > 1 {
		return "", fmt.Errorf("multiple specs found for %s (ambiguous)", reqID)
	}

	return matches[0], nil
}

// FindSpecBySearch performs fuzzy search across spec directories
// Reuses CBIN-133 fuzzy matcher for scoring and ranking
func FindSpecBySearch(query string, limit int) ([]matcher.Match, error) {
	specsDir := ".canary/specs"

	// Check if specs directory exists
	if _, err := os.Stat(specsDir); err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("specs directory not found: %s", specsDir)
		}
		return nil, fmt.Errorf("failed to access specs directory: %w", err)
	}

	// Use existing fuzzy matcher from CBIN-133
	matches, err := matcher.FindBestMatches(query, specsDir, limit)
	if err != nil {
		return nil, fmt.Errorf("fuzzy search failed: %w", err)
	}

	return matches, nil
}

// FindSpecInDB queries database for fast spec lookup (optional fallback)
// If database is unavailable, returns error and caller should use FindSpecByID
func FindSpecInDB(db *storage.DB, reqID string) (string, error) {
	if db == nil {
		return "", fmt.Errorf("database not available")
	}

	if reqID == "" {
		return "", fmt.Errorf("requirement ID cannot be empty")
	}

	// Query database for tokens matching the requirement ID
	// Note: This depends on CBIN-123 TokenStorage implementation
	tokens, err := db.GetTokensByReqID(reqID)
	if err != nil {
		return "", fmt.Errorf("database query failed: %w", err)
	}

	if len(tokens) == 0 {
		return "", fmt.Errorf("spec not found in database: %s", reqID)
	}

	// Try to find spec.md file based on spec directory pattern
	specPattern := fmt.Sprintf(".canary/specs/%s-*/spec.md", reqID)
	matches, err := filepath.Glob(specPattern)
	if err != nil {
		return "", fmt.Errorf("glob pattern error: %w", err)
	}

	if len(matches) == 0 {
		return "", fmt.Errorf("spec file not found for %s (database entry exists but no spec.md)", reqID)
	}

	return matches[0], nil
}
