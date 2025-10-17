// Copyright (c) 2024 by Developer Network.
//
// For more details, see the LICENSE file in the root directory of this
// source code repository or contact Developer Network at info@devnw.com.

// CANARY: REQ=CBIN-136; FEATURE="DocIntegrationTests"; ASPECT=CLI; STATUS=TESTED; TEST=TestCANARY_CBIN_136_CLI_DocWorkflow; UPDATED=2025-10-16

package main

import (
	"os"
	"path/filepath"
	"testing"

	"go.spyder.org/canary/internal/docs"
	"go.spyder.org/canary/internal/storage"
)

// TestCANARY_CBIN_136_CLI_DocWorkflow verifies end-to-end documentation workflow
func TestCANARY_CBIN_136_CLI_DocWorkflow(t *testing.T) {
	// Setup: Create temporary directory
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")
	docPath := filepath.Join(tmpDir, "test-doc.md")

	// Create and migrate database
	if err := storage.MigrateDB(dbPath, storage.MigrateAll); err != nil {
		t.Fatalf("failed to migrate database: %v", err)
	}

	db, err := storage.Open(dbPath)
	if err != nil {
		t.Fatalf("failed to open database: %v", err)
	}
	defer db.Close()

	// Step 1: Create documentation file
	docContent := []byte("# Test Documentation\n\nThis is test content.")
	if err := os.WriteFile(docPath, docContent, 0644); err != nil {
		t.Fatalf("failed to write doc file: %v", err)
	}

	// Step 2: Calculate initial hash
	initialHash, err := docs.CalculateHash(docPath)
	if err != nil {
		t.Fatalf("failed to calculate hash: %v", err)
	}

	// Step 3: Create token with documentation reference
	token := &storage.Token{
		ReqID:      "CBIN-TEST",
		Feature:    "TestFeature",
		Aspect:     "Engine",
		Status:     "IMPL",
		FilePath:   "test.go",
		LineNumber: 10,
		UpdatedAt:  "2025-10-16",
		DocPath:    docPath,
		DocHash:    initialHash,
		DocType:    "user",
		DocStatus:  "DOC_CURRENT",
	}

	if err := db.UpsertToken(token); err != nil {
		t.Fatalf("failed to upsert token: %v", err)
	}

	// Step 4: Verify documentation is current
	status, err := docs.CheckStaleness(token)
	if err != nil {
		t.Fatalf("CheckStaleness failed: %v", err)
	}
	if status != "DOC_CURRENT" {
		t.Errorf("expected DOC_CURRENT, got %s", status)
	}

	// Step 5: Modify documentation
	modifiedContent := []byte("# Test Documentation\n\nThis is MODIFIED content.")
	if err := os.WriteFile(docPath, modifiedContent, 0644); err != nil {
		t.Fatalf("failed to modify doc file: %v", err)
	}

	// Step 6: Verify documentation is now stale
	status, err = docs.CheckStaleness(token)
	if err != nil {
		t.Fatalf("CheckStaleness failed: %v", err)
	}
	if status != "DOC_STALE" {
		t.Errorf("expected DOC_STALE after modification, got %s", status)
	}

	// Step 7: Update hash
	newHash, err := docs.CalculateHash(docPath)
	if err != nil {
		t.Fatalf("failed to calculate new hash: %v", err)
	}

	token.DocHash = newHash
	token.DocStatus = "DOC_CURRENT"
	if err := db.UpsertToken(token); err != nil {
		t.Fatalf("failed to update token: %v", err)
	}

	// Step 8: Verify documentation is current again
	status, err = docs.CheckStaleness(token)
	if err != nil {
		t.Fatalf("CheckStaleness failed: %v", err)
	}
	if status != "DOC_CURRENT" {
		t.Errorf("expected DOC_CURRENT after hash update, got %s", status)
	}

	// Verify hash changed
	if initialHash == newHash {
		t.Errorf("hash should have changed after modification")
	}
}

