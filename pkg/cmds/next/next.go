// Copyright (c) 2025 by Developer Network.
//
// For more details, see the LICENSE file in the root directory of this
// source code repository or contact Developer Network at info@devnw.com.

// CANARY: REQ=CP-252; FEATURE="NextPriorityCommand"; ASPECT=CLI; STATUS=BENCHED; TEST=TestCANARY_CBIN_132_CLI_NextPrioritySelection; BENCH=BenchmarkCANARY_CBIN_132_CLI_PriorityQuery; OWNER=canary; UPDATED=2026-08-29
package next

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"text/template"
	"time"

	"github.com/spf13/cobra"

	"devnw.dev/canary/pkg/cmds/internal/utils"
	"devnw.dev/canary/pkg/config"
	"devnw.dev/canary/pkg/external"
	"devnw.dev/canary/pkg/sources"
	"devnw.dev/canary/pkg/storage"
	"errors"
	"io/fs"
)

// CANARY: REQ=CP-252; FEATURE="NextCmd"; ASPECT=CLI; STATUS=BENCHED; TEST=TestCANARY_CBIN_132_CLI_NextPrioritySelection; BENCH=BenchmarkCANARY_CBIN_132_CLI_PriorityQuery; OWNER=canary; DOC=user:docs/user/next-priority-guide.md; DOC_HASH=17524f7a14d2c410; UPDATED=2026-08-29
var NextCmd = &cobra.Command{
	Use:   "next [flags]",
	Short: "Identify and implement next highest priority requirement",
	Long: `Identify the next highest priority unimplemented requirement and generate
comprehensive implementation guidance.

This command automatically:
- Queries database or scans filesystem for CANARY tokens
- Identifies highest priority STUB or IMPL requirement
- Excludes hidden requirements (test files, templates, examples)
- Verifies dependencies are satisfied
- Generates comprehensive implementation prompt with:
  - Specification details
  - Constitutional principles
  - Test-first guidance
  - Token placement examples

Priority determination factors:
1. PRIORITY field (1=highest, 10=lowest)
2. STATUS (STUB > IMPL > TESTED)
3. DEPENDS_ON (dependencies must be TESTED/BENCHED)
4. UPDATED field (older tokens get priority boost)`,
	RunE: func(cmd *cobra.Command, args []string) error {
		promptArg, _ := cmd.Flags().GetString("prompt-arg")
		if promptArg != "" {
			if _, err := utils.LoadPrompt(promptArg); err != nil {
				return err
			}
		}
		dbPath, _ := cmd.Flags().GetString("db")
		promptFlag, _ := cmd.Flags().GetBool("prompt")
		jsonOutput, _ := cmd.Flags().GetBool("json")
		dryRun, _ := cmd.Flags().GetBool("dry-run")
		filterStatus, _ := cmd.Flags().GetString("status")
		filterAspect, _ := cmd.Flags().GetString("aspect")
		strictExternal, _ := cmd.Flags().GetBool("strict-external")

		// Build filters
		filters := make(map[string]string)
		if filterStatus != "" {
			filters["status"] = filterStatus
		}
		if filterAspect != "" {
			filters["aspect"] = filterAspect
		}

		// Select next priority
		projectID := utils.ReadProjectID(cmd)

		token, err := selectNextPriorityStrict(dbPath, projectID, filters, strictExternal)
		if err != nil {
			// A PROJECT_REQUIRED refusal reaching here is the contract, not
			// a failure to select: it must arrive as the machine-readable
			// line on stdout, unwrapped.
			if guarded := utils.GuardContract(cmd, err); errors.Is(guarded, utils.ErrContractFailed) {
				return guarded
			}
			return fmt.Errorf("select next priority: %w", err)
		}

		if token == nil {
			fmt.Println("🎉 All requirements completed! No work available.")
			fmt.Println("\nSuggestions:")
			fmt.Println("  • Run: canary scan --verify GAP_ANALYSIS.md")
			fmt.Println("  • Review completed requirements")
			fmt.Println("  • Consider creating new specifications")
			return nil
		}

		if dryRun {
			fmt.Printf("Next priority (dry run): %s - %s\n", token.ReqID, token.Feature)
			fmt.Printf("Priority: %d | Status: %s | Aspect: %s\n", token.Priority, token.Status, token.Aspect)
			fmt.Printf("Location: %s\n", token.FilePath)
			return nil
		}

		// Render prompt
		output, err := renderPrompt(token, projectID, promptFlag)
		if err != nil {
			return fmt.Errorf("render prompt: %w", err)
		}

		if jsonOutput {
			out := nextJSONOutput{
				ReqID:    token.ReqID,
				Feature:  token.Feature,
				Aspect:   token.Aspect,
				Status:   token.Status,
				Priority: token.Priority,
				FilePath: token.FilePath,
				Updated:  token.UpdatedAt,
			}
			enc := json.NewEncoder(os.Stdout)
			enc.SetIndent("", "  ")
			if err := enc.Encode(out); err != nil {
				return fmt.Errorf("json encode: %w", err)
			}
			return nil
		}

		fmt.Println(output)
		return nil
	},
}

