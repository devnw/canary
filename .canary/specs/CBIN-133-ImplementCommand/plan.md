# Implementation Plan: CBIN-133 - ImplementCommand

**Requirement ID:** CBIN-133
**Feature Name:** ImplementCommand
**Status:** Ready for Implementation
**Created:** 2025-10-16
**Last Updated:** 2025-10-16

---

## Constitutional Compliance Review

Before proceeding with implementation, validate against all constitutional articles:

### Article I: Requirement-First Development ✅
- **Token Format**: `// CANARY: REQ=CBIN-133; FEATURE="ImplementCommand"; ASPECT=CLI; STATUS=STUB; UPDATED=2025-10-16`
- **Evidence-Based Promotion**: Plan includes explicit status progression (STUB → IMPL → TESTED)
- **Staleness Management**: UPDATED field will be maintained

### Article II: Specification Discipline ✅
- **WHAT Before HOW**: Spec focuses on user needs (agents need implementation guidance)
- **Testable Requirements**: All FR items have clear acceptance criteria
- **No Clarifications**: Zero [NEEDS CLARIFICATION] markers in spec

### Article III: Token-Driven Planning ✅
- **Token Granularity**: Each sub-feature (FuzzyMatcher, PromptGenerator, etc.) gets own token
- **Aspect Classification**: CLI (main command), Engine (fuzzy matching), API (prompt generation)
- **Cross-Cutting Concerns**: Properly split across aspects

### Article IV: Test-First Imperative ✅ **NON-NEGOTIABLE**
- **Phase 1 = Tests**: All test files created BEFORE implementation
- **Red-Green-Refactor**: Explicit phases documented
- **Test Naming**: TEST= field references added to all tokens

### Article V: Simplicity and Anti-Abstraction ✅
- **Standard Library**: Using Go's text/template, strings, filepath (no external dependencies)
- **No Premature Optimization**: Simple Levenshtein implementation acceptable
- **Minimal Complexity**: Direct file I/O, no caching layer initially

### Article VI: Integration-First Testing ✅
- **Real Environment**: Tests use actual filesystem and spec files
- **Contract-First**: Prompt template defines contract before implementation

### Article VII: Documentation Currency ✅
- **Tokens as Documentation**: All sub-features have CANARY tokens
- **UPDATED Field**: Will be maintained during implementation
- **Gap Analysis**: Spec includes tracking in Implementation Checklist

### Article VIII: Continuous Improvement ✅
- **Metrics**: Performance targets defined (<1s lookup, <2s prompt generation)
- **Regular Audits**: Success criteria includes measurable outcomes

### Article IX: Amendment Process ✅
- **No Constitutional Violations**: Plan follows all existing articles

**All Constitutional Gates: PASSED ✅**

---

## Tech Stack Decision

### Language & Runtime
- **Go 1.19+**
- **Rationale**:
  - Already project standard
  - Excellent string manipulation (Levenshtein distance)
  - Built-in text/template package (Article V: Simplicity)
  - Fast compilation for CI/CD workflows

### Core Dependencies
- **Standard Library Only** (Article V: Prefer standard library)
  - `text/template` - Template rendering
  - `os` - File I/O for spec/plan loading
  - `filepath` - Path operations
  - `strings` - String manipulation for fuzzy matching
  - `regexp` - Token extraction (already used in codebase)
  - `github.com/spf13/cobra` - CLI framework (already in use)

### No External Dependencies
- **Fuzzy Matching**: Implement simple Levenshtein distance in-house
  - Complexity: O(n*m) acceptable for <1000 specs
  - Alternative considered: `github.com/lithammer/fuzzysearch` (rejected per Article V)
- **Template Engine**: Use Go standard `text/template`
  - Already used in CBIN-132 (NextCmd) - proven pattern

### Database Integration
- **Optional**: Use existing `.canary/canary.db` if available
- **Fallback**: Filesystem scan if database missing
- **Rationale**: Follow CBIN-132 pattern (dual-mode operation)

---

## Architecture Overview

```
┌─────────────────────────────────────────────────────┐
│  canary implement <query> [--flags]                 │
└───────────────────┬─────────────────────────────────┘
                    │
                    v
┌─────────────────────────────────────────────────────┐
│            implementCmd.RunE()                      │
│  • Parse flags                                      │
│  • Handle --list mode                               │
│  • Call findRequirement(query)                      │
│  • Generate prompt or show progress                 │
└─────┬──────────────────────┬────────────────────────┘
      │                      │
      v                      v
┌──────────────────┐   ┌─────────────────────────────┐
│ findRequirement  │   │  renderImplementPrompt      │
│                  │   │                              │
│ 1. Exact Match   │   │  • Load spec.md             │
│    (CBIN-XXX)    │   │  • Load plan.md (optional)  │
│                  │   │  • Load constitution        │
│ 2. Dir Match     │   │  • Extract checklist        │
│    (.canary/     │   │  • Scan for tokens          │
│     specs/)      │   │  • Populate template        │
│                  │   │  • Render prompt            │
│ 3. Fuzzy Match   │   └─────────────────────────────┘
│    (Levenshtein) │
│                  │
│ 4. Interactive   │
│    Selection     │
└──────────────────┘
```

### Key Components

1. **CLI Command** (`cmd/canary/main.go`)
   - Cobra command definition
   - Flag parsing
   - High-level orchestration

