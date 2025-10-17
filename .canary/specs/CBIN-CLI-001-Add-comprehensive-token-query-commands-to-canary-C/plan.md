# CANARY: REQ=CBIN-CLI-001; FEATURE="TokenQueryCommands"; ASPECT=CLI; STATUS=STUB; OWNER=canary; UPDATED=2025-10-16
# Implementation Plan: CBIN-CLI-001 - Token Query Commands

**Requirement:** CBIN-CLI-001
**Feature Name:** TokenQueryCommands
**Specification:** [../spec.md](./spec.md)
**Status:** STUB → Ready for Implementation
**Created:** 2025-10-16
**Updated:** 2025-10-16

---

## Constitutional Compliance Review

### Article I: Requirement-First Development ✅
- **Token Format**: `// CANARY: REQ=CBIN-CLI-001; FEATURE="TokenQueryCommands"; ASPECT=CLI; STATUS=STUB; UPDATED=2025-10-16`
- **Evidence-Based Promotion**: Plan includes explicit status progression (STUB → IMPL → TESTED)
- **Staleness Management**: UPDATED field will be maintained

### Article II: Specification Discipline ✅
- **WHAT Before HOW**: Spec focuses on user needs (agents need better query tools)
- **Testable Requirements**: All FR items have clear acceptance criteria
- **No Clarifications**: Zero [NEEDS CLARIFICATION] markers in spec

### Article III: Token-Driven Planning ✅
- **Token Granularity**: Each command (show, files, status, grep) gets own token
- **Aspect Classification**: CLI (commands), Storage (query abstraction), Docs (templates)
- **Cross-Cutting Concerns**: Properly split across aspects

### Article IV: Test-First Imperative ✅ **NON-NEGOTIABLE**
- **Phase 1 = Tests**: All test files created BEFORE implementation
- **Red-Green-Refactor**: Explicit phases documented
- **Test Naming**: TEST= field references added to all tokens

### Article V: Simplicity and Anti-Abstraction ✅
- **Standard Library**: Using Go's database/sql, fmt, strings (no external dependencies)
- **No Premature Optimization**: Simple query functions, no caching initially
- **Minimal Complexity**: Direct database access, no ORM

### Article VI: Integration-First Testing ✅
- **Real Environment**: Tests use actual database and filesystem
- **Contract-First**: SQL queries defined before implementation

### Article VII: Documentation Currency ✅
- **Tokens as Documentation**: All commands have CANARY tokens
- **UPDATED Field**: Will be maintained during implementation
- **Gap Analysis**: Spec includes tracking in Implementation Checklist

**All Constitutional Gates: PASSED ✅**

---

## Tech Stack Decision

### Language & Runtime
- **Go 1.19+**
- **Rationale**:
  - Already project standard
  - Excellent database/sql support
  - Built-in fmt package for formatted output
  - Fast compilation for quick iteration

### Core Dependencies
- **Standard Library Only** (Article V: Prefer standard library)
  - `database/sql` - Database queries
  - `os` - Filesystem fallback
  - `fmt` - Formatted output
  - `strings` - String manipulation
  - `regexp` - Pattern matching for grep
  - `github.com/spf13/cobra` - CLI framework (already in use)
  - `github.com/fatih/color` - Terminal colors (already in use for other commands)

### No External Dependencies
- **Database Access**: Use existing `internal/storage` package
- **Output Formatting**: Use Go's fmt and tabwriter (standard library)

### Database Integration
- **Primary**: Use existing `.canary/canary.db` SQLite database
- **Fallback**: Filesystem grep if database missing
- **Rationale**: Follow existing pattern from list/search commands

---

## Architecture Overview

```
┌─────────────────────────────────────────────────────┐
│  canary show/files/status/grep <args> [--flags]    │
└───────────────────┬─────────────────────────────────┘
                    │
                    v
┌─────────────────────────────────────────────────────┐
│            Command Handlers (cmd/canary/)           │
│  • showCmd.RunE() - Display tokens                  │
│  • filesCmd.RunE() - List files                     │
│  • statusCmd.RunE() - Show progress                 │
│  • grepCmd.RunE() - Search tokens                   │
└─────┬──────────────────────┬────────────────────────┘
      │                      │
      v                      v
┌──────────────────┐   ┌─────────────────────────────┐
│ Query Functions  │   │  Output Formatters          │
│ (storage layer)  │   │                              │
│                  │   │  • FormatTokensTable        │
│ • GetTokensByID  │   │  • FormatFilesList          │
│ • GetFilesByID   │   │  • FormatStatusSummary      │
│ • SearchTokens   │   │  • HighlightMatches         │
│                  │   └─────────────────────────────┘
│ • Filesystem     │
│   Fallback       │
└──────────────────┘
```

