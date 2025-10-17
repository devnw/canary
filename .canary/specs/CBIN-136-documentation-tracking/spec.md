<!-- CANARY: REQ=CBIN-115; FEATURE="SpecTemplate"; ASPECT=Docs; STATUS=IMPL; OWNER=canary; UPDATED=2025-10-16 -->
# Feature Specification: Documentation Tracking and Consistency

**Requirement ID:** CBIN-136
**Status:** STUB
**Created:** 2025-10-16
**Last Updated:** 2025-10-16

## Overview

**Purpose:** Enable AI agents and developers to maintain consistent, up-to-date documentation across multiple documentation types (user guides, technical docs, feature docs, API docs) by linking CANARY tokens directly to documentation files and tracking documentation freshness through content hashing. This ensures documentation remains synchronized with code changes and provides clear signals when documentation needs updating.

**Scope:**
- Included: Link CANARY tokens to documentation files via DOC= field
- Included: Multiple documentation types (user, technical, feature, api, architecture)
- Included: Content hash tracking to detect stale documentation
- Included: Automated staleness detection comparing doc hash vs actual content
- Included: Documentation templates for consistent structure
- Included: Agent-driven documentation generation and updates
- Excluded: Automated doc generation from code (manual/agent-written preferred)
- Excluded: Version control integration (git handles versioning)
- Excluded: Real-time doc watching (batch scan approach)

## User Stories

### Primary User Stories

**US-1: Link Requirements to Documentation**
As a developer,
I want to associate CANARY tokens with their documentation files,
So that I can quickly find relevant documentation for any requirement.

**Acceptance Criteria:**
- [ ] Can add DOC= field to CANARY token pointing to documentation file path
- [ ] Can specify multiple documentation files (comma-separated)
- [ ] Can specify documentation type (user, technical, feature, api, architecture)
- [ ] Documentation links are validated during scan (file exists check)
- [ ] Can query tokens by documentation file path

**US-2: Track Documentation Freshness**
As a developer,
I want to know when documentation becomes outdated,
So that I can prioritize documentation updates alongside code changes.

**Acceptance Criteria:**
- [ ] Documentation files have content hash stored in CANARY token (DOC_HASH= field)
- [ ] Scanning detects when actual file hash differs from stored hash
- [ ] Stale documentation flagged in scan reports with "DOCS_STALE" indicator
- [ ] Can query for all requirements with stale documentation
- [ ] Hash algorithm is stable (SHA256) and deterministic

**US-3: Generate Consistent Documentation**
As an AI agent,
I want documentation templates for different doc types,
So that I can create consistent, well-structured documentation.

**Acceptance Criteria:**
- [ ] Templates exist for: user guide, technical doc, feature doc, API doc, architecture doc
- [ ] Templates include standard sections (Overview, Usage, Examples, etc.)
- [ ] Templates include CANARY token placeholder for doc tracking
- [ ] Agent can invoke template generation via command/slash command
- [ ] Templates follow markdown format for universal compatibility

**US-4: Update Documentation with Hash Tracking**
As an AI agent,
I want to update documentation and automatically update the hash,
So that the system recognizes documentation is now current.

**Acceptance Criteria:**
- [ ] After updating documentation file, hash is recalculated
- [ ] CANARY token in code updated with new DOC_HASH= value
- [ ] Update command/slash command handles both doc file and token update
- [ ] Batch update capability for multiple documentation files
- [ ] Dry-run mode to preview hash changes without committing

### Secondary User Stories

**US-5: Query Documentation Status**
As a project maintainer,
I want to see which requirements have missing or stale documentation,
So that I can prioritize documentation work.

**Acceptance Criteria:**
- [ ] Can list all tokens without DOC= field (missing docs)
- [ ] Can list all tokens with DOC_HASH mismatch (stale docs)
- [ ] Can filter by documentation type (e.g., show only stale user guides)
- [ ] Report shows: req_id, feature, doc path, staleness reason
- [ ] Export to CSV/JSON for tracking