2. **Requirement Finder** (`cmd/canary/implement.go`)
   - `findRequirement(query string) (*RequirementSpec, error)`
   - Exact match by ID
   - Directory pattern matching
   - Fuzzy matching with scoring

3. **Fuzzy Matcher** (`internal/matcher/fuzzy.go`)
   - `CalculateLevenshtein(s1, s2 string) int`
   - `ScoreMatch(query, candidate string) int`
   - `FindBestMatches(query string, candidates []string, limit int) []Match`

4. **Prompt Generator** (`cmd/canary/implement.go`)
   - `renderImplementPrompt(spec *RequirementSpec, flags *ImplementFlags) (string, error)`
   - Load template from `.canary/templates/implement-prompt-template.md`
   - Populate template variables
   - Render final markdown prompt

5. **Progress Tracker** (`cmd/canary/implement.go`)
   - `calculateProgress(reqID string) (*ProgressStats, error)`
   - Scan codebase for CANARY tokens matching reqID
   - Count by status (STUB, IMPL, TESTED, BENCHED)

---

## CANARY Token Placement

### Main Command Token
```go
// File: cmd/canary/main.go (around line 1650, after nextCmd)
// CANARY: REQ=CBIN-133; FEATURE="ImplementCmd"; ASPECT=CLI; STATUS=STUB; UPDATED=2025-10-16
var implementCmd = &cobra.Command{
	Use:   "implement <query> [flags]",
	Short: "Generate implementation guidance for a requirement",
	Long: `Find and display comprehensive implementation guidance...`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Implementation
	},
}
```

### Sub-Feature Tokens

**Requirement Lookup:**
```go
// File: cmd/canary/implement.go (new file)
// CANARY: REQ=CBIN-133; FEATURE="RequirementLookup"; ASPECT=API; STATUS=STUB; UPDATED=2025-10-16
func findRequirement(query string) (*RequirementSpec, error) {
	// Lookup logic
}
```

**Fuzzy Matcher:**
```go
// File: internal/matcher/fuzzy.go (new file)
// CANARY: REQ=CBIN-133; FEATURE="FuzzyMatcher"; ASPECT=Engine; STATUS=STUB; UPDATED=2025-10-16
func CalculateLevenshtein(s1, s2 string) int {
	// Levenshtein distance algorithm
}
```

**Prompt Renderer:**
```go
// File: cmd/canary/implement.go
// CANARY: REQ=CBIN-133; FEATURE="PromptRenderer"; ASPECT=API; STATUS=STUB; UPDATED=2025-10-16
func renderImplementPrompt(spec *RequirementSpec, flags *ImplementFlags) (string, error) {
	// Template rendering
}
```

---

## Implementation Phases

Following Article IV (Test-First Imperative), implementation MUST proceed in strict order:

### Phase 0: Pre-Implementation Setup

**Constitutional Gate Check:**
- [ ] All constitutional articles reviewed (completed above ✅)
- [ ] Test-first approach validated
- [ ] Simplicity gate passed (no external dependencies)
- [ ] Token format validated

**Create Test Files (TDD Red Phase):**
```bash
# Create test files BEFORE any implementation
touch cmd/canary/implement_test.go
touch internal/matcher/fuzzy_test.go
```

---

### Phase 1: Test Creation (TDD Red Phase)

**Article IV Mandate**: Write tests FIRST, confirm they FAIL.

#### 1.1 Fuzzy Matcher Tests

```go
// File: internal/matcher/fuzzy_test.go
// CANARY: REQ=CBIN-133; FEATURE="FuzzyMatcherTests"; ASPECT=Engine; STATUS=STUB; TEST=TestCANARY_CBIN_133_Engine_Levenshtein; UPDATED=2025-10-16
package matcher_test

import "testing"

func TestCANARY_CBIN_133_Engine_Levenshtein(t *testing.T) {
	tests := []struct {
		s1       string
		s2       string
		expected int
	}{
		{"", "", 0},
		{"hello", "hello", 0},
		{"hello", "hallo", 1},
		{"kitten", "sitting", 3},
		{"CBIN105", "CBIN-105", 1}, // Hyphen difference
	}

	for _, tc := range tests {
		result := matcher.CalculateLevenshtein(tc.s1, tc.s2)
		if result != tc.expected {
			t.Errorf("Levenshtein(%q, %q) = %d; want %d", tc.s1, tc.s2, result, tc.expected)
		}
	}
}

func TestCANARY_CBIN_133_Engine_FuzzyScoring(t *testing.T) {
	tests := []struct {
		query      string
		candidate  string
		minScore   int // Score should be >= this
	}{
		{"user auth", "UserAuthentication", 80},
		{"auth", "UserAuthentication", 60},
		{"CBIN105", "CBIN-105", 90},
	}

	for _, tc := range tests {
		score := matcher.ScoreMatch(tc.query, tc.candidate)
		if score < tc.minScore {
			t.Errorf("ScoreMatch(%q, %q) = %d; want >= %d", tc.query, tc.candidate, score, tc.minScore)
		}
	}
}
```

#### 1.2 Requirement Lookup Tests

