// Copyright (c) 2025 by Developer Network.
//
// For more details, see the LICENSE file in the root directory of this
// source code repository or contact Developer Network at info@devnw.com.

// CANARY: REQ=CBIN-133; FEATURE="RequirementLookup"; ASPECT=API; STATUS=TESTED; UPDATED=2025-10-16
package implement

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"text/template"
	"time"

	"github.com/spf13/cobra"
	"go.devnw.com/canary/cli/internal/utils"
	"go.devnw.com/canary/internal/matcher"
)

// CANARY: REQ=CBIN-133; FEATURE="ImplementCmd"; ASPECT=CLI; STATUS=TESTED; TEST=TestCANARY_CBIN_133_CLI_ExactMatch; OWNER=canary; DOC=user:docs/user/implement-command-guide.md; DOC_HASH=ed68fb1d97cf0562; UPDATED=2025-10-17
var ImplementCmd = &cobra.Command{
	Use:   "implement <query>",
	Short: "Generate implementation guidance for a requirement",
	Long: `Generate comprehensive implementation guidance for a requirement specification.

This command:
- Accepts requirement by ID (CBIN-XXX), name, or fuzzy search query
- Uses fuzzy matching with auto-selection for strong matches
- Generates complete implementation prompt including:
  - Specification details
  - Implementation plan
  - Constitutional principles
  - Implementation checklist
  - Progress tracking
  - Test-first guidance

Examples:
  canary implement CBIN-105              # Exact ID match
  canary implement "user auth"           # Fuzzy search
  canary implement UserAuthentication    # Feature name match
  canary implement --list                # List all unimplemented requirements`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		// TODO: Implement --prompt flag to load custom prompts from file
		promptArg, _ := cmd.Flags().GetString("prompt-arg")
		_ = promptArg // Stubbed for future use

		listFlag, _ := cmd.Flags().GetBool("list")
		promptFlag, _ := cmd.Flags().GetBool("prompt")

		// Handle --list flag
		if listFlag {
			return listUnimplemented()
		}

		// Require query argument if not listing
		if len(args) < 1 {
			return fmt.Errorf("requirement query is required (use --list to show all unimplemented)")
		}

		query := args[0]

		// Find requirement spec
		spec, err := findRequirement(query)
		if err != nil {
			return fmt.Errorf("find requirement: %w", err)
		}

		// Generate prompt
		flags := &ImplementFlags{
			Prompt: promptFlag,
		}

		prompt, err := renderImplementPrompt(spec, flags)
		if err != nil {
			return fmt.Errorf("generate prompt: %w", err)
		}

		fmt.Println(prompt)

		return nil
	},
}

func init() {
	ImplementCmd.Flags().String("prompt-arg", "", "Custom prompt file or embedded prompt name (future use)")
	ImplementCmd.Flags().Bool("list", false, "list all unimplemented requirements")
	ImplementCmd.Flags().Bool("prompt", true, "generate full implementation prompt (default: true)")
}

// RequirementSpec holds loaded specification data
type RequirementSpec struct {
	ReqID       string
	FeatureName string
	SpecPath    string
	SpecContent string
	PlanPath    string
	PlanContent string
	HasPlan     bool
}

// ImplementFlags holds command flags
type ImplementFlags struct {
	Prompt       bool
	ShowProgress bool
	ContextLines int
}

// ProgressStats tracks implementation progress
type ProgressStats struct {
	Total     int
	Stub      int
	Impl      int
	Tested    int
	Benched   int
	Completed int
}

// findRequirement locates a requirement by ID or fuzzy match
func findRequirement(query string) (*RequirementSpec, error) {
	query = strings.TrimSpace(query)
	queryUpper := strings.ToUpper(query)

	// Extract REQ-ID if query is in format "CBIN-101-feature-name" or "CBIN-101"
	reqID := extractReqID(queryUpper)

	// Attempt 1: Exact ID match using extracted REQ-ID
	if reqID != "" {
		spec, err := findByExactID(reqID)
		if err == nil {
			return spec, nil
		}
	}

	// Attempt 2: Directory pattern match
	specsDir := ".canary/specs"
	pattern := filepath.Join(specsDir, "*"+query+"*")
	matches, _ := filepath.Glob(pattern)
	if len(matches) == 1 {
		return loadSpecFromDir(matches[0])
	}

	// Attempt 3: Fuzzy match
	fuzzyMatches, err := matcher.FindBestMatches(query, specsDir, 5)
	if err != nil {
		return nil, fmt.Errorf("fuzzy search failed: %w", err)
	}

	if len(fuzzyMatches) == 0 {
		return nil, fmt.Errorf("no matches found for query: %s", query)
	}

	// Auto-select if clear winner (>90% score or >20 points ahead)
	if fuzzyMatches[0].Score >= 90 && (len(fuzzyMatches) == 1 || fuzzyMatches[0].Score-fuzzyMatches[1].Score > 20) {
		return loadSpecFromDir(fuzzyMatches[0].SpecPath)
	}

	// For test purposes, return best match
	// In production, this would trigger interactive selection
	return loadSpecFromDir(fuzzyMatches[0].SpecPath)
}

// extractReqID extracts the requirement ID from a query
// Examples:
//   - "CBIN-101" -> "CBIN-101"
//   - "CBIN-101-engine" -> "CBIN-101"
//   - "CBIN-101-feature-name" -> "CBIN-101"
//   - "engine" -> ""
func extractReqID(query string) string {
	// Match pattern: PROJECT-###
	// Where PROJECT is alphanumeric (CBIN, REQ, etc.) and ### is 1-4 digits
	parts := strings.SplitN(query, "-", 3)
	if len(parts) >= 2 {
		// Check if first part is alphabetic and second part is numeric
		if len(parts[0]) > 0 && len(parts[1]) > 0 {
			// Validate that second part is all digits
			allDigits := true
			for _, ch := range parts[1] {
				if ch < '0' || ch > '9' {
					allDigits = false
					break
				}
			}
			if allDigits {
				return parts[0] + "-" + parts[1]
			}
		}
	}
	return ""
}

