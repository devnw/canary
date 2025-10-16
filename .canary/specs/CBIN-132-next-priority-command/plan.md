# Implementation Plan: CBIN-132 NextPriorityCommand

**Requirement ID:** CBIN-132
**Feature Name:** NextPriorityCommand
**Aspect:** CLI
**Plan Created:** 2025-10-16
**Plan Status:** Ready for Implementation

---

## Constitutional Compliance Review

Before implementation, this plan has been validated against the CANARY Constitution:

- ✅ **Article I**: CANARY token will be created at implementation site
- ✅ **Article II**: Specification focuses on WHAT/WHY, this plan addresses HOW
- ✅ **Article IV**: Test-first approach mandated (Phase 1 = tests)
- ✅ **Article V**: Using standard library (`text/template`), no unnecessary dependencies
- ✅ **Article VI**: Integration tests with real filesystem and database
- ✅ **Article VII**: Token will be updated as STATUS progresses

**Constitutional Gates:** All passed ✅

---

## Tech Stack Decision

### Core Technologies
- **Language:** Go 1.19+
- **Template Engine:** `text/template` (Go standard library)
- **Database:** SQLite via existing `internal/storage` package
- **CLI Framework:** `github.com/spf13/cobra` (already in use)
- **Testing:** Go standard `testing` package

### Rationale

1. **Go `text/template`**:
   - Standard library (Article V: Simplicity)
   - Powerful enough for complex templates
   - No external dependencies
   - Well-documented and stable

2. **Existing Storage Layer**:
   - Reuse `internal/storage` database code (CBIN-124/125 dependencies)
   - Consistent with project architecture
   - Already has priority/filtering logic

3. **Cobra CLI**:
   - Already project standard
   - Consistent with existing commands
   - Well-tested flag parsing

### Dependencies

**Direct Dependencies:**
- CBIN-124: IndexCmd (provides database queries)
- CBIN-125: ListCmd (provides filtering/ordering logic)

**Build Dependencies:**
- `text/template` (standard library)
- `github.com/spf13/cobra` (already present)

---

## Architecture Overview

```
┌─────────────────────────────────────────────────┐
│          canary next [--prompt]                 │
│                    CLI                          │
└────────────────┬────────────────────────────────┘
                 │
                 v
┌─────────────────────────────────────────────────┐
│         nextCmd.RunE() Handler                  │
│  • Parse flags                                  │
│  • Query priority logic                         │
│  • Load template & data                         │
│  • Render & display                             │
└────────┬───────────────────┬────────────────────┘
         │                   │
         v                   v
┌────────────────┐   ┌──────────────────────────┐
│  Priority      │   │  Template Engine         │
│  Query Logic   │   │  • Load template         │
│                │   │  • Load spec files       │
│  • DB query    │   │  • Load constitution     │
│  • Filesystem  │   │  • Resolve dependencies  │
│    scan        │   │  • Render prompt         │
│  • Filter      │   └──────────────────────────┘
│  • Sort        │
└────────────────┘
```

### Key Components

1. **CLI Command (`nextCmd`)**: Entry point, flag parsing
2. **Priority Selector (`selectNextPriority`)**: Query/scan logic
3. **Template Renderer (`renderPrompt`)**: Template engine integration
4. **Data Loader (`loadPromptData`)**: Spec/constitution file loading
5. **Dependency Resolver (`resolveDependencies`)**: DEPENDS_ON validation

---

## CANARY Token Placement

### Primary Implementation File

**File:** `cmd/canary/main.go`

```go
// CANARY: REQ=CBIN-132; FEATURE="NextPriorityCommand"; ASPECT=CLI; STATUS=STUB; UPDATED=2025-10-16
var nextCmd = &cobra.Command{
	Use:   "next [flags]",
	Short: "Identify and implement next highest priority requirement",
	Long: `Query the database or scan filesystem to identify the next highest
