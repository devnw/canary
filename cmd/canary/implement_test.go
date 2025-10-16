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

// CANARY: REQ=CBIN-133; FEATURE="ImplementCmdTests"; ASPECT=CLI; STATUS=TESTED; TEST=TestCANARY_CBIN_133_CLI_ExactMatch; OWNER=canary; UPDATED=2025-10-16
package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCANARY_CBIN_133_CLI_ExactMatch(t *testing.T) {
	// Setup: Create test spec directory
	tmpDir := t.TempDir()
	specDir := filepath.Join(tmpDir, ".canary", "specs", "CBIN-105-test-feature")
	if err := os.MkdirAll(specDir, 0755); err != nil {
		t.Fatalf("failed to create spec dir: %v", err)
	}

	specContent := `# Feature Specification: TestFeature

**Requirement ID:** CBIN-105
**Feature Name:** TestFeature
**Status:** STUB

## Feature Overview
Test feature for exact ID matching.
`
	if err := os.WriteFile(filepath.Join(specDir, "spec.md"), []byte(specContent), 0644); err != nil {
		t.Fatalf("failed to write spec: %v", err)
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
	if !strings.Contains(spec.SpecContent, "TestFeature") {
		t.Error("Spec content not loaded correctly")
	}
}

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
			t.Fatalf("failed to create spec dir: %v", err)
		}
		content := fmt.Sprintf("# Feature Specification: %s\n\n**Requirement ID:** %s", s.name, s.id)
		if err := os.WriteFile(filepath.Join(specDir, "spec.md"), []byte(content), 0644); err != nil {
			t.Fatalf("failed to write spec: %v", err)
		}
	}

	originalWd, _ := os.Getwd()

	defer os.Chdir(originalWd)
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("failed to change directory: %v", err)
	}

	// Execute: Fuzzy search for "user auth" (should auto-select UserAuthentication)
	spec, err := findRequirementAutoSelect("user auth")

	// Verify: Matches UserAuthentication
	if err != nil {
		t.Fatalf("findRequirement failed: %v", err)
	}
	if spec.ReqID != "CBIN-105" {
		t.Errorf("Expected CBIN-105 (UserAuthentication), got %s", spec.ReqID)
	}
}

func TestCANARY_CBIN_133_CLI_PromptGeneration(t *testing.T) {
	// Setup: Create spec with Implementation Checklist
	tmpDir := t.TempDir()
	specDir := filepath.Join(tmpDir, ".canary", "specs", "CBIN-105-test-feature")
	templatesDir := filepath.Join(tmpDir, ".canary", "templates")
	memoryDir := filepath.Join(tmpDir, ".canary", "memory")

	for _, dir := range []string{specDir, templatesDir, memoryDir} {
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatalf("failed to create dir: %v", err)
		}
	}

	specContent := `# Feature Specification: TestFeature

**Requirement ID:** CBIN-105

## Implementation Checklist

### Phase 1: Core Features

<!-- CANARY: REQ=CBIN-105; FEATURE="Feature1"; ASPECT=API; STATUS=STUB; UPDATED=2025-10-16 -->
**Feature 1: Test Feature**
- [ ] Implement test functionality
`
	if err := os.WriteFile(filepath.Join(specDir, "spec.md"), []byte(specContent), 0644); err != nil {
		t.Fatalf("failed to write spec: %v", err)
	}

	// Create minimal template
	templateContent := `# Implementation Guidance: {{.FeatureName}}

**Requirement:** {{.ReqID}}

## Specification
{{.SpecContent}}

## Checklist
{{.Checklist}}
`
	if err := os.WriteFile(filepath.Join(templatesDir, "implement-prompt-template.md"), []byte(templateContent), 0644); err != nil {
		t.Fatalf("failed to write template: %v", err)
	}

	// Create minimal constitution
	constitutionContent := `# Constitution
## Article IV: Test-First Imperative
Tests must be written before implementation.
`
	if err := os.WriteFile(filepath.Join(memoryDir, "constitution.md"), []byte(constitutionContent), 0644); err != nil {
		t.Fatalf("failed to write constitution: %v", err)
	}

	originalWd, _ := os.Getwd()

	defer os.Chdir(originalWd)
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("failed to change directory: %v", err)
	}

	// Execute: Generate prompt
	spec, err := findRequirement("CBIN-105")
	if err != nil {
		t.Fatalf("findRequirement failed: %v", err)
	}

	prompt, err := renderImplementPrompt(spec, &ImplementFlags{Prompt: true})

	// Verify: Prompt contains key sections
	if err != nil {
		t.Fatalf("renderImplementPrompt failed: %v", err)
	}

	requiredSections := []string{
		"Implementation Guidance:",
		"CBIN-105",
		"Specification",
		"Checklist",
	}

	for _, section := range requiredSections {
		if !strings.Contains(prompt, section) {
			t.Errorf("Prompt missing section: %s", section)
		}
	}
}

func TestCANARY_CBIN_133_CLI_ProgressTracking(t *testing.T) {
	// Setup: Create files with CANARY tokens
	tmpDir := t.TempDir()

	files := []struct {
		name    string
		content string
	}{
		{"feature1.go", "// CANARY: REQ=CBIN-105; FEATURE=\"Feature1\"; ASPECT=API; STATUS=IMPL; UPDATED=2025-10-16\npackage main"},
		{"feature2.go", "// CANARY: REQ=CBIN-105; FEATURE=\"Feature2\"; ASPECT=API; STATUS=TESTED; TEST=TestFeature2; UPDATED=2025-10-16\npackage main"},
		{"feature3.go", "// CANARY: REQ=CBIN-105; FEATURE=\"Feature3\"; ASPECT=API; STATUS=STUB; UPDATED=2025-10-16\npackage main"},
	}

	for _, f := range files {
		if err := os.WriteFile(filepath.Join(tmpDir, f.name), []byte(f.content), 0644); err != nil {
			t.Fatalf("failed to write file: %v", err)
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
	if progress.Tested != 1 {
		t.Errorf("Expected 1 tested feature, got %d", progress.Tested)
	}
	if progress.Impl != 1 {
		t.Errorf("Expected 1 impl feature, got %d", progress.Impl)
	}
	if progress.Stub != 1 {
		t.Errorf("Expected 1 stub feature, got %d", progress.Stub)
	}
}

func TestCANARY_CBIN_133_CLI_MissingSpec(t *testing.T) {
	// Setup: Empty specs directory
	tmpDir := t.TempDir()
	specsDir := filepath.Join(tmpDir, ".canary", "specs")
	if err := os.MkdirAll(specsDir, 0755); err != nil {
		t.Fatalf("failed to create specs dir: %v", err)
	}

	originalWd, _ := os.Getwd()

	defer os.Chdir(originalWd)
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("failed to change directory: %v", err)
	}

	// Execute: Try to find non-existent spec
	_, err := findRequirement("CBIN-999")

	// Verify: Error returned
	if err == nil {
		t.Error("Expected error for missing spec, got nil")
	}
	if !strings.Contains(err.Error(), "not found") && !strings.Contains(err.Error(), "no matches") {
		t.Errorf("Expected 'not found' error, got: %v", err)
	}
}

// Helper function for auto-select testing (bypasses interactive selection)
func findRequirementAutoSelect(query string) (*RequirementSpec, error) {
	// This will be implemented to force auto-selection behavior
	return findRequirement(query)
}
