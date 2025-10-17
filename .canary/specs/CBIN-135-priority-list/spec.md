<!-- CANARY: REQ=CBIN-115; FEATURE="SpecTemplate"; ASPECT=Docs; STATUS=IMPL; OWNER=canary; UPDATED=2025-10-16 -->
# Feature Specification: Priority List Command

**Requirement ID:** CBIN-135
**Status:** STUB
**Created:** 2025-10-16
**Last Updated:** 2025-10-16

## Overview

**Purpose:** Provide developers and AI agents with a quick way to view prioritized work items without loading full specification details. Enable filtering, sorting, and limiting results to focus on the most relevant requirements for the current workflow.

**Scope:**
- Included: CLI `list` subcommand with sort, filter, and count options
- Included: Agent-accessible `/canary.list` slash command
- Included: Priority-based ranking of requirements
- Included: Status, aspect, and owner filtering
- Included: Database-backed queries for performance
- Excluded: Detailed requirement content in list view
- Excluded: Interactive editing from list view
- Excluded: Visualization or graphical representations

## User Stories

### Primary User Stories

**US-1: View Top Priority Work Items**
As a developer,
I want to see a sorted list of top priority requirements,
So that I can quickly identify what to work on next.

**Acceptance Criteria:**
- [ ] Can run `canary list` to see default view (top 10 by priority)
- [ ] List shows requirement ID, feature name, status, priority, and last updated
- [ ] Results display in < 1 second
- [ ] Output is readable in terminal with clear formatting
- [ ] Shows "No requirements found" message when list is empty

**US-2: Filter Requirements by Status**
As a developer,
I want to filter requirements by status (STUB, IMPL, TESTED),
So that I can focus on work at a specific stage.

**Acceptance Criteria:**
- [ ] Can use `--status STUB` to show only STUB requirements
- [ ] Can specify multiple statuses: `--status STUB,IMPL`
- [ ] Invalid status values show clear error message
- [ ] Filter applies before sorting and limiting

**US-3: Sort by Different Criteria**
As a developer,
I want to sort requirements by priority, status, or last updated date,
So that I can view requirements in the order most relevant to my workflow.

**Acceptance Criteria:**
- [ ] Can use `--sort priority` (default), `--sort updated`, `--sort status`
- [ ] Can reverse sort order with `--sort priority --desc`
- [ ] Invalid sort fields show helpful error with valid options
- [ ] Sort is stable (maintains secondary ordering)

**US-4: Limit Result Count**
As an AI agent with limited context,
I want to limit the number of results returned,
So that I don't overflow my context window with irrelevant information.

**Acceptance Criteria:**
- [ ] Can use `--count N` or `-n N` to limit results
- [ ] Default limit is 10 items
- [ ] Can use `--count 0` or `--all` to show unlimited results
- [ ] Shows "Showing X of Y total" when results are limited

### Secondary User Stories

**US-5: Filter by Aspect**
As a developer,
I want to filter requirements by aspect (CLI, API, Storage, etc.),
So that I can focus on work in my area of expertise.

**Acceptance Criteria:**
- [ ] Can use `--aspect CLI` to filter by aspect
- [ ] Can specify multiple aspects: `--aspect CLI,API`
- [ ] Shows all valid aspect values in error message

**US-6: Filter by Owner**
As a team member,
I want to see requirements assigned to specific owners,
So that I can track team responsibilities.

**Acceptance Criteria:**
- [ ] Can use `--owner alice` to filter by owner
- [ ] Shows unassigned requirements when owner field is empty
- [ ] Owner names are case-insensitive

**US-7: Agent Slash Command Access**
As an AI agent,
I want to use `/canary.list` with parameters,
So that I can quickly query work items during development sessions.

**Acceptance Criteria:**
- [ ] `/canary.list` shows top 10 priorities
- [ ] `/canary.list --status STUB --count 5` filters and limits
- [ ] Results format is agent-friendly (parseable, concise)
- [ ] Agent template `.claude/commands/canary.list.md` exists

## Functional Requirements

### FR-1: List Subcommand
**Priority:** High
**Description:** Add `list` subcommand to main CLI that displays requirement summary information
**Acceptance:** Command executes successfully and displays formatted table of requirements

### FR-2: Priority Sorting
**Priority:** High
**Description:** Default sort order is by priority (1=highest), then by status (STUB > IMPL > TESTED > BENCHED)
**Acceptance:** Requirements appear with highest priority first, breaking ties by status progression

### FR-3: Status Filtering
**Priority:** High
**Description:** Support `--status` flag accepting comma-separated status values
**Acceptance:** Only requirements matching specified statuses are displayed

