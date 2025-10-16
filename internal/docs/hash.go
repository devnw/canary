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

// CANARY: REQ=CBIN-136; FEATURE="DocHashCalculation"; ASPECT=Engine; STATUS=TESTED; TEST=TestCANARY_CBIN_136_Engine_HashCalculation; UPDATED=2025-10-16

package docs

import (
	"crypto/sha256"
	"encoding/hex"
	"os"
	"strings"
)

// CalculateHash computes SHA256 hash of documentation file with line ending normalization.
// Returns first 16 characters (64 bits) for abbreviated display in CANARY tokens.
//
// Line endings are normalized to LF (\n) before hashing to ensure cross-platform consistency
// between Windows (CRLF) and Unix/Mac (LF) systems.
//
// Example:
//
//	hash, err := docs.CalculateHash("docs/user/auth.md")
//	// Returns: "8f434346648f6b96" (first 16 chars of SHA256)
func CalculateHash(filePath string) (string, error) {
	// Read file content
	content, err := os.ReadFile(filePath)
	if err != nil {
		return "", err
	}

	// Normalize line endings: Convert CRLF to LF
	// This ensures deterministic hashing across platforms
	normalized := strings.ReplaceAll(string(content), "\r\n", "\n")

	// Calculate SHA256
	hash := sha256.Sum256([]byte(normalized))

	// Encode to hex string
	fullHash := hex.EncodeToString(hash[:])

	// Return abbreviated hash (first 16 characters)
	// 16 hex chars = 64 bits, sufficient collision resistance for doc tracking
	return fullHash[:16], nil
}

// CalculateFullHash returns full 64-character SHA256 hash.
// Use this for database storage if full hash precision is needed.
func CalculateFullHash(filePath string) (string, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return "", err
	}

	normalized := strings.ReplaceAll(string(content), "\r\n", "\n")
	hash := sha256.Sum256([]byte(normalized))
	return hex.EncodeToString(hash[:]), nil
}
