package specs

import (
	"strings"
	"testing"
)

// CANARY: REQ=CBIN-134; FEATURE="SectionLoader"; ASPECT=Engine; STATUS=STUB; TEST=TestCANARY_CBIN_134_Engine_SectionParser; UPDATED=2025-10-16

func TestCANARY_CBIN_134_Engine_SectionParser(t *testing.T) {
	// Test ParseSections extracts specific sections
	// Expected to FAIL initially (parser doesn't exist)

	testMarkdown := `# CANARY: REQ=CBIN-999; FEATURE="Test"; ASPECT=Docs; STATUS=STUB; UPDATED=2025-10-16
# Test Specification

**Requirement ID:** CBIN-999
**Status:** STUB

## Overview

This is the overview section with some content.

## User Stories

**US-1: First Story**
Some user story content here.

**US-2: Second Story**
More user story content.

## Functional Requirements

### FR-1: First Requirement
Requirement description here.

### FR-2: Second Requirement
Another requirement.

## Success Criteria

- Criterion 1
- Criterion 2

## Dependencies

- Dependency 1
- Dependency 2
`

	tests := []struct {
		name         string
		content      string
		sections     []string
		wantContains []string
		wantErr      bool
	}{
		{
			name:         "extract single section - overview",
			content:      testMarkdown,
			sections:     []string{"overview"},
			wantContains: []string{"## Overview", "overview section with some content"},
			wantErr:      false,
		},
		{
			name:         "extract multiple sections",
			content:      testMarkdown,
			sections:     []string{"overview", "user stories"},
			wantContains: []string{"## Overview", "## User Stories", "US-1: First Story"},
			wantErr:      false,
		},
		{
			name:         "extract all (empty sections list)",
			content:      testMarkdown,
			sections:     []string{},
			wantContains: []string{"## Overview", "## User Stories", "## Functional Requirements"},
			wantErr:      false,
		},
		{
			name:     "invalid section name",
			content:  testMarkdown,
			sections: []string{"nonexistent"},
			wantErr:  true,
		},
		{
			name:         "case insensitive matching",
			content:      testMarkdown,
			sections:     []string{"OVERVIEW", "User Stories"},
			wantContains: []string{"## Overview", "## User Stories"},
			wantErr:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// This will fail because ParseSections doesn't exist yet
			result, err := ParseSections(tt.content, tt.sections)

			if (err != nil) != tt.wantErr {
				t.Errorf("ParseSections() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				// Verify metadata is preserved (first lines before ##)
				if !strings.Contains(result, "CBIN-999") {
					t.Error("ParseSections() should preserve metadata at top")
				}

				// Verify requested sections are present
				for _, want := range tt.wantContains {
					if !strings.Contains(result, want) {
						t.Errorf("ParseSections() result missing: %q", want)
					}
				}

				// For specific section requests, verify excluded sections are NOT present
				if len(tt.sections) > 0 && !tt.wantErr {
					// Dependencies section should not be in result if not requested
					if !sliceContains(tt.sections, "dependencies") && !sliceContains(tt.sections, "Dependencies") {
						if strings.Contains(result, "## Dependencies") {
							t.Error("ParseSections() included unrequested section: Dependencies")
						}
					}
				}
			}
		})
	}
}

func TestCANARY_CBIN_134_Engine_ListSections(t *testing.T) {
	// Test ListSections returns all section headers
	// Expected to FAIL initially (function doesn't exist)

	testMarkdown := `# Test Specification

## Overview
Content here.

## User Stories
More content.

## Functional Requirements
Requirements here.

## Success Criteria
Criteria here.
`

	tests := []struct {
		name         string
		content      string
		wantSections []string
		wantErr      bool
	}{
		{
			name:    "list all sections",
			content: testMarkdown,
			wantSections: []string{
				"Overview",
				"User Stories",
				"Functional Requirements",
				"Success Criteria",
			},
			wantErr: false,
		},
		{
			name:         "empty content",
			content:      "",
			wantSections: []string{},
			wantErr:      false,
		},
		{
			name:         "no sections",
			content:      "Just some text without headers",
			wantSections: []string{},
			wantErr:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// This will fail because ListSections doesn't exist yet
			sections, err := ListSections(tt.content)

			if (err != nil) != tt.wantErr {
				t.Errorf("ListSections() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if len(sections) != len(tt.wantSections) {
					t.Errorf("ListSections() got %d sections, want %d", len(sections), len(tt.wantSections))
				}

				// Verify each expected section is present
				for _, want := range tt.wantSections {
					found := false
					for _, got := range sections {
						if got == want {
							found = true
							break
						}
					}
					if !found {
						t.Errorf("ListSections() missing section: %q", want)
					}
				}
			}
		})
	}
}

func TestCANARY_CBIN_134_Engine_SectionParserContextReduction(t *testing.T) {
	// Test that section-specific loading reduces content size significantly

	largeMarkdown := `# Test Specification

## Overview
` + strings.Repeat("Lorem ipsum dolor sit amet. ", 100) + `

## User Stories
` + strings.Repeat("User story content here. ", 100) + `

## Functional Requirements
` + strings.Repeat("Functional requirement text. ", 100) + `

## Success Criteria
` + strings.Repeat("Success criteria content. ", 100) + `

## Dependencies
` + strings.Repeat("Dependency information. ", 100) + `
`

	// This will fail because ParseSections doesn't exist yet
	fullContent, err := ParseSections(largeMarkdown, []string{})
	if err != nil {
		t.Fatalf("ParseSections() with empty sections failed: %v", err)
	}

	// Extract only overview section
	overviewOnly, err := ParseSections(largeMarkdown, []string{"overview"})
	if err != nil {
		t.Fatalf("ParseSections() with overview failed: %v", err)
	}

	// Calculate reduction percentage
	fullSize := len(fullContent)
	reducedSize := len(overviewOnly)
	reductionPct := float64(fullSize-reducedSize) / float64(fullSize) * 100

	// Per spec: Section-specific loading should reduce context by 50-80%
	if reductionPct < 50 {
		t.Errorf("Context reduction %0.1f%% is less than expected 50%%", reductionPct)
	}

	t.Logf("Context reduction: %0.1f%% (full: %d bytes, reduced: %d bytes)", reductionPct, fullSize, reducedSize)
}

// Helper to check if slice contains string (case-insensitive)
func sliceContains(slice []string, str string) bool {
	lowerStr := strings.ToLower(str)
	for _, s := range slice {
		if strings.ToLower(s) == lowerStr {
			return true
		}
	}
	return false
}
