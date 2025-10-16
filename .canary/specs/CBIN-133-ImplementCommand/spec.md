# Feature Specification: Implement Command

**Requirement ID:** CBIN-133
**Feature Name:** ImplementCommand
**Status:** STUB
**Owner:** canary
**Created:** 2025-10-16
**Updated:** 2025-10-16

---

## Feature Overview

### Summary
Provide a `canary implement` CLI command and `/canary.implement` slash command that guides AI coding agents through the implementation of a specified requirement. The command should intelligently locate requirements using fuzzy matching, load the complete specification and implementation plan, and generate a comprehensive implementation prompt that includes all necessary context for safe and correct implementation.

### Business Value
- **Reduced Implementation Errors**: Agents receive complete context (spec, plan, constitution) before coding
- **Faster Onboarding**: New agents can immediately understand what to implement and how
- **Consistent Quality**: Every implementation follows the same workflow (specify → plan → implement)
- **Traceability**: CANARY token placement instructions ensure all implementations are tracked
- **Self-Service**: Agents can autonomously select and implement requirements without human intervention

### Target Users
- AI coding agents (Claude Code, Cursor, GitHub Copilot, etc.)
- Human developers following the CANARY workflow
- Automated CI/CD systems implementing features

---

## User Stories

### US-1: Agent Implements Requirement by ID
**As an** AI coding agent
**I want to** request implementation guidance for a specific requirement ID
**So that** I can implement the feature with full context about spec, plan, and token placement

**Acceptance Criteria:**
- Agent runs `/canary.implement CBIN-105` or `canary implement CBIN-105`
- System loads specification from `.canary/specs/CBIN-105-*/spec.md`
- System loads implementation plan from `.canary/specs/CBIN-105-*/plan.md`
- System loads constitutional principles from `.canary/memory/constitution.md`
- System generates comprehensive implementation prompt including:
  - What to build (from spec)
  - How to build it (from plan)
  - Governing principles (from constitution)
  - Where to place CANARY tokens
  - Test-first approach guidance
- Agent receives actionable implementation instructions

### US-2: Agent Searches for Requirement by Name
**As an** AI coding agent
**I want to** search for a requirement using a feature name or keyword
**So that** I can find the right requirement without knowing the exact ID

**Acceptance Criteria:**
- Agent runs `canary implement "user authentication"` with partial name
- System performs fuzzy matching against:
  - Requirement IDs (CBIN-XXX)
  - Feature names from spec.md files
  - Keywords from spec content
- System displays top 5 matches with scores
- If single clear match (score > 80%), auto-selects it
- If multiple matches, prompts user to select from list
- After selection, proceeds with implementation prompt generation

### US-3: Developer Lists Available Requirements
**As a** human developer
**I want to** see all unimplemented requirements
**So that** I can choose what to work on next

**Acceptance Criteria:**
- Developer runs `canary implement --list`
- System displays all requirements with STATUS=STUB or STATUS=IMPL (not yet TESTED)
- Output shows: requirement ID, feature name, status, priority (if available)
- List is sorted by priority (highest first)
- Developer can then run `canary implement <REQ-ID>` for chosen requirement

### US-4: Agent Receives File Location Hints
**As an** AI coding agent
**I want to** know exactly where to implement each sub-feature
**So that** I don't waste time searching the codebase or create files in wrong locations

**Acceptance Criteria:**
- Implementation prompt includes "Implementation Checklist" from spec
- Each sub-feature shows:
  - Feature name
  - CANARY token to place
  - Suggested file locations (from spec "Location hint")
  - Dependencies on other features
- Agent can navigate directly to suggested locations
- If location doesn't exist, agent creates it following project conventions

---

## Functional Requirements

### FR-1: Requirement Lookup
The system shall support multiple lookup methods:
1. **By Requirement ID**: Exact match on CBIN-XXX format
2. **By Feature Name**: Fuzzy matching on feature names from spec files
3. **By Keyword**: Full-text search in spec content
4. **By Directory Name**: Match against `.canary/specs/CBIN-XXX-feature-name/` patterns

**Acceptance:** Running `canary implement <query>` successfully finds and loads the correct specification within 1 second for codebases with <1000 requirements.

