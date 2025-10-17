# Feature Specification: Legacy Token Migration

**Requirement ID:** CBIN-145
**Aspect:** CLI
**Status:** STUB
**Created:** 2025-10-17
**Last Updated:** 2025-10-17

## Overview

**Purpose:** Enable CANARY CLI to automatically generate specifications and implementation plans for requirements that were migrated from legacy codebases, where CANARY tokens exist in code but formal specification files were never created. This prevents orphaned requirements from appearing in priority lists without actionable context.

**Scope:**
- Included: Auto-generation of spec.md and plan.md from existing CANARY tokens
- Included: Reconstruction of requirement context from token metadata (status, features, aspects)
- Included: Detection of legacy tokens (tokens without corresponding specs)
- Included: Migration command to batch-process orphaned requirements
- Excluded: Modification of existing specifications (only creates missing ones)
- Excluded: Code refactoring or token format changes
- Excluded: Automatic implementation of missing features

## User Stories

### Primary User Stories

**US-1: Detect Orphaned Requirements**
As a project maintainer,
I want to identify requirements that have CANARY tokens in code but no specification files,
So that I can understand which legacy requirements need documentation.

**Acceptance Criteria:**
- [ ] CLI can scan database and identify requirements with tokens but no spec directory
- [ ] Report shows requirement ID, token count, file locations, and status breakdown
- [ ] Distinguishes between documentation examples (in /docs/, /.claude/) and real code
- [ ] Report is actionable with clear next steps

**US-2: Auto-Generate Specifications from Tokens**
As a project maintainer,
I want to automatically generate specification files from existing CANARY tokens,
So that legacy requirements have proper documentation without manual reconstruction.

