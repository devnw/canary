<!-- CANARY: REQ=CBIN-115; FEATURE="SpecTemplate"; ASPECT=Docs; STATUS=IMPL; OWNER=canary; UPDATED=2025-10-16 -->
# Feature Specification: Requirement History Tracking

**Requirement ID:** CBIN-137
**Status:** STUB
**Created:** 2025-10-16
**Last Updated:** 2025-10-16

## Overview

**Purpose:** Track changes to requirements and specifications over time, enabling developers and AI agents to understand how requirements evolved, retrieve historical versions for context, and maintain an audit trail of requirement modifications.

**Scope:**
- Included: Version tracking for CANARY tokens and specification files
- Included: Historical version retrieval for context loading
- Included: Change auditing (what changed, when, why)
- Included: Version comparison and diff capabilities
- Included: Database storage of historical versions
- Excluded: Full Git integration (use CANARY-specific versioning)
- Excluded: Branching or merging of requirement versions
- Excluded: Real-time collaborative editing
- Excluded: Visual timeline or graphical history views

## User Stories

### Primary User Stories

**US-1: Track Requirement Changes**
As a developer,
I want to automatically version CANARY tokens and specifications when they change,
So that I can maintain an audit trail of requirement evolution.

**Acceptance Criteria:**
- [ ] Each CANARY token modification creates a new version record
- [ ] Version records capture: timestamp, author, change description, old values, new values
- [ ] Specification file changes are versioned when STATUS or critical fields change
- [ ] Version numbers increment automatically (e.g., v1, v2, v3)
- [ ] Can query version history for any requirement ID

**US-2: Retrieve Historical Versions**
As an AI agent,
I want to load previous versions of requirements into context,
So that I can understand the evolution of a feature and make informed decisions.

**Acceptance Criteria:**
- [ ] Can retrieve specific version: `canary history CBIN-137 --version 3`
- [ ] Can retrieve version by date: `canary history CBIN-137 --date 2025-09-15`
- [ ] Can retrieve all versions: `canary history CBIN-137 --all`
- [ ] Output includes full CANARY token state and specification content snapshot
- [ ] Response time < 100ms for single version, < 500ms for full history

**US-3: Compare Requirement Versions**
As a developer,
I want to see what changed between two versions of a requirement,
So that I can understand the impact of requirement modifications.

**Acceptance Criteria:**
- [ ] Can compare adjacent versions: `canary history CBIN-137 --diff v2..v3`
- [ ] Can compare any two versions: `canary history CBIN-137 --diff v1..v5`
- [ ] Diff shows field-level changes (STATUS: STUB → IMPL)
- [ ] Diff highlights specification content changes
- [ ] Output is readable in terminal (colored diff format)

**US-4: Annotate Version Changes**
As a developer,
I want to add notes explaining why a requirement changed,
So that future developers understand the reasoning behind modifications.

**Acceptance Criteria:**
- [ ] Can add change notes: `canary history CBIN-137 --annotate "Changed scope to exclude X"`
- [ ] Notes are stored with version record
- [ ] Notes appear in history output
- [ ] Can edit notes after creation

### Secondary User Stories

**US-5: Automated Version Creation**
As a system,
I want to automatically create versions when scanning detects changes,
So that no manual version management is required.

**Acceptance Criteria:**
- [ ] `canary scan` detects CANARY token changes since last version
- [ ] New version created automatically with change summary
- [ ] Version number increments from last known version
- [ ] No duplicate versions created for identical content

**US-6: Version Rollback Guidance**
As a developer,
I want to see how to restore a previous requirement version,
So that I can revert problematic changes.

**Acceptance Criteria:**
- [ ] History output shows rollback commands
- [ ] Can generate diff to restore old version
- [ ] Rollback guidance includes affected files and tokens
- [ ] Warning shown for destructive rollback operations

## Functional Requirements