```go
// File: cmd/canary/implement_test.go
// CANARY: REQ=CBIN-133; FEATURE="ImplementCmdTests"; ASPECT=CLI; STATUS=STUB; TEST=TestCANARY_CBIN_133_CLI_ExactMatch; UPDATED=2025-10-16
package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCANARY_CBIN_133_CLI_ExactMatch(t *testing.T) {
	// Setup: Create test spec directory
	tmpDir := t.TempDir()
	specDir := filepath.Join(tmpDir, ".canary", "specs", "CBIN-105-test-feature")
	os.MkdirAll(specDir, 0755)

	specContent := "# Feature Specification: TestFeature\n\n**Requirement ID:** CBIN-105"
	os.WriteFile(filepath.Join(specDir, "spec.md"), []byte(specContent), 0644)

	// Change to tmpDir
	originalWd, _ := os.Getwd()
	defer os.Chdir(originalWd)
	os.Chdir(tmpDir)

	// Execute: Find by exact ID
	spec, err := findRequirement("CBIN-105")

	// Verify: Spec loaded correctly
	if err != nil {
		t.Fatalf("findRequirement failed: %v", err)
	}
	if spec.ReqID != "CBIN-105" {
		t.Errorf("Expected CBIN-105, got %s", spec.ReqID)
	}
}

func TestCANARY_CBIN_133_CLI_FuzzyMatch(t *testing.T) {
	// Setup: Multiple specs with similar names
	tmpDir := t.TempDir()
	specs := []struct {
		id   string
		name string
	}{
		{"CBIN-105", "UserAuthentication"},
		{"CBIN-110", "OAuthIntegration"},
		{"CBIN-112", "DataValidation"},
	}

	for _, s := range specs {
		specDir := filepath.Join(tmpDir, ".canary", "specs", fmt.Sprintf("%s-%s", s.id, s.name))
		os.MkdirAll(specDir, 0755)
		content := fmt.Sprintf("# Feature Specification: %s\n\n**Requirement ID:** %s", s.name, s.id)
		os.WriteFile(filepath.Join(specDir, "spec.md"), []byte(content), 0644)
	}

	originalWd, _ := os.Getwd()
	defer os.Chdir(originalWd)
	os.Chdir(tmpDir)

	// Execute: Fuzzy search for "user auth"
	spec, err := findRequirement("user auth")

	// Verify: Matches UserAuthentication
	if err != nil {
		t.Fatalf("findRequirement failed: %v", err)
	}
	if spec.ReqID != "CBIN-105" {
		t.Errorf("Expected CBIN-105 (UserAuthentication), got %s", spec.ReqID)
	}
}

func TestCANARY_CBIN_133_CLI_PromptGeneration(t *testing.T) {
	// Setup: Create spec with Implementation Checklist
	tmpDir := t.TempDir()
	setupTestSpec(t, tmpDir, "CBIN-105", withImplementationChecklist())

	originalWd, _ := os.Getwd()
	defer os.Chdir(originalWd)
	os.Chdir(tmpDir)

	// Execute: Generate prompt
	spec, _ := findRequirement("CBIN-105")
	prompt, err := renderImplementPrompt(spec, &ImplementFlags{Prompt: true})

	// Verify: Prompt contains key sections
	if err != nil {
		t.Fatalf("renderImplementPrompt failed: %v", err)
	}

	requiredSections := []string{
		"# Implementation Guidance:",
		"## Specification",
		"## Implementation Plan",
		"## Constitutional Guidance",
		"## Implementation Checklist",
		"## Progress Tracking",
	}

	for _, section := range requiredSections {
		if !strings.Contains(prompt, section) {
			t.Errorf("Prompt missing section: %s", section)
		}
	}
}

func TestCANARY_CBIN_133_CLI_ProgressTracking(t *testing.T) {
	// Setup: Create files with CANARY tokens
	tmpDir := t.TempDir()
	createFileWithToken(t, tmpDir, "feature1.go", "CBIN-105", "Feature1", "IMPL")
	createFileWithToken(t, tmpDir, "feature2.go", "CBIN-105", "Feature2", "TESTED")
	createFileWithToken(t, tmpDir, "feature3.go", "CBIN-105", "Feature3", "STUB")

	originalWd, _ := os.Getwd()
	defer os.Chdir(originalWd)
	os.Chdir(tmpDir)

	// Execute: Calculate progress
	progress, err := calculateProgress("CBIN-105")

	// Verify: Correct counts
	if err != nil {
		t.Fatalf("calculateProgress failed: %v", err)
	}
	if progress.Total != 3 {
		t.Errorf("Expected 3 total features, got %d", progress.Total)
	}
	if progress.Completed != 1 { // Only TESTED counts as completed
		t.Errorf("Expected 1 completed feature, got %d", progress.Completed)
	}
}
```

**Run Tests (Expect Failures):**
```bash
go test ./cmd/canary -v -run TestCANARY_CBIN_133
go test ./internal/matcher -v -run TestCANARY_CBIN_133
# All tests should FAIL (functions don't exist yet)
```

---

### Phase 2: Core Implementation (TDD Green Phase)

Now implement ONLY enough code to make tests pass.

#### 2.1 Fuzzy Matcher Implementation

