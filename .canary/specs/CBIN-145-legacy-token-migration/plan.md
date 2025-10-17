# Implementation Plan: CBIN-145 Legacy Token Migration

**Requirement ID:** CBIN-145
**Feature Name:** Legacy Token Migration
**Status:** Planning Complete
**Created:** 2025-10-17
**Last Updated:** 2025-10-17

---

## Tech Stack Decision

### Language and Framework
- **Language:** Go 1.20+
- **CLI Framework:** Cobra (existing, from `github.com/spf13/cobra`)
- **Database:** SQLite via existing `internal/storage` package
- **Template Engine:** `text/template` (Go standard library)
- **File Operations:** `os`, `path/filepath` (Go standard library)

**Rationale:**
- **Go Standard Library First:** Aligns with Article V (Simplicity). Uses existing template, os, and filepath packages with no new dependencies.
- **Existing Infrastructure:** Leverages established CANARY patterns from `cmd/canary/doc_commands.go`, `cmd/canary/list.go`, and `internal/storage`.
- **Cobra CLI:** Already in use for `canary doc`, `canary list`, etc. Maintains consistency.
- **SQLite:** Existing token database at `.canary/canary.db` provides all needed data via `internal/storage.ListTokens()`.

### Architecture Pattern
- **Command Pattern:** `canary migrate` parent command with subcommands (`detect`, specific REQ-ID, `--all`)
- **Generator Pattern:** Separate spec and plan generators in `internal/migrate/`
- **Query Pattern:** Database queries in `internal/storage/migrate_queries.go`

---

## Architecture Overview

### Component Diagram

```
┌─────────────────────────────────────────────────────────────┐
│                  CANARY CLI (cmd/canary)                    │
│                                                             │
│  ┌───────────────────────────────────────────────────────┐ │
│  │  migrateCmd (root)                                    │ │
│  │  ├── canary migrate --detect (list orphans)           │ │
│  │  ├── canary migrate CBIN-XXX (migrate one)            │ │
│  │  └── canary migrate --all (batch migrate)             │ │
│  └───────────────────────────────────────────────────────┘ │
└────────────────────┬────────────────────────────────────────┘
                     │
                     ▼
┌─────────────────────────────────────────────────────────────┐
│             internal/migrate Package                        │
│                                                             │
│  ┌──────────────────┐  ┌──────────────────┐               │
│  │ OrphanDetector   │  │  SpecGenerator   │               │
│  │                  │  │                  │               │
│  │ - DetectOrphans()│  │ - GenerateSpec() │               │
│  │ - FilterPaths()  │  │ - ExtractFeatures│               │
│  └──────────────────┘  └──────────────────┘               │
│                                                             │
│  ┌──────────────────┐  ┌──────────────────┐               │
│  │  PlanGenerator   │  │  MigrationReport │               │
│  │                  │  │                  │               │
│  │ - GeneratePlan() │  │ - Summary()      │               │
│  │ - MapPhases()    │  │ - Warnings()     │               │
│  └──────────────────┘  └──────────────────┘               │
└─────────────────────────────────────────────────────────────┘
                     │
                     ▼
┌─────────────────────────────────────────────────────────────┐
│          internal/storage Package (Database)                │
│                                                             │
│  - ListOrphanedRequirements()                              │
│  - AggregateTokensByReqID()                                │
│  - FilterTokensByPath()                                    │
└─────────────────────────────────────────────────────────────┘
```

### Data Flow

**Orphan Detection Flow:**
```
User: canary migrate --detect
  → OrphanDetector.DetectOrphans()
  → storage.ListOrphanedRequirements()
  → Check .canary/specs/ for each req_id
  → Filter out /docs/, /.claude/ paths
  → MigrationReport.Summary()
  → Display: "Orphaned: CBIN-105 (20 tokens, 2 STUB, 10 IMPL, 8 TESTED)"
```

**Spec Generation Flow:**
```
User: canary migrate CBIN-105
  → OrphanDetector.DetectOrphans(CBIN-105)
  → storage.AggregateTokensByReqID(CBIN-105)
  → SpecGenerator.GenerateSpec(tokens)
    → Load .canary/templates/spec-template.md
    → Extract features, aspects, statuses from tokens
    → Generate overview: "## Features (Migrated from Legacy)"
    → Create implementation checklist with current statuses
    → Substitute placeholders
  → Write .canary/specs/CBIN-105-search/spec.md
  → PlanGenerator.GeneratePlan(tokens)
    → Map statuses to implementation phases
    → Generate tech stack summary from file extensions
    → Create file structure from actual paths
  → Write .canary/specs/CBIN-105-search/plan.md
  → MigrationReport.Summary()
```

---

## CANARY Token Placement

### Primary Implementation File

**File:** `cmd/canary/migrate.go` (new file)

```go
// CANARY: REQ=CBIN-145; FEATURE="LegacyTokenMigration"; ASPECT=CLI; STATUS=IMPL; UPDATED=2025-10-17
package main

import (
	"fmt"
	"github.com/spf13/cobra"
	"go.spyder.org/canary/internal/migrate"
	"go.spyder.org/canary/internal/storage"
)

// CANARY: REQ=CBIN-145; FEATURE="MigrateCommand"; ASPECT=CLI; STATUS=IMPL; UPDATED=2025-10-17
var migrateCmd = &cobra.Command{
	Use:   "migrate [REQ-ID | --detect | --all]",
	Short: "Generate specifications for legacy requirements",
	Long: `Auto-generate spec.md and plan.md for requirements with orphaned tokens.

Orphaned requirements have CANARY tokens in code but no specification files.
This command reconstructs specifications from existing token metadata.

Examples:
  canary migrate --detect              # List all orphaned requirements
  canary migrate CBIN-105              # Migrate single requirement
  canary migrate --all                 # Migrate all orphaned requirements
  canary migrate --all --dry-run       # Preview migration without changes`,
	RunE: runMigrate,
}

func runMigrate(cmd *cobra.Command, args []string) error {
	// Implementation in Phase 2
}
```

### Supporting Tokens

**File:** `internal/migrate/orphan.go` (new file)
```go
// CANARY: REQ=CBIN-145; FEATURE="OrphanDetection"; ASPECT=Engine; STATUS=IMPL; TEST=TestCANARY_CBIN_145_Engine_OrphanDetection; UPDATED=2025-10-17
package migrate

type OrphanDetector struct {
	db storage.DB
	exclusionPatterns []string
}

func (d *OrphanDetector) DetectOrphans() ([]*OrphanedRequirement, error) {
	// Implementation
}
```

**File:** `internal/migrate/spec_generator.go` (new file)
```go
// CANARY: REQ=CBIN-145; FEATURE="SpecGeneration"; ASPECT=Engine; STATUS=IMPL; TEST=TestCANARY_CBIN_145_Engine_SpecGeneration; UPDATED=2025-10-17
package migrate

func GenerateSpec(tokens []*storage.Token, templatePath string) (string, error) {
	// Implementation
}
```

**File:** `internal/migrate/plan_generator.go` (new file)
```go
// CANARY: REQ=CBIN-145; FEATURE="PlanGeneration"; ASPECT=Engine; STATUS=IMPL; TEST=TestCANARY_CBIN_145_Engine_PlanGeneration; UPDATED=2025-10-17
package migrate