### FR-2: Fuzzy Matching Algorithm
The system shall implement fuzzy string matching with:
- **Levenshtein distance** for typo tolerance (max distance: 3)
- **Substring matching** for partial names
- **Abbreviation matching** (e.g., "auth" matches "authentication")
- **Score threshold**: Minimum 60% similarity to show in results
- **Auto-selection**: If top match scores >80% and 20+ points ahead of second match, auto-select

**Acceptance:**
- "user auth" successfully matches "UserAuthentication" feature
- "CBIN105" matches "CBIN-105" requirement
- "jwt token" matches specification containing "JWT token validation"

### FR-3: Implementation Prompt Generation
The system shall generate prompts including:
- **Specification Content**: Full spec.md content
- **Implementation Plan**: Full plan.md content (if exists)
- **Constitutional Guidance**: Relevant articles from constitution.md
- **Token Placement Instructions**: CANARY token examples for each sub-feature
- **Test-First Guidance**: Article IV requirements and test examples
- **File Location Hints**: Suggested directories/files from Implementation Checklist
- **Success Criteria**: Measurable outcomes from spec
- **Dependencies**: DEPENDS_ON requirements with status check

**Acceptance:** Generated prompt is 3,000-8,000 words and includes all sections above.

### FR-4: Interactive Selection
When multiple matches exist (score difference <20 points), the system shall:
- Display numbered list of top 5 matches
- Show for each match:
  - Requirement ID
  - Feature name
  - Match score percentage
  - Current status
  - Brief description (first 100 chars of spec)
- Accept numeric selection input
- Re-prompt if invalid selection
- Allow "q" to quit without selection

**Acceptance:** User can select from ambiguous matches within 3 seconds of seeing the list.

### FR-5: Implementation Status Tracking
The system shall track implementation progress by:
- Scanning codebase for CANARY tokens matching the requirement ID
- Counting sub-features by status (STUB, IMPL, TESTED, BENCHED)
- Showing progress percentage in prompt header
- Warning if plan.md missing (suggesting to run `canary plan CBIN-XXX` first)

**Acceptance:** Implementation prompt header shows "Progress: 3/10 sub-features completed (30%)"

### FR-6: Slash Command Integration
The system shall:
- Create `.canary/templates/commands/implement.md` template
- Install to all supported agent directories (`.claude/`, `.cursor/`, etc.)
- Command invokes `canary implement <query> --prompt` to generate guidance
- Slash command passes through all CLI flags

**Acceptance:** Typing `/canary.implement CBIN-105` in Claude Code generates full implementation prompt.

### FR-7: CLI Flags
```bash
canary implement <query> [flags]

Flags:
  --list              List all unimplemented requirements (ignores query)
  --prompt            Generate full implementation prompt (default: true for slash commands)
  --json              Output in JSON format
  --show-progress     Show implementation progress without generating prompt
  --context-lines N   Show N lines of code context around existing tokens (default: 3)
```

**Acceptance:** All flags work as documented and are validated on input.

### FR-8: Error Handling
The system shall handle:
- **No Match Found**: Display helpful message with suggestions (run `canary specify`, check spelling)
- **Missing Spec File**: Explain that spec must be created first via `/canary.specify`
- **Missing Plan**: Warn that plan is recommended, offer to run `canary plan CBIN-XXX`
- **Invalid Requirement ID**: Suggest valid format (CBIN-XXX) and show example
- **Filesystem Access Errors**: Gracefully degrade with clear error messages

**Acceptance:** All error cases display actionable guidance without stack traces.

---

## Success Criteria

### Measurability

**Quantitative Metrics:**
1. **Lookup Speed**: 95% of queries resolve in <1 second for codebases with <1000 requirements
2. **Fuzzy Match Accuracy**: 90% of partial queries correctly identify intended requirement
3. **Auto-Selection Accuracy**: 85% of auto-selections (score >80%) are correct
4. **Prompt Completeness**: 100% of generated prompts include all required sections (spec, plan, constitution, tokens)
5. **Agent Satisfaction**: Post-implementation surveys show >4.0/5.0 rating for clarity and completeness

**Qualitative Measures:**
- Agents successfully implement requirements without requesting additional context
- Implementation errors due to missing information reduced by 50%
- Time from `/canary.implement` to first code commit reduced by 30%

### Quality Gates

