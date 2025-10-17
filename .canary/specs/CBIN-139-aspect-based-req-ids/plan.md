<!-- CANARY: REQ=CBIN-116; FEATURE="PlanTemplate"; ASPECT=Docs; STATUS=IMPL; OWNER=canary; UPDATED=2025-10-16 -->
# Implementation Plan: CBIN-139 Aspect-Based Requirement IDs

**Requirement:** CBIN-139
**Specification:** [spec.md](./spec.md)
**Status:** STUB → IMPL
**Created:** 2025-10-16
**Updated:** 2025-10-16

## Tech Stack Decision

### Primary Technologies
- **Language:** Go 1.24
- **Framework:** Standard library
- **Database:** SQLite via modernc.org/sqlite (existing)
- **Testing:** Go testing package + testify
- **Fuzzy Matching:** Existing internal/matcher/fuzzy.go (CBIN-133)

### Rationale
- **Go 1.24**: Already project standard, mature string processing
- **Standard library**: Keeps with Article V (Simplicity), no new dependencies needed
- **Existing SQLite schema**: Extend req_id column from TEXT to accommodate longer IDs
- **Existing fuzzy matcher**: Reuse CBIN-133 implementation for aspect validation suggestions
- **No CGO**: modernc.org/sqlite already provides pure Go implementation

## CANARY Token Placement

### Primary Implementation Tokens

```go
// File: internal/reqid/parser.go (NEW FILE)
// CANARY: REQ=CBIN-139; FEATURE="AspectIDParser"; ASPECT=Engine; STATUS=IMPL; OWNER=canary; UPDATED=2025-10-16
package reqid

// ParseRequirementID parses both old (CBIN-001) and new (CBIN-CLI-001) formats
func ParseRequirementID(reqID string) (*RequirementID, error) {
    // Implementation
}
```

```go
// File: internal/reqid/validator.go (NEW FILE)
// CANARY: REQ=CBIN-139; FEATURE="AspectValidation"; ASPECT=Engine; STATUS=IMPL; OWNER=canary; UPDATED=2025-10-16
package reqid

// ValidateAspect validates aspect against constitution's approved list
func ValidateAspect(aspect string) error {
    // Implementation
}
```

```go
// File: internal/reqid/generator.go (NEW FILE)
// CANARY: REQ=CBIN-139; FEATURE="AspectScopedIDGen"; ASPECT=Engine; STATUS=IMPL; OWNER=canary; UPDATED=2025-10-16
package reqid

// GenerateNextID generates next available aspect-scoped ID
func GenerateNextID(projectKey, aspect string) (string, error) {
    // Implementation
}
```

```go
// File: cmd/canary/migrate_ids.go (NEW FILE)
// CANARY: REQ=CBIN-139; FEATURE="MigrationScript"; ASPECT=CLI; STATUS=IMPL; OWNER=canary; UPDATED=2025-10-16
package main

var migrateIDsCmd = &cobra.Command{
    Use:   "migrate-ids [--dry-run] [--confirm]",
    Short: "Migrate requirement IDs from CBIN-XXX to CBIN-<ASPECT>-XXX format",
    // Implementation
}
```

### File Structure
```
project/
├── internal/
│   └── reqid/                # New package for requirement ID logic
│       ├── parser.go         # ID parser (new + old format)
│       ├── parser_test.go    # Parser tests
│       ├── validator.go      # Aspect validation
│       ├── validator_test.go # Validator tests
│       ├── generator.go      # ID generation
│       └── generator_test.go # Generator tests
├── cmd/canary/
│   ├── migrate_ids.go        # Migration command (NEW)
│   └── migrate_ids_test.go   # Migration tests (NEW)
├── internal/storage/
│   └── migrations/
│       ├── 000002_extend_reqid.up.sql   # Schema migration (NEW)
│       └── 000002_extend_reqid.down.sql # Rollback migration (NEW)
└── .canary/
    ├── templates/
    │   ├── spec-template.md  # Updated with new ID format
    │   └── plan-template.md  # Updated with new ID format
    ├── memory/
    │   └── migration-log.json # Migration record (NEW)
    └── backup/               # Migration backups (NEW)
```

## Architecture Overview