### FR-1: Version Field in CANARY Tokens
**Priority:** High
**Description:** Add optional VERSION= field to CANARY tokens to track requirement versions
**Acceptance:** Token parser recognizes VERSION= field, stores in database, displays in reports

### FR-2: Version History Database Schema
**Priority:** High
**Description:** Extend database with version_history table storing snapshots of requirement states
**Acceptance:** Schema includes: version_id, req_id, version_number, created_at, author, change_summary, token_snapshot (JSON), spec_snapshot (text), parent_version_id

### FR-3: Automatic Version Creation
**Priority:** High
**Description:** Scan command detects changes to CANARY tokens and creates new version records
**Acceptance:** New version created when: STATUS changes, FEATURE renamed, ASPECT changed, OWNER changed, or specification file modified

### FR-4: Version Retrieval Command
**Priority:** High
**Description:** CLI command `canary history <REQ-ID>` to retrieve version history
**Acceptance:** Command displays version list, allows filtering by date/version, outputs structured format

### FR-5: Version Comparison
**Priority:** Medium
**Description:** Diff functionality to compare two requirement versions
**Acceptance:** Shows field changes, specification content diff, timestamp delta, author changes

### FR-6: Change Annotation
**Priority:** Medium
**Description:** Allow adding notes to version records explaining change rationale
**Acceptance:** Notes stored in database, displayed in history output, editable via CLI

### FR-7: Agent Context Loading
**Priority:** High
**Description:** Agent slash command `/canary.history` to load historical versions into context
**Acceptance:** Agents can retrieve specific versions, compare versions, see change timeline

### FR-8: Version Pruning
**Priority:** Low
**Description:** Configurable retention policy for old versions (e.g., keep last 50 versions)
**Acceptance:** Pruning command removes old versions, preserves major milestones (STUB→IMPL→TESTED)

## Success Criteria

**Quantitative Metrics:**
- [ ] Version retrieval completes in < 100ms for single version
- [ ] Full history query (50+ versions) completes in < 500ms
- [ ] Version diff output is < 2000 tokens for agent context
- [ ] Automatic versioning adds < 50ms to scan operations
- [ ] Version history storage is < 100KB per requirement (50 versions)

**Qualitative Measures:**
- [ ] Developers can understand requirement evolution in < 1 minute
- [ ] Agents successfully use historical context to inform decisions
- [ ] Version diffs clearly highlight critical changes (status, scope)
- [ ] Change annotations provide meaningful context for future readers
- [ ] Version history is reliable and never loses data

## User Scenarios & Testing

### Scenario 1: Automatic Versioning During Scan (Happy Path)
**Given:** CBIN-137 exists with STATUS=STUB (current version v1)
**When:** Developer changes STATUS=IMPL and runs `canary scan`
**Then:** System creates version v2 with change summary "STATUS: STUB → IMPL", stores token snapshot, preserves v1 in history

### Scenario 2: Retrieve Specific Historical Version
**Given:** CBIN-137 has 5 versions in history
**When:** Developer runs `canary history CBIN-137 --version 3`
**Then:** System displays v3 token snapshot, specification content, creation timestamp, change summary

### Scenario 3: Compare Two Versions
**Given:** CBIN-137 has versions v1 (STATUS=STUB) and v3 (STATUS=IMPL, FEATURE renamed)
**When:** Developer runs `canary history CBIN-137 --diff v1..v3`
**Then:** System shows field changes: STATUS (STUB → IMPL), FEATURE (old → new), specification diff highlighting scope changes

### Scenario 4: Agent Loads Historical Context
**Given:** AI agent is implementing CBIN-137 and needs to understand past decisions
**When:** Agent runs `/canary.history CBIN-137 --all`
**Then:** System returns version timeline with change summaries, agent uses context to avoid repeating past mistakes

### Scenario 5: Annotate Version with Rationale
**Given:** Developer just created v4 by changing ASPECT from API to CLI
**When:** Developer runs `canary history CBIN-137 --annotate "Moved to CLI aspect due to user feedback"`
**Then:** System stores note with v4, displays note in future history queries

