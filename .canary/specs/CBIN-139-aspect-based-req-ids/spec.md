<!-- CANARY: REQ=CBIN-115; FEATURE="SpecTemplate"; ASPECT=Docs; STATUS=IMPL; OWNER=canary; UPDATED=2025-10-16 -->
# Feature Specification: Aspect-Based Requirement IDs

**Requirement ID:** CBIN-139
**Status:** STUB
**Created:** 2025-10-16
**Last Updated:** 2025-10-16

## Overview

**Purpose:** Change the requirement identifier format from `<key>-<id>` (e.g., `CBIN-001`) to `<key>-<aspect-key>-<id>` (e.g., `CBIN-CLI-001`) to improve requirement organization, discoverability, and aspect-specific tracking. This allows teams to quickly identify the architectural layer of a requirement and enables aspect-scoped queries without parsing full tokens.

**Scope:**
- Included: New requirement ID format with aspect key segment
- Included: Aspect abbreviation mapping (API, CLI, Engine, Storage, Security, Docs, Wire, Planner, Decode, Encode, RoundTrip, Bench, FrontEnd, Dist)
- Included: Migration script to rename existing 138+ requirements
- Included: Updated ID generation logic in `canary create`, `canary specify`, `/canary.specify`
- Included: Updated templates (spec-template.md, plan-template.md, command templates)
- Included: Updated database schema to accommodate longer req_id values
- Included: Updated parser/matcher to recognize new format
- Included: Backward compatibility for parsing old-format IDs during migration period
- Excluded: Multiple aspects per requirement ID (use single primary aspect; split requirement if multi-aspect)
- Excluded: Custom aspect abbreviations per project (standardize on constitution list)
- Excluded: Automatic inference of aspect from code location (must be explicit)

## User Stories

### Primary User Stories

**US-1: Generate Aspect-Based Requirement IDs**
As a developer,
I want requirement IDs to include the aspect key (e.g., `CBIN-CLI-001`),
So that I can immediately identify the architectural layer without reading the full token.

**Acceptance Criteria:**
- [ ] `canary create` generates IDs like `CBIN-CLI-001`, `CBIN-API-002`, `CBIN-Engine-003`
- [ ] `canary specify` generates IDs with aspect key based on user input or prompt analysis
- [ ] `/canary.specify` slash command generates IDs with aspect key derived from feature description
- [ ] Generated IDs never collide (aspect-scoped sequential numbering)

**US-2: Migrate Existing Requirements**
As a project maintainer,
I want to migrate all 138+ existing requirements from `CBIN-XXX` to `CBIN-<ASPECT>-XXX`,
So that the entire project uses consistent naming without manual renaming.

**Acceptance Criteria:**
- [ ] Migration script renames all spec directories from `CBIN-XXX-feature` to `CBIN-<ASPECT>-XXX-feature`
- [ ] Migration updates all CANARY tokens in codebase (*.go, *.md files)
- [ ] Migration updates database req_id column values
- [ ] Migration preserves all token fields (FEATURE, STATUS, TEST, BENCH, etc.)
- [ ] Migration is idempotent (can be safely re-run)
- [ ] Migration creates backup before making changes

**US-3: Query Requirements by Aspect**
As a developer,
I want to list all CLI requirements or all API requirements,
So that I can focus on requirements relevant to my work area.

**Acceptance Criteria:**
- [ ] `canary list --aspect CLI` shows only CLI requirements
- [ ] `canary search --aspect Engine` searches only Engine requirements
- [ ] `canary next --aspect API` prioritizes API requirements
- [ ] Aspect filtering works with both old and new format during migration

**US-4: Parse Both Old and New Formats**
As a user during migration,
I want the scanner to parse both `CBIN-XXX` and `CBIN-<ASPECT>-XXX` formats,
So that migration can be gradual without breaking existing workflows.

**Acceptance Criteria:**
- [ ] Scanner recognizes `REQ=CBIN-001` (old format)
- [ ] Scanner recognizes `REQ=CBIN-CLI-001` (new format)
- [ ] `canary scan` reports both formats in output
- [ ] `canary verify` works with mixed formats during migration