// nextJSONOutput is the structure emitted for next --json.
type nextJSONOutput struct {
	ReqID    string `json:"req_id"`
	Feature  string `json:"feature"`
	Aspect   string `json:"aspect"`
	Status   string `json:"status"`
	Priority int    `json:"priority"`
	FilePath string `json:"file_path"`
	Updated  string `json:"updated"`
}

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
// Uses database if available, falls back to filesystem scan. External
// dependencies with unknown (uncached) status are treated as non-blocking
// (the safe default); use selectNextPriorityStrict for --strict-external.
func selectNextPriority(dbPath, projectID string, filters map[string]string) (*storage.Token, error) {
	return selectNextPriorityStrict(dbPath, projectID, filters, false)
}

// selectNextPriorityStrict is selectNextPriority with control over whether
// external dependencies of unknown (uncached) status block selection.
// CANARY: REQ=ENG-3960; FEATURE="ExternalDeps"; ASPECT=CLI; STATUS=TESTED; TEST=TestCANARY_ENG_3960_Next_ExternalSatisfied_NotBlocking,TestCANARY_ENG_3960_Next_ExternalUnsatisfied_Blocking,TestCANARY_ENG_3960_Next_ExternalUnknown_NotBlockingByDefault,TestCANARY_ENG_3960_Next_ExternalUnknown_StrictBlocks,TestCANARY_ENG_3960_Next_LocalMissingDep_StillBlocking; UPDATED=2026-08-29
func selectNextPriorityStrict(dbPath, projectID string, filters map[string]string, strictExternal bool) (*storage.Token, error) {
	// Read-only: a repository with no index falls back to a filesystem scan
	// rather than creating an empty database and reporting nothing to do.
	db, err := storage.OpenRO(dbPath)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return selectFromFilesystem(filters)
		}
		// An index that exists but cannot be opened is still not a reason to
		// invent an answer from an empty database.
		return selectFromFilesystem(filters)
	}

	defer func() { _ = db.Close() }()
	return selectFromDatabase(db, projectID, filters, strictExternal)
}

// selectFromDatabase queries the database for next priority
func selectFromDatabase(db *storage.DB, projectID string, filters map[string]string, strictExternal bool) (*storage.Token, error) {
	// Build filters for incomplete requirements
	if filters == nil {
		filters = make(map[string]string)
	}

	// Load project config for ID pattern filtering
	cfg, err := config.Load(".")
	if err != nil {
		return nil, fmt.Errorf("load .canary/project.yaml: %w", err)
	}
	idPattern := ""
	if cfg.Requirements.IDPattern != "" {
		idPattern = cfg.Requirements.IDPattern
	}

	// Source registry + per-run dedup for the "no cached status" stderr
	// note, shared across every hasUnresolvedDependencies call this run.
	reg, err := sources.FromProjectConfig(cfg)
	if err != nil {
		return nil, fmt.Errorf("load .canary/project.yaml: %w", err)
	}
	warned := map[string]bool{}

	// If no status filter, only select STUB or IMPL by default
	if _, hasStatusFilter := filters["status"]; !hasStatusFilter {
		// Query separately for STUB and IMPL, prioritizing STUB
		stubFilters := make(map[string]string)
		for k, v := range filters {
			stubFilters[k] = v
		}
		stubFilters["status"] = "STUB"

		// Try STUB first
		tokens, err := db.ListTokens(projectID, anyFilters(stubFilters), idPattern, "", 50)
		if err != nil {
			return nil, fmt.Errorf("query STUB tokens: %w", err)
		}

		// Filter out blocked tokens
		for _, token := range tokens {
			blocked, err := hasUnresolvedDependencies(db, projectID, token, reg, ".", strictExternal, warned)
			if err != nil {
				return nil, err
			}
			if !blocked {
				return token, nil
			}
		}

		// Try IMPL if no STUB available
		implFilters := make(map[string]string)
		for k, v := range filters {
			implFilters[k] = v
		}
		implFilters["status"] = "IMPL"

		tokens, err = db.ListTokens(projectID, anyFilters(implFilters), idPattern, "", 50)
		if err != nil {
			return nil, fmt.Errorf("query IMPL tokens: %w", err)
		}

		for _, token := range tokens {
			blocked, err := hasUnresolvedDependencies(db, projectID, token, reg, ".", strictExternal, warned)
			if err != nil {
				return nil, err
			}
			if !blocked {
				return token, nil
			}
		}

		return nil, nil // No work available
	}

	// Use provided filters
	tokens, err := db.ListTokens(projectID, anyFilters(filters), idPattern, "", 50)
	if err != nil {
		return nil, fmt.Errorf("query tokens: %w", err)
	}

	// Find first unblocked token
	for _, token := range tokens {
		blocked, err := hasUnresolvedDependencies(db, projectID, token, reg, ".", strictExternal, warned)
		if err != nil {
			return nil, err
		}
		if !blocked {
			return token, nil
		}
	}

	return nil, nil // No unblocked work available
}

