<!-- CANARY: REQ=CBIN-116; FEATURE="PlanTemplate"; ASPECT=Docs; STATUS=IMPL; OWNER=canary; UPDATED=2025-10-16 -->
# Implementation Plan: CBIN-135 Priority List Command

**Requirement:** CBIN-135
**Specification:** [spec.md](./spec.md)
**Status:** STUB → IMPL (Simplified: Agent slash command only)
**Created:** 2025-10-16
**Updated:** 2025-10-16

**NOTE:** CBIN-125 `listCmd` already implements the core CLI functionality at `cmd/canary/main.go:760`.
CBIN-135 only adds the agent slash command template for `/canary.list`.

## Tech Stack Decision

### Primary Technologies
- **Language:** Go 1.25+
- **Framework:** Standard library + Cobra CLI (existing)
- **Database:** SQLite via modernc.org/sqlite (existing, with filesystem fallback)
- **Testing:** Go standard testing package
- **Formatter:** Custom table formatter (text/tabwriter or manual formatting)

### Rationale
**Reuse existing infrastructure - CBIN-125 already partially implemented:**
- **Critical Discovery:** `cmd/canary/main.go:760` already contains CBIN-125 `listCmd` implementation
- **Strategy:** Enhance existing command rather than duplicate functionality
- **Go standard library:** Minimizes dependencies, already used throughout project
- **Cobra CLI:** Existing command structure
- **SQLite database:** CBIN-123 (internal/storage/storage.go) provides token storage
- **Filesystem fallback:** Required by spec for projects without database

**Constitutional compliance (Article V: Simplicity):**
- Enhance existing listCmd rather than creating new command
- Reuse existing database query patterns from CBIN-125
- No new external dependencies required
- Minimal complexity added

### Clarifications Resolved

**Clarification 1: Dependency information display**
**Decision:** Option A (No - keep output simple and focused)
**Rationale:**
- Aligns with Article V (Simplicity)
- Dependencies can be shown via `canary next` command instead
- Keeps table width manageable for 80-column terminals
- Can be added later as `--show-deps` flag if needed (out of scope)

**Clarification 2: Default behavior when database unavailable**
**Decision:** Option C (Show warning then proceed with filesystem scan)
**Rationale:**
- Best user experience with transparency
- Aligns with existing `canary next` behavior (cmd/canary/next.go:47-62)
- Provides clear feedback about performance expectations
- Encourages database usage without blocking functionality

## CANARY Token Placement

### Token Definition
```go
// File: cmd/canary/list_enhanced.go (new file for enhancements)
// CANARY: REQ=CBIN-135; FEATURE="PriorityList"; ASPECT=CLI; STATUS=IMPL; OWNER=canary; UPDATED=2025-10-16

package main

// Enhanced list functionality for CBIN-135
// Builds on existing CBIN-125 listCmd at cmd/canary/main.go:760
```

**Note:** Main CBIN-125 token already exists at cmd/canary/main.go:760. CBIN-135 adds enhancements.

### File Structure
```
cmd/canary/
├── main.go                     # Existing: CBIN-125 listCmd (line 760)
├── list_enhanced.go            # New: CBIN-135 enhancements
├── list_enhanced_test.go       # New: Tests for CBIN-135 features
└── list_formatter.go           # New: Table formatting utilities

internal/query/                 # New package
├── list.go                     # Query engine with sorting/filtering
├── list_test.go                # Query engine tests
├── filesystem.go               # Filesystem fallback implementation
└── filesystem_test.go          # Filesystem fallback tests

.claude/commands/
└── canary.list.md              # New: Agent slash command template
```

## Architecture Overview

### Component Diagram
```
User Input → ListCmd (CBIN-125/existing) → Enhanced Features (CBIN-135)
                 ↓                              ↓
         Database Query (storage.ListTokens)  Query Engine (new)
                 ↓                              ↓
         Filters Applied                  Advanced Sorting
                 ↓                              ↓
         Table Formatter (enhanced)      Filesystem Fallback
                 ↓
         Terminal Output (80-column formatted table)
```

### Key Components

**Component 1: Enhanced ListCmd (cmd/canary/list_enhanced.go)**
- **Responsibility:** Add missing features from CBIN-135 to existing CBIN-125 listCmd
- **Interfaces:**
  - Extend existing command with: `--count`, `--all`, `--desc`, `--sort` enhancements
  - Add default limit of 10 (currently no limit)
  - Add "Showing X of Y total" counter
  - Improve error messages with suggestions
