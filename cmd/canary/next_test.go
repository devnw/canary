// CANARY: REQ=CBIN-132; FEATURE="NextCmd"; ASPECT=CLI; STATUS=BENCHED; TEST=TestCANARY_CBIN_132_CLI_NextPrioritySelection; BENCH=BenchmarkCANARY_CBIN_132_CLI_PriorityQuery; OWNER=canary; UPDATED=2025-10-16
package main

import (
	"os"
	"path/filepath"
	"testing"

	"go.spyder.org/canary/internal/storage"
)

// TestCANARY_CBIN_132_CLI_NextPrioritySelection verifies priority-based selection logic
func TestCANARY_CBIN_132_CLI_NextPrioritySelection(t *testing.T) {
	// Setup: Create test database with 5 requirements of different priorities
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	db, err := storage.Open(dbPath)
	if err != nil {
		t.Fatalf("failed to open database: %v", err)
	}

	defer db.Close()

	// Initialize schema
	if err := storage.AutoMigrate(dbPath); err != nil {
		t.Fatalf("failed to migrate database: %v", err)
	}

	// Insert test tokens with different priorities
	testTokens := []*storage.Token{
		{
			ReqID:    "CBIN-201",
			Feature:  "LowPriority",
			Aspect:   "API",
			Status:   "STUB",
			Priority: 10, // Lowest
			FilePath: "test1.go",
		},
		{
			ReqID:    "CBIN-202",
			Feature:  "HighPriority",
			Aspect:   "CLI",
			Status:   "STUB",
			Priority: 1, // Highest
			FilePath: "test2.go",
		},
		{
			ReqID:    "CBIN-203",
			Feature:  "MediumPriority",
			Aspect:   "API",
			Status:   "STUB",
			Priority: 5,
			FilePath: "test3.go",
		},
		{
			ReqID:    "CBIN-204",
			Feature:  "AlreadyTested",
			Aspect:   "API",
			Status:   "TESTED",
			Priority: 2,
			FilePath: "test4.go",
		},
	}

	for _, token := range testTokens {
		if err := db.UpsertToken(token); err != nil {
			t.Fatalf("failed to insert test token: %v", err)
		}
	}

	// Execute: Select next priority with no filters
	selected, err := selectNextPriority(dbPath, nil)
	if err != nil {
		t.Fatalf("selectNextPriority failed: %v", err)
	}

	// Verify: Returns requirement with PRIORITY=1 and STATUS=STUB
	if selected == nil {
		t.Fatal("expected a token to be selected, got nil")
	}
	if selected.ReqID != "CBIN-202" {
		t.Errorf("expected CBIN-202 (highest priority STUB), got %s", selected.ReqID)
	}
	if selected.Priority != 1 {
		t.Errorf("expected priority 1, got %d", selected.Priority)
	}
	if selected.Status != "STUB" {
		t.Errorf("expected STATUS=STUB, got %s", selected.Status)
	}
}

// TestCANARY_CBIN_132_CLI_DependencyBlocking verifies dependency resolution
func TestCANARY_CBIN_132_CLI_DependencyBlocking(t *testing.T) {
	// Setup: Create test database where highest priority has unresolved DEPENDS_ON
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	db, err := storage.Open(dbPath)
	if err != nil {
		t.Fatalf("failed to open database: %v", err)
	}

	defer db.Close()

	if err := storage.AutoMigrate(dbPath); err != nil {
		t.Fatalf("failed to migrate database: %v", err)
	}

	// Insert tokens with dependency chain
	testTokens := []*storage.Token{
		{
			ReqID:     "CBIN-301",
			Feature:   "BlockedHighPriority",
			Aspect:    "API",
			Status:    "STUB",
			Priority:  1,
			FilePath:  "test1.go",
			DependsOn: "CBIN-302", // Depends on unresolved requirement
		},
		{
			ReqID:    "CBIN-302",
			Feature:  "DependencyRequirement",
			Aspect:   "API",
			Status:   "STUB",
			Priority: 5,
			FilePath: "test2.go",
		},
		{
			ReqID:    "CBIN-303",
			Feature:  "IndependentLowPriority",
			Aspect:   "CLI",
			Status:   "STUB",
			Priority: 8,
			FilePath: "test3.go",
		},
	}

	for _, token := range testTokens {
		if err := db.UpsertToken(token); err != nil {
			t.Fatalf("failed to insert test token: %v", err)
		}
	}

	// Execute: Select next priority
	selected, err := selectNextPriority(dbPath, nil)
	if err != nil {
		t.Fatalf("selectNextPriority failed: %v", err)
	}

	// Verify: Skips blocked CBIN-301, returns CBIN-302 (the dependency)
	if selected == nil {
		t.Fatal("expected a token to be selected, got nil")
	}
	if selected.ReqID != "CBIN-302" {
		t.Errorf("expected CBIN-302 (dependency), got %s", selected.ReqID)
	}
}