### Key Components

1. **CLI Commands** (`cmd/canary/main.go`, `cmd/canary/{show,files,status,grep}.go`)
   - Cobra command definitions
   - Flag parsing
   - High-level orchestration

2. **Query Abstraction** (`internal/storage/queries.go` - new file)
   - `GetTokensByReqID(reqID string) ([]*Token, error)`
   - `GetFilesByReqID(reqID string, excludeSpecs bool) (map[string][]Token, error)`
   - `GetStatusSummary(reqID string) (*StatusSummary, error)`
   - `SearchTokens(pattern string, fields []string, regex bool) ([]*Token, error)`

3. **Output Formatters** (`cmd/canary/format.go` - new file)
   - `FormatTokensTable(tokens []*Token, groupBy string) string`
   - `FormatFilesList(fileGroups map[string][]Token) string`
   - `FormatStatusSummary(summary *StatusSummary) string`

4. **Filesystem Fallback** (`cmd/canary/fallback.go` - new file)
   - `GrepTokensByReqID(reqID string) ([]*Token, error)`
   - Optimized grep patterns when DB missing

---

## CANARY Token Placement

### Main Commands

**Show Command:**
```go
// File: cmd/canary/show.go (new file)
// CANARY: REQ=CBIN-CLI-001; FEATURE="ShowCmd"; ASPECT=CLI; STATUS=STUB; UPDATED=2025-10-16
var showCmd = &cobra.Command{
	Use:   "show <REQ-ID>",
	Short: "Display all CANARY tokens for a requirement",
	RunE: func(cmd *cobra.Command, args []string) error {
		// Implementation
	},
}
```

**Files Command:**
```go
// File: cmd/canary/files.go (new file)
// CANARY: REQ=CBIN-CLI-001; FEATURE="FilesCmd"; ASPECT=CLI; STATUS=STUB; UPDATED=2025-10-16
var filesCmd = &cobra.Command{
	Use:   "files <REQ-ID>",
	Short: "List implementation files for a requirement",
	RunE: func(cmd *cobra.Command, args []string) error {
		// Implementation
	},
}
```

**Status Command:**
```go
// File: cmd/canary/status.go (new file)
// CANARY: REQ=CBIN-CLI-001; FEATURE="StatusCmd"; ASPECT=CLI; STATUS=STUB; UPDATED=2025-10-16
var statusCmd = &cobra.Command{
	Use:   "status <REQ-ID>",
	Short: "Show implementation progress for a requirement",
	RunE: func(cmd *cobra.Command, args []string) error {
		// Implementation
	},
}
```

**Grep Command:**
```go
// File: cmd/canary/grep.go (new file)
// CANARY: REQ=CBIN-CLI-001; FEATURE="GrepCmd"; ASPECT=CLI; STATUS=STUB; UPDATED=2025-10-16
var grepCmd = &cobra.Command{
	Use:   "grep <pattern>",
	Short: "Search CANARY tokens by pattern",
	RunE: func(cmd *cobra.Command, args []string) error {
		// Implementation
	},
}
```

### Query Abstraction

```go
// File: internal/storage/queries.go (new file)
// CANARY: REQ=CBIN-CLI-001; FEATURE="QueryAbstraction"; ASPECT=Storage; STATUS=STUB; UPDATED=2025-10-16

// GetTokensByReqID retrieves all tokens for a requirement ID
func (db *DB) GetTokensByReqID(reqID string) ([]*Token, error) {
	// Implementation
}

// GetFilesByReqID groups tokens by file path, excluding specs if requested
func (db *DB) GetFilesByReqID(reqID string, excludeSpecs bool) (map[string][]Token, error) {
	// Implementation
}
```

---

## Implementation Phases

Following Article IV (Test-First Imperative), implementation MUST proceed in strict order:

### Phase 0: Pre-Implementation Setup

**Constitutional Gate Check:**
- [x] All constitutional articles reviewed (completed above ✅)
- [x] Test-first approach validated
- [x] Simplicity gate passed (standard library only)
- [x] Token format validated

**Create Test Files (TDD Red Phase):**
```bash
# Create test files BEFORE any implementation
touch cmd/canary/show_test.go
touch cmd/canary/files_test.go
touch cmd/canary/status_test.go
touch cmd/canary/grep_test.go
touch internal/storage/queries_test.go
```