### Component Diagram
```
Migration Flow:
User → migrateIDsCmd → MigrationPlanner → [Scan existing specs]
                                        ↓
                                   [Extract ASPECT from tokens]
                                        ↓
                                   [Generate new IDs: CBIN-XXX → CBIN-<ASPECT>-XXX]
                                        ↓
                                   [Create backup]
                                        ↓
    [Rename spec directories] ← MigrationExecutor → [Update all CANARY tokens in source]
                                        ↓
                                   [Update database req_id]
                                        ↓
                                   [Write migration log]

ID Generation Flow:
canary specify → GenerateNextID → [Scan .canary/specs for aspect-specific IDs]
                                ↓
                           [Find max ID for aspect]
                                ↓
                           [Increment and zero-pad]
                                ↓
                           [Return CBIN-<ASPECT>-###]

Parsing Flow:
Scanner/Commands → ParseRequirementID → [Detect format version]
                                      ↓
                              [v1: CBIN-XXX  → {key: CBIN, id: XXX}]
                              [v2: CBIN-CLI-XXX → {key: CBIN, aspect: CLI, id: XXX}]
```

### Key Components

**Component 1: RequirementID Parser**
- **Responsibility:** Parse both old (CBIN-001) and new (CBIN-CLI-001) formats, extract segments
- **Interfaces:** `ParseRequirementID(reqID string) (*RequirementID, error)`
- **Dependencies:** regexp package

**Component 2: Aspect Validator**
- **Responsibility:** Validate aspect against constitution's 14 approved aspects, suggest corrections
- **Interfaces:** `ValidateAspect(aspect string) error`, `SuggestAspect(typo string) string`
- **Dependencies:** internal/matcher/fuzzy.go (CBIN-133)

**Component 3: ID Generator**
- **Responsibility:** Generate next available aspect-scoped ID by scanning filesystem and database
- **Interfaces:** `GenerateNextID(projectKey, aspect string) (string, error)`
- **Dependencies:** RequirementID Parser, filesystem scanning, database queries

**Component 4: Migration Executor**
- **Responsibility:** Execute migration plan (rename specs, update tokens, update database)
- **Interfaces:** `PlanMigration() (*MigrationPlan, error)`, `ExecuteMigration(plan *MigrationPlan) error`
- **Dependencies:** RequirementID Parser, Aspect Validator, file I/O, database storage

## Implementation Phases

### Phase 0: Pre-Implementation Gates

**Simplicity Gate (Constitution Article V):**
- ✅ Using standard library (regexp, strings, os, path/filepath)
- ✅ Minimal dependencies (only reusing existing fuzzy matcher)
- ✅ No premature optimization (straightforward parsing and validation)
- ✅ No speculative features (focused on requirement ID format only)

**Anti-Abstraction Gate (Constitution Article V):**
- ✅ Using SQL directly (no ORM overhead)
- ✅ Using os package directly (no file system abstraction layer)
- ✅ Single representation of requirement IDs (RequirementID struct)
- ✅ No unnecessary interfaces (concrete functions)

**Test-First Gate (Constitution Article IV):**
- ✅ Test strategy defined below
- ✅ Test functions named (TestCBIN139_AspectIDParser, etc.)
- ✅ Tests will be written before implementation

**Integration-First Gate (Constitution Article VI):**
- ✅ Real filesystem testing (create actual spec directories)
- ✅ Real database testing (create test SQLite database)
- ✅ No mocks for file I/O or database operations

### Phase 1: Test Creation (Red Phase)

**Step 1.1: Create parser test file**
```go
// File: internal/reqid/parser_test.go
// CANARY: REQ=CBIN-139; FEATURE="ParserTests"; ASPECT=Engine; STATUS=STUB; TEST=TestCBIN139_AspectIDParser; UPDATED=2025-10-16
package reqid

import "testing"

func TestCBIN139_AspectIDParser(t *testing.T) {
    tests := []struct {
        name    string
        reqID   string
        want    *RequirementID
        wantErr bool
    }{
        // New format tests
        {
            name:  "new format CLI",
            reqID: "CBIN-CLI-001",
            want:  &RequirementID{Key: "CBIN", Aspect: "CLI", ID: "001", Format: "v2"},
        },
        {
            name:  "new format Engine with higher ID",
            reqID: "CBIN-Engine-042",
            want:  &RequirementID{Key: "CBIN", Aspect: "Engine", ID: "042", Format: "v2"},
        },
        // Old format tests (backward compatibility)
        {
            name:  "old format",
            reqID: "CBIN-001",
            want:  &RequirementID{Key: "CBIN", ID: "001", Format: "v1"},
        },
        // Error cases
        {
            name:    "invalid aspect",
            reqID:   "CBIN-InvalidAspect-001",
            wantErr: true,
        },
        {
            name:    "missing ID segment",
            reqID:   "CBIN-CLI",
            wantErr: true,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got, err := ParseRequirementID(tt.reqID)
            // Assertions - EXPECTED TO FAIL initially
        })
    }
}
```