// TestCANARY_CBIN_132_CLI_TemplateRendering verifies prompt generation from template
func TestCANARY_CBIN_132_CLI_TemplateRendering(t *testing.T) {
	// Setup: Create test token with spec file and constitution
	tmpDir := t.TempDir()

	// Create .canary structure
	canaryDir := filepath.Join(tmpDir, ".canary")
	specsDir := filepath.Join(canaryDir, "specs", "CBIN-401-test-feature")
	memoryDir := filepath.Join(canaryDir, "memory")
	templatesDir := filepath.Join(canaryDir, "templates")

	for _, dir := range []string{specsDir, memoryDir, templatesDir} {
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatalf("failed to create directory %s: %v", dir, err)
		}
	}

	// Create minimal spec file
	specContent := `# Feature Specification: Test Feature

**Requirement ID:** CBIN-401
**Feature Name:** TestFeature
**Status:** STUB

## Success Criteria
- Requirement must be verified
- All tests must pass
`
	specFile := filepath.Join(specsDir, "spec.md")
	if err := os.WriteFile(specFile, []byte(specContent), 0644); err != nil {
		t.Fatalf("failed to write spec file: %v", err)
	}

	// Create minimal constitution
	constitutionContent := `# Constitution

## Article IV: Test-First Imperative
Tests must be written before implementation.
`
	constitutionFile := filepath.Join(memoryDir, "constitution.md")
	if err := os.WriteFile(constitutionFile, []byte(constitutionContent), 0644); err != nil {
		t.Fatalf("failed to write constitution: %v", err)
	}

	// Create minimal template
	templateContent := `# Implementation Guidance: {{.Feature}}

**Requirement:** {{.ReqID}}
**Priority:** {{.Priority}}

## Constitution
{{.Constitution}}

## Specification
{{.SpecContent}}
`
	templateFile := filepath.Join(templatesDir, "next-prompt-template.md")
	if err := os.WriteFile(templateFile, []byte(templateContent), 0644); err != nil {
		t.Fatalf("failed to write template: %v", err)
	}

	// Create test token
	token := &storage.Token{
		ReqID:    "CBIN-401",
		Feature:  "TestFeature",
		Aspect:   "API",
		Status:   "STUB",
		Priority: 3,
		FilePath: "test.go",
	}

	// Change to tmpDir for relative path resolution
	originalWd, _ := os.Getwd()

	defer os.Chdir(originalWd)
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("failed to change directory: %v", err)
	}

	// Execute: Render prompt with --prompt flag
	prompt, err := renderPrompt(token, true)
	if err != nil {
		t.Fatalf("renderPrompt failed: %v", err)
	}

	// Verify: Prompt contains spec content, constitution, and test guidance
	if prompt == "" {
		t.Fatal("expected non-empty prompt")
	}

	// Check for key sections
	if !contains(prompt, "CBIN-401") {
		t.Error("prompt missing requirement ID")
	}
	if !contains(prompt, "TestFeature") {
		t.Error("prompt missing feature name")
	}
	if !contains(prompt, "Test-First Imperative") {
		t.Error("prompt missing constitution content")
	}
	if !contains(prompt, "Success Criteria") {
		t.Error("prompt missing spec content")
	}
}

// TestCANARY_CBIN_132_CLI_NoWorkAvailable verifies behavior when all requirements are complete
func TestCANARY_CBIN_132_CLI_NoWorkAvailable(t *testing.T) {
	// Setup: Create database with only TESTED/BENCHED tokens
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	db, err := storage.Open(dbPath)
	if err != nil {
		t.Fatalf("failed to open database: %v", err)
	}

	defer db.Close()

	if err := storage.AutoMigrate(dbPath); err != nil {
		t.Fatalf("failed to migrate database: %v", err)
	}

	// Insert only completed tokens
	testTokens := []*storage.Token{
		{
			ReqID:    "CBIN-501",
			Feature:  "CompleteFeature1",
			Aspect:   "API",
			Status:   "TESTED",
			Priority: 1,
			FilePath: "test1.go",
		},
		{
			ReqID:    "CBIN-502",
			Feature:  "CompleteFeature2",
			Aspect:   "CLI",
			Status:   "BENCHED",
			Priority: 2,
			FilePath: "test2.go",
		},
	}

	for _, token := range testTokens {
		if err := db.UpsertToken(token); err != nil {
			t.Fatalf("failed to insert test token: %v", err)
		}
	}

	// Execute: Select next priority
	selected, err := selectNextPriority(dbPath, nil)

	// Verify: Returns nil (no work available) without error
	if err != nil {
		t.Errorf("expected no error, got: %v", err)
	}
	if selected != nil {
		t.Errorf("expected nil (no work available), got token: %s", selected.ReqID)
	}
}

