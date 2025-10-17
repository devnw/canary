# CANARY: REQ=CBIN-CLI-001; FEATURE="TokenQueryCommands"; ASPECT=CLI; STATUS=TESTED; OWNER=canary; DOC=user:docs/user/query-commands-guide.md; DOC_HASH=8eb4be8db164a0cf; UPDATED=2025-10-17
# Feature Specification: Token Query Commands

**Requirement ID:** CBIN-CLI-001
**Feature Name:** TokenQueryCommands
**Status:** STUB
**Owner:** canary
**Created:** 2025-10-16
**Updated:** 2025-10-16

---

## Overview

**Purpose:** Replace manual grep and sqlite3 usage with dedicated canary CLI commands for querying and analyzing CANARY tokens. AI agents and developers should be able to inspect tokens, check implementation status, find files, and analyze progress without resorting to low-level shell commands.

**Scope:** Add high-level query commands (`show`, `tokens`, `files`, `status`, `grep`) that provide formatted output for token analysis. Update agent templates and documentation to guide agents toward using these commands instead of manual bash/SQL queries.

## User Stories

### Primary User Stories

**US-1: Agent Shows Tokens for Requirement**
As an AI coding agent,
I want to see all CANARY tokens for a specific requirement ID,
So that I can verify implementation status, find file locations, and check test coverage without using grep or sqlite3.

**Acceptance Criteria:**
- [ ] Agent runs `canary show CBIN-133`
- [ ] System displays all tokens for CBIN-133 with: feature name, aspect, status, file path, line number, test name (if present)
- [ ] Output is formatted in human-readable table or list
- [ ] Exit code 0 if tokens found, 1 if requirement doesn't exist

**US-2: Agent Lists Implementation Files**
As an AI coding agent,
I want to see which files contain implementations for a requirement,
So that I know exactly where to make changes without grepping the codebase.

**Acceptance Criteria:**
- [ ] Agent runs `canary files CBIN-133`
- [ ] System displays unique list of implementation files (*.go, *.md, etc.)
- [ ] Files are grouped by aspect (CLI, API, Engine, etc.)
- [ ] Each file shows count of tokens it contains
- [ ] Output excludes template/plan files (only actual implementation)

**US-3: Agent Checks Requirement Status**
As an AI coding agent,
I want to check overall implementation progress for a requirement,
So that I can understand how much work remains and what status levels exist.

**Acceptance Criteria:**
- [ ] Agent runs `canary status CBIN-133`
- [ ] System displays summary: total tokens, count by status (STUB/IMPL/TESTED/BENCHED)
- [ ] System shows completion percentage
- [ ] System lists any STUB or IMPL tokens that need work
- [ ] Output includes links to files needing updates

**US-4: Developer Searches Tokens by Pattern**
As a developer,
I want to search for tokens matching a pattern (feature name, aspect, owner),
So that I can find related implementations across the codebase.

**Acceptance Criteria:**
- [ ] Developer runs `canary grep "FuzzyMatch"`
- [ ] System searches token database for matches in: req_id, feature, aspect, owner, keywords
- [ ] Results show matching tokens with context
- [ ] Case-insensitive search by default
- [ ] Supports regex patterns with `--regex` flag

### Secondary User Stories

**US-5: Agent Updates Templates with Proper Commands**
As a template maintainer,
I want agent templates to reference `canary show/files/status` commands,
So that agents stop using raw grep/sqlite3 and use proper abstractions.

## Functional Requirements

### FR-1: Show Command (canary show <REQ-ID>)
**Priority:** High
**Description:** Display all CANARY tokens for a specific requirement ID with detailed information including feature name, aspect, status, file location, line number, test references, and owner.
**Acceptance:**
- Command accepts requirement ID (e.g., CBIN-133)
- Queries database for all tokens matching req_id
- Displays formatted output with all token fields
- Groups tokens by aspect or status (configurable with --group-by flag)
- Supports --json flag for machine-readable output
- Returns exit code 0 if found, 1 if not found

