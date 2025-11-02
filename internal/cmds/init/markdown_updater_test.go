package init

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestUpdateMarkdownSection(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.md")

	// Test 1: Create new file with gated content
	t.Run("CreateNewFile", func(t *testing.T) {
		content := "## CANARY Section\n\nThis is the CANARY content."
		err := updateMarkdownSection(testFile, content)
		if err != nil {
			t.Fatalf("updateMarkdownSection failed: %v", err)
		}

		result, err := os.ReadFile(testFile)
		if err != nil {
			t.Fatalf("read file failed: %v", err)
		}

		resultStr := string(result)
		if !strings.Contains(resultStr, "<!-- CANARY:START -->") {
			t.Error("Missing start marker")
		}
		if !strings.Contains(resultStr, "<!-- CANARY:END -->") {
			t.Error("Missing end marker")
		}
		if !strings.Contains(resultStr, "This is the CANARY content") {
			t.Error("Missing content")
		}
	})

	// Test 2: Update existing file with user content
	t.Run("PreserveUserContent", func(t *testing.T) {
		// First, write a file with user content and CANARY section
		initialContent := `# My Project

This is my custom content that should be preserved.

<!-- CANARY:START -->
Old CANARY content here
<!-- CANARY:END -->

More user content below.
`
		if err := os.WriteFile(testFile, []byte(initialContent), 0644); err != nil {
			t.Fatalf("write initial file failed: %v", err)
		}

		// Now update the CANARY section
		newContent := "## Updated CANARY Section\n\nThis is NEW CANARY content."
		err := updateMarkdownSection(testFile, newContent)
		if err != nil {
			t.Fatalf("updateMarkdownSection failed: %v", err)
		}

		result, err := os.ReadFile(testFile)
		if err != nil {
			t.Fatalf("read file failed: %v", err)
		}

		resultStr := string(result)

		// Check that user content is preserved
		if !strings.Contains(resultStr, "This is my custom content that should be preserved") {
			t.Error("User content above CANARY section was lost")
		}
		if !strings.Contains(resultStr, "More user content below") {
			t.Error("User content below CANARY section was lost")
		}

		// Check that CANARY content was updated
		if strings.Contains(resultStr, "Old CANARY content here") {
			t.Error("Old CANARY content was not replaced")
		}
		if !strings.Contains(resultStr, "This is NEW CANARY content") {
			t.Error("New CANARY content was not added")
		}
	})

	// Test 3: Add CANARY section to existing file without markers
	t.Run("AddSectionToExistingFile", func(t *testing.T) {
		// Write a file without CANARY section
		existingContent := `# Existing Project

Some existing content here.

## Another Section

More content.
`
		if err := os.WriteFile(testFile, []byte(existingContent), 0644); err != nil {
			t.Fatalf("write existing file failed: %v", err)
		}

		// Add CANARY section
		canaryContent := "## CANARY Integration\n\nCANARY tracking enabled."
		err := updateMarkdownSection(testFile, canaryContent)
		if err != nil {
			t.Fatalf("updateMarkdownSection failed: %v", err)
		}

		result, err := os.ReadFile(testFile)
		if err != nil {
			t.Fatalf("read file failed: %v", err)
		}

		resultStr := string(result)

		// Check that existing content is preserved
		if !strings.Contains(resultStr, "Some existing content here") {
			t.Error("Existing content was lost")
		}

		// Check that CANARY section was appended
		if !strings.Contains(resultStr, "CANARY tracking enabled") {
			t.Error("CANARY section was not added")
		}
		if !strings.Contains(resultStr, "<!-- CANARY:START -->") {
			t.Error("CANARY markers were not added")
		}
	})

	// Test 4: Multiple updates should be idempotent
	t.Run("IdempotentUpdates", func(t *testing.T) {
		content := "Test content v1"
		_ = updateMarkdownSection(testFile, content)

		result1, _ := os.ReadFile(testFile)

		// Update again with same content
		_ = updateMarkdownSection(testFile, content)
		result2, _ := os.ReadFile(testFile)

		// Should have same result
		if string(result1) != string(result2) {
			t.Error("Multiple updates with same content produced different results")
		}

		// Count markers - should only have one pair
		markerCount := strings.Count(string(result2), "<!-- CANARY:START -->")
		if markerCount != 1 {
			t.Errorf("Expected 1 start marker, got %d", markerCount)
		}
	})
}

