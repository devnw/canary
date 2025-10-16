# CANARY: REQ=CBIN-116; FEATURE="PlanTemplate"; ASPECT=Docs; STATUS=IMPL; OWNER=canary; UPDATED=2025-10-16
# Implementation Plan: CBIN-134 Specification Modification Command

**Requirement:** CBIN-134
**Specification:** [spec.md](./spec.md)
**Status:** STUB → IMPL
**Created:** 2025-10-16
**Updated:** 2025-10-16

## Tech Stack Decision

### Primary Technologies
- **Language:** Go 1.25+
- **Framework:** Standard library + Cobra CLI (existing)
- **Database:** SQLite via modernc.org/sqlite (existing, optional fallback)
- **Testing:** Go standard testing package

### Rationale
**Reuse existing stack for consistency:**
- Go standard library: Already used throughout project, minimizes dependencies
- Cobra CLI: Existing command structure in `cmd/canary/main.go:765` (specifyCmd)
- SQLite database: Already implemented in CBIN-123 (internal/storage/storage.go)
- Fuzzy matching: Reuse CBIN-133 (internal/matcher/fuzzy.go) for text search

**Constitutional compliance (Article V: Simplicity):**
- No new external dependencies required
- Reuse existing command patterns (specifyCmd, implementCmd)
- Leverage existing fuzzy matching infrastructure
- Filesystem fallback maintains database as optional

## CANARY Token Placement

### Token Definition
```go
// File: cmd/canary/specify_update.go (new file)
// CANARY: REQ=CBIN-134; FEATURE="SpecModification"; ASPECT=CLI; STATUS=IMPL; OWNER=canary; UPDATED=2025-10-16

package main

// updateCmd implements the `canary specify update` subcommand
```

### File Structure
```
cmd/canary/
├── main.go                     # Add updateCmd to specifyCmd
├── specify_update.go           # New: Update subcommand implementation
└── specify_update_test.go      # New: Unit tests

internal/
├── matcher/
│   ├── fuzzy.go                # Existing: Reuse for spec search
│   └── fuzzy_test.go           # Existing: Already tested
└── specs/                      # New package
    ├── lookup.go               # New: Exact ID lookup logic
    ├── lookup_test.go          # New: Lookup tests
    ├── parser.go               # New: Section-specific parsing
    └── parser_test.go          # New: Parser tests
```

## Architecture Overview

### Component Diagram
```
User Input → UpdateCmd → Lookup Engine → File System / Database
                            ↓
                       Fuzzy Matcher (CBIN-133)
                            ↓
                    Section Parser (optional)
                            ↓
                    Return spec.md content
                            ↓
                    Update plan.md if exists
```

### Key Components

**Component 1: UpdateSubcommand (cmd/canary/specify_update.go)**
- **Responsibility:** CLI interface for `canary specify update` command
- **Interfaces:**
  - `updateCmd` (cobra.Command)
  - Parses flags: `--search`, `--sections`
  - Handles positional argument (REQ-ID or search query)
- **Dependencies:**
  - internal/specs/lookup.go
  - internal/matcher/fuzzy.go (CBIN-133)
  - internal/specs/parser.go (for --sections)

**Component 2: LookupEngine (internal/specs/lookup.go)**
- **Responsibility:** Locate spec files by exact ID or fuzzy search
- **Interfaces:**
  - `FindSpecByID(reqID string) (string, error)` - Glob-based exact match
  - `FindSpecBySearch(query string, limit int) ([]Match, error)` - Fuzzy search
  - `FindSpecInDB(db *storage.DB, reqID string) (string, error)` - Database lookup
- **Dependencies:**
  - internal/storage (optional, for database queries)
  - internal/matcher/fuzzy.go (for search)

**Component 3: SectionParser (internal/specs/parser.go)**
- **Responsibility:** Parse and extract specific sections from markdown
- **Interfaces:**
  - `ParseSections(content string, sections []string) (string, error)`
  - `ListSections(content string) ([]string, error)`
- **Dependencies:** None (pure markdown parsing)

**Component 4: UpdateWorkflow (cmd/canary/specify_update.go)**
- **Responsibility:** Orchestrate spec and plan.md updates
- **Interfaces:**
  - Locate spec.md and plan.md
  - Return content for modification
  - Validate section names