**US-6: Documentation Coverage Metrics**
As a project maintainer,
I want metrics on documentation coverage and freshness,
So that I can track documentation quality over time.

**Acceptance Criteria:**
- [ ] Percentage of requirements with documentation
- [ ] Percentage of documentation that is current (hash matches)
- [ ] Breakdown by documentation type
- [ ] Trend tracking via checkpoints
- [ ] Integration with existing `canary scan` reporting

## Functional Requirements

### FR-1: CANARY Token DOC Field
**Priority:** High
**Description:** Extend CANARY token format to include DOC= field pointing to documentation file paths, with support for multiple docs and type specification
**Acceptance:** Token parsing recognizes DOC=path/to/doc.md and DOC=type:path/to/doc.md formats

### FR-2: CANARY Token DOC_HASH Field
**Priority:** High
**Description:** Extend CANARY token format to include DOC_HASH= field containing SHA256 hash of documentation content
**Acceptance:** Token parsing recognizes DOC_HASH=abc123... format, validates hash format

### FR-3: Documentation Scanning
**Priority:** High
**Description:** Scan command reads documentation files referenced in DOC= fields and compares content hashes against DOC_HASH= values
**Acceptance:** Scanning detects stale documentation and reports mismatches with clear indicators

### FR-4: Documentation Templates
**Priority:** High
**Description:** Provide templates for user, technical, feature, API, and architecture documentation with standard sections
**Acceptance:** Templates are accessible via command or file system, include CANARY token placeholders

### FR-5: Documentation Update Command
**Priority:** High
**Description:** Command to update documentation and automatically recalculate and update hash in CANARY token
**Acceptance:** Single command updates both doc file and token DOC_HASH= field atomically

### FR-6: Documentation Status Queries
**Priority:** Medium
**Description:** Query capabilities to find requirements with missing, stale, or current documentation
**Acceptance:** Queries return accurate results with filtering by doc type and staleness

### FR-7: Documentation Metrics
**Priority:** Medium
**Description:** Reporting on documentation coverage percentage and freshness percentage
**Acceptance:** Metrics displayed in scan output and checkpoint reports

### FR-8: Agent Slash Command Integration
**Priority:** High
**Description:** Slash commands for agents to generate, update, and verify documentation
**Acceptance:** Commands accessible via `/canary.doc-create`, `/canary.doc-update`, `/canary.doc-verify`

## Success Criteria

**Quantitative Metrics:**
- [ ] Documentation coverage tracked for 100% of TESTED/BENCHED requirements
- [ ] Documentation staleness detected within 1 second during scan
- [ ] Hash calculation adds < 100ms overhead per documentation file
- [ ] 90% of requirements with documentation have current hashes
- [ ] Agents can generate documentation in < 30 seconds using templates

**Qualitative Measures:**
- [ ] Developers can locate documentation for any requirement in < 5 seconds
- [ ] Documentation consistency improves (measured by template adherence)
- [ ] Documentation updates tracked alongside code changes
- [ ] Agents successfully use templates to create well-structured docs
- [ ] Stale documentation visible in regular scans, prompting updates

## User Scenarios & Testing

### Scenario 1: Link Requirement to Documentation (Happy Path)
**Given:** Developer has implemented feature CBIN-105 with user guide at `docs/user/authentication.md`
**When:** They add `DOC=user:docs/user/authentication.md` to CANARY token
**Then:** Scan command validates file exists, calculates hash, and suggests adding `DOC_HASH=abc123...` to token

### Scenario 2: Detect Stale Documentation
**Given:** Documentation file `docs/user/authentication.md` has been modified
**When:** Scan command runs and compares file hash to DOC_HASH= in token
**Then:** Report shows "DOCS_STALE: CBIN-105 - docs/user/authentication.md (hash mismatch)"

### Scenario 3: Generate Documentation from Template (Agent Workflow)
**Given:** AI agent needs to document new feature CBIN-107
**When:** Agent runs `/canary.doc-create CBIN-107 --type feature --output docs/features/oauth2.md`
**Then:** System creates documentation file from template, includes CANARY token, returns file path for agent to edit

