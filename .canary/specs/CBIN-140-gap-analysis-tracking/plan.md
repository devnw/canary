# Implementation Plan: CBIN-140 Gap Analysis Tracking

**Requirement ID:** CBIN-140
**Feature:** GapAnalysisTracking
**Created:** 2025-10-17
**Status:** Ready for Implementation

## Executive Summary

This plan implements a gap analysis tracking system that enables AI agents and developers to record, query, and learn from implementation mistakes. The system stores gap data in SQLite, provides CLI commands for gap management, and auto-injects relevant gaps into planning prompts to prevent repeated mistakes.

## Clarification Decisions

### Decision 1: Gap Data Storage Location
**Selected:** Option C - Support both with configuration flag
**Rationale:**
- Default to SQLite database (already in use per constitution Article V - simplicity)
- Allow export/import for team sharing via version control
- Provides flexibility without requiring cloud infrastructure
- Aligns with existing storage patterns in `internal/storage/db.go`

### Decision 2: Gap Prioritization Strategy
**Selected:** Option D - Configurable ranking
**Rationale:**
- Default ranking: `helpful_count DESC, created_at DESC` (most useful + recent first)
- Allow override via config for different workflows
- Supports constitution Article VIII (metrics-driven development)
- Enables future refinement based on actual usage patterns

## Tech Stack Decision

**Language:** Go 1.21+
**Database:** SQLite (existing - `modernc.org/sqlite`)
**Migration:** golang-migrate/v4 (existing pattern)
**CLI Framework:** Cobra (existing - used in `cmd/canary/`)
**Testing:** Go standard library `testing` package

**Rationale:**
- ✅ Article V: Use existing stack (simplicity, no new dependencies)
- ✅ Reuse proven patterns from `internal/storage/`
- ✅ SQLite already handles 10,000+ CANARY tokens efficiently
- ✅ Cobra provides consistent CLI UX with existing commands

## Architecture Overview

```
┌─────────────────────────────────────────────────────────────┐
│                    CLI Layer (cmd/canary/)                  │
│  ┌──────────┐  ┌──────────┐  ┌──────────┐  ┌───────────┐  │
│  │ gap mark │  │gap query │  │gap report│  │gap helpful│  │
│  └────┬─────┘  └────┬─────┘  └────┬─────┘  └─────┬─────┘  │
└───────┼────────────────┼──────────────┼──────────────┼──────┘
        │                │              │              │
        └────────┬───────┴──────┬───────┴──────────────┘
                 ▼              ▼
        ┌────────────────────────────────────┐
        │   Gap Service (internal/gap/)      │
        │  ┌──────────────────────────────┐  │
        │  │ GapService                   │  │
        │  │ - Mark(req, feature, cat, desc)│
        │  │ - Query(filters) []Gap       │  │
        │  │ - Rate(gapID, helpful bool)  │  │
        │  │ - Report(groupBy) Report     │  │
        │  └──────────────┬───────────────┘  │
        └─────────────────┼──────────────────┘
                          ▼
        ┌─────────────────────────────────────┐
        │  Storage Layer (internal/storage/)  │
        │  ┌───────────────────────────────┐  │
        │  │ GapRepository                 │  │
        │  │ - CreateGap(gap) int64       │  │
        │  │ - GetGap(id) Gap             │  │
        │  │ - ListGaps(filter) []Gap     │  │
        │  │ - UpdateGap(gap)             │  │
        │  │ - IncrementHelpful(id)       │  │
        │  └───────────────┬───────────────┘  │
        └──────────────────┼──────────────────┘
                           ▼
          ┌────────────────────────────────┐
          │  SQLite Database               │
          │  Tables:                       │
          │  - gap_entries                 │
          │  - gap_categories              │
          │  - gap_configurations          │
          └────────────────────────────────┘
```

## Database Schema

### Migration 004: Gap Analysis Tables