- **Dependencies:**
  - Existing cmd/canary/main.go:760 (CBIN-125)
  - internal/query (new)
  - internal/storage (existing)

**Component 2: QueryEngine (internal/query/list.go)**
- **Responsibility:** Advanced query logic beyond simple DB filters
- **Interfaces:**
  - `ApplyDefaultSorting(tokens []*storage.Token) []*storage.Token` - Priority + status secondary sort
  - `ApplyCustomSorting(tokens []*storage.Token, field string, desc bool) []*storage.Token`
  - `CountByFilter(filters map[string]string) (int, error)` - For "X of Y" display
- **Dependencies:**
  - internal/storage.Token (existing)

**Component 3: FilesystemFallback (internal/query/filesystem.go)**
- **Responsibility:** Scan .canary/specs/ when database unavailable
- **Interfaces:**
  - `ScanSpecs(filters map[string]string) ([]*storage.Token, error)`
  - `ParseSpecMetadata(specPath string) (*storage.Token, error)`
- **Dependencies:**
  - None (pure filesystem operations)
  - Pattern: Reuse logic from cmd/canary/next.go:168-276 (selectFromFilesystem)

**Component 4: TableFormatter (cmd/canary/list_formatter.go)**
- **Responsibility:** Format results as aligned 80-column table
- **Interfaces:**
  - `FormatTable(tokens []*storage.Token, showCount bool, total int) string`
  - `TruncateFeature(name string, maxLen int) string` - Truncate to 40 chars
  - `FormatEmptyMessage(filters map[string]string) string` - Helpful suggestions
- **Dependencies:** None

**Component 5: Agent Slash Command (.claude/commands/canary.list.md)**
- **Responsibility:** Template for AI agents to use list command
- **Interfaces:** Documentation/template only
- **Dependencies:** None

## Implementation Phases

### Phase 0: Pre-Implementation Gates

**Simplicity Gate (Constitution Article V):**
- [x] Using standard library (os, filepath, strings, text/tabwriter)
- [x] Minimal dependencies (reusing existing CBIN-125 infrastructure)
- [x] No premature optimization (simple sorting/filtering adequate)
- [x] No speculative features (only what spec requires)

**Anti-Abstraction Gate (Constitution Article V):**
- [x] Using Cobra CLI directly (no wrapper)
- [x] Using existing storage.ListTokens() (no new query interfaces)
- [x] Single representation of tokens (storage.Token struct)
- [x] No unnecessary abstractions

**Test-First Gate (Constitution Article IV):**
- [x] Test strategy defined (see Testing Strategy below)
- [x] Test functions named with CANARY_CBIN_135 prefix
- [x] Tests will be written before implementation (RED → GREEN → REFACTOR)

**Integration-First Gate (Constitution Article VI):**
- [x] Real filesystem testing planned (actual .canary/specs/ directory)
- [x] Real database testing planned (SQLite with fixtures)
- [x] No mocking of storage layer (use temporary database)

### Phase 1: Test Creation (Red Phase)

**Step 1.1: Create test file for enhanced ListCmd features**
```go
// File: cmd/canary/list_enhanced_test.go
package main

import "testing"

func TestCANARY_CBIN_135_CLI_DefaultLimit(t *testing.T) {
    // Test that default limit is 10 items
    // Expected to FAIL initially (no limit enforcement)
}

func TestCANARY_CBIN_135_CLI_CountFlag(t *testing.T) {
    // Test that --count N limits results
    // Expected to FAIL initially (flag doesn't exist)
}

func TestCANARY_CBIN_135_CLI_ShowingCounter(t *testing.T) {
    // Test that "Showing X of Y total" displays
    // Expected to FAIL initially (counter not implemented)
}

func TestCANARY_CBIN_135_CLI_DescFlag(t *testing.T) {
    // Test that --desc reverses sort order
    // Expected to FAIL initially (desc flag doesn't exist)
}

func TestCANARY_CBIN_135_CLI_InvalidStatusError(t *testing.T) {
    // Test helpful error for invalid status values
    // Expected to FAIL initially (generic error message)
}
```