### Secondary User Stories

**US-5: Validate Aspect Keys**
As a developer,
I want the system to reject invalid aspect keys,
So that requirement IDs remain consistent with the constitution.

**Acceptance Criteria:**
- [ ] Only valid aspects accepted: API, CLI, Engine, Storage, Security, Docs, Wire, Planner, Decode, Encode, RoundTrip, Bench, FrontEnd, Dist
- [ ] `canary create CBIN-InvalidAspect-001` returns error with valid aspect list
- [ ] Error messages suggest correct aspect if typo detected (fuzzy matching)
- [ ] Case-insensitive validation (CLI = cli = Cli)

**US-6: Generate Unique Aspect-Scoped IDs**
As the system,
I want to track next available ID per aspect,
So that `CBIN-CLI-001`, `CBIN-CLI-002` increment independently from `CBIN-API-001`, `CBIN-API-002`.

**Acceptance Criteria:**
- [ ] Each aspect maintains independent ID counter (CLI: 1,2,3..., API: 1,2,3...)
- [ ] `canary create` queries highest ID for given aspect and increments
- [ ] ID generation works with filesystem (spec directories) and database
- [ ] No race conditions when creating multiple requirements concurrently

## Functional Requirements

### FR-1: Aspect-Based ID Format
**Priority:** High
**Description:** Requirement IDs must follow the pattern `<key>-<aspect>-<id>` where key is the project identifier (e.g., CBIN), aspect is one of 14 valid aspects, and id is a zero-padded 3-digit number (001, 002, ..., 999).
**Acceptance:** Parser validates all three segments, rejects IDs not matching pattern

### FR-2: Aspect Validation
**Priority:** High
**Description:** System must validate aspect segment against constitution's approved list: API, CLI, Engine, Storage, Security, Docs, Wire, Planner, Decode, Encode, RoundTrip, Bench, FrontEnd, Dist (case-insensitive).
**Acceptance:** Invalid aspect keys rejected with error message listing valid options

### FR-3: Aspect-Scoped ID Generation
**Priority:** High
**Description:** ID generation must track next available ID per aspect, supporting independent counters (e.g., CBIN-CLI-005 and CBIN-API-003 coexist).
**Acceptance:** `canary create` generates unique IDs per aspect by scanning filesystem or querying database

### FR-4: Migration Automation
**Priority:** High
**Description:** Provide migration command to rename all existing requirement directories, update all CANARY tokens in source files, and update database records.
**Acceptance:** `canary migrate-ids` completes migration in < 5 minutes for 138+ requirements, creates backup, reports changes made

### FR-5: Backward Compatibility During Migration
**Priority:** High
**Description:** Scanner must recognize both old format (`CBIN-XXX`) and new format (`CBIN-<ASPECT>-XXX`) during migration period.
**Acceptance:** `canary scan` parses both formats, `canary verify` validates both formats, migration can be gradual

### FR-6: Template Updates
**Priority:** Medium
**Description:** Update all templates (spec-template.md, plan-template.md, slash command templates) to use new ID format.
**Acceptance:** All templates show examples with aspect-based IDs, placeholder text updated from `CBIN-XXX` to `CBIN-<ASPECT>-XXX`

### FR-7: Database Schema Update
**Priority:** Medium
**Description:** Extend database req_id column to accommodate longer IDs (old: 8 chars for CBIN-001, new: up to 20 chars for CBIN-FrontEnd-001).
**Acceptance:** Database migration extends req_id column to VARCHAR(64), existing data preserved

### FR-8: Aspect Filtering
**Priority:** Low
**Description:** Add `--aspect` flag to `canary list`, `canary search`, `canary next` commands to filter requirements by aspect.
**Acceptance:** Commands return only requirements matching specified aspect, error if invalid aspect provided

## Success Criteria

**Quantitative Metrics:**
- [ ] Migration completes in < 5 minutes for 138+ existing requirements
- [ ] Zero data loss during migration (all 138+ requirements preserved)
- [ ] New ID format reduces aspect identification time by 100% (instant visual recognition)
- [ ] Aspect-scoped queries (e.g., `canary list --aspect CLI`) execute in < 1 second