```sql
-- 004_create_gap_tables.up.sql
CREATE TABLE IF NOT EXISTS gap_categories (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL UNIQUE,
    description TEXT,
    priority INTEGER DEFAULT 5,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Seed standard categories
INSERT INTO gap_categories (name, description, priority) VALUES
    ('implementation', 'Logic or algorithm implementation errors', 8),
    ('testing', 'Missing or inadequate test coverage', 7),
    ('performance', 'Performance or efficiency issues', 6),
    ('security', 'Security vulnerabilities or concerns', 10),
    ('design', 'API or architectural design problems', 7),
    ('documentation', 'Missing or incorrect documentation', 4);

CREATE TABLE IF NOT EXISTS gap_entries (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    gap_id TEXT NOT NULL UNIQUE, -- Format: GAP-NNN
    req_id TEXT NOT NULL, -- e.g., CBIN-042
    feature TEXT NOT NULL, -- Feature name from token
    aspect TEXT, -- CANARY aspect (API, CLI, etc.)
    category_id INTEGER NOT NULL,
    description TEXT NOT NULL,
    corrective_action TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    created_by TEXT DEFAULT 'unknown',
    helpful_count INTEGER DEFAULT 0,
    unhelpful_count INTEGER DEFAULT 0,
    FOREIGN KEY (category_id) REFERENCES gap_categories(id)
);

CREATE INDEX idx_gap_entries_req_id ON gap_entries(req_id);
CREATE INDEX idx_gap_entries_aspect ON gap_entries(aspect);
CREATE INDEX idx_gap_entries_category ON gap_entries(category_id);
CREATE INDEX idx_gap_entries_created_at ON gap_entries(created_at DESC);
CREATE INDEX idx_gap_entries_helpful ON gap_entries(helpful_count DESC);

CREATE TABLE IF NOT EXISTS gap_configurations (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    agent_id TEXT NOT NULL UNIQUE DEFAULT 'default',
    enabled_categories TEXT NOT NULL DEFAULT 'implementation,testing,performance,security,design', -- CSV
    max_gaps_per_query INTEGER DEFAULT 10,
    similarity_threshold REAL DEFAULT 0.7,
    ranking_strategy TEXT DEFAULT 'helpful_recent', -- helpful_recent, recent, category_priority
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Insert default configuration
INSERT INTO gap_configurations (agent_id) VALUES ('default');
```

```sql
-- 004_create_gap_tables.down.sql
DROP TABLE IF EXISTS gap_configurations;
DROP TABLE IF EXISTS gap_entries;
DROP TABLE IF NOT EXISTS gap_categories;
```

## CANARY Token Placement

**Primary Token:**
```go
// File: internal/gap/service.go
// CANARY: REQ=CBIN-140; FEATURE="GapAnalysisTracking"; ASPECT=CLI; STATUS=IMPL; OWNER=canary; UPDATED=2025-10-17
package gap

// GapService provides gap analysis tracking operations
type GapService struct {
    repo *storage.GapRepository
}
```

**Sub-Feature Tokens:**
```go
// File: cmd/canary/gap_mark.go
// CANARY: REQ=CBIN-140; FEATURE="GapMarkCmd"; ASPECT=CLI; STATUS=IMPL; TEST=TestGapMarkCommand; UPDATED=2025-10-17

// File: cmd/canary/gap_query.go
// CANARY: REQ=CBIN-140; FEATURE="GapQueryCmd"; ASPECT=CLI; STATUS=IMPL; TEST=TestGapQueryCommand; UPDATED=2025-10-17

// File: cmd/canary/gap_report.go
// CANARY: REQ=CBIN-140; FEATURE="GapReportCmd"; ASPECT=CLI; STATUS=IMPL; TEST=TestGapReportCommand; UPDATED=2025-10-17

// File: cmd/canary/gap_helpful.go
// CANARY: REQ=CBIN-140; FEATURE="GapHelpfulCmd"; ASPECT=CLI; STATUS=IMPL; TEST=TestGapHelpfulCommand; UPDATED=2025-10-17

// File: internal/storage/gap_repository.go
// CANARY: REQ=CBIN-140; FEATURE="GapStorage"; ASPECT=Storage; STATUS=IMPL; TEST=TestGapRepository; UPDATED=2025-10-17

// File: internal/gap/query.go
// CANARY: REQ=CBIN-140; FEATURE="GapQuery"; ASPECT=Storage; STATUS=IMPL; TEST=TestGapQueryEngine; UPDATED=2025-10-17
```

