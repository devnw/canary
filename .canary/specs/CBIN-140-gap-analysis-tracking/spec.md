# CANARY: REQ=CBIN-140; FEATURE="GapAnalysisTracking"; ASPECT=CLI; STATUS=STUB; UPDATED=2025-10-17
# Feature Specification: Gap Analysis Tracking

**Requirement ID:** CBIN-140
**Status:** STUB
**Created:** 2025-10-17
**Last Updated:** 2025-10-17

## Overview

**Purpose:** Enable AI agents and developers to record, track, and learn from implementation mistakes by annotating CANARY tokens with gap analysis data. When implementations, tests, benchmarks, or fuzzing are incorrect, the system captures what went wrong and uses this information to adjust future AI agent behavior and implementation strategies.

**Scope:** This feature includes CLI commands to mark implementations as incorrect, annotate CANARY tokens with gap analysis data, track patterns of mistakes, and use historical gap data to inform future agent prompts and decision-making. Excludes automatic bug detection or AI model retraining.

## User Stories

### Primary User Stories

**US-1: Record Implementation Mistakes**
As an AI agent or developer,
I want to mark a CANARY implementation as incorrect and record what went wrong,
So that the system learns from mistakes and avoids repeating them in future implementations.

**Acceptance Criteria:**
- [ ] Agent can mark a specific CANARY token (by REQ-ID + FEATURE) as having incorrect implementation
- [ ] System prompts for categorization of the mistake (implementation logic, test coverage, performance, security, API design)
- [ ] Gap annotation is persisted and associated with the original CANARY token
- [ ] Agent can add free-form notes about what was wrong and why

**US-2: Query Historical Mistakes**
As an AI agent,
I want to retrieve gap analysis data for similar features before implementing new requirements,
So that I can avoid making the same mistakes that were made previously.

**Acceptance Criteria:**
- [ ] Agent can query gap analysis by requirement ID, feature aspect, or mistake category
- [ ] System returns relevant historical mistakes with context and lessons learned
- [ ] Query results include frequency of specific mistake patterns
- [ ] Results are formatted for inclusion in agent prompts

**US-3: Adjust Future Behavior**
As an AI agent,
I want the system to inject relevant gap analysis into my planning and implementation prompts,
So that I automatically receive guidance based on past mistakes without manual lookups.

**Acceptance Criteria:**
- [ ] When creating a plan for a requirement, system includes relevant gap analysis from similar past work
- [ ] Agent configuration can specify which categories of gaps to prioritize
- [ ] Gap data appears in agent context with clear recommendations
- [ ] Agent can update gap priority/relevance based on whether it prevented future mistakes

### Secondary User Stories (if applicable)

**US-4: Generate Gap Analysis Reports**
As a project maintainer,
I want to view aggregated reports of common implementation mistakes,
So that I can identify systemic issues and improve project documentation or constitutional principles.

**Acceptance Criteria:**
- [ ] CLI command generates summary report of all gap analysis entries
- [ ] Report groups mistakes by category, aspect, and frequency
- [ ] Report identifies most common mistake patterns across requirements

## Functional Requirements

### FR-1: Mark Implementation as Incorrect
**Priority:** High
**Description:** System must provide CLI command to mark a CANARY implementation as incorrect and record gap analysis data including mistake category, description, and corrective actions taken.
**Acceptance:** User can run command with REQ-ID and FEATURE, system prompts for details, and gap data is persisted in queryable storage.

### FR-2: Annotate CANARY Tokens
**Priority:** High
**Description:** System must associate gap analysis data with specific CANARY tokens in a way that preserves original token integrity while adding metadata about mistakes and corrections.
**Acceptance:** Gap data is linked to tokens without modifying source code tokens; annotations are retrievable when querying tokens.

### FR-3: Query Gap Analysis
**Priority:** High
**Description:** System must allow filtering and retrieval of gap analysis by requirement ID, feature name, aspect, mistake category, and date range.
**Acceptance:** CLI command accepts filter parameters and returns matching gap entries with full context in human-readable and machine-parseable formats.

### FR-4: Inject Gaps into Agent Context
**Priority:** Medium
**Description:** System must provide mechanism to retrieve relevant gap analysis data during planning and implementation phases to inform agent decision-making.
**Acceptance:** When agent runs planning commands, system automatically retrieves and formats relevant gap data; agent can configure which gap categories to include.

### FR-5: Update Gap Relevance
**Priority:** Medium
**Description:** System must allow agents to mark gap analysis entries as helpful/unhelpful and track which gaps successfully prevented future mistakes.
**Acceptance:** Agent can rate gap entries; system tracks effectiveness metrics; low-relevance gaps can be filtered from future queries.