**Acceptance Criteria:**
- [ ] Command creates `.canary/specs/CBIN-XXX-feature/spec.md` for orphaned requirements
- [ ] Spec includes overview extracted from token metadata (features, aspects, statuses)
- [ ] Spec documents existing implementation state (what's TESTED vs STUB vs IMPL)
- [ ] Spec marks incomplete sections with [MIGRATED FROM LEGACY] placeholders
- [ ] Generated spec passes validation (required sections present)

**US-3: Auto-Generate Implementation Plans**
As a developer,
I want legacy requirements to have implementation plans showing current state,
So that I can understand what's already implemented and what's missing.

**Acceptance Criteria:**
- [ ] Command creates `.canary/specs/CBIN-XXX-feature/plan.md` from token analysis
- [ ] Plan lists all features found in tokens with their current STATUS
- [ ] Plan includes file locations and line numbers for implemented features
- [ ] Plan identifies missing features (STUB tokens) as TODOs
- [ ] Plan can be used with `canary implement` command immediately

**US-4: Batch Migration**
As a project maintainer,
I want to migrate all orphaned requirements in one command,
So that I can quickly establish proper documentation for legacy code.

**Acceptance Criteria:**
- [ ] Single command migrates all detected orphaned requirements
- [ ] Migration is idempotent (can run multiple times safely)
- [ ] Migration preserves existing specs (only creates missing ones)
- [ ] Migration generates summary report with statistics
- [ ] Migration can be previewed with dry-run mode

### Secondary User Stories

**US-5: Migration Quality Validation**
As a project maintainer,
I want migrated specifications to meet minimum quality standards,
So that they're useful for future development and not just placeholders.

**Acceptance Criteria:**
- [ ] Generated specs include all mandatory sections
- [ ] Generated specs reference actual code locations
- [ ] Generated specs document current implementation state accurately
- [ ] Migration report flags low-quality migrations for manual review

**US-6: Documentation Example Filtering**
As a developer,
I want migration to ignore documentation examples and test fixtures,
So that only real requirements are migrated.

**Acceptance Criteria:**
- [ ] Tokens in /docs/, /.claude/, /.cursor/ directories are excluded
- [ ] Tokens in `.canary/specs/` are excluded (already have specs)
- [ ] Test fixture tokens (CBIN-XXX in test files) can be optionally excluded
- [ ] Exclusion rules are configurable via flags

## Functional Requirements

### FR-1: Orphan Detection
**Priority:** High
**Description:** System must identify requirements with CANARY tokens but no specification directory
**Acceptance:** Query database for distinct req_id values, check for `.canary/specs/CBIN-XXX-*` existence, report mismatches

### FR-2: Specification Generation
**Priority:** High
**Description:** System must generate valid spec.md files from token metadata
**Acceptance:** Generated spec.md contains: requirement overview, feature list with statuses, file locations, implementation checklist with current progress

### FR-3: Plan Generation
**Priority:** High
**Description:** System must generate plan.md files reflecting current implementation state
**Acceptance:** Generated plan.md includes: tech stack summary, implementation phases mapping to existing statuses, file structure with actual paths, token placement examples from real code

### FR-4: Idempotent Migration
**Priority:** High
**Description:** Migration command must safely run multiple times without corrupting existing specifications
**Acceptance:** Command skips requirements where `.canary/specs/CBIN-XXX-*` already exists, dry-run mode shows what would be created without modifications, error handling prevents partial migrations

### FR-5: Batch Processing
**Priority:** Medium
**Description:** System must process multiple orphaned requirements efficiently in single command
**Acceptance:** Migration processes all orphaned requirements in under 30 seconds for typical projects (<1000 tokens), progress indicator shows current requirement being processed, summary statistics reported at completion

### FR-6: Quality Heuristics
**Priority:** Medium
**Description:** Generated specifications must meet minimum quality thresholds for usefulness
**Acceptance:** Spec includes at least 3 features with distinct aspects, Plan references at least 2 actual code files, Spec has meaningful overview text (not just "[MIGRATED]" placeholder)

### FR-7: Path Filtering
**Priority:** Medium
**Description:** Migration must exclude documentation examples and test fixtures
**Acceptance:** Default exclusion patterns: /docs/, /.claude/, /.cursor/, .canary/specs/, Custom exclusion patterns configurable via --exclude flag, Inclusion override available via --include-all flag

## Success Criteria

**Quantitative Metrics:**
- [ ] Migration completes in < 30 seconds for projects with 100 orphaned requirements
- [ ] 90% of generated specifications pass validation checks
- [ ] Generated plans include accurate file locations for 95% of IMPL/TESTED tokens
- [ ] Dry-run preview matches actual migration results 100% of time
- [ ] Migration reduces "orphaned requirement" warnings by 100%

**Qualitative Measures:**
- [ ] Developers can run `canary list` and see actionable context for all requirements
- [ ] Generated specifications provide enough context to continue development
- [ ] Legacy requirements integrate seamlessly with normal workflow (specify/plan/implement)
- [ ] Migration requires zero manual file editing for basic cases
- [ ] Generated plans are immediately usable with `canary implement CBIN-XXX`

## User Scenarios & Testing

### Scenario 1: Detect Single Orphaned Requirement (Happy Path)
**Given:** CBIN-105 has 20 tokens in codebase but no `.canary/specs/CBIN-105-*` directory
**When:** Developer runs `canary migrate --detect`
**Then:** Report shows:
```
Orphaned Requirements: 1

CBIN-105:
  Tokens: 20 (2 STUB, 10 IMPL, 8 TESTED)
  Features: Search, FuzzySearch, UserAuth
  Aspects: API (15), Engine (5)
  Files: internal/search/search.go, internal/auth/auth.go

Recommendation: Run 'canary migrate CBIN-105' to generate spec
```

### Scenario 2: Auto-Generate Specification
**Given:** CBIN-105 detected as orphaned with tokens in database
**When:** Developer runs `canary migrate CBIN-105`
**Then:**
- `.canary/specs/CBIN-105-search/spec.md` created
- Spec includes: "## Features (Migrated from Legacy)\n\n- Search (API, TESTED)\n- FuzzySearch (Engine, TESTED)\n- UserAuth (API, IMPL)"
- Spec references actual files: "Implementation locations: internal/search/search.go:45, internal/auth/auth.go:12"
- Spec includes CANARY tokens in Implementation Checklist matching current statuses

### Scenario 3: Batch Migration
**Given:** 5 orphaned requirements detected (CBIN-105, CBIN-107, CBIN-110, CBIN-115, CBIN-120)
**When:** Developer runs `canary migrate --all`
**Then:**
- 5 spec directories created
- Summary report shows: "Migrated 5 requirements (45 tokens total, 12 STUB, 20 IMPL, 13 TESTED)"
- Each spec contains accurate feature lists and file references
- Migration completes in < 10 seconds

### Scenario 4: Exclude Documentation Examples
**Given:** CBIN-105 has tokens in both real code and `/docs/user/getting-started.md`
**When:** Developer runs `canary migrate CBIN-105`
**Then:**
- Generated spec only references tokens from real code files
- Tokens from /docs/, /.claude/, /.cursor/ directories excluded
- Report shows: "Excluded 8 documentation example tokens"

### Scenario 5: Idempotent Dry-Run
**Given:** CBIN-105 already has a spec at `.canary/specs/CBIN-105-search/spec.md`
**When:** Developer runs `canary migrate --all --dry-run`
**Then:**
- CBIN-105 skipped with message: "Already has spec, skipping"
- Other orphaned requirements shown as "Would create: CBIN-107, CBIN-110"
- No files actually modified

### Scenario 6: Low-Quality Migration Warning
**Given:** CBIN-999 has only 2 tokens, both in test files
**When:** Developer runs `canary migrate CBIN-999`
**Then:**
- Spec created but flagged with warning: "⚠️ Low confidence migration: Only 2 tokens found, all in test files"
- Spec includes [NEEDS MANUAL REVIEW] marker in overview
- Developer advised to manually verify or delete spec

## Key Entities

### Entity 1: OrphanedRequirement
**Attributes:**
- req_id: Requirement identifier (CBIN-XXX)
- token_count: Number of tokens found for this requirement
- features: List of feature names extracted from tokens
- aspects: List of aspects (API, CLI, Engine, etc.)
- status_breakdown: Count of tokens by status (STUB, IMPL, TESTED)
- file_paths: List of files containing tokens
- is_documentation: Boolean indicating if tokens are in docs directories

**Relationships:**
- Has many Token entities
- Will generate one Specification entity

### Entity 2: MigrationPlan
**Attributes:**
- target_requirements: List of OrphanedRequirement to migrate
- exclusion_patterns: Path patterns to ignore
- dry_run: Boolean for preview mode
- created_specs: List of generated specification paths
- skipped_requirements: List of requirements with existing specs

**Relationships:**
- References multiple OrphanedRequirement entities
- Produces MigrationReport entity

### Entity 3: MigrationReport
**Attributes:**
- total_orphaned: Count of orphaned requirements found
- migrated_count: Count of specs created
- skipped_count: Count of requirements with existing specs
- excluded_token_count: Count of tokens excluded (docs/examples)
- quality_warnings: List of low-confidence migrations
- execution_time: Duration of migration

**Relationships:**
- Summarizes MigrationPlan results
- References all OrphanedRequirement entities processed

## Assumptions

- Existing CANARY tokens in code are syntactically valid and parsable
- Token metadata (REQ, FEATURE, ASPECT, STATUS) accurately reflects implementation state
- File paths in database are relative to project root and still valid
- Developers prefer auto-generated specs over manual creation for legacy code
- Legacy code has at least 2-3 features per requirement for meaningful migration
- Specification template structure is stable and compatible with generated content

## Constraints

**Technical Constraints:**
- Migration must not modify existing specifications or code files
- Generated specifications must pass same validation as manual specifications
- File system operations must be atomic (no partial spec creation)
- Database queries must complete in < 5 seconds for large projects (10,000+ tokens)

**Business Constraints:**
- Migration feature must be backwards-compatible with existing workflows
- Generated specs should be indistinguishable from manual specs where possible
- No external dependencies beyond existing CANARY CLI infrastructure

**Regulatory Constraints:**
- None

## Out of Scope

- Automatic generation of test cases for STUB features
- Code refactoring or token format standardization
- Migration from other requirement tracking systems (Jira, GitHub Issues, etc.)
- Natural language processing to infer requirements from code comments
- Version control integration (git commit generation)
- Specification quality scoring beyond basic heuristics
- Interactive migration wizard (use CLI flags instead)

## Dependencies

- CBIN-123: TokenStorage (database must have indexed tokens)
- CBIN-135: ListCmd (uses similar token querying logic)
- CBIN-121: PlanCmd (generated plans follow same template structure)
- CBIN-120: SpecifyCmd (uses specification template)
- File system access to `.canary/specs/` directory

## Risks & Mitigation

| Risk | Impact | Probability | Mitigation |
|------|--------|-------------|------------|
| Generated specs have low quality/usefulness | Medium | High | Implement quality heuristics, flag low-confidence migrations, provide manual review checklist |
| Migration overwrites manually edited specs | High | Low | Check for existing spec directory before creation, require --force flag to overwrite, backup existing specs |
| Documentation tokens pollute migration | Low | Medium | Default exclusion patterns for /docs/, /.claude/, test files, configurable via flags |
| Large projects timeout during migration | Medium | Low | Optimize database queries, process in batches, show progress indicator, allow partial migration |
| Token metadata is inaccurate or stale | Medium | Medium | Include migration timestamp, mark as [MIGRATED FROM LEGACY], encourage manual review, provide update path |

## Clarifications Needed

[NEEDS CLARIFICATION: Should migration update existing specs if they contain [MIGRATED FROM LEGACY] markers?]
**Options:**
A) Yes - Re-run migration to refresh legacy specs with latest token data
B) No - Migration only creates new specs, never modifies existing
C) Optional - Provide --refresh flag to update legacy-migrated specs only