## Implementation Phases

### Phase 0: Pre-Implementation Gates ✅

**Constitutional Compliance:**
- [x] **Article I (Requirement-First):** CANARY tokens defined above
- [x] **Article IV (Test-First):** Tests written before implementation (see Phase 1)
- [x] **Article V (Simplicity):** Reusing existing SQLite, no new dependencies
- [x] **Article VI (Integration-First):** Real SQLite database in tests (not mocks)

**Complexity Justification:**
- Gap ID generation (`GAP-001`, `GAP-002`): Simple counter, similar to existing REQ-ID pattern
- Ranking algorithm: SQL `ORDER BY` with configurable columns, no custom complexity
- Plan injection: Template modification only, leverages existing prompt system

**Readiness Checklist:**
- [x] Specification reviewed and clarifications resolved
- [x] Constitution principles mapped to design decisions
- [x] Database schema designed with indexes for performance (2-second query target)
- [x] Token placement identified in existing code structure

---

### Phase 1: Database Schema & Migration (Test-First) 🔴

**Test File:** `internal/storage/gap_repository_test.go`

**Tests to Write (Red Phase):**
```go
func TestGapRepository_CreateGap(t *testing.T) {
    // Setup: Create test database with migration 004
    // Test: Insert gap entry with all required fields
    // Assert: Gap ID generated, all fields persisted correctly
}

func TestGapRepository_GetGap(t *testing.T) {
    // Setup: Create gap entry
    // Test: Retrieve by gap_id (e.g., GAP-001)
    // Assert: All fields match, timestamps present
}

func TestGapRepository_ListGaps_ByRequirement(t *testing.T) {
    // Setup: Create 3 gaps for CBIN-042, 2 for CBIN-050
    // Test: Query gaps for CBIN-042
    // Assert: Returns exactly 3 gaps, correct req_id
}

func TestGapRepository_ListGaps_ByCategory(t *testing.T) {
    // Setup: Create gaps in security, performance categories
    // Test: Query security category
    // Assert: Only security gaps returned
}

func TestGapRepository_IncrementHelpful(t *testing.T) {
    // Setup: Create gap with helpful_count=0
    // Test: Call IncrementHelpful twice
    // Assert: helpful_count=2
}

func TestGapRepository_QueryPerformance(t *testing.T) {
    // Setup: Insert 1000 gap entries
    // Test: Query with filters
    // Assert: Query completes in < 2 seconds (success criteria)
}
```

**Implementation Steps:**
1. Create `internal/storage/migrations/004_create_gap_tables.up.sql`
2. Create `internal/storage/migrations/004_create_gap_tables.down.sql`
3. Update `internal/storage/db.go` `LatestVersion` to 4
4. Run tests - **VERIFY ALL FAIL** (Red Phase ✅)

**Token Update After Red Phase:**
```go
// CANARY: REQ=CBIN-140; FEATURE="GapSchema"; ASPECT=Storage; STATUS=STUB; TEST=TestGapRepository; UPDATED=2025-10-17
```

---

### Phase 2: Repository Implementation (Green Phase) 🟢

**File:** `internal/storage/gap_repository.go`

