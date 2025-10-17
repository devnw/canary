// Copyright (c) 2024 by Developer Network.
//
// For more details, see the LICENSE file in the root directory of this
// source code repository or contact Developer Network at info@devnw.com.


// CANARY: REQ=CBIN-139; FEATURE="AspectIDGenerator"; ASPECT=Engine; STATUS=IMPL; UPDATED=2025-10-16
package reqid

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// GenerateNextID generates the next requirement ID for a given aspect
func GenerateNextID(key, aspect string) (string, error) {
	// Validate aspect
	if err := ValidateAspect(aspect); err != nil {
		return "", err
	}

	// Normalize aspect to canonical casing
	aspect = NormalizeAspect(aspect)

	// Find .canary/specs directory
	specsDir := filepath.Join(".canary", "specs")
	if _, err := os.Stat(specsDir); os.IsNotExist(err) {
		return "", fmt.Errorf("specs directory not found: %s", specsDir)
	}

	// Read all spec directories
	entries, err := os.ReadDir(specsDir)
	if err != nil {
		return "", fmt.Errorf("failed to read specs directory: %w", err)
	}

	// Find maximum ID for this aspect
	maxID := 0
	prefix := fmt.Sprintf("%s-%s-", key, aspect)

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		name := entry.Name()

		// Check if this directory matches our aspect (case-insensitive prefix match)
		if !strings.HasPrefix(strings.ToLower(name), strings.ToLower(prefix)) {
			continue
		}

		// Extract ID from directory name: CBIN-CLI-001-feature-name
		// Remove prefix to get: 001-feature-name
		remainder := name[len(prefix):]

		// Split by '-' and take first part (the ID)
		parts := strings.SplitN(remainder, "-", 2)
		if len(parts) == 0 {
			continue
		}

		// Parse ID as integer
		idStr := parts[0]
		id, err := strconv.Atoi(idStr)
		if err != nil {
			// Skip if not a valid integer
			continue
		}

		if id > maxID {
			maxID = id
		}
	}

	// Generate next ID
	nextID := maxID + 1

	// Format as 3-digit zero-padded string
	return fmt.Sprintf("%s-%s-%03d", key, aspect, nextID), nil
}