### Scenario 6: Retrieve Version by Date
**Given:** CBIN-137 has 10 versions created over 3 months
**When:** Developer runs `canary history CBIN-137 --date 2025-09-15`
**Then:** System returns version closest to date (before or on 2025-09-15), shows timestamp and version number

### Scenario 7: No Changes Detected (Duplicate Prevention)
**Given:** CBIN-137 is at v5 with STATUS=IMPL
**When:** Developer runs `canary scan` twice without changes
**Then:** System detects no changes, does not create v6 or v7, version remains v5

### Scenario 8: Version Rollback Guidance
**Given:** CBIN-137 v6 introduced errors, developer wants to revert to v4
**When:** Developer runs `canary history CBIN-137 --rollback v4`
**Then:** System displays: files to modify, token changes to apply, specification content to restore, warning about losing v5 and v6 changes

## Key Entities

### Entity 1: VersionHistory
**Attributes:**
- version_id: Primary key (UUID or auto-increment)
- req_id: Requirement ID (CBIN-XXX)
- version_number: Integer version (1, 2, 3, ...)
- created_at: Timestamp of version creation
- author: Username or email of modifier (if available)
- change_summary: Brief description of what changed
- token_snapshot: JSON object storing all CANARY token fields at this version
- spec_snapshot: Full text of specification file at this version (optional, if spec changed)
- parent_version_id: Link to previous version (for history chain)

**Relationships:**
- One requirement (req_id) has many versions
- Each version has one parent version (except v1)
- Versions form a linear history chain

### Entity 2: VersionAnnotation
**Attributes:**
- annotation_id: Primary key
- version_id: Foreign key to VersionHistory
- created_at: Timestamp of annotation
- author: Who added the note
- note: Text explanation of change rationale

**Relationships:**
- One version can have multiple annotations
- Annotations are ordered by created_at

### Entity 3: VersionComparison
**Attributes:**
- version_a_id: First version in comparison
- version_b_id: Second version in comparison
- field_changes: JSON object mapping field names to (old_value, new_value) pairs
- spec_diff: Text diff of specification content (if both versions have specs)
- timestamp_delta: Time elapsed between versions

**Relationships:**
- Derived entity (computed on demand, not stored)
- References two VersionHistory records

## Assumptions

- CANARY tokens are uniquely identified by REQ= field (CBIN-XXX)
- Specification files in `.canary/specs/CBIN-XXX-feature/spec.md` map to tokens
- Version history is stored in `.canary/canary.db` (SQLite)
- Users may or may not have Git configured (CANARY versioning is independent)
- Version numbers are monotonically increasing integers (no semantic versioning)
- Authors can be inferred from Git commits or environment variables (optional)
- Specification snapshots may be large (10-50 KB per version), so storage is selectively enabled

## Constraints

**Technical Constraints:**
- SQLite database maximum size (~140 TB, but prefer < 100 MB for portability)
- Version snapshots stored as JSON or TEXT, limited by database field size
- Must work without Git (pure CANARY-based versioning)
- Version diffs must be deterministic (same inputs = same output)
- No network calls (all operations are local)

**Business Constraints:**
- Minimal performance impact on existing scan operations (< 10% overhead)
- Version history should not significantly increase database size (< 2x growth)
- CLI commands should be consistent with existing `canary` command patterns
- No breaking changes to existing CANARY token format

**Regulatory Constraints:**
- None

## Out of Scope

- **Git Integration:** CANARY versioning is independent of Git history (though Git metadata may be used if available)
- **Branching/Merging:** Only linear version history (no parallel version branches)
- **Real-time Collaboration:** No conflict resolution or multi-user editing
- **Visual Timeline:** No web UI or graphical history visualization
- **Cross-Requirement Versioning:** Only tracks individual requirements, not project-wide snapshots
- **Semantic Versioning:** Simple integer versioning (v1, v2, v3), not semver (v1.2.3)
- **Automated Rollback:** Provides guidance only, does not automatically modify files
- **Version Tagging:** No named versions or milestone markers (e.g., "beta", "release")

