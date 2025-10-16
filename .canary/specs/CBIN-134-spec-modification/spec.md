# CANARY: REQ=CBIN-115; FEATURE="SpecTemplate"; ASPECT=Docs; STATUS=IMPL; OWNER=canary; UPDATED=2025-10-16
# Feature Specification: Specification Modification Command

**Requirement ID:** CBIN-134
**Status:** STUB
**Created:** 2025-10-16
**Last Updated:** 2025-10-16

## Overview

**Purpose:** Enable AI agents and developers to efficiently locate and modify existing specifications without loading excessive context. The current workflow requires agents to scan through all specs to find the right one, consuming valuable context window space. This feature provides targeted lookup capabilities for spec modification.

**Scope:**
- Included: Subcommands for modifying existing specs via exact ID lookup, fuzzy text search, and keyword matching
- Included: Modification of both spec.md and plan.md files
- Included: CLI flags to minimize context usage during lookup
- Included: Integration with `/canary.specify` template for agent workflows
- Excluded: Bulk modification of multiple specs simultaneously
- Excluded: Automated spec merging or conflict resolution
- Excluded: Version control/history tracking (handled by git)

## User Stories

### Primary User Stories

**US-1: Modify Spec by Exact ID**
As an AI agent,
I want to update an existing specification using its exact requirement ID,
So that I can make targeted changes without loading all specs into context.

**Acceptance Criteria:**
- [ ] Can call `canary specify update CBIN-XXX` with exact requirement ID
- [ ] System locates spec file in < 1 second
- [ ] Returns spec content for modification
- [ ] Updates both spec.md and plan.md if plan exists
- [ ] Returns error if requirement ID not found

**US-2: Find Spec by Text Search**
As an AI agent,
I want to search for specifications by feature name or keywords,
So that I can locate the right spec when I don't know the exact ID.

**Acceptance Criteria:**
- [ ] Can use `--search` flag with fuzzy text matching
- [ ] Returns top 5 matching specifications ranked by relevance
- [ ] Shows requirement ID, feature name, and match score for each result
- [ ] Allows selection from search results
- [ ] Handles partial matches and typos gracefully

**US-3: Modify Spec with Minimal Context**
As an AI agent with limited context window,
I want to modify only specific sections of a spec,
So that I don't need to reload the entire specification file.

**Acceptance Criteria:**
- [ ] Can specify which sections to load (e.g., `--sections overview,requirements`)
- [ ] Returns only requested sections plus metadata
- [ ] Preserves other sections unchanged during update
- [ ] Validates section names and provides helpful errors

### Secondary User Stories

**US-4: List Modification Candidates**
As a developer,
I want to see which specs are marked as needing updates,
So that I can prioritize specification maintenance work.

**Acceptance Criteria:**
- [ ] Can use `--list-stale` to show specs with outdated UPDATED fields
- [ ] Shows specs with [NEEDS CLARIFICATION] markers
- [ ] Displays specs in STATUS=STUB or STATUS=IMPL state

## Functional Requirements

### FR-1: Update Subcommand
**Priority:** High
**Description:** Add `update` or `modify` subcommand to `canary specify` that accepts requirement ID or search criteria
**Acceptance:** Command successfully locates and opens spec for editing using provided lookup method

### FR-2: Exact ID Lookup
**Priority:** High
**Description:** Support `canary specify update CBIN-XXX` to directly load spec by requirement ID
**Acceptance:** Spec located in < 1 second, returns file path and opens in editor/returns content

### FR-3: Fuzzy Text Search
**Priority:** High
**Description:** Implement `--search "feature keywords"` flag that performs fuzzy matching across spec files
**Acceptance:** Returns ranked list of matching specs with scores, handles typos and partial matches

### FR-4: Section-Specific Loading
**Priority:** Medium
**Description:** Support `--sections` flag to load only specific sections of the specification
**Acceptance:** Returns only requested sections, reduces context usage by 50-80% for targeted updates