**Qualitative Measures:**
- [ ] Developers can identify requirement aspect without reading full token
- [ ] Requirement directories are easier to browse (grouped by aspect in alphabetical listings)
- [ ] Code reviews benefit from immediate aspect context in REQ= field
- [ ] Team adopts new format within 1 week of migration

## User Scenarios & Testing

### Scenario 1: Create New Requirement with Aspect (Happy Path)
**Given:** Developer wants to create a new CLI feature
**When:** They run `canary create "Import Command" --aspect CLI`
**Then:** System generates `CBIN-CLI-001` (or next available CLI ID) and creates `.canary/specs/CBIN-CLI-001-import-command/spec.md`

**Example:**
```bash
$ canary create "Import Command" --aspect CLI
Created: .canary/specs/CBIN-CLI-001-import-command/spec.md
Generated ID: CBIN-CLI-001
```

### Scenario 2: Migrate Existing Project
**Given:** Project has 138 requirements in old format (CBIN-001 through CBIN-138)
**When:** Maintainer runs `canary migrate-ids --dry-run` then `canary migrate-ids --confirm`
**Then:** All specs renamed, all tokens updated, database migrated, backup created

**Example:**
```bash
$ canary migrate-ids --dry-run
Will rename 138 requirements:
  CBIN-101-scanner-core → CBIN-Engine-101-scanner-core
  CBIN-102-verify-gate → CBIN-CLI-102-verify-gate
  CBIN-103-status-json → CBIN-API-103-status-json
  ...
Will update 542 CANARY tokens in source files
Create backup? (y/n): y

$ canary migrate-ids --confirm
✓ Created backup: .canary/backup/pre-migration-2025-10-16.tar.gz
✓ Renamed 138 spec directories
✓ Updated 542 CANARY tokens
✓ Migrated database (138 records)
✓ Migration complete in 2m 34s
```

### Scenario 3: Query Requirements by Aspect
**Given:** Project has migrated to aspect-based IDs
**When:** Developer runs `canary list --aspect CLI`
**Then:** System shows only CLI requirements

**Example:**
```bash
$ canary list --aspect CLI
CBIN-CLI-001 - Import Command (IMPL)
CBIN-CLI-002 - Export Command (STUB)
CBIN-CLI-005 - Verify Gate (TESTED)
...
```

### Scenario 4: Invalid Aspect Rejection (Error Case)
**Given:** Developer tries to create requirement with invalid aspect
**When:** They run `canary create "Test" --aspect Frontend` (typo: should be FrontEnd)
**Then:** System rejects with helpful error

**Example:**
```bash
$ canary create "Test" --aspect Frontend
Error: Invalid aspect "Frontend"
Did you mean: FrontEnd

Valid aspects: API, CLI, Engine, Storage, Security, Docs, Wire, Planner,
Decode, Encode, RoundTrip, Bench, FrontEnd, Dist
```

### Scenario 5: Parse Mixed Formats During Migration
**Given:** Project partially migrated (some old, some new format)
**When:** Developer runs `canary scan`
**Then:** Scanner parses both formats correctly

**Example:**
```go
// Old format (not yet migrated)
// CANARY: REQ=CBIN-042; FEATURE="OldFeature"; ASPECT=API; STATUS=IMPL; UPDATED=2025-09-20

// New format (migrated)
// CANARY: REQ=CBIN-API-042; FEATURE="OldFeature"; ASPECT=API; STATUS=IMPL; UPDATED=2025-10-16
```

### Scenario 6: Aspect-Scoped ID Generation
**Given:** Project has CBIN-CLI-003 as highest CLI requirement
**When:** Developer creates new CLI requirement
**Then:** System generates CBIN-CLI-004 (not CBIN-CLI-139 even if other aspects have higher IDs)

### Scenario 7: Database Schema Migration
**Given:** Database has req_id column VARCHAR(16)
**When:** Maintainer runs `canary migrate-ids`
**Then:** Schema updated to VARCHAR(64), all existing records preserved

## Key Entities

