# CANARY: REQ=CBIN-143; FEATURE="EvidenceField"; ASPECT=Storage; STATUS=STUB; UPDATED=2025-10-17
# Feature Specification: Evidence Field in CANARY Tokens

**Requirement ID:** CBIN-143
**Status:** STUB
**Created:** 2025-10-17
**Last Updated:** 2025-10-17

## Overview

**Purpose:** Enable developers and AI agents to attach concrete evidence to CANARY tokens that documents what was actually implemented, why design decisions were made, and where specific code addresses the requirement. This creates an audit trail linking requirements to implementation details, making code archaeology faster and providing context for future maintenance.

**Scope:** This feature adds an optional `EVIDENCE` field to CANARY token syntax, stores evidence in the database alongside token metadata, and provides query/display capabilities for evidence. Evidence can be multi-line, contain specific code references, rationale, or links to related documentation. Excludes automatic evidence generation or validation - evidence is manually provided by developers/agents.

## User Stories

### Primary User Stories

**US-1: Attach Implementation Evidence**
As a developer or AI agent,
I want to add evidence to a CANARY token explaining what was implemented and why,
So that future maintainers understand the reasoning behind implementation decisions without reading entire files.

**Acceptance Criteria:**
- [ ] CANARY token syntax supports optional `EVIDENCE="..."` field
- [ ] Evidence can span multiple lines in token comments
- [ ] Scanner parses and stores evidence in database
- [ ] Evidence persists across scans and updates

**US-2: Query Evidence**
As a developer,
I want to retrieve evidence for a specific requirement or feature,
So that I can understand why code was written a certain way before making changes.

**Acceptance Criteria:**
- [ ] `canary show CBIN-XXX` displays all evidence for requirement
- [ ] Evidence appears alongside token status and location
- [ ] Can filter tokens with/without evidence
- [ ] Evidence is searchable by keyword

**US-3: Multi-Line Evidence**
As a developer,
I want to write detailed evidence that includes code snippets or multiple points,
So that I can provide comprehensive context without cramming everything into one line.

**Acceptance Criteria:**
- [ ] Evidence can continue across multiple comment lines
- [ ] Scanner correctly concatenates multi-line evidence
- [ ] Formatting (line breaks, indentation) is preserved where meaningful
- [ ] Long evidence doesn't break token parsing

**US-4: Update Evidence**
As a developer,
I want to update evidence when implementation details change,
So that evidence stays current with code evolution.

**Acceptance Criteria:**
- [ ] Changing evidence in source code updates database on next scan
- [ ] Can view evidence history to see what changed
- [ ] System detects when evidence was added/modified via UPDATED field

### Secondary User Stories (if applicable)

**US-5: Link Evidence to Gap Analysis**
As a maintainer,
I want evidence to reference gap analysis entries when implementations were corrected,
So that I can see the learning path from mistake to fix.

**Acceptance Criteria:**
- [ ] Evidence can reference gap IDs (if CBIN-140 is implemented)
- [ ] Gap analysis entries can link back to evidence
- [ ] Evidence shows what was learned from previous mistakes

## Functional Requirements

### FR-1: Evidence Syntax Extension
**Priority:** High
**Description:** System must extend CANARY token syntax to support optional `EVIDENCE` field with quoted text that can span multiple comment lines. Evidence should be structured to include filename and line numbers when referencing specific code locations.
**Acceptance:** Parser recognizes `EVIDENCE="text"`; handles multi-line evidence; preserves evidence text exactly as written; evidence is optional (tokens without evidence remain valid); supports filename:line notation.

### FR-2: Evidence Storage
**Priority:** High
**Description:** System must store evidence in database with association to specific CANARY token (REQ-ID + FEATURE + file location).
**Acceptance:** Database schema includes evidence field; evidence is persisted; can be queried independently; supports UTF-8 and special characters.

### FR-3: Evidence Display
**Priority:** High
**Description:** System must display evidence when showing token information via CLI commands.
**Acceptance:** `canary show`, `canary list`, and similar commands include evidence in output; evidence is formatted for readability; long evidence is not truncated unless user requests summary.

### FR-4: Evidence Search
**Priority:** Medium
**Description:** System must allow searching tokens by evidence content.
**Acceptance:** `canary grep` or `canary search` can filter by evidence keywords; search is case-insensitive; supports partial matches; returns tokens with matching evidence.