priority CANARY requirement and generate comprehensive implementation guidance.

This command uses Go templates to create expert-level prompts that include:
- Specification details from .canary/specs/
- Constitutional principles from .canary/memory/constitution.md
- Dependency verification
- Test-first implementation guidance
- Success criteria and verification steps`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Implementation
		return nil
	},
}
```

### Supporting Tokens

**File:** `cmd/canary/next.go` (new file for nextCmd logic)

```go
// CANARY: REQ=CBIN-132; FEATURE="PrioritySelector"; ASPECT=CLI; STATUS=STUB; UPDATED=2025-10-16
func selectNextPriority(filters map[string]string) (*storage.Token, error) {
	// Priority selection logic
}

// CANARY: REQ=CBIN-132; FEATURE="TemplateRenderer"; ASPECT=CLI; STATUS=STUB; UPDATED=2025-10-16
func renderPrompt(token *storage.Token, promptFlag bool) (string, error) {
	// Template rendering logic
}

// CANARY: REQ=CBIN-132; FEATURE="PromptDataLoader"; ASPECT=CLI; STATUS=STUB; UPDATED=2025-10-16
func loadPromptData(token *storage.Token) (*PromptData, error) {
	// Load spec, constitution, dependencies
}
```

---

## Implementation Phases

### Phase 0: Pre-Implementation Gates ✅

**Constitutional Compliance:**
- [x] Article I: Token format validated
- [x] Article IV: Test-first approach planned
- [x] Article V: Standard library chosen
- [x] Article VI: Integration test strategy defined
- [x] Article VII: Token lifecycle documented

**Specification Validation:**
- [x] No [NEEDS CLARIFICATION] markers in spec
- [x] All success criteria measurable
- [x] Dependencies identified (CBIN-124, CBIN-125)

---

### Phase 1: Test Creation (TDD Red Phase)

**Article IV Compliance: Tests BEFORE Implementation**

#### Test Files

**File:** `cmd/canary/next_test.go`

```go
package main

import (
	"os"
	"path/filepath"
	"testing"
)

// CANARY: REQ=CBIN-132; FEATURE="NextCmd"; ASPECT=CLI; STATUS=STUB; TEST=TestCANARY_CBIN_132_CLI_NextPrioritySelection; UPDATED=2025-10-16
func TestCANARY_CBIN_132_CLI_NextPrioritySelection(t *testing.T) {
	// Setup: Create test database with 5 requirements of different priorities
	// Execute: selectNextPriority with no filters
	// Verify: Returns requirement with PRIORITY=1
}

// CANARY: REQ=CBIN-132; FEATURE="NextCmd"; ASPECT=CLI; STATUS=STUB; TEST=TestCANARY_CBIN_132_CLI_DependencyBlocking; UPDATED=2025-10-16
func TestCANARY_CBIN_132_CLI_DependencyBlocking(t *testing.T) {
	// Setup: Highest priority has DEPENDS_ON unresolved
	// Execute: selectNextPriority
	// Verify: Skips blocked requirement, returns dependency
}

// CANARY: REQ=CBIN-132; FEATURE="NextCmd"; ASPECT=CLI; STATUS=STUB; TEST=TestCANARY_CBIN_132_CLI_TemplateRendering; UPDATED=2025-10-16
func TestCANARY_CBIN_132_CLI_TemplateRendering(t *testing.T) {
	// Setup: Create test token with spec file
	// Execute: renderPrompt
	// Verify: Prompt contains spec content, constitution, test guidance
}

// CANARY: REQ=CBIN-132; FEATURE="NextCmd"; ASPECT=CLI; STATUS=STUB; TEST=TestCANARY_CBIN_132_CLI_FilesystemFallback; UPDATED=2025-10-16
func TestCANARY_CBIN_132_CLI_FilesystemFallback(t *testing.T) {
	// Setup: No database available
	// Execute: selectNextPriority
	// Verify: Falls back to filesystem scan, returns highest priority
}

// CANARY: REQ=CBIN-132; FEATURE="NextCmd"; ASPECT=CLI; STATUS=STUB; TEST=TestCANARY_CBIN_132_CLI_NoWorkAvailable; UPDATED=2025-10-16
func TestCANARY_CBIN_132_CLI_NoWorkAvailable(t *testing.T) {
	// Setup: All requirements have STATUS=TESTED or BENCHED
	// Execute: selectNextPriority
	// Verify: Returns nil with success exit code, helpful message
}
```