- [ ] Fuzzy matching correctly identifies requirements with typos (Levenshtein distance ≤3)
- [ ] Generated prompts pass constitutional validation (reference Articles I, IV, V, VII)
- [ ] Implementation Checklist locations exist in codebase or are valid creation targets
- [ ] Token placement examples are syntactically correct for target language
- [ ] All dependencies (DEPENDS_ON) are checked and reported if incomplete
- [ ] Prompts are formatted as valid markdown
- [ ] Interactive selection handles invalid input gracefully

---

## Acceptance Scenarios

### Scenario 1: Exact ID Match (Happy Path)
**Given:** Specification exists at `.canary/specs/CBIN-105-user-authentication/spec.md`
**And:** Implementation plan exists at `.canary/specs/CBIN-105-user-authentication/plan.md`
**When:** Agent runs `canary implement CBIN-105 --prompt`
**Then:**
- System loads spec.md content
- System loads plan.md content
- System loads constitution.md content
- System generates prompt with all sections
- Prompt includes CANARY token placement examples
- Prompt displays progress: "0/5 sub-features completed (0%)"
- Agent receives comprehensive implementation guidance

### Scenario 2: Fuzzy Match with Auto-Selection
**Given:** Specification CBIN-107 has feature name "UserAuthentication"
**When:** Agent runs `canary implement "user auth"`
**Then:**
- System calculates fuzzy match scores
- "UserAuthentication" scores 85%
- Second-best match scores 60%
- Score difference is >20 points
- System auto-selects CBIN-107
- Displays: "Auto-selected: CBIN-107 - UserAuthentication (85% match)"
- Proceeds with prompt generation

### Scenario 3: Fuzzy Match with Interactive Selection
**Given:** Three specifications match query "auth":
  - CBIN-105: "UserAuthentication" (78% match)
  - CBIN-110: "OAuthIntegration" (75% match)
  - CBIN-112: "AuthorizationRules" (72% match)
**When:** Agent runs `canary implement "auth"`
**Then:**
- System displays:
  ```
  Multiple matches found. Select one:

  1. CBIN-105 - UserAuthentication (78% match, STATUS=STUB)
     Email/password and OAuth2 user authentication

  2. CBIN-110 - OAuthIntegration (75% match, STATUS=IMPL)
     OAuth2 provider integration with Google and GitHub

  3. CBIN-112 - AuthorizationRules (72% match, STATUS=STUB)
     Role-based authorization and permissions

  Enter number (1-3) or 'q' to quit:
  ```
- Agent enters "1"
- System loads CBIN-105 and generates prompt

### Scenario 4: List All Unimplemented Requirements
**Given:** Database contains 10 STUB requirements and 5 IMPL requirements
**When:** Developer runs `canary implement --list`
**Then:**
- System displays:
  ```
  Unimplemented Requirements (15):

  Priority 1:
    CBIN-105 - UserAuthentication (STUB)
    CBIN-110 - OAuthIntegration (STUB)

  Priority 5:
    CBIN-112 - AuthorizationRules (STUB)
    CBIN-115 - DataValidation (IMPL, 3/5 features completed)

  No Priority Set:
    CBIN-120 - EmailNotifications (STUB)
  ```
- Exit code 0

### Scenario 5: Missing Specification
**Given:** No specification exists for CBIN-999
**When:** Agent runs `canary implement CBIN-999`
**Then:**
- System displays:
  ```
  ❌ Specification not found for CBIN-999

  Possible reasons:
  - Requirement hasn't been created yet
  - Incorrect requirement ID format

  Next steps:
  1. Create specification: /canary.specify "your feature description"
  2. List existing requirements: canary implement --list
  3. Search by name: canary implement "feature name"
  ```
- Exit code 1

### Scenario 6: Missing Implementation Plan
**Given:** Specification exists at `.canary/specs/CBIN-105-*/spec.md`
**And:** No plan.md file exists
**When:** Agent runs `canary implement CBIN-105 --prompt`
**Then:**
- System displays warning:
  ```
  ⚠️  Warning: Implementation plan not found for CBIN-105

  Recommendation: Create plan first for better guidance
  Run: canary plan CBIN-105

  Continue without plan? (y/n):
  ```
- If "y": Generates prompt without plan section
- If "n": Exit code 1

---