### FR-4: Result Limiting
**Priority:** High
**Description:** Support `--count N` flag to limit number of results, default to 10
**Acceptance:** Exactly N results shown (or fewer if less than N match filters)

### FR-5: Aspect Filtering
**Priority:** Medium
**Description:** Support `--aspect` flag to filter by requirement aspect
**Acceptance:** Only requirements with matching aspect are displayed

### FR-6: Custom Sorting
**Priority:** Medium
**Description:** Support `--sort field` flag with options: priority, updated, status, aspect
**Acceptance:** Results sorted by specified field in ascending order (or descending with `--desc`)

### FR-7: Owner Filtering
**Priority:** Medium
**Description:** Support `--owner` flag to filter by owner name
**Acceptance:** Only requirements assigned to specified owner are displayed

### FR-8: Database Query Optimization
**Priority:** Medium
**Description:** Use database queries when available, fall back to filesystem scan
**Acceptance:** Queries complete in < 100ms with database, < 1s without

### FR-9: Agent Slash Command
**Priority:** High
**Description:** Create `.claude/commands/canary.list.md` template with usage guidance
**Acceptance:** Agents can invoke `/canary.list` and receive formatted results

### FR-10: Output Formatting
**Priority:** High
**Description:** Display results as readable table with columns: ID, Feature, Status, Priority, Aspect, Updated
**Acceptance:** Table aligns columns, truncates long feature names, shows concise data

## Success Criteria

**Quantitative Metrics:**
- [ ] List command completes in < 1 second without database
- [ ] Database queries complete in < 100ms
- [ ] Default view (10 items) uses < 2000 tokens for agent context
- [ ] Filtered views reduce results by 60-90% on average
- [ ] 90% of user queries require only 1 command invocation

**Qualitative Measures:**
- [ ] Developers can identify next work item in < 5 seconds
- [ ] Agents successfully use list output to guide development
- [ ] Output is readable without horizontal scrolling on 80-column terminal
- [ ] Error messages are helpful and suggest corrections

## User Scenarios & Testing

### Scenario 1: View Default Priority List (Happy Path)
**Given:** User has initialized CANARY project with 25 requirements
**When:** They run `canary list`
**Then:** System displays top 10 requirements sorted by priority, showing ID, feature, status, priority, aspect, and last updated

### Scenario 2: Filter STUB Requirements for New Work
**Given:** User wants to find new work to implement
**When:** They run `canary list --status STUB --count 5`
**Then:** System displays 5 highest priority STUB requirements

### Scenario 3: Find Stale Requirements Needing Updates
**Given:** User wants to identify outdated requirements
**When:** They run `canary list --sort updated --count 20`
**Then:** System displays 20 requirements sorted by last updated (oldest first)

### Scenario 4: View All CLI-Related Work
**Given:** Developer focuses on CLI features
**When:** They run `canary list --aspect CLI --all`
**Then:** System displays all CLI requirements without limit

### Scenario 5: Agent Queries Top Priorities (Agent Workflow)
**Given:** AI agent starts development session
**When:** Agent runs `/canary.list --count 3`
**Then:** System returns top 3 priorities in parseable format, using < 500 tokens

### Scenario 6: Empty Results with Helpful Message
**Given:** User filters for requirements that don't exist
**When:** They run `canary list --status REMOVED --owner nobody`
**Then:** System displays "No requirements found matching filters" with suggestion to adjust filters

### Scenario 7: Invalid Filter Value (Error Case)
**Given:** User provides invalid status value
**When:** They run `canary list --status INVALID`
**Then:** System shows error: "Invalid status 'INVALID'. Valid values: STUB, IMPL, TESTED, BENCHED, REMOVED"

### Scenario 8: Combine Multiple Filters
**Given:** User wants specific subset of requirements
**When:** They run `canary list --status STUB,IMPL --aspect CLI --owner alice --count 10`
**Then:** System applies all filters and shows up to 10 matching requirements

## Key Entities

### Entity 1: RequirementSummary
**Attributes:**
- req_id: Requirement ID (CBIN-XXX)
- feature: Feature name (shortened if > 40 chars)
- status: Current status (STUB, IMPL, TESTED, BENCHED, REMOVED)
- priority: Priority number (1=highest)
- aspect: Requirement aspect (CLI, API, Storage, etc.)
- owner: Assigned owner name
- updated: Last updated timestamp (YYYY-MM-DD)

**Relationships:**
- Derived from Token entity in database
- Maps to spec.md files in .canary/specs/ directory