// TestCANARY_CBIN_132_CLI_FilesystemFallback verifies filesystem scan when database unavailable
func TestCANARY_CBIN_132_CLI_FilesystemFallback(t *testing.T) {
	// Setup: Create test directory with CANARY tokens in files
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "nonexistent.db") // Database doesn't exist

	// Create test files with CANARY tokens
	testFile := filepath.Join(tmpDir, "test.go")
	fileContent := `package test

// CANARY: REQ=CBIN-601; FEATURE="FilesystemTest"; ASPECT=API; STATUS=STUB; PRIORITY=1; UPDATED=2025-10-16
func FilesystemTest() {
	// stub
}
`
	if err := os.WriteFile(testFile, []byte(fileContent), 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	// Change to tmpDir
	originalWd, _ := os.Getwd()

	defer os.Chdir(originalWd)
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("failed to change directory: %v", err)
	}

	// Execute: Select next priority (should fall back to filesystem)
	selected, err := selectNextPriority(dbPath, nil)

	// Verify: Falls back to filesystem scan and finds token
	if err != nil {
		t.Fatalf("filesystem fallback failed: %v", err)
	}
	if selected == nil {
		t.Fatal("expected token from filesystem, got nil")
	}
	if selected.ReqID != "CBIN-601" {
		t.Errorf("expected CBIN-601 from filesystem, got %s", selected.ReqID)
	}
}

// TestCANARY_CBIN_132_CLI_StatusFiltering verifies filtering by status
func TestCANARY_CBIN_132_CLI_StatusFiltering(t *testing.T) {
	// Setup: Create database with mixed statuses
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	db, err := storage.Open(dbPath)
	if err != nil {
		t.Fatalf("failed to open database: %v", err)
	}

	defer db.Close()

	if err := storage.AutoMigrate(dbPath); err != nil {
		t.Fatalf("failed to migrate database: %v", err)
	}

	testTokens := []*storage.Token{
		{
			ReqID:    "CBIN-701",
			Feature:  "StubFeature",
			Aspect:   "API",
			Status:   "STUB",
			Priority: 2,
			FilePath: "test1.go",
		},
		{
			ReqID:    "CBIN-702",
			Feature:  "ImplFeature",
			Aspect:   "API",
			Status:   "IMPL",
			Priority: 1,
			FilePath: "test2.go",
		},
	}

	for _, token := range testTokens {
		if err := db.UpsertToken(token); err != nil {
			t.Fatalf("failed to insert test token: %v", err)
		}
	}

	// Execute: Select with IMPL filter
	filters := map[string]string{"status": "IMPL"}
	selected, err := selectNextPriority(dbPath, filters)
	if err != nil {
		t.Fatalf("selectNextPriority with filter failed: %v", err)
	}

	// Verify: Returns IMPL token even though STUB has lower numerical priority
	if selected == nil {
		t.Fatal("expected a token to be selected")
	}
	if selected.Status != "IMPL" {
		t.Errorf("expected STATUS=IMPL, got %s", selected.Status)
	}
	if selected.ReqID != "CBIN-702" {
		t.Errorf("expected CBIN-702, got %s", selected.ReqID)
	}
}

// Helper function to check if string contains substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && containsAt(s, substr, 0))
}