**Step 1.2: Create validator test file**
```go
// File: internal/reqid/validator_test.go
// CANARY: REQ=CBIN-139; FEATURE="ValidatorTests"; ASPECT=Engine; STATUS=STUB; TEST=TestCBIN139_AspectValidation; UPDATED=2025-10-16
package reqid

import "testing"

func TestCBIN139_AspectValidation(t *testing.T) {
    tests := []struct {
        aspect  string
        wantErr bool
    }{
        {"API", false},
        {"CLI", false},
        {"Engine", false},
        {"Storage", false},
        {"api", false},      // Case-insensitive
        {"Frontend", true},  // Typo (should be FrontEnd)
        {"Invalid", true},
    }

    for _, tt := range tests {
        t.Run(tt.aspect, func(t *testing.T) {
            err := ValidateAspect(tt.aspect)
            // Assertions - EXPECTED TO FAIL initially
        })
    }
}

func TestCBIN139_AspectSuggestion(t *testing.T) {
    tests := []struct {
        typo string
        want string
    }{
        {"Frontend", "FrontEnd"},
        {"Engin", "Engine"},
        {"Storge", "Storage"},
    }

    for _, tt := range tests {
        t.Run(tt.typo, func(t *testing.T) {
            got := SuggestAspect(tt.typo)
            // Assertions - EXPECTED TO FAIL initially
        })
    }
}
```

**Step 1.3: Create generator test file**
```go
// File: internal/reqid/generator_test.go
// CANARY: REQ=CBIN-139; FEATURE="GeneratorTests"; ASPECT=Engine; STATUS=STUB; TEST=TestCBIN139_AspectScopedIDGen; UPDATED=2025-10-16
package reqid

import (
    "os"
    "path/filepath"
    "testing"
)

func TestCBIN139_AspectScopedIDGen(t *testing.T) {
    // Create temporary .canary/specs directory
    tmpDir := t.TempDir()
    specsDir := filepath.Join(tmpDir, ".canary", "specs")
    os.MkdirAll(specsDir, 0755)

    // Create some existing specs
    os.MkdirAll(filepath.Join(specsDir, "CBIN-CLI-001-feature1"), 0755)
    os.MkdirAll(filepath.Join(specsDir, "CBIN-CLI-003-feature2"), 0755) // Gap at 002
    os.MkdirAll(filepath.Join(specsDir, "CBIN-API-001-feature3"), 0755)

    tests := []struct {
        aspect string
        want   string
    }{
        {"CLI", "CBIN-CLI-004"},  // Next after 003 (ignores gap)
        {"API", "CBIN-API-002"},  // Next after 001
        {"Engine", "CBIN-Engine-001"}, // First for this aspect
    }

    for _, tt := range tests {
        t.Run(tt.aspect, func(t *testing.T) {
            got, err := GenerateNextID("CBIN", tt.aspect)
            // Assertions - EXPECTED TO FAIL initially
        })
    }
}
```

**Step 1.4: Create migration test file**
```go
// File: cmd/canary/migrate_ids_test.go
// CANARY: REQ=CBIN-139; FEATURE="MigrationTests"; ASPECT=CLI; STATUS=STUB; TEST=TestCBIN139_Migration; UPDATED=2025-10-16
package main

import (
    "os"
    "path/filepath"
    "testing"
)

func TestCBIN139_Migration_DryRun(t *testing.T) {
    // Create test project structure
    tmpDir := t.TempDir()
    specsDir := filepath.Join(tmpDir, ".canary", "specs")
    os.MkdirAll(specsDir, 0755)

    // Create spec with old format
    specDir := filepath.Join(specsDir, "CBIN-101-scanner-core")
    os.MkdirAll(specDir, 0755)

    specContent := `# CANARY: REQ=CBIN-101; FEATURE="ScannerCore"; ASPECT=Engine; STATUS=IMPL; UPDATED=2025-10-16
# Spec content
`
    os.WriteFile(filepath.Join(specDir, "spec.md"), []byte(specContent), 0644)

    // Run dry-run migration
    plan, err := PlanMigration(tmpDir)
    // Assertions - EXPECTED TO FAIL initially

    // Verify plan shows CBIN-101 → CBIN-Engine-101
    // Verify no actual file changes occurred (dry-run)
}

func TestCBIN139_Migration_Execute(t *testing.T) {
    // Similar setup
    // Run actual migration
    // Verify spec directory renamed
    // Verify spec.md token updated
    // Verify migration log created
}
```

**Step 1.5: Update CANARY tokens**
```
Add TEST= field to all feature tokens in spec.md
```