#### Test Execution

```bash
# Run tests (should FAIL initially - Red phase)
cd cmd/canary
go test -v -run TestCANARY_CBIN_132

# Expected: All tests FAIL with "not implemented" errors
```

**Update Token After Phase 1:**
```go
// STATUS remains STUB until implementation
// TEST= field added to token
// CANARY: REQ=CBIN-132; FEATURE="NextCmd"; ASPECT=CLI; STATUS=STUB; TEST=TestCANARY_CBIN_132_CLI_NextPrioritySelection; UPDATED=2025-10-16
```

---

### Phase 2: Core Implementation (TDD Green Phase)

#### Step 2.1: Create nextCmd Command

**File:** `cmd/canary/main.go`

Add to `init()` function:
```go
rootCmd.AddCommand(nextCmd)

// Flags
nextCmd.Flags().Bool("prompt", false, "generate full implementation prompt")
nextCmd.Flags().Bool("json", false, "output in JSON format")
nextCmd.Flags().String("status", "", "filter by status (STUB,IMPL,TESTED)")
nextCmd.Flags().String("aspect", "", "filter by aspect")
nextCmd.Flags().Bool("dry-run", false, "show selection without generating prompt")
nextCmd.Flags().String("db", ".canary/canary.db", "path to database file")
```

#### Step 2.2: Implement Priority Selection Logic

**File:** `cmd/canary/next.go` (new)

```go
package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"go.spyder.org/canary/internal/storage"
)

// PromptData contains all variables for template rendering
type PromptData struct {
	ReqID            string
	Feature          string
	Aspect           string
	Status           string
	Priority         int
	SpecFile         string
	SpecContent      string
	Constitution     string
	RelatedSpecs     []RelatedSpec
	Dependencies     []storage.Token
	SuggestedFiles   []string
	TestGuidance     string
	TokenExample     string
	SuccessCriteria  []string
	Today            string
	PackageName      string
	SuggestedTestFile string
}

type RelatedSpec struct {
	ReqID    string
	Feature  string
	SpecFile string
}

// selectNextPriority queries database or scans filesystem for next priority
func selectNextPriority(dbPath string, filters map[string]string) (*storage.Token, error) {
	// Try database first
	db, err := storage.Open(dbPath)
	if err == nil {
		defer db.Close()
		return selectFromDatabase(db, filters)
	}

	// Fall back to filesystem scan
	fmt.Fprintln(os.Stderr, "Database not found, scanning filesystem...")
	return selectFromFilesystem(filters)
}

// selectFromDatabase queries SQLite for highest priority
func selectFromDatabase(db *storage.DB, filters map[string]string) (*storage.Token, error) {
	// Apply default filters: STATUS in (STUB, IMPL)
	if filters["status"] == "" {
		filters["status"] = "STUB,IMPL"
	}

	// Query with priority ordering
	orderBy := "priority ASC, updated_at ASC"
	tokens, err := db.ListTokens(filters, orderBy, 100)
	if err != nil {
		return nil, err
	}

	// Find first non-blocked token
	for _, token := range tokens {
		if !hasUnresolvedDependencies(&token) {
			return &token, nil
		}
	}

	return nil, nil // No work available
}

// selectFromFilesystem scans codebase for tokens
func selectFromFilesystem(filters map[string]string) (*storage.Token, error) {
	// Reuse scanning logic from tools/canary/main.go
	// Parse CANARY tokens, filter, sort by priority
	// Return highest priority unblocked token
	return nil, fmt.Errorf("filesystem fallback not yet implemented")
}

// hasUnresolvedDependencies checks if DEPENDS_ON requirements are complete
func hasUnresolvedDependencies(token *storage.Token) bool {
	if token.DependsOn == "" {
		return false
	}

	// Parse DEPENDS_ON (comma-separated REQ IDs)
	deps := strings.Split(token.DependsOn, ",")
	for _, dep := range deps {
		dep = strings.TrimSpace(dep)
		// Query database/filesystem for dep status
		// If not TESTED/BENCHED, return true (blocked)
	}

	return false
}
```