func containsAt(s, substr string, start int) bool {
	if start+len(substr) > len(s) {
		return false
	}
	for i := start; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// CANARY: REQ=CBIN-132; FEATURE="NextCmd"; ASPECT=CLI; STATUS=BENCHED; BENCH=BenchmarkCANARY_CBIN_132_CLI_PriorityQuery; OWNER=canary; UPDATED=2025-10-16
// BenchmarkCANARY_CBIN_132_CLI_PriorityQuery measures priority query performance
// Target: <100ms per operation (for <10,000 requirements)
func BenchmarkCANARY_CBIN_132_CLI_PriorityQuery(b *testing.B) {
	// Setup: Create database with 1000 requirements
	tmpDir := b.TempDir()
	dbPath := filepath.Join(tmpDir, "bench.db")

	db, err := storage.Open(dbPath)
	if err != nil {
		b.Fatalf("failed to open database: %v", err)
	}

	defer db.Close()

	if err := storage.AutoMigrate(dbPath); err != nil {
		b.Fatalf("failed to migrate database: %v", err)
	}

	// Insert 1000 test tokens with various priorities
	for i := 0; i < 1000; i++ {
		token := &storage.Token{
			ReqID:    filepath.Join("CBIN-", filepath.Base(tmpDir), "-", string(rune(i))),
			Feature:  filepath.Join("Feature", string(rune(i))),
			Aspect:   "API",
			Status:   "STUB",
			Priority: (i % 10) + 1, // Priority 1-10
			FilePath: filepath.Join("test", string(rune(i)), ".go"),
		}
		if err := db.UpsertToken(token); err != nil {
			b.Fatalf("failed to insert token: %v", err)
		}
	}

	// Reset timer to exclude setup
	b.ResetTimer()

	// Benchmark: selectNextPriority
	for i := 0; i < b.N; i++ {
		_, err := selectNextPriority(dbPath, nil)
		if err != nil {
			b.Fatalf("selectNextPriority failed: %v", err)
		}
	}
}

// CANARY: REQ=CBIN-132; FEATURE="NextCmd"; ASPECT=CLI; STATUS=BENCHED; BENCH=BenchmarkCANARY_CBIN_132_CLI_PromptGeneration; OWNER=canary; UPDATED=2025-10-16
// BenchmarkCANARY_CBIN_132_CLI_PromptGeneration measures prompt rendering performance
// Target: <500ms per operation
func BenchmarkCANARY_CBIN_132_CLI_PromptGeneration(b *testing.B) {
	// Setup: Create test environment with spec and constitution
	tmpDir := b.TempDir()

	// Create .canary structure
	canaryDir := filepath.Join(tmpDir, ".canary")
	specsDir := filepath.Join(canaryDir, "specs", "CBIN-999-bench-feature")
	memoryDir := filepath.Join(canaryDir, "memory")
	templatesDir := filepath.Join(canaryDir, "templates")

	for _, dir := range []string{specsDir, memoryDir, templatesDir} {
		if err := os.MkdirAll(dir, 0755); err != nil {
			b.Fatalf("failed to create directory %s: %v", dir, err)
		}
	}

	// Create spec file
	specContent := `# Feature Specification: Benchmark Feature

**Requirement ID:** CBIN-999
**Feature Name:** BenchmarkFeature
**Status:** STUB

## Feature Overview
This is a test specification for benchmarking prompt generation.

## Success Criteria
1. Performance meets target
2. All sections rendered correctly
3. Template variables properly substituted
`
	specFile := filepath.Join(specsDir, "spec.md")
	if err := os.WriteFile(specFile, []byte(specContent), 0644); err != nil {
		b.Fatalf("failed to write spec file: %v", err)
	}

	// Create constitution file
	constitutionContent := `# CANARY Development Constitution

## Article I: Requirement-First Development
Every feature MUST begin with a CANARY token.

## Article IV: Test-First Imperative
All implementation MUST follow Test-Driven Development.

## Article V: Simplicity and Anti-Abstraction
Features SHOULD use the simplest approach that meets requirements.
`
	constitutionFile := filepath.Join(memoryDir, "constitution.md")
	if err := os.WriteFile(constitutionFile, []byte(constitutionContent), 0644); err != nil {
		b.Fatalf("failed to write constitution: %v", err)
	}

	// Create template file
	templateContent := `# Implementation Guidance: {{.Feature}}

**Requirement ID:** {{.ReqID}}
**Feature Name:** {{.Feature}}
**Aspect:** {{.Aspect}}
**Priority:** {{.Priority}}
**Today's Date:** {{.Today}}

## Constitutional Principles
{{.Constitution}}

## Full Specification
{{.SpecContent}}

## Test Guidance
{{.TestGuidance}}

## Token Example
{{.TokenExample}}
`
	templateFile := filepath.Join(templatesDir, "next-prompt-template.md")
	if err := os.WriteFile(templateFile, []byte(templateContent), 0644); err != nil {
		b.Fatalf("failed to write template: %v", err)
	}

	// Create test token
	token := &storage.Token{
		ReqID:    "CBIN-999",
		Feature:  "BenchmarkFeature",
		Aspect:   "API",
		Status:   "STUB",
		Priority: 1,
		FilePath: "test.go",
	}

	// Change to tmpDir for relative path resolution
	originalWd, _ := os.Getwd()

	defer os.Chdir(originalWd)
	if err := os.Chdir(tmpDir); err != nil {
		b.Fatalf("failed to change directory: %v", err)
	}

	// Reset timer to exclude setup
	b.ResetTimer()

	// Benchmark: renderPrompt with full prompt generation
	for i := 0; i < b.N; i++ {
		_, err := renderPrompt(token, true)
		if err != nil {
			b.Fatalf("renderPrompt failed: %v", err)
		}
	}
}