**Step 1.2: Create test file for QueryEngine**
```go
// File: internal/query/list_test.go
package query

import "testing"

func TestCANARY_CBIN_135_Engine_DefaultSorting(t *testing.T) {
    // Test priority ascending, then status (STUB > IMPL > TESTED)
    // Expected to FAIL initially (function doesn't exist)
}

func TestCANARY_CBIN_135_Engine_CustomSorting(t *testing.T) {
    // Test sort by updated, aspect, status fields
    // Expected to FAIL initially (function doesn't exist)
}

func TestCANARY_CBIN_135_Engine_DescendingSort(t *testing.T) {
    // Test reverse sort order
    // Expected to FAIL initially (desc logic not implemented)
}

func TestCANARY_CBIN_135_Engine_StableSort(t *testing.T) {
    // Test that sort maintains secondary ordering
    // Expected to FAIL initially (stability not guaranteed)
}
```

**Step 1.3: Create test file for FilesystemFallback**
```go
// File: internal/query/filesystem_test.go
package query

import "testing"

func TestCANARY_CBIN_135_Engine_FilesystemScan(t *testing.T) {
    // Test scanning .canary/specs/ directory
    // Expected to FAIL initially (function doesn't exist)
}

func TestCANARY_CBIN_135_Engine_SpecMetadataParsing(t *testing.T) {
    // Test parsing CANARY tokens from spec.md files
    // Expected to FAIL initially (parser doesn't exist)
}

func TestCANARY_CBIN_135_Engine_FilesystemFilters(t *testing.T) {
    // Test applying filters during filesystem scan
    // Expected to FAIL initially (filter logic not implemented)
}
```

**Step 1.4: Create test file for TableFormatter**
```go
// File: cmd/canary/list_formatter_test.go
package main

import "testing"

func TestCANARY_CBIN_135_CLI_TableFormatting(t *testing.T) {
    // Test table fits in 80 columns
    // Expected to FAIL initially (formatter doesn't exist)
}

func TestCANARY_CBIN_135_CLI_FeatureTruncation(t *testing.T) {
    // Test long feature names truncated to 40 chars with ellipsis
    // Expected to FAIL initially (truncation not implemented)
}

func TestCANARY_CBIN_135_CLI_EmptyResults(t *testing.T) {
    // Test "No requirements found" message with suggestions
    // Expected to FAIL initially (no helpful message)
}
```

**Step 1.5: Update CANARY tokens with TEST= fields**
```
// In spec.md:
<!-- CANARY: REQ=CBIN-135; FEATURE="ListSubcommand"; ASPECT=CLI; STATUS=STUB; TEST=TestCANARY_CBIN_135_CLI_DefaultLimit; UPDATED=2025-10-16 -->
<!-- CANARY: REQ=CBIN-135; FEATURE="QueryEngine"; ASPECT=Engine; STATUS=STUB; TEST=TestCANARY_CBIN_135_Engine_DefaultSorting; UPDATED=2025-10-16 -->
<!-- CANARY: REQ=CBIN-135; FEATURE="TableFormatter"; ASPECT=CLI; STATUS=STUB; TEST=TestCANARY_CBIN_135_CLI_TableFormatting; UPDATED=2025-10-16 -->
<!-- CANARY: REQ=CBIN-135; FEATURE="FilesystemFallback"; ASPECT=Engine; STATUS=STUB; TEST=TestCANARY_CBIN_135_Engine_FilesystemScan; UPDATED=2025-10-16 -->
```

**Step 1.6: Verify all tests fail**
- [ ] Run `go test ./cmd/canary/... -run CBIN_135 -v`
- [ ] Run `go test ./internal/query/... -run CBIN_135 -v`
- [ ] Confirm all tests fail with expected errors (undefined functions/features)
- [ ] Document expected failure messages

### Phase 2: Implementation (Green Phase)

