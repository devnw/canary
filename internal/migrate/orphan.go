// Copyright (c) 2025 by Developer Network.
//
// For more details, see the LICENSE file in the root directory of this
// source code repository or contact Developer Network at info@devnw.com.

// CANARY: REQ=CBIN-145; FEATURE="OrphanDetection"; ASPECT=Engine; STATUS=IMPL; UPDATED=2025-10-17

package migrate

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"go.devnw.com/canary/internal/storage"
)

// DetectOrphans finds all requirements with tokens but no specification
func DetectOrphans(db *storage.DB, rootDir string, excludePaths []string) ([]*OrphanedRequirement, error) {
	// Get all tokens from database
	tokens, err := db.ListTokens(map[string]string{}, "", "req_id ASC", 0)
	if err != nil {
		return nil, fmt.Errorf("failed to list tokens: %w", err)
	}

	// Filter tokens by path exclusions
	filteredTokens := []*storage.Token{}
	for _, token := range tokens {
		if !shouldExcludePath(token.FilePath, excludePaths) {
			filteredTokens = append(filteredTokens, token)
		}
	}

	// Group tokens by requirement ID
	grouped := groupByRequirement(filteredTokens)

	// Find orphans (requirements without specs)
	orphans := []*OrphanedRequirement{}
	for reqID, reqTokens := range grouped {
		if !specExists(rootDir, reqID) {
			orphan := createOrphanRequirement(reqID, reqTokens)
			orphans = append(orphans, orphan)
		}
	}

	return orphans, nil
}

// DryRun simulates migration without creating files
func DryRun(db *storage.DB, rootDir string, excludePaths []string) (*OrphanPlan, error) {
	orphans, err := DetectOrphans(db, rootDir, excludePaths)
	if err != nil {
		return nil, err
	}

	plan := &OrphanPlan{
		Orphans:      orphans,
		TotalOrphans: len(orphans),
		Excluded:     excludePaths,
	}

	return plan, nil
}

// CalculateConfidence determines confidence level based on orphan characteristics
func CalculateConfidence(orphan *OrphanedRequirement) string {
	score := 0

	// Feature count scoring
	if orphan.FeatureCount >= 5 {
		score += 3
	} else if orphan.FeatureCount >= 3 {
		score += 2
	} else if orphan.FeatureCount >= 2 {
		score += 1
	}

	// Test coverage scoring
	hasTests := false
	hasBenchmarks := false
	for _, token := range orphan.Features {
		if token.Test != "" {
			hasTests = true
		}
		if token.Bench != "" {
			hasBenchmarks = true
		}
	}

	if hasTests {
		score += 2
	}
	if hasBenchmarks {
		score += 1
	}

	// Status progression scoring
	hasImpl := false
	hasTested := false
	hasBenched := false
	for _, token := range orphan.Features {
		if token.Status == "IMPL" {
			hasImpl = true
		}
		if token.Status == "TESTED" {
			hasTested = true
		}
		if token.Status == "BENCHED" {
			hasBenched = true
		}
	}

	if hasBenched {
		score += 2
	} else if hasTested {
		score += 1
	} else if hasImpl {
		score += 0
	}

	// Convert score to confidence level
	if score >= 5 {
		return ConfidenceHigh
	} else if score >= 2 {
		return ConfidenceMedium
	}
	return ConfidenceLow
}

// shouldExcludePath checks if a file path should be excluded
func shouldExcludePath(filePath string, excludePaths []string) bool {
	for _, exclude := range excludePaths {
		// Remove leading/trailing slashes for consistent matching
		exclude = strings.Trim(exclude, "/")
		// Check if path contains the exclusion pattern
		if strings.Contains(filePath, exclude+"/") || strings.HasPrefix(filePath, exclude+"/") {
			return true
		}
	}
	return false
}

// groupByRequirement groups tokens by requirement ID
func groupByRequirement(tokens []*storage.Token) map[string][]*storage.Token {
	grouped := make(map[string][]*storage.Token)
	for _, token := range tokens {
		grouped[token.ReqID] = append(grouped[token.ReqID], token)
	}
	return grouped
}

// specExists checks if a specification exists for the given requirement ID
func specExists(rootDir string, reqID string) bool {
	specsDir := filepath.Join(rootDir, ".canary", "specs")

	// Read specs directory
	entries, err := os.ReadDir(specsDir)
	if err != nil {
		return false
	}

	// Look for directory matching reqID pattern (e.g., CBIN-100-*)
	prefix := reqID + "-"
	for _, entry := range entries {
		if entry.IsDir() && strings.HasPrefix(entry.Name(), prefix) {
			// Check if spec.md exists in this directory
			specPath := filepath.Join(specsDir, entry.Name(), "spec.md")
			if _, err := os.Stat(specPath); err == nil {
				return true
			}
		}
	}

	return false
}

// createOrphanRequirement creates an OrphanedRequirement from tokens
func createOrphanRequirement(reqID string, tokens []*storage.Token) *OrphanedRequirement {
	orphan := &OrphanedRequirement{
		ReqID:        reqID,
		Features:     tokens,
		FeatureCount: len(tokens),
	}

	// Calculate confidence
	orphan.Confidence = CalculateConfidence(orphan)

	return orphan
}