### FR-5: Plan File Updates
**Priority:** Medium
**Description:** When modifying a spec, also update corresponding plan.md file if it exists
**Acceptance:** Both spec.md and plan.md updated when changes affect implementation details

### FR-6: Database-Backed Search
**Priority:** Medium
**Description:** Use `.canary/canary.db` for fast lookups when database is available
**Acceptance:** Search queries complete in < 100ms using database index, falls back to filesystem if DB unavailable

### FR-7: Agent Template Integration
**Priority:** High
**Description:** Update `/canary.specify` template to include modification workflow instructions
**Acceptance:** Template provides clear guidance on using update vs. create workflows

## Success Criteria

**Quantitative Metrics:**
- [ ] Spec lookup completes in < 1 second for exact ID matches
- [ ] Fuzzy search returns results in < 2 seconds
- [ ] Section-specific loading reduces context usage by 50-80%
- [ ] Database-backed searches complete in < 100ms
- [ ] 95% of spec modifications use < 5000 tokens of context

**Qualitative Measures:**
- [ ] AI agents can locate correct spec in first attempt
- [ ] Developers find fuzzy search intuitive and helpful
- [ ] Spec modification workflow feels faster than full reload

## User Scenarios & Testing

### Scenario 1: Update Spec by Exact ID (Happy Path)
**Given:** Developer knows the requirement ID is CBIN-105
**When:** They run `canary specify update CBIN-105`
**Then:** System loads .canary/specs/CBIN-105-*/spec.md and opens for editing

### Scenario 2: Find Spec by Feature Name (Fuzzy Search)
**Given:** Developer remembers feature is about "authentication"
**When:** They run `canary specify update --search "auth"`
**Then:** System returns ranked list: 1) CBIN-107 UserAuthentication (95%), 2) CBIN-109 TokenAuth (78%), etc.

### Scenario 3: Update Specific Section Only
**Given:** Agent needs to update success criteria only
**When:** They run `canary specify update CBIN-105 --sections success-criteria`
**Then:** System returns only the success criteria section, preserving < 2000 tokens of context

### Scenario 4: Spec Not Found (Error Case)
**Given:** User searches for non-existent requirement
**When:** They run `canary specify update CBIN-999`
**Then:** System returns error "Requirement CBIN-999 not found" with suggestion to use `--search`

### Scenario 5: Multiple Match Disambiguation
**Given:** Search term matches multiple specs equally
**When:** User runs `canary specify update --search "data validation"`
**Then:** System presents numbered list of matches and prompts for selection (1-5)

### Scenario 6: Update Both Spec and Plan
**Given:** Spec CBIN-105 has both spec.md and plan.md files
**When:** User modifies functional requirements affecting implementation
**Then:** System updates both files and reports which files were modified

## Key Entities

### Entity 1: SpecificationLookup
**Attributes:**
- req_id: Unique requirement identifier (CBIN-XXX)
- spec_path: File system path to spec.md
- plan_path: Optional path to plan.md
- feature_name: Human-readable feature name
- status: Current requirement status
- last_updated: Timestamp of last modification

**Relationships:**
- Related to DatabaseToken entries for fast lookup
- Linked to filesystem directory structure

### Entity 2: SearchResult
**Attributes:**
- req_id: Matching requirement ID
- feature_name: Feature name text
- match_score: Relevance score (0-100)
- matched_text: Snippet showing why it matched
- spec_path: Path to specification file

**Relationships:**
- Derived from SpecificationLookup via search query

## Assumptions

- `.canary/specs/` directory follows naming convention: `CBIN-XXX-feature-name/`
- Requirement IDs follow `CBIN-XXX` pattern (3-digit zero-padded)
- Database `.canary/canary.db` may or may not be available (graceful fallback required)
- Specs are UTF-8 encoded markdown files
- Users have basic familiarity with command-line interfaces

## Constraints

**Technical Constraints:**
- Must work without database (filesystem-only fallback)
- Search results limited to top 5 matches to avoid context overflow
- Section names must match predefined list for validation
- File paths must be absolute or relative to project root