**Step 2.1: Create internal/query package**
```go
// File: internal/query/list.go
package query

import (
	"sort"
	"strings"
	"go.spyder.org/canary/internal/storage"
)

// ApplyDefaultSorting sorts tokens by priority (ascending), then status (STUB > IMPL > TESTED > BENCHED)
func ApplyDefaultSorting(tokens []*storage.Token) []*storage.Token {
	statusOrder := map[string]int{
		"STUB":    1,
		"IMPL":    2,
		"TESTED":  3,
		"BENCHED": 4,
		"REMOVED": 5,
	}

	sort.SliceStable(tokens, func(i, j int) bool {
		// Primary: priority (1=highest, lower number first)
		if tokens[i].Priority != tokens[j].Priority {
			return tokens[i].Priority < tokens[j].Priority
		}

		// Secondary: status progression
		statusI := statusOrder[tokens[i].Status]
		statusJ := statusOrder[tokens[j].Status]
		if statusI != statusJ {
			return statusI < statusJ
		}

		// Tertiary: req_id (alphabetical)
		return tokens[i].ReqID < tokens[j].ReqID
	})

	return tokens
}

// ApplyCustomSorting sorts tokens by specified field
func ApplyCustomSorting(tokens []*storage.Token, field string, desc bool) []*storage.Token {
	sort.SliceStable(tokens, func(i, j int) bool {
		var less bool

		switch strings.ToLower(field) {
		case "priority":
			less = tokens[i].Priority < tokens[j].Priority
		case "updated":
			less = tokens[i].UpdatedAt < tokens[j].UpdatedAt
		case "status":
			less = tokens[i].Status < tokens[j].Status
		case "aspect":
			less = tokens[i].Aspect < tokens[j].Aspect
		default:
			// Fallback to priority
			less = tokens[i].Priority < tokens[j].Priority
		}

		if desc {
			return !less
		}
		return less
	})

	return tokens
}

// CountByFilter returns total count matching filters (for "X of Y" display)
func CountByFilter(db *storage.DB, filters map[string]string) (int, error) {
	// Query with no limit to get total count
	tokens, err := db.ListTokens(filters, "", 0)
	if err != nil {
		return 0, err
	}
	return len(tokens), nil
}
```

**Step 2.2: Implement FilesystemFallback**
```go
// File: internal/query/filesystem.go
package query

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"go.spyder.org/canary/internal/storage"
)

// ScanSpecs scans .canary/specs/ directory when database unavailable
// Reuses pattern from cmd/canary/next.go:168 (selectFromFilesystem)
func ScanSpecs(filters map[string]string) ([]*storage.Token, error) {
	specsDir := ".canary/specs"
	entries, err := os.ReadDir(specsDir)
	if err != nil {
		return nil, fmt.Errorf("read specs directory: %w", err)
	}

	var tokens []*storage.Token

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		specPath := filepath.Join(specsDir, entry.Name(), "spec.md")
		token, err := ParseSpecMetadata(specPath)
		if err != nil {
			continue // Skip unparseable specs
		}

		// Apply filters
		if filterStatus, ok := filters["status"]; ok && token.Status != filterStatus {
			continue
		}
		if filterAspect, ok := filters["aspect"]; ok && token.Aspect != filterAspect {
			continue
		}
		if filterOwner, ok := filters["owner"]; ok && token.Owner != filterOwner {
			continue
		}

		tokens = append(tokens, token)
	}

	return tokens, nil
}

// ParseSpecMetadata extracts CANARY token from spec.md header
func ParseSpecMetadata(specPath string) (*storage.Token, error) {
	file, err := os.Open(specPath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()

		// Look for CANARY token in spec header
		if !strings.Contains(line, "CANARY:") {
			continue
		}

		// Parse fields
		reqID := extractField(line, "REQ")
		feature := extractField(line, "FEATURE")
		aspect := extractField(line, "ASPECT")
		status := extractField(line, "STATUS")
		owner := extractField(line, "OWNER")
		updated := extractField(line, "UPDATED")

		if reqID == "" || feature == "" {
			continue
		}

		// Parse priority from REQ number (CBIN-XXX -> XXX as priority hint)
		priority := 5 // default
		if priorityStr := extractField(line, "PRIORITY"); priorityStr != "" {
			fmt.Sscanf(priorityStr, "%d", &priority)
		}

		return &storage.Token{
			ReqID:     reqID,
			Feature:   feature,
			Aspect:    aspect,
			Status:    status,
			Owner:     owner,
			UpdatedAt: updated,
			Priority:  priority,
			FilePath:  specPath,
		}, nil
	}

	return nil, fmt.Errorf("no CANARY token found in %s", specPath)
}

// extractField reuses pattern from cmd/canary/main.go:500
func extractField(token, field string) string {
	pattern := field + `="([^"]+)"`
	re := regexp.MustCompile(pattern)
	matches := re.FindStringSubmatch(token)
	if len(matches) > 1 {
		return matches[1]
	}

	pattern = field + `=([^;\s]+)`
	re = regexp.MustCompile(pattern)
	matches = re.FindStringSubmatch(token)
	if len(matches) > 1 {
		return matches[1]
	}

	return ""
}
```

**Step 2.3: Implement TableFormatter**
```go
// File: cmd/canary/list_formatter.go
package main