**Impact:** Option A enables continuous sync with code evolution but risks overwriting manual edits. Option B is safest but requires manual spec updates. Option C provides flexibility.

[NEEDS CLARIFICATION: How should migration handle requirements with only test tokens?]
**Options:**
A) Create spec anyway with warning (test fixtures may indicate feature exists)
B) Skip entirely (test tokens are not real requirements)
C) Require --include-tests flag to process

**Impact:** Option A creates noise but captures edge cases. Option B misses some valid requirements. Option C provides control.

[NEEDS CLARIFICATION: Should generated specs include AI-generated text or just structured data?]
**Options:**
A) Generate full natural language descriptions using token context
B) Use minimal placeholders ([MIGRATED: Add description here])
C) Hybrid - Generate basic overview, leave details as placeholders

**Impact:** Option A provides richer specs but may be inaccurate. Option B requires more manual work. Option C balances automation and accuracy.

## Review & Acceptance Checklist

**Content Quality:**
- [x] No implementation details (languages, frameworks, APIs)
- [x] Focused on user value and business needs
- [x] Written for non-technical stakeholders
- [x] All mandatory sections completed

**Requirement Completeness:**
- [x] Only 3 [NEEDS CLARIFICATION] markers (migration update policy, test token handling, spec content richness)
- [x] Requirements are testable and unambiguous
- [x] Success criteria are measurable and technology-agnostic
- [x] All acceptance scenarios defined
- [x] Edge cases identified (low-quality migrations, documentation tokens, existing specs)
- [x] Scope clearly bounded (no code refactoring, no other system imports)
- [x] Dependencies and assumptions identified