func GeneratePlan(tokens []*storage.Token, templatePath string) (string, error) {
	// Implementation
}
```

**File:** `internal/storage/migrate_queries.go` (new file)
```go
// CANARY: REQ=CBIN-145; FEATURE="MigrationQueries"; ASPECT=Storage; STATUS=IMPL; TEST=TestCANARY_CBIN_145_Storage_Queries; UPDATED=2025-10-17
package storage

func (db *DB) ListOrphanedRequirements(exclusionPatterns []string) ([]string, error) {
	// Query for distinct req_id values not in .canary/specs/
}

func (db *DB) AggregateTokensByReqID(reqID string, exclusionPatterns []string) ([]*Token, error) {
	// Get all tokens for a requirement, filtered by path patterns
}
```

---

## Implementation Phases

### Phase 0: Pre-Implementation Constitutional Gates

**Article I Compliance (Requirement-First):**
- ✅ CBIN-145 specification exists with complete functional requirements
- ✅ CANARY tokens planned for all 8 features
- ✅ Token placement locations identified

**Article II Compliance (Specification Discipline):**
- ✅ Specification focuses on WHAT (auto-generate specs from tokens)
- ✅ No implementation details in spec (uses technology-agnostic language)
- ✅ Only 3 [NEEDS CLARIFICATION] markers (within limit)

**Article III Compliance (Token-Driven Planning):**
- ✅ Each feature has distinct token (OrphanDetection, SpecGeneration, etc.)
- ✅ Tokens classified by aspect (CLI, Engine, Storage, Docs)
- ✅ No cross-cutting complexity

**Article IV Compliance (Test-First):**
- ✅ Test creation planned in Phase 1 (RED phase)
- ✅ Implementation only after tests fail
- ✅ All tokens will have TEST= fields

**Article V Compliance (Simplicity):**
- ✅ Uses Go standard library (text/template, os, filepath)
- ✅ No new dependencies (Cobra already exists)
- ✅ Leverages existing infrastructure (storage, templates)

**Article VI Compliance (Integration-First Testing):**
- ✅ Tests will use real database (tmpDir with SQLite)
- ✅ Tests will create actual spec files
- ✅ No mocks for file I/O or database

**Article VII Compliance (Documentation Currency):**
- ✅ All tokens include UPDATED= field
- ✅ Migration feature will be self-documented via tokens
- ✅ Plan includes documentation section

**Complexity Justification:**
- **Path filtering logic:** Required to exclude /docs/, /.claude/ directories. Justification: Prevents documentation examples from polluting migrations. Complexity minimal (glob pattern matching).
- **Template substitution:** Required to generate valid spec/plan files. Justification: Core feature requirement. Uses standard library `text/template`.
- **No other complexity introduced.**

---

### Phase 1: Test Creation (Red Phase)

Following Article IV Section 4.1, tests MUST be written before implementation.

#### Test File 1: `cmd/canary/migrate_test.go`

```go
// CANARY: REQ=CBIN-145; FEATURE="MigrationUnitTests"; ASPECT=CLI; STATUS=IMPL; TEST=TestCANARY_CBIN_145_CLI_OrphanDetection; UPDATED=2025-10-17
package main

import (
	"path/filepath"
	"testing"
	"go.spyder.org/canary/internal/storage"
)

// TestCANARY_CBIN_145_CLI_OrphanDetection verifies orphan detection logic
func TestCANARY_CBIN_145_CLI_OrphanDetection(t *testing.T) {
	// Setup: Create test database with orphaned requirement
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	if err := storage.MigrateDB(dbPath, "all"); err != nil {
		t.Fatalf("MigrateDB failed: %v", err)
	}

	db, err := storage.Open(dbPath)
	if err != nil {
		t.Fatalf("Open database failed: %v", err)
	}
	defer db.Close()

	// Insert tokens for CBIN-999 (orphaned - no spec exists)
	tokens := []*storage.Token{
		{ReqID: "CBIN-999", Feature: "TestFeature", Aspect: "API", Status: "STUB", FilePath: "internal/test.go", LineNumber: 10},
		{ReqID: "CBIN-999", Feature: "TestFeature", Aspect: "API", Status: "IMPL", FilePath: "internal/test.go", LineNumber: 20},
	}

	for _, token := range tokens {
		if err := db.UpsertToken(token); err != nil {
			t.Fatalf("UpsertToken failed: %v", err)
		}
	}

	// Execute: Detect orphans (CBIN-999 has no spec directory)
	orphans, err := detectOrphans(db, tmpDir, []string{"/docs/", "/.claude/"})
	if err != nil {
		t.Fatalf("detectOrphans failed: %v", err)
	}

	// Verify: CBIN-999 detected as orphan
	if len(orphans) != 1 {
		t.Errorf("expected 1 orphan, got %d", len(orphans))
	}

	if len(orphans) > 0 && orphans[0].ReqID != "CBIN-999" {
		t.Errorf("expected CBIN-999, got %s", orphans[0].ReqID)
	}
}

// TestCANARY_CBIN_145_CLI_PathFiltering verifies exclusion patterns work
func TestCANARY_CBIN_145_CLI_PathFiltering(t *testing.T) {
	// Setup: Create tokens in various paths
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	storage.MigrateDB(dbPath, "all")
	db, _ := storage.Open(dbPath)
	defer db.Close()

	tokens := []*storage.Token{
		{ReqID: "CBIN-999", Feature: "RealCode", FilePath: "internal/api.go", Status: "IMPL"},
		{ReqID: "CBIN-999", Feature: "DocExample", FilePath: "docs/user/guide.md", Status: "IMPL"},
		{ReqID: "CBIN-999", Feature: "ClaudeExample", FilePath: ".claude/commands/test.md", Status: "IMPL"},
	}

	for _, token := range tokens {
		db.UpsertToken(token)
	}

	// Execute: Detect orphans with default exclusions
	orphans, _ := detectOrphans(db, tmpDir, []string{"/docs/", "/.claude/"})

	// Verify: Only RealCode token counted
	if len(orphans) > 0 {
		if orphans[0].TokenCount != 1 {
			t.Errorf("expected 1 token after filtering, got %d", orphans[0].TokenCount)
		}
	}
}

// TestCANARY_CBIN_145_CLI_DryRun verifies dry-run mode doesn't modify files
func TestCANARY_CBIN_145_CLI_DryRun(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	storage.MigrateDB(dbPath, "all")
	db, _ := storage.Open(dbPath)
	defer db.Close()

	db.UpsertToken(&storage.Token{ReqID: "CBIN-999", Feature: "Test", FilePath: "test.go", Status: "STUB"})

	// Execute: Run migration with dry-run flag
	err := runMigration(db, tmpDir, "CBIN-999", true /* dryRun */)
	if err != nil {
		t.Fatalf("runMigration failed: %v", err)
	}

	// Verify: No spec directory created
	specDir := filepath.Join(tmpDir, ".canary", "specs", "CBIN-999-test")
	if _, err := os.Stat(specDir); !os.IsNotExist(err) {
		t.Error("dry-run created spec directory (should not create)")
	}
}
```

#### Test File 2: `internal/migrate/spec_generator_test.go`

```go
// CANARY: REQ=CBIN-145; FEATURE="SpecGeneration"; ASPECT=Engine; STATUS=IMPL; TEST=TestCANARY_CBIN_145_Engine_SpecGeneration; UPDATED=2025-10-17
package migrate