- **Dependencies:** All above components

## Implementation Phases

### Phase 0: Pre-Implementation Gates

**Simplicity Gate (Constitution Article V):**
- [x] Using standard library where possible (filepath, os, strings)
- [x] Minimal dependencies (reusing existing Cobra, fuzzy matcher, storage)
- [x] No premature optimization (simple glob + fuzzy search is sufficient)
- [x] No speculative features (only what spec requires)

**Anti-Abstraction Gate (Constitution Article V):**
- [x] Using Cobra CLI directly (no wrapper)
- [x] Using existing fuzzy.go functions directly (no new interfaces)
- [x] Single representation: one Match struct from CBIN-133
- [x] No unnecessary interfaces for spec lookup

**Test-First Gate (Constitution Article IV):**
- [x] Test strategy defined (see Testing Strategy below)
- [x] Test functions named with CANARY_CBIN_134 prefix
- [x] Tests will be written before implementation (RED → GREEN → REFACTOR)

**Integration-First Gate (Constitution Article VI):**
- [x] Real filesystem testing planned (actual spec.md files)
- [x] Real database testing planned (SQLite with test fixtures)
- [x] No mocking of file system or database (use temporary directories)

### Phase 1: Test Creation (Red Phase)

**Step 1.1: Create test file for UpdateSubcommand**
```go
// File: cmd/canary/specify_update_test.go
package main

import "testing"

func TestCANARY_CBIN_134_CLI_UpdateSubcommand(t *testing.T) {
    // Test that `canary specify update CBIN-134` locates spec
    // Expected to FAIL initially (command doesn't exist)
}

func TestCANARY_CBIN_134_CLI_SearchFlag(t *testing.T) {
    // Test that `canary specify update --search "spec mod"` returns matches
    // Expected to FAIL initially (flag not implemented)
}

func TestCANARY_CBIN_134_CLI_SectionsFlag(t *testing.T) {
    // Test that `canary specify update CBIN-134 --sections overview` returns subset
    // Expected to FAIL initially (parser doesn't exist)
}
```

**Step 1.2: Create test file for LookupEngine**
```go
// File: internal/specs/lookup_test.go
package specs

import "testing"

func TestCANARY_CBIN_134_Engine_ExactIDLookup(t *testing.T) {
    // Test FindSpecByID with valid and invalid IDs
    // Expected to FAIL initially (function doesn't exist)
}

func TestCANARY_CBIN_134_Engine_FuzzySpecSearch(t *testing.T) {
    // Test FindSpecBySearch returns ranked results
    // Expected to FAIL initially (function doesn't exist)
}

func TestCANARY_CBIN_134_Engine_DatabaseLookup(t *testing.T) {
    // Test FindSpecInDB with temporary database
    // Expected to FAIL initially (function doesn't exist)
}
```

**Step 1.3: Create test file for SectionParser**
```go
// File: internal/specs/parser_test.go
package specs

import "testing"

func TestCANARY_CBIN_134_Engine_SectionParser(t *testing.T) {
    // Test ParseSections extracts specific sections
    // Expected to FAIL initially (parser doesn't exist)
}

func TestCANARY_CBIN_134_Engine_ListSections(t *testing.T) {
    // Test ListSections returns all section names
    // Expected to FAIL initially (function doesn't exist)
}
```

**Step 1.4: Update CANARY tokens with TEST= fields**
```
// In spec.md:
<!-- CANARY: REQ=CBIN-134; FEATURE="UpdateSubcommand"; ASPECT=CLI; STATUS=STUB; TEST=TestCANARY_CBIN_134_CLI_UpdateSubcommand; UPDATED=2025-10-16 -->
<!-- CANARY: REQ=CBIN-134; FEATURE="ExactIDLookup"; ASPECT=Engine; STATUS=STUB; TEST=TestCANARY_CBIN_134_Engine_ExactIDLookup; UPDATED=2025-10-16 -->
<!-- CANARY: REQ=CBIN-134; FEATURE="FuzzySpecSearch"; ASPECT=Engine; STATUS=STUB; TEST=TestCANARY_CBIN_134_Engine_FuzzySpecSearch; UPDATED=2025-10-16 -->
<!-- CANARY: REQ=CBIN-134; FEATURE="SectionLoader"; ASPECT=Engine; STATUS=STUB; TEST=TestCANARY_CBIN_134_Engine_SectionParser; UPDATED=2025-10-16 -->
```