// TestCANARY_CBIN_136_CLI_DocMissingFile verifies handling of missing documentation
func TestCANARY_CBIN_136_CLI_DocMissingFile(t *testing.T) {
	tmpDir := t.TempDir()
	nonexistentPath := filepath.Join(tmpDir, "nonexistent.md")

	token := &storage.Token{
		ReqID:   "CBIN-TEST",
		Feature: "TestFeature",
		DocPath: nonexistentPath,
		DocHash: "abc123",
	}

	status, err := docs.CheckStaleness(token)
	if err != nil {
		t.Fatalf("expected no error for missing file, got: %v", err)
	}

	if status != "DOC_MISSING" {
		t.Errorf("expected DOC_MISSING, got %s", status)
	}
}

// TestCANARY_CBIN_136_CLI_DocUnhashed verifies handling of unhashed documentation
func TestCANARY_CBIN_136_CLI_DocUnhashed(t *testing.T) {
	tmpDir := t.TempDir()
	docPath := filepath.Join(tmpDir, "unhashed.md")

	// Create file but no hash
	if err := os.WriteFile(docPath, []byte("content"), 0644); err != nil {
		t.Fatalf("failed to write file: %v", err)
	}

	token := &storage.Token{
		ReqID:   "CBIN-TEST",
		Feature: "TestFeature",
		DocPath: docPath,
		DocHash: "", // No hash
	}

	status, err := docs.CheckStaleness(token)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if status != "DOC_UNHASHED" {
		t.Errorf("expected DOC_UNHASHED, got %s", status)
	}
}

// TestCANARY_CBIN_136_CLI_MultipleDocuments verifies handling of multiple docs
func TestCANARY_CBIN_136_CLI_MultipleDocuments(t *testing.T) {
	tmpDir := t.TempDir()

	// Create two documentation files
	doc1Path := filepath.Join(tmpDir, "user.md")
	doc2Path := filepath.Join(tmpDir, "api.md")

	content1 := []byte("# User Documentation")
	content2 := []byte("# API Documentation")

	if err := os.WriteFile(doc1Path, content1, 0644); err != nil {
		t.Fatalf("failed to write doc1: %v", err)
	}
	if err := os.WriteFile(doc2Path, content2, 0644); err != nil {
		t.Fatalf("failed to write doc2: %v", err)
	}

	// Calculate hashes
	hash1, _ := docs.CalculateHash(doc1Path)
	hash2, _ := docs.CalculateHash(doc2Path)

	// Create token with multiple docs
	token := &storage.Token{
		ReqID:   "CBIN-TEST",
		Feature: "TestFeature",
		DocPath: "user:" + doc1Path + ",api:" + doc2Path,
		DocHash: hash1 + "," + hash2,
	}

	// Check multiple documentation
	results, err := docs.CheckMultipleDocumentation(token)
	if err != nil {
		t.Fatalf("CheckMultipleDocumentation failed: %v", err)
	}

	if len(results) != 2 {
		t.Errorf("expected 2 results, got %d", len(results))
	}

	if results[doc1Path] != "DOC_CURRENT" {
		t.Errorf("expected doc1 to be DOC_CURRENT, got %s", results[doc1Path])
	}

	if results[doc2Path] != "DOC_CURRENT" {
		t.Errorf("expected doc2 to be DOC_CURRENT, got %s", results[doc2Path])
	}
}

