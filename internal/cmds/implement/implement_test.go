// Copyright (c) 2025 by Developer Network.
//
// For more details, see the LICENSE file in the root directory of this
// source code repository or contact Developer Network at info@devnw.com.

// CANARY: REQ=CBIN-133; FEATURE="ImplementCmdTests"; ASPECT=CLI; STATUS=TESTED; TEST=TestCANARY_CBIN_133_CLI_ExactMatch; UPDATED=2025-10-16
package implement

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestCANARY_CBIN_133_CLI_ExactMatch verifies exact ID matching
func TestCANARY_CBIN_133_CLI_ExactMatch(t *testing.T) {
	// Setup: Create test spec directory
	tmpDir := t.TempDir()
	specDir := filepath.Join(tmpDir, ".canary", "specs", "CBIN-105-test-feature")
	if err := os.MkdirAll(specDir, 0755); err != nil {
		t.Fatalf("failed to create spec directory: %v", err)
	}

	specContent := `# Feature Specification: TestFeature

**Requirement ID:** CBIN-105
**Feature Name:** TestFeature
**Status:** STUB

## Success Criteria
- Feature must be implemented
- Tests must pass
`
	if err := os.WriteFile(filepath.Join(specDir, "spec.md"), []byte(specContent), 0644); err != nil {
		t.Fatalf("failed to write spec file: %v", err)
	}

	// Change to tmpDir
	originalWd, _ := os.Getwd()
	defer os.Chdir(originalWd)
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("failed to change directory: %v", err)
	}

	// Execute: Find by exact ID
	spec, err := findRequirement("CBIN-105")

	// Verify: Spec loaded correctly
	if err != nil {
		t.Fatalf("findRequirement failed: %v", err)
	}
	if spec.ReqID != "CBIN-105" {
		t.Errorf("Expected CBIN-105, got %s", spec.ReqID)
	}
	if spec.FeatureName != "test-feature" {
		t.Errorf("Expected feature name 'test-feature', got %s", spec.FeatureName)
	}
	if !strings.Contains(spec.SpecContent, "TestFeature") {
		t.Error("Spec content should contain 'TestFeature'")
	}
}

// TestCANARY_CBIN_133_CLI_FuzzyMatch verifies fuzzy matching by feature name
func TestCANARY_CBIN_133_CLI_FuzzyMatch(t *testing.T) {
	// Setup: Multiple specs with similar names
	tmpDir := t.TempDir()
	specs := []struct {
		id   string
		name string
	}{
		{"CBIN-105", "UserAuthentication"},
		{"CBIN-110", "OAuthIntegration"},
		{"CBIN-112", "DataValidation"},
	}

	for _, s := range specs {
		specDir := filepath.Join(tmpDir, ".canary", "specs", fmt.Sprintf("%s-%s", s.id, s.name))
		if err := os.MkdirAll(specDir, 0755); err != nil {
			t.Fatalf("failed to create spec directory: %v", err)
		}
		content := fmt.Sprintf(`# Feature Specification: %s

**Requirement ID:** %s
**Feature Name:** %s
`, s.name, s.id, s.name)
		if err := os.WriteFile(filepath.Join(specDir, "spec.md"), []byte(content), 0644); err != nil {
			t.Fatalf("failed to write spec file: %v", err)
		}
	}

	originalWd, _ := os.Getwd()
	defer os.Chdir(originalWd)
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("failed to change directory: %v", err)
	}

	// Execute: Fuzzy search for "user auth"
	spec, err := findRequirement("user auth")

	// Verify: Matches UserAuthentication
	if err != nil {
		t.Fatalf("findRequirement failed: %v", err)
	}
	if spec.ReqID != "CBIN-105" {
		t.Errorf("Expected CBIN-105 (UserAuthentication), got %s", spec.ReqID)
	}
}