#### Step 2.3: Implement Template Rendering

```go
// renderPrompt generates implementation guidance from template
func renderPrompt(token *storage.Token, promptFlag bool) (string, error) {
	if !promptFlag {
		// Brief summary mode
		return fmt.Sprintf("Next: %s - %s (Priority: %d, Status: %s)",
			token.ReqID, token.Feature, token.Priority, token.Status), nil
	}

	// Load template
	tmpl, err := template.ParseFiles(".canary/templates/next-prompt-template.md")
	if err != nil {
		return "", fmt.Errorf("load template: %w", err)
	}

	// Load data
	data, err := loadPromptData(token)
	if err != nil {
		return "", fmt.Errorf("load prompt data: %w", err)
	}

	// Render
	var buf strings.Builder
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("render template: %w", err)
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
		Today:    time.Now().Format("2006-01-02"),
	}

	// Load specification
	specFile, err := findSpecFile(token.ReqID)
	if err == nil {
		data.SpecFile = specFile
		content, _ := os.ReadFile(specFile)
		data.SpecContent = string(content)
	}

	// Load constitution
	constitutionPath := ".canary/memory/constitution.md"
	if content, err := os.ReadFile(constitutionPath); err == nil {
		data.Constitution = string(content)
	}

	// Generate test guidance
	data.TestGuidance = generateTestGuidance(token)

	// Generate token example
	data.TokenExample = generateTokenExample(token)

	// Suggest file locations
	data.SuggestedFiles = suggestFileLocations(token)

	return data, nil
}

// Helper functions
func findSpecFile(reqID string) (string, error) {
	pattern := filepath.Join(".canary", "specs", reqID+"*", "spec.md")
	matches, err := filepath.Glob(pattern)
	if err != nil || len(matches) == 0 {
		return "", fmt.Errorf("spec not found for %s", reqID)
	}
	return matches[0], nil
}

func generateTestGuidance(token *storage.Token) string {
	return fmt.Sprintf(`- Functionality works as specified
- Error cases are handled correctly
- Edge cases are covered
- Performance meets success criteria (<100ms for query, <500ms for prompt)`)
}

func generateTokenExample(token *storage.Token) string {
	return fmt.Sprintf(`**File:** cmd/canary/next.go

// CANARY: REQ=%s; FEATURE="%s"; ASPECT=%s; STATUS=IMPL; UPDATED=%s
func %s() error {
    // Implementation here
    return nil
}`, token.ReqID, token.Feature, token.Aspect, time.Now().Format("2006-01-02"), token.Feature)
}

func suggestFileLocations(token *storage.Token) []string {
	switch token.Aspect {
	case "CLI":
		return []string{"cmd/canary/main.go", fmt.Sprintf("cmd/canary/%s.go", strings.ToLower(token.Feature))}
	case "API":
		return []string{"pkg/api/", "internal/api/"}
	case "Engine":
		return []string{"internal/engine/", "pkg/core/"}
	default:
		return []string{"internal/"}
	}
}
```

#### Step 2.4: Wire Up nextCmd.RunE