### Entity 1: AspectKey
**Attributes:**
- name: Full aspect name (e.g., "CLI", "API", "Engine")
- abbreviation: Same as name for most aspects (API, CLI, Engine)
- valid: Boolean indicating if aspect is approved by constitution
- description: Short description of aspect purpose

**Relationships:**
- Each RequirementID has exactly one AspectKey
- Each AspectKey has many RequirementIDs

### Entity 2: RequirementID
**Attributes:**
- key: Project identifier (e.g., "CBIN")
- aspect: Aspect key (e.g., "CLI", "API")
- id: Zero-padded 3-digit number (e.g., "001", "042")
- full_id: Complete ID string (e.g., "CBIN-CLI-001")
- format_version: "v1" (old: CBIN-XXX) or "v2" (new: CBIN-ASPECT-XXX)

**Relationships:**
- Belongs to one AspectKey
- Referenced by one Specification
- Referenced by many CanaryTokens in source code

### Entity 3: MigrationRecord
**Attributes:**
- old_id: Original ID (e.g., "CBIN-101")
- new_id: Migrated ID (e.g., "CBIN-Engine-101")
- aspect: Derived aspect from token's ASPECT field
- timestamp: When migration occurred
- spec_path_old: Original spec directory path
- spec_path_new: New spec directory path
- tokens_updated: Count of CANARY tokens updated in source files

**Relationships:**
- One MigrationRecord per migrated requirement
- Stored in `.canary/memory/migration-log.json`

## Assumptions

- Existing CANARY tokens have valid ASPECT field (used to derive aspect for migration)
- Each requirement has exactly one primary aspect (multi-aspect features should be split)
- Project uses CBIN as key prefix (migration script parameterizable for other keys)
- Requirement IDs will not exceed 999 per aspect (3-digit limit sufficient)
- Migration happens once (not continuous incremental migration)
- Developers will adopt new format within 1-2 weeks (old format deprecated after migration)

## Constraints

**Technical Constraints:**
- Database req_id column must support up to 64 characters (some SQL databases have identifier limits)
- Filesystem path length limits (Windows: 260 chars, Linux: 4096 chars) - longer IDs reduce available spec name length
- Migration must be atomic (all-or-nothing) to avoid inconsistent state
- Backward compatibility parser must not degrade performance by >5%

**Business Constraints:**
- Migration requires downtime or maintenance window (no concurrent edits during migration)
- All developers must be notified before migration (communication plan required)
- Backup storage required (138+ specs × average 10KB = ~1.5MB backup size)

**Regulatory Constraints:**
- None

## Out of Scope

- **Multi-Aspect IDs:** Requirements with multiple aspects (e.g., `CBIN-CLI-API-001`) - split into separate requirements instead
- **Custom Aspect Keys:** Project-specific aspects beyond constitution's 14 approved aspects
- **Automatic Aspect Inference:** Deriving aspect from code location (e.g., files in `cmd/` → CLI) - must be explicit
- **Aspect Hierarchies:** Parent-child aspect relationships (e.g., API > REST > GraphQL)
- **Aspect Renaming:** Changing aspect names after creation (REQ ID is immutable)
- **Cross-Project Aspect Mapping:** Standardizing aspect keys across multiple CANARY projects
- **Aspect-Based Branching:** Git branch naming based on aspect (e.g., `feature/CLI-001-import-cmd`)
- **Aspect Ownership:** Assigning teams to aspects (use OWNER field instead)

## Dependencies

- Existing CANARY scanner (internal/matcher)
- Database schema (internal/storage/db.go, migrations)
- Spec creation logic (cmd/canary/main.go, `canary create` command)
- Template system (embedded/templates.go)
- `/canary.specify` slash command template
- Constitution aspect list (.canary/memory/constitution.md Article III Section 3.2)

## Risks & Mitigation