**Implementation:**
```go
package storage

type GapRepository struct {
    db *sqlx.DB
}

func NewGapRepository(db *sqlx.DB) *GapRepository {
    return &GapRepository{db: db}
}

// CreateGap inserts a new gap entry and returns the generated gap_id
func (r *GapRepository) CreateGap(gap *GapEntry) (string, error) {
    // Generate next GAP-ID (GAP-001, GAP-002, etc.)
    // Insert into gap_entries table
    // Return gap_id
}

// GetGap retrieves a gap entry by gap_id
func (r *GapRepository) GetGap(gapID string) (*GapEntry, error) {
    // SELECT from gap_entries WHERE gap_id = ?
}

// ListGaps retrieves gaps matching filter criteria
func (r *GapRepository) ListGaps(filter *GapFilter) ([]*GapEntry, error) {
    // Build dynamic SQL query based on filter
    // Apply ORDER BY per ranking strategy
    // LIMIT based on config max_gaps_per_query
}

// IncrementHelpful increments helpful_count for a gap
func (r *GapRepository) IncrementHelpful(gapID string) error {
    // UPDATE gap_entries SET helpful_count = helpful_count + 1 WHERE gap_id = ?
}

// IncrementUnhelpful increments unhelpful_count
func (r *GapRepository) IncrementUnhelpful(gapID string) error {
    // UPDATE gap_entries SET unhelpful_count = unhelpful_count + 1 WHERE gap_id = ?
}
```

**Acceptance Criteria:**
- All tests from Phase 1 pass (Green Phase ✅)
- Query performance test confirms < 2-second response for 1000 entries
- Gap ID generation is sequential and collision-free

**Token Update After Green Phase:**
```go
// CANARY: REQ=CBIN-140; FEATURE="GapStorage"; ASPECT=Storage; STATUS=TESTED; TEST=TestGapRepository; UPDATED=2025-10-17
```

---

### Phase 3: Gap Service Layer (Test-First) 🔴

**Test File:** `internal/gap/service_test.go`

**Tests to Write (Red Phase):**
```go
func TestGapService_MarkGap(t *testing.T) {
    // Setup: Service with mock repo
    // Test: Mark CBIN-042/OAuth as incorrect with security category
    // Assert: Gap created with correct req_id, feature, category
}

func TestGapService_QueryGaps_ByAspect(t *testing.T) {
    // Setup: Gaps across multiple aspects (API, CLI, Storage)
    // Test: Query for aspect=API
    // Assert: Only API-related gaps returned
}

func TestGapService_RateGap_Helpful(t *testing.T) {
    // Setup: Existing gap GAP-005
    // Test: Rate as helpful
    // Assert: helpful_count incremented, gap effectiveness tracked
}

func TestGapService_GenerateReport(t *testing.T) {
    // Setup: 100 gaps across categories
    // Test: Generate report grouped by category
    // Assert: "Security: 15, Implementation: 40, Testing: 30"
}
```

**Implementation:** `internal/gap/service.go`

**Token Update:**
```go
// CANARY: REQ=CBIN-140; FEATURE="GapQuery"; ASPECT=Storage; STATUS=TESTED; TEST=TestGapQueryEngine; UPDATED=2025-10-17
```

---

### Phase 4: CLI Commands (Test-First) 🔴

**Test Files:**
- `cmd/canary/gap_mark_test.go`
- `cmd/canary/gap_query_test.go`
- `cmd/canary/gap_report_test.go`
- `cmd/canary/gap_helpful_test.go`

**Tests to Write (Red Phase):**
```go
func TestGapMarkCommand(t *testing.T) {
    // Test: canary gap mark CBIN-042 OAuth --category security
    // Assert: Gap created, confirmation message displayed
}

func TestGapQueryCommand(t *testing.T) {
    // Test: canary gap query --aspect API --category security
    // Assert: Matching gaps displayed with formatting
}

func TestGapReportCommand(t *testing.T) {
    // Test: canary gap report --summary
    // Assert: Report shows category breakdown
}

func TestGapHelpfulCommand(t *testing.T) {
    // Test: canary gap helpful GAP-005
    // Assert: Helpful count incremented, success message
}
```