### FR-6: Gap Analysis Storage
**Priority:** High
**Description:** System must persist gap analysis data in structured format that supports efficient querying, versioning, and migration.
**Acceptance:** Gap data survives across CLI sessions; supports concurrent access; can be exported/imported for sharing across projects.

## Success Criteria

**Quantitative Metrics:**
- [ ] Agent can record a gap analysis entry in under 30 seconds
- [ ] Gap query returns results in under 2 seconds for 1000+ gap entries
- [ ] 80% of agents report that injected gap data helped avoid repeating mistakes
- [ ] Gap analysis storage supports at least 10,000 entries without performance degradation

**Qualitative Measures:**
- [ ] Gap analysis entries contain actionable information that can inform future implementations
- [ ] Agents successfully use gap data to make different decisions than they would have without it
- [ ] Project maintainers find gap reports useful for identifying systemic issues
- [ ] Gap annotation process does not disrupt normal development workflow

## User Scenarios & Testing

### Scenario 1: Recording Implementation Mistake
**Given:** Agent implemented CBIN-042 OAuth feature but forgot to validate redirect URIs
**When:** Developer runs `canary gap mark CBIN-042 OAuth --category security`
**Then:** System prompts for description, agent provides "Missing redirect URI validation in OAuth flow - security vulnerability", system persists gap with timestamp and category

### Scenario 2: Querying Before Implementation
**Given:** Multiple past authentication features had security gaps
**When:** Agent runs `canary gap query --aspect Security --category security` before implementing new auth feature
**Then:** System returns list of security mistakes from past auth work, agent incorporates lessons into new implementation plan

### Scenario 3: Auto-Injection into Planning
**Given:** Agent is creating plan for CBIN-150 which is an API feature
**When:** Agent runs `/canary.plan CBIN-150`
**Then:** System detects ASPECT=API, retrieves relevant API-related gaps, injects formatted gap data into planning context, agent sees warnings about common API mistakes

### Scenario 4: Gap Effectiveness Tracking
**Given:** Gap entry warned about missing error handling in parsers
**When:** Agent implements new parser and successfully includes error handling due to gap warning
**Then:** Agent runs `canary gap helpful GAP-005` to mark gap as effective, system increments usefulness counter

### Scenario 5: Gap Analysis Report Generation
**Given:** Project has 100+ gap entries across 50 requirements
**When:** Maintainer runs `canary gap report --summary`
**Then:** System generates report showing "Security: 15 issues, Implementation Logic: 40 issues, Test Coverage: 30 issues" with top patterns listed

## Key Entities (if data-driven feature)

### Entity 1: GapAnalysisEntry
**Attributes:**
- gap_id: Unique identifier for the gap entry
- req_id: CANARY requirement ID (e.g., CBIN-042)
- feature: Feature name from CANARY token
- aspect: CANARY aspect (API, CLI, Storage, Security, etc.)
- category: Type of mistake (implementation, testing, performance, security, design)
- description: Free-form text explaining what went wrong
- corrective_action: What was done to fix the issue
- created_at: Timestamp when gap was recorded
- created_by: Agent or user who recorded the gap
- helpful_count: Number of times marked as helpful
- unhelpful_count: Number of times marked as unhelpful

**Relationships:**
- Links to CANARY token via req_id + feature
- May reference multiple related gap entries for pattern tracking

### Entity 2: GapCategory
**Attributes:**
- category_name: Standard category (implementation, testing, performance, security, design, documentation)
- description: What types of mistakes fall into this category
- priority: Default priority for this category in agent context

**Relationships:**
- One-to-many with GapAnalysisEntry

### Entity 3: GapConfiguration
**Attributes:**
- agent_id: Identifier for agent or user configuration
- enabled_categories: List of gap categories to include in prompts
- max_gaps_per_query: Limit on how many gaps to inject
- similarity_threshold: How closely gaps must match current work to be included

**Relationships:**
- Controls which gaps are surfaced to which agents

## Assumptions

- Agents have ability to identify when implementation is incorrect (through testing, user feedback, or code review)
- Gap analysis data will be stored locally in project repository (not remote service)
- Agents will be honest in recording mistakes and marking gaps as helpful/unhelpful
- Gap data will be reviewed periodically to ensure quality and remove obsolete entries

## Constraints

**Technical Constraints:**
- Must work with existing CANARY token format without breaking current scanning
- Must not require modification of source code files to store gap data
- Query performance must scale to thousands of gap entries