### Scenario 4: Update Documentation and Hash
**Given:** Developer updates `docs/api/endpoints.md` for CBIN-108
**When:** They run `canary doc-update CBIN-108 docs/api/endpoints.md`
**Then:** System recalculates hash, updates CANARY token DOC_HASH= field in source code, confirms update

### Scenario 5: Query Stale Documentation
**Given:** Project has multiple requirements with outdated documentation
**When:** Maintainer runs `canary doc-status --stale`
**Then:** Report lists all requirements with DOC_HASH mismatches, grouped by documentation type

### Scenario 6: Documentation Coverage Metrics
**Given:** Project has 50 requirements, 35 with documentation
**When:** Maintainer runs `canary scan --doc-metrics`
**Then:** Report shows "Documentation Coverage: 70% (35/50), Fresh: 80% (28/35), Stale: 20% (7/35)"

### Scenario 7: Multiple Documentation Types
**Given:** Feature CBIN-110 has user guide, API doc, and architecture doc
**When:** Developer adds `DOC=user:docs/user/feature.md,api:docs/api/feature.md,arch:docs/arch/feature.md`
**Then:** Scan tracks all three documentation files independently with separate hash checks

### Scenario 8: Missing Documentation File (Error Case)
**Given:** CANARY token references non-existent file `DOC=docs/missing.md`
**When:** Scan command runs
**Then:** Report shows "DOC_MISSING: CBIN-112 - docs/missing.md (file not found)" with suggestion to create or remove DOC= field

## Key Entities

### Entity 1: DocumentationLink
**Attributes:**
- req_id: Requirement identifier (CBIN-XXX)
- doc_type: Documentation type (user, technical, feature, api, architecture)
- doc_path: Relative path to documentation file
- doc_hash: SHA256 hash of documentation content
- last_checked: Timestamp of last staleness check
- status: DOC_CURRENT, DOC_STALE, DOC_MISSING

**Relationships:**
- Linked to Token entity via req_id
- Maps to physical documentation file on filesystem

### Entity 2: DocumentationTemplate
**Attributes:**
- template_type: Type (user, technical, feature, api, architecture)
- template_path: Path to template file
- sections: List of standard sections
- canary_token_placeholder: Where to insert token

**Relationships:**
- Used to generate new documentation files
- Referenced by doc-create command

### Entity 3: DocumentationMetrics
**Attributes:**
- total_requirements: Count of all requirements
- documented_count: Requirements with DOC= field
- current_count: Requirements with matching DOC_HASH
- stale_count: Requirements with mismatched DOC_HASH
- missing_count: Requirements without DOC= field
- coverage_percentage: documented_count / total_requirements
- freshness_percentage: current_count / documented_count

**Relationships:**
- Aggregated from Token and DocumentationLink entities
- Included in checkpoint snapshots

## Assumptions

- Documentation files are markdown format (.md)
- Documentation files stored in `docs/` directory by convention
- SHA256 hash algorithm sufficient for change detection (no need for content-aware diffing)
- Documentation updates happen less frequently than code changes (acceptable to require manual hash update)
- Agents have write access to both documentation files and source code files
- Git handles version control; CANARY tracks synchronization status only

## Constraints

**Technical Constraints:**
- Hash calculation must be deterministic (line endings normalized)
- Multiple documentation files per requirement supported (comma-separated)
- DOC= field limited to 500 characters to fit in CANARY token
- Hash displayed in abbreviated form (first 16 characters) for readability
- File paths must be relative to project root

**Business Constraints:**
- Reuse existing CANARY token infrastructure
- Minimal performance impact on scan command (< 10% overhead)
- Compatible with existing CANARY tools and workflows
- No external dependencies (use standard library for hashing)

**Regulatory Constraints:**
- None

## Out of Scope

- Automated documentation generation from code comments (prefer manual/agent-written)
- Real-time documentation watching (use batch scan approach)
- Documentation versioning (git handles this)
- Documentation diff/merge tools (use git diff)
- Documentation translation/localization
- Documentation rendering/preview (use markdown viewers)
- Documentation search indexing (use text search tools)
- Documentation link checking (external links)
- Auto-updating documentation content (only hash tracking)