**Implementation:**
```go
// cmd/canary/gap_mark.go
var gapMarkCmd = &cobra.Command{
    Use:   "mark REQ-ID FEATURE --category CATEGORY",
    Short: "Mark a CANARY implementation as incorrect",
    Run: func(cmd *cobra.Command, args []string) {
        // Parse args: req_id, feature, category flag
        // Prompt for description (interactive)
        // Call gapService.Mark(...)
        // Display: "✅ Gap GAP-042 created for CBIN-042/OAuth"
    },
}
```

**Token Updates:**
```go
// CANARY: REQ=CBIN-140; FEATURE="GapMarkCmd"; ASPECT=CLI; STATUS=TESTED; TEST=TestGapMarkCommand; UPDATED=2025-10-17
// CANARY: REQ=CBIN-140; FEATURE="GapQueryCmd"; ASPECT=CLI; STATUS=TESTED; TEST=TestGapQueryCommand; UPDATED=2025-10-17
// CANARY: REQ=CBIN-140; FEATURE="GapReportCmd"; ASPECT=CLI; STATUS=TESTED; TEST=TestGapReportCommand; UPDATED=2025-10-17
// CANARY: REQ=CBIN-140; FEATURE="GapHelpfulCmd"; ASPECT=CLI; STATUS=TESTED; TEST=TestGapHelpfulCommand; UPDATED=2025-10-17
```

---

### Phase 5: Plan Command Integration (Test-First) 🔴

**Test File:** `test/integration/plan_gap_injection_test.go`

**Tests to Write:**
```go
func TestPlanGapInjection_APIFeature(t *testing.T) {
    // Setup: Create gaps for aspect=API
    // Test: Run /canary.plan for CBIN-150 (API aspect)
    // Assert: Plan includes "⚠️ Gap Analysis: 3 relevant gaps found"
    // Assert: Gap details injected into plan context
}

func TestPlanGapInjection_NoRelevantGaps(t *testing.T) {
    // Setup: Gaps exist but different aspect
    // Test: Run /canary.plan for CLI feature
    // Assert: No gap warnings in plan output
}
```

**Implementation:**
1. Modify `.canary/templates/commands/plan.md`:
```markdown
## Relevant Gap Analysis

{{gap_analysis_injection aspect={{ASPECT}} max=5}}

**If gaps shown above:** Review these past mistakes before proceeding.
```

2. Implement gap injection in plan command execution

**Token Update:**
```go
// CANARY: REQ=CBIN-140; FEATURE="PlanIntegration"; ASPECT=Planner; STATUS=TESTED; TEST=TestPlanGapInjection; UPDATED=2025-10-17
```

---

### Phase 6: Integration Testing (Full Workflow) 🔴

**Test File:** `test/integration/gap_workflow_test.go`

**Tests to Write:**
```go
func TestGapWorkflow_EndToEnd(t *testing.T) {
    // Step 1: Mark implementation as incorrect
    // Step 2: Query gaps before new implementation
    // Step 3: Rate gap as helpful
    // Step 4: Generate report
    // Step 5: Verify gap appears in plan injection
    // Assert: All steps succeed, data flows correctly
}

func TestGapWorkflow_PerformanceAtScale(t *testing.T) {
    // Setup: Create 10,000 gap entries
    // Test: Query, report, injection all work
    // Assert: All operations complete in < 2 seconds (success criteria)
}
```

**Token Update:**
```go
// CANARY: REQ=CBIN-140; FEATURE="GapIntegrationTests"; ASPECT=CLI; STATUS=TESTED; TEST=TestGapWorkflow; UPDATED=2025-10-17
```

---

### Phase 7: Documentation 📝

**Files to Create:**
1. `docs/cli/gap.md` - CLI command reference
2. `docs/workflow/gap-tracking.md` - When and how to use gap analysis
3. Update `.canary/AGENT_CONTEXT.md` - Agent instructions for gap tracking