// TestCANARY_CBIN_136_CLI_BatchUpdate verifies batch update operations
func TestCANARY_CBIN_136_CLI_BatchUpdate(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	// Create and migrate database
	if err := storage.MigrateDB(dbPath, storage.MigrateAll); err != nil {
		t.Fatalf("failed to migrate database: %v", err)
	}

	db, err := storage.Open(dbPath)
	if err != nil {
		t.Fatalf("failed to open database: %v", err)
	}
	defer db.Close()

	// Create multiple documentation files
	doc1Path := filepath.Join(tmpDir, "doc1.md")
	doc2Path := filepath.Join(tmpDir, "doc2.md")
	doc3Path := filepath.Join(tmpDir, "doc3.md")

	os.WriteFile(doc1Path, []byte("# Doc 1 - Original"), 0644)
	os.WriteFile(doc2Path, []byte("# Doc 2 - Original"), 0644)
	os.WriteFile(doc3Path, []byte("# Doc 3 - Original"), 0644)

	// Calculate initial hashes
	hash1, _ := docs.CalculateHash(doc1Path)
	hash2, _ := docs.CalculateHash(doc2Path)
	hash3, _ := docs.CalculateHash(doc3Path)

	// Create tokens
	token1 := &storage.Token{
		ReqID:     "CBIN-TEST-1",
		Feature:   "Feature1",
		Aspect:    "Engine",
		Status:    "IMPL",
		FilePath:  "test1.go",
		UpdatedAt: "2025-10-16",
		DocPath:   doc1Path,
		DocHash:   hash1,
		DocStatus: "DOC_CURRENT",
	}

	token2 := &storage.Token{
		ReqID:     "CBIN-TEST-2",
		Feature:   "Feature2",
		Aspect:    "Engine",
		Status:    "IMPL",
		FilePath:  "test2.go",
		UpdatedAt: "2025-10-16",
		DocPath:   doc2Path,
		DocHash:   hash2,
		DocStatus: "DOC_CURRENT",
	}

	token3 := &storage.Token{
		ReqID:     "CBIN-TEST-3",
		Feature:   "Feature3",
		Aspect:    "Engine",
		Status:    "IMPL",
		FilePath:  "test3.go",
		UpdatedAt: "2025-10-16",
		DocPath:   doc3Path,
		DocHash:   hash3,
		DocStatus: "DOC_CURRENT",
	}

	db.UpsertToken(token1)
	db.UpsertToken(token2)
	db.UpsertToken(token3)

	// Modify doc2 to make it stale
	os.WriteFile(doc2Path, []byte("# Doc 2 - Modified"), 0644)

	// Verify doc2 is now stale
	status, _ := docs.CheckStaleness(token2)
	if status != "DOC_STALE" {
		t.Errorf("expected doc2 to be stale, got %s", status)
	}

	// Get all tokens and check staleness
	tokens, _ := db.ListTokens(map[string]string{}, "", "req_id ASC", 0)

	staleCount := 0
	currentCount := 0

	for _, token := range tokens {
		if token.DocPath == "" {
			continue
		}

		results, err := docs.CheckMultipleDocumentation(token)
		if err != nil {
			continue
		}

		for _, s := range results {
			if s == "DOC_STALE" {
				staleCount++
			} else if s == "DOC_CURRENT" {
				currentCount++
			}
		}
	}

	// Should have 1 stale and 2 current
	if staleCount != 1 {
		t.Errorf("expected 1 stale doc before update, got %d", staleCount)
	}
	if currentCount != 2 {
		t.Errorf("expected 2 current docs before update, got %d", currentCount)
	}

	// Update only stale documentation
	staleOnlyUpdated := 0
	for _, token := range tokens {
		if token.DocPath == "" {
			continue
		}

		results, err := docs.CheckMultipleDocumentation(token)
		if err != nil {
			continue
		}

		hasStale := false
		for _, status := range results {
			if status == "DOC_STALE" {
				hasStale = true
				break
			}
		}

		if !hasStale {
			continue // Skip non-stale docs (simulating --stale-only)
		}

		// Recalculate hash
		newHash, err := docs.CalculateHash(token.DocPath)
		if err != nil {
			continue
		}

		token.DocHash = newHash
		token.DocStatus = "DOC_CURRENT"
		db.UpsertToken(token)
		staleOnlyUpdated++
	}

	if staleOnlyUpdated != 1 {
		t.Errorf("expected to update 1 stale doc, updated %d", staleOnlyUpdated)
	}

	// Verify all docs are now current
	tokens, _ = db.ListTokens(map[string]string{}, "", "req_id ASC", 0)

	allCurrent := true
	for _, token := range tokens {
		if token.DocPath == "" {
			continue
		}

		status, err := docs.CheckStaleness(token)
		if err != nil {
			continue
		}

		if status != "DOC_CURRENT" {
			allCurrent = false
			t.Errorf("expected all docs to be current after update, but %s is %s", token.ReqID, status)
		}
	}

	if !allCurrent {
		t.Error("not all documentation is current after batch update")
	}
}