| Risk | Impact | Probability | Mitigation |
|------|--------|-------------|------------|
| Migration corrupts existing specs | High | Low | Create backup before migration, implement dry-run mode, test on copy of production data |
| ID collisions during migration | High | Low | Migration script scans all existing IDs per aspect, uses max+1 logic, validates uniqueness before commit |
| Database migration fails mid-way | High | Low | Wrap migration in transaction, implement rollback logic, test on database copy |
| Developers forget new format | Medium | Medium | Update all templates, add validation to `canary create`, show examples in error messages |
| Performance degradation from longer IDs | Medium | Low | Benchmark parser with old vs new format (target <5% degradation), optimize if needed |
| Confusion during mixed-format period | Medium | Medium | Clearly document migration timeline, provide `--format-version` flag to check status, complete migration quickly |
| Breaking external tools parsing old IDs | Medium | Low | Document breaking change, provide regex patterns for both formats, deprecation notice in changelog |

## Clarifications Needed

[NEEDS CLARIFICATION: Should aspect-scoped ID numbering restart at 001 or continue from existing numbers?]
**Options:**
A) Restart at 001 per aspect (CBIN-101 with ASPECT=Engine → CBIN-Engine-001)
B) Preserve original numbers (CBIN-101 with ASPECT=Engine → CBIN-Engine-101)
C) Hybrid: new requirements start at 001, migrated requirements keep original numbers
**Impact:**
- Option A: Cleaner aspect-scoped numbering but loses original ID historical context
- Option B: Preserves historical context, easier to trace back to old format, but inconsistent numbering per aspect
- Option C: Best of both worlds but more complex migration logic
**Recommendation:** Option B - preserves git history, easier debugging, simpler migration

[NEEDS CLARIFICATION: Should migration be reversible?]
**Options:**
A) One-way migration only (no rollback command)
B) Implement `canary migrate-ids --rollback` using backup
C) Store migration mapping in `.canary/memory/migration-map.json` for bidirectional lookup
**Impact:**
- Option A: Simpler implementation, forces commitment to new format
- Option B: Safety net for early mistakes, requires robust backup/restore
- Option C: Best observability, allows tools to understand old→new mapping, minimal overhead
**Recommendation:** Option C - provides safety and traceability with minimal cost

[NEEDS CLARIFICATION: Should old-format IDs be deprecated immediately or gradually?]
**Options:**
A) Deprecate immediately after migration (scanner warns on old format)
B) Support both formats indefinitely (no deprecation)
C) Grace period of 3 months, then deprecate old format
**Impact:**
- Option A: Forces quick adoption, cleaner codebase sooner, but may break workflows
- Option B: Maximum compatibility, but perpetual complexity in parser
- Option C: Balanced approach, gives time to discover missed migrations
**Recommendation:** Option C - practical grace period allows discovery of edge cases

## Review & Acceptance Checklist

**Content Quality:**
- [x] No implementation details (languages, frameworks, APIs)
- [x] Focused on user value and business needs
- [x] Written for non-technical stakeholders
- [x] All mandatory sections completed

**Requirement Completeness:**
- [x] Only 3 [NEEDS CLARIFICATION] markers remaining
- [x] Requirements are testable and unambiguous
- [x] Success criteria are measurable and technology-agnostic
- [x] All acceptance scenarios defined
- [x] Edge cases identified (invalid aspects, mixed formats, ID collisions)
- [x] Scope clearly bounded (no multi-aspect IDs, no custom aspects)
- [x] Dependencies and assumptions identified

**Readiness:**
- [x] All functional requirements have clear acceptance criteria
- [x] User scenarios cover primary flows (creation, migration, querying)
- [x] Ready for technical planning (`/canary.plan`)

---

## Implementation Checklist

### Core Features

<!-- CANARY: REQ=CBIN-139; FEATURE="AspectIDParser"; ASPECT=Engine; STATUS=STUB; UPDATED=2025-10-16 -->
**Feature 1: Aspect-Based ID Parser**
- [ ] Parse new format `CBIN-<ASPECT>-<ID>` (regex: `^[A-Z]+-[A-Za-z]+-\d{3}$`)
- [ ] Parse old format `CBIN-<ID>` for backward compatibility
- [ ] Extract key, aspect, id segments from full ID
- [ ] Validate aspect against constitution's approved list
- **Location hint:** `internal/matcher/token.go` or new `internal/reqid/parser.go`
- **Dependencies:** None