**Business Constraints:**
- Should reuse existing fuzzy matching logic from `canary implement`
- Must integrate cleanly with current `canary specify` command structure
- No external dependencies beyond current project stack

**Regulatory Constraints:**
- None

## Out of Scope

- Bulk modification of multiple specifications in single command
- Automated conflict resolution when multiple agents modify same spec
- Version control history tracking (use git for this)
- Graphical user interface or web-based editor
- Real-time collaboration features
- Spec validation beyond basic format checking
- Migration of specs from old to new format
- Cross-project spec sharing or templates

## Dependencies

- CBIN-133: FuzzyMatching (reuse existing Levenshtein distance implementation)
- CBIN-123: TokenStorage (use database for fast lookups when available)
- Existing `.canary/specs/` directory structure
- `canary specify` command infrastructure

## Risks & Mitigation

| Risk | Impact | Probability | Mitigation |
|------|--------|-------------|------------|
| Fuzzy search returns too many false positives | Medium | Medium | Tune relevance threshold, limit to top 5 results, show match scores |
| Database becomes required dependency | High | Low | Implement robust filesystem fallback, make DB optional |
| Section-specific loading breaks spec integrity | High | Low | Validate section names, preserve unmodified sections, add integrity checks |
| Spec and plan files become out of sync | Medium | Medium | Always update both files when modifications affect both, warn on discrepancies |

## Clarifications Needed

[NEEDS CLARIFICATION: Should the command support interactive selection UI or pure CLI arguments?]
**Options:**
A) Interactive numbered menu when multiple matches found (like implement command)
B) Pure CLI with `--select N` flag for programmatic use
C) Both modes with `--interactive` flag to choose behavior
**Impact:** Option C provides best flexibility for both human and agent users, but adds complexity

## Review & Acceptance Checklist

**Content Quality:**
- [x] No implementation details (languages, frameworks, APIs)
- [x] Focused on user value and business needs
- [x] Written for non-technical stakeholders
- [x] All mandatory sections completed

**Requirement Completeness:**
- [x] Only 1 [NEEDS CLARIFICATION] marker remaining
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

<!-- CANARY: REQ=CBIN-134; FEATURE="UpdateSubcommand"; ASPECT=CLI; STATUS=IMPL; TEST=TestCANARY_CBIN_134_CLI_UpdateSubcommand; UPDATED=2025-10-16 -->
**Feature 1: Update Subcommand**
- [ ] Add `update` or `modify` subcommand to specify command
- [ ] Parse requirement ID from positional argument
- [ ] Support both `canary specify update` and `canary specify modify` aliases
- **Location hint:** `cmd/canary/main.go` near existing specifyCmd
- **Dependencies:** None

<!-- CANARY: REQ=CBIN-134; FEATURE="ExactIDLookup"; ASPECT=Engine; STATUS=IMPL; TEST=TestCANARY_CBIN_134_Engine_ExactIDLookup; UPDATED=2025-10-16 -->
**Feature 2: Exact ID Lookup**
- [ ] Glob for .canary/specs/CBIN-XXX-*/spec.md pattern
- [ ] Return spec path or error if not found
- [ ] Validate requirement ID format (CBIN-\d{3})
- **Location hint:** `internal/matcher/` or new `internal/specs/lookup.go`
- **Dependencies:** None

<!-- CANARY: REQ=CBIN-134; FEATURE="FuzzySpecSearch"; ASPECT=Engine; STATUS=IMPL; TEST=TestCANARY_CBIN_134_Engine_FuzzySpecSearch; UPDATED=2025-10-16 -->
**Feature 3: Fuzzy Spec Search**
- [ ] Implement --search flag with fuzzy matching
- [ ] Reuse Levenshtein distance from CBIN-133
- [ ] Return top 5 ranked results with scores
- [ ] Search across feature names and requirement IDs
- **Location hint:** `internal/matcher/fuzzy.go` (extend existing)
- **Dependencies:** CBIN-133 FuzzyMatching