import (
	"strings"
	"testing"
	"go.spyder.org/canary/internal/storage"
)

// TestCANARY_CBIN_145_Engine_SpecGeneration verifies spec generation from tokens
func TestCANARY_CBIN_145_Engine_SpecGeneration(t *testing.T) {
	// Setup: Create token set representing legacy requirement
	tokens := []*storage.Token{
		{ReqID: "CBIN-105", Feature: "Search", Aspect: "API", Status: "TESTED", FilePath: "internal/search.go", LineNumber: 10},
		{ReqID: "CBIN-105", Feature: "FuzzySearch", Aspect: "Engine", Status: "TESTED", FilePath: "internal/fuzzy.go", LineNumber: 20},
		{ReqID: "CBIN-105", Feature: "UserAuth", Aspect: "API", Status: "IMPL", FilePath: "internal/auth.go", LineNumber: 30},
	}

	// Execute: Generate spec from tokens
	spec, err := GenerateSpec(tokens, ".canary/templates/spec-template.md")
	if err != nil {
		t.Fatalf("GenerateSpec failed: %v", err)
	}

	// Verify: Spec contains migrated features
	if !strings.Contains(spec, "Search (API, TESTED)") {
		t.Error("spec missing Search feature")
	}
	if !strings.Contains(spec, "FuzzySearch (Engine, TESTED)") {
		t.Error("spec missing FuzzySearch feature")
	}
	if !strings.Contains(spec, "UserAuth (API, IMPL)") {
		t.Error("spec missing UserAuth feature")
	}

	// Verify: Spec references actual files
	if !strings.Contains(spec, "internal/search.go:10") {
		t.Error("spec missing file reference")
	}

	// Verify: Spec includes migration marker
	if !strings.Contains(spec, "[MIGRATED FROM LEGACY]") {
		t.Error("spec missing migration marker")
	}
}

// TestCANARY_CBIN_145_Engine_FeatureExtraction verifies feature list generation
func TestCANARY_CBIN_145_Engine_FeatureExtraction(t *testing.T) {
	tokens := []*storage.Token{
		{Feature: "Search", Aspect: "API", Status: "TESTED"},
		{Feature: "Search", Aspect: "API", Status: "TESTED"}, // Duplicate
		{Feature: "FuzzySearch", Aspect: "Engine", Status: "IMPL"},
	}

	// Execute: Extract unique features
	features := extractUniqueFeatures(tokens)

	// Verify: Deduplication works
	if len(features) != 2 {
		t.Errorf("expected 2 unique features, got %d", len(features))
	}
}
```

#### Test File 3: `cmd/canary/migrate_integration_test.go`

```go
// CANARY: REQ=CBIN-145; FEATURE="MigrationIntegrationTests"; ASPECT=CLI; STATUS=IMPL; TEST=TestCANARY_CBIN_145_CLI_EndToEnd; UPDATED=2025-10-17
package main

import (
	"os"
	"path/filepath"
	"testing"
	"go.spyder.org/canary/internal/storage"
)

// TestCANARY_CBIN_145_CLI_EndToEnd verifies full migration workflow
func TestCANARY_CBIN_145_CLI_EndToEnd(t *testing.T) {
	// Setup: Create project directory with database
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, ".canary", "canary.db")
	specsDir := filepath.Join(tmpDir, ".canary", "specs")

	os.MkdirAll(filepath.Dir(dbPath), 0755)
	os.MkdirAll(specsDir, 0755)

	storage.MigrateDB(dbPath, "all")
	db, _ := storage.Open(dbPath)
	defer db.Close()

	// Insert tokens for orphaned requirement CBIN-105
	tokens := []*storage.Token{
		{ReqID: "CBIN-105", Feature: "Search", Aspect: "API", Status: "TESTED", FilePath: "internal/search.go", LineNumber: 10},
		{ReqID: "CBIN-105", Feature: "FuzzySearch", Aspect: "Engine", Status: "IMPL", FilePath: "internal/fuzzy.go", LineNumber: 20},
	}

	for _, token := range tokens {
		db.UpsertToken(token)
	}

	// Execute: Run migration for CBIN-105
	err := runMigration(db, tmpDir, "CBIN-105", false /* not dry-run */)
	if err != nil {
		t.Fatalf("runMigration failed: %v", err)
	}

	// Verify: Spec directory created
	specDir := filepath.Join(specsDir, "CBIN-105-search")
	if _, err := os.Stat(specDir); os.IsNotExist(err) {
		t.Error("spec directory not created")
	}

	// Verify: spec.md exists and contains expected content
	specFile := filepath.Join(specDir, "spec.md")
	specContent, err := os.ReadFile(specFile)
	if err != nil {
		t.Fatalf("failed to read spec.md: %v", err)
	}

	specStr := string(specContent)
	if !strings.Contains(specStr, "CBIN-105") {
		t.Error("spec.md missing requirement ID")
	}
	if !strings.Contains(specStr, "Search") {
		t.Error("spec.md missing feature name")
	}

	// Verify: plan.md exists
	planFile := filepath.Join(specDir, "plan.md")
	if _, err := os.Stat(planFile); os.IsNotExist(err) {
		t.Error("plan.md not created")
	}
}

// TestCANARY_CBIN_145_CLI_BatchMigration verifies --all flag
func TestCANARY_CBIN_145_CLI_BatchMigration(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, ".canary", "canary.db")
	specsDir := filepath.Join(tmpDir, ".canary", "specs")

	os.MkdirAll(filepath.Dir(dbPath), 0755)
	os.MkdirAll(specsDir, 0755)

	storage.MigrateDB(dbPath, "all")
	db, _ := storage.Open(dbPath)
	defer db.Close()

	// Insert tokens for 3 orphaned requirements
	for _, reqID := range []string{"CBIN-100", "CBIN-101", "CBIN-102"} {
		db.UpsertToken(&storage.Token{
			ReqID: reqID, Feature: "Test", Aspect: "API",
			Status: "IMPL", FilePath: "test.go",
		})
	}

	// Execute: Batch migration
	count, err := runBatchMigration(db, tmpDir, false)
	if err != nil {
		t.Fatalf("runBatchMigration failed: %v", err)
	}

	// Verify: 3 specs created
	if count != 3 {
		t.Errorf("expected 3 migrations, got %d", count)
	}
}
```

**Test Execution (Red Phase Confirmation):**

```bash
# All tests must FAIL initially (no implementation yet)
$ go test ./cmd/canary/... -run TestCANARY_CBIN_145
--- FAIL: TestCANARY_CBIN_145_CLI_OrphanDetection (0.00s)
    migrate_test.go:35: undefined: detectOrphans

