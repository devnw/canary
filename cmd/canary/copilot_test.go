// CANARY: REQ=CBIN-148; FEATURE="CopilotInstructionCreator"; ASPECT=CLI; STATUS=BENCHED; TEST=TestCreateCopilotInstructions; BENCH=BenchmarkCreateCopilotInstructions; UPDATED=2025-10-19
package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestCreateCopilotInstructions verifies that Copilot instruction files are created correctly
func TestCreateCopilotInstructions(t *testing.T) {
	// Create temporary directory for testing
	tmpDir := t.TempDir()
	projectKey := "TEST"

	// Act: Create Copilot instructions
	err := createCopilotInstructions(tmpDir, projectKey)
	if err != nil {
		t.Fatalf("createCopilotInstructions failed: %v", err)
	}

	// Assert: Directory structure exists
	instructionsDir := filepath.Join(tmpDir, ".github", "instructions")
	if _, err := os.Stat(instructionsDir); os.IsNotExist(err) {
		t.Errorf(".github/instructions/ directory not created")
	}

	// Assert: Repository-wide instruction file exists
	repoFile := filepath.Join(instructionsDir, "repository.md")
	content, err := os.ReadFile(repoFile)
	if err != nil {
		t.Errorf("repository.md not found: %v", err)
	}

	// Assert: Project key is substituted in template
	if !strings.Contains(string(content), projectKey) {
		t.Errorf("repository.md does not contain project key %q", projectKey)
	}

	// Assert: Path-specific files exist
	expectedFiles := []string{
		filepath.Join(instructionsDir, ".canary", "specs", "instruction.md"),
		filepath.Join(instructionsDir, "tests", "instruction.md"),
		filepath.Join(instructionsDir, ".canary", "instruction.md"),
	}

	for _, file := range expectedFiles {
		if _, err := os.Stat(file); os.IsNotExist(err) {
			t.Errorf("expected instruction file not found: %s", file)
		}
	}
}

// TestCreateCopilotInstructionsPreservesExisting verifies re-init doesn't overwrite
func TestCreateCopilotInstructionsPreservesExisting(t *testing.T) {
	tmpDir := t.TempDir()
	projectKey := "TEST"

	// Arrange: Create custom instruction file
	instructionsDir := filepath.Join(tmpDir, ".github", "instructions")
	os.MkdirAll(instructionsDir, 0755)

	customContent := "# Custom Instructions\n\nDo not overwrite me!"
	customFile := filepath.Join(instructionsDir, "repository.md")
	os.WriteFile(customFile, []byte(customContent), 0644)

	// Act: Run createCopilotInstructions again
	err := createCopilotInstructions(tmpDir, projectKey)
	if err != nil {
		t.Fatalf("createCopilotInstructions failed on re-run: %v", err)
	}

	// Assert: Custom content preserved
	content, err := os.ReadFile(customFile)
	if err != nil {
		t.Fatalf("repository.md not readable: %v", err)
	}

	if string(content) != customContent {
		t.Errorf("existing repository.md was overwritten:\ngot: %s\nwant: %s",
			string(content), customContent)
	}
}

// TestCopilotInstructionTemplateValidity verifies templates are valid markdown
func TestCopilotInstructionTemplateValidity(t *testing.T) {
	templates := []string{
		"base/copilot/repository.md",
		"base/copilot/specs.md",
		"base/copilot/tests.md",
		"base/copilot/canary.md",
	}

	for _, tmpl := range templates {
		content, err := readEmbeddedFile(tmpl)
		if err != nil {
			t.Errorf("template %s not found: %v", tmpl, err)
			continue
		}

		// Basic validation: non-empty, contains CANARY token
		if len(content) == 0 {
			t.Errorf("template %s is empty", tmpl)
		}

		if !strings.Contains(string(content), "CANARY:") {
			t.Errorf("template %s missing CANARY token", tmpl)
		}

		// Verify templates reference shared command files (not duplicate content)
		if !strings.Contains(string(content), ".canary/commands/") {
			t.Errorf("template %s should reference .canary/commands/ files", tmpl)
		}
	}
}

// BenchmarkCreateCopilotInstructions measures performance of instruction file creation
// Target: <100ms per operation (well under the 2-second init overhead spec)
func BenchmarkCreateCopilotInstructions(b *testing.B) {
	// Create a temporary base directory for all benchmark runs
	tmpBase := b.TempDir()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Each iteration gets its own directory to avoid file conflicts
		tmpDir := filepath.Join(tmpBase, fmt.Sprintf("bench-%d", i))
		if err := os.MkdirAll(tmpDir, 0755); err != nil {
			b.Fatalf("failed to create benchmark directory: %v", err)
		}

		// Benchmark the instruction file creation
		if err := createCopilotInstructions(tmpDir, "BENCH"); err != nil {
			b.Fatalf("createCopilotInstructions failed: %v", err)
		}
	}
}

// BenchmarkCreateCopilotInstructionsReInit measures re-initialization performance
// This tests the existing file detection and skip logic
func BenchmarkCreateCopilotInstructionsReInit(b *testing.B) {
	tmpDir := b.TempDir()

	// Initial creation (not benchmarked)
	if err := createCopilotInstructions(tmpDir, "BENCH"); err != nil {
		b.Fatalf("initial creation failed: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Re-run on existing files (should skip them)
		if err := createCopilotInstructions(tmpDir, "BENCH"); err != nil {
			b.Fatalf("re-init failed: %v", err)
		}
	}
}