**Step 1.6: Verify tests fail**
- [ ] Run `go test ./internal/reqid`
- [ ] Run `go test ./cmd/canary -run TestCBIN139`
- [ ] Confirm all tests fail with "undefined: ParseRequirementID" etc.
- [ ] Document expected failure messages

### Phase 2: Implementation (Green Phase)

**Step 2.1: Implement RequirementID parser**
```go
// File: internal/reqid/parser.go
package reqid

import (
    "fmt"
    "regexp"
    "strings"
)

var (
    // New format: CBIN-CLI-001, CBIN-Engine-042
    newFormatRegex = regexp.MustCompile(`^([A-Z]+)-([A-Za-z]+)-(\d{3})$`)
    // Old format: CBIN-001, CBIN-138
    oldFormatRegex = regexp.MustCompile(`^([A-Z]+)-(\d{3})$`)
)

type RequirementID struct {
    Key    string // CBIN
    Aspect string // CLI, API, Engine (empty for old format)
    ID     string // 001, 042, 138
    Format string // "v1" or "v2"
}

func (r *RequirementID) String() string {
    if r.Format == "v2" && r.Aspect != "" {
        return fmt.Sprintf("%s-%s-%s", r.Key, r.Aspect, r.ID)
    }
    return fmt.Sprintf("%s-%s", r.Key, r.ID)
}

func ParseRequirementID(reqID string) (*RequirementID, error) {
    // Try new format first
    if matches := newFormatRegex.FindStringSubmatch(reqID); matches != nil {
        aspect := matches[2]
        // Validate aspect
        if err := ValidateAspect(aspect); err != nil {
            return nil, fmt.Errorf("invalid aspect in %s: %w", reqID, err)
        }
        return &RequirementID{
            Key:    matches[1],
            Aspect: aspect,
            ID:     matches[3],
            Format: "v2",
        }, nil
    }

    // Try old format
    if matches := oldFormatRegex.FindStringSubmatch(reqID); matches != nil {
        return &RequirementID{
            Key:    matches[1],
            ID:     matches[2],
            Format: "v1",
        }, nil
    }

    return nil, fmt.Errorf("invalid requirement ID format: %s (expected CBIN-XXX or CBIN-<ASPECT>-XXX)", reqID)
}
```

**Step 2.2: Implement aspect validator**
```go
// File: internal/reqid/validator.go
package reqid

import (
    "fmt"
    "strings"

    "go.spyder.org/canary/internal/matcher" // CBIN-133
)

var validAspects = map[string]bool{
    "API":      true,
    "CLI":      true,
    "Engine":   true,
    "Storage":  true,
    "Security": true,
    "Docs":     true,
    "Wire":     true,
    "Planner":  true,
    "Decode":   true,
    "Encode":   true,
    "RoundTrip": true,
    "Bench":    true,
    "FrontEnd": true,
    "Dist":     true,
}

func ValidateAspect(aspect string) error {
    // Case-insensitive validation
    for valid := range validAspects {
        if strings.EqualFold(aspect, valid) {
            return nil
        }
    }

    // Invalid aspect - suggest correction
    suggestion := SuggestAspect(aspect)
    if suggestion != "" {
        return fmt.Errorf("invalid aspect %q, did you mean %q?", aspect, suggestion)
    }

    return fmt.Errorf("invalid aspect %q (valid: API, CLI, Engine, Storage, Security, Docs, Wire, Planner, Decode, Encode, RoundTrip, Bench, FrontEnd, Dist)", aspect)
}

func SuggestAspect(typo string) string {
    // Use CBIN-133 fuzzy matcher
    aspectList := make([]string, 0, len(validAspects))
    for aspect := range validAspects {
        aspectList = append(aspectList, aspect)
    }

    matches := matcher.FindBestMatches(typo, aspectList, 1)
    if len(matches) > 0 && matches[0].Score > 0.6 {
        return matches[0].Text
    }

    return ""
}
```