### FR-5: Evidence History
**Priority:** Low
**Description:** System should track evidence changes over time to show evolution of understanding.
**Acceptance:** When evidence is modified, old version is archived; can view evidence at specific dates; history shows who/when evidence changed.

### FR-6: Evidence Validation
**Priority:** Low
**Description:** System should warn when evidence is missing for implemented features (status IMPL or TESTED).
**Acceptance:** `canary validate` reports tokens without evidence; can configure whether evidence is required; warnings don't block operations.

### FR-7: Evidence Length Limits
**Priority:** Medium
**Description:** System must enforce soft length limits on evidence to encourage concise, actionable documentation while allowing detailed context when necessary.
**Acceptance:** Evidence exceeding 500 characters triggers warning (not error); warning suggests breaking into multiple tokens or referencing external docs; database supports up to 4KB evidence; display shows character count for long evidence.

## Success Criteria

**Quantitative Metrics:**
- [ ] 80% of IMPL/TESTED tokens include evidence within 1 month of feature deployment
- [ ] Evidence queries return results in under 2 seconds for 1000+ tokens
- [ ] Scanner parses multi-line evidence without errors in 100% of test cases
- [ ] Evidence field adds less than 10% to database size

**Qualitative Measures:**
- [ ] Developers find evidence helpful when maintaining unfamiliar code
- [ ] Evidence reduces time spent understanding implementation decisions by 50%
- [ ] Evidence contains actionable information (not just rephrasing requirement)
- [ ] Evidence is consistently formatted and easy to read

## User Scenarios & Testing

### Scenario 1: Adding Single-Line Evidence with Location
**Given:** Developer implements memory leak fix
**When:** Developer adds CANARY token with evidence:
```
// CANARY: REQ=REQ-C4-001; FEATURE="MemoryLeakFix"; ASPECT=Environment; STATUS=IMPLEMENTED; OWNER=core; UPDATED=2025-10-17;
//         EVIDENCE="Free old arrays in withBinding() at binding.go:145 to prevent memory leaks"
```
**Then:** Scanner parses evidence; stores evidence with file:line reference in database; evidence appears when querying REQ-C4-001 with clickable location

### Scenario 2: Adding Structured Multi-Line Evidence
**Given:** Developer implements complex algorithm with multiple decision points
**When:** Developer adds multi-line evidence with file:line references:
```
// CANARY: REQ=CBIN-055; FEATURE="GraphTraversal"; ASPECT=API; STATUS=IMPL; UPDATED=2025-10-17;
//         EVIDENCE="Implemented BFS at graph.go:78 instead of DFS because:
//                   1. Shorter path is priority for user experience
//                   2. Graph is wide (avg branching factor 10)
//                   3. Memory is acceptable (max 10k nodes)
//                   Queue impl: queue.go:45; See benchmark: graph_test.go:234"
```
**Then:** Scanner concatenates lines; preserves structure; extracts file:line references; stores complete evidence; displays with clickable locations

### Scenario 3: Querying Evidence
**Given:** Multiple tokens exist with evidence about "performance optimization"
**When:** Developer runs `canary search --evidence "performance"`
**Then:** System returns all tokens with "performance" in evidence field; displays token ID, feature, and evidence snippet; developer can drill down for full details

### Scenario 4: Soft Length Limit Warning
**Given:** Developer writes detailed evidence exceeding 500 characters
**When:** Developer runs `canary scan` on file with long evidence
**Then:** System parses and stores evidence; warns "Evidence for CBIN-055 is 612 chars (>500); consider breaking into multiple tokens or referencing external docs"; scan completes successfully

### Scenario 5: Missing Evidence Warning
**Given:** Token has STATUS=IMPL but no EVIDENCE field
**When:** Developer runs `canary validate --check-evidence`
**Then:** System reports "Warning: CBIN-042 UserAuth has STATUS=IMPL but no EVIDENCE"; suggests adding evidence to document implementation

### Scenario 6: Updating Evidence
**Given:** Implementation approach changed; original evidence is outdated
**When:** Developer updates evidence in source code and runs `canary scan`
**Then:** Database updates evidence for that token; old evidence is archived in history; UPDATED field reflects change date