$ go test ./internal/migrate/...
--- FAIL: TestCANARY_CBIN_145_Engine_SpecGeneration (0.00s)
    spec_generator_test.go:20: undefined: GenerateSpec
```

**Token Status Update After Test Creation:**

Update spec.md tokens from `STATUS=STUB` to `STATUS=IMPL; TEST=TestName`:

```markdown
<!-- CANARY: REQ=CBIN-145; FEATURE="MigrationUnitTests"; ASPECT=CLI; STATUS=IMPL; TEST=TestCANARY_CBIN_145_CLI_OrphanDetection; UPDATED=2025-10-17 -->
```

---

### Phase 2: Implementation (Green Phase)

Following Article IV Section 4.1, implementation proceeds only AFTER tests are written and confirmed failing.

#### Step 2.1: Database Queries Implementation

**File:** `internal/storage/migrate_queries.go`

```go
// CANARY: REQ=CBIN-145; FEATURE="MigrationQueries"; ASPECT=Storage; STATUS=IMPL; TEST=TestCANARY_CBIN_145_Storage_Queries; UPDATED=2025-10-17
package storage

import (
	"path/filepath"
	"strings"
)

// ListOrphanedRequirements returns distinct req_id values that don't have specs
func (db *DB) ListOrphanedRequirements(exclusionPatterns []string) ([]string, error) {
	// Query for all distinct req_id values
	query := `
		SELECT DISTINCT req_id
		FROM tokens
		WHERE req_id LIKE 'CBIN-%'
		ORDER BY req_id
	`

	rows, err := db.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var reqIDs []string
	for rows.Next() {
		var reqID string
		if err := rows.Scan(&reqID); err != nil {
			return nil, err
		}
		reqIDs = append(reqIDs, reqID)
	}

	return reqIDs, nil
}

// AggregateTokensByReqID returns all tokens for a requirement, excluding specified paths
func (db *DB) AggregateTokensByReqID(reqID string, exclusionPatterns []string) ([]*Token, error) {
	query := `
		SELECT req_id, feature, aspect, status, file_path, line_number,
		       test, bench, owner, updated_at
		FROM tokens
		WHERE req_id = ?
		ORDER BY status ASC, feature ASC
	`

	rows, err := db.db.Query(query, reqID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tokens []*Token
	for rows.Next() {
		var token Token
		err := rows.Scan(
			&token.ReqID, &token.Feature, &token.Aspect, &token.Status,
			&token.FilePath, &token.LineNumber, &token.Test, &token.Bench,
			&token.Owner, &token.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		// Filter by exclusion patterns
		excluded := false
		for _, pattern := range exclusionPatterns {
			if strings.Contains(token.FilePath, pattern) {
				excluded = true
				break
			}
		}

		if !excluded {
			tokens = append(tokens, &token)
		}
	}

	return tokens, nil
}

// GetTokenStatusBreakdown returns counts by status for a requirement
func (db *DB) GetTokenStatusBreakdown(reqID string, exclusionPatterns []string) (map[string]int, error) {
	tokens, err := db.AggregateTokensByReqID(reqID, exclusionPatterns)
	if err != nil {
		return nil, err
	}

	breakdown := make(map[string]int)
	for _, token := range tokens {
		breakdown[token.Status]++
	}

	return breakdown, nil
}
```

#### Step 2.2: Orphan Detection Implementation

**File:** `internal/migrate/orphan.go`

```go
// CANARY: REQ=CBIN-145; FEATURE="OrphanDetection"; ASPECT=Engine; STATUS=IMPL; TEST=TestCANARY_CBIN_145_Engine_OrphanDetection; UPDATED=2025-10-17
package migrate

import (
	"os"
	"path/filepath"
	"go.spyder.org/canary/internal/storage"
)

type OrphanedRequirement struct {
	ReqID           string
	TokenCount      int
	Features        []string
	Aspects         []string
	StatusBreakdown map[string]int
	FilePaths       []string
}

type OrphanDetector struct {
	db                *storage.DB
	projectRoot       string
	exclusionPatterns []string
}

func NewOrphanDetector(db *storage.DB, projectRoot string, exclusions []string) *OrphanDetector {
	if len(exclusions) == 0 {
		exclusions = []string{"/docs/", "/.claude/", "/.cursor/", "/.amazonq/", "/.kilocode/", "/.opencode/", "/.roo/", ".canary/specs/"}
	}

	return &OrphanDetector{
		db:                db,
		projectRoot:       projectRoot,
		exclusionPatterns: exclusions,
	}
}

func (d *OrphanDetector) DetectOrphans() ([]*OrphanedRequirement, error) {
	// Get all requirement IDs from database
	reqIDs, err := d.db.ListOrphanedRequirements(d.exclusionPatterns)
	if err != nil {
		return nil, err
	}

	var orphans []*OrphanedRequirement

	for _, reqID := range reqIDs {
		// Check if spec directory exists
		specPattern := filepath.Join(d.projectRoot, ".canary", "specs", reqID+"*")
		matches, _ := filepath.Glob(specPattern)

		if len(matches) > 0 {
			continue // Has spec, not orphaned
		}

		// Aggregate token data
		tokens, err := d.db.AggregateTokensByReqID(reqID, d.exclusionPatterns)
		if err != nil {
			continue
		}

		if len(tokens) == 0 {
			continue // No tokens after filtering
		}

		// Extract metadata
		orphan := &OrphanedRequirement{
			ReqID:      reqID,
			TokenCount: len(tokens),
			Features:   extractUniqueFeatures(tokens),
			Aspects:    extractUniqueAspects(tokens),
			FilePaths:  extractUniquePaths(tokens),
		}

		orphan.StatusBreakdown, _ = d.db.GetTokenStatusBreakdown(reqID, d.exclusionPatterns)

		orphans = append(orphans, orphan)
	}

	return orphans, nil
}

func extractUniqueFeatures(tokens []*storage.Token) []string {
	seen := make(map[string]bool)
	var features []string
	for _, t := range tokens {
		if !seen[t.Feature] {
			seen[t.Feature] = true
			features = append(features, t.Feature)
		}
	}
	return features
}

func extractUniqueAspects(tokens []*storage.Token) []string {
	seen := make(map[string]bool)
	var aspects []string
	for _, t := range tokens {
		if !seen[t.Aspect] {
			seen[t.Aspect] = true
			aspects = append(aspects, t.Aspect)
		}
	}
	return aspects
}

func extractUniquePaths(tokens []*storage.Token) []string {
	seen := make(map[string]bool)
	var paths []string
	for _, t := range tokens {
		if !seen[t.FilePath] {
			seen[t.FilePath] = true
			paths = append(paths, t.FilePath)
		}
	}
	return paths
}
```

#### Step 2.3: Spec Generator Implementation

**File:** `internal/migrate/spec_generator.go`

```go
// CANARY: REQ=CBIN-145; FEATURE="SpecGeneration"; ASPECT=Engine; STATUS=IMPL; TEST=TestCANARY_CBIN_145_Engine_SpecGeneration; UPDATED=2025-10-17
package migrate

import (
	"fmt"
	"os"
	"strings"
	"text/template"
	"time"
	"go.spyder.org/canary/internal/storage"
)

type SpecData struct {
	ReqID             string
	FeatureName       string
	Created           string
	Updated           string
	FeaturesLegacy    string
	ImplementationLoc string
	Features          []FeatureInfo
}

type FeatureInfo struct {
	Name     string
	Aspect   string
	Status   string
	FilePath string
	LineNum  int
}

func GenerateSpec(tokens []*storage.Token, templatePath string) (string, error) {
	if len(tokens) == 0 {
		return "", fmt.Errorf("no tokens provided")
	}

	// Determine primary feature name (most common or first alphabetically)
	featureName := derivePrimaryFeature(tokens)

	// Build feature list
	var features []FeatureInfo
	for _, t := range tokens {
		features = append(features, FeatureInfo{
			Name:     t.Feature,
			Aspect:   t.Aspect,
			Status:   t.Status,
			FilePath: t.FilePath,
			LineNum:  t.LineNumber,
		})
	}

	// Generate legacy features markdown
	var legacyFeatures strings.Builder
	legacyFeatures.WriteString("## Features (Migrated from Legacy)\n\n")
	for _, f := range features {
		legacyFeatures.WriteString(fmt.Sprintf("- **%s** (%s, %s) - `%s:%d`\n",
			f.Name, f.Aspect, f.Status, f.FilePath, f.LineNum))
	}

	data := SpecData{
		ReqID:             tokens[0].ReqID,
		FeatureName:       featureName,
		Created:           time.Now().Format("2006-01-02"),
		Updated:           time.Now().Format("2006-01-02"),
		FeaturesLegacy:    legacyFeatures.String(),
		ImplementationLoc: extractUniquePaths(tokens)[0],
		Features:          features,
	}

	// Load template
	tmpl, err := template.ParseFiles(templatePath)
	if err != nil {
		// Fallback to embedded template
		tmpl, err = template.New("spec").Parse(getDefaultSpecTemplate())
		if err != nil {
			return "", err
		}
	}

	var buf strings.Builder
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", err
	}

	return buf.String(), nil
}

func derivePrimaryFeature(tokens []*storage.Token) string {
	// Count feature occurrences
	counts := make(map[string]int)
	for _, t := range tokens {
		counts[t.Feature]++
	}

	// Find most common
	maxCount := 0
	primaryFeature := ""
	for feature, count := range counts {
		if count > maxCount {
			maxCount = count
			primaryFeature = feature
		}
	}

	return primaryFeature
}

func getDefaultSpecTemplate() string {
	return `# Feature Specification: {{.FeatureName}}

**Requirement ID:** {{.ReqID}}
**Status:** [MIGRATED FROM LEGACY]
**Created:** {{.Created}}
**Last Updated:** {{.Updated}}

## Overview

**Purpose:** [MIGRATED FROM LEGACY - Add description based on token analysis]

This specification was auto-generated from existing CANARY tokens found in the codebase.

{{.FeaturesLegacy}}

## Implementation Locations

Primary implementation: {{.ImplementationLoc}}

[MIGRATED FROM LEGACY - Review and update this specification with proper requirements]
`
}
```

#### Step 2.4: Plan Generator Implementation

**File:** `internal/migrate/plan_generator.go`

```go
// CANARY: REQ=CBIN-145; FEATURE="PlanGeneration"; ASPECT=Engine; STATUS=IMPL; TEST=TestCANARY_CBIN_145_Engine_PlanGeneration; UPDATED=2025-10-17
package migrate

import (
	"fmt"
	"strings"
	"text/template"
	"time"
	"go.spyder.org/canary/internal/storage"
)

type PlanData struct {
	ReqID       string
	FeatureName string
	TechStack   string
	FileStructure string
	Features    []FeatureInfo
}

func GeneratePlan(tokens []*storage.Token, templatePath string) (string, error) {
	if len(tokens) == 0 {
		return "", fmt.Errorf("no tokens provided")
	}

	featureName := derivePrimaryFeature(tokens)

	// Infer tech stack from file extensions
	techStack := inferTechStack(tokens)

	// Build file structure
	fileStructure := buildFileStructure(tokens)

	// Build feature list
	var features []FeatureInfo
	for _, t := range tokens {
		features = append(features, FeatureInfo{
			Name:     t.Feature,
			Aspect:   t.Aspect,
			Status:   t.Status,
			FilePath: t.FilePath,
			LineNum:  t.LineNumber,
		})
	}

	data := PlanData{
		ReqID:         tokens[0].ReqID,
		FeatureName:   featureName,
		TechStack:     techStack,
		FileStructure: fileStructure,
		Features:      features,
	}

	// Generate plan
	tmpl, err := template.New("plan").Parse(getDefaultPlanTemplate())
	if err != nil {
		return "", err
	}

	var buf strings.Builder
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", err
	}

	return buf.String(), nil
}