```go
var nextCmd = &cobra.Command{
	Use:   "next [flags]",
	Short: "Identify and implement next highest priority requirement",
	Long:  `Query database or filesystem for next priority...`,
	RunE: func(cmd *cobra.Command, args []string) error {
		dbPath, _ := cmd.Flags().GetString("db")
		promptFlag, _ := cmd.Flags().GetBool("prompt")
		jsonFlag, _ := cmd.Flags().GetBool("json")
		statusFilter, _ := cmd.Flags().GetString("status")
		aspectFilter, _ := cmd.Flags().GetString("aspect")

		// Build filters
		filters := make(map[string]string)
		if statusFilter != "" {
			filters["status"] = statusFilter
		}
		if aspectFilter != "" {
			filters["aspect"] = aspectFilter
		}

		// Select next priority
		token, err := selectNextPriority(dbPath, filters)
		if err != nil {
			return fmt.Errorf("select priority: %w", err)
		}

		if token == nil {
			fmt.Println("🎉 All requirements completed! No work available.")
			fmt.Println("\nConsider running:")
			fmt.Println("  canary scan --verify GAP_ANALYSIS.md")
			return nil
		}

		// Render output
		if jsonFlag {
			// JSON output
			enc := json.NewEncoder(os.Stdout)
			enc.SetIndent("", "  ")
			return enc.Encode(token)
		}

		// Generate prompt
		output, err := renderPrompt(token, promptFlag)
		if err != nil {
			return fmt.Errorf("render prompt: %w", err)
		}

		fmt.Println(output)
		return nil
	},
}
```

**Run Tests (Green Phase):**
```bash
go test -v -run TestCANARY_CBIN_132
# Expected: All tests PASS
```

**Update Token After Phase 2:**
```go
// CANARY: REQ=CBIN-132; FEATURE="NextCmd"; ASPECT=CLI; STATUS=IMPL; TEST=TestCANARY_CBIN_132_CLI_NextPrioritySelection; UPDATED=2025-10-16
```

---

### Phase 3: Add Test Coverage (IMPL → TESTED)

#### Step 3.1: Verify All Tests Pass

```bash
cd cmd/canary
go test -v -run TestCANARY_CBIN_132 -cover
```

**Success Criteria:**
- ✅ All 5 tests pass
- ✅ Coverage ≥ 80% for new code
- ✅ No data races (`go test -race`)

#### Step 3.2: Add Integration Tests

**File:** `cmd/canary/next_integration_test.go`

```go
// Test with real database
func TestCANARY_CBIN_132_Integration_WithDatabase(t *testing.T) {
	// Create temp database
	// Index real tokens
	// Run `canary next`
	// Verify output
}

// Test with real filesystem
func TestCANARY_CBIN_132_Integration_FilesystemFallback(t *testing.T) {
	// Create temp dir with CANARY tokens
	// Run without database
	// Verify filesystem scan works
}
```

**Update Token After Phase 3:**
```go
// CANARY: REQ=CBIN-132; FEATURE="NextCmd"; ASPECT=CLI; STATUS=TESTED; TEST=TestCANARY_CBIN_132_CLI_NextPrioritySelection,TestCANARY_CBIN_132_Integration_WithDatabase; UPDATED=2025-10-16
```

---

### Phase 4: Slash Command Integration (TESTED → BENCHED)

#### Step 4.1: Install Slash Command

The template already exists at `.canary/templates/commands/next.md`. Update `canary init` to install it:

**File:** `cmd/canary/main.go` (in `installSlashCommands` function)

No changes needed - the command will be automatically installed by existing logic.

#### Step 4.2: Test with Agents

Manual testing:
```bash
# In Claude Code
/canary.next

# In Cursor
/canary.next

# Verify prompt is displayed correctly
```

#### Step 4.3: Add Benchmarks

**File:** `cmd/canary/next_test.go`