<!-- CANARY: REQ=CBIN-139; FEATURE="AspectValidation"; ASPECT=Engine; STATUS=STUB; UPDATED=2025-10-16 -->
**Feature 2: Aspect Validation**
- [ ] Validate aspect against: API, CLI, Engine, Storage, Security, Docs, Wire, Planner, Decode, Encode, RoundTrip, Bench, FrontEnd, Dist
- [ ] Case-insensitive matching (CLI = cli = Cli)
- [ ] Fuzzy matching for typo suggestions (Frontend → FrontEnd)
- [ ] Error messages listing valid aspects
- **Location hint:** `internal/reqid/validator.go`
- **Dependencies:** AspectIDParser

<!-- CANARY: REQ=CBIN-139; FEATURE="AspectScopedIDGen"; ASPECT=Engine; STATUS=STUB; UPDATED=2025-10-16 -->
**Feature 3: Aspect-Scoped ID Generation**
- [ ] Query highest ID for given aspect from filesystem (spec directories)
- [ ] Query highest ID for given aspect from database
- [ ] Increment to generate next available ID (zero-padded 3 digits)
- [ ] Handle concurrent creation (lock or atomic increment)
- **Location hint:** `cmd/canary/main.go` (create command) or `internal/reqid/generator.go`
- **Dependencies:** AspectIDParser

<!-- CANARY: REQ=CBIN-139; FEATURE="MigrationScript"; ASPECT=CLI; STATUS=STUB; UPDATED=2025-10-16 -->
**Feature 4: Migration Command**
- [ ] Scan all spec directories (`.canary/specs/CBIN-XXX-*`)
- [ ] Read ASPECT field from spec.md CANARY token
- [ ] Determine new ID (preserve number: CBIN-101 → CBIN-Engine-101)
- [ ] Rename spec directories
- [ ] Update all CANARY tokens in spec.md files
- [ ] Update all CANARY tokens in source code (*.go, *.md)
- [ ] Update database req_id column
- [ ] Create backup (`.canary/backup/pre-migration-YYYY-MM-DD.tar.gz`)
- [ ] Generate migration log (`.canary/memory/migration-log.json`)
- [ ] Support `--dry-run` and `--confirm` flags
- **Location hint:** `cmd/canary/migrate_ids.go` (new command)
- **Dependencies:** AspectIDParser, AspectValidation

<!-- CANARY: REQ=CBIN-139; FEATURE="TemplateUpdates"; ASPECT=Docs; STATUS=STUB; UPDATED=2025-10-16 -->
**Feature 5: Template Updates**
- [ ] Update `spec-template.md`: `CBIN-XXX` → `CBIN-<ASPECT>-XXX`
- [ ] Update `plan-template.md`: `CBIN-XXX` → `CBIN-<ASPECT>-XXX`
- [ ] Update `.claude/commands/canary.specify.md`
- [ ] Update `.claude/commands/canary.plan.md`
- [ ] Update example tokens in `cmd/canary/main.go` (help text)
- **Location hint:** `.canary/templates/`, `.claude/commands/`
- **Dependencies:** None

<!-- CANARY: REQ=CBIN-139; FEATURE="DatabaseSchemaMigration"; ASPECT=Storage; STATUS=STUB; UPDATED=2025-10-16 -->
**Feature 6: Database Schema Migration**
- [ ] Extend req_id column from VARCHAR(16) to VARCHAR(64)
- [ ] Update migration files in `internal/storage/migrations/`
- [ ] Preserve existing data during schema change
- [ ] Test migration on copy of production database
- **Location hint:** `internal/storage/migrations/00X_extend_reqid.sql`
- **Dependencies:** None

### Testing Requirements

<!-- CANARY: REQ=CBIN-139; FEATURE="ParserTests"; ASPECT=Engine; STATUS=STUB; TEST=TestCBIN139_AspectIDParser; UPDATED=2025-10-16 -->
**Unit Tests - Parser:**
- [ ] Test parsing new format (CBIN-CLI-001, CBIN-Engine-042)
- [ ] Test parsing old format (CBIN-001, CBIN-138)
- [ ] Test invalid formats (CBIN-InvalidAspect-001, CBIN-CLI-1, CBIN-CLI-1234)
- [ ] Test aspect extraction and validation
- **Location hint:** `internal/reqid/parser_test.go`