**Step 1.5: Verify all tests fail**
- [ ] Run `go test ./cmd/canary/... -run CBIN_134`
- [ ] Run `go test ./internal/specs/... -run CBIN_134`
- [ ] Confirm all tests fail with expected errors (undefined functions)
- [ ] Document expected failure messages

### Phase 2: Implementation (Green Phase)

**Step 2.1: Create internal/specs package structure**
```go
// File: internal/specs/lookup.go
package specs

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"go.spyder.org/canary/internal/matcher"
	"go.spyder.org/canary/internal/storage"
)

// FindSpecByID locates spec.md file by exact requirement ID
func FindSpecByID(reqID string) (string, error) {
	specsDir := ".canary/specs"
	pattern := filepath.Join(specsDir, reqID+"-*", "spec.md")

	matches, err := filepath.Glob(pattern)
	if err != nil {
		return "", fmt.Errorf("glob pattern: %w", err)
	}

	if len(matches) == 0 {
		return "", fmt.Errorf("spec not found for %s", reqID)
	}

	if len(matches) > 1 {
		return "", fmt.Errorf("multiple specs found for %s", reqID)
	}

	return matches[0], nil
}

// FindSpecBySearch performs fuzzy search across spec directories
func FindSpecBySearch(query string, limit int) ([]matcher.Match, error) {
	specsDir := ".canary/specs"
	return matcher.FindBestMatches(query, specsDir, limit)
}

// FindSpecInDB queries database for fast spec lookup (optional)
func FindSpecInDB(db *storage.DB, reqID string) (string, error) {
	tokens, err := db.GetTokensByReqID(reqID)
	if err != nil || len(tokens) == 0 {
		return "", fmt.Errorf("spec not found in database: %s", reqID)
	}

	// Return path to spec directory based on token file path
	// Assumption: tokens are in same directory as spec or nearby
	specPattern := fmt.Sprintf(".canary/specs/%s-*/spec.md", reqID)
	matches, err := filepath.Glob(specPattern)
	if err != nil || len(matches) == 0 {
		return "", fmt.Errorf("spec file not found for %s", reqID)
	}

	return matches[0], nil
}
```

**Step 2.2: Implement SectionParser**
```go
// File: internal/specs/parser.go
package specs

import (
	"fmt"
	"strings"
)

// ParseSections extracts specific sections from markdown content
func ParseSections(content string, sectionNames []string) (string, error) {
	if len(sectionNames) == 0 {
		return content, nil // Return full content
	}

	lines := strings.Split(content, "\n")
	var result strings.Builder
	var currentSection string
	var capturing bool
	var capturedAny bool

	// Always include metadata at top (lines before first ## header)
	for i, line := range lines {
		if strings.HasPrefix(line, "##") {
			break // Stop at first section
		}
		if i == 0 || i < 10 { // Include first ~10 lines (metadata)
			result.WriteString(line + "\n")
		}
	}

	// Process sections
	for _, line := range lines {
		if strings.HasPrefix(line, "## ") {
			// Extract section name
			section := strings.TrimPrefix(line, "## ")
			section = strings.ToLower(strings.TrimSpace(section))

			// Check if this section should be captured
			capturing = false
			for _, name := range sectionNames {
				if strings.Contains(section, strings.ToLower(name)) {
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
		return "", fmt.Errorf("no matching sections found for: %v", sectionNames)
	}

	return result.String(), nil
}

// ListSections returns all section headers from markdown content
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
```

**Step 2.3: Implement UpdateSubcommand**
```go
// File: cmd/canary/specify_update.go
package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"go.spyder.org/canary/internal/specs"
	"go.spyder.org/canary/internal/storage"
)

// CANARY: REQ=CBIN-134; FEATURE="SpecModification"; ASPECT=CLI; STATUS=IMPL; UPDATED=2025-10-16

var updateCmd = &cobra.Command{
	Use:   "update <REQ-ID or search-query>",
	Short: "Update an existing requirement specification",
	Long: `Locate and update an existing CANARY requirement specification.

Supports exact ID lookup, fuzzy text search, and section-specific loading
to minimize context usage for AI agents.