## Dependencies

- CBIN-104: CanaryCLI (extend scan command)
- CBIN-123: TokenStorage (add DOC, DOC_HASH fields to database schema)
- Existing file I/O infrastructure
- SHA256 hashing (standard library)

## Risks & Mitigation

| Risk | Impact | Probability | Mitigation |
|------|--------|-------------|------------|
| Hash changes from line ending differences | Medium | Medium | Normalize line endings before hashing, document normalization rules |
| Documentation paths become stale when files move | Medium | Low | Validate paths during scan, show clear errors, provide update command |
| Large documentation files slow scanning | Low | Low | Hash calculation is fast (< 10ms per file), cache hashes in database |
| Multiple documentation types confuse users | Low | Medium | Provide clear documentation type taxonomy, show examples |
| Agents create inconsistent documentation | Medium | Medium | Enforce templates, provide clear examples, validate structure |

## Clarifications Needed

[NEEDS CLARIFICATION: Should DOC_HASH be required or optional in CANARY tokens?]
**Options:**
A) Required for all TESTED/BENCHED requirements (enforce documentation)
B) Optional (documentation nice-to-have but not mandatory)
C) Required only for specific aspects (e.g., API, Docs)
**Impact:** Option A ensures documentation completeness but adds burden. Option B provides flexibility. Option C balances both.

[NEEDS CLARIFICATION: How should documentation type taxonomy be extended?]
**Options:**
A) Fixed list: user, technical, feature, api, architecture
B) Extensible via configuration file
C) Free-form user-defined types
**Impact:** Option A ensures consistency. Option B provides flexibility with governance. Option C risks inconsistency.

[NEEDS CLARIFICATION: Should agents auto-update hashes when modifying documentation?]
**Options:**
A) Yes - agent command updates both doc and token hash atomically
B) No - separate manual step to update hash
C) Configurable via flag (default: auto-update)
**Impact:** Option A reduces friction but requires careful implementation. Option B ensures developer awareness. Option C provides flexibility.

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

<!-- CANARY: REQ=CBIN-136; FEATURE="TokenDocField"; ASPECT=Engine; STATUS=IMPL; UPDATED=2025-10-17 -->
**Feature 1: CANARY Token DOC Field Parsing**
- [ ] Extend token parser to recognize DOC= field
- [ ] Support comma-separated multiple docs
- [ ] Support type prefix (e.g., `user:path/to/doc.md`)
- [ ] Validate paths during parsing
- **Location hint:** `tools/canary/main.go` (scanner) or `internal/scanner/parser.go`
- **Dependencies:** None

<!-- CANARY: REQ=CBIN-136; FEATURE="TokenDocHashField"; ASPECT=Engine; STATUS=IMPL; UPDATED=2025-10-17 -->
**Feature 2: CANARY Token DOC_HASH Field Parsing**
- [ ] Extend token parser to recognize DOC_HASH= field
- [ ] Validate hash format (hex string, 64 chars for SHA256)
- [ ] Support abbreviated hash display (first 16 chars)
- **Location hint:** `tools/canary/main.go` (scanner) or `internal/scanner/parser.go`
- **Dependencies:** None

<!-- CANARY: REQ=CBIN-136; FEATURE="DocHashCalculation"; ASPECT=Engine; STATUS=TESTED; TEST=TestCANARY_CBIN_136_Engine_HashCalculation; UPDATED=2025-10-17 -->
**Feature 3: Documentation Hash Calculation**
- [ ] Read documentation file content
- [ ] Normalize line endings (LF)
- [ ] Calculate SHA256 hash
- [ ] Return hex-encoded hash string
- **Location hint:** `internal/docs/hash.go` (new file)
- **Dependencies:** Standard library crypto/sha256