## Dependencies

- CBIN-123: TokenStorage (database schema for tokens)
- Existing `canary scan` command (for change detection)
- Specification file structure (`.canary/specs/CBIN-XXX-feature/spec.md`)

## Risks & Mitigation

| Risk | Impact | Probability | Mitigation |
|------|--------|-------------|------------|
| Version history grows too large (database bloat) | Medium | Medium | Implement retention policy (keep last 50 versions), compress spec snapshots, prune old versions periodically |
| Change detection misses subtle modifications | Medium | Low | Use content hashing for deterministic change detection, include UPDATED field in comparison |
| Version diff output is too verbose for agents | Low | Medium | Provide summary mode showing only critical field changes (STATUS, FEATURE, ASPECT) |
| Users forget to annotate important changes | Low | High | Auto-generate change summaries from field diffs, make annotation optional |
| Version retrieval is slow with many versions | Medium | Low | Index version_history table by (req_id, version_number), cache recent queries |

## Clarifications Needed

[NEEDS CLARIFICATION: Should specification snapshots be stored for every version or only when spec.md content changes?]
**Options:**
A) Store spec snapshot for every version (comprehensive but storage-heavy)
B) Store spec snapshot only when spec.md file content changes (storage-efficient but requires change detection)
C) Make it configurable via `--snapshot-specs` flag
**Impact:** Option B balances storage efficiency with completeness, reduces database size by 60-80%

[NEEDS CLARIFICATION: How should authors be determined when creating versions?]
**Options:**
A) Use Git author from last commit touching the requirement
B) Use environment variable (e.g., $CANARY_AUTHOR or $USER)
C) Prompt user to configure author in `.canary/config.yaml`
D) Leave author field empty (system-generated versions)
**Impact:** Option B is simplest and works without Git, option A provides best accuracy when Git is available

[NEEDS CLARIFICATION: Should VERSION= field be required in CANARY tokens or optional?]
**Options:**
A) Required - all tokens must have VERSION= field
B) Optional - only added if history tracking is enabled
C) Auto-generated - scanner adds VERSION= automatically, tokens don't need it
**Impact:** Option C provides best UX (no manual version management), maintains backward compatibility

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
- [x] Edge cases identified
- [x] Scope clearly bounded
- [x] Dependencies and assumptions identified

**Readiness:**
- [x] All functional requirements have clear acceptance criteria
- [x] User scenarios cover primary flows
- [x] Ready for technical planning (`/canary.plan`)

---

## Implementation Checklist

### Core Features

<!-- CANARY: REQ=CBIN-137; FEATURE="VersionField"; ASPECT=Engine; STATUS=STUB; UPDATED=2025-10-16 -->
**Feature 1: Version Field Parsing**
- [ ] Add VERSION= field to token parser
- [ ] Store version in database tokens table
- [ ] Display version in scan reports
- [ ] Validate version format (integer or empty)
- **Location hint:** `internal/matcher/token.go` (parser), `internal/storage/` (database)
- **Dependencies:** None

<!-- CANARY: REQ=CBIN-137; FEATURE="VersionHistorySchema"; ASPECT=Storage; STATUS=STUB; UPDATED=2025-10-16 -->
**Feature 2: Version History Database Schema**
- [ ] Create version_history table with schema: version_id, req_id, version_number, created_at, author, change_summary, token_snapshot, spec_snapshot, parent_version_id
- [ ] Create version_annotations table: annotation_id, version_id, created_at, author, note
- [ ] Add indexes: (req_id, version_number), (created_at)
- [ ] Migration script for existing databases
- **Location hint:** `internal/storage/schema.go`, `migrations/`
- **Dependencies:** CBIN-123 TokenStorage