### FR-2: Files Command (canary files <REQ-ID>)
**Priority:** High
**Description:** List all implementation files containing tokens for a requirement, grouped by aspect, excluding template and specification files.
**Acceptance:**
- Command accepts requirement ID
- Queries database and extracts unique file paths
- Filters out .canary/specs/, .canary/templates/, base/ directories
- Groups files by aspect (CLI, API, Engine, etc.)
- Shows token count per file
- Supports --all flag to include spec/template files

### FR-3: Status Command (canary status <REQ-ID>)
**Priority:** High
**Description:** Display implementation progress summary for a requirement including total tokens, breakdown by status, completion percentage, and list of incomplete work.
**Acceptance:**
- Command accepts requirement ID
- Calculates total, STUB, IMPL, TESTED, BENCHED counts
- Computes completion % (TESTED+BENCHED / total)
- Lists STUB and IMPL tokens that need work
- Shows file locations for incomplete items
- Colored output (green=complete, yellow=in-progress, red=stub)

### FR-4: Grep Command (canary grep <pattern>)
**Priority:** Medium
**Description:** Search tokens by pattern matching across req_id, feature, aspect, owner, and keywords fields.
**Acceptance:**
- Accepts search pattern as argument
- Case-insensitive search by default (--case-sensitive flag)
- Searches across multiple fields
- Supports regex with --regex flag
- Returns matching tokens with context
- Highlights matched portions of text

### FR-5: Database Query Abstraction
**Priority:** High
**Description:** All commands must use database when available, fall back to filesystem grep if database doesn't exist, and provide helpful error messages.
**Acceptance:**
- Commands check for .canary/canary.db existence
- If DB exists, use SQL queries for fast retrieval
- If DB missing, suggest running `canary index`
- Filesystem fallback uses optimized grep patterns
- Performance: DB queries <100ms, filesystem <1s for <1000 files

## Success Criteria

**Quantitative Metrics:**
- [ ] 95% of token queries complete in <100ms using database
- [ ] 90% of agents use `canary show/files/status` instead of grep/sqlite3 (measured by documentation updates)
- [ ] Zero instances of grep/sqlite3 in updated agent templates
- [ ] Command output fits within 80-column terminal width
- [ ] Database queries handle 10,000+ tokens without performance degradation

**Qualitative Measures:**
- [ ] Agents can find implementation files without asking for help
- [ ] Output is immediately understandable without documentation
- [ ] Error messages guide users to correct commands
- [ ] Templates demonstrate proper command usage with examples
- [ ] Agents prefer canary commands over bash alternatives

## User Scenarios & Testing

### Scenario 1: Agent Shows Tokens (Happy Path)
**Given:** Database contains 4 tokens for CBIN-133
**When:** Agent runs `canary show CBIN-133`
**Then:**
- System displays 4 tokens in table format
- Each token shows: feature name, aspect, status, file path:line, test name
- Output grouped by aspect (CLI, API, Engine)
- Exit code 0

### Scenario 2: Agent Checks Files (Happy Path)
**Given:** CBIN-133 has tokens in 3 files (main.go, implement.go, fuzzy.go)
**When:** Agent runs `canary files CBIN-133`
**Then:**
- System displays 3 files grouped by aspect
- Shows: CLI: cmd/canary/main.go (1 token), API: cmd/canary/implement.go (2 tokens), Engine: internal/matcher/fuzzy.go (1 token)
- Excludes spec.md and plan.md files
- Exit code 0

### Scenario 3: Agent Checks Status (Happy Path)
**Given:** CBIN-133 has 4 TESTED tokens
**When:** Agent runs `canary status CBIN-133`
**Then:**
- Displays: Total: 4, TESTED: 4, Completion: 100%
- Shows green checkmarks for completed features
- No incomplete work listed
- Exit code 0

### Scenario 4: Database Missing (Fallback)
**Given:** No .canary/canary.db file exists
**When:** Agent runs `canary show CBIN-133`
**Then:**
- System displays warning: "Database not found, using filesystem search (slower)"
- Suggests running: `canary index` to build database
- Falls back to grep-based search
- Still displays results if tokens found
- Exit code 0 if found, 1 if not

### Scenario 5: Requirement Not Found (Error Case)
**Given:** No tokens exist for CBIN-999
**When:** Agent runs `canary show CBIN-999`
**Then:**
- Displays: "No tokens found for CBIN-999"
- Suggests: "Run `canary list` to see all requirements"
- Exit code 1