func TestUpdateMultipleMarkdownSections(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "multi-test.md")

	t.Run("CreateMultipleSections", func(t *testing.T) {
		sections := map[string]string{
			"intro":    "# Introduction\n\nThis is the intro.",
			"features": "## Features\n\n- Feature 1\n- Feature 2",
			"setup":    "## Setup\n\nInstallation steps here.",
		}

		err := updateMultipleMarkdownSections(testFile, sections)
		if err != nil {
			t.Fatalf("updateMultipleMarkdownSections failed: %v", err)
		}

		result, err := os.ReadFile(testFile)
		if err != nil {
			t.Fatalf("read file failed: %v", err)
		}

		resultStr := string(result)

		// Check all sections are present
		for key := range sections {
			if !strings.Contains(resultStr, "<!-- CANARY:"+key+":START -->") {
				t.Errorf("Missing start marker for %s", key)
			}
			if !strings.Contains(resultStr, "<!-- CANARY:"+key+":END -->") {
				t.Errorf("Missing end marker for %s", key)
			}
			// Check for some content from each section
			if key == "intro" && !strings.Contains(resultStr, "This is the intro") {
				t.Error("Missing intro content")
			}
		}
	})

	t.Run("UpdateOneSectionPreserveOthers", func(t *testing.T) {
		// First create multiple sections
		sections := map[string]string{
			"section1": "Content 1",
			"section2": "Content 2",
			"section3": "Content 3",
		}
		_ = updateMultipleMarkdownSections(testFile, sections)

		// Now update just one section
		updatedSections := map[string]string{
			"section2": "Updated Content 2",
		}
		err := updateMultipleMarkdownSections(testFile, updatedSections)
		if err != nil {
			t.Fatalf("update failed: %v", err)
		}

		result, _ := os.ReadFile(testFile)
		resultStr := string(result)

		// section1 and section3 should still be there
		if !strings.Contains(resultStr, "Content 1") {
			t.Error("Section 1 was lost")
		}
		if !strings.Contains(resultStr, "Content 3") {
			t.Error("Section 3 was lost")
		}

		// section2 should be updated
		if strings.Contains(resultStr, "Content 2") && !strings.Contains(resultStr, "Updated Content 2") {
			t.Error("Section 2 was not updated")
		}
		if !strings.Contains(resultStr, "Updated Content 2") {
			t.Error("Section 2 update was not applied")
		}
	})
}

func TestBuildMarkdownGatedBodyHelpers(t *testing.T) {
	unnamed := buildMarkdownGatedBody("Body Line")
	if !strings.Contains(unnamed, "<!-- CANARY:START -->") || !strings.Contains(unnamed, "<!-- CANARY:END -->") {
		t.Errorf("unnamed gated body missing markers: %s", unnamed)
	}
	if !strings.Contains(unnamed, "Body Line") {
		t.Errorf("unnamed gated body missing content")
	}
	keyed := buildMarkdownGatedBodyKey("intro", "Intro Line")
	if !strings.Contains(keyed, "<!-- CANARY:intro:START -->") || !strings.Contains(keyed, "<!-- CANARY:intro:END -->") {
		t.Errorf("keyed gated body missing markers: %s", keyed)
	}
	if !strings.Contains(keyed, "Intro Line") {
		t.Errorf("keyed gated body missing content")
	}
	// Ensure trailing newline
	if !strings.HasSuffix(unnamed, "\n") || !strings.HasSuffix(keyed, "\n") {
		t.Errorf("expected trailing newline in gated body snippets")
	}
}