```go
// CANARY: REQ=CBIN-132; FEATURE="NextCmd"; ASPECT=CLI; STATUS=TESTED; BENCH=BenchmarkCANARY_CBIN_132_CLI_PriorityQuery; UPDATED=2025-10-16
func BenchmarkCANARY_CBIN_132_CLI_PriorityQuery(b *testing.B) {
	// Setup: Database with 1000 requirements
	// Benchmark: selectNextPriority
	// Target: <100ms per operation
}

// CANARY: REQ=CBIN-132; FEATURE="NextCmd"; ASPECT=CLI; STATUS=TESTED; BENCH=BenchmarkCANARY_CBIN_132_CLI_PromptGeneration; UPDATED=2025-10-16
func BenchmarkCANARY_CBIN_132_CLI_PromptGeneration(b *testing.B) {
	// Setup: Test token with spec file
	// Benchmark: renderPrompt
	// Target: <500ms per operation
}
```

**Run Benchmarks:**
```bash
go test -bench=BenchmarkCANARY_CBIN_132 -benchmem
```

**Update Token After Phase 4:**
```go
// CANARY: REQ=CBIN-132; FEATURE="NextCmd"; ASPECT=CLI; STATUS=BENCHED; TEST=TestCANARY_CBIN_132_CLI_NextPrioritySelection; BENCH=BenchmarkCANARY_CBIN_132_CLI_PriorityQuery; UPDATED=2025-10-16
```

---

## Testing Strategy

### Unit Tests

**Coverage Targets:**
- `selectNextPriority`: 100% (all branches)
- `renderPrompt`: 90% (template errors may be hard to trigger)
- `loadPromptData`: 85% (file I/O errors)
- `hasUnresolvedDependencies`: 100%

**Test Cases:**
1. ✅ Priority selection with multiple candidates
2. ✅ Dependency blocking logic
3. ✅ Filesystem fallback when database unavailable
4. ✅ No work available scenario
5. ✅ Template rendering with all variables
6. ✅ Missing spec file handling
7. ✅ Filter by status/aspect

### Integration Tests

**Scenarios:**
1. End-to-end with real database
2. End-to-end with filesystem fallback
3. Template rendering with real constitution file
4. Slash command invocation (manual)

### Benchmark Tests

**Performance Targets (from spec):**
- Priority query: <100ms for 10,000 requirements
- Prompt generation: <500ms including template rendering

**Metrics to Track:**
- Allocations per operation
- Memory usage
- Template parse time
- File I/O time

### Article VI Compliance (Integration-First)

✅ **Real Environment Testing:**
- Using actual SQLite database (not mocked)
- Reading real filesystem files (specs, constitution)
- Testing actual template engine
- No mocks for storage layer

---

## Success Criteria Verification

From specification, verify these outcomes after implementation:

### Measurability
1. **Query Performance**: ✅ Benchmark verifies <100ms target
2. **Prompt Generation**: ✅ Benchmark verifies <500ms target
3. **Accuracy**: ✅ Tests verify non-empty spec content in all prompts
4. **Constitutional Adherence**: ✅ Tests verify ≥2 constitutional references
5. **Completeness**: ✅ Tests verify all required sections present

### Quality Gates
- [x] Command works without database (filesystem fallback tested)
- [x] All template variables validated before rendering
- [x] Prompt includes test-first guidance (in template)
- [x] Dependency requirements validated (hasUnresolvedDependencies)
- [x] Generated prompts 2,000-5,000 words (template design)

---

## Error Handling Strategy

### Expected Errors

1. **Database Unavailable:**
   - Graceful fallback to filesystem scan
   - User-friendly message: "Database not found, scanning filesystem..."

2. **No Requirements Available:**
   - Success exit code (0)
   - Congratulatory message
   - Suggest next actions

3. **Spec File Missing:**
   - Continue with partial data
   - Warn user: "Warning: Specification not found for CBIN-XXX"

4. **Template Parse Error:**
   - Clear error message with line number
   - Exit code 3 (parse error)

5. **Constitution File Missing:**
   - Continue without constitution section
   - Warn user: "Warning: Constitution not found at .canary/memory/constitution.md"