## Assumptions

- Database exists at .canary/canary.db (or commands fall back gracefully)
- CANARY tokens follow standard format
- Agents have access to terminal for formatted output
- Color output supported in terminal (with --no-color fallback)

## Constraints

**Technical Constraints:**
- Must work with existing database schema
- Must support both database and filesystem modes
- Performance: queries must complete in <100ms (DB) or <1s (filesystem)

**Business Constraints:**
- No external dependencies (use Go standard library)
- Must integrate with existing canary CLI architecture

## Out of Scope

- Modifying tokens (use existing update commands)
- Creating new requirements (use `canary specify`)
- Visual dashboards or web UI
- Real-time token monitoring
- Token analytics beyond basic counts

## Dependencies

- CBIN-124: IndexCmd (database must exist or fallback works)
- CBIN-125: ListCmd (shares some query logic)
- Existing database schema in internal/storage

## Risks & Mitigation

| Risk | Impact | Probability | Mitigation |
|------|--------|-------------|------------|
| Agents continue using grep/sqlite3 | Medium | Medium | Update all templates with new commands, add deprecation warnings |
| Performance issues with large codebases | Medium | Low | Optimize queries, add caching, limit results |
| Database schema changes break queries | High | Low | Use abstraction layer, version queries |

## Review & Acceptance Checklist

**Content Quality:**
- [x] No implementation details (languages, frameworks, APIs)
- [x] Focused on user value and business needs
- [x] Written for non-technical stakeholders
- [x] All mandatory sections completed

**Requirement Completeness:**
- [x] No [NEEDS CLARIFICATION] markers remain
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

### CLI Commands

<!-- CANARY: REQ=CBIN-CLI-001; FEATURE="ShowCmd"; ASPECT=CLI; STATUS=TESTED; UPDATED=2025-10-16 -->
**Feature 1: Show Command**
- [ ] Add `showCmd` to cmd/canary/main.go
- [ ] Implement token query by requirement ID
- [ ] Format output as table grouped by aspect
- [ ] Support --json, --group-by, --no-color flags
- **Location hint:** cmd/canary/main.go, cmd/canary/show.go (new)
- **Dependencies:** None

<!-- CANARY: REQ=CBIN-CLI-001; FEATURE="FilesCmd"; ASPECT=CLI; STATUS=TESTED; UPDATED=2025-10-16 -->
**Feature 2: Files Command**
- [ ] Add `filesCmd` to cmd/canary/main.go
- [ ] Query unique file paths for requirement
- [ ] Filter out spec/template directories
- [ ] Group by aspect, show token counts
- [ ] Support --all flag to include spec files
- **Location hint:** cmd/canary/main.go, cmd/canary/files.go (new)
- **Dependencies:** ShowCmd (shares query logic)

<!-- CANARY: REQ=CBIN-CLI-001; FEATURE="StatusCmd"; ASPECT=CLI; STATUS=TESTED; UPDATED=2025-10-16 -->
**Feature 3: Status Command**
- [ ] Add `statusCmd` to cmd/canary/main.go
- [ ] Calculate token counts by status
- [ ] Compute completion percentage
- [ ] Display colored progress bar
- [ ] List incomplete work with file locations
- **Location hint:** cmd/canary/main.go, cmd/canary/status.go (new)
- **Dependencies:** ShowCmd (uses same queries)

<!-- CANARY: REQ=CBIN-CLI-001; FEATURE="GrepCmd"; ASPECT=CLI; STATUS=TESTED; UPDATED=2025-10-16 -->
**Feature 4: Grep Command**
- [ ] Add `grepCmd` to cmd/canary/main.go
- [ ] Implement multi-field search (req_id, feature, aspect, owner, keywords)
- [ ] Support case-insensitive (default) and --case-sensitive
- [ ] Support --regex flag for regex patterns
- [ ] Highlight matched text in output
- **Location hint:** cmd/canary/main.go, cmd/canary/grep.go (new)
- **Dependencies:** None

### Database Abstraction