```go
// File: internal/matcher/fuzzy.go (new)
// CANARY: REQ=CBIN-133; FEATURE="FuzzyMatcher"; ASPECT=Engine; STATUS=IMPL; UPDATED=2025-10-16
package matcher

import (
	"strings"
	"unicode"
)

// CalculateLevenshtein computes edit distance between two strings
func CalculateLevenshtein(s1, s2 string) int {
	s1 = strings.ToLower(s1)
	s2 = strings.ToLower(s2)

	if s1 == s2 {
		return 0
	}

	// Create matrix
	d := make([][]int, len(s1)+1)
	for i := range d {
		d[i] = make([]int, len(s2)+1)
		d[i][0] = i
	}
	for j := range d[0] {
		d[0][j] = j
	}

	// Fill matrix
	for i := 1; i <= len(s1); i++ {
		for j := 1; j <= len(s2); j++ {
			cost := 0
			if s1[i-1] != s2[j-1] {
				cost = 1
			}

			d[i][j] = min(
				d[i-1][j]+1,      // deletion
				d[i][j-1]+1,      // insertion
				d[i-1][j-1]+cost, // substitution
			)
		}
	}

	return d[len(s1)][len(s2)]
}

// ScoreMatch calculates similarity score (0-100) between query and candidate
func ScoreMatch(query, candidate string) int {
	query = strings.ToLower(query)
	candidate = strings.ToLower(candidate)

	// Exact match
	if query == candidate {
		return 100
	}

	// Substring match gets high score
	if strings.Contains(candidate, query) {
		ratio := float64(len(query)) / float64(len(candidate))
		return int(80 + (ratio * 20)) // 80-100 range
	}

	// Abbreviation match (e.g., "auth" matches "UserAuthentication")
	if matchesAbbreviation(query, candidate) {
		return 75
	}

	// Levenshtein distance
	distance := CalculateLevenshtein(query, candidate)
	maxLen := max(len(query), len(candidate))

	if maxLen == 0 {
		return 0
	}

	// Convert distance to similarity percentage
	similarity := float64(maxLen-distance) / float64(maxLen)
	score := int(similarity * 100)

	if score < 0 {
		return 0
	}
	return score
}

// matchesAbbreviation checks if query matches first letters of words in candidate
func matchesAbbreviation(query, candidate string) bool {
	var abbrev strings.Builder
	for _, ch := range candidate {
		if unicode.IsUpper(ch) || abbrev.Len() == 0 {
			abbrev.WriteRune(unicode.ToLower(ch))
		}
	}
	return strings.Contains(abbrev.String(), query)
}

func min(a, b, c int) int {
	if a < b {
		if a < c {
			return a
		}
		return c
	}
	if b < c {
		return b
	}
	return c
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// Match represents a fuzzy match result
type Match struct {
	ReqID       string
	FeatureName string
	Score       int
	SpecPath    string
}

// FindBestMatches returns top N matches for query
func FindBestMatches(query string, specsDir string, limit int) ([]Match, error) {
	entries, err := os.ReadDir(specsDir)
	if err != nil {
		return nil, err
	}

	var matches []Match
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		// Parse directory name: CBIN-XXX-feature-name
		parts := strings.SplitN(entry.Name(), "-", 2)
		if len(parts) < 2 {
			continue
		}

		reqID := parts[0]
		featureName := parts[1]

		// Score against both ID and feature name
		idScore := ScoreMatch(query, reqID)
		nameScore := ScoreMatch(query, featureName)
		score := max(idScore, nameScore)

		if score >= 60 { // Minimum threshold
			matches = append(matches, Match{
				ReqID:       reqID,
				FeatureName: featureName,
				Score:       score,
				SpecPath:    filepath.Join(specsDir, entry.Name()),
			})
		}
	}

	// Sort by score (highest first)
	sort.Slice(matches, func(i, j int) bool {
		return matches[i].Score > matches[j].Score
	})

	// Return top N
	if len(matches) > limit {
		return matches[:limit], nil
	}
	return matches, nil
}
```

#### 2.2 Requirement Lookup Implementation