**Step 2.3: Implement ID generator**
```go
// File: internal/reqid/generator.go
package reqid

import (
    "fmt"
    "os"
    "path/filepath"
    "regexp"
    "strconv"
)

func GenerateNextID(projectKey, aspect string) (string, error) {
    // Validate aspect first
    if err := ValidateAspect(aspect); err != nil {
        return "", err
    }

    // Normalize aspect to constitution casing
    aspect = NormalizeAspect(aspect)

    // Scan filesystem for highest ID
    specsDir := ".canary/specs"
    maxID := 0

    // Pattern: CBIN-CLI-001-feature-name
    pattern := regexp.MustCompile(fmt.Sprintf(`^%s-%s-(\d{3})`, projectKey, aspect))

    entries, err := os.ReadDir(specsDir)
    if err != nil && !os.IsNotExist(err) {
        return "", fmt.Errorf("read specs directory: %w", err)
    }

    for _, entry := range entries {
        if !entry.IsDir() {
            continue
        }

        matches := pattern.FindStringSubmatch(entry.Name())
        if len(matches) > 1 {
            id, err := strconv.Atoi(matches[1])
            if err == nil && id > maxID {
                maxID = id
            }
        }
    }

    // Generate next ID
    nextID := maxID + 1
    return fmt.Sprintf("%s-%s-%03d", projectKey, aspect, nextID), nil
}

func NormalizeAspect(aspect string) string {
    // Find exact casing from validAspects map
    for valid := range validAspects {
        if strings.EqualFold(aspect, valid) {
            return valid
        }
    }
    return aspect // Return as-is if not found (will fail validation)
}
```

**Step 2.4: Implement migration command**
```go
// File: cmd/canary/migrate_ids.go
package main

import (
    "encoding/json"
    "fmt"
    "os"
    "path/filepath"
    "regexp"
    "strings"
    "time"

    "github.com/spf13/cobra"
    "go.spyder.org/canary/internal/reqid"
    "go.spyder.org/canary/internal/storage"
)

type MigrationRecord struct {
    OldID         string `json:"old_id"`
    NewID         string `json:"new_id"`
    Aspect        string `json:"aspect"`
    SpecPathOld   string `json:"spec_path_old"`
    SpecPathNew   string `json:"spec_path_new"`
    TokensUpdated int    `json:"tokens_updated"`
}

type MigrationPlan struct {
    Records   []MigrationRecord `json:"records"`
    Timestamp string            `json:"timestamp"`
}

var migrateIDsCmd = &cobra.Command{
    Use:   "migrate-ids [--dry-run] [--confirm]",
    Short: "Migrate requirement IDs from CBIN-XXX to CBIN-<ASPECT>-XXX format",
    Long: `Migrate all existing requirements from old format (CBIN-XXX) to new aspect-based format (CBIN-<ASPECT>-XXX).

This command:
- Scans all spec directories in .canary/specs/
- Extracts ASPECT from spec.md CANARY tokens
- Renames directories from CBIN-XXX-feature to CBIN-<ASPECT>-XXX-feature
- Updates all CANARY tokens in spec files and source code
- Updates database req_id column
- Creates backup before migration
- Generates migration log

Use --dry-run to preview changes without applying them.
Use --confirm to execute migration.`,
    RunE: func(cmd *cobra.Command, args []string) error {
        dryRun, _ := cmd.Flags().GetBool("dry-run")
        confirm, _ := cmd.Flags().GetBool("confirm")

        if !dryRun && !confirm {
            return fmt.Errorf("must specify either --dry-run or --confirm")
        }

        // Plan migration
        plan, err := PlanMigration()
        if err != nil {
            return fmt.Errorf("plan migration: %w", err)
        }

        // Show plan summary
        fmt.Printf("📋 Migration Plan\n\n")
        fmt.Printf("Will migrate %d requirements:\n\n", len(plan.Records))
        for _, record := range plan.Records {
            fmt.Printf("  %s → %s (aspect: %s)\n", record.OldID, record.NewID, record.Aspect)
        }

        if dryRun {
            fmt.Println("\n🔍 Dry run complete - no changes made")
            return nil
        }

        // Execute migration
        fmt.Println("\n🚀 Executing migration...")

        // Create backup
        backupPath, err := CreateBackup()
        if err != nil {
            return fmt.Errorf("create backup: %w", err)
        }
        fmt.Printf("✅ Created backup: %s\n", backupPath)

        // Execute migration
        if err := ExecuteMigration(plan); err != nil {
            return fmt.Errorf("execute migration: %w", err)
        }

        // Write migration log
        logPath := ".canary/memory/migration-log.json"
        if err := WriteMigrationLog(plan, logPath); err != nil {
            return fmt.Errorf("write migration log: %w", err)
        }

        fmt.Printf("\n✅ Migration complete!\n")
        fmt.Printf("  - Migrated %d requirements\n", len(plan.Records))
        fmt.Printf("  - Backup: %s\n", backupPath)
        fmt.Printf("  - Log: %s\n", logPath)

        return nil
    },
}

func PlanMigration() (*MigrationPlan, error) {
    // Implementation: scan .canary/specs, extract aspects, plan renames
}

func ExecuteMigration(plan *MigrationPlan) error {
    // Implementation: rename directories, update tokens, update database
}

func CreateBackup() (string, error) {
    // Implementation: tar.gz .canary/specs and database
}
```

