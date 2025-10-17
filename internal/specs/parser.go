// Copyright (c) 2024 by Developer Network.
//
// For more details, see the LICENSE file in the root directory of this
// source code repository or contact Developer Network at info@devnw.com.


// CANARY: REQ=CBIN-134; FEATURE="SectionLoader"; ASPECT=Engine; STATUS=IMPL; UPDATED=2025-10-16
package specs

import (
	"fmt"
	"strings"
)

// ParseSections extracts specific sections from markdown content
// If sections is empty, returns full content
// Preserves metadata at top (lines before first ##)
// Section names are case-insensitive
func ParseSections(content string, sections []string) (string, error) {
	if len(sections) == 0 {
		return content, nil // Return full content
	}

	lines := strings.Split(content, "\n")
	var result strings.Builder
	var capturing bool
	var capturedAny bool

	// Always include metadata at top (lines before first ## header)
	for _, line := range lines {
		if strings.HasPrefix(line, "##") {
			break // Stop at first section header
		}
		result.WriteString(line + "\n")
	}

	// Process sections
	for _, line := range lines {
		if strings.HasPrefix(line, "## ") {
			// Extract section name from header
			section := strings.TrimPrefix(line, "## ")
			section = strings.ToLower(strings.TrimSpace(section))

			// Check if this section should be captured (case-insensitive match)
			capturing = false
			for _, requestedSection := range sections {
				requestedLower := strings.ToLower(requestedSection)
				// Match if section name contains the requested text
				// This allows "user stories" to match "## User Stories"
				if strings.Contains(section, requestedLower) || section == requestedLower {
					capturing = true
					capturedAny = true
					break
				}
			}
		}

		if capturing {
			result.WriteString(line + "\n")
		}
	}

	if !capturedAny {
		return "", fmt.Errorf("no matching sections found for: %v", sections)
	}

	return result.String(), nil
}

// ListSections returns all section headers from markdown content
// Extracts all ## level headers
func ListSections(content string) ([]string, error) {
	var sections []string
	lines := strings.Split(content, "\n")

	for _, line := range lines {
		if strings.HasPrefix(line, "## ") {
			section := strings.TrimPrefix(line, "## ")
			sections = append(sections, strings.TrimSpace(section))
		}
	}

	return sections, nil
}