### Entity 2: ListQuery
**Attributes:**
- status_filter: List of status values to include
- aspect_filter: List of aspects to include
- owner_filter: Owner name to match
- sort_field: Field to sort by (priority, updated, status)
- sort_desc: Boolean for descending order
- limit: Maximum results to return

**Relationships:**
- Used to construct database query or filesystem filter
- Produces RequirementSummary results

## Assumptions

- `.canary/canary.db` may or may not exist (filesystem fallback required)
- Requirement IDs follow CBIN-XXX pattern
- CANARY tokens in codebase contain all necessary metadata
- Terminal width is at least 80 characters for readable output
- Users understand CANARY status lifecycle (STUB → IMPL → TESTED → BENCHED)

## Constraints

**Technical Constraints:**
- Must work without database (filesystem scan as fallback)
- Table output must fit in 80-column terminal
- Feature names truncated to 40 characters with ellipsis
- Default limit of 10 to prevent context overflow for agents
- Case-insensitive matching for owner and aspect filters

**Business Constraints:**
- Reuse existing database schema and query patterns
- Integrate with current CLI command structure
- No external dependencies beyond current stack
- Output format should be consistent with `canary implement` style

**Regulatory Constraints:**
- None

## Out of Scope

- Interactive selection or editing from list view (use `canary implement` or `canary specify update`)
- Graphical visualization or charts
- Export to formats other than terminal output (JSON, CSV, HTML)
- Filtering by test names or bench names
- Search across specification file contents (use `canary specify update --search`)
- Grouping or aggregation statistics
- Saved filter presets or aliases
- Real-time updates or watch mode

## Dependencies

- CBIN-123: TokenStorage (use database for fast queries)
- Existing CLI command infrastructure
- Database schema with tokens table

## Risks & Mitigation

| Risk | Impact | Probability | Mitigation |
|------|--------|-------------|------------|
| Large codebases slow filesystem scan | Medium | Medium | Implement database-first strategy, cache results, optimize file reading |
| Table formatting breaks on narrow terminals | Low | Low | Use responsive formatting, abbreviate columns when width < 80 |
| Filter combinations return zero results | Low | High | Show helpful message suggesting less restrictive filters |
| Sort stability inconsistent across runs | Low | Low | Use stable sort with secondary keys (priority -> status -> req_id) |

## Clarifications Needed

[NEEDS CLARIFICATION: Should list output include dependency information?]
**Options:**
A) No - keep output simple and focused
B) Yes - add column showing DEPENDS_ON count
C) Optional flag `--show-deps` to include dependency info
**Impact:** Option C provides flexibility but adds complexity. Option A keeps output clean and fast.

[NEEDS CLARIFICATION: What should default behavior be when database is not available?]
**Options:**
A) Show error and require database
B) Silently fall back to filesystem scan
C) Show warning then proceed with filesystem scan
**Impact:** Option C provides best user experience with transparency

## Review & Acceptance Checklist

**Content Quality:**
- [x] No implementation details (languages, frameworks, APIs)
- [x] Focused on user value and business needs
- [x] Written for non-technical stakeholders
- [x] All mandatory sections completed

**Requirement Completeness:**
- [x] Only 2 [NEEDS CLARIFICATION] markers remaining
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

<!-- CANARY: REQ=CBIN-135; FEATURE="ListSubcommand"; ASPECT=CLI; STATUS=TESTED; TEST=TestCANARY_CBIN_135_CLI_ListCommand; UPDATED=2025-10-17 -->
**Feature 1: List Subcommand**
- [x] Add `list` subcommand to main CLI
- [x] Parse flags: --status, --aspect, --owner, --sort, --count, --desc, --all
- [x] Call query engine with parsed filters
- [x] Format and display results table
- **Location:** `cmd/canary/main.go:1544` (CBIN-125 listCmd)
- **Tests:** `cmd/canary/list_test.go`, `cmd/canary/list_integration_test.go`
- **Dependencies:** None

<!-- CANARY: REQ=CBIN-135; FEATURE="QueryEngine"; ASPECT=Engine; STATUS=TESTED; TEST=TestCANARY_CBIN_135_CLI_ListCommand_Sorting; UPDATED=2025-10-17 -->
**Feature 2: Query Engine**
- [x] Implement filter logic (status, aspect, owner)
- [x] Implement sort logic (priority, updated, status, aspect)
- [x] Apply limit and return RequirementSummary list
- [x] Use database when available, fall back to filesystem
- **Location:** `internal/storage/storage.go:238` (ListTokens method, CBIN-145)
- **Tests:** `cmd/canary/list_test.go` (sorting, filtering, performance tests)
- **Dependencies:** CBIN-123 TokenStorage