### Scenario 7: Evidence Linking to Gap Analysis
**Given:** Implementation was initially wrong (GAP-003) and then corrected
**When:** Developer adds evidence: `EVIDENCE="Fixed incorrect mutex usage (see GAP-003); now using RWMutex for read-heavy workload"`
**Then:** Evidence references gap analysis; gap entry can be retrieved from evidence; learning is documented

## Key Entities (if data-driven feature)

### Entity 1: CANARYToken (extended)
**Attributes (new/modified):**
- evidence: Text field containing implementation evidence
- evidence_updated: Timestamp when evidence was last modified
- evidence_lines: Number of lines in evidence (for display purposes)

**Relationships:**
- Links to EvidenceHistory for tracking changes
- May reference GapAnalysisEntry via evidence text

### Entity 2: EvidenceHistory
**Attributes:**
- history_id: Unique identifier
- token_id: Reference to CANARY token
- evidence_text: Historical evidence content
- changed_at: When evidence was changed
- changed_by: Who modified evidence (if trackable)
- change_reason: Optional note about why evidence changed

**Relationships:**
- Belongs to CANARYToken
- Ordered by timestamp for viewing history

## Assumptions

- Evidence is written in natural language (English) with optional code snippets
- Evidence follows structured conventions: include filename:line references when possible (e.g., "at binding.go:145")
- Evidence is manually authored by developers/agents, not auto-generated
- Evidence quality depends on developer discipline and guidelines
- Evidence will be version-controlled as part of source code
- Most evidence will be concise (under 500 chars); longer evidence is allowed but warned
- Evidence does not replace code comments; provides requirement-level context
- File:line references use format "filename.ext:line" or "at filename.ext:line"

## Constraints

**Technical Constraints:**
- Must maintain backward compatibility with existing CANARY tokens without EVIDENCE
- Evidence field must not break existing parser or scanner
- Database schema must accommodate very long evidence (up to 4KB)
- Multi-line evidence must follow comment syntax rules

**Business Constraints:**
- Adding evidence should take less than 2 minutes per token
- Evidence should not significantly increase scan time
- Storage overhead for evidence should be minimal

**Regulatory Constraints:**
- Evidence may contain proprietary information; must respect access controls
- Evidence should not include sensitive data (passwords, keys, PII)

## Out of Scope

- Automatic generation of evidence from commit messages or code analysis
- AI-powered evidence suggestion or completion
- Evidence validation for accuracy or completeness (only format/length validation)
- Evidence translation to other languages
- Diff/merge tools for conflicting evidence
- Formal evidence schemas (JSON/YAML/XML) - evidence uses natural language with conventions
- Integration with external documentation systems
- Automatic clickable links in terminal (CLI can show file:line but navigation depends on terminal/editor)

## Dependencies

- CANARY scanner and parser (CBIN-101, CBIN-102)
- Database schema and storage (existing token storage)
- CLI display commands (`canary show`, `canary list`)
- (Optional) Gap analysis system (CBIN-140) for cross-referencing

## Risks & Mitigation

| Risk | Impact | Probability | Mitigation |
|------|--------|-------------|------------|
| Evidence becomes stale and misleading | High | Medium | Track evidence age; warn when evidence is old; encourage updates |
| Evidence quality is inconsistent | Medium | High | Provide guidelines and examples; review evidence in code reviews |
| Multi-line parsing is fragile | Medium | Medium | Extensive testing; clear syntax rules; error messages |
| Evidence field is overused (too verbose) | Low | Medium | Suggest 1-3 sentence limit; warn on very long evidence |
| Evidence contains sensitive information | High | Low | Code review checks; security scanning; guidelines prohibit |

## Clarifications Resolved

**Clarification 1: Evidence Length Limits**
**Decision:** Soft limit with 500 character warning threshold (Option A)
**Rationale:** Allows flexibility for detailed context when needed while encouraging concise documentation. Warnings guide developers toward best practices without blocking their work. Database supports up to 4KB for edge cases.

**Clarification 2: Evidence Structure**
**Decision:** Structured free-form text with filename:line conventions
**Rationale:** Evidence remains human-readable natural language but follows conventions for referencing code locations (e.g., "at binding.go:145"). This provides structure without requiring formal schemas, making it easy to write and parse. File:line references can be extracted for clickable navigation in CLI/tools.