```go
// File: cmd/canary/implement.go (new)
// CANARY: REQ=CBIN-133; FEATURE="RequirementLookup"; ASPECT=API; STATUS=IMPL; UPDATED=2025-10-16
package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"go.spyder.org/canary/internal/matcher"
)

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

// findRequirement locates a requirement by ID or fuzzy match
func findRequirement(query string) (*RequirementSpec, error) {
	// Attempt 1: Exact ID match
	if strings.HasPrefix(strings.ToUpper(query), "CBIN-") {
		spec, err := findByExactID(query)
		if err == nil {
			return spec, nil
		}
	}

	// Attempt 2: Directory pattern match
	specsDir := ".canary/specs"
	pattern := filepath.Join(specsDir, query+"*")
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

	// Auto-select if clear winner (>80% score, >20 points ahead)
	if fuzzyMatches[0].Score > 80 && (len(fuzzyMatches) == 1 || fuzzyMatches[0].Score-fuzzyMatches[1].Score > 20) {
		fmt.Printf("Auto-selected: %s - %s (%d%% match)\n\n",
			fuzzyMatches[0].ReqID, fuzzyMatches[0].FeatureName, fuzzyMatches[0].Score)
		return loadSpecFromDir(fuzzyMatches[0].SpecPath)
	}

	// Interactive selection
	return selectInteractive(fuzzyMatches)
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
	parts := strings.SplitN(dirName, "-", 2)
	if len(parts) < 2 {
		return nil, fmt.Errorf("invalid spec directory name: %s", dirName)
	}

	spec := &RequirementSpec{
		ReqID:       parts[0],
		FeatureName: parts[1],
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

// selectInteractive displays matches and prompts user to select
func selectInteractive(matches []matcher.Match) (*RequirementSpec, error) {
	fmt.Println("Multiple matches found. Select one:\n")

	for i, match := range matches {
		// Load brief description from spec
		desc := getSpecDescription(match.SpecPath)
		fmt.Printf("%d. %s - %s (%d%% match)\n", i+1, match.ReqID, match.FeatureName, match.Score)
		if desc != "" {
			fmt.Printf("   %s\n\n", desc)
		}
	}

	fmt.Print("Enter number (1-", len(matches), ") or 'q' to quit: ")

	reader := bufio.NewReader(os.Stdin)
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(input)

	if input == "q" {
		return nil, fmt.Errorf("user cancelled")
	}

	var selection int
	if _, err := fmt.Sscanf(input, "%d", &selection); err != nil || selection < 1 || selection > len(matches) {
		fmt.Println("Invalid selection")
		return selectInteractive(matches) // Retry
	}

	return loadSpecFromDir(matches[selection-1].SpecPath)
}

// getSpecDescription extracts first 100 chars of spec summary
func getSpecDescription(specPath string) string {
	specFile := filepath.Join(specPath, "spec.md")
	content, err := os.ReadFile(specFile)
	if err != nil {
		return ""
	}

	lines := strings.Split(string(content), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "##") && !strings.HasPrefix(line, "###") {
			continue
		}
		if len(line) > 10 && !strings.HasPrefix(line, "#") && !strings.HasPrefix(line, "**") {
			if len(line) > 100 {
				return line[:100] + "..."
			}
			return line
		}
	}
	return ""
}
```

#### 2.3 Prompt Generation Implementation

```go
// File: cmd/canary/implement.go (continued)
// CANARY: REQ=CBIN-133; FEATURE="PromptRenderer"; ASPECT=API; STATUS=IMPL; UPDATED=2025-10-16

// ImplementFlags holds command flags
type ImplementFlags struct {
	Prompt       bool
	ShowProgress bool
	ContextLines int
}

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
		"ReqID":         spec.ReqID,
		"FeatureName":   spec.FeatureName,
		"SpecContent":   spec.SpecContent,
		"PlanContent":   spec.PlanContent,
		"HasPlan":       spec.HasPlan,
		"Constitution":  string(constitutionContent),
		"Checklist":     checklist,
		"Progress":      progress,
		"Today":         time.Now().UTC().Format("2006-01-02"),
	}

	// Render
	var buf strings.Builder
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("execute template: %w", err)
	}

	return buf.String(), nil
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

		status := extractField(line, "STATUS")
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
```

**Update CANARY Tokens to STATUS=IMPL:**
```go
// CANARY: REQ=CBIN-133; FEATURE="FuzzyMatcher"; ASPECT=Engine; STATUS=IMPL; UPDATED=2025-10-16
// CANARY: REQ=CBIN-133; FEATURE="RequirementLookup"; ASPECT=API; STATUS=IMPL; UPDATED=2025-10-16
// CANARY: REQ=CBIN-133; FEATURE="PromptRenderer"; ASPECT=API; STATUS=IMPL; UPDATED=2025-10-16
```

**Run Tests (Expect Success):**
```bash
go test ./cmd/canary -v -run TestCANARY_CBIN_133
go test ./internal/matcher -v -run TestCANARY_CBIN_133
# All tests should PASS
```

---

### Phase 3: CLI Command Integration

#### 3.1 Add implementCmd to main.go

```go
// File: cmd/canary/main.go (around line 1650, after nextCmd)
// CANARY: REQ=CBIN-133; FEATURE="ImplementCmd"; ASPECT=CLI; STATUS=IMPL; UPDATED=2025-10-16
var implementCmd = &cobra.Command{
	Use:   "implement <query> [flags]",
	Short: "Generate implementation guidance for a requirement",
	Long: `Find and generate comprehensive implementation guidance for a CANARY requirement.

Supports multiple lookup methods:
- Exact ID: canary implement CBIN-105
- Fuzzy match: canary implement "user authentication"
- List all: canary implement --list