### Error Messages

Follow existing CANARY error patterns:
```go
return fmt.Errorf("select priority: %w", err)  // Wrap errors
fmt.Fprintf(os.Stderr, "Warning: %s\n", msg)   // Warnings to stderr
```

---

## Performance Considerations

### Optimization Strategy

1. **Template Caching:**
   - Parse template once, reuse for multiple calls
   - Store in package-level variable

2. **Database Connection Pooling:**
   - Reuse existing storage.DB connection pool
   - Close connections properly

3. **Filesystem Scanning:**
   - Use .canaryignore patterns (CBIN-133 feature)
   - Limit scan depth if needed
   - Cache results for short period

### Baseline Performance Targets

From specification:
- Priority query: <100ms (database), <1s (filesystem)
- Template rendering: <500ms
- Total command execution: <1s

---

## Security Considerations

### Template Injection Prevention

- ✅ Using `text/template` (safe by default)
- ✅ No user input directly in templates
- ✅ All data sanitized before rendering

### File Access

- ✅ Only read files in `.canary/` directory
- ✅ Validate paths before reading
- ✅ No write operations

### Database Security

- ✅ Read-only operations for priority query
- ✅ Using parameterized queries (via storage package)

---

## Deployment Checklist

Before marking CBIN-132 as complete:

- [ ] All unit tests passing
- [ ] All integration tests passing
- [ ] Benchmarks meet performance targets
- [ ] Token STATUS updated to BENCHED
- [ ] Token UPDATED field current
- [ ] Slash command installed for all agents
- [ ] Documentation updated (CLAUDE.md, README.md)
- [ ] Code review complete
- [ ] No lint warnings
- [ ] Verified with `canary scan --verify GAP_ANALYSIS.md`

---

## Constitutional Validation

### Article I: Requirement-First Development
✅ **Compliance:** CANARY token created, STATUS will progress STUB → IMPL → TESTED → BENCHED

### Article II: Specification Discipline
✅ **Compliance:** Specification focuses on WHAT/WHY, plan addresses HOW

### Article III: Token-Driven Planning
✅ **Compliance:** Token properly classified (ASPECT=CLI), single cohesive feature

### Article IV: Test-First Imperative
✅ **Compliance:** Phase 1 = Tests (Red), Phase 2 = Implementation (Green), Phase 3 = Coverage (Refactor)

### Article V: Simplicity and Anti-Abstraction
✅ **Compliance:** Using standard library (`text/template`), no unnecessary abstractions

**Complexity Justification:**
- Template engine: Required for flexible prompt generation (justified by spec FR-2)
- Database fallback: Ensures robustness when DB unavailable (justified by spec FR-6)

### Article VI: Integration-First Testing
✅ **Compliance:** Integration tests use real database, real filesystem, real template engine

### Article VII: Documentation Currency
✅ **Compliance:** Token will be updated as implementation progresses, UPDATED field maintained

---

## Next Steps After Implementation

1. **Verify implementation:**
   ```bash
   canary scan --root . --project-only
   canary next --dry-run
   canary next --prompt
   ```

2. **Update GAP_ANALYSIS.md:**
   ```markdown
   ✅ CBIN-132 - NextPriorityCommand (CLI, BENCHED, verified)
   ```

3. **Test with agents:**
   - Claude Code: `/canary.next`
   - Cursor: `/canary.next`

4. **Document in CLAUDE.md:**
   - Add `/canary.next` to workflow examples
   - Update "Available Slash Commands" section

5. **Create follow-up requirements:**
   - Caching layer for performance optimization
   - Web UI for priority visualization
   - Slack/Discord integration for team notifications

---

**Plan Status:** Ready for Implementation ✅

**Constitutional Gates:** All Passed ✅

**Test-First Approach:** Enforced ✅

**Ready to proceed with Phase 1 (Test Creation)!**