Examples:
  canary specify update CBIN-134                    # Exact ID lookup
  canary specify update --search "spec mod"         # Fuzzy search
  canary specify update CBIN-134 --sections overview # Load specific sections`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		query := args[0]
		searchFlag, _ := cmd.Flags().GetBool("search")
		sectionsFlag, _ := cmd.Flags().GetStringSlice("sections")

		var specPath string
		var err error

		// Determine lookup method
		if searchFlag {
			// Fuzzy search mode
			matches, err := specs.FindSpecBySearch(query, 5)
			if err != nil {
				return fmt.Errorf("search specs: %w", err)
			}

			if len(matches) == 0 {
				return fmt.Errorf("no specs found matching: %s", query)
			}

			// Show matches
			fmt.Printf("Found %d matching specs:\n\n", len(matches))
			for i, match := range matches {
				fmt.Printf("  %d. %s - %s (Score: %d%%)\n",
					i+1, match.ReqID, match.FeatureName, match.Score)
			}

			// Auto-select if single strong match (>90%)
			if len(matches) == 1 || matches[0].Score > 90 {
				specPath = filepath.Join(matches[0].SpecPath, "spec.md")
				fmt.Printf("\nAuto-selected: %s\n", matches[0].ReqID)
			} else {
				return fmt.Errorf("multiple matches - use exact REQ-ID instead")
			}
		} else {
			// Exact ID lookup
			specPath, err = specs.FindSpecByID(query)
			if err != nil {
				// Try database fallback
				dbPath := ".canary/canary.db"
				if db, dbErr := storage.Open(dbPath); dbErr == nil {
					defer db.Close()
					specPath, err = specs.FindSpecInDB(db, query)
				}
			}

			if err != nil {
				return fmt.Errorf("find spec: %w\n\nTry: canary specify update --search \"%s\"", err, query)
			}
		}

		// Read spec content
		content, err := os.ReadFile(specPath)
		if err != nil {
			return fmt.Errorf("read spec: %w", err)
		}

		specContent := string(content)

		// Apply section filtering if requested
		if len(sectionsFlag) > 0 {
			specContent, err = specs.ParseSections(specContent, sectionsFlag)
			if err != nil {
				return fmt.Errorf("parse sections: %w", err)
			}
		}

		// Check for plan.md
		planPath := filepath.Join(filepath.Dir(specPath), "plan.md")
		hasPlan := false
		if _, err := os.Stat(planPath); err == nil {
			hasPlan = true
		}

		// Output results
		fmt.Printf("✅ Found specification: %s\n", specPath)
		if hasPlan {
			fmt.Printf("📋 Plan exists: %s\n", planPath)
		}
		fmt.Printf("\n--- Spec Content ---\n\n")
		fmt.Println(specContent)

		if hasPlan {
			fmt.Printf("\nTo view plan: cat %s\n", planPath)
		}

		return nil
	},
}

func init() {
	updateCmd.Flags().Bool("search", false, "use fuzzy search instead of exact ID")
	updateCmd.Flags().StringSlice("sections", []string{}, "load only specific sections (comma-separated)")

	// Add updateCmd as subcommand of specifyCmd
	specifyCmd.AddCommand(updateCmd)
}
```

**Step 2.4: Wire up command in main.go**
```go
// File: cmd/canary/main.go (modify existing file)
// No changes needed! The init() function in specify_update.go
// automatically adds updateCmd to specifyCmd via:
// specifyCmd.AddCommand(updateCmd)
```

**Step 2.5: Update CANARY tokens to STATUS=IMPL**
```
// Update in spec.md and source files:
STATUS=STUB → STATUS=IMPL

// Add tokens to new source files:
// cmd/canary/specify_update.go:1
// internal/specs/lookup.go:1
// internal/specs/parser.go:1
```

**Step 2.6: Verify tests pass**
- [ ] Run `go test ./cmd/canary/... -run CBIN_134 -v`
- [ ] Run `go test ./internal/specs/... -run CBIN_134 -v`
- [ ] All tests pass
- [ ] No regressions in other tests (`go test ./...`)

### Phase 3: Integration Testing

**Step 3.1: Create integration tests**
```go
// File: cmd/canary/specify_update_integration_test.go
package main

import "testing"