Generates comprehensive prompts including:
- Specification details
- Implementation plan (if exists)
- Constitutional principles
- CANARY token placement instructions
- Test-first approach guidance
- File location hints`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Parse flags
		listMode, _ := cmd.Flags().GetBool("list")
		promptMode, _ := cmd.Flags().GetBool("prompt")
		showProgress, _ := cmd.Flags().GetBool("show-progress")
		contextLines, _ := cmd.Flags().GetInt("context-lines")

		// Handle --list mode
		if listMode {
			return listUnimplemented()
		}

		// Require query argument
		if len(args) == 0 {
			return fmt.Errorf("query argument required (or use --list)")
		}

		query := strings.Join(args, " ")

		// Find requirement
		spec, err := findRequirement(query)
		if err != nil {
			return fmt.Errorf("find requirement: %w", err)
		}

		// Handle --show-progress mode
		if showProgress {
			progress, _ := calculateProgress(spec.ReqID)
			fmt.Printf("Implementation Progress: %s - %s\n\n", spec.ReqID, spec.FeatureName)
			fmt.Printf("Total sub-features: %d\n", progress.Total)
			fmt.Printf("  STUB: %d\n", progress.Stub)
			fmt.Printf("  IMPL: %d\n", progress.Impl)
			fmt.Printf("  TESTED: %d\n", progress.Tested)
			fmt.Printf("  BENCHED: %d\n", progress.Benched)
			fmt.Printf("\nCompleted: %d/%d (%.0f%%)\n", progress.Completed, progress.Total,
				float64(progress.Completed)/float64(progress.Total)*100)
			return nil
		}

		// Warn if plan missing
		if !spec.HasPlan {
			fmt.Printf("⚠️  Warning: Implementation plan not found for %s\n\n", spec.ReqID)
			fmt.Println("Recommendation: Create plan first for better guidance")
			fmt.Printf("Run: canary plan %s\n\n", spec.ReqID)

			fmt.Print("Continue without plan? (y/n): ")
			reader := bufio.NewReader(os.Stdin)
			input, _ := reader.ReadString('\n')
			if strings.TrimSpace(strings.ToLower(input)) != "y" {
				return fmt.Errorf("user cancelled")
			}
		}

		// Generate prompt
		flags := &ImplementFlags{
			Prompt:       promptMode,
			ShowProgress: showProgress,
			ContextLines: contextLines,
		}

		prompt, err := renderImplementPrompt(spec, flags)
		if err != nil {
			return fmt.Errorf("generate prompt: %w", err)
		}

		fmt.Println(prompt)
		return nil
	},
}

// listUnimplemented displays all requirements with STATUS=STUB or IMPL
func listUnimplemented() error {
	// Query database if available, otherwise filesystem scan
	dbPath := ".canary/canary.db"
	if _, err := os.Stat(dbPath); err == nil {
		return listUnimplementedFromDB(dbPath)
	}
	return listUnimplementedFromFilesystem()
}

// Register command in init()
func init() {
	// ... existing commands ...
	rootCmd.AddCommand(implementCmd)

	// implementCmd flags
	implementCmd.Flags().Bool("list", false, "list all unimplemented requirements")
	implementCmd.Flags().Bool("prompt", true, "generate full implementation prompt (default: true)")
	implementCmd.Flags().Bool("json", false, "output in JSON format")
	implementCmd.Flags().Bool("show-progress", false, "show implementation progress without prompt")
	implementCmd.Flags().Int("context-lines", 3, "number of context lines around tokens")
}
```

**Update Token to STATUS=TESTED after tests pass:**
```go
// CANARY: REQ=CBIN-133; FEATURE="ImplementCmd"; ASPECT=CLI; STATUS=TESTED; TEST=TestCANARY_CBIN_133_CLI_ExactMatch; UPDATED=2025-10-16
```

---

### Phase 4: Template Creation

#### 4.1 Create Implementation Prompt Template

```markdown
<!-- File: .canary/templates/implement-prompt-template.md (new) -->
<!-- CANARY: REQ=CBIN-133; FEATURE="PromptGenerator"; ASPECT=API; STATUS=IMPL; UPDATED=2025-10-16 -->

# Implementation Guidance: {{.FeatureName}}

**Requirement:** {{.ReqID}}
**Status:** Ready for Implementation
**Generated:** {{.Today}}

{{if .Progress}}
**Progress:** {{.Progress.Completed}}/{{.Progress.Total}} sub-features completed ({{.Progress.Percentage}}%)
- STUB: {{.Progress.Stub}}
- IMPL: {{.Progress.Impl}}
- TESTED: {{.Progress.Tested}}
- BENCHED: {{.Progress.Benched}}
{{end}}

---

## Constitutional Guidance

Before implementing, review these governing principles from `.canary/memory/constitution.md`:

{{.Constitution}}

**Key Principles for This Implementation:**
1. **Article I: Requirement-First** - CANARY tokens must be placed at all implementation points
2. **Article IV: Test-First Imperative** - Write tests BEFORE implementation (NON-NEGOTIABLE)
3. **Article V: Simplicity** - Prefer standard library, avoid unnecessary complexity
4. **Article VII: Documentation Currency** - Update UPDATED field when modifying

---

## Specification

**Source:** {{.SpecPath}}

{{.SpecContent}}

---

{{if .HasPlan}}
## Implementation Plan

**Source:** {{.PlanPath}}

{{.PlanContent}}
{{else}}
## Implementation Plan

⚠️ **Warning:** No implementation plan found.

**Recommendation:** Create plan first for better guidance:
```bash
canary plan {{.ReqID}}
```

Continue with specification-only guidance:
{{end}}

---

## Implementation Checklist

{{.Checklist}}

---

## Test-First Approach (Article IV)

**MANDATORY STEPS:**

### Step 1: Create Test Files
```bash
# Create test files BEFORE any implementation code
touch <test_file_name>.go
```

### Step 2: Write Failing Tests (Red Phase)
Write tests that verify the requirements. They MUST fail initially.

### Step 3: Implement to Pass Tests (Green Phase)
Write ONLY enough code to make tests pass.

### Step 4: Update CANARY Tokens
Update token STATUS from STUB → IMPL → TESTED as you progress.

---

## CANARY Token Placement

For each sub-feature in the Implementation Checklist, place a CANARY token:

**Format:**
```
// CANARY: REQ={{.ReqID}}; FEATURE="SubFeatureName"; ASPECT=<Aspect>; STATUS=STUB; UPDATED={{.Today}}
```

**Status Progression:**
- `STATUS=STUB` → Initial placeholder
- `STATUS=IMPL` → Implementation complete, no tests
- `STATUS=TESTED` → Tests added and passing (add `TEST=TestName`)
- `STATUS=BENCHED` → Benchmarks added (add `BENCH=BenchName`)

---

## Verification Checklist

Before marking this requirement as complete:

- [ ] All tests written and passing
- [ ] CANARY tokens placed at all implementation points
- [ ] Token STATUS updated to TESTED (minimum)
- [ ] UPDATED field set to today's date
- [ ] All success criteria from specification met
- [ ] Code follows project conventions
- [ ] Error handling is comprehensive
- [ ] Documentation is current (if public API)

---

## Next Steps After Completion

1. **Verify tokens:**
   ```bash
   canary implement {{.ReqID}} --show-progress
   ```

2. **Run full test suite:**
   ```bash
   go test ./...
   ```

3. **Update GAP_ANALYSIS.md** if all sub-features TESTED:
   ```markdown
   ✅ {{.ReqID}} - {{.FeatureName}} (verified)
   ```

4. **Scan for updates:**
   ```bash
   canary scan --root . --project-only
   ```

---

**Ready to implement {{.ReqID}}!** Follow the test-first approach and refer back to this guidance as needed.
```