## Review & Acceptance Checklist

**Content Quality:**
- [x] No implementation details (languages, frameworks, APIs)
- [x] Focused on user value and business needs
- [x] Written for non-technical stakeholders
- [x] All mandatory sections completed

**Requirement Completeness:**
- [x] No [NEEDS CLARIFICATION] markers remain (all resolved)
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

Break down this requirement into specific implementation points. Each point gets its own CANARY token to help agents locate where to implement changes.

### Core Features

<!-- CANARY: REQ=CBIN-143; FEATURE="EvidenceSyntax"; ASPECT=Engine; STATUS=STUB; UPDATED=2025-10-17 -->
**Feature 1: Evidence Token Syntax**
- [ ] Extend CANARY token regex/parser to recognize EVIDENCE field
- **Location hint:** "internal/scanner/token_parser.go"
- **Dependencies:** Token parser

<!-- CANARY: REQ=CBIN-143; FEATURE="MultiLineEvidence"; ASPECT=Engine; STATUS=STUB; UPDATED=2025-10-17 -->
**Feature 2: Multi-Line Evidence Parsing**
- [ ] Implement logic to concatenate evidence across multiple comment lines
- **Location hint:** "internal/scanner/token_parser.go"
- **Dependencies:** Evidence syntax

<!-- CANARY: REQ=CBIN-143; FEATURE="EvidenceStorage"; ASPECT=Storage; STATUS=STUB; UPDATED=2025-10-17 -->
**Feature 3: Evidence Database Storage**
- [ ] Add evidence field to database schema and migration
- **Location hint:** "internal/storage/schema.go", migration files
- **Dependencies:** Database schema

<!-- CANARY: REQ=CBIN-143; FEATURE="EvidenceQuery"; ASPECT=Storage; STATUS=STUB; UPDATED=2025-10-17 -->
**Feature 4: Evidence Query Support**
- [ ] Implement queries to retrieve and filter by evidence
- **Location hint:** "internal/storage/token_repository.go"
- **Dependencies:** Evidence storage

<!-- CANARY: REQ=CBIN-143; FEATURE="EvidenceDisplay"; ASPECT=CLI; STATUS=STUB; UPDATED=2025-10-17 -->
**Feature 5: Evidence Display in CLI**
- [ ] Update CLI commands to show evidence alongside token info
- **Location hint:** "cmd/show.go", "cmd/list.go"
- **Dependencies:** Evidence query

<!-- CANARY: REQ=CBIN-143; FEATURE="EvidenceSearch"; ASPECT=CLI; STATUS=STUB; UPDATED=2025-10-17 -->
**Feature 6: Evidence Search Command**
- [ ] Implement search/grep for evidence content
- **Location hint:** "cmd/search.go" or extend existing grep
- **Dependencies:** Evidence query

### Optional Features

<!-- CANARY: REQ=CBIN-143; FEATURE="EvidenceHistory"; ASPECT=Storage; STATUS=STUB; UPDATED=2025-10-17 -->
**Evidence History Tracking:**
- [ ] Track evidence changes over time
- **Location hint:** "internal/storage/evidence_history.go"
- **Dependencies:** Evidence storage

<!-- CANARY: REQ=CBIN-143; FEATURE="EvidenceValidation"; ASPECT=CLI; STATUS=STUB; UPDATED=2025-10-17 -->
**Evidence Validation:**
- [ ] Warn when implemented features lack evidence
- **Location hint:** "cmd/validate.go"
- **Dependencies:** Evidence query, token status

<!-- CANARY: REQ=CBIN-143; FEATURE="EvidenceLengthWarning"; ASPECT=Engine; STATUS=STUB; UPDATED=2025-10-17 -->
**Evidence Length Warning:**
- [ ] Warn when evidence exceeds 500 character soft limit during scan
- **Location hint:** "internal/scanner/evidence_validator.go"
- **Dependencies:** Evidence parsing

<!-- CANARY: REQ=CBIN-143; FEATURE="FileLineExtraction"; ASPECT=Engine; STATUS=STUB; UPDATED=2025-10-17 -->
**File:Line Reference Extraction:**
- [ ] Extract and parse filename:line references from evidence for clickable navigation
- **Location hint:** "internal/scanner/evidence_parser.go"
- **Dependencies:** Evidence storage