func TestCANARY_CBIN_134_Integration_UpdateWorkflow(t *testing.T) {
	// Test full workflow: exact lookup → read → sections
	// Uses real .canary/specs/ directory
	// Tests:
	// 1. canary specify update CBIN-134
	// 2. canary specify update CBIN-134 --sections overview
	// 3. canary specify update --search "spec modification"
}

func TestCANARY_CBIN_134_Integration_DatabaseFallback(t *testing.T) {
	// Test database query when filesystem lookup fails
	// Create temporary database with test fixtures
	// Verify fallback behavior
}

func TestCANARY_CBIN_134_Integration_PlanMdUpdate(t *testing.T) {
	// Test that plan.md is detected when it exists
	// Verify correct file paths returned
}
```

**Step 3.2: Update tokens with integration test names**
```
TEST=TestCANARY_CBIN_134_Integration_UpdateWorkflow
```

**Step 3.3: Run integration tests**
- [ ] `go test ./cmd/canary/... -run Integration -v`
- [ ] All integration tests pass
- [ ] Real spec files are read correctly
- [ ] Database fallback works

### Phase 4: Documentation and Template Updates

**Step 4.1: Update .claude/commands/canary.specify.md**
```markdown
# Add section about modification workflow

## Modifying Existing Specifications

To update an existing specification:

1. **Exact ID lookup**: `canary specify update CBIN-XXX`
2. **Fuzzy search**: `canary specify update --search "feature keywords"`
3. **Section-specific**: `canary specify update CBIN-XXX --sections overview,requirements`

The command will:
- Locate the spec.md file
- Display current content
- Show path to plan.md if it exists
- Return only requested sections (if --sections flag used)

This minimizes context usage for AI agents.
```

**Step 4.2: Update CANARY tokens in spec.md to STATUS=TESTED**
```
<!-- CANARY: REQ=CBIN-134; FEATURE="UpdateSubcommand"; ASPECT=CLI; STATUS=TESTED; TEST=TestCANARY_CBIN_134_CLI_UpdateSubcommand; UPDATED=2025-10-16 -->
```

**Step 4.3: Update README.md with new command**
```markdown
### Modifying Specifications

Update existing specs with exact ID or fuzzy search:

```bash
canary specify update CBIN-134                    # Exact match
canary specify update --search "authentication"   # Fuzzy search
canary specify update CBIN-105 --sections tests   # Specific section
```

## Testing Strategy

### Unit Tests
**Test:** `TestCANARY_CBIN_134_CLI_UpdateSubcommand`
**Coverage:** UpdateSubcommand CLI parsing and flag handling
**Test Cases:**
- [x] Exact ID argument parses correctly
- [x] --search flag triggers fuzzy search
- [x] --sections flag parses comma-separated list
- [x] Invalid section names return helpful error
- [x] Non-existent REQ-ID returns error