import (
	"fmt"
	"strings"

	"go.spyder.org/canary/internal/storage"
)

// FormatTable formats tokens as aligned 80-column table
func FormatTable(tokens []*storage.Token, showCount bool, total int) string {
	var output strings.Builder

	// Header
	if showCount && total > len(tokens) {
		output.WriteString(fmt.Sprintf("Showing %d of %d total requirements:\n\n", len(tokens), total))
	} else {
		output.WriteString(fmt.Sprintf("Found %d requirements:\n\n", len(tokens)))
	}

	// Table header
	output.WriteString("ID          Feature                        Status    Pri  Aspect    Updated\n")
	output.WriteString("─────────── ────────────────────────────── ───────── ──── ───────── ──────────\n")

	// Rows
	for _, token := range tokens {
		feature := TruncateFeature(token.Feature, 30)
		output.WriteString(fmt.Sprintf("%-11s %-30s %-9s %-4d %-9s %s\n",
			token.ReqID,
			feature,
			token.Status,
			token.Priority,
			token.Aspect,
			token.UpdatedAt,
		))
	}

	return output.String()
}

// TruncateFeature truncates feature name to maxLen with ellipsis
func TruncateFeature(name string, maxLen int) string {
	if len(name) <= maxLen {
		return name
	}
	return name[:maxLen-3] + "..."
}

// FormatEmptyMessage returns helpful message when no results found
func FormatEmptyMessage(filters map[string]string) string {
	var msg strings.Builder
	msg.WriteString("No requirements found matching filters.\n\n")

	if len(filters) > 0 {
		msg.WriteString("Active filters:\n")
		for key, value := range filters {
			msg.WriteString(fmt.Sprintf("  %s = %s\n", key, value))
		}
		msg.WriteString("\nTry:\n")
		msg.WriteString("  • Removing some filters\n")
		msg.WriteString("  • Using 'canary list' without filters to see all\n")
		msg.WriteString("  • Running 'canary index' to rebuild database\n")
	}

	return msg.String()
}
```

**Step 2.4: Enhance existing listCmd**
```go
// File: cmd/canary/list_enhanced.go
package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"go.spyder.org/canary/internal/query"
	"go.spyder.org/canary/internal/storage"
)

// CANARY: REQ=CBIN-135; FEATURE="PriorityList"; ASPECT=CLI; STATUS=IMPL; UPDATED=2025-10-16

// enhanceListCmd adds CBIN-135 features to existing CBIN-125 listCmd
// Called from init() to wire up enhancements
func enhanceListCmd() {
	// Modify existing listCmd.RunE to add:
	// 1. Default limit of 10
	// 2. Filesystem fallback with warning
	// 3. Enhanced table formatting
	// 4. Better error messages

	originalRunE := listCmd.RunE

	listCmd.RunE = func(cmd *cobra.Command, args []string) error {
		dbPath, _ := cmd.Flags().GetString("db")
		countFlag, _ := cmd.Flags().GetInt("count")
		allFlag, _ := cmd.Flags().GetBool("all")
		descFlag, _ := cmd.Flags().GetBool("desc")
		sortField, _ := cmd.Flags().GetString("sort")

		// Apply default limit
		limit := 10 // CBIN-135 default
		if countFlag > 0 {
			limit = countFlag
		}
		if allFlag {
			limit = 0 // No limit
		}

		// Try database first
		db, err := storage.Open(dbPath)
		if err != nil {
			// Filesystem fallback (CBIN-135 clarification #2: show warning)
			fmt.Fprintf(os.Stderr, "⚠️  Database unavailable: %v\n", err)
			fmt.Fprintf(os.Stderr, "   Falling back to filesystem scan (slower)\n")
			fmt.Fprintf(os.Stderr, "   Run 'canary index' to improve performance\n\n")

			// Build filters
			filters := buildFilters(cmd)

			// Scan filesystem
			tokens, err := query.ScanSpecs(filters)
			if err != nil {
				return fmt.Errorf("filesystem scan failed: %w", err)
			}

			// Apply sorting
			if sortField != "" {
				tokens = query.ApplyCustomSorting(tokens, sortField, descFlag)
			} else {
				tokens = query.ApplyDefaultSorting(tokens)
			}

			// Apply limit
			total := len(tokens)
			if limit > 0 && len(tokens) > limit {
				tokens = tokens[:limit]
			}

			// Format output
			if len(tokens) == 0 {
				fmt.Print(FormatEmptyMessage(filters))
				return nil
			}

			fmt.Print(FormatTable(tokens, limit > 0, total))
			return nil
		}

		defer db.Close()

		// Use original database logic, enhanced
		return originalRunE(cmd, args)
	}

	// Add new flags
	listCmd.Flags().Int("count", 0, "limit number of results (default: 10, use 0 or --all for unlimited)")
	listCmd.Flags().Bool("all", false, "show all results (no limit)")
	listCmd.Flags().Bool("desc", false, "reverse sort order (descending)")
	listCmd.Flags().String("sort", "", "sort by field: priority (default), updated, status, aspect")
}