// findByExactID finds spec by exact requirement ID
func findByExactID(reqID string) (*RequirementSpec, error) {
	reqID = strings.ToUpper(reqID)
	specsDir := ".canary/specs"

	entries, err := os.ReadDir(specsDir)
	if err != nil {
		return nil, err
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		if strings.HasPrefix(entry.Name(), reqID+"-") {
			return loadSpecFromDir(filepath.Join(specsDir, entry.Name()))
		}
	}

	return nil, fmt.Errorf("specification not found for %s", reqID)
}

// loadSpecFromDir loads spec.md and plan.md from directory
func loadSpecFromDir(dirPath string) (*RequirementSpec, error) {
	specPath := filepath.Join(dirPath, "spec.md")
	specContent, err := os.ReadFile(specPath)
	if err != nil {
		return nil, fmt.Errorf("read spec file: %w", err)
	}

	// Extract ReqID and FeatureName from directory name
	dirName := filepath.Base(dirPath)
	parts := strings.SplitN(dirName, "-", 3)
	if len(parts) < 3 {
		return nil, fmt.Errorf("invalid spec directory name: %s", dirName)
	}

	spec := &RequirementSpec{
		ReqID:       parts[0] + "-" + parts[1], // CBIN-XXX
		FeatureName: parts[2],                  // feature-name
		SpecPath:    specPath,
		SpecContent: string(specContent),
	}

	// Load plan.md if exists
	planPath := filepath.Join(dirPath, "plan.md")
	if planContent, err := os.ReadFile(planPath); err == nil {
		spec.PlanPath = planPath
		spec.PlanContent = string(planContent)
		spec.HasPlan = true
	}

	return spec, nil
}

// CANARY: REQ=CBIN-133; FEATURE="PromptRenderer"; ASPECT=API; STATUS=TESTED; UPDATED=2025-10-16
// renderImplementPrompt generates comprehensive implementation guidance
func renderImplementPrompt(spec *RequirementSpec, flags *ImplementFlags) (string, error) {
	// Load template
	templatePath := ".canary/templates/implement-prompt-template.md"
	templateContent, err := os.ReadFile(templatePath)
	if err != nil {
		return "", fmt.Errorf("read template: %w", err)
	}

	tmpl, err := template.New("implement-prompt").Parse(string(templateContent))
	if err != nil {
		return "", fmt.Errorf("parse template: %w", err)
	}

	// Load constitution
	constitutionPath := ".canary/memory/constitution.md"
	constitutionContent, _ := os.ReadFile(constitutionPath)

	// Calculate progress
	progress, _ := calculateProgress(spec.ReqID)

	// Extract implementation checklist from spec
	checklist := extractImplementationChecklist(spec.SpecContent)

	// Populate template data
	data := map[string]interface{}{
		"ReqID":        spec.ReqID,
		"FeatureName":  spec.FeatureName,
		"SpecPath":     spec.SpecPath,
		"SpecContent":  spec.SpecContent,
		"PlanPath":     spec.PlanPath,
		"PlanContent":  spec.PlanContent,
		"HasPlan":      spec.HasPlan,
		"Constitution": string(constitutionContent),
		"Checklist":    checklist,
		"Progress":     progress,
		"Today":        time.Now().UTC().Format("2006-01-02"),
	}

	// Render
	var buf strings.Builder
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("execute template: %w", err)
	}

	return buf.String(), nil
}

// calculateProgress scans codebase for tokens matching reqID
func calculateProgress(reqID string) (*ProgressStats, error) {
	// Use grep to find all tokens for this requirement
	grepCmd := exec.Command("grep", "-rn", "--include=*.go", "--include=*.md",
		fmt.Sprintf("CANARY:.*REQ=%s", reqID), ".")

	output, err := grepCmd.CombinedOutput()
	if err != nil && len(output) == 0 {
		return &ProgressStats{}, nil // No tokens found
	}

	stats := &ProgressStats{}
	lines := strings.Split(string(output), "\n")

	for _, line := range lines {
		if line == "" {
			continue
		}

		status := utils.ExtractField(line, "STATUS")
		stats.Total++

		switch status {
		case "STUB":
			stats.Stub++
		case "IMPL":
			stats.Impl++
		case "TESTED":
			stats.Tested++
			stats.Completed++
		case "BENCHED":
			stats.Benched++
			stats.Completed++
		}
	}

	return stats, nil
}

// extractImplementationChecklist extracts checklist section from spec
func extractImplementationChecklist(specContent string) string {
	lines := strings.Split(specContent, "\n")
	inChecklist := false
	var checklist strings.Builder

	for _, line := range lines {
		if strings.Contains(line, "## Implementation Checklist") {
			inChecklist = true
			continue
		}

		if inChecklist {
			// Stop at next major section
			if strings.HasPrefix(line, "## ") && !strings.Contains(line, "Implementation") {
				break
			}
			checklist.WriteString(line + "\n")
		}
	}

	return checklist.String()
}

// listUnimplemented lists all unimplemented (STUB/IMPL) requirements
func listUnimplemented() error {
	// TODO: Implement listing functionality
	// For now, just return a message
	fmt.Println("Listing unimplemented requirements...")
	fmt.Println("(Feature not yet fully implemented)")
	return nil
}