---

### Phase 1: Query Abstraction (Test-First)

#### 1.1 Storage Layer Tests

```go
// File: internal/storage/queries_test.go
// CANARY: REQ=CBIN-CLI-001; FEATURE="QueryAbstractionTests"; ASPECT=Storage; STATUS=STUB; TEST=TestCANARY_CBIN_CLI_001_Storage_GetTokensByReqID; UPDATED=2025-10-16
package storage_test

import "testing"

func TestCANARY_CBIN_CLI_001_Storage_GetTokensByReqID(t *testing.T) {
	// Setup: Create test database with sample tokens
	db := setupTestDB(t)
	defer db.Close()

	// Insert test tokens for CBIN-999
	insertTestToken(t, db, "CBIN-999", "TestFeature1", "CLI", "TESTED")
	insertTestToken(t, db, "CBIN-999", "TestFeature2", "API", "IMPL")

	// Execute: Query tokens
	tokens, err := db.GetTokensByReqID("CBIN-999")

	// Verify: Should return 2 tokens
	if err != nil {
		t.Fatalf("GetTokensByReqID failed: %v", err)
	}
	if len(tokens) != 2 {
		t.Errorf("Expected 2 tokens, got %d", len(tokens))
	}
}

func TestCANARY_CBIN_CLI_001_Storage_GetFilesByReqID(t *testing.T) {
	// Test file grouping and spec filtering
	// Expected to FAIL initially
}

func TestCANARY_CBIN_CLI_001_Storage_SearchTokens(t *testing.T) {
	// Test pattern matching across fields
	// Expected to FAIL initially
}
```

**Run Tests (Expect Failures):**
```bash
go test ./internal/storage -v -run TestCANARY_CBIN_CLI_001
# All tests should FAIL (functions don't exist yet)
```

#### 1.2 Implement Query Functions (Green Phase)

```go
// File: internal/storage/queries.go (new)
// CANARY: REQ=CBIN-CLI-001; FEATURE="QueryAbstraction"; ASPECT=Storage; STATUS=IMPL; UPDATED=2025-10-16
package storage

// GetTokensByReqID retrieves all tokens matching requirement ID
func (db *DB) GetTokensByReqID(reqID string) ([]*Token, error) {
	query := `
		SELECT req_id, feature, aspect, status, file_path, line_number,
		       test, bench, owner, priority, updated_at
		FROM tokens
		WHERE req_id = ?
		ORDER BY aspect, feature
	`

	rows, err := db.conn.Query(query, reqID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tokens []*Token
	for rows.Next() {
		token := &Token{}
		err := rows.Scan(
			&token.ReqID, &token.Feature, &token.Aspect, &token.Status,
			&token.FilePath, &token.LineNumber, &token.Test, &token.Bench,
			&token.Owner, &token.Priority, &token.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		tokens = append(tokens, token)
	}

	return tokens, rows.Err()
}

// GetFilesByReqID groups tokens by file path
func (db *DB) GetFilesByReqID(reqID string, excludeSpecs bool) (map[string][]*Token, error) {
	tokens, err := db.GetTokensByReqID(reqID)
	if err != nil {
		return nil, err
	}

	// Group by file path, filter specs if requested
	fileGroups := make(map[string][]*Token)
	for _, token := range tokens {
		if excludeSpecs && shouldExcludeFile(token.FilePath) {
			continue
		}
		fileGroups[token.FilePath] = append(fileGroups[token.FilePath], token)
	}

	return fileGroups, nil
}

// shouldExcludeFile checks if file is spec/template/plan
func shouldExcludeFile(path string) bool {
	excludePatterns := []string{
		".canary/specs/",
		".canary/templates/",
		"base/",
		"/plan.md",
		"/spec.md",
	}
	for _, pattern := range excludePatterns {
		if strings.Contains(path, pattern) {
			return true
		}
	}
	return false
}

// SearchTokens searches across multiple fields
func (db *DB) SearchTokens(pattern string, fields []string, regex bool) ([]*Token, error) {
	// Build dynamic WHERE clause
	var conditions []string
	var args []interface{}

	for _, field := range fields {
		if regex {
			conditions = append(conditions, fmt.Sprintf("%s REGEXP ?", field))
		} else {
			conditions = append(conditions, fmt.Sprintf("%s LIKE ?", field))
			pattern = "%" + pattern + "%"
		}
		args = append(args, pattern)
	}

	query := fmt.Sprintf(`
		SELECT req_id, feature, aspect, status, file_path, line_number,
		       test, bench, owner, priority, updated_at
		FROM tokens
		WHERE %s
		ORDER BY req_id, aspect
	`, strings.Join(conditions, " OR "))

	// Execute query and scan results
	// (similar to GetTokensByReqID)
}
```