**Token Update:**
```go
// CANARY: REQ=CBIN-140; FEATURE="GapCLIDocs"; ASPECT=Docs; STATUS=IMPL; UPDATED=2025-10-17
// CANARY: REQ=CBIN-140; FEATURE="GapWorkflowGuide"; ASPECT=Docs; STATUS=IMPL; UPDATED=2025-10-17
```

---

## Testing Strategy

### Unit Tests
- **Repository Layer:** All CRUD operations, query filtering, performance
- **Service Layer:** Business logic, gap categorization, ranking
- **CLI Commands:** Argument parsing, output formatting, error handling

### Integration Tests
- **Database Migrations:** Migration 004 applies cleanly, rolls back correctly
- **End-to-End Workflow:** Mark → Query → Rate → Report → Injection
- **Performance:** 10,000 gap entries, < 2-second queries (success criteria)

### Acceptance Tests (Based on Spec Success Criteria)
- ✅ Agent records gap in < 30 seconds (CLI responsiveness)
- ✅ Query returns results in < 2 seconds for 1000+ entries
- ✅ Gap injection appears in plan prompts automatically
- ✅ Report generation works with 100+ gaps

### Test Coverage Targets
- Repository: 90%+ (critical path)
- Service: 85%+
- CLI: 80%+ (user-facing)

---

## Constitutional Compliance Validation

### Article I: Requirement-First Development ✅
- [x] CANARY tokens created for all features
- [x] Token progression: STUB → IMPL → TESTED (no BENCHED for CLI/Storage)
- [x] Token placement documented in plan

### Article IV: Test-First Imperative ✅
- [x] Phase 1: Tests written BEFORE migration implementation
- [x] Phase 2: Tests written BEFORE repository implementation
- [x] Phase 3: Tests written BEFORE service implementation
- [x] Phase 4: Tests written BEFORE CLI implementation
- [x] All phases follow Red → Green → Refactor cycle

### Article V: Simplicity and Anti-Abstraction ✅
- [x] Reuses existing SQLite database (no new database)
- [x] Reuses existing migration pattern (no custom schema manager)
- [x] Reuses Cobra CLI framework (no new CLI abstraction)
- [x] No premature optimization (indexes only for 2-second query target)

### Article VI: Integration-First Testing ✅
- [x] Tests use real SQLite database (not mocks)
- [x] Tests use actual file I/O for migration files
- [x] Integration tests cover full workflow
- [x] Performance tests use realistic data volumes (10,000 entries)

### Article VII: Documentation Currency ✅
- [x] CANARY tokens include UPDATED field
- [x] Documentation phase planned (Phase 7)
- [x] Agent context updated with gap workflow

---

## Performance Considerations

### Query Optimization
- **Indexes Created:**
  - `idx_gap_entries_req_id` - Fast lookup by requirement
  - `idx_gap_entries_aspect` - Fast filtering by aspect
  - `idx_gap_entries_category` - Fast filtering by category
  - `idx_gap_entries_created_at DESC` - Fast sorting by recency
  - `idx_gap_entries_helpful DESC` - Fast sorting by usefulness

- **Query Pattern:**
```sql
SELECT * FROM gap_entries
WHERE aspect = ? AND category_id IN (?)
ORDER BY helpful_count DESC, created_at DESC
LIMIT 10;
```
- **Expected Performance:** < 100ms for 10,000 entries (well under 2-second target)

### Database Sizing
- **Average Gap Entry:** ~500 bytes (text fields)
- **10,000 Entries:** ~5 MB storage
- **100,000 Entries:** ~50 MB storage
- **Acceptable:** SQLite handles hundreds of MB efficiently

---

## Security Considerations

### Sensitive Data in Gaps
- **Risk:** Security gaps may contain vulnerability details
- **Mitigation:**
  - Gap descriptions stored in local SQLite (not cloud)
  - Export/import feature uses file permissions (user-controlled)
  - No automatic sharing or telemetry
  - Documentation warns against including passwords/keys in gap descriptions