### Testing Requirements

<!-- CANARY: REQ=CBIN-143; FEATURE="EvidenceParsingTests"; ASPECT=Engine; STATUS=STUB; TEST=TestEvidenceParsing; UPDATED=2025-10-17 -->
**Unit Tests:**
- [ ] Test evidence parsing (single-line, multi-line, edge cases)
- **Location hint:** "internal/scanner/token_parser_test.go"

<!-- CANARY: REQ=CBIN-143; FEATURE="EvidenceStorageTests"; ASPECT=Storage; STATUS=STUB; TEST=TestEvidenceStorage; UPDATED=2025-10-17 -->
**Storage Tests:**
- [ ] Test evidence CRUD operations and queries
- **Location hint:** "internal/storage/token_repository_test.go"

<!-- CANARY: REQ=CBIN-143; FEATURE="EvidenceIntegrationTests"; ASPECT=CLI; STATUS=STUB; TEST=TestEvidenceWorkflow; UPDATED=2025-10-17 -->
**Integration Tests:**
- [ ] Test end-to-end: add evidence → scan → query → display
- **Location hint:** "test/integration/evidence_test.go"

### Documentation

<!-- CANARY: REQ=CBIN-143; FEATURE="EvidenceSyntaxDocs"; ASPECT=Docs; STATUS=STUB; UPDATED=2025-10-17 -->
**Syntax Documentation:**
- [ ] Document EVIDENCE field syntax and examples
- **Location hint:** "docs/token-syntax.md", "README.md"

<!-- CANARY: REQ=CBIN-143; FEATURE="EvidenceGuidelines"; ASPECT=Docs; STATUS=STUB; UPDATED=2025-10-17 -->
**Evidence Guidelines:**
- [ ] Create guide for writing effective evidence
- **Location hint:** "docs/guides/writing-evidence.md"

<!-- CANARY: REQ=CBIN-143; FEATURE="EvidenceCLIDocs"; ASPECT=Docs; STATUS=STUB; UPDATED=2025-10-17 -->
**CLI Documentation:**
- [ ] Document evidence-related commands and flags
- **Location hint:** "docs/cli/evidence.md"

---

**Agent Instructions:**

After implementing each feature:
1. Update the CANARY token in the spec from `STATUS=STUB` to `STATUS=IMPL`
2. Add the same token to your source code at the implementation location
3. Add `TEST=TestName` when tests are written
4. Run `canary implement CBIN-143` to see implementation progress

---

## CANARY Tokens Reference

**Main requirement token** (add to primary implementation file):
```
// CANARY: REQ=CBIN-143; FEATURE="EvidenceField"; ASPECT=Storage; STATUS=IMPL; UPDATED=2025-10-17
```

**Example token with structured evidence** (use in actual implementations):
```
// CANARY: REQ=CBIN-143; FEATURE="EvidenceSyntax"; ASPECT=Engine; STATUS=IMPL; UPDATED=2025-10-17;
//         EVIDENCE="Extended regex at token_parser.go:156 to match EVIDENCE field; handles quoted strings with escaped quotes"
```

**Multi-line evidence example with file:line references**:
```
// CANARY: REQ=CBIN-143; FEATURE="MultiLineEvidence"; ASPECT=Engine; STATUS=IMPL; UPDATED=2025-10-17;
//         EVIDENCE="Multi-line evidence parsing at token_parser.go:234:
//                   - Continues until next field or end of comment block
//                   - Strips leading '//' and whitespace (token_parser.go:245)
//                   - Preserves intentional line breaks
//                   - See TestMultiLineEvidence at token_parser_test.go:89"
```

**Evidence with gap analysis reference**:
```
// CANARY: REQ=CBIN-143; FEATURE="EvidenceValidation"; ASPECT=CLI; STATUS=IMPL; UPDATED=2025-10-17;
//         EVIDENCE="Fixed validation logic at validate.go:67 (see GAP-015); now correctly handles missing EVIDENCE field"
```

**Use `canary implement CBIN-143` to find:**
- Which features are implemented vs. still TODO
- Exact file locations and line numbers
- Context around each implementation point
