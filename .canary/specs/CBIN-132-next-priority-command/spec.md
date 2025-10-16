# Feature Specification: Next Priority Command

**Requirement ID:** CBIN-132
**Feature Name:** NextPriorityCommand
**Status:** STUB
**Owner:** canary
**Created:** 2025-10-16
**Updated:** 2025-10-16

---

## Feature Overview

### Summary
Provide a `/canary.next` slash command and `canary next` CLI command that automatically identifies the highest priority unimplemented requirement and generates a comprehensive implementation prompt for coding agents. The command should leverage Go templates to create expert-level guidance that includes specification details, constitutional principles, and contextual references to ensure safe and correct implementation.

### Business Value
- **Reduced Context Switching**: Agents automatically know what to work on next without manual prioritization
- **Consistent Implementation**: Template-driven prompts ensure all requirements are implemented following project standards
- **Constitutional Adherence**: Every implementation references governing principles from constitution.md
- **Reduced Errors**: Comprehensive prompts reduce ambiguity and implementation mistakes
- **Priority Management**: Focuses development effort on highest-value requirements first

### Target Users
- AI coding agents (Claude Code, Cursor, Copilot, etc.)
- Human developers using the CANARY workflow
- CI/CD automation systems

---

## User Stories

### US-1: Agent Requests Next Task
**As an** AI coding agent
**I want to** request the next highest priority requirement to implement
**So that** I can automatically proceed with the most valuable work without user intervention

**Acceptance Criteria:**
- Agent types `/canary.next` in their interface
- System identifies highest priority STUB or IMPL requirement
- System generates comprehensive implementation prompt
- Prompt includes specification details, constitution references, and related specs
- Agent receives actionable, unambiguous guidance

### US-2: Developer Queries Priority Queue
**As a** human developer
**I want to** see what the next highest priority requirement is
**So that** I can make informed decisions about what to work on next

**Acceptance Criteria:**
- Developer runs `canary next` command
- System displays requirement ID, feature name, priority, and brief description
- System shows estimated complexity and dependencies
- Developer can optionally generate full implementation prompt

### US-3: CI/CD Automation
**As a** CI/CD pipeline
**I want to** query the next priority requirement programmatically
**So that** automated agents can implement requirements without manual intervention

**Acceptance Criteria:**
- Command supports `--json` flag for machine-readable output
- Output includes all template variables needed for prompt generation
- Exit codes indicate success/no-work-available states

---

## Functional Requirements

### FR-1: Priority Determination
The system shall determine priority based on:
1. Explicit PRIORITY field in CANARY tokens (1=highest, 10=lowest)
2. STATUS (STUB > IMPL > TESTED)
3. DEPENDS_ON relationships (dependencies must be completed first)
4. UPDATED field (older requirements get priority boost)

### FR-2: Prompt Template Engine
The system shall use Go's `text/template` package to generate prompts with:
- Requirement specification content
- Constitutional principles (from `.canary/memory/constitution.md`)
- Related specifications (DEPENDS_ON, BLOCKS, RELATED_TO)
- Test-first approach guidance
- Token placement examples
- Success criteria from spec

### FR-3: Template Variables
The prompt template shall have access to:
```go
type PromptData struct {
    ReqID          string
    Feature        string
    Aspect         string
    Status         string
    Priority       int
    SpecFile       string
    SpecContent    string
    Constitution   string
    RelatedSpecs   []RelatedSpec
    Dependencies   []Requirement
    SuggestedFiles []string
    TestGuidance   string
    TokenExample   string
}
```

### FR-4: Slash Command Integration
- Create `.canary/templates/commands/next.md` template
- Install to all supported agents (`.claude/`, `.cursor/`, etc.)
- Command invokes `canary next --prompt` to generate implementation guidance

### FR-5: CLI Command
```bash
canary next [flags]

Flags:
  --prompt          Generate full implementation prompt (default: false)
  --json            Output in JSON format
  --status string   Filter by status (STUB,IMPL,TESTED)
  --aspect string   Filter by aspect
  --dry-run         Show what would be selected without generating prompt
```

### FR-6: Database Integration
The command shall:
- Query `.canary/canary.db` for priorities if available
- Fall back to scanning filesystem if database doesn't exist
- Use `canary index` workflow if database is stale

### FR-7: Error Handling
- No requirements available → Display helpful message and exit 0
- All requirements completed → Congratulatory message and exit 0
- Database unavailable → Fall back to filesystem scan
- Template errors → Display clear error with template line number

---

## Success Criteria