**Step 2.5: Create database schema migration**
```sql
-- File: internal/storage/migrations/000002_extend_reqid.up.sql
-- CANARY: REQ=CBIN-139; FEATURE="DatabaseSchemaMigration"; ASPECT=Storage; STATUS=IMPL; UPDATED=2025-10-16
-- Extend req_id column to accommodate aspect-based IDs (up to 20 chars for CBIN-FrontEnd-001)

-- SQLite doesn't support ALTER COLUMN, so we need to:
-- 1. Create new table with larger req_id
-- 2. Copy data
-- 3. Drop old table
-- 4. Rename new table

CREATE TABLE IF NOT EXISTS tokens_new (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    req_id TEXT NOT NULL,  -- Extended from implicit limit to explicit unlimited TEXT
    feature TEXT NOT NULL,
    aspect TEXT NOT NULL,
    status TEXT NOT NULL,
    file_path TEXT NOT NULL,
    line_number INTEGER NOT NULL,
    test TEXT,
    bench TEXT,
    owner TEXT,
    priority INTEGER DEFAULT 5,
    phase TEXT,
    keywords TEXT,
    spec_status TEXT DEFAULT 'draft',
    created_at TEXT,
    updated_at TEXT NOT NULL,
    started_at TEXT,
    completed_at TEXT,
    commit_hash TEXT,
    branch TEXT,
    depends_on TEXT,
    blocks TEXT,
    related_to TEXT,
    raw_token TEXT NOT NULL,
    indexed_at TEXT NOT NULL,
    UNIQUE(req_id, feature, file_path, line_number)
);

INSERT INTO tokens_new SELECT * FROM tokens;
DROP TABLE tokens;
ALTER TABLE tokens_new RENAME TO tokens;

-- Recreate indexes
CREATE INDEX IF NOT EXISTS idx_tokens_req_id ON tokens(req_id);
CREATE INDEX IF NOT EXISTS idx_tokens_status ON tokens(status);
CREATE INDEX IF NOT EXISTS idx_tokens_priority ON tokens(priority);
CREATE INDEX IF NOT EXISTS idx_tokens_aspect ON tokens(aspect);
CREATE INDEX IF NOT EXISTS idx_tokens_spec_status ON tokens(spec_status);
CREATE INDEX IF NOT EXISTS idx_tokens_phase ON tokens(phase);
CREATE INDEX IF NOT EXISTS idx_tokens_keywords ON tokens(keywords);
```

**Step 2.6: Update templates**
- Update `.canary/templates/spec-template.md`: Replace `CBIN-XXX` with `CBIN-<ASPECT>-XXX` placeholders
- Update `.canary/templates/plan-template.md`: Replace `CBIN-XXX` with `CBIN-<ASPECT>-XXX` placeholders
- Update `.claude/commands/canary.specify.md`: Document aspect requirement
- Update `.claude/commands/canary.plan.md`: Show aspect-based examples

**Step 2.7: Update canary create and canary specify commands**
- Modify `cmd/canary/main.go` specifyCmd to use `GenerateNextID` with aspect detection
- Add aspect extraction from feature description or prompt user for aspect

**Step 2.8: Update CANARY token STATUS**
```
Update all CANARY tokens in spec.md from STATUS=STUB to STATUS=IMPL
Update tokens in implementation files with TEST= field
```

**Step 2.9: Verify tests pass**
- [ ] Run `go test ./internal/reqid`
- [ ] Run `go test ./cmd/canary -run TestCBIN139`
- [ ] All tests pass
- [ ] No regressions in other tests

### Phase 3: Database Schema Migration (if applicable)

**Step 3.1: Update LatestVersion constant**
```go
// File: internal/storage/db.go
const LatestVersion = 2 // Update from 1 to 2
```

**Step 3.2: Test database migration**
```go
func TestCBIN139_DatabaseMigration(t *testing.T) {
    // Create v1 database
    // Apply migration 000002
    // Verify schema updated
    // Verify data preserved
}
```