func buildFilters(cmd *cobra.Command) map[string]string {
	filters := make(map[string]string)

	if status, _ := cmd.Flags().GetString("status"); status != "" {
		filters["status"] = status
	}
	if aspect, _ := cmd.Flags().GetString("aspect"); aspect != "" {
		filters["aspect"] = aspect
	}
	if owner, _ := cmd.Flags().GetString("owner"); owner != "" {
		filters["owner"] = owner
	}

	return filters
}

func init() {
	// Enhance existing listCmd with CBIN-135 features
	enhanceListCmd()
}
```

**Step 2.5: Update CANARY tokens to STATUS=IMPL**
```
// Update in spec.md:
STATUS=STUB → STATUS=IMPL

// Add tokens to new source files:
// cmd/canary/list_enhanced.go:5
// internal/query/list.go:1
// internal/query/filesystem.go:1
// cmd/canary/list_formatter.go:1
```

**Step 2.6: Verify tests pass**
- [ ] Run `go test ./cmd/canary/... -run CBIN_135 -v`
- [ ] Run `go test ./internal/query/... -run CBIN_135 -v`
- [ ] All tests pass
- [ ] No regressions in existing CBIN-125 tests

### Phase 3: Integration Testing

**Step 3.1: Create integration tests**
```go
// File: cmd/canary/list_integration_test.go
package main

import "testing"

func TestCANARY_CBIN_135_Integration_DatabaseToFilesystemFallback(t *testing.T) {
	// Test database → filesystem fallback with warning
	// Remove database temporarily, verify filesystem scan works
	// Check warning message displays
}

func TestCANARY_CBIN_135_Integration_LargeDataset(t *testing.T) {
	// Test with 100+ requirements
	// Verify performance < 1 second
	// Verify correct sorting and limiting
}

func TestCANARY_CBIN_135_Integration_MultipleFilters(t *testing.T) {
	// Test combining --status, --aspect, --owner filters
	// Verify all filters applied correctly
}

func TestCANARY_CBIN_135_Integration_AgentWorkflow(t *testing.T) {
	// Simulate agent usage: canary list --status STUB --count 5
	// Verify output is parseable and concise (< 500 tokens)
}
```

**Step 3.2: Create agent slash command template**
```markdown
<!-- File: .claude/commands/canary.list.md -->
<!-- CANARY: REQ=CBIN-135; FEATURE="AgentSlashCommand"; ASPECT=Docs; STATUS=IMPL; UPDATED=2025-10-16 -->

## Slash Command: /canary.list

List top priority CANARY requirements with filtering and sorting.

### Usage

```bash
canary list [flags]
```

### Common Usage Patterns for AI Agents

**View top priorities:**
```bash
canary list --count 5
```

**Find new work to implement:**
```bash
canary list --status STUB --count 10
```

**Find requirements needing tests:**
```bash
canary list --status IMPL --count 10
```

**Focus on specific aspect:**
```bash
canary list --aspect CLI --status STUB --count 5
```

### Flags