// hasUnresolvedDependencies checks if a token has blocking dependencies.
//
// A dependency with at least one local CANARY token keeps the original
// rule: blocking unless every token for it is TESTED/BENCHED.
//
// A dependency with ZERO local tokens is consulted against
// external.Resolve(dep, reg, root):
//   - Detail "not external" (local/flatfile id, or unconfigured prefix):
//     unchanged legacy behavior — missing = blocking.
//   - satisfied: not blocking.
//   - unsatisfied: blocking.
//   - unknown (no cached ticket status): NOT blocking by default
//     (degradation is sacred) — a one-line stderr note is printed the
//     first time a given dep is seen this run (warned dedups by id).
//     strictExternal flips unknown to blocking.
func hasUnresolvedDependencies(db *storage.DB, projectID string, token *storage.Token, reg *sources.Registry, root string, strictExternal bool, warned map[string]bool) (bool, error) {
	if token.DependsOn == "" {
		return false, nil
	}

	// Parse comma-separated dependencies
	deps := strings.Split(token.DependsOn, ",")
	for _, dep := range deps {
		dep = strings.TrimSpace(dep)
		if dep == "" {
			continue
		}

		// Query dependency status. A contract refusal -- the same id under
		// two projects with no --project to disambiguate -- is a question
		// canary cannot answer, not a dependency that happens to be
		// missing. Swallowing it reported every requirement as blocked and
		// told the user "no work available", which is a lie about the tree
		// rather than a refusal to guess.
		depTokens, err := db.GetTokensByReqID(projectID, dep)
		if err != nil && errors.Is(err, storage.ErrProjectRequired) {
			return false, err
		}
		if err != nil || len(depTokens) == 0 {
			res := external.Resolve(dep, reg, root)
			if !res.IsExternal() {
				return true, nil // Local dependency not found = blocking
			}
			switch res.State {
			case external.StateSatisfied:
				continue
			case external.StateUnsatisfied:
				return true, nil
			default: // unknown
				if strictExternal {
					return true, nil
				}
				if warned != nil && !warned[dep] {
					warned[dep] = true
					fmt.Fprintf(os.Stderr, "note: external dependency %s has no cached status (canary ticket status --refresh)\n", dep)
				}
				continue
			}
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
			return true, nil // Dependency incomplete = blocking
		}
	}

	return false, nil
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
		reqID := utils.ExtractField(content, "REQ")
		feature := utils.ExtractField(content, "FEATURE")
		aspect := utils.ExtractField(content, "ASPECT")
		status := utils.ExtractField(content, "STATUS")
		priorityStr := utils.ExtractField(content, "PRIORITY")

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
func renderPrompt(token *storage.Token, projectID string, promptFlag bool) (string, error) {
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
	data, err := loadPromptData(token, projectID)
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
// projectID scopes the dependency lookups; "" spans every project.
func loadPromptData(token *storage.Token, projectID string) (*PromptData, error) {
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
	if db, err := storage.OpenRO(dbPath); err == nil {
		defer func() { _ = db.Close() }()
		if token.DependsOn != "" {
			deps := strings.Split(token.DependsOn, ",")
			for _, dep := range deps {
				dep = strings.TrimSpace(dep)
				if dep == "" {
					continue
				}
				depTokens, err := db.GetTokensByReqID(projectID, dep)
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
// extractFieldInternal is used for internal parsing; kept for compatibility.
//
//nolint:unused
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

// anyFilters widens a string filter map to the map[string]any shape
// storage.ListTokens takes. Filter values are bound as query parameters, so
// only their type changes here, never their meaning.
func anyFilters(in map[string]string) map[string]any {
	if in == nil {
		return nil
	}
	out := make(map[string]any, len(in))
	for k, v := range in {
		out[k] = v
	}
	return out
}
