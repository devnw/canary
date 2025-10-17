// Copyright (c) 2025 by Developer Network.
//
// For more details, see the LICENSE file in the root directory of this
// source code repository or contact Developer Network at info@devnw.com.

// CANARY: REQ=CBIN-132; FEATURE="NextPriorityCommand"; ASPECT=CLI; STATUS=BENCHED; TEST=TestCANARY_CBIN_132_CLI_NextPrioritySelection; BENCH=BenchmarkCANARY_CBIN_132_CLI_PriorityQuery; OWNER=canary; UPDATED=2025-10-16
package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"text/template"
	"time"

	"go.spyder.org/canary/internal/config"
	"go.spyder.org/canary/internal/storage"
)

// PromptData holds template variables for prompt generation
type PromptData struct {
	ReqID             string
	Feature           string
	Aspect            string
	Status            string
	Priority          int
	SpecFile          string
	SpecContent       string
	Constitution      string
	RelatedSpecs      []RelatedSpec
	Dependencies      []*storage.Token
	SuggestedFiles    []string
	TestGuidance      string
	TokenExample      string
	SuccessCriteria   []string
	Today             string
	SuggestedTestFile string
	PackageName       string
}

// RelatedSpec represents a related specification reference
type RelatedSpec struct {
	ReqID    string
	Feature  string
	SpecFile string
}

// selectNextPriority identifies the highest priority unimplemented requirement
// Uses database if available, falls back to filesystem scan
func selectNextPriority(dbPath string, filters map[string]string) (*storage.Token, error) {
	// Check if database file exists
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		// Fall back to filesystem scan if database doesn't exist
		return selectFromFilesystem(filters)
	}

	// Try database first
	db, err := storage.Open(dbPath)
	if err != nil {
		// Fall back to filesystem scan if database unavailable
		return selectFromFilesystem(filters)
	}

	defer db.Close()
	return selectFromDatabase(db, filters)
}

// selectFromDatabase queries the database for next priority
func selectFromDatabase(db *storage.DB, filters map[string]string) (*storage.Token, error) {
	// Build filters for incomplete requirements
	if filters == nil {
		filters = make(map[string]string)
	}

	// Load project config for ID pattern filtering
	cfg, _ := config.Load(".")
	idPattern := ""
	if cfg != nil && cfg.Requirements.IDPattern != "" {
		idPattern = cfg.Requirements.IDPattern
	}

	// If no status filter, only select STUB or IMPL by default
	if _, hasStatusFilter := filters["status"]; !hasStatusFilter {
		// Query separately for STUB and IMPL, prioritizing STUB
		stubFilters := make(map[string]string)
		for k, v := range filters {
			stubFilters[k] = v
		}
		stubFilters["status"] = "STUB"

		// Try STUB first
		tokens, err := db.ListTokens(stubFilters, idPattern, "priority ASC, updated_at DESC", 50)
		if err != nil {
			return nil, fmt.Errorf("query STUB tokens: %w", err)
		}

		// Filter out blocked tokens
		for _, token := range tokens {
			if !hasUnresolvedDependencies(db, token) {
				return token, nil
			}
		}

		// Try IMPL if no STUB available
		implFilters := make(map[string]string)
		for k, v := range filters {
			implFilters[k] = v
		}
		implFilters["status"] = "IMPL"

		tokens, err = db.ListTokens(implFilters, idPattern, "priority ASC, updated_at DESC", 50)
		if err != nil {
			return nil, fmt.Errorf("query IMPL tokens: %w", err)
		}

		for _, token := range tokens {
			if !hasUnresolvedDependencies(db, token) {
				return token, nil
			}
		}

		return nil, nil // No work available
	}

	// Use provided filters
	tokens, err := db.ListTokens(filters, idPattern, "priority ASC, updated_at DESC", 50)
	if err != nil {
		return nil, fmt.Errorf("query tokens: %w", err)
	}

	// Find first unblocked token
	for _, token := range tokens {
		if !hasUnresolvedDependencies(db, token) {
			return token, nil
		}
	}

	return nil, nil // No unblocked work available
}

// hasUnresolvedDependencies checks if a token has blocking dependencies
func hasUnresolvedDependencies(db *storage.DB, token *storage.Token) bool {
	if token.DependsOn == "" {
		return false
	}

	// Parse comma-separated dependencies
	deps := strings.Split(token.DependsOn, ",")
	for _, dep := range deps {
		dep = strings.TrimSpace(dep)
		if dep == "" {
			continue
		}

		// Query dependency status
		depTokens, err := db.GetTokensByReqID(dep)
		if err != nil || len(depTokens) == 0 {
			return true // Dependency not found = blocking
		}

		// Check if any token for this requirement is incomplete
		allComplete := true
		for _, depToken := range depTokens {
			if depToken.Status != "TESTED" && depToken.Status != "BENCHED" {
				allComplete = false
				break
			}
		}

		if !allComplete {
			return true // Dependency incomplete = blocking
		}
	}

	return false
}