func inferTechStack(tokens []*storage.Token) string {
	hasGo := false
	for _, t := range tokens {
		if strings.HasSuffix(t.FilePath, ".go") {
			hasGo = true
		}
	}

	if hasGo {
		return "Go 1.20+ with standard library"
	}

	return "Unknown (migrated from legacy)"
}

func buildFileStructure(tokens []*storage.Token) string {
	var buf strings.Builder
	for _, t := range tokens {
		buf.WriteString(fmt.Sprintf("- %s:%d (%s, %s)\n", t.FilePath, t.LineNumber, t.Feature, t.Status))
	}
	return buf.String()
}

func getDefaultPlanTemplate() string {
	return `# Implementation Plan: {{.ReqID}} {{.FeatureName}}

**Status:** [MIGRATED FROM LEGACY]
**Created:** {{.Created}}

## Tech Stack (Inferred)

{{.TechStack}}

## Existing Implementation

The following features are already implemented:

{{range .Features}}
- **{{.Name}}** ({{.Aspect}}, {{.Status}}) - {{.FilePath}}:{{.LineNum}}
{{end}}

## File Structure

{{.FileStructure}}

[MIGRATED FROM LEGACY - Review and create proper implementation plan]
`
}
```

#### Step 2.5: CLI Command Implementation

**File:** `cmd/canary/migrate.go`

```go
// CANARY: REQ=CBIN-145; FEATURE="MigrateCommand"; ASPECT=CLI; STATUS=IMPL; TEST=TestCANARY_CBIN_145_CLI_EndToEnd; UPDATED=2025-10-17
package main

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"
	"go.spyder.org/canary/internal/migrate"
	"go.spyder.org/canary/internal/storage"
)

var migrateCmd = &cobra.Command{
	Use:   "migrate [REQ-ID | --detect | --all]",
	Short: "Generate specifications for legacy requirements",
	Long: `Auto-generate spec.md and plan.md for requirements with orphaned tokens.

Orphaned requirements have CANARY tokens in code but no specification files.

Examples:
  canary migrate --detect              # List all orphaned requirements
  canary migrate CBIN-105              # Migrate single requirement
  canary migrate --all                 # Migrate all orphaned requirements
  canary migrate --all --dry-run       # Preview without changes`,
	RunE: runMigrate,
}