**Update Token Status:**
```go
// CANARY: REQ=CBIN-CLI-001; FEATURE="QueryAbstraction"; ASPECT=Storage; STATUS=TESTED; TEST=TestCANARY_CBIN_CLI_001_Storage_GetTokensByReqID; UPDATED=2025-10-16
```

---

### Phase 2: Show Command Implementation

#### 2.1 Show Command Tests

```go
// File: cmd/canary/show_test.go (new)
// CANARY: REQ=CBIN-CLI-001; FEATURE="ShowCmdTests"; ASPECT=CLI; STATUS=STUB; TEST=TestCANARY_CBIN_CLI_001_ShowCmd; UPDATED=2025-10-16
package main

func TestCANARY_CBIN_CLI_001_ShowCmd_ExactMatch(t *testing.T) {
	// Setup: Create test database with tokens
	tmpDir := t.TempDir()
	setupTestDB(t, tmpDir, "CBIN-999")

	// Execute: Run show command
	output, err := runCommand("show", "CBIN-999")

	// Verify: Output contains token details
	if err != nil {
		t.Fatalf("show command failed: %v", err)
	}
	if !strings.Contains(output, "CBIN-999") {
		t.Errorf("Output missing requirement ID: %s", output)
	}
	if !strings.Contains(output, "CLI") && !strings.Contains(output, "API") {
		t.Errorf("Output missing aspect grouping: %s", output)
	}
}

func TestCANARY_CBIN_CLI_001_ShowCmd_NotFound(t *testing.T) {
	// Test error handling for non-existent requirement
}

func TestCANARY_CBIN_CLI_001_ShowCmd_JSONOutput(t *testing.T) {
	// Test --json flag
}
```

#### 2.2 Implement Show Command

```go
// File: cmd/canary/show.go (new)
// CANARY: REQ=CBIN-CLI-001; FEATURE="ShowCmd"; ASPECT=CLI; STATUS=IMPL; UPDATED=2025-10-16
package main

var showCmd = &cobra.Command{
	Use:   "show <REQ-ID>",
	Short: "Display all CANARY tokens for a requirement",
	Long: `Show displays all CANARY tokens for a specific requirement ID.

Displays:
- Feature name, aspect, status
- File location and line number
- Test and benchmark references
- Owner and priority

Grouping:
- By default, groups by aspect (CLI, API, Engine, etc.)
- Use --group-by status to group by implementation status
- Use --json for machine-readable output`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		reqID := args[0]
		groupBy, _ := cmd.Flags().GetString("group-by")
		jsonOutput, _ := cmd.Flags().GetBool("json")

		// Open database
		db, err := storage.Open(".canary/canary.db")
		if err != nil {
			// Fallback to filesystem grep
			return showWithFilesystemFallback(reqID)
		}
		defer db.Close()

		// Query tokens
		tokens, err := db.GetTokensByReqID(reqID)
		if err != nil {
			return fmt.Errorf("query tokens: %w", err)
		}

		if len(tokens) == 0 {
			fmt.Printf("No tokens found for %s\n", reqID)
			fmt.Println("\nSuggestions:")
			fmt.Println("  • Run: canary list")
			fmt.Println("  • Check requirement ID format (e.g., CBIN-XXX)")
			return fmt.Errorf("requirement not found")
		}

		// Format output
		if jsonOutput {
			return outputJSON(tokens)
		}

		fmt.Printf("Tokens for %s:\n\n", reqID)
		output := formatTokensTable(tokens, groupBy)
		fmt.Println(output)

		return nil
	},
}
```

---

### Phase 3: Files, Status, Grep Commands

Similar TDD approach for remaining commands:

1. **Files Command** - List implementation files grouped by aspect
2. **Status Command** - Show progress summary with colored output
3. **Grep Command** - Pattern search across token fields

(Implementation follows same pattern as Show command)

---

### Phase 4: Output Formatting