## Assumptions and Constraints

### Assumptions
1. Specifications exist in `.canary/specs/CBIN-XXX-*/spec.md` format
2. Implementation plans exist in `.canary/specs/CBIN-XXX-*/plan.md` format (optional but recommended)
3. Constitution exists at `.canary/memory/constitution.md`
4. Specs follow template structure with "Implementation Checklist" section
5. Agents can process markdown-formatted prompts up to 10,000 words
6. Fuzzy matching library available (or simple Levenshtein implementation acceptable)

### Constraints
1. **Performance**: Prompt generation must complete in <2 seconds for responsive agent experience
2. **Fuzzy Matching**: Levenshtein distance algorithm has O(n*m) complexity - limit to first 1000 specs for large codebases
3. **Prompt Size**: Generated prompts must fit within agent context windows (max 10,000 words)
4. **Interactive Mode**: Only works in TTY environments (batch/CI mode must use exact IDs)
5. **Memory**: Loading all spec files for fuzzy matching requires <100MB RAM

### Technical Constraints
- Go 1.19+ required
- Filesystem access required for spec/plan file loading
- UTF-8 encoding for all template files
- Markdown rendering not required (agents display raw markdown)

---

## Out of Scope

- **Automated Code Generation**: This command provides guidance, not generated code
- **Git Integration**: No automatic commit/branch creation (agents handle this)
- **Spec Validation**: Assumes specs are valid (validation is `/canary.verify` responsibility)
- **Multi-Requirement Implementation**: Only handles one requirement at a time
- **Progress Persistence**: Progress calculation is real-time scan, not persisted
- **Dependency Auto-Resolution**: Warns about unresolved dependencies but doesn't auto-implement them

---

## Related Requirements

**DEPENDS_ON:**
- CBIN-120: SpecifyCmd (specifications must exist before implementation)
- CBIN-121: PlanCmd (plans should exist for guided implementation)

**BLOCKS:**
- None

**RELATED_TO:**
- CBIN-132: NextCmd (shares priority/selection logic, different use case)
- CBIN-110: SpecifyCmd (shared spec file loading)
- CBIN-106: AgentContext (shares template rendering approach)

---

## Open Questions

None - requirement is fully specified.

---

## Implementation Checklist

### Phase 1: CLI Command (STUB → IMPL)

<!-- CANARY: REQ=CBIN-133; FEATURE="ImplementCmd"; ASPECT=CLI; STATUS=STUB; UPDATED=2025-10-16 -->
**Feature 1: Cobra Command Definition**
- [ ] Add `implementCmd` to `cmd/canary/main.go`
- [ ] Define flags: --list, --prompt, --json, --show-progress, --context-lines
- [ ] Register with rootCmd in init()
- **Location hint:** `cmd/canary/main.go` (already has similar commands as examples)
- **Dependencies:** None

<!-- CANARY: REQ=CBIN-133; FEATURE="RequirementLookup"; ASPECT=API; STATUS=STUB; UPDATED=2025-10-16 -->
**Feature 2: Requirement Lookup Logic**
- [ ] Implement `findRequirement(query string) (*Requirement, error)`
- [ ] Support exact ID match (CBIN-XXX)
- [ ] Support directory name match
- [ ] Load spec.md and plan.md files
- **Location hint:** `cmd/canary/implement.go` (new file, similar to `next.go`)
- **Dependencies:** None

### Phase 2: Fuzzy Matching (IMPL → TESTED)

<!-- CANARY: REQ=CBIN-133; FEATURE="FuzzyMatcher"; ASPECT=Engine; STATUS=STUB; UPDATED=2025-10-16 -->
**Feature 3: Fuzzy String Matching**
- [ ] Implement Levenshtein distance algorithm (or use library)
- [ ] Implement scoring function (0-100%)
- [ ] Implement substring and abbreviation matching
- [ ] Return top N matches with scores
- **Location hint:** `internal/matcher/fuzzy.go` (new file)
- **Dependencies:** None

<!-- CANARY: REQ=CBIN-133; FEATURE="InteractiveSelection"; ASPECT=CLI; STATUS=STUB; UPDATED=2025-10-16 -->
**Feature 4: Interactive Selection UI**
- [ ] Display numbered list of matches
- [ ] Accept numeric input
- [ ] Validate and re-prompt on invalid input
- [ ] Handle quit command
- **Location hint:** `cmd/canary/implement.go`
- **Dependencies:** FuzzyMatcher