**Business Constraints:**
- Implementation should reuse existing storage mechanisms (SQLite database) if possible
- CLI interface must be simple enough for both humans and AI agents to use

**Regulatory Constraints:**
- Gap data may contain sensitive information about security vulnerabilities; ensure appropriate access controls

## Out of Scope

- Automatic detection of implementation mistakes (requires manual marking)
- AI model fine-tuning or retraining based on gap data
- Integration with external issue tracking systems
- Real-time gap detection during implementation
- Automatic remediation of gaps
- Gap data synchronization across multiple projects or teams

## Dependencies

- Existing CANARY scanning infrastructure (CBIN-101, CBIN-102)
- SQLite database for gap storage
- `/canary.plan` command integration for auto-injection
- CANARY token parser to link gaps to tokens

## Risks & Mitigation

| Risk | Impact | Probability | Mitigation |
|------|--------|-------------|------------|
| Gap data becomes stale and misleading | High | Medium | Implement gap aging/review system; allow gaps to be marked obsolete |
| Too many gaps overwhelm agent context | Medium | High | Implement smart filtering; limit injection to most relevant gaps |
| Agents don't record gaps consistently | High | Medium | Make recording process quick and simple; provide templates |
| Gap data grows unbounded | Medium | Medium | Implement archival strategy; allow pruning of low-value gaps |
| Security-sensitive gaps leaked in reports | High | Low | Add sensitivity markers; restrict access to security gaps |

## Clarifications Needed

[NEEDS CLARIFICATION: Should gap data be committed to version control or kept in local-only database?]
**Options:** A) Commit to git for team sharing, B) Keep local only, C) Support both with configuration flag
**Impact:** Determines whether team members can learn from each other's mistakes or if gap data is per-developer

[NEEDS CLARIFICATION: How should gaps be prioritized when multiple gaps match a query?]
**Options:** A) By recency, B) By helpful_count, C) By category priority, D) Configurable ranking
**Impact:** Affects which gaps agents see first and whether old or frequently-confirmed gaps take precedence

## Review & Acceptance Checklist

**Content Quality:**
- [x] No implementation details (languages, frameworks, APIs)
- [x] Focused on user value and business needs
- [x] Written for non-technical stakeholders
- [x] All mandatory sections completed

**Requirement Completeness:**
- [ ] No [NEEDS CLARIFICATION] markers remain
- [x] Requirements are testable and unambiguous
- [x] Success criteria are measurable and technology-agnostic
- [x] All acceptance scenarios defined
- [x] Edge cases identified
- [x] Scope clearly bounded
- [x] Dependencies and assumptions identified

**Readiness:**
- [x] All functional requirements have clear acceptance criteria
- [x] User scenarios cover primary flows
- [ ] Ready for technical planning (`/canary.plan`) - after clarifications resolved

---

## Implementation Checklist

Break down this requirement into specific implementation points. Each point gets its own CANARY token to help agents locate where to implement changes.

### Core Features

<!-- CANARY: REQ=CBIN-140; FEATURE="GapMarkCmd"; ASPECT=CLI; STATUS=STUB; UPDATED=2025-10-17 -->
**Feature 1: Gap Mark Command**
- [ ] Implement CLI command to mark CANARY as incorrect with gap details
- **Location hint:** "cmd/gap.go" or "cmd/gap_mark.go"
- **Dependencies:** Gap storage system

<!-- CANARY: REQ=CBIN-140; FEATURE="GapQueryCmd"; ASPECT=CLI; STATUS=STUB; UPDATED=2025-10-17 -->
**Feature 2: Gap Query Command**
- [ ] Implement CLI command to search and filter gap entries
- **Location hint:** "cmd/gap_query.go"
- **Dependencies:** Gap storage system

<!-- CANARY: REQ=CBIN-140; FEATURE="GapReportCmd"; ASPECT=CLI; STATUS=STUB; UPDATED=2025-10-17 -->
**Feature 3: Gap Report Command**
- [ ] Implement CLI command to generate aggregated gap analysis reports
- **Location hint:** "cmd/gap_report.go"
- **Dependencies:** Gap storage, query system

<!-- CANARY: REQ=CBIN-140; FEATURE="GapHelpfulCmd"; ASPECT=CLI; STATUS=STUB; UPDATED=2025-10-17 -->
**Feature 4: Gap Helpful/Unhelpful Rating**
- [ ] Implement CLI command to rate gap usefulness
- **Location hint:** "cmd/gap_rate.go"
- **Dependencies:** Gap storage

### Data Layer