<!-- CANARY: REQ=CBIN-134; FEATURE="SectionLoader"; ASPECT=Engine; STATUS=IMPL; TEST=TestCANARY_CBIN_134_Engine_SectionParser; UPDATED=2025-10-16 -->
**Feature 4: Section-Specific Loading**
- [ ] Implement --sections flag parsing
- [ ] Parse markdown sections by ## headers
- [ ] Extract and return only requested sections
- [ ] Preserve section ordering and formatting
- **Location hint:** `internal/specs/parser.go` (new file)
- **Dependencies:** None

### Data Layer

<!-- CANARY: REQ=CBIN-134; FEATURE="DatabaseLookup"; ASPECT=Storage; STATUS=IMPL; UPDATED=2025-10-16 -->
**Database Lookup:**
- [ ] Query tokens table for fast spec lookups
- [ ] Join with file paths from token data
- [ ] Fall back to filesystem if DB unavailable
- **Location hint:** `internal/storage/storage.go` (extend existing)
- **Dependencies:** CBIN-123 TokenStorage

### Testing Requirements

<!-- CANARY: REQ=CBIN-134; FEATURE="UnitTests"; ASPECT=CLI; STATUS=STUB; TEST=TestCANARY_CBIN_134_CLI; UPDATED=2025-10-16 -->
**Unit Tests:**
- [ ] Test exact ID lookup with valid and invalid IDs
- [ ] Test fuzzy search ranking and scoring
- [ ] Test section parsing and extraction
- [ ] Test database fallback behavior
- **Location hint:** `cmd/canary/*_test.go` and `internal/matcher/*_test.go`

<!-- CANARY: REQ=CBIN-134; FEATURE="IntegrationTests"; ASPECT=CLI; STATUS=STUB; TEST=TestCANARY_CBIN_134_Integration; UPDATED=2025-10-16 -->
**Integration Tests:**
- [ ] Test end-to-end update workflow
- [ ] Test multi-match disambiguation
- [ ] Test plan.md update when spec.md modified
- **Location hint:** `cmd/canary/*_test.go`

### Documentation

<!-- CANARY: REQ=CBIN-134; FEATURE="TemplateUpdate"; ASPECT=Docs; STATUS=IMPL; UPDATED=2025-10-16 -->
**Template Update:**
- [ ] Update `/canary.specify` template with modification workflow
- [ ] Document update vs. create command decision tree
- [ ] Provide examples of search and exact ID usage
- **Location hint:** `.claude/commands/canary.specify.md`

<!-- CANARY: REQ=CBIN-134; FEATURE="CLIDocs"; ASPECT=Docs; STATUS=IMPL; UPDATED=2025-10-16 -->
**CLI Documentation:**
- [ ] Add `canary specify update --help` documentation
- [ ] Document all flags and usage examples
- [ ] Update README with modification workflow
- **Location hint:** `cmd/canary/main.go` (cobra command help), `README.md`

---

**Agent Instructions:**

After implementing each feature:
1. Update the CANARY token in the spec from `STATUS=STUB` to `STATUS=IMPL`
2. Add the same token to your source code at the implementation location
3. Add `TEST=TestName` when tests are written
4. Run `canary implement CBIN-134` to see implementation progress

---

## CANARY Tokens Reference

**Main requirement token** (add to primary implementation file):
```go
// CANARY: REQ=CBIN-134; FEATURE="SpecModification"; ASPECT=CLI; STATUS=IMPL; UPDATED=2025-10-16
```

**Sub-feature tokens** (use the specific feature names from Implementation Checklist):
```go
// CANARY: REQ=CBIN-134; FEATURE="UpdateSubcommand"; ASPECT=CLI; STATUS=IMPL; TEST=TestUpdateSubcommand; UPDATED=2025-10-16
```

**Use `canary implement CBIN-134` to find:**
- Which features are implemented vs. still TODO
- Exact file locations and line numbers
- Context around each implementation point