### Phase 3: Prompt Generation (TESTED → BENCHED)

<!-- CANARY: REQ=CBIN-133; FEATURE="PromptGenerator"; ASPECT=API; STATUS=STUB; UPDATED=2025-10-16 -->
**Feature 5: Implementation Prompt Template**
- [ ] Create `.canary/templates/implement-prompt-template.md`
- [ ] Include sections: spec, plan, constitution, tokens, tests, success criteria
- [ ] Use Go text/template syntax
- **Location hint:** `.canary/templates/implement-prompt-template.md` (new file)
- **Dependencies:** None

<!-- CANARY: REQ=CBIN-133; FEATURE="PromptRenderer"; ASPECT=API; STATUS=STUB; UPDATED=2025-10-16 -->
**Feature 6: Prompt Rendering Engine**
- [ ] Load template file
- [ ] Populate template variables from spec/plan/constitution
- [ ] Extract Implementation Checklist from spec
- [ ] Generate CANARY token examples
- [ ] Render final prompt
- **Location hint:** `cmd/canary/implement.go`
- **Dependencies:** PromptGenerator

### Phase 4: Progress Tracking

<!-- CANARY: REQ=CBIN-133; FEATURE="ProgressTracker"; ASPECT=Engine; STATUS=STUB; UPDATED=2025-10-16 -->
**Feature 7: Implementation Progress Calculation**
- [ ] Scan codebase for CANARY tokens matching requirement ID
- [ ] Count tokens by status
- [ ] Calculate completion percentage
- [ ] Include in prompt header
- **Location hint:** `cmd/canary/implement.go`
- **Dependencies:** None (uses existing grep/scanner logic)

### Phase 5: Slash Command Integration

<!-- CANARY: REQ=CBIN-133; FEATURE="ImplementSlashCmd"; ASPECT=CLI; STATUS=STUB; UPDATED=2025-10-16 -->
**Feature 8: Slash Command Template**
- [ ] Create `.canary/templates/commands/implement.md`
- [ ] Include usage examples
- [ ] Document workflow (specify → plan → implement)
- **Location hint:** `.canary/templates/commands/implement.md` (new file)
- **Dependencies:** None

### Testing Requirements

<!-- CANARY: REQ=CBIN-133; FEATURE="ImplementCmdTests"; ASPECT=CLI; STATUS=STUB; TEST=TestCANARY_CBIN_133_CLI_ImplementCommand; UPDATED=2025-10-16 -->
**Unit Tests:**
- [ ] Test exact ID lookup
- [ ] Test fuzzy matching with various queries
- [ ] Test interactive selection
- [ ] Test prompt generation with mock spec/plan
- [ ] Test error handling (missing spec, invalid ID)
- **Location hint:** `cmd/canary/implement_test.go` (new file)

<!-- CANARY: REQ=CBIN-133; FEATURE="FuzzyMatcherTests"; ASPECT=Engine; STATUS=STUB; TEST=TestCANARY_CBIN_133_Engine_FuzzyMatching; UPDATED=2025-10-16 -->
**Fuzzy Matcher Tests:**
- [ ] Test Levenshtein distance calculation
- [ ] Test scoring accuracy
- [ ] Test abbreviation matching
- [ ] Test threshold filtering
- **Location hint:** `internal/matcher/fuzzy_test.go` (new file)

---

## CANARY Tokens Reference

**Main requirement token** (add to primary implementation file):
```
// CANARY: REQ=CBIN-133; FEATURE="ImplementCommand"; ASPECT=CLI; STATUS=STUB; UPDATED=2025-10-16
```

**Example sub-feature token**:
```
// CANARY: REQ=CBIN-133; FEATURE="FuzzyMatcher"; ASPECT=Engine; STATUS=IMPL; TEST=TestCANARY_CBIN_133_Engine_FuzzyMatching; UPDATED=2025-10-16
```

---

**Next Steps:**
1. Run: `/canary.plan CBIN-133` to create technical implementation plan
2. Run: `/canary.implement CBIN-133` to begin implementation (once plan exists)
3. Follow test-first approach (Article IV)