<!-- CANARY: REQ=CBIN-136; FEATURE="DocStalenessDetection"; ASPECT=Engine; STATUS=TESTED; TEST=TestCANARY_CBIN_136_Engine_StalenessDetection; UPDATED=2025-10-17 -->
**Feature 4: Documentation Staleness Detection**
- [ ] Compare DOC_HASH field to calculated file hash
- [ ] Flag mismatches as DOCS_STALE
- [ ] Flag missing files as DOC_MISSING
- [ ] Flag missing DOC_HASH as DOC_UNHASHED
- **Location hint:** `tools/canary/main.go` (scanner) or `internal/docs/checker.go`
- **Dependencies:** Feature 3 (DocHashCalculation)

### Documentation Templates

<!-- CANARY: REQ=CBIN-136; FEATURE="DocTemplates"; ASPECT=Docs; STATUS=IMPL; UPDATED=2025-10-17 -->
**Documentation Templates:**
- [ ] Create user guide template
- [ ] Create technical doc template
- [ ] Create feature doc template
- [ ] Create API doc template
- [ ] Create architecture doc template
- **Location hint:** `.canary/templates/docs/` directory
- **Dependencies:** None

### CLI Commands

<!-- CANARY: REQ=CBIN-136; FEATURE="DocCreateCommand"; ASPECT=CLI; STATUS=TESTED; TEST=TestCANARY_CBIN_136_CLI_DocWorkflow; UPDATED=2025-10-17 -->
**Doc Create Command:**
- [ ] Add `canary doc-create <REQ-ID> --type <type> --output <path>` command
- [ ] Copy template to output path
- [ ] Replace placeholders with requirement details
- [ ] Calculate and suggest DOC_HASH value
- **Location hint:** `cmd/canary/doc_commands.go` (new file)
- **Dependencies:** Feature 5 (DocTemplates)

<!-- CANARY: REQ=CBIN-136; FEATURE="DocUpdateCommand"; ASPECT=CLI; STATUS=TESTED; TEST=TestCANARY_CBIN_136_CLI_BatchUpdate; UPDATED=2025-10-17 -->
**Doc Update Command:**
- [ ] Add `canary doc-update <REQ-ID> <doc-path>` command
- [ ] Recalculate documentation file hash
- [ ] Find CANARY token in source code
- [ ] Update DOC_HASH= field in token
- [ ] Support dry-run mode
- **Location hint:** `cmd/canary/doc_commands.go`
- **Dependencies:** Feature 3 (DocHashCalculation)

<!-- CANARY: REQ=CBIN-136; FEATURE="DocStatusCommand"; ASPECT=CLI; STATUS=TESTED; TEST=TestCANARY_CBIN_136_CLI_DocReport; UPDATED=2025-10-17 -->
**Doc Status Command:**
- [ ] Add `canary doc-status [--stale] [--missing]` command
- [ ] Query database for tokens with DOC= field
- [ ] Check staleness for each documentation file
- [ ] Display report with req_id, doc_path, status
- [ ] Support CSV/JSON export
- **Location hint:** `cmd/canary/doc_commands.go`
- **Dependencies:** Feature 4 (DocStalenessDetection)

### Database Schema

<!-- CANARY: REQ=CBIN-136; FEATURE="DocDatabaseSchema"; ASPECT=Storage; STATUS=IMPL; UPDATED=2025-10-17 -->
**Database Schema Extension:**
- [ ] Add `doc_path` column to tokens table (TEXT)
- [ ] Add `doc_hash` column to tokens table (TEXT)
- [ ] Add `doc_type` column to tokens table (TEXT)
- [ ] Add `doc_checked_at` column to tokens table (DATETIME)
- [ ] Add `doc_status` column to tokens table (TEXT)
- [ ] Create migration script
- **Location hint:** `internal/storage/migrations/` or schema definition
- **Dependencies:** CBIN-123 (TokenStorage)

### Scan Integration