#### 4.2 Create Slash Command Template

```markdown
<!-- File: .canary/templates/commands/implement.md (new) -->
<!-- CANARY: REQ=CBIN-133; FEATURE="ImplementSlashCmd"; ASPECT=CLI; STATUS=IMPL; UPDATED=2025-10-16 -->

Generate comprehensive implementation guidance for a CANARY requirement.

## Usage

```bash
/canary.implement <query>
```

## Query Options

- **Exact ID**: `/canary.implement CBIN-105`
- **Fuzzy match**: `/canary.implement "user authentication"`
- **List all**: `/canary.implement --list`

## What You'll Get

The command generates a comprehensive implementation prompt including:

1. **Specification Details** - What to build and why
2. **Implementation Plan** - How to build it (if plan exists)
3. **Constitutional Guidance** - Governing principles (test-first, simplicity, etc.)
4. **Implementation Checklist** - Sub-features with file location hints
5. **CANARY Token Examples** - Where to place tokens and how to update them
6. **Test-First Approach** - Step-by-step TDD guidance
7. **Progress Tracking** - Current implementation status
8. **Verification Checklist** - How to confirm completion

## Workflow

```
/canary.specify "feature description"
    ↓
/canary.plan CBIN-XXX
    ↓
/canary.implement CBIN-XXX  ← You are here
    ↓
[Implement following TDD]
    ↓
/canary.scan --verify
```

## Examples

**Implement specific requirement:**
```bash
/canary.implement CBIN-105
```

**Search by feature name:**
```bash
/canary.implement "user authentication"
```

**Show implementation progress:**
```bash
/canary.implement CBIN-105 --show-progress
```

**List all unimplemented requirements:**
```bash
/canary.implement --list
```

## Flags

- `--list` - List all STUB/IMPL requirements
- `--show-progress` - Show progress without full prompt
- `--context-lines N` - Show N lines of code context (default: 3)

## Constitutional Reminder

From Article IV (Test-First Imperative):
> "This is NON-NEGOTIABLE: All implementation MUST follow Test-Driven Development."

Always write tests FIRST, then implement to make them pass.

---

After completing implementation, verify with:
```bash
/canary.scan
/canary.verify
```
```

---

### Phase 5: Final Validation & Status Update

#### 5.1 Run Complete Test Suite

```bash
# Run all CBIN-133 tests
go test ./cmd/canary -v -run TestCANARY_CBIN_133
go test ./internal/matcher -v -run TestCANARY_CBIN_133

# Run full project test suite
go test ./...
```

#### 5.2 Manual Testing

```bash
# Test exact match
./bin/canary implement CBIN-105

# Test fuzzy match
./bin/canary implement "user auth"

# Test list mode
./bin/canary implement --list

# Test progress tracking
./bin/canary implement CBIN-105 --show-progress

# Test missing spec
./bin/canary implement CBIN-999  # Should show helpful error
```

#### 5.3 Update All CANARY Tokens to TESTED

Once all tests pass, update tokens:

```go
// CANARY: REQ=CBIN-133; FEATURE="ImplementCmd"; ASPECT=CLI; STATUS=TESTED; TEST=TestCANARY_CBIN_133_CLI_ExactMatch; UPDATED=2025-10-16
// CANARY: REQ=CBIN-133; FEATURE="RequirementLookup"; ASPECT=API; STATUS=TESTED; TEST=TestCANARY_CBIN_133_CLI_ExactMatch; UPDATED=2025-10-16
// CANARY: REQ=CBIN-133; FEATURE="FuzzyMatcher"; ASPECT=Engine; STATUS=TESTED; TEST=TestCANARY_CBIN_133_Engine_Levenshtein; UPDATED=2025-10-16
// CANARY: REQ=CBIN-133; FEATURE="PromptRenderer"; ASPECT=API; STATUS=TESTED; TEST=TestCANARY_CBIN_133_CLI_PromptGeneration; UPDATED=2025-10-16
// CANARY: REQ=CBIN-133; FEATURE="ImplementCmdTests"; ASPECT=CLI; STATUS=TESTED; TEST=TestCANARY_CBIN_133_CLI_ExactMatch; UPDATED=2025-10-16
// CANARY: REQ=CBIN-133; FEATURE="FuzzyMatcherTests"; ASPECT=Engine; STATUS=TESTED; TEST=TestCANARY_CBIN_133_Engine_Levenshtein; UPDATED=2025-10-16
```