```go
// File: cmd/canary/format.go (new)
// CANARY: REQ=CBIN-CLI-001; FEATURE="OutputFormatting"; ASPECT=CLI; STATUS=IMPL; UPDATED=2025-10-16

func formatTokensTable(tokens []*Token, groupBy string) string {
	var buf strings.Builder

	// Group tokens
	groups := groupTokens(tokens, groupBy)

	// Format each group
	for groupName, groupTokens := range groups {
		buf.WriteString(fmt.Sprintf("## %s\n\n", groupName))

		for _, token := range groupTokens {
			buf.WriteString(fmt.Sprintf("📌 %s - %s\n", token.ReqID, token.Feature))
			buf.WriteString(fmt.Sprintf("   Status: %s | Aspect: %s\n", token.Status, token.Aspect))
			buf.WriteString(fmt.Sprintf("   Location: %s:%d\n", token.FilePath, token.LineNumber))

			if token.Test != "" {
				buf.WriteString(fmt.Sprintf("   Test: %s\n", token.Test))
			}
			buf.WriteString("\n")
		}
	}

	return buf.String()
}
```

---

### Phase 5: Template Updates

```markdown
<!-- File: .canary/templates/implement-prompt-template.md -->
<!-- Update to use new commands instead of grep/sqlite3 -->

## Implementation Files

Instead of using grep or sqlite3, use:
```bash
canary show {{.ReqID}}      # View all tokens
canary files {{.ReqID}}     # List implementation files
canary status {{.ReqID}}    # Check progress
```

**Example Output:**
```
$ canary files CBIN-133
Implementation files for CBIN-133:

CLI:
  cmd/canary/main.go (1 token)
  cmd/canary/implement.go (2 tokens)

Engine:
  internal/matcher/fuzzy.go (1 token)
```
```

---

## Testing Strategy

### Unit Tests
- **GetTokensByReqID**: Test database queries, result parsing
- **GetFilesByReqID**: Test file grouping, spec filtering
- **SearchTokens**: Test pattern matching, regex support
- **Show Command**: Test output formatting, grouping, JSON mode
- **Files Command**: Test file listing, aspect grouping
- **Status Command**: Test progress calculation, colored output
- **Grep Command**: Test search across fields, highlighting

### Integration Tests
- **End-to-End Show**: Create test tokens, run show command, verify output
- **Filesystem Fallback**: Test behavior when database missing
- **Template Usage**: Verify templates use new commands

### Manual Testing
- **Real Database**: Test with actual .canary/canary.db
- **Large Datasets**: Test performance with >1000 tokens
- **Terminal Colors**: Verify colored output works correctly

---

## Performance Targets

- **Database Queries**: <100ms for <10,000 tokens
- **Filesystem Fallback**: <1 second for <1000 files
- **Output Formatting**: <50ms for <100 tokens
- **Memory Usage**: <10MB for typical queries

---

## Constitutional Compliance Final Check

- ✅ **Article I**: CANARY tokens placed at all implementation points
- ✅ **Article II**: Spec is technology-agnostic, focuses on user needs
- ✅ **Article III**: Sub-features split by aspect (CLI, Storage, Docs)
- ✅ **Article IV**: Tests written FIRST (Phase 1), implementation SECOND (Phase 2-4)
- ✅ **Article V**: Standard library only (database/sql, fmt, strings)
- ✅ **Article VI**: Integration tests use real database and filesystem
- ✅ **Article VII**: All tokens have UPDATED field, STATUS progression documented

**Plan Status**: Ready for Implementation ✅

---

## Implementation Checklist

- [ ] Phase 0 gates all passed
- [ ] Query abstraction tests created
- [ ] Query abstraction implemented
- [ ] Show command tests created
- [ ] Show command implemented
- [ ] Files command tests created
- [ ] Files command implemented
- [ ] Status command tests created
- [ ] Status command implemented
- [ ] Grep command tests created
- [ ] Grep command implemented
- [ ] Output formatting implemented
- [ ] Template updates completed
- [ ] All tests passing
- [ ] Documentation updated
- [ ] Constitutional compliance verified

---

## Next Steps

1. **Begin Implementation**: Run `/canary.implement CBIN-CLI-001`
2. **Follow TDD**: Red (failing tests) → Green (passing implementation) → Refactor
3. **Update Tokens**: Progress STATUS from STUB → IMPL → TESTED
4. **Update Templates**: Replace grep/sqlite3 with new commands
5. **Verify**: Run `canary scan` and `canary status CBIN-CLI-001` after completion

---

**Implementation Guidance Generated:** 2025-10-16
**Plan Version:** 1.0
**Status:** APPROVED FOR IMPLEMENTATION
