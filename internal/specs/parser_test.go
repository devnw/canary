// Copyright (c) 2025 by Developer Network.
//
// For more details, see the LICENSE file in the root directory of this
// source code repository or contact Developer Network at info@devnw.com.

// CANARY: REQ=CBIN-134; FEATURE="SectionLoader"; ASPECT=Engine; STATUS=TESTED; TEST=TestCANARY_CBIN_134_Engine_SectionParser; UPDATED=2025-10-17
package specs

import (
	"strings"
	"testing"
)

// TestCANARY_CBIN_134_Engine_SectionParser verifies ParseSections extracts specific sections
func TestCANARY_CBIN_134_Engine_SectionParser(t *testing.T) {
	testContent := `# Feature Spec

## Overview
This is the overview section.

## Requirements  
These are requirements.

## Success Criteria
These are success criteria.
`

	tests := []struct {
		name     string
		sections []string
		want     string
	}{
		{
			name:     "single section",
			sections: []string{"overview"},
			want:     "## Overview",
		},
		{
			name:     "multiple sections",
			sections: []string{"requirements", "success"},
			want:     "## Requirements",
		},
		{
			name:     "no sections (full content)",
			sections: []string{},
			want:     "# Feature Spec",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Execute: Parse sections
			result, err := ParseSections(testContent, tt.sections)
			if err != nil {
				t.Fatalf("ParseSections failed: %v", err)
			}

			// Verify: Contains expected section
			if !strings.Contains(result, tt.want) {
				t.Errorf("result does not contain %q", tt.want)
			}
		})
	}
}

// TestCANARY_CBIN_134_Engine_ListSections verifies ListSections returns all section headers
func TestCANARY_CBIN_134_Engine_ListSections(t *testing.T) {
	content := `# Title

## Section 1
Content

## Section 2
More content
`

	// Execute: List sections
	sections, err := ListSections(content)
	if err != nil {
		t.Fatalf("ListSections failed: %v", err)
	}

	// Verify: Found both sections
	if len(sections) < 2 {
		t.Errorf("expected at least 2 sections, got %d", len(sections))
	}

	t.Logf("Found sections: %v", sections)
}
