package init

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"strings"
)

// CANARY: REQ=CBIN-149; FEATURE="MarkdownSectionUpdater"; ASPECT=CLI; STATUS=IMPL; OWNER=canary; UPDATED=2025-11-01
// updateMarkdownSection updates or inserts a gated section in a markdown file
// The section is marked with HTML comments: <!-- CANARY:START --> ... <!-- CANARY:END -->
// If the file doesn't exist, it creates it with the content
// If the section exists, it replaces the content between the markers
// If the markers don't exist, it appends the section to the end
func updateMarkdownSection(filePath, sectionContent string) error {
	startMarker := "<!-- CANARY:START -->"
	endMarker := "<!-- CANARY:END -->"

	// Check if file exists
	existingContent, err := os.ReadFile(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			// File doesn't exist, create it with gated content
			content := fmt.Sprintf("%s\n%s\n%s\n", startMarker, sectionContent, endMarker)
			return os.WriteFile(filePath, []byte(content), 0644)
		}
		return fmt.Errorf("read file: %w", err)
	}

	// Parse existing file to find CANARY section
	lines := strings.Split(string(existingContent), "\n")
	var result []string
	inCanarySection := false
	foundCanarySection := false

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		if trimmed == startMarker {
			// Found start marker, enter CANARY section
			inCanarySection = true
			foundCanarySection = true
			result = append(result, line)
			// Add new content
			result = append(result, sectionContent)
			continue
		}

		if trimmed == endMarker {
			// Found end marker, exit CANARY section
			inCanarySection = false
			result = append(result, line)
			continue
		}

		// If we're in the CANARY section, skip old content (it's being replaced)
		if inCanarySection {
			continue
		}

		// Keep all non-CANARY content
		result = append(result, line)
	}

	// If no CANARY section was found, append it to the end
	if !foundCanarySection {
		// Ensure there's a blank line before the section
		if len(result) > 0 && result[len(result)-1] != "" {
			result = append(result, "")
		}
		result = append(result, startMarker)
		result = append(result, sectionContent)
		result = append(result, endMarker)
	}

	// Write updated content
	updated := strings.Join(result, "\n")
	return os.WriteFile(filePath, []byte(updated), 0644)
}

// CANARY: REQ=CBIN-149; FEATURE="MarkdownSectionUpdater"; ASPECT=CLI; STATUS=IMPL; OWNER=canary; UPDATED=2025-11-01
// updateMultipleMarkdownSections updates multiple gated sections in a markdown file
// Each section is identified by a unique key: <!-- CANARY:key:START --> ... <!-- CANARY:key:END -->
func updateMultipleMarkdownSections(filePath string, sections map[string]string) error {
	// Check if file exists
	existingContent, err := os.ReadFile(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			// File doesn't exist, create it with all gated sections
			var buf bytes.Buffer
			for key, content := range sections {
				buf.WriteString(fmt.Sprintf("<!-- CANARY:%s:START -->\n", key))
				buf.WriteString(content)
				buf.WriteString(fmt.Sprintf("\n<!-- CANARY:%s:END -->\n\n", key))
			}
			return os.WriteFile(filePath, buf.Bytes(), 0644)
		}
		return fmt.Errorf("read file: %w", err)
	}

	// Parse existing file to find all CANARY sections
	scanner := bufio.NewScanner(bytes.NewReader(existingContent))
	var result []string
	inCanarySection := ""
	foundSections := make(map[string]bool)

	for scanner.Scan() {
		line := scanner.Text()
		trimmed := strings.TrimSpace(line)

		// Check for section start marker
		if strings.HasPrefix(trimmed, "<!-- CANARY:") && strings.HasSuffix(trimmed, ":START -->") {
			// Extract section key
			key := strings.TrimPrefix(trimmed, "<!-- CANARY:")
			key = strings.TrimSuffix(key, ":START -->")

			inCanarySection = key
			foundSections[key] = true
			result = append(result, line)

			// Add content for this section
			// If we have new content, use it; otherwise preserve old content
			if content, ok := sections[key]; ok {
				result = append(result, content)
			}
			// If no new content provided, we'll preserve the old content (don't skip lines below)
			continue
		}

		// Check for section end marker
		if strings.HasPrefix(trimmed, "<!-- CANARY:") && strings.HasSuffix(trimmed, ":END -->") {
			// Extract section key to see if we're updating this section
			key := strings.TrimPrefix(trimmed, "<!-- CANARY:")
			key = strings.TrimSuffix(key, ":END -->")

			// If we provided new content for this section, skip the old content
			// Otherwise, the old content has already been added above
			inCanarySection = ""
			result = append(result, line)
			continue
		}

		// If we're in a CANARY section that we're updating, skip old content
		if inCanarySection != "" {
			if _, shouldUpdate := sections[inCanarySection]; shouldUpdate {
				// Skip old content - we already added new content
				continue
			}
			// Not updating this section, preserve its content
			result = append(result, line)
			continue
		}

		// Keep all non-CANARY content
		result = append(result, line)
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("scan file: %w", err)
	}

	// Append any sections that weren't found in the file
	for key, content := range sections {
		if !foundSections[key] {
			// Ensure there's a blank line before the section
			if len(result) > 0 && result[len(result)-1] != "" {
				result = append(result, "")
			}
			result = append(result, fmt.Sprintf("<!-- CANARY:%s:START -->", key))
			result = append(result, content)
			result = append(result, fmt.Sprintf("<!-- CANARY:%s:END -->", key))
		}
	}

	// Write updated content
	updated := strings.Join(result, "\n")
	// Ensure file ends with newline
	if !strings.HasSuffix(updated, "\n") {
		updated += "\n"
	}
	return os.WriteFile(filePath, []byte(updated), 0644)
}

// CANARY: REQ=CBIN-149; FEATURE="MarkdownSectionUpdater"; ASPECT=CLI; STATUS=IMPL; OWNER=canary; UPDATED=2025-11-01
// removeMarkdownSection removes a gated section from a markdown file
func removeMarkdownSection(filePath, sectionKey string) error {
	existingContent, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("read file: %w", err)
	}

	startMarker := fmt.Sprintf("<!-- CANARY:%s:START -->", sectionKey)
	endMarker := fmt.Sprintf("<!-- CANARY:%s:END -->", sectionKey)

	scanner := bufio.NewScanner(bytes.NewReader(existingContent))
	var result []string
	inSection := false

	for scanner.Scan() {
		line := scanner.Text()
		trimmed := strings.TrimSpace(line)

		if trimmed == startMarker {
			inSection = true
			continue
		}

		if trimmed == endMarker {
			inSection = false
			continue
		}

		if !inSection {
			result = append(result, line)
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("scan file: %w", err)
	}

	updated := strings.Join(result, "\n")
	if !strings.HasSuffix(updated, "\n") {
		updated += "\n"
	}
	return os.WriteFile(filePath, []byte(updated), 0644)
}