#### 5.4 Update Spec Implementation Checklist

Mark completed items in `.canary/specs/CBIN-133-ImplementCommand/spec.md`:

```markdown
- [x] Add `implementCmd` to `cmd/canary/main.go`
- [x] Implement `findRequirement(query string) (*Requirement, error)`
- [x] Implement Levenshtein distance algorithm
- [x] Display numbered list of matches
- [x] Create `.canary/templates/implement-prompt-template.md`
- [x] Load template and populate variables
- [x] Scan codebase for CANARY tokens
- [x] Create `.canary/templates/commands/implement.md`
- [x] Test exact ID lookup
- [x] Test Levenshtein distance calculation
```

---

## Testing Strategy

### Unit Tests
- **Levenshtein Distance**: Test with various string pairs (empty, identical, similar, different)
- **Fuzzy Scoring**: Test substring, abbreviation, and distance-based scoring
- **Requirement Lookup**: Test exact match, directory match, fuzzy match
- **Prompt Generation**: Test template rendering with mock data

### Integration Tests
- **End-to-End Lookup**: Create test specs, search by various queries
- **Template Rendering**: Use real spec/plan/constitution files
- **Progress Tracking**: Create files with CANARY tokens, verify counts

### Manual Testing
- **Interactive Selection**: Test with multiple ambiguous matches
- **Error Handling**: Test missing spec, invalid ID, malformed query
- **Performance**: Time lookup and prompt generation (<2 seconds target)

### Benchmarks (Optional, for BENCHED status)
```go
func BenchmarkCANARY_CBIN_133_Engine_Levenshtein(b *testing.B) {
	for i := 0; i < b.N; i++ {
		matcher.CalculateLevenshtein("UserAuthentication", "user auth")
	}
}

func BenchmarkCANARY_CBIN_133_CLI_PromptGeneration(b *testing.B) {
	spec := setupBenchmarkSpec(b)
	for i := 0; i < b.N; i++ {
		renderImplementPrompt(spec, &ImplementFlags{Prompt: true})
	}
}
```

---

## Performance Considerations

### Lookup Performance
- **Target**: <1 second for <1000 specs
- **Optimization**: Cache spec directory list if repeated queries needed
- **Levenshtein**: O(n*m) complexity acceptable for short strings (<100 chars)

### Template Rendering
- **Target**: <2 seconds for full prompt generation
- **File I/O**: Reading 3-5 files (spec, plan, constitution, template) acceptable
- **Template Parsing**: Parse template once, reuse if generating multiple prompts

### Memory Usage
- **Target**: <100MB for fuzzy matching all specs
- **Trade-off**: Load all spec directories into memory for matching (acceptable for <1000 specs)

---

## Security Considerations

### Input Validation
- **Query Sanitization**: Prevent directory traversal (../ sequences)
- **File Path Validation**: Ensure spec files are within `.canary/specs/`
- **Template Injection**: Go's `text/template` auto-escapes by default

### File Access
- **Read-Only**: Command only reads files, never writes
- **Permissions**: Respect filesystem permissions for spec/plan files

---

## Complexity Justification (Article V)

### Levenshtein Distance Algorithm
**Complexity**: O(n*m) where n and m are string lengths
**Justification**: Essential for fuzzy matching with typo tolerance. No simpler algorithm provides accurate edit distance. Limited to strings <100 chars.

### Interactive Selection UI
**Complexity**: Synchronous user input loop
**Justification**: Required for disambiguating multiple matches. Follows standard CLI pattern (select from numbered list).

### Template System
**Complexity**: Go's text/template with 10+ variables
**Justification**: Essential for generating comprehensive prompts with spec/plan/constitution content. Standard library solution (Article V: prefer standard library).

**No Unnecessary Complexity**: All complexity serves spec requirements directly.

---

## Constitutional Compliance Final Check

- ✅ **Article I**: CANARY tokens placed at all implementation points
- ✅ **Article II**: Spec is technology-agnostic, focuses on user needs
- ✅ **Article III**: Sub-features split by aspect (CLI, Engine, API)
- ✅ **Article IV**: Tests written FIRST (Phase 1), implementation SECOND (Phase 2)
- ✅ **Article V**: Standard library only (text/template, strings, os, filepath)
- ✅ **Article VI**: Integration tests use real filesystem and specs
- ✅ **Article VII**: All tokens have UPDATED field, STATUS progression documented

**Plan Status**: Ready for Implementation ✅

---

## Next Steps

1. **Begin Implementation**: Start with Phase 1 (Test Creation)
2. **Follow TDD**: Red (failing tests) → Green (passing implementation) → Refactor
3. **Update Tokens**: Progress STATUS from STUB → IMPL → TESTED
4. **Verify**: Run `/canary.scan` and `/canary.verify` after completion

---

**Implementation Guidance Generated:** 2025-10-16
**Plan Version:** 1.0
**Status:** APPROVED FOR IMPLEMENTATION