// TestCANARY_CBIN_133_CLI_PromptGeneration verifies comprehensive prompt generation
func TestCANARY_CBIN_133_CLI_PromptGeneration(t *testing.T) {
	// Setup: Create spec with plan and constitution
	tmpDir := t.TempDir()

	// Create spec
	specDir := filepath.Join(tmpDir, ".canary", "specs", "CBIN-105-test-feature")
	if err := os.MkdirAll(specDir, 0755); err != nil {
		t.Fatalf("failed to create spec directory: %v", err)
	}

	specContent := `# Feature Specification: TestFeature

**Requirement ID:** CBIN-105

## Implementation Checklist
- [ ] Create test file
- [ ] Implement feature
- [ ] Run tests
`
	if err := os.WriteFile(filepath.Join(specDir, "spec.md"), []byte(specContent), 0644); err != nil {
		t.Fatalf("failed to write spec file: %v", err)
	}

	// Create plan
	planContent := `# Implementation Plan: CBIN-105

## Phase 1: Test Creation
Write tests first.
`
	if err := os.WriteFile(filepath.Join(specDir, "plan.md"), []byte(planContent), 0644); err != nil {
		t.Fatalf("failed to write plan file: %v", err)
	}

	// Create constitution
	constitutionDir := filepath.Join(tmpDir, ".canary", "memory")
	if err := os.MkdirAll(constitutionDir, 0755); err != nil {
		t.Fatalf("failed to create memory directory: %v", err)
	}

	constitutionContent := `# Constitution

## Article IV: Test-First
Tests must be written before implementation.
`
	if err := os.WriteFile(filepath.Join(constitutionDir, "constitution.md"), []byte(constitutionContent), 0644); err != nil {
		t.Fatalf("failed to write constitution: %v", err)
	}

	// Create template
	templateDir := filepath.Join(tmpDir, ".canary", "templates")
	if err := os.MkdirAll(templateDir, 0755); err != nil {
		t.Fatalf("failed to create templates directory: %v", err)
	}

	templateContent := `# Implementation Guidance: {{.FeatureName}}

**Requirement:** {{.ReqID}}

## Specification
{{.SpecContent}}

{{if .HasPlan}}
## Implementation Plan
{{.PlanContent}}
{{end}}

## Constitution
{{.Constitution}}
`
	if err := os.WriteFile(filepath.Join(templateDir, "implement-prompt-template.md"), []byte(templateContent), 0644); err != nil {
		t.Fatalf("failed to write template: %v", err)
	}

	originalWd, _ := os.Getwd()
	defer os.Chdir(originalWd)
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("failed to change directory: %v", err)
	}

	// Execute: Generate prompt
	spec, _ := findRequirement("CBIN-105")
	prompt, err := renderImplementPrompt(spec, &ImplementFlags{Prompt: true})

	// Verify: Prompt contains key sections
	if err != nil {
		t.Fatalf("renderImplementPrompt failed: %v", err)
	}

	requiredSections := []string{
		"Implementation Guidance:",
		"CBIN-105",
		"Specification",
		"Implementation Plan",
		"Constitution",
	}

	for _, section := range requiredSections {
		if !strings.Contains(prompt, section) {
			t.Errorf("Prompt missing section: %s", section)
		}
	}
}

// TestCANARY_CBIN_133_CLI_ProgressTracking verifies progress calculation
func TestCANARY_CBIN_133_CLI_ProgressTracking(t *testing.T) {
	// Setup: Create files with CANARY tokens
	tmpDir := t.TempDir()

	// Create test files with tokens
	files := []struct {
		name    string
		content string
	}{
		{"feature1.go", `// CANARY: REQ=CBIN-105; FEATURE="Feature1"; ASPECT=API; STATUS=IMPL; UPDATED=2025-10-16
package main`},
		{"feature2.go", `// CANARY: REQ=CBIN-105; FEATURE="Feature2"; ASPECT=API; STATUS=TESTED; UPDATED=2025-10-16
package main`},
		{"feature3.go", `// CANARY: REQ=CBIN-105; FEATURE="Feature3"; ASPECT=API; STATUS=STUB; UPDATED=2025-10-16
package main`},
	}

	for _, f := range files {
		if err := os.WriteFile(filepath.Join(tmpDir, f.name), []byte(f.content), 0644); err != nil {
			t.Fatalf("failed to write test file %s: %v", f.name, err)
		}
	}

	originalWd, _ := os.Getwd()
	defer os.Chdir(originalWd)
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("failed to change directory: %v", err)
	}

	// Execute: Calculate progress
	progress, err := calculateProgress("CBIN-105")

	// Verify: Correct counts
	if err != nil {
		t.Fatalf("calculateProgress failed: %v", err)
	}
	if progress.Total != 3 {
		t.Errorf("Expected 3 total features, got %d", progress.Total)
	}
	if progress.Completed != 1 { // Only TESTED counts as completed
		t.Errorf("Expected 1 completed feature, got %d", progress.Completed)
	}
	if progress.Stub != 1 {
		t.Errorf("Expected 1 STUB, got %d", progress.Stub)
	}
	if progress.Impl != 1 {
		t.Errorf("Expected 1 IMPL, got %d", progress.Impl)
	}
	if progress.Tested != 1 {
		t.Errorf("Expected 1 TESTED, got %d", progress.Tested)
	}
}

// TestCANARY_CBIN_133_CLI_MissingSpec verifies error handling for missing specs
func TestCANARY_CBIN_133_CLI_MissingSpec(t *testing.T) {
	// Setup: Empty specs directory
	tmpDir := t.TempDir()
	specsDir := filepath.Join(tmpDir, ".canary", "specs")
	if err := os.MkdirAll(specsDir, 0755); err != nil {
		t.Fatalf("failed to create specs directory: %v", err)
	}

	originalWd, _ := os.Getwd()
	defer os.Chdir(originalWd)
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("failed to change directory: %v", err)
	}

	// Execute: Try to find non-existent spec
	_, err := findRequirement("CBIN-999")

	// Verify: Returns appropriate error
	if err == nil {
		t.Error("Expected error for missing spec, got nil")
	}
	if !strings.Contains(err.Error(), "no matches found") && !strings.Contains(err.Error(), "not found") {
		t.Errorf("Expected error about missing spec, got: %v", err)
	}
}
