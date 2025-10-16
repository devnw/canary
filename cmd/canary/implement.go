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

// CANARY: REQ=CBIN-133; FEATURE="RequirementLookup"; ASPECT=API; STATUS=TESTED; TEST=TestCANARY_CBIN_133_CLI_ExactMatch; OWNER=canary; UPDATED=2025-10-16
package main

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"text/template"

	"go.spyder.org/canary/internal/matcher"
)

// RequirementSpec represents a loaded requirement specification
type RequirementSpec struct {
	ReqID       string
	FeatureName string
	SpecPath    string
	SpecContent string
	PlanContent string
}

// ImplementFlags holds command-line flags for implement command
type ImplementFlags struct {
	Prompt bool
	List   bool
}

// findRequirement locates a requirement spec by query (exact ID, fuzzy match, etc.)
func findRequirement(query string) (*RequirementSpec, error) {
	// Try exact ID match first
	spec, err := findByExactID(query)
	if err == nil {
		return spec, nil
	}

	// Fall back to fuzzy matching
	specsDir := filepath.Join(".canary", "specs")
	if _, err := os.Stat(specsDir); os.IsNotExist(err) {
		return nil, fmt.Errorf("specs directory not found: %s", specsDir)
	}

	matches, err := matcher.FindBestMatches(query, specsDir, 10)
	if err != nil {
		return nil, fmt.Errorf("fuzzy search failed: %w", err)
	}

	if len(matches) == 0 {
		return nil, fmt.Errorf("no matches found for query: %s", query)
	}

	// Auto-select if top match is strong and significantly better than others
	if len(matches) == 1 || (matches[0].Score > 80 && matches[0].Score-matches[1].Score > 20) {
		return loadSpecFromDir(matches[0].SpecPath)
	}

	// Interactive selection for ambiguous matches
	return selectInteractive(matches)
}

// findByExactID attempts to find a spec by exact requirement ID
func findByExactID(reqID string) (*RequirementSpec, error) {
	specsDir := filepath.Join(".canary", "specs")
	entries, err := os.ReadDir(specsDir)
	if err != nil {
		return nil, err
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		// Directory format: CBIN-XXX-FeatureName
		// Need to extract CBIN-XXX as reqID
		parts := strings.Split(entry.Name(), "-")
		if len(parts) < 3 {
			continue
		}

		// ReqID is first two parts: "CBIN" + "-" + "XXX"
		dirReqID := parts[0] + "-" + parts[1]
		if strings.EqualFold(dirReqID, reqID) {
			specPath := filepath.Join(specsDir, entry.Name())
			return loadSpecFromDir(specPath)
		}
	}

	return nil, fmt.Errorf("exact match not found for ID: %s", reqID)
}

// loadSpecFromDir loads a RequirementSpec from a spec directory
func loadSpecFromDir(dirPath string) (*RequirementSpec, error) {
	// Parse directory name for ReqID and FeatureName
	dirName := filepath.Base(dirPath)
	parts := strings.Split(dirName, "-")
	if len(parts) < 3 {
		return nil, fmt.Errorf("invalid spec directory name: %s", dirName)
	}

	// ReqID is first two parts: "CBIN" + "-" + "XXX"
	reqID := parts[0] + "-" + parts[1]
	// Feature name is everything after that
	featureName := strings.Join(parts[2:], "-")

	// Load spec.md
	specPath := filepath.Join(dirPath, "spec.md")
	specContent, err := os.ReadFile(specPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read spec.md: %w", err)
	}

	// Load plan.md (optional)
	planPath := filepath.Join(dirPath, "plan.md")
	planContent, _ := os.ReadFile(planPath) // Ignore error, plan is optional

	return &RequirementSpec{
		ReqID:       reqID,
		FeatureName: featureName,
		SpecPath:    dirPath,
		SpecContent: string(specContent),
		PlanContent: string(planContent),
	}, nil
}

// selectInteractive prompts the user to select from multiple matches
func selectInteractive(matches []matcher.Match) (*RequirementSpec, error) {
	fmt.Println("Multiple matches found:")
	fmt.Println()

	for i, match := range matches {
		fmt.Printf("%d. %s - %s (Score: %d%%)\n", i+1, match.ReqID, match.FeatureName, match.Score)
	}

	fmt.Println()
	fmt.Print("Select a requirement (1-", len(matches), "): ")

	reader := bufio.NewReader(os.Stdin)
	input, err := reader.ReadString('\n')
	if err != nil {
		return nil, fmt.Errorf("failed to read input: %w", err)
	}

	input = strings.TrimSpace(input)
	selection, err := strconv.Atoi(input)
	if err != nil || selection < 1 || selection > len(matches) {
		return nil, fmt.Errorf("invalid selection: %s", input)
	}

	selectedMatch := matches[selection-1]
	return loadSpecFromDir(selectedMatch.SpecPath)
}

// CANARY: REQ=CBIN-133; FEATURE="PromptGeneration"; ASPECT=API; STATUS=TESTED; TEST=TestCANARY_CBIN_133_CLI_PromptGeneration; OWNER=canary; UPDATED=2025-10-16