// isHiddenPath determines if a token should be hidden based on its file path
func isHiddenPath(filePath string) bool {
	hiddenPatterns := []string{
		// Test files
		"_test.go", "Test.", "/tests/", "/test/",
		// Template directories
		".canary/templates/", "/templates/", "/base/", "/embedded/",
		// Documentation examples
		"IMPLEMENTATION_SUMMARY", "FINAL_SUMMARY", "README_CANARY.md", "GAP_ANALYSIS.md",
		// AI agent directories
		".claude/", ".cursor/", ".github/prompts/", ".windsurf/", ".kilocode/",
		".roo/", ".opencode/", ".codex/", ".augment/", ".codebuddy/", ".amazonq/",
	}

	for _, pattern := range hiddenPatterns {
		if strings.Contains(filePath, pattern) {
			return true
		}
	}
	return false
}

// selectFromFilesystem scans filesystem for CANARY tokens when database unavailable
func selectFromFilesystem(filters map[string]string) (*storage.Token, error) {
	// Use grep to find all CANARY tokens
	grepCmd := exec.Command("grep",
		"-rn",
		"--include=*.go", "--include=*.md", "--include=*.py",
		"--include=*.js", "--include=*.ts", "--include=*.java",
		"--include=*.rb", "--include=*.rs",
		"CANARY:",
		".",
	)

	output, err := grepCmd.CombinedOutput()
	if err != nil && len(output) == 0 {
		return nil, nil // No tokens found
	}

	// Parse tokens from grep output
	var candidates []*storage.Token
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if line == "" {
			continue
		}

		// Parse grep output: file:line:content
		parts := strings.SplitN(line, ":", 3)
		if len(parts) < 3 {
			continue
		}

		file := parts[0]
		content := parts[2]

		// Extract CANARY fields
		reqID := extractField(content, "REQ")
		feature := extractField(content, "FEATURE")
		aspect := extractField(content, "ASPECT")
		status := extractField(content, "STATUS")
		priorityStr := extractField(content, "PRIORITY")

		if reqID == "" || feature == "" {
			continue
		}

		// Apply filters
		if filterStatus, ok := filters["status"]; ok && status != filterStatus {
			continue
		}
		if filterAspect, ok := filters["aspect"]; ok && aspect != filterAspect {
			continue
		}

		// Parse priority
		priority := 5 // default
		if priorityStr != "" {
			//nolint:errcheck // Best-effort parse, default to 5 on failure
			fmt.Sscanf(priorityStr, "%d", &priority)
		}

		// Only include STUB or IMPL unless filtered
		if _, hasFilter := filters["status"]; !hasFilter {
			if status != "STUB" && status != "IMPL" {
				continue
			}
		}

		// Skip hidden paths unless include_hidden is set
		if includeHidden, ok := filters["include_hidden"]; !ok || includeHidden != "true" {
			if isHiddenPath(file) {
				continue
			}
		}

		token := &storage.Token{
			ReqID:    reqID,
			Feature:  feature,
			Aspect:   aspect,
			Status:   status,
			Priority: priority,
			FilePath: file,
			RawToken: content,
		}

		candidates = append(candidates, token)
	}

	if len(candidates) == 0 {
		return nil, nil
	}

	// Sort by priority (1=highest), then by status (STUB > IMPL)
	var best *storage.Token
	for _, candidate := range candidates {
		if best == nil {
			best = candidate
			continue
		}

		// Prefer higher priority (lower number)
		if candidate.Priority < best.Priority {
			best = candidate
			continue
		}
		if candidate.Priority > best.Priority {
			continue
		}

		// Same priority: prefer STUB over IMPL
		if candidate.Status == "STUB" && best.Status == "IMPL" {
			best = candidate
		}
	}

	return best, nil
}

// renderPrompt generates implementation prompt from template
func renderPrompt(token *storage.Token, promptFlag bool) (string, error) {
	if !promptFlag {
		// Simple summary output
		return fmt.Sprintf("Next: %s - %s (Priority: %d, Status: %s)\n"+
			"Run with --prompt for full implementation guidance.",
			token.ReqID, token.Feature, token.Priority, token.Status), nil
	}

	// Load template
	templatePath := ".canary/templates/next-prompt-template.md"
	templateContent, err := os.ReadFile(templatePath)
	if err != nil {
		return "", fmt.Errorf("read template: %w", err)
	}

	tmpl, err := template.New("next-prompt").Parse(string(templateContent))
	if err != nil {
		return "", fmt.Errorf("parse template: %w", err)
	}

	// Load prompt data
	data, err := loadPromptData(token)
	if err != nil {
		return "", fmt.Errorf("load prompt data: %w", err)
	}

	// Render template
	var buf strings.Builder
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("execute template: %w", err)
	}

	return buf.String(), nil
}