**Step 3.3: Test rollback**
```sql
-- File: internal/storage/migrations/000002_extend_reqid.down.sql
-- Rollback: restore original schema (data may be truncated if IDs exceed old limit)

CREATE TABLE IF NOT EXISTS tokens_old (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    req_id TEXT NOT NULL,
    feature TEXT NOT NULL,
    aspect TEXT NOT NULL,
    status TEXT NOT NULL,
    file_path TEXT NOT NULL,
    line_number INTEGER NOT NULL,
    test TEXT,
    bench TEXT,
    owner TEXT,
    priority INTEGER DEFAULT 5,
    phase TEXT,
    keywords TEXT,
    spec_status TEXT DEFAULT 'draft',
    created_at TEXT,
    updated_at TEXT NOT NULL,
    started_at TEXT,
    completed_at TEXT,
    commit_hash TEXT,
    branch TEXT,
    depends_on TEXT,
    blocks TEXT,
    related_to TEXT,
    raw_token TEXT NOT NULL,
    indexed_at TEXT NOT NULL,
    UNIQUE(req_id, feature, file_path, line_number)
);

INSERT INTO tokens_old SELECT * FROM tokens;
DROP TABLE tokens;
ALTER TABLE tokens_old RENAME TO tokens;

-- Recreate indexes (same as before)
```

### Phase 4: Integration Testing

**Step 4.1: End-to-end migration test**
```bash
# Create test project
canary init test-project --key TEST
cd test-project

# Create old-format specs
canary specify "Feature A"  # TEST-001
canary specify "Feature B"  # TEST-002

# Manually edit specs to set ASPECT fields

# Run migration
canary migrate-ids --dry-run
canary migrate-ids --confirm

# Verify all specs renamed
# Verify database updated
# Verify migration log exists
```

**Step 4.2: Aspect filtering test**
```bash
# After migration
canary list --aspect CLI
canary search --aspect Engine "feature"
canary next --aspect API
```

**Step 4.3: New ID generation test**
```bash
# After migration
canary specify "New Feature"  # Should generate aspect-based ID
# Expected: TEST-<ASPECT>-003 or TEST-<ASPECT>-001 (depending on aspect)
```

## Testing Strategy

### Unit Tests
**Test:** `TestCBIN139_AspectIDParser`
**Coverage:** Parse new format, parse old format, error handling
**Test Cases:**
- ✅ Parse CBIN-CLI-001 (new format)
- ✅ Parse CBIN-Engine-042 (new format)
- ✅ Parse CBIN-001 (old format, backward compat)
- ✅ Invalid aspect (CBIN-InvalidAspect-001)
- ✅ Missing segments (CBIN-CLI, CBIN-001-)
- ✅ Case-insensitive aspect parsing (CBIN-cli-001 → CBIN-CLI-001)

**Test:** `TestCBIN139_AspectValidation`
**Coverage:** Validate all 14 aspects, case-insensitive, error messages
**Test Cases:**
- ✅ All valid aspects (API, CLI, Engine, Storage, Security, Docs, Wire, Planner, Decode, Encode, RoundTrip, Bench, FrontEnd, Dist)
- ✅ Case variations (api, Api, API all valid)
- ✅ Invalid aspects with suggestions (Frontend → FrontEnd)
- ✅ Invalid aspects without suggestions (XYZ)

**Test:** `TestCBIN139_AspectScopedIDGen`
**Coverage:** Aspect-scoped incrementing, gap handling, new aspects
**Test Cases:**
- ✅ Generate CBIN-CLI-002 when CLI-001 exists
- ✅ Generate CBIN-CLI-004 when CLI-001, CLI-003 exist (gap at 002)
- ✅ Generate CBIN-Engine-001 when no Engine specs exist
- ✅ Generate independent IDs for different aspects

### Integration Tests
**Test:** `TestCBIN139_Migration`
**Coverage:** Full migration workflow (plan, backup, execute, log)
**Test Cases:**
- ✅ Dry-run produces correct plan without changes
- ✅ Migration renames spec directories correctly
- ✅ Migration updates spec.md tokens
- ✅ Migration updates source code tokens
- ✅ Migration updates database req_id
- ✅ Migration creates backup tar.gz
- ✅ Migration writes migration log JSON
- ✅ Migration is idempotent (can re-run safely)

**Test:** `TestCBIN139_AspectFilter`
**Coverage:** Aspect filtering in list/search/next commands
**Test Cases:**
- ✅ `canary list --aspect CLI` returns only CLI requirements
- ✅ `canary search --aspect Engine "query"` searches only Engine requirements
- ✅ `canary next --aspect API` prioritizes API requirements
- ✅ Invalid aspect returns helpful error

### Acceptance Tests
**Based on spec success criteria:**
- [ ] Migration completes in < 5 minutes for 138+ requirements (Performance)
- [ ] Zero data loss during migration (Reliability)
- [ ] Aspect-scoped queries execute in < 1 second (Performance)
- [ ] New ID format visually distinguishes aspects (UX)

### Performance Benchmarks
Not applicable for this feature (migration is one-time operation, ID parsing is trivial performance)

## Constitutional Compliance