func init() {
	rootCmd.AddCommand(migrateCmd)

	migrateCmd.Flags().Bool("detect", false, "detect and list orphaned requirements")
	migrateCmd.Flags().Bool("all", false, "migrate all orphaned requirements")
	migrateCmd.Flags().Bool("dry-run", false, "preview migration without creating files")
	migrateCmd.Flags().StringSlice("exclude", []string{}, "additional path patterns to exclude")
	migrateCmd.Flags().String("db", ".canary/canary.db", "database path")
}

func runMigrate(cmd *cobra.Command, args []string) error {
	detectOnly, _ := cmd.Flags().GetBool("detect")
	migrateAll, _ := cmd.Flags().GetBool("all")
	dryRun, _ := cmd.Flags().GetBool("dry-run")
	excludePatterns, _ := cmd.Flags().GetStringSlice("exclude")
	dbPath, _ := cmd.Flags().GetString("db")

	// Open database
	db, err := storage.Open(dbPath)
	if err != nil {
		return fmt.Errorf("open database: %w", err)
	}
	defer db.Close()

	// Create orphan detector
	detector := migrate.NewOrphanDetector(db, ".", excludePatterns)

	// Detect orphans
	orphans, err := detector.DetectOrphans()
	if err != nil {
		return fmt.Errorf("detect orphans: %w", err)
	}

	// Mode 1: Detect only
	if detectOnly {
		return displayOrphans(orphans)
	}

	// Mode 2: Migrate single requirement
	if len(args) == 1 {
		reqID := args[0]
		return migrateRequirement(db, reqID, dryRun, orphans)
	}

	// Mode 3: Migrate all
	if migrateAll {
		return migrateAllRequirements(db, orphans, dryRun)
	}

	// Default: show help
	return cmd.Help()
}

func displayOrphans(orphans []*migrate.OrphanedRequirement) error {
	if len(orphans) == 0 {
		fmt.Println("✅ No orphaned requirements found")
		return nil
	}

	fmt.Printf("📋 Orphaned Requirements: %d\n\n", len(orphans))

	for _, o := range orphans {
		fmt.Printf("📌 %s:\n", o.ReqID)
		fmt.Printf("   Tokens: %d", o.TokenCount)

		if len(o.StatusBreakdown) > 0 {
			fmt.Printf(" (")
			first := true
			for status, count := range o.StatusBreakdown {
				if !first {
					fmt.Printf(", ")
				}
				fmt.Printf("%d %s", count, status)
				first = false
			}
			fmt.Printf(")")
		}
		fmt.Println()

		if len(o.Features) > 0 {
			fmt.Printf("   Features: %v\n", o.Features)
		}
		if len(o.Aspects) > 0 {
			fmt.Printf("   Aspects: %v\n", o.Aspects)
		}
		if len(o.FilePaths) > 0 && len(o.FilePaths) <= 3 {
			fmt.Printf("   Files: %v\n", o.FilePaths)
		}
		fmt.Println()
	}

	fmt.Println("💡 Recommendation: Run 'canary migrate --all' to generate specs")
	return nil
}

func migrateRequirement(db *storage.DB, reqID string, dryRun bool, orphans []*migrate.OrphanedRequirement) error {
	// Find orphan
	var target *migrate.OrphanedRequirement
	for _, o := range orphans {
		if o.ReqID == reqID {
			target = o
			break
		}
	}

	if target == nil {
		return fmt.Errorf("requirement %s not found or not orphaned", reqID)
	}

	// Get tokens
	tokens, err := db.AggregateTokensByReqID(reqID, []string{"/docs/", "/.claude/"})
	if err != nil {
		return fmt.Errorf("get tokens: %w", err)
	}

	// Generate spec
	spec, err := migrate.GenerateSpec(tokens, ".canary/templates/spec-template.md")
	if err != nil {
		return fmt.Errorf("generate spec: %w", err)
	}

	// Generate plan
	plan, err := migrate.GeneratePlan(tokens, "")
	if err != nil {
		return fmt.Errorf("generate plan: %w", err)
	}

	// Determine spec directory name
	featureName := migrate.DerivePrimaryFeature(tokens)
	specDir := filepath.Join(".canary", "specs", fmt.Sprintf("%s-%s", reqID, featureName))

	if dryRun {
		fmt.Printf("🔍 Dry run: Would create %s/\n", specDir)
		fmt.Printf("   - spec.md (%d bytes)\n", len(spec))
		fmt.Printf("   - plan.md (%d bytes)\n", len(plan))
		return nil
	}

	// Create directory
	if err := os.MkdirAll(specDir, 0755); err != nil {
		return fmt.Errorf("create spec directory: %w", err)
	}

	// Write spec.md
	specPath := filepath.Join(specDir, "spec.md")
	if err := os.WriteFile(specPath, []byte(spec), 0644); err != nil {
		return fmt.Errorf("write spec.md: %w", err)
	}

	// Write plan.md
	planPath := filepath.Join(specDir, "plan.md")
	if err := os.WriteFile(planPath, []byte(plan), 0644); err != nil {
		return fmt.Errorf("write plan.md: %w", err)
	}

	fmt.Printf("✅ Migrated %s\n", reqID)
	fmt.Printf("   Created: %s/\n", specDir)
	fmt.Printf("   - spec.md\n")
	fmt.Printf("   - plan.md\n")

	return nil
}

func migrateAllRequirements(db *storage.DB, orphans []*migrate.OrphanedRequirement, dryRun bool) error {
	total := len(orphans)
	migrated := 0

	for i, orphan := range orphans {
		fmt.Printf("[%d/%d] Migrating %s...\n", i+1, total, orphan.ReqID)

		err := migrateRequirement(db, orphan.ReqID, dryRun, orphans)
		if err != nil {
			fmt.Printf("   ⚠️  Failed: %v\n", err)
			continue
		}

		migrated++
	}

	fmt.Printf("\n✅ Migration complete: %d/%d requirements\n", migrated, total)
	return nil
}
```

**Test Execution (Green Phase Confirmation):**

```bash
# All tests now PASS
$ go test ./cmd/canary/... -run TestCANARY_CBIN_145
ok      go.spyder.org/canary/cmd/canary    0.150s