// TestCANARY_CBIN_136_CLI_DocReport verifies documentation reporting
func TestCANARY_CBIN_136_CLI_DocReport(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	// Create and migrate database
	if err := storage.MigrateDB(dbPath, storage.MigrateAll); err != nil {
		t.Fatalf("failed to migrate database: %v", err)
	}

	db, err := storage.Open(dbPath)
	if err != nil {
		t.Fatalf("failed to open database: %v", err)
	}
	defer db.Close()

	// Create test documentation files
	userDocPath := filepath.Join(tmpDir, "user.md")
	apiDocPath := filepath.Join(tmpDir, "api.md")
	techDocPath := filepath.Join(tmpDir, "tech.md")
	missingDocPath := filepath.Join(tmpDir, "missing.md")

	os.WriteFile(userDocPath, []byte("# User Doc"), 0644)
	os.WriteFile(apiDocPath, []byte("# API Doc"), 0644)
	os.WriteFile(techDocPath, []byte("# Tech Doc"), 0644)
	// Don't create missingDocPath

	// Calculate hashes
	userHash, _ := docs.CalculateHash(userDocPath)
	apiHash, _ := docs.CalculateHash(apiDocPath)
	_, _ = docs.CalculateHash(techDocPath) // Not used - token3 has old hash intentionally

	// Create tokens with various documentation states
	// Token 1: Has user documentation (current)
	token1 := &storage.Token{
		ReqID:     "CBIN-TEST-DOC-1",
		Feature:   "Feature1",
		Aspect:    "Engine",
		Status:    "IMPL",
		FilePath:  "test1.go",
		UpdatedAt: "2025-10-16",
		DocPath:   "user:" + userDocPath,
		DocHash:   userHash,
		DocType:   "user",
		DocStatus: "DOC_CURRENT",
	}

	// Token 2: Has API documentation (current)
	token2 := &storage.Token{
		ReqID:     "CBIN-TEST-DOC-2",
		Feature:   "Feature2",
		Aspect:    "API",
		Status:    "IMPL",
		FilePath:  "test2.go",
		UpdatedAt: "2025-10-16",
		DocPath:   "api:" + apiDocPath,
		DocHash:   apiHash,
		DocType:   "api",
		DocStatus: "DOC_CURRENT",
	}

	// Token 3: Has technical documentation (stale)
	token3 := &storage.Token{
		ReqID:     "CBIN-TEST-DOC-3",
		Feature:   "Feature3",
		Aspect:    "Engine",
		Status:    "IMPL",
		FilePath:  "test3.go",
		UpdatedAt: "2025-10-16",
		DocPath:   "technical:" + techDocPath,
		DocHash:   "oldHash123",
		DocType:   "technical",
		DocStatus: "DOC_STALE",
	}

	// Token 4: References missing documentation
	token4 := &storage.Token{
		ReqID:     "CBIN-TEST-DOC-4",
		Feature:   "Feature4",
		Aspect:    "CLI",
		Status:    "IMPL",
		FilePath:  "test4.go",
		UpdatedAt: "2025-10-16",
		DocPath:   "user:" + missingDocPath,
		DocHash:   "hash456",
		DocType:   "user",
		DocStatus: "DOC_MISSING",
	}

	// Token 5: No documentation
	token5 := &storage.Token{
		ReqID:     "CBIN-TEST-DOC-5",
		Feature:   "Feature5",
		Aspect:    "Storage",
		Status:    "IMPL",
		FilePath:  "test5.go",
		UpdatedAt: "2025-10-16",
	}

	// Token 6: Different requirement, no docs
	token6 := &storage.Token{
		ReqID:     "CBIN-TEST-OTHER",
		Feature:   "OtherFeature",
		Aspect:    "Engine",
		Status:    "IMPL",
		FilePath:  "test6.go",
		UpdatedAt: "2025-10-16",
	}

	db.UpsertToken(token1)
	db.UpsertToken(token2)
	db.UpsertToken(token3)
	db.UpsertToken(token4)
	db.UpsertToken(token5)
	db.UpsertToken(token6)

	// Get all tokens to verify reporting logic
	tokens, err := db.ListTokens(map[string]string{}, "", "req_id ASC", 0)
	if err != nil {
		t.Fatalf("failed to list tokens: %v", err)
	}

	// Simulate report generation
	seenRequirements := make(map[string]bool)
	requirementsWithDocs := make(map[string]bool)
	byType := make(map[string]int)
	byStatus := make(map[string]int)
	tokensWithDocs := 0

	for _, token := range tokens {
		if !seenRequirements[token.ReqID] {
			seenRequirements[token.ReqID] = true
		}

		if token.DocPath == "" {
			continue
		}

		tokensWithDocs++
		requirementsWithDocs[token.ReqID] = true

		if token.DocType != "" {
			byType[token.DocType]++
		}

		results, err := docs.CheckMultipleDocumentation(token)
		if err != nil {
			continue
		}

		for _, status := range results {
			byStatus[status]++
		}
	}

	// Verify statistics
	if len(seenRequirements) != 6 {
		t.Errorf("expected 6 unique requirements, got %d", len(seenRequirements))
	}

	if len(requirementsWithDocs) != 4 {
		t.Errorf("expected 4 requirements with docs, got %d", len(requirementsWithDocs))
	}

	if tokensWithDocs != 4 {
		t.Errorf("expected 4 tokens with docs, got %d", tokensWithDocs)
	}

	// Check documentation by type
	if byType["user"] != 2 {
		t.Errorf("expected 2 user docs, got %d", byType["user"])
	}
	if byType["api"] != 1 {
		t.Errorf("expected 1 api doc, got %d", byType["api"])
	}
	if byType["technical"] != 1 {
		t.Errorf("expected 1 technical doc, got %d", byType["technical"])
	}

	// Check documentation status
	if byStatus["DOC_CURRENT"] != 2 {
		t.Errorf("expected 2 current docs, got %d", byStatus["DOC_CURRENT"])
	}
	if byStatus["DOC_STALE"] != 1 {
		t.Errorf("expected 1 stale doc, got %d", byStatus["DOC_STALE"])
	}
	if byStatus["DOC_MISSING"] != 1 {
		t.Errorf("expected 1 missing doc, got %d", byStatus["DOC_MISSING"])
	}

	// Verify coverage calculation
	coveragePercent := float64(len(requirementsWithDocs)) / float64(len(seenRequirements)) * 100
	expectedCoverage := 66.67 // 4/6 * 100
	if coveragePercent < expectedCoverage-1 || coveragePercent > expectedCoverage+1 {
		t.Errorf("expected coverage around %.2f%%, got %.2f%%", expectedCoverage, coveragePercent)
	}

	// Verify undocumented requirements list
	undocumented := []string{}
	for reqID := range seenRequirements {
		if !requirementsWithDocs[reqID] {
			undocumented = append(undocumented, reqID)
		}
	}

	if len(undocumented) != 2 {
		t.Errorf("expected 2 undocumented requirements, got %d", len(undocumented))
	}
}