**Readiness:**
- [x] All functional requirements have clear acceptance criteria
- [x] User scenarios cover primary flows (detect, generate, batch, filter, idempotent)
- [x] Ready for technical planning (`/canary.plan`)

---

## Implementation Checklist

### Core Features

<!-- CANARY: REQ=CBIN-145; FEATURE="OrphanDetection"; ASPECT=Engine; STATUS=TESTED; TEST=TestCANARY_CBIN_145_Engine_PathFilter; UPDATED=2025-10-17 -->
**Feature 1: Orphan Detection Engine**
- [x] Query database for distinct req_id values
- [x] Check filesystem for spec directory existence
- [x] Filter out documentation examples (configurable patterns)
- [x] Calculate status breakdown per requirement
- **Location:** `internal/migrate/orphan.go`
- **Tests:** `internal/migrate/orphan_test.go`

<!-- CANARY: REQ=CBIN-145; FEATURE="SpecGeneration"; ASPECT=Engine; STATUS=TESTED; TEST=TestCANARY_CBIN_145_Engine_SpecTemplate; UPDATED=2025-10-17 -->
**Feature 2: Specification Generator**
- [x] Extract feature list from tokens
- [x] Generate overview text from token metadata
- [x] Create implementation checklist with current statuses
- [x] Populate file locations from database
- **Location:** `internal/migrate/spec_generator.go`
- **Tests:** `internal/migrate/spec_generator_test.go`