// ImplementPromptData holds template data for implementation prompt generation
type ImplementPromptData struct {
	ReqID        string
	FeatureName  string
	SpecContent  string
	PlanContent  string
	Checklist    string
	Progress     ProgressStats
	Constitution string
}

// ProgressStats holds implementation progress statistics
type ProgressStats struct {
	Total  int
	Tested int
	Impl   int
	Stub   int
}

// renderImplementPrompt generates the implementation guidance prompt
func renderImplementPrompt(spec *RequirementSpec, flags *ImplementFlags) (string, error) {
	// Load constitution
	constitutionPath := filepath.Join(".canary", "memory", "constitution.md")
	constitutionContent, err := os.ReadFile(constitutionPath)
	if err != nil {
		constitutionContent = []byte("") // Optional, continue without it
	}

	// Calculate progress
	progress, err := calculateProgress(spec.ReqID)
	if err != nil {
		// Non-fatal, use empty progress
		progress = ProgressStats{}
	}

	// Extract implementation checklist from spec
	checklist := extractImplementationChecklist(spec.SpecContent)

	// Load template
	templatePath := filepath.Join(".canary", "templates", "implement-prompt-template.md")
	templateContent, err := os.ReadFile(templatePath)
	if err != nil {
		return "", fmt.Errorf("failed to read template: %w", err)
	}

	// Prepare template data
	data := ImplementPromptData{
		ReqID:        spec.ReqID,
		FeatureName:  spec.FeatureName,
		SpecContent:  spec.SpecContent,
		PlanContent:  spec.PlanContent,
		Checklist:    checklist,
		Progress:     progress,
		Constitution: string(constitutionContent),
	}

	// Parse and execute template
	tmpl, err := template.New("implement-prompt").Parse(string(templateContent))
	if err != nil {
		return "", fmt.Errorf("failed to parse template: %w", err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("failed to execute template: %w", err)
	}

	return buf.String(), nil
}

// calculateProgress scans the codebase for CANARY tokens and counts by status
func calculateProgress(reqID string) (ProgressStats, error) {
	// Use grep to find all CANARY tokens for this requirement
	cmd := exec.Command("grep",
		"-rn",
		"--include=*.go",
		"--include=*.md",
		"--include=*.py",
		"--include=*.js",
		"--include=*.ts",
		"--include=*.java",
		"--include=*.rb",
		"--include=*.rs",
		fmt.Sprintf("CANARY:.*REQ=%s", reqID),
		".",
	)

	output, err := cmd.CombinedOutput()
	if err != nil && len(output) == 0 {
		// No matches found - not an error, just empty
		return ProgressStats{}, nil
	}

	stats := ProgressStats{}
	statusRegex := regexp.MustCompile(`STATUS=([A-Z]+)`)

	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if line == "" {
			continue
		}

		stats.Total++

		// Extract STATUS field
		matches := statusRegex.FindStringSubmatch(line)
		if len(matches) > 1 {
			status := matches[1]
			switch status {
			case "TESTED", "BENCHED":
				stats.Tested++
			case "IMPL":
				stats.Impl++
			case "STUB":
				stats.Stub++
			}
		}
	}

	return stats, nil
}

// extractImplementationChecklist extracts the Implementation Checklist section from spec
func extractImplementationChecklist(specContent string) string {
	lines := strings.Split(specContent, "\n")
	var checklist strings.Builder
	inChecklist := false

	for _, line := range lines {
		// Start capturing at "## Implementation Checklist"
		if strings.HasPrefix(line, "## Implementation Checklist") {
			inChecklist = true
			checklist.WriteString(line)
			checklist.WriteString("\n")
			continue
		}

		// Stop at next ## heading
		if inChecklist && strings.HasPrefix(line, "## ") && !strings.HasPrefix(line, "## Implementation Checklist") {
			break
		}

		if inChecklist {
			checklist.WriteString(line)
			checklist.WriteString("\n")
		}
	}

	result := checklist.String()
	if result == "" {
		return "No implementation checklist found in specification."
	}

	return result
}

// listUnimplemented lists all requirements with incomplete implementation
func listUnimplemented() error {
	specsDir := filepath.Join(".canary", "specs")
	entries, err := os.ReadDir(specsDir)
	if err != nil {
		return fmt.Errorf("failed to read specs directory: %w", err)
	}

	fmt.Println("Requirements with incomplete implementation:")
	fmt.Println()

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		// Parse directory name
		parts := strings.SplitN(entry.Name(), "-", 2)
		if len(parts) < 2 {
			continue
		}

		reqID := parts[0]
		featureName := parts[1]

		// Calculate progress
		progress, err := calculateProgress(reqID)
		if err != nil {
			continue // Skip on error
		}

		// Show if has stub or impl tokens (not fully tested)
		if progress.Stub > 0 || progress.Impl > 0 {
			completionRate := 0
			if progress.Total > 0 {
				completionRate = (progress.Tested * 100) / progress.Total
			}

			fmt.Printf("%s - %s [%d%% complete]\n", reqID, featureName, completionRate)
			fmt.Printf("  Total: %d | Tested: %d | Impl: %d | Stub: %d\n",
				progress.Total, progress.Tested, progress.Impl, progress.Stub)
			fmt.Println()
		}
	}

	return nil
}
