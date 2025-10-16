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

// CANARY: REQ=CBIN-136; FEATURE="DocStalenessDetection"; ASPECT=Engine; STATUS=TESTED; TEST=TestCANARY_CBIN_136_Engine_StalenessDetection; DOC=architecture:docs/architecture/adr-001-documentation-tracking.md; DOC_HASH=9c40f77ae6604be5; UPDATED=2025-10-16

package docs

import (
	"os"
	"strings"

	"go.spyder.org/canary/internal/storage"
)

// CheckStaleness compares documentation file hash to token DOC_HASH field.
// Returns one of: DOC_CURRENT, DOC_STALE, DOC_MISSING, or DOC_UNHASHED
//
// Status meanings:
//   - DOC_CURRENT: File hash matches token DOC_HASH (documentation is up-to-date)
//   - DOC_STALE: File hash differs from token DOC_HASH (documentation needs updating)
//   - DOC_MISSING: Documentation file does not exist at specified path
//   - DOC_UNHASHED: Token has no DOC_HASH field (hash tracking not enabled)
//
// Example:
//
//	status, err := docs.CheckStaleness(token)
//	if status == "DOC_STALE" {
//	    fmt.Printf("Documentation for %s is outdated\n", token.ReqID)
//	}
func CheckStaleness(token *storage.Token) (string, error) {
	// Case 1: No DOC_HASH field in token
	if token.DocHash == "" {
		return "DOC_UNHASHED", nil
	}

	// Case 2: Documentation file missing
	if _, err := os.Stat(token.DocPath); os.IsNotExist(err) {
		return "DOC_MISSING", nil
	}

	// Case 3: Calculate current hash and compare
	currentHash, err := CalculateHash(token.DocPath)
	if err != nil {
		return "", err
	}

	// Compare abbreviated hash (first 16 chars)
	tokenHash := token.DocHash
	if len(tokenHash) > 16 {
		tokenHash = tokenHash[:16]
	}

	if currentHash == tokenHash {
		return "DOC_CURRENT", nil
	}

	return "DOC_STALE", nil
}

// CheckMultipleDocumentation handles tokens with multiple DOC paths (comma-separated).
// Returns a map of doc path to status for each documentation file.
//
// Example:
//
//	// Token with: DOC=user:docs/user.md,api:docs/api.md
//	results, err := docs.CheckMultipleDocumentation(token)
//	// Returns: {"docs/user.md": "DOC_CURRENT", "docs/api.md": "DOC_STALE"}
func CheckMultipleDocumentation(token *storage.Token) (map[string]string, error) {
	// Parse comma-separated DOC paths
	docPaths := strings.Split(token.DocPath, ",")
	docHashes := strings.Split(token.DocHash, ",")

	results := make(map[string]string)

	for i, docPath := range docPaths {
		// Trim whitespace and type prefix (e.g., "user:docs/file.md" -> "docs/file.md")
		docPath = strings.TrimSpace(docPath)
		if strings.Contains(docPath, ":") {
			parts := strings.SplitN(docPath, ":", 2)
			if len(parts) == 2 {
				docPath = parts[1]
			}
		}

		// Create temporary token for single doc check
		singleDocToken := &storage.Token{
			DocPath: docPath,
			DocHash: "",
		}

		if i < len(docHashes) {
			singleDocToken.DocHash = strings.TrimSpace(docHashes[i])
		}

		status, err := CheckStaleness(singleDocToken)
		if err != nil {
			return nil, err
		}

		results[docPath] = status
	}

	return results, nil
}