### Measurability
1. **Query Performance**: Next priority identified in <100ms for codebases with <10,000 requirements
2. **Prompt Generation**: Full prompt generated in <500ms including all template rendering
3. **Accuracy**: 100% of generated prompts include non-empty specification content
4. **Constitutional Adherence**: 100% of prompts reference at least 2 constitutional principles
5. **Completeness**: Generated prompts include all required sections (spec, constitution, examples, tests)

### Quality Gates
- [ ] Command works without database (filesystem fallback)
- [ ] All template variables are validated before rendering
- [ ] Prompt includes test-first guidance from Article IV
- [ ] Dependency requirements are validated before selection
- [ ] Generated prompts are 2,000-5,000 words (comprehensive but not overwhelming)

---

## Implementation Checklist

### Phase 1: CLI Command (STUB → IMPL)
- [ ] Add `nextCmd` to `cmd/canary/main.go`
- [ ] Implement priority query logic
- [ ] Add database fallback to filesystem scan
- [ ] Create flag parsing and validation

### Phase 2: Template Engine (IMPL → TESTED)
- [ ] Create `.canary/templates/next-prompt-template.md`
- [ ] Implement `text/template` rendering
- [ ] Load specification files
- [ ] Load constitution content
- [ ] Resolve DEPENDS_ON relationships
- [ ] Generate test examples

### Phase 3: Slash Command Integration (TESTED → BENCHED)
- [ ] Create `.canary/templates/commands/next.md`
- [ ] Install to agent directories via `canary init`
- [ ] Test with Claude Code
- [ ] Test with Cursor
- [ ] Document workflow in CLAUDE.md

---

## Acceptance Scenarios

### Scenario 1: Happy Path
**Given:** Database contains 5 STUB requirements with priorities 1-5
**When:** Agent runs `/canary.next`
**Then:**
- System selects requirement with PRIORITY=1
- Generates prompt with specification content
- Includes constitution references
- Provides test-first guidance
- Includes token placement example

### Scenario 2: Dependency Blocking
**Given:** Highest priority requirement (CBIN-105, PRIORITY=1) has DEPENDS_ON=CBIN-104
**And:** CBIN-104 has STATUS=STUB
**When:** System queries next priority
**Then:**
- System skips CBIN-105 (blocked)
- Selects CBIN-104 (dependency must be resolved first)
- Includes note in prompt about CBIN-105 being unblocked after completion

### Scenario 3: No Work Available
**Given:** All requirements have STATUS=TESTED or STATUS=BENCHED
**When:** Agent runs `/canary.next`
**Then:**
- System displays: "🎉 All requirements completed! No work available."
- Suggests running `canary scan --verify GAP_ANALYSIS.md`
- Exits with code 0

### Scenario 4: Database Unavailable
**Given:** `.canary/canary.db` does not exist
**When:** Agent runs `canary next`
**Then:**
- System displays: "Database not found, scanning filesystem..."
- Falls back to filesystem scan
- Selects highest priority from CANARY tokens
- Generates prompt normally

---

## Assumptions and Constraints

### Assumptions
1. Specifications exist in `.canary/specs/CBIN-XXX-*/spec.md` format
2. Constitution file exists at `.canary/memory/constitution.md`
3. CANARY tokens follow standard format with PRIORITY field (or default to 5)
4. Agents can process markdown-formatted prompts
5. Go `text/template` package is available

### Constraints
1. Prompt generation must complete in <1 second for responsive agent experience
2. Template must be maintainable by non-Go developers (readable markdown with {{.Variable}} syntax)
3. Cannot require database (must work with filesystem-only scans)
4. Must work across all supported agents (agent-agnostic prompt format)
5. Prompt must fit within agent context windows (max 5,000 words)

### Technical Constraints
- Go 1.19+ required for `text/template` features
- Filesystem access required for spec file loading
- UTF-8 encoding for all template files

---

## Related Requirements

**DEPENDS_ON:**
- CBIN-124: IndexCmd (for database priority queries)
- CBIN-125: ListCmd (for filtering and ordering logic)

**BLOCKS:**
- None (this is a new feature)

**RELATED_TO:**
- CBIN-110: SpecifyCmd (shares specification loading logic)
- CBIN-121: PlanCmd (shares template rendering approach)
- CBIN-106: AgentContext (provides context for agent integration)

---

## Open Questions

None - requirement is fully specified.

---

## References

- [CANARY Constitutional Principles](./../memory/constitution.md)
- [Template Syntax](https://pkg.go.dev/text/template)
- [Priority Management Design](../../docs/PRIORITY_DESIGN.md)