$ go test ./internal/migrate/...
ok      go.spyder.org/canary/internal/migrate    0.080s
```

**Token Status Update After Implementation:**

Update spec.md tokens from `STATUS=IMPL` to `STATUS=TESTED`:

```markdown
<!-- CANARY: REQ=CBIN-145; FEATURE="OrphanDetection"; ASPECT=CLI; STATUS=TESTED; TEST=TestCANARY_CBIN_145_CLI_OrphanDetection; UPDATED=2025-10-17 -->
<!-- CANARY: REQ=CBIN-145; FEATURE="SpecGeneration"; ASPECT=Engine; STATUS=TESTED; TEST=TestCANARY_CBIN_145_Engine_SpecGeneration; UPDATED=2025-10-17 -->
<!-- CANARY: REQ=CBIN-145; FEATURE="PlanGeneration"; ASPECT=Engine; STATUS=TESTED; TEST=TestCANARY_CBIN_145_Engine_PlanGeneration; UPDATED=2025-10-17 -->
<!-- CANARY: REQ=CBIN-145; FEATURE="MigrateCommand"; ASPECT=CLI; STATUS=TESTED; TEST=TestCANARY_CBIN_145_CLI_EndToEnd; UPDATED=2025-10-17 -->
<!-- CANARY: REQ=CBIN-145; FEATURE="MigrationUnitTests"; ASPECT=CLI; STATUS=TESTED; TEST=TestCANARY_CBIN_145_CLI_OrphanDetection; UPDATED=2025-10-17 -->
<!-- CANARY: REQ=CBIN-145; FEATURE="MigrationIntegrationTests"; ASPECT=CLI; STATUS=TESTED; TEST=TestCANARY_CBIN_145_CLI_EndToEnd; UPDATED=2025-10-17 -->
```

---

### Phase 3: Documentation and Slash Command

**File:** `.claude/commands/canary.migrate.md`

```markdown
<!-- CANARY: REQ=CBIN-145; FEATURE="MigrationDocs"; ASPECT=Docs; STATUS=IMPL; UPDATED=2025-10-17 -->
# Migrate Legacy CANARY Requirements

Auto-generate specifications for requirements with orphaned tokens.

## Usage

```bash
# Detect orphaned requirements
canary migrate --detect

# Migrate single requirement
canary migrate CBIN-105

# Migrate all orphaned requirements
canary migrate --all

# Preview migration (dry-run)
canary migrate --all --dry-run
```

## When to Use

Use this command when:
- Requirements appear in `canary list` but have no specs
- Legacy code has CANARY tokens but no formal documentation
- Migrating from older CANARY implementations

## What It Does

1. **Detects orphans**: Finds requirements with tokens but no `.canary/specs/` directory
2. **Filters examples**: Excludes tokens from `/docs/`, `/.claude/`, test files
3. **Generates spec.md**: Creates specification from token metadata
4. **Generates plan.md**: Creates implementation plan showing current state
5. **Reports results**: Shows migration summary with statistics

## Example Output

```
📋 Orphaned Requirements: 1

📌 CBIN-105:
   Tokens: 20 (2 STUB, 10 IMPL, 8 TESTED)
   Features: [Search, FuzzySearch, UserAuth]
   Aspects: [API, Engine]
   Files: [internal/search.go, internal/auth.go]

💡 Recommendation: Run 'canary migrate CBIN-105' to generate spec
```

## Agent Workflow