<!-- CANARY: REQ=CBIN-137; FEATURE="ChangeDetection"; ASPECT=Engine; STATUS=STUB; UPDATED=2025-10-16 -->
**Feature 3: Change Detection**
- [ ] Compare current token state to last version in history
- [ ] Detect field changes: STATUS, FEATURE, ASPECT, OWNER, UPDATED, TEST, BENCH
- [ ] Detect specification file content changes (hash-based)
- [ ] Generate change summary from diff
- [ ] Determine when to create new version (any critical field change)
- **Location hint:** `internal/matcher/version.go` (new file)
- **Dependencies:** VersionField, VersionHistorySchema

<!-- CANARY: REQ=CBIN-137; FEATURE="AutoVersioning"; ASPECT=Engine; STATUS=STUB; UPDATED=2025-10-16 -->
**Feature 4: Automatic Version Creation**
- [ ] Integrate change detection into `canary scan`
- [ ] Create new version record when changes detected
- [ ] Increment version number from last version
- [ ] Store token snapshot (JSON serialization of all fields)
- [ ] Optionally store spec snapshot (if spec.md changed)
- [ ] Prevent duplicate versions (idempotent)
- **Location hint:** `cmd/canary/scan.go`, `internal/storage/version.go`
- **Dependencies:** ChangeDetection

<!-- CANARY: REQ=CBIN-137; FEATURE="HistoryCmd"; ASPECT=CLI; STATUS=STUB; UPDATED=2025-10-16 -->
**Feature 5: History Command**
- [ ] Add `canary history <REQ-ID>` subcommand
- [ ] Support flags: --version, --date, --all, --diff, --annotate, --rollback
- [ ] Query version_history table by req_id
- [ ] Format output: version number, timestamp, author, change summary
- [ ] Display token snapshot for specific version
- **Location hint:** `cmd/canary/history.go` (new file)
- **Dependencies:** VersionHistorySchema

<!-- CANARY: REQ=CBIN-137; FEATURE="VersionRetrieval"; ASPECT=Engine; STATUS=STUB; UPDATED=2025-10-16 -->
**Feature 6: Version Retrieval Logic**
- [ ] Retrieve specific version by version_number
- [ ] Retrieve version by date (closest before or on date)
- [ ] Retrieve all versions (ordered by version_number)
- [ ] Deserialize token snapshot to display fields
- [ ] Retrieve specification snapshot if available
- **Location hint:** `internal/storage/version.go`
- **Dependencies:** VersionHistorySchema

<!-- CANARY: REQ=CBIN-137; FEATURE="VersionDiff"; ASPECT=Engine; STATUS=STUB; UPDATED=2025-10-16 -->
**Feature 7: Version Comparison**
- [ ] Compare two version snapshots (JSON diff)
- [ ] Generate field-level change report (STATUS: STUB → IMPL)
- [ ] Compute specification content diff (unified diff format)
- [ ] Calculate timestamp delta
- [ ] Format output for terminal (colored diff)
- **Location hint:** `internal/matcher/diff.go` (new file)
- **Dependencies:** VersionRetrieval

<!-- CANARY: REQ=CBIN-137; FEATURE="VersionAnnotation"; ASPECT=CLI; STATUS=STUB; UPDATED=2025-10-16 -->
**Feature 8: Change Annotation**
- [ ] Add `--annotate "note"` flag to history command
- [ ] Store annotation in version_annotations table
- [ ] Retrieve and display annotations with version history
- [ ] Support editing annotations (update note by annotation_id)
- **Location hint:** `cmd/canary/history.go`, `internal/storage/annotation.go`
- **Dependencies:** VersionHistorySchema