<!-- CANARY: REQ=CBIN-145; FEATURE="PlanGeneration"; ASPECT=Engine; STATUS=TESTED; TEST=TestCANARY_CBIN_145_Engine_PlanTemplate; UPDATED=2025-10-17 -->
**Feature 3: Plan Generator**
- [x] Map token statuses to implementation phases
- [x] Extract tech stack from file extensions
- [x] Generate file structure from actual paths
- [x] Create token placement examples from real tokens
- **Location:** `internal/migrate/plan_generator.go`
- **Tests:** `internal/migrate/plan_generator_test.go`

<!-- CANARY: REQ=CBIN-145; FEATURE="MigrateCommand"; ASPECT=CLI; STATUS=IMPL; UPDATED=2025-10-17 -->
**Feature 4: Migration CLI Command**
- [x] Add `canary orphan` command with subcommands (detect, run)
- [x] Implement dry-run mode
- [x] Add exclusion pattern flags
- [x] Generate migration report
- **Location:** `cmd/canary/migrate.go`
- **Tests:** `cmd/canary/migrate_test.go`, `cmd/canary/migrate_integration_test.go`

### Data Layer

<!-- CANARY: REQ=CBIN-145; FEATURE="MigrationQueries"; ASPECT=Storage; STATUS=IMPL; UPDATED=2025-10-17 -->
**Migration Queries:**
- [x] Query for orphaned requirements (using existing db.ListTokens)
- [x] Aggregate token metadata by req_id
- [x] Filter by path patterns
- **Location:** `internal/migrate/orphan.go` (uses existing storage APIs)

### Testing Requirements

<!-- CANARY: REQ=CBIN-145; FEATURE="MigrationUnitTests"; ASPECT=CLI; STATUS=TESTED; TEST=TestCANARY_CBIN_145_CLI_OrphanDetection; UPDATED=2025-10-17 -->
**Unit Tests:**
- [x] Test orphan detection logic
- [x] Test spec generation with various token sets
- [x] Test exclusion pattern matching
- [x] Test idempotency (existing specs not overwritten)
- **Location:** `cmd/canary/migrate_test.go`, `internal/migrate/*_test.go`

<!-- CANARY: REQ=CBIN-145; FEATURE="MigrationIntegrationTests"; ASPECT=CLI; STATUS=IMPL; TEST=TestCANARY_CBIN_145_CLI_EndToEnd; UPDATED=2025-10-17 -->
**Integration Tests:**
- [x] Test full migration workflow with test database
- [x] Test batch migration of multiple requirements
- [x] Verify generated specs are valid and parseable
- [ ] Verify generated plans work with `canary implement` (pending manual verification)
- **Location:** `cmd/canary/migrate_integration_test.go`

### Documentation

<!-- CANARY: REQ=CBIN-145; FEATURE="MigrationDocs"; ASPECT=Docs; STATUS=IMPL; UPDATED=2025-10-17 -->
**Migration Command Documentation:**
- [ ] Update README with migration workflow (pending)
- [ ] Create docs/migration-guide.md (pending)
- [x] Add slash command: `.claude/commands/canary.migrate.md`
- [x] Document common migration patterns (in slash command)
- **Location:** `.claude/commands/canary.migrate.md`

---

**Agent Instructions:**

After implementing each feature:
1. Update the CANARY token in the spec from `STATUS=STUB` to `STATUS=IMPL`
2. Add the same token to your source code at the implementation location
3. Add `TEST=TestName` when tests are written
4. Run `canary implement CBIN-145` to see implementation progress

---

## CANARY Tokens Reference

**Main requirement token** (add to primary implementation file):
```go
// CANARY: REQ=CBIN-145; FEATURE="LegacyTokenMigration"; ASPECT=CLI; STATUS=IMPL; UPDATED=2025-10-17
```

**Sub-feature tokens** (use the specific feature names from Implementation Checklist):
```go
// CANARY: REQ=CBIN-145; FEATURE="OrphanDetection"; ASPECT=CLI; STATUS=IMPL; TEST=TestCANARY_CBIN_145_CLI_OrphanDetection; UPDATED=2025-10-17
```

**Use `canary implement CBIN-145` to find:**
- Which features are implemented vs. still TODO
- Exact file locations and line numbers
- Context around each implementation point