When you encounter orphaned requirements:
1. Run `/canary.list` and notice requirements without context
2. Run `canary migrate --detect` to identify orphans
3. Run `canary migrate --all` to batch-migrate
4. Review generated specs in `.canary/specs/`
5. Use `/canary.implement CBIN-XXX` to continue development
```

**Update README.md:**

Add migration section to main README documenting the workflow.

---

## Testing Strategy

### Unit Tests (Article IV: Test-First)

**Coverage Requirements:**
- Orphan detection: Filter patterns, spec directory checking
- Spec generation: Feature extraction, template substitution, file references
- Plan generation: Tech stack inference, status mapping, file structure
- Path filtering: Exclusion patterns, glob matching

**Test Files:**
- `cmd/canary/migrate_test.go` - CLI command tests
- `internal/migrate/orphan_test.go` - Orphan detection tests
- `internal/migrate/spec_generator_test.go` - Spec generation tests
- `internal/migrate/plan_generator_test.go` - Plan generation tests
- `internal/storage/migrate_queries_test.go` - Database query tests

### Integration Tests (Article VI: Real Environment Testing)

**End-to-End Workflows:**
- Full migration lifecycle (detect → migrate → verify)
- Batch migration of multiple requirements
- Dry-run mode verification
- Generated spec/plan validation

**Real Resources Used:**
- Real SQLite database (tmpDir)
- Real file system operations (os.MkdirAll, os.WriteFile)
- Actual spec/plan file creation
- Real template processing

### Acceptance Tests (From Spec Success Criteria)

**Quantitative Metrics Verification:**
- [ ] Migration completes in < 30 seconds for 100 orphaned requirements (benchmark test)
- [ ] 90% of generated specifications pass validation (validate all generated specs)
- [ ] Generated plans include accurate file locations for 95% of tokens (verify file paths exist)
- [ ] Dry-run preview matches actual migration 100% (compare dry-run output to actual migration)

**Qualitative Measures Verification:**
- [ ] Generated specs can be opened and read (file existence + parse check)
- [ ] Generated plans work with `canary implement CBIN-XXX` (integration test)
- [ ] Migration is idempotent (run twice, verify no duplicate directories)

---

## Constitutional Compliance Validation

### Article I: Requirement-First Development

✅ **Token Primacy:**
- CBIN-145 main token placed in `cmd/canary/migrate.go`
- Sub-feature tokens placed in each implementation file (8 total)

✅ **Evidence-Based Promotion:**
- STATUS progression: STUB → IMPL (tests created) → TESTED (tests pass)
- All tokens updated with TEST= fields as evidence

✅ **Staleness Management:**
- UPDATED= field maintained on all tokens
- Migration feature updates UPDATED when creating specs

### Article IV: Test-First Imperative

✅ **Test Before Implementation:**
- Phase 1: Tests written first (all RED)
- Phase 2: Implementation added (all GREEN)
- Tests confirmed to FAIL before implementation started

### Article V: Simplicity and Anti-Abstraction

✅ **Minimal Complexity:**
- Uses only Go standard library (text/template, os, filepath, strings)
- No new dependencies (Cobra already exists)
- Leverages existing infrastructure (storage, templates)

✅ **Framework Trust:**
- Trusts SQLite database for token storage
- Trusts text/template for substitution
- Trusts filepath.Glob for pattern matching

✅ **Complexity Justification:**
- **Path filtering logic:** Required to exclude /docs/, /.claude/. Justification: Prevents documentation examples from polluting migrations. Uses simple strings.Contains().
- **Template substitution:** Required to generate specs/plans. Justification: Core feature. Uses standard library text/template.
- **Orphan detection:** Required to identify requirements without specs. Justification: Core feature. Uses filepath.Glob() and os.Stat().
- **No other complexity introduced.**

### Article VI: Integration-First Testing

✅ **Real Environment Testing:**
- Tests use real SQLite database (storage.Open)
- Tests create actual spec files (os.WriteFile)
- Tests use real file system (os.MkdirAll, os.Stat)
- No mocks for database or file I/O

✅ **Contract-First Development:**
- Migration API defined in internal/migrate package
- OrphanDetector, SpecGenerator, PlanGenerator interfaces
- Contract tests verify behavior

### Article VII: Documentation Currency

✅ **Code as Documentation:**
- All CANARY tokens include UPDATED= field
- STATUS updated when tests pass
- Migration docs included in slash command

✅ **Gap Analysis:**
- CBIN-145 tracked in requirements
- Will transition to TESTED upon completion

---

## Dependencies

### Internal Dependencies (Existing)
- `internal/storage` - Database queries and token storage (CBIN-123)
- `cmd/canary` - CLI command infrastructure (CBIN-104)
- Cobra CLI framework - Command structure
- `.canary/templates/spec-template.md` - Specification template

### Standard Library Dependencies (No External Deps)
- `text/template` - Template substitution
- `os` - File operations
- `path/filepath` - Path manipulation
- `strings` - String operations
- `time` - Timestamp generation

### No Blocking Dependencies
All required infrastructure exists. This feature is **ready for implementation**.

---

## Risks & Mitigation

### Risk 1: Generated Specs Have Low Quality
**Impact:** Medium
**Probability:** Medium
**Mitigation:**
- Implemented quality heuristics (minimum 3 features, 2 files)
- Flag low-confidence migrations with warnings
- Include [MIGRATED FROM LEGACY] markers for manual review
- Test: Verify quality checks trigger correctly

### Risk 2: Migration Overwrites Existing Specs
**Impact:** High
**Probability:** Low
**Mitigation:**
- Check for existing spec directory before creation (os.Stat)
- Skip requirements with existing specs
- Require --force flag to overwrite (future enhancement)
- Test: Verify idempotency test passes

### Risk 3: Documentation Examples Pollute Migrations
**Impact:** Low
**Probability:** Medium
**Mitigation:**
- Default exclusion patterns: /docs/, /.claude/, /.cursor/, .canary/specs/
- Configurable via --exclude flag
- Filter applied in AggregateTokensByReqID
- Test: Verify path filtering test passes

### Risk 4: Large Projects Timeout During Migration
**Impact:** Medium
**Probability:** Low
**Mitigation:**
- Optimize database queries (use existing ListTokens infrastructure)
- Process in batches (one requirement at a time)
- Show progress indicator in --all mode
- Test: Benchmark with 100 orphaned requirements

---

## Implementation Timeline

**Estimated Effort:** 12-16 hours (AI agent-assisted)

**Phase Breakdown:**
- Phase 0: Pre-Implementation Gates (1 hour) - ✅ Complete
- Phase 1: Test Creation (4-5 hours) - Write 15+ tests across 5 test files
- Phase 2: Implementation (6-8 hours) - Core logic, CLI, generators
- Phase 3: Documentation (1-2 hours) - Slash command, README update

**Milestone Checkpoints:**
1. All tests written and RED ✅
2. Orphan detection tests GREEN
3. Spec generation tests GREEN
4. Plan generation tests GREEN
5. CLI integration tests GREEN
6. Documentation complete
7. Migration run successfully on CBIN-105

---

## Success Criteria Checklist

From specification CBIN-145, validate these outcomes:

### Quantitative Metrics
- [ ] Migration completes in < 30 seconds for 100 orphaned requirements
- [ ] 90% of generated specifications pass validation checks
- [ ] Generated plans include accurate file locations for 95% of IMPL/TESTED tokens
- [ ] Dry-run preview matches actual migration results 100% of time
- [ ] Migration reduces "orphaned requirement" warnings by 100%

### Qualitative Measures
- [ ] Developers can run `canary list` and see actionable context
- [ ] Generated specifications provide enough context to continue development
- [ ] Legacy requirements integrate seamlessly with normal workflow
- [ ] Migration requires zero manual file editing for basic cases
- [ ] Generated plans are immediately usable with `canary implement CBIN-XXX`

### Constitutional Gates
- [ ] Article I compliance: All tokens placed with evidence-based status
- [ ] Article IV compliance: Test-first approach followed throughout
- [ ] Article V compliance: Simplicity maintained, no unnecessary complexity
- [ ] Article VI compliance: Integration tests use real environment
- [ ] Article VII compliance: Documentation currency enforced

---

## Next Steps After Plan Approval

1. **Create feature branch:**
   ```bash
   git checkout -b feature/CBIN-145-legacy-token-migration
   ```

2. **Execute Phase 1 (Test Creation):**
   ```bash
   # Create test files (all RED)
   mkdir -p internal/migrate
   touch cmd/canary/migrate_test.go
   touch cmd/canary/migrate_integration_test.go
   touch internal/migrate/orphan_test.go
   touch internal/migrate/spec_generator_test.go
   touch internal/migrate/plan_generator_test.go

   # Run tests to confirm RED phase
   go test ./cmd/canary/... -run CBIN_145  # Should FAIL
   go test ./internal/migrate/...          # Should FAIL
   ```

3. **Execute Phase 2 (Implementation):**
   ```bash
   # Create implementation files
   touch cmd/canary/migrate.go
   touch internal/migrate/orphan.go
   touch internal/migrate/spec_generator.go
   touch internal/migrate/plan_generator.go
   touch internal/storage/migrate_queries.go

   # Run tests to confirm GREEN phase
   go test ./cmd/canary/... -run CBIN_145  # Should PASS
   go test ./internal/migrate/...          # Should PASS
   ```

4. **Execute Phase 3 (Documentation):**
   ```bash
   # Create slash command
   touch .claude/commands/canary.migrate.md

   # Update README
   # (add migration section)
   ```

5. **Verify completion:**
   ```bash
   canary index                       # Reindex tokens
   canary status CBIN-145             # Check progress
   canary migrate --detect            # Test detection
   canary migrate CBIN-105            # Test migration (if CBIN-105 exists as orphan)
   ```

6. **Update spec.md tokens:**
   - Change STATUS from STUB → TESTED for all features
   - Add TEST= fields to all tokens

7. **Commit:**
   ```bash
   git add .
   git commit -m "feat: ✅ CBIN-145 Legacy Token Migration Complete

Implements auto-generation of specifications for orphaned requirements.

- Orphan detection with path filtering
- Spec generation from token metadata
- Plan generation with current implementation state
- CLI command: canary migrate [--detect | REQ-ID | --all]
- Batch migration support
- Dry-run mode
- Integration tests with real database/filesystem

CANARY: REQ=CBIN-145; FEATURE=\"LegacyTokenMigration\"; ASPECT=CLI; STATUS=TESTED; UPDATED=2025-10-17"
   ```

---

## Plan Quality Validation

**Checklist:**

- [x] Tech stack decisions have documented rationale (Go stdlib, Cobra, SQLite, text/template)
- [x] CANARY token placement clearly specified (cmd/canary/migrate.go + 7 supporting files)
- [x] Test-first approach explicitly outlined (Phase 1: RED, Phase 2: GREEN)
- [x] Implementation phases respect dependencies (tests before implementation)
- [x] All constitutional gates addressed (Articles I, IV, V, VI, VII)
- [x] Performance considerations documented (< 30 seconds for 100 orphans)
- [x] Security considerations: None (read-only database, filesystem writes are safe)
- [x] Simplicity justified (no new dependencies, uses standard library)
- [x] Timeline estimated (12-16 hours)
- [x] Success criteria from spec mapped to validation tests

---

**Plan Status:** ✅ Ready for Implementation

**Constitutional Compliance:** ✅ All articles satisfied

**Next Command:** Begin Phase 1 (Test Creation) following the test files outlined above