### SQL Injection
- **Mitigation:** All queries use prepared statements (`sqlx` placeholders)
- **Example:** `db.Get(&gap, "SELECT * FROM gap_entries WHERE gap_id = ?", gapID)`

---

## Migration Strategy

### For Existing CANARY Projects
1. Run `canary scan` to auto-migrate to version 4 (adds gap tables)
2. Gap tracking is optional - no impact on existing workflows
3. Export/import allows gradual team adoption

### For New CANARY Projects
- Migration 004 runs automatically on first `canary init`
- Gap categories pre-seeded with defaults
- Default configuration created

---

## Rollback Plan

### If Migration Fails
```bash
# Rollback migration 004
canary migrate down 1

# Verify database at version 3
canary status --db-version
```

### If Critical Bug Found
- Gap tracking is isolated from core scanning functionality
- Can disable gap injection by removing template directive
- Core CANARY operations (scan, verify, status) unaffected

---

## Success Metrics (Post-Implementation)

### Functionality Metrics
- [ ] All 4 CLI commands (mark, query, report, helpful) operational
- [ ] Database migration 004 applies successfully
- [ ] Gap injection appears in `/canary.plan` output
- [ ] Export/import works across projects

### Performance Metrics
- [ ] Gap marking completes in < 30 seconds (interactive prompt)
- [ ] Queries return in < 2 seconds for 1000+ entries
- [ ] Reports generate in < 5 seconds for 100+ entries
- [ ] Database supports 10,000 entries without degradation

### Quality Metrics
- [ ] Test coverage: Repository 90%+, Service 85%+, CLI 80%+
- [ ] Zero critical bugs in first week of usage
- [ ] Documentation complete and reviewed

---

## Next Steps

1. **Begin Phase 1:** Create database migration and repository tests (Red Phase)
2. **Constitutional Review:** Verify test-first approach in action
3. **Iterate:** Follow Red → Green → Refactor for each phase
4. **Track Progress:** Update CANARY tokens as features complete
5. **Final Validation:** Run full integration test suite before release

---

## Dependencies

- ✅ `internal/storage/db.go` - Existing migration framework
- ✅ `modernc.org/sqlite` - SQLite driver already in use
- ✅ `github.com/jmoiron/sqlx` - Query framework already in use
- ✅ `github.com/spf13/cobra` - CLI framework already in use
- ✅ `.canary/templates/commands/plan.md` - Template for injection

**No New Dependencies Required** ✅

---

## Appendix: Example Usage

### Recording a Gap
```bash
$ canary gap mark CBIN-042 OAuth --category security
Description: Missing redirect URI validation in OAuth flow - security vulnerability
Corrective Action: Added URI validation in auth handler at auth.go:145
✅ Gap GAP-042 created for CBIN-042/OAuth (category: security)
```

### Querying Gaps
```bash
$ canary gap query --aspect Security --category security
Found 3 gaps matching filters:

GAP-042 | CBIN-042/OAuth | security | 2025-10-15
  Missing redirect URI validation in OAuth flow
  Helpful: 5 | Unhelpful: 0

GAP-038 | CBIN-038/AuthN | security | 2025-10-12
  Missing rate limiting on login endpoint
  Helpful: 3 | Unhelpful: 1
...
```

### Rating a Gap
```bash
$ canary gap helpful GAP-042
✅ Marked GAP-042 as helpful (count: 6)
```

### Generating Report
```bash
$ canary gap report --summary
Gap Analysis Report
===================
Total Gaps: 127

By Category:
  Implementation: 45 (35%)
  Testing: 32 (25%)
  Security: 18 (14%)
  Performance: 17 (13%)
  Design: 10 (8%)
  Documentation: 5 (4%)

Top Patterns:
  1. Missing error handling (12 occurrences)
  2. Incomplete input validation (8 occurrences)
  3. Missing test coverage (8 occurrences)
```

---

**Plan Status:** ✅ Ready for Implementation
**Constitutional Gates:** ✅ All Passed
**Next Command:** Begin Phase 1 - Database Schema Implementation