<!-- CANARY: REQ=CBIN-137; FEATURE="AgentHistoryCmd"; ASPECT=Docs; STATUS=STUB; UPDATED=2025-10-16 -->
**Feature 9: Agent Slash Command**
- [ ] Create `.claude/commands/canary.history.md` template
- [ ] Document usage patterns for agents (load historical context)
- [ ] Provide examples: `canary history CBIN-137 --all`
- [ ] Explain output format for parsing
- **Location hint:** `.claude/commands/canary.history.md`
- **Dependencies:** HistoryCmd

<!-- CANARY: REQ=CBIN-137; FEATURE="VersionPruning"; ASPECT=CLI; STATUS=STUB; UPDATED=2025-10-16 -->
**Feature 10: Version Pruning**
- [ ] Add `canary history <REQ-ID> --prune` command
- [ ] Configurable retention policy (e.g., keep last 50 versions)
- [ ] Preserve milestone versions (STUB→IMPL, IMPL→TESTED)
- [ ] Delete old version records and snapshots
- [ ] Report pruned versions and space saved
- **Location hint:** `cmd/canary/history.go`, `internal/storage/prune.go`
- **Dependencies:** VersionHistorySchema

### Testing Requirements

<!-- CANARY: REQ=CBIN-137; FEATURE="UnitTests"; ASPECT=Engine; STATUS=STUB; TEST=TestCBIN137_Engine; UPDATED=2025-10-16 -->
**Unit Tests:**
- [ ] Test change detection logic (detect STATUS, FEATURE changes)
- [ ] Test version number incrementing
- [ ] Test token snapshot serialization/deserialization
- [ ] Test version retrieval by number and date
- [ ] Test version diff computation
- [ ] Test annotation CRUD operations
- **Location hint:** `internal/matcher/version_test.go`, `internal/storage/version_test.go`

<!-- CANARY: REQ=CBIN-137; FEATURE="IntegrationTests"; ASPECT=CLI; STATUS=STUB; TEST=TestCBIN137_Integration; UPDATED=2025-10-16 -->
**Integration Tests:**
- [ ] Test end-to-end version creation during scan
- [ ] Test history command with real database
- [ ] Test version diff output formatting
- [ ] Test agent slash command usage
- [ ] Test version pruning with retention policy
- **Location hint:** `cmd/canary/history_test.go`

### Documentation

<!-- CANARY: REQ=CBIN-137; FEATURE="HistoryCmdDocs"; ASPECT=Docs; STATUS=STUB; UPDATED=2025-10-16 -->
**CLI Documentation:**
- [ ] Add `canary history --help` documentation
- [ ] Document all flags: --version, --date, --all, --diff, --annotate, --rollback, --prune
- [ ] Provide usage examples in README
- [ ] Explain version numbering scheme
- **Location hint:** `cmd/canary/history.go` (cobra help), `README.md`

---

**Agent Instructions:**

After implementing each feature:
1. Update the CANARY token in the spec from `STATUS=STUB` to `STATUS=IMPL`
2. Add the same token to your source code at the implementation location
3. Add `TEST=TestName` when tests are written
4. Run `canary implement CBIN-137` to see implementation progress

---

## CANARY Tokens Reference

**Main requirement token** (add to primary implementation file):
```go
// CANARY: REQ=CBIN-137; FEATURE="RequirementHistory"; ASPECT=Engine; STATUS=IMPL; UPDATED=2025-10-16
```

**Sub-feature tokens** (use the specific feature names from Implementation Checklist):
```go
// CANARY: REQ=CBIN-137; FEATURE="VersionField"; ASPECT=Engine; STATUS=IMPL; TEST=TestVersionField; UPDATED=2025-10-16
```

**Example token with version field:**
```go
// CANARY: REQ=CBIN-137; FEATURE="RequirementHistory"; ASPECT=Engine; STATUS=IMPL; VERSION=3; UPDATED=2025-10-16
```

**Use `canary implement CBIN-137` to find:**
- Which features are implemented vs. still TODO
- Exact file locations and line numbers
- Context around each implementation point