<!-- CANARY: REQ=CBIN-136; FEATURE="ScanDocumentation"; ASPECT=CLI; STATUS=IMPL; UPDATED=2025-10-17 -->
**Scan Command Integration:**
- [ ] Extend scan command to check documentation
- [ ] Report documentation coverage metrics
- [ ] Report stale documentation count
- [ ] Add `--doc-report` flag for detailed doc status
- [ ] Add `--skip-docs` flag to disable doc checking
- **Location hint:** `tools/canary/main.go` (scan command)
- **Dependencies:** Feature 4 (DocStalenessDetection)

### Agent Integration

<!-- CANARY: REQ=CBIN-136; FEATURE="AgentDocCommands"; ASPECT=Docs; STATUS=IMPL; UPDATED=2025-10-17 -->
**Agent Slash Commands:**
- [ ] Create `/canary.doc-create` slash command
- [ ] Create `/canary.doc-update` slash command
- [ ] Create `/canary.doc-verify` slash command
- [ ] Provide usage examples and patterns
- **Location hint:** `.claude/commands/` directory
- **Dependencies:** Features 6, 7, 8 (CLI commands)

### Testing Requirements

<!-- CANARY: REQ=CBIN-136; FEATURE="DocUnitTests"; ASPECT=Engine; STATUS=TESTED; TEST=TestCANARY_CBIN_136_Engine_HashCalculation; UPDATED=2025-10-17 -->
**Unit Tests:**
- [ ] Test hash calculation determinism
- [ ] Test line ending normalization
- [ ] Test token parsing with DOC=/DOC_HASH= fields
- [ ] Test staleness detection logic
- **Location hint:** `internal/docs/hash_test.go`, `internal/scanner/parser_test.go`

<!-- CANARY: REQ=CBIN-136; FEATURE="DocIntegrationTests"; ASPECT=CLI; STATUS=TESTED; TEST=TestCANARY_CBIN_136_CLI_DocWorkflow; UPDATED=2025-10-17 -->
**Integration Tests:**
- [ ] Test end-to-end doc-create workflow
- [ ] Test end-to-end doc-update workflow
- [ ] Test scan with documentation checking
- [ ] Test documentation metrics reporting
- **Location hint:** `cmd/canary/doc_commands_test.go`

### Documentation

<!-- CANARY: REQ=CBIN-136; FEATURE="DocSystemDocs"; ASPECT=Docs; STATUS=IMPL; UPDATED=2025-10-17 -->
**Documentation System Documentation:**
- [ ] Update README with DOC=/DOC_HASH= field documentation
- [ ] Create docs/documentation-tracking.md guide
- [ ] Document documentation types taxonomy
- [ ] Provide hash calculation examples
- [ ] Document agent workflows
- **Location hint:** `README.md`, `docs/` directory

---

**Agent Instructions:**

After implementing each feature:
1. Update the CANARY token in the spec from `STATUS=STUB` to `STATUS=IMPL`
2. Add the same token to your source code at the implementation location
3. Add `TEST=TestName` when tests are written
4. Run `canary implement CBIN-136` to see implementation progress
5. For documentation features, ensure self-documentation (eat our own dog food!)

---

## CANARY Tokens Reference

**Main requirement token** (add to primary implementation file):
```go
// CANARY: REQ=CBIN-136; FEATURE="DocumentationTracking"; ASPECT=Docs; STATUS=IMPL; DOC=user:docs/user/documentation-tracking-guide.md; DOC_HASH=1e32f44252c80284; UPDATED=2025-10-16
```

**Sub-feature tokens** (use the specific feature names from Implementation Checklist):
```go
// CANARY: REQ=CBIN-136; FEATURE="DocHashCalculation"; ASPECT=Engine; STATUS=IMPL; TEST=TestDocHashCalculation; UPDATED=2025-10-16
```

**Documentation file example with CANARY token:**
```markdown
<!-- CANARY: REQ=CBIN-136; FEATURE="DocumentationTracking"; ASPECT=Docs; STATUS=CURRENT; UPDATED=2025-10-16 -->
# Documentation Tracking Guide

This feature enables tracking of documentation freshness...
```

**Use `canary implement CBIN-136` to find:**
- Which features are implemented vs. still TODO
- Exact file locations and line numbers
- Context around each implementation point
- Documentation files linked to this requirement