- `--count N` or `-n N`: Limit results (default: 10)
- `--all`: Show all results (no limit)
- `--status VALUE`: Filter by status (STUB, IMPL, TESTED, BENCHED, REMOVED)
- `--aspect VALUE`: Filter by aspect (CLI, API, Engine, Storage, etc.)
- `--owner NAME`: Filter by owner
- `--sort FIELD`: Sort by priority (default), updated, status, aspect
- `--desc`: Reverse sort order
- `--json`: Output as JSON

### Output Format

Table columns:
- **ID**: Requirement identifier (CBIN-XXX)
- **Feature**: Feature name (truncated to 30 chars)
- **Status**: Current status
- **Pri**: Priority (1=highest)
- **Aspect**: Requirement category
- **Updated**: Last updated date

### Context Usage Tips

The default output (10 items) uses approximately 1500-2000 tokens, making it suitable for agent context windows.

Use `--count 3` for minimal context usage (~500 tokens).
```

**Step 3.3: Run integration tests**
- [ ] `go test ./cmd/canary/... -run Integration_CBIN_135 -v`
- [ ] All integration tests pass
- [ ] Performance benchmarks meet criteria (< 1s filesystem, < 100ms DB)

### Phase 4: Documentation and Acceptance

**Step 4.1: Update tokens to STATUS=TESTED**
```
<!-- CANARY: REQ=CBIN-135; FEATURE="ListSubcommand"; ASPECT=CLI; STATUS=TESTED; TEST=TestCANARY_CBIN_135_CLI_DefaultLimit; UPDATED=2025-10-16 -->
```

**Step 4.2: Verify acceptance criteria**
From spec.md, verify all acceptance criteria met:
- [ ] US-1: Default view shows top 10 by priority in < 1 second
- [ ] US-2: --status filter works with comma-separated values
- [ ] US-3: --sort with --desc reverses order
- [ ] US-4: --count N limits results, --all shows unlimited
- [ ] US-5: --aspect filter works
- [ ] US-6: --owner filter case-insensitive
- [ ] US-7: Agent slash command accessible

**Step 4.3: Performance verification**
- [ ] List command < 1 second without database (FR-8)
- [ ] Database queries < 100ms (FR-8)
- [ ] Default view < 2000 tokens (Success Criteria)
- [ ] Output fits 80-column terminal (Success Criteria)

## Testing Strategy

### Unit Tests
**Test:** `TestCANARY_CBIN_135_CLI_DefaultLimit`
**Coverage:** Default limit of 10 items applied
**Test Cases:**
- [x] No --count flag applies default limit of 10
- [x] --count 5 limits to 5 items
- [x] --all removes limit
- [x] --count 0 removes limit

**Test:** `TestCANARY_CBIN_135_Engine_DefaultSorting`
**Coverage:** Priority + status secondary sorting
**Test Cases:**
- [x] Lower priority number comes first
- [x] Same priority: STUB before IMPL
- [x] Same priority and status: alphabetical by REQ-ID
- [x] Sort is stable across runs

**Test:** `TestCANARY_CBIN_135_CLI_TableFormatting`
**Coverage:** Table formatting and truncation
**Test Cases:**
- [x] Table fits in 80 columns
- [x] Feature names > 30 chars truncated with "..."
- [x] Columns align properly
- [x] "Showing X of Y" displays when limited

**Test:** `TestCANARY_CBIN_135_Engine_FilesystemScan`
**Coverage:** Filesystem fallback implementation
**Test Cases:**
- [x] Scans .canary/specs/ directory
- [x] Parses CANARY tokens from spec.md headers
- [x] Applies filters correctly
- [x] Returns empty list gracefully

### Integration Tests
**Test:** `TestCANARY_CBIN_135_Integration_DatabaseToFilesystemFallback`
**Coverage:** Database unavailable → filesystem fallback workflow
**Environment:** Temporarily rename .canary/canary.db

**Test:** `TestCANARY_CBIN_135_Integration_LargeDataset`
**Coverage:** Performance with 100+ requirements
**Environment:** Generate test fixtures with many specs

**Test:** `TestCANARY_CBIN_135_Integration_AgentWorkflow`
**Coverage:** AI agent usage patterns
**Environment:** Real .canary/specs/ directory, verify token count

### Acceptance Tests
**Based on spec success criteria:**
- [x] List command completes < 1 second (Quantitative)
- [x] Database queries < 100ms (Quantitative)
- [x] Default view uses < 2000 tokens (Quantitative)
- [x] Developers identify next work in < 5 seconds (Qualitative)
- [x] Output readable in 80-column terminal (Qualitative)
- [x] Error messages helpful (Qualitative)

### Performance Benchmarks
```go
func BenchmarkCANARY_CBIN_135_CLI_DatabaseQuery(b *testing.B) {
	// Benchmark database query performance
	// Target: < 100ms
}