### Article I: Requirement-First Development
- ✅ CANARY token CBIN-139 defined
- ✅ Tokens placed in all implementation files (parser.go, validator.go, generator.go, migrate_ids.go)
- ✅ All tokens include REQ, FEATURE, ASPECT, STATUS, UPDATED fields

### Article IV: Test-First Imperative
- ✅ Tests written before implementation (Phase 1)
- ✅ Tests fail initially (red phase)
- ✅ Implementation makes tests pass (green phase)
- ✅ All tests named with TestCBIN139_ prefix

### Article V: Simplicity and Anti-Abstraction
- ✅ Using standard library (regexp, strings, os, filepath)
- ✅ No new dependencies (reusing existing fuzzy matcher)
- ✅ Direct SQL for database operations (no ORM)
- ✅ Direct file I/O (no abstraction layer)

### Article VI: Integration-First Testing
- ✅ Real filesystem testing (create actual spec directories in tests)
- ✅ Real database testing (create test SQLite database)
- ✅ No mocks for file I/O or database

### Article VII: Documentation Currency
- ✅ CANARY tokens include OWNER field
- ✅ UPDATED field maintained (2025-10-16)
- ✅ Status progression: STUB → IMPL → TESTED

## Complexity Tracking

### Justified Complexity
**Exception:** Migration command involves multiple file system operations (rename directories, update tokens in many files, update database)
**Justification:** Migration is inherently complex - must atomically update specs, source code, and database. Backup/rollback adds necessary safety.
**Constitutional Article:** Article V (Simplicity)
**Mitigation:**
- Break migration into clear phases (plan, backup, execute, log)
- Use dry-run mode for preview
- Create backup before any changes
- Write migration log for audit trail
- Keep migration logic in separate package/file

**Exception:** Backward compatibility parser must handle both old and new formats
**Justification:** Critical for gradual migration - spec requires 3-month grace period
**Constitutional Article:** Article V (Simplicity)
**Mitigation:**
- Use clear regex patterns for each format
- Try new format first (most common after migration)
- Fall back to old format
- Centralize parsing logic in single function

### Dependencies Added
- None (reusing existing internal/matcher/fuzzy.go for aspect suggestions)

## Implementation Checklist

- [ ] Phase 0 gates all passed
- [ ] Parser test file created (internal/reqid/parser_test.go)
- [ ] Validator test file created (internal/reqid/validator_test.go)
- [ ] Generator test file created (internal/reqid/generator_test.go)
- [ ] Migration test file created (cmd/canary/migrate_ids_test.go)
- [ ] Tests fail initially (red)
- [ ] Parser implementation created (internal/reqid/parser.go)
- [ ] Validator implementation created (internal/reqid/validator.go)
- [ ] Generator implementation created (internal/reqid/generator.go)
- [ ] Migration command created (cmd/canary/migrate_ids.go)
- [ ] Database migration created (000002_extend_reqid.up/down.sql)
- [ ] Templates updated (spec-template.md, plan-template.md)
- [ ] Tests pass (green)
- [ ] CANARY tokens updated with TEST= field
- [ ] Token STATUS updated to TESTED
- [ ] Database migration tested
- [ ] End-to-end migration tested
- [ ] All acceptance criteria met
- [ ] Constitutional compliance verified
- [ ] Ready for code review

---

## Next Steps

1. Review this plan for accuracy and completeness
2. Use `/canary.implement CBIN-139` to execute implementation following TDD workflow
3. Run `canary index` after implementation to rebuild database
4. Run `canary scan` to verify token status
5. Run `canary migrate-ids --dry-run` to test migration on current project
6. Run `canary migrate-ids --confirm` to execute migration
7. Verify GAP_ANALYSIS.md and update with CBIN-139 completion

## Decision Log

**Decision 1: Preserve original ID numbers during migration**
- **Rationale:** Spec clarification recommends Option B (preserve numbers) for git history traceability
- **Impact:** CBIN-101 with ASPECT=Engine becomes CBIN-Engine-101 (not CBIN-Engine-001)
- **Tradeoff:** Less "clean" aspect-scoped numbering, but easier debugging and git blame

**Decision 2: Store migration mapping in migration-log.json**
- **Rationale:** Spec clarification recommends Option C (bidirectional lookup)
- **Impact:** Enables tools to understand old→new ID mapping
- **Tradeoff:** Minimal overhead (~10KB JSON file), high observability value

**Decision 3: 3-month grace period for old format**
- **Rationale:** Spec clarification recommends Option C (grace period)
- **Impact:** Parser must support both formats until deprecation date
- **Tradeoff:** Parser complexity, but allows discovery of missed migrations