<!-- CANARY: REQ=CBIN-CLI-001; FEATURE="QueryAbstraction"; ASPECT=Storage; STATUS=TESTED; UPDATED=2025-10-16 -->
**Feature 5: Query Abstraction Layer**
- [ ] Create GetTokensByReqID(reqID string) in internal/storage
- [ ] Create GetFilesByReqID(reqID string) in internal/storage
- [ ] Create SearchTokens(pattern string, fields []string) in internal/storage
- [ ] Implement filesystem fallback when DB missing
- **Location hint:** internal/storage/queries.go (new)
- **Dependencies:** Existing storage package

### Testing

<!-- CANARY: REQ=CBIN-CLI-001; FEATURE="ShowCmdTests"; ASPECT=CLI; STATUS=TESTED; TEST=TestCANARY_CBIN_CLI_001_ShowCmd; UPDATED=2025-10-16 -->
**Unit Tests: Show Command**
- [ ] Test exact ID match
- [ ] Test output formatting
- [ ] Test --json flag
- [ ] Test --group-by flag
- [ ] Test error handling (not found)
- **Location hint:** cmd/canary/show_test.go (new)

<!-- CANARY: REQ=CBIN-CLI-001; FEATURE="FilesCmdTests"; ASPECT=CLI; STATUS=TESTED; TEST=TestCANARY_CBIN_CLI_001_FilesCmd; UPDATED=2025-10-16 -->
**Unit Tests: Files Command**
- [ ] Test file grouping by aspect
- [ ] Test spec/template filtering
- [ ] Test --all flag
- [ ] Test token counting
- **Location hint:** cmd/canary/files_test.go (new)

<!-- CANARY: REQ=CBIN-CLI-001; FEATURE="StatusCmdTests"; ASPECT=CLI; STATUS=TESTED; TEST=TestCANARY_CBIN_CLI_001_StatusCmd; UPDATED=2025-10-16 -->
**Unit Tests: Status Command**
- [ ] Test status counting
- [ ] Test completion percentage
- [ ] Test colored output
- [ ] Test incomplete work listing
- **Location hint:** cmd/canary/status_test.go (new)

### Documentation Updates

<!-- CANARY: REQ=CBIN-CLI-001; FEATURE="TemplateUpdates"; ASPECT=Docs; STATUS=TESTED; UPDATED=2025-10-16 -->
**Feature 9: Update Agent Templates**
- [ ] Update .canary/templates/next-prompt-template.md to use `canary show`
- [ ] Update .canary/templates/implement-prompt-template.md to use `canary files`
- [ ] Update CLAUDE.md with new command examples
- [ ] Add deprecation warnings for grep/sqlite3 usage
- **Location hint:** .canary/templates/, CLAUDE.md
- **Dependencies:** All commands implemented

<!-- CANARY: REQ=CBIN-CLI-001; FEATURE="CommandDocs"; ASPECT=Docs; STATUS=TESTED; UPDATED=2025-10-16 -->
**Feature 10: Command Documentation**
- [ ] Add help text for show command
- [ ] Add help text for files command
- [ ] Add help text for status command
- [ ] Add help text for grep command
- [ ] Update README with new commands
- **Location hint:** cmd/canary/main.go (help text), README.md
- **Dependencies:** All commands implemented

---

**Agent Instructions:**

After implementing each feature:
1. Update the CANARY token in the spec from `STATUS=STUB` to `STATUS=IMPL`
2. Add the same token to your source code at the implementation location
3. Add `TEST=TestName` when tests are written
4. Run `canary implement CBIN-CLI-001` to see implementation progress

---

## CANARY Tokens Reference

**Main requirement token** (add to primary implementation file):
```
// CANARY: REQ=CBIN-CLI-001; FEATURE="FeatureName"; ASPECT=API; STATUS=TESTED; UPDATED=2025-10-16
```

**Sub-feature tokens** (use the specific feature names from Implementation Checklist):
```
// CANARY: REQ=CBIN-CLI-001; FEATURE="CoreFeature1"; ASPECT=API; STATUS=TESTED; TEST=TestCoreFeature1; UPDATED=2025-10-16
```

**Use `canary implement CBIN-CLI-001` to find:**
- Which features are implemented vs. still TODO
- Exact file locations and line numbers
- Context around each implementation point