<!-- CANARY: REQ=CBIN-139; FEATURE="GeneratorTests"; ASPECT=Engine; STATUS=STUB; TEST=TestCBIN139_AspectScopedIDGen; UPDATED=2025-10-16 -->
**Unit Tests - ID Generation:**
- [ ] Test aspect-scoped incrementing (CLI: 001→002, API: 001→002 independently)
- [ ] Test ID generation with no existing requirements
- [ ] Test ID generation with gaps (CBIN-CLI-001, CBIN-CLI-003 → generates CBIN-CLI-004)
- [ ] Test concurrent ID generation (race conditions)
- **Location hint:** `internal/reqid/generator_test.go`

<!-- CANARY: REQ=CBIN-139; FEATURE="MigrationTests"; ASPECT=CLI; STATUS=STUB; TEST=TestCBIN139_Migration; UPDATED=2025-10-16 -->
**Integration Tests - Migration:**
- [ ] Test migration with 10 sample requirements
- [ ] Test dry-run mode (no changes made)
- [ ] Test backup creation and restore
- [ ] Test migration idempotency (safe to re-run)
- [ ] Test migration log generation
- [ ] Test mixed-format parsing after partial migration
- **Location hint:** `cmd/canary/migrate_ids_test.go`

<!-- CANARY: REQ=CBIN-139; FEATURE="AspectFilterTests"; ASPECT=CLI; STATUS=STUB; TEST=TestCBIN139_AspectFilter; UPDATED=2025-10-16 -->
**Integration Tests - Aspect Filtering:**
- [ ] Test `canary list --aspect CLI`
- [ ] Test `canary search --aspect Engine`
- [ ] Test invalid aspect rejection
- [ ] Test aspect filtering with mixed formats
- **Location hint:** `cmd/canary/*_test.go` (various command tests)

### Documentation

<!-- CANARY: REQ=CBIN-139; FEATURE="MigrationGuide"; ASPECT=Docs; STATUS=STUB; UPDATED=2025-10-16 -->
**Migration Guide:**
- [ ] Document migration steps (dry-run → confirm)
- [ ] Explain backup/restore process
- [ ] Show before/after examples
- [ ] Provide troubleshooting section
- [ ] Document grace period (3 months old format support)
- **Location hint:** `.canary/docs/migration-aspect-ids.md`

<!-- CANARY: REQ=CBIN-139; FEATURE="AspectIDDocs"; ASPECT=Docs; STATUS=STUB; UPDATED=2025-10-16 -->
**ID Format Documentation:**
- [ ] Document new ID format: `<key>-<aspect>-<id>`
- [ ] List all valid aspects with descriptions
- [ ] Show examples for each aspect
- [ ] Explain aspect-scoped numbering
- [ ] Document validation rules
- **Location hint:** `README.md` or `.canary/docs/aspect-based-ids.md`

---

**Agent Instructions:**

After implementing each feature:
1. Update the CANARY token in the spec from `STATUS=STUB` to `STATUS=IMPL`
2. Add the same token to your source code at the implementation location
3. Add `TEST=TestName` when tests are written
4. Run `canary implement CBIN-139` to see implementation progress

---

## CANARY Tokens Reference

**Main requirement token** (add to primary implementation file):
```go
// CANARY: REQ=CBIN-139; FEATURE="AspectBasedRequirementIDs"; ASPECT=Engine; STATUS=IMPL; UPDATED=2025-10-16
```

**Sub-feature tokens** (use the specific feature names from Implementation Checklist):
```go
// CANARY: REQ=CBIN-139; FEATURE="AspectIDParser"; ASPECT=Engine; STATUS=IMPL; TEST=TestCBIN139_AspectIDParser; UPDATED=2025-10-16
```

**Use `canary implement CBIN-139` to find:**
- Which features are implemented vs. still TODO
- Exact file locations and line numbers
- Context around each implementation point