<!-- CANARY: REQ=CBIN-140; FEATURE="GapSchema"; ASPECT=Storage; STATUS=STUB; UPDATED=2025-10-17 -->
**Data Model:**
- [ ] Define schema for gap analysis entries, categories, configurations
- **Location hint:** "internal/storage/schema.go" or migration files

<!-- CANARY: REQ=CBIN-140; FEATURE="GapStorage"; ASPECT=Storage; STATUS=STUB; UPDATED=2025-10-17 -->
**Data Access:**
- [ ] Implement CRUD operations for gap entries (Create, Read, Update, Delete)
- **Location hint:** "internal/storage/gap_repository.go"

<!-- CANARY: REQ=CBIN-140; FEATURE="GapQuery"; ASPECT=Storage; STATUS=STUB; UPDATED=2025-10-17 -->
**Gap Query Engine:**
- [ ] Implement filtering, sorting, and relevance ranking for gap queries
- **Location hint:** "internal/gap/query.go"

### Integration

<!-- CANARY: REQ=CBIN-140; FEATURE="PlanIntegration"; ASPECT=Planner; STATUS=STUB; UPDATED=2025-10-17 -->
**Plan Command Integration:**
- [ ] Modify plan command to auto-inject relevant gap analysis
- **Location hint:** ".canary/templates/commands/plan.md" or plan execution logic
- **Dependencies:** Gap query system

<!-- CANARY: REQ=CBIN-140; FEATURE="CanaryTokenLink"; ASPECT=Engine; STATUS=STUB; UPDATED=2025-10-17 -->
**CANARY Token Linking:**
- [ ] Link gap entries to CANARY tokens for bidirectional lookup
- **Location hint:** "internal/scanner/token_parser.go" or "internal/gap/linker.go"
- **Dependencies:** Token scanner

### Testing Requirements

<!-- CANARY: REQ=CBIN-140; FEATURE="GapUnitTests"; ASPECT=CLI; STATUS=STUB; TEST=TestGapCommands; UPDATED=2025-10-17 -->
**Unit Tests:**
- [ ] Test gap CRUD operations, query filtering, rating system
- **Location hint:** "internal/gap/*_test.go", "cmd/gap_test.go"

<!-- CANARY: REQ=CBIN-140; FEATURE="GapIntegrationTests"; ASPECT=CLI; STATUS=STUB; TEST=TestGapWorkflow; UPDATED=2025-10-17 -->
**Integration Tests:**
- [ ] Test end-to-end workflow: mark gap → query gaps → rate gap → generate report
- **Location hint:** "test/integration/gap_test.go"

<!-- CANARY: REQ=CBIN-140; FEATURE="GapInjectionTests"; ASPECT=Planner; STATUS=STUB; TEST=TestPlanGapInjection; UPDATED=2025-10-17 -->
**Plan Integration Tests:**
- [ ] Test that relevant gaps are injected into plan command context
- **Location hint:** "test/integration/plan_gap_test.go"

### Documentation

<!-- CANARY: REQ=CBIN-140; FEATURE="GapCLIDocs"; ASPECT=Docs; STATUS=STUB; UPDATED=2025-10-17 -->
**CLI Documentation:**
- [ ] Document all gap-related commands with examples
- **Location hint:** "docs/cli/gap.md", "README.md"

<!-- CANARY: REQ=CBIN-140; FEATURE="GapWorkflowGuide"; ASPECT=Docs; STATUS=STUB; UPDATED=2025-10-17 -->
**Workflow Guide:**
- [ ] Create guide for agents on when and how to record gaps
- **Location hint:** ".canary/AGENT_CONTEXT.md", "docs/workflow/gap-tracking.md"

---

**Agent Instructions:**

After implementing each feature:
1. Update the CANARY token in the spec from `STATUS=STUB` to `STATUS=IMPL`
2. Add the same token to your source code at the implementation location
3. Add `TEST=TestName` when tests are written
4. Run `canary implement CBIN-140` to see implementation progress

---

## CANARY Tokens Reference

**Main requirement token** (add to primary implementation file):
```
// CANARY: REQ=CBIN-140; FEATURE="GapAnalysisTracking"; ASPECT=CLI; STATUS=IMPL; UPDATED=2025-10-17
```

**Sub-feature tokens** (use the specific feature names from Implementation Checklist):
```
// CANARY: REQ=CBIN-140; FEATURE="GapMarkCmd"; ASPECT=CLI; STATUS=IMPL; TEST=TestGapMark; UPDATED=2025-10-17
```

**Use `canary implement CBIN-140` to find:**
- Which features are implemented vs. still TODO
- Exact file locations and line numbers
- Context around each implementation point