<!-- CANARY: REQ=CBIN-135; FEATURE="TableFormatter"; ASPECT=CLI; STATUS=TESTED; TEST=TestCANARY_CBIN_135_CLI_ListCommand; UPDATED=2025-10-17 -->
**Feature 3: Table Formatter**
- [x] Format RequirementSummary as aligned table
- [x] Truncate long feature names to 40 chars
- [x] Handle empty results with helpful message
- [x] Show "Showing X of Y total" when limited
- **Location:** `cmd/canary/main.go:1559-1650` (listCmd RunE function)
- **Tests:** `cmd/canary/list_test.go` (integrated in ListCommand tests)
- **Dependencies:** None

<!-- CANARY: REQ=CBIN-135; FEATURE="FilesystemFallback"; ASPECT=Engine; STATUS=REMOVED; UPDATED=2025-10-17 -->
**Feature 4: Filesystem Fallback**
- [x] ~~Scan .canary/specs/ directory when DB unavailable~~
- [x] ~~Parse spec.md files for requirement metadata~~
- [x] ~~Apply filters and sorting on parsed data~~
- [x] ~~Optimize for performance (parallel reads, early termination)~~
- **Status:** REMOVED - Implementation is database-only (CBIN-125), no filesystem fallback needed
- **Rationale:** Database provides fast queries; `canary index` creates DB from scan
- **Dependencies:** None

### Agent Integration

<!-- CANARY: REQ=CBIN-135; FEATURE="AgentSlashCommand"; ASPECT=Docs; STATUS=IMPL; UPDATED=2025-10-17 -->
**Agent Slash Command:**
- [x] Create `.claude/commands/canary.list.md` template
- [x] Document usage patterns for agents
- [x] Provide examples of common queries
- [x] Explain output format for parsing
- **Location:** `.claude/commands/canary.list.md:5`
- **Dependencies:** None

### Testing Requirements

<!-- CANARY: REQ=CBIN-135; FEATURE="UnitTests"; ASPECT=CLI; STATUS=TESTED; TEST=TestCANARY_CBIN_135_CLI_ListCommand; UPDATED=2025-10-17 -->
**Unit Tests:**
- [x] Test filter logic (status, aspect, owner)
- [x] Test sort logic (all sort fields, ascending/descending)
- [x] Test limit and default behaviors
- [x] Test table formatting edge cases
- **Location:** `cmd/canary/list_test.go` (4 test functions, 34 subtests)

<!-- CANARY: REQ=CBIN-135; FEATURE="IntegrationTests"; ASPECT=CLI; STATUS=TESTED; TEST=TestCANARY_CBIN_135_Integration_EndToEnd; UPDATED=2025-10-17 -->
**Integration Tests:**
- [x] Test end-to-end list workflow with real database
- [x] Test JSON output format for programmatic parsing
- [x] Test agent workflow (context-constrained queries)
- [x] Test combined filters and edge cases
- [x] Test performance with large requirement sets (200+ tokens)
- **Location:** `cmd/canary/list_integration_test.go` (5 test functions, all passing)

### Documentation

<!-- CANARY: REQ=CBIN-135; FEATURE="CLIDocs"; ASPECT=Docs; STATUS=IMPL; UPDATED=2025-10-17 -->
**CLI Documentation:**
- [x] Add `canary list --help` documentation
- [x] Document all flags and usage examples
- [ ] Update README with list command section (TODO)
- **Location:** `cmd/canary/main.go:1546-1558` (listCmd help text), `.claude/commands/canary.list.md`

---

**Agent Instructions:**

After implementing each feature:
1. Update the CANARY token in the spec from `STATUS=STUB` to `STATUS=IMPL`
2. Add the same token to your source code at the implementation location
3. Add `TEST=TestName` when tests are written
4. Run `canary implement CBIN-135` to see implementation progress

---

## CANARY Tokens Reference

**Main requirement token** (add to primary implementation file):
```go
// CANARY: REQ=CBIN-135; FEATURE="PriorityList"; ASPECT=CLI; STATUS=IMPL; UPDATED=2025-10-16
```

**Sub-feature tokens** (use the specific feature names from Implementation Checklist):
```go
// CANARY: REQ=CBIN-135; FEATURE="ListSubcommand"; ASPECT=CLI; STATUS=IMPL; TEST=TestListSubcommand; UPDATED=2025-10-16
```

**Use `canary implement CBIN-135` to find:**
- Which features are implemented vs. still TODO
- Exact file locations and line numbers
- Context around each implementation point