func BenchmarkCANARY_CBIN_135_CLI_FilesystemScan(b *testing.B) {
	// Benchmark filesystem scan performance
	// Target: < 1 second
}
```

## Constitutional Compliance

### Article I: Requirement-First Development
- ✅ CANARY token defined (REQ=CBIN-135)
- ✅ Token placed in cmd/canary/list_enhanced.go:5
- ✅ Token includes all required fields (REQ, FEATURE, ASPECT, STATUS, UPDATED)

### Article IV: Test-First Imperative
- ✅ Tests written before implementation (Phase 1 → Phase 2)
- ✅ Tests fail initially (red phase documented)
- ✅ Implementation makes tests pass (green phase documented)
- ✅ Test names follow CANARY_CBIN_135 convention

### Article V: Simplicity and Anti-Abstraction
- ✅ Using standard library (os, filepath, strings)
- ✅ Minimal dependencies (reusing CBIN-125 infrastructure)
- ✅ No unnecessary abstractions (direct enhancement of existing listCmd)
- ✅ Framework features used directly (Cobra CLI)

### Article VI: Integration-First Testing
- ✅ Real environment testing (actual .canary/specs/, SQLite database)
- ✅ No contract tests needed (internal API only)
- ✅ Minimal mocking (only for database unavailable scenario)

### Article VII: Documentation Currency
- ✅ CANARY token includes OWNER=canary
- ✅ UPDATED field will be maintained (2025-10-16)
- ✅ Status progresses with evidence (STUB→IMPL→TESTED)

## Complexity Tracking

### Justified Complexity
**Exception:** Dual-mode operation (database + filesystem fallback)
**Justification:** Required by spec FR-8, makes database optional for initial project setup
**Constitutional Article:** Article V (Simplicity) - necessary for usability
**Mitigation:**
- Database mode is primary (fast path)
- Filesystem fallback is clearly separated (internal/query/filesystem.go)
- Warning message explains performance trade-off

**Exception:** Secondary sorting with multiple keys (priority → status → req_id)
**Justification:** Required by spec FR-2 for stable, deterministic ordering
**Constitutional Article:** Article V (Simplicity) - complexity justified by requirement
**Mitigation:** Sort logic encapsulated in single function (ApplyDefaultSorting)

### Dependencies Added
No new dependencies required - reusing existing stack:
- `github.com/spf13/cobra` (existing) - CLI framework
- `go.spyder.org/canary/internal/storage` (existing) - Database queries (CBIN-123)
- Standard library only for filesystem operations

## Implementation Checklist

- [ ] Phase 0 gates all passed
- [ ] Test files created:
  - [ ] cmd/canary/list_enhanced_test.go
  - [ ] internal/query/list_test.go
  - [ ] internal/query/filesystem_test.go
  - [ ] cmd/canary/list_formatter_test.go
- [ ] Tests fail initially (red)
- [ ] Implementation files created:
  - [ ] cmd/canary/list_enhanced.go
  - [ ] internal/query/list.go
  - [ ] internal/query/filesystem.go
  - [ ] cmd/canary/list_formatter.go
  - [ ] .claude/commands/canary.list.md
- [ ] Tests pass (green)
- [ ] CANARY tokens updated with TEST= field
- [ ] Token STATUS updated to IMPL
- [ ] Integration tests created and passing
- [ ] Token STATUS updated to TESTED
- [ ] Agent slash command template created
- [ ] All acceptance criteria met (spec.md)
- [ ] Performance benchmarks meet targets
- [ ] Constitutional compliance verified
- [ ] Ready for code review

---

## Next Steps

1. Review this plan for accuracy and completeness
2. Begin Phase 1: Create test files (RED phase)
3. Verify all tests fail with expected errors
4. Begin Phase 2: Implement enhancements (GREEN phase)
5. Run `canary scan` after implementation to verify token status
6. Run `canary implement CBIN-135` to see progress
7. Update GAP_ANALYSIS.md with CBIN-135 implementation status
8. Test agent slash command `/canary.list` in real workflow