// loadPromptData loads all data needed for template rendering
func loadPromptData(token *storage.Token) (*PromptData, error) {
	data := &PromptData{
		ReqID:    token.ReqID,
		Feature:  token.Feature,
		Aspect:   token.Aspect,
		Status:   token.Status,
		Priority: token.Priority,
		Today:    time.Now().UTC().Format("2006-01-02"),
	}

	// Load specification file
	specPattern := fmt.Sprintf(".canary/specs/%s-*/spec.md", token.ReqID)
	matches, err := filepath.Glob(specPattern)
	if err == nil && len(matches) > 0 {
		data.SpecFile = matches[0]
		specContent, err := os.ReadFile(matches[0])
		if err == nil {
			data.SpecContent = string(specContent)

			// Extract success criteria from spec
			data.SuccessCriteria = extractSuccessCriteria(data.SpecContent)
		}
	}

	// Load constitution
	constitutionPath := ".canary/memory/constitution.md"
	constitutionContent, err := os.ReadFile(constitutionPath)
	if err == nil {
		data.Constitution = string(constitutionContent)
	}

	// Generate suggested files based on aspect
	data.SuggestedFiles = suggestFileLocations(token.Aspect)

	// Generate test guidance
	data.TestGuidance = generateTestGuidance(token)

	// Generate token example
	data.TokenExample = generateTokenExample(token)

	// Determine package name and test file
	data.PackageName = guessPackageName(token.Aspect)
	data.SuggestedTestFile = fmt.Sprintf("cmd/canary/%s_test.go", strings.ToLower(token.Feature))

	// Load dependencies if in database
	dbPath := ".canary/canary.db"
	if db, err := storage.Open(dbPath); err == nil {
		defer db.Close()
		if token.DependsOn != "" {
			deps := strings.Split(token.DependsOn, ",")
			for _, dep := range deps {
				dep = strings.TrimSpace(dep)
				if dep == "" {
					continue
				}
				depTokens, err := db.GetTokensByReqID(dep)
				if err == nil && len(depTokens) > 0 {
					data.Dependencies = append(data.Dependencies, depTokens[0])
				}
			}
		}
	}

	return data, nil
}

// extractSuccessCriteria extracts success criteria from specification
func extractSuccessCriteria(specContent string) []string {
	var criteria []string
	inSection := false

	lines := strings.Split(specContent, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)

		// Look for success criteria section
		if strings.Contains(strings.ToLower(line), "success criteria") {
			inSection = true
			continue
		}

		// Stop at next major section
		if inSection && strings.HasPrefix(line, "##") {
			break
		}

		// Extract list items
		if inSection && (strings.HasPrefix(line, "-") || strings.HasPrefix(line, "*")) {
			criterion := strings.TrimLeft(line, "-* \t")
			if criterion != "" {
				criteria = append(criteria, criterion)
			}
		}
	}

	if len(criteria) == 0 {
		criteria = []string{
			"Implementation meets specification requirements",
			"All tests pass",
			"Code follows project conventions",
		}
	}

	return criteria
}

// suggestFileLocations suggests file locations based on aspect
func suggestFileLocations(aspect string) []string {
	suggestions := map[string][]string{
		"CLI":      {"cmd/canary/main.go", "cmd/canary/*.go"},
		"API":      {"internal/*/api.go", "pkg/*/api.go"},
		"Engine":   {"internal/engine/*.go", "pkg/engine/*.go"},
		"Storage":  {"internal/storage/*.go"},
		"Security": {"internal/security/*.go", "pkg/security/*.go"},
	}

	if files, ok := suggestions[aspect]; ok {
		return files
	}

	return []string{"cmd/", "internal/", "pkg/"}
}

// generateTestGuidance creates test-first guidance
func generateTestGuidance(token *storage.Token) string {
	return fmt.Sprintf(`Create tests that verify the %s functionality:
- Test happy path with valid inputs
- Test error cases with invalid inputs
- Test edge cases and boundary conditions
- Test integration with existing components

Use table-driven tests where appropriate for multiple scenarios.`, token.Feature)
}

// generateTokenExample creates CANARY token placement example
func generateTokenExample(token *storage.Token) string {
	today := time.Now().UTC().Format("2006-01-02")
	return fmt.Sprintf(`// CANARY: REQ=%s; FEATURE="%s"; ASPECT=%s; STATUS=STUB; UPDATED=%s
func %s() error {
    // TODO: implement
    return nil
}`, token.ReqID, token.Feature, token.Aspect, today, token.Feature)
}

// guessPackageName guesses package name from aspect
func guessPackageName(aspect string) string {
	names := map[string]string{
		"CLI":      "main",
		"API":      "api",
		"Engine":   "engine",
		"Storage":  "storage",
		"Security": "security",
	}

	if name, ok := names[aspect]; ok {
		return name
	}

	return "main"
}

// extractField extracts a field value from a CANARY token string (already defined in main.go)
// This is a duplicate for use in next.go - consider moving to shared utility
func extractFieldInternal(token, field string) string {
	// Look for FIELD="value" or FIELD=value
	pattern := field + `="([^"]+)"`
	re := regexp.MustCompile(pattern)
	matches := re.FindStringSubmatch(token)
	if len(matches) > 1 {
		return matches[1]
	}

	// Try without quotes
	pattern = field + `=([^;\s]+)`
	re = regexp.MustCompile(pattern)
	matches = re.FindStringSubmatch(token)
	if len(matches) > 1 {
		return matches[1]
	}

	return ""
}