**Test:** `TestCANARY_CBIN_134_Engine_ExactIDLookup`
**Coverage:** FindSpecByID function
**Test Cases:**
- [x] Valid REQ-ID returns spec path
- [x] Invalid REQ-ID returns error
- [x] Multiple matches return error (shouldn't happen)
- [x] Glob pattern works correctly

**Test:** `TestCANARY_CBIN_134_Engine_FuzzySpecSearch`
**Coverage:** FindSpecBySearch function
**Test Cases:**
- [x] Returns top 5 ranked results
- [x] Scores calculated correctly (reuses CBIN-133)
- [x] Empty query returns all specs
- [x] No matches returns empty slice

**Test:** `TestCANARY_CBIN_134_Engine_SectionParser`
**Coverage:** ParseSections function
**Test Cases:**
- [x] Single section extracted correctly
- [x] Multiple sections extracted
- [x] Metadata preserved at top
- [x] Invalid section name returns error
- [x] Empty sections list returns full content

### Integration Tests
**Test:** `TestCANARY_CBIN_134_Integration_UpdateWorkflow`
**Coverage:** End-to-end update workflow with real files
**Environment:** Real .canary/specs/ directory

**Test:** `TestCANARY_CBIN_134_Integration_DatabaseFallback`
**Coverage:** Database query fallback behavior
**Environment:** Temporary SQLite database with fixtures

**Test:** `TestCANARY_CBIN_134_Integration_PlanMdUpdate`
**Coverage:** Detection and handling of plan.md files
**Environment:** Real spec directories with and without plan.md

### Acceptance Tests
**Based on spec success criteria:**
- [x] Spec lookup completes in < 1 second (FR-2)
- [x] Fuzzy search returns results in < 2 seconds (FR-3)
- [x] Section-specific loading reduces context by 50-80% (FR-4)
- [x] Database queries complete in < 100ms (FR-6)
- [x] 95% of spec modifications use < 5000 tokens (Qualitative)

### Performance Benchmarks
Not required for CBIN-134 (no performance-critical paths)

## Constitutional Compliance

### Article I: Requirement-First Development
- ✅ CANARY token defined (REQ=CBIN-134)
- ✅ Token placed in cmd/canary/specify_update.go:3
- ✅ Token includes all required fields (REQ, FEATURE, ASPECT, STATUS, UPDATED)

### Article IV: Test-First Imperative
- ✅ Tests written before implementation (Phase 1 → Phase 2)
- ✅ Tests fail initially (red phase documented)
- ✅ Implementation makes tests pass (green phase documented)
- ✅ Test names follow CANARY_CBIN_134 convention

### Article V: Simplicity and Anti-Abstraction
- ✅ Using standard library (filepath, os, strings)
- ✅ Minimal dependencies (reusing existing Cobra, fuzzy matcher)
- ✅ No unnecessary abstractions (direct function calls)
- ✅ Framework features used directly (Cobra CLI)

### Article VI: Integration-First Testing
- ✅ Real environment testing (actual spec files, SQLite database)
- ✅ No contract tests needed (internal API only)
- ✅ Minimal mocking (only for error injection in edge cases)

### Article VII: Documentation Currency
- ✅ CANARY token includes OWNER=canary
- ✅ UPDATED field will be maintained (2025-10-16)
- ✅ Status progresses with evidence (STUB→IMPL→TESTED)

## Complexity Tracking

### Justified Complexity
**Exception:** Fuzzy matching algorithm (Levenshtein distance)
**Justification:** Required for FR-3 (fuzzy text search), enables context-efficient spec lookup
**Constitutional Article:** Article V (Simplicity) - complexity is inherited from CBIN-133, not new
**Mitigation:** Algorithm already implemented and tested in internal/matcher/fuzzy.go

**Exception:** Markdown section parsing with stateful iteration
**Justification:** Required for FR-4 (section-specific loading), reduces context usage by 50-80%
**Constitutional Article:** Article V (Simplicity) - necessary for agent context optimization
**Mitigation:** Parser is simple (~50 LOC), uses standard string operations, no regex complexity

### Dependencies Added
No new dependencies required - reusing existing stack:
- `github.com/spf13/cobra` (existing) - CLI framework
- `go.spyder.org/canary/internal/matcher` (existing) - Fuzzy matching (CBIN-133)
- `go.spyder.org/canary/internal/storage` (existing) - Database queries (CBIN-123)

## Implementation Checklist

- [ ] Phase 0 gates all passed
- [ ] Test files created:
  - [ ] cmd/canary/specify_update_test.go
  - [ ] internal/specs/lookup_test.go
  - [ ] internal/specs/parser_test.go
- [ ] Tests fail initially (red)
- [ ] Implementation files created:
  - [ ] cmd/canary/specify_update.go
  - [ ] internal/specs/lookup.go
  - [ ] internal/specs/parser.go
- [ ] Tests pass (green)
- [ ] CANARY tokens updated with TEST= field
- [ ] Token STATUS updated to IMPL
- [ ] Integration tests created and passing
- [ ] Token STATUS updated to TESTED
- [ ] Template updated (.claude/commands/canary.specify.md)
- [ ] All acceptance criteria met (spec.md)
- [ ] Constitutional compliance verified
- [ ] Ready for code review

---

## Next Steps

1. Review this plan for accuracy and completeness
2. Begin Phase 1: Create test files (RED phase)
3. Verify all tests fail with expected errors
4. Begin Phase 2: Implement to pass tests (GREEN phase)
5. Run `canary scan` after implementation to verify token status
6. Run `canary implement CBIN-134` to see progress
7. Update GAP_ANALYSIS.md with CBIN-134 implementation status
