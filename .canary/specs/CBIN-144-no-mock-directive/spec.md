<!-- CANARY: REQ=CBIN-144; FEATURE="NoMockDirective"; ASPECT=Planner; STATUS=STUB; UPDATED=2025-10-17 -->
# Feature Specification: No Mock Data Implementation Directive

**Requirement ID:** CBIN-144
**Status:** STUB
**Created:** 2025-10-17
**Last Updated:** 2025-10-17

## Overview

**Purpose:** Ensure all AI agent prompts across CANARY slash commands include explicit directives prohibiting mock data, simulations, and placeholders, forcing agents to implement complete, production-ready features with full complexity from the start. This prevents a common anti-pattern where agents create stub implementations, leaving actual work for humans.

**Scope:** This feature adds a standardized "Implementation Quality Directive" section to all slash command templates (`/canary.specify`, `/canary.plan`, etc.) that explicitly instructs agents to avoid shortcuts, implement real functionality, and handle edge cases. Includes template updates, validation that directives are present, and documentation of when exceptions are acceptable. Excludes enforcement at runtime (relies on prompt engineering).

## User Stories

### Primary User Stories

**US-1: Prevent Mock Implementations**
As a project maintainer,
I want all slash command prompts to explicitly forbid mock data and simulations,
So that agents deliver complete, production-ready implementations instead of placeholders.

**Acceptance Criteria:**
- [ ] All slash command templates include "NO MOCK DATA, NO SIMULATION, IMPLEMENT THE COMPLEX FEATURES" directive
- [ ] Directive appears prominently in agent prompts before task instructions
- [ ] Directive is mandatory; templates without it fail validation
- [ ] Directive text is standardized across all commands

**US-2: Enforce Complete Implementations**
As an AI agent,
I want clear instructions about implementation expectations,
So that I understand I must deliver full functionality, not stubs or TODO comments.

**Acceptance Criteria:**
- [ ] Directive explicitly lists prohibited shortcuts (mock data, TODO markers, placeholder functions)
- [ ] Directive specifies what constitutes acceptable implementation (working code, error handling, edge cases)
- [ ] Directive provides examples of complete vs. incomplete implementations
- [ ] Agent prompts reinforce directive at multiple points in workflow

**US-3: Validate Template Compliance**
As a maintainer,
I want to verify that all templates contain the implementation directive,
So that I can catch missing or weakened directives during development.

**Acceptance Criteria:**
- [ ] `canary validate-templates` command checks for directive presence
- [ ] Validation fails if directive is missing or improperly formatted
- [ ] CI/CD integration ensures new templates include directive
- [ ] Validation reports which templates lack directive

**US-4: Document Acceptable Exceptions**
As a developer,
I want to know when mock data or stubs are acceptable,
So that I can make appropriate trade-offs during prototyping or testing.

**Acceptance Criteria:**
- [ ] Documentation clearly states when exceptions are allowed (tests, examples, prototypes)
- [ ] Exception cases are explicitly called out in directive
- [ ] Agents can request clarification when unsure if exception applies
- [ ] Guidelines prevent exception abuse

### Secondary User Stories (if applicable)

**US-5: Strengthen Directive Over Time**
As a maintainer,
I want to refine the directive based on observed agent behavior,
So that it adapts to new patterns of incomplete implementations.

**Acceptance Criteria:**
- [ ] Directive is version-controlled with change history
- [ ] Can add specific prohibitions as new anti-patterns emerge
- [ ] Directive updates propagate to all templates automatically
- [ ] Changes are documented with rationale

## Functional Requirements

### FR-1: Standardized Directive Text
**Priority:** High
**Description:** System must define a canonical "Implementation Quality Directive" that prohibits mock data, simulations, placeholders, and TODO markers while requiring complete, production-ready implementations.
**Acceptance:** Directive text is stored in single authoritative location; all templates reference this directive; directive is version-controlled; changes update all templates.

### FR-2: Template Integration
**Priority:** High
**Description:** All slash command templates must include the Implementation Quality Directive in a prominent, consistent location within agent prompts.
**Acceptance:** Directive appears in all templates (specify.md, plan.md, scan.md, verify.md, constrain.md, etc.); directive is placed before task-specific instructions; directive uses consistent formatting; templates fail to load if directive is missing.

### FR-3: Directive Content
**Priority:** High
**Description:** Directive must explicitly prohibit common shortcuts and define expectations for complete implementations.
**Acceptance:** Directive lists prohibited items (mock data, simulations, TODO/FIXME, placeholder functions, stub classes, hardcoded values); directive defines required elements (error handling, edge cases, validation, real data processing); directive uses clear, imperative language.

### FR-4: Template Validation
**Priority:** Medium
**Description:** System must provide validation command to verify all templates contain proper implementation directive.
**Acceptance:** `canary validate-templates` scans all command templates; reports missing or malformed directives; exits with error code if validation fails; suitable for CI/CD integration.

### FR-5: Exception Documentation
**Priority:** Medium
**Description:** Directive must clearly document when mock data or incomplete implementations are acceptable.
**Acceptance:** Exceptions listed in directive (unit tests, documentation examples, proof-of-concept prototypes); exceptions are specific and limited; directive warns against exception abuse; agents can check if their case qualifies.

### FR-6: Multi-Point Reinforcement
**Priority:** Low
**Description:** Directive should be reinforced at multiple points in agent workflow to prevent agents from forgetting or ignoring it.
**Acceptance:** Directive appears in initial prompt, before implementation phase, and in summary/validation sections; repetition is non-intrusive but consistent; key phrases are emphasized.

## Success Criteria

**Quantitative Metrics:**
- [ ] 100% of slash command templates include implementation directive
- [ ] 90% reduction in mock/stub implementations after directive deployment
- [ ] Zero instances of "TODO: implement this" in agent-generated code within 1 month
- [ ] Template validation runs in under 2 seconds

**Qualitative Measures:**
- [ ] Agents consistently deliver complete implementations on first attempt
- [ ] Code reviews find fewer "this is just a placeholder" issues
- [ ] Maintainers spend less time requesting agents to "actually implement" features
- [ ] Directive wording is clear and unambiguous to both humans and agents

## User Scenarios & Testing

### Scenario 1: Adding Directive to Existing Template
**Given:** `/canary.plan` template exists without implementation directive
**When:** Maintainer updates template to include directive at top of prompt section
**Then:** Directive reads "IMPLEMENTATION QUALITY DIRECTIVE: NO MOCK DATA, NO SIMULATION, IMPLEMENT THE COMPLEX FEATURES..."; appears before planning instructions; validation passes

### Scenario 2: Agent Receives Directive
**Given:** Agent executes `/canary.plan CBIN-050`
**When:** Slash command expands with directive in prompt
**Then:** Agent sees directive before task instructions; directive explicitly prohibits stubs, mocks, placeholders; agent implements complete solution with error handling and edge cases

### Scenario 3: Validating Templates
**Given:** Project has 10 slash command templates
**When:** Developer runs `canary validate-templates`
**Then:** System scans all .md files in `.canary/templates/commands/`; reports "9/10 templates valid, 1 missing directive: scan.md"; exits with code 1; output shows exact problem

### Scenario 4: CI/CD Integration
**Given:** New slash command template added in pull request
**When:** CI runs `canary validate-templates` as part of checks
**Then:** Validation detects missing directive; CI fails with clear message "Template validation failed: new-command.md missing Implementation Quality Directive"; PR cannot merge until fixed

### Scenario 5: Exception Case - Unit Tests
**Given:** Agent is creating unit tests for CBIN-055
**When:** Agent reads directive with exception clause
**Then:** Directive states "Exception: Unit tests may use mock objects for external dependencies"; agent uses mock HTTP client in tests; production code remains fully implemented

### Scenario 6: Strengthening Directive
**Given:** Agents repeatedly use hardcoded example URLs in implementations
**When:** Maintainer updates directive to explicitly prohibit "hardcoded example.com URLs"
**Then:** Directive version increments; all templates auto-update; future agents avoid this pattern

## Key Entities (if data-driven feature)

### Entity 1: ImplementationDirective
**Attributes:**
- directive_id: Unique identifier
- version: Version number (semver)
- directive_text: Full text of directive
- prohibited_items: List of banned shortcuts (mock data, TODOs, etc.)
- required_elements: List of mandatory implementation aspects (error handling, validation)
- exceptions: Cases where incomplete implementation is acceptable
- updated_at: When directive was last modified
- change_reason: Why directive was updated

**Relationships:**
- Referenced by all SlashCommandTemplates
- Versioned for tracking evolution

### Entity 2: DirectiveViolation
**Attributes:**
- violation_id: Unique identifier
- requirement_id: Which CANARY requirement had violation
- violation_type: Type of shortcut used (mock_data, todo_marker, placeholder, etc.)
- detected_at: When violation was found
- file_path: Where violation occurred
- evidence: Code snippet showing violation
- resolved: Whether violation was fixed

**Relationships:**
- Links to CANARY requirement
- May link to gap analysis for learning

## Assumptions

- Agents process and respect prominent directives in prompts
- Directive enforcement relies on prompt engineering, not code analysis
- Most violations are accidental (agents default to shortcuts) rather than intentional
- Directive needs periodic updates as new anti-patterns emerge
- Exceptions are rare and well-defined
- Templates are stored as markdown files in known location

## Constraints

**Technical Constraints:**
- Directive must be included in prompt without exceeding token limits
- Validation must work with various template formats
- Directive text must be compatible with all slash commands

**Business Constraints:**
- Directive should not significantly lengthen prompts
- Validation must be fast enough for CI/CD pipelines
- Updates to directive should be easy to deploy

**Regulatory Constraints:**
- None identified

## Out of Scope

- Automated detection of mock data in generated code (relies on code review)
- Runtime enforcement of directive (prompt-only)
- Penalties for agents that violate directive
- Machine learning to detect new types of incomplete implementations
- Integration with external code quality tools
- Directive translations to other languages
- Dynamic directive generation based on context

## Dependencies

- Slash command template system in `.canary/templates/commands/`
- All existing slash commands (specify, plan, scan, verify, constrain, etc.)
- Template loading mechanism
- (Optional) CI/CD pipeline for validation

## Risks & Mitigation

| Risk | Impact | Probability | Mitigation |
|------|--------|-------------|------------|
| Agents ignore directive despite prominence | High | Medium | Multi-point reinforcement; track violations; strengthen wording |
| Directive is too restrictive for valid use cases | Medium | Medium | Document clear exceptions; allow clarification requests |
| Directive adds too many tokens to prompts | Low | Low | Keep directive concise; measure token impact |
| Validation gives false positives | Medium | Low | Test validation thoroughly; allow override flag for special cases |
| Directive becomes outdated as AI capabilities evolve | Medium | Medium | Regular reviews; version tracking; community feedback |

## Clarifications Resolved

**Clarification 1: Directive Placement**
**Decision:** Directive appears at the beginning of agent prompt, before task-specific instructions
**Rationale:** Ensures agents see directive before starting work; establishes expectations upfront; reduces chance of being overlooked in long prompts.

**Clarification 2: Exception Scope**
**Decision:** Exceptions limited to: (1) unit tests with mock external dependencies, (2) documentation examples, (3) explicit prototypes marked as such
**Rationale:** Narrow exceptions prevent abuse while allowing legitimate use cases; agents can easily determine if exception applies; production code always requires full implementation.

**Clarification 3: Validation Enforcement**
**Decision:** Validation is mandatory in CI/CD; blocks merges if templates lack directive
**Rationale:** Ensures directive never accidentally removed; prevents new templates without directive; maintains consistency across all commands.

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

<!-- CANARY: REQ=CBIN-144; FEATURE="DirectiveDefinition"; ASPECT=Docs; STATUS=STUB; UPDATED=2025-10-17 -->
**Feature 1: Canonical Directive Definition**
- [ ] Create authoritative directive text in single source file
- **Location hint:** ".canary/templates/directives/implementation-quality.md"
- **Dependencies:** None

<!-- CANARY: REQ=CBIN-144; FEATURE="DirectiveContent"; ASPECT=Docs; STATUS=STUB; UPDATED=2025-10-17 -->
**Feature 2: Directive Content**
- [ ] Write full directive text with prohibited items, required elements, and exceptions
- **Location hint:** ".canary/templates/directives/implementation-quality.md"
- **Dependencies:** Directive definition

<!-- CANARY: REQ=CBIN-144; FEATURE="TemplateIntegration"; ASPECT=Planner; STATUS=STUB; UPDATED=2025-10-17 -->
**Feature 3: Template Integration**
- [ ] Update all slash command templates to include directive
- **Location hint:** ".canary/templates/commands/*.md"
- **Dependencies:** Directive content

<!-- CANARY: REQ=CBIN-144; FEATURE="SpecifyIntegration"; ASPECT=Planner; STATUS=STUB; UPDATED=2025-10-17 -->
**Feature 4: Specify Command Integration**
- [ ] Add directive to `/canary.specify` template
- **Location hint:** ".canary/templates/commands/specify.md"
- **Dependencies:** Template integration

<!-- CANARY: REQ=CBIN-144; FEATURE="PlanIntegration"; ASPECT=Planner; STATUS=STUB; UPDATED=2025-10-17 -->
**Feature 5: Plan Command Integration**
- [ ] Add directive to `/canary.plan` template
- **Location hint:** ".canary/templates/commands/plan.md"
- **Dependencies:** Template integration

<!-- CANARY: REQ=CBIN-144; FEATURE="AllCommandsIntegration"; ASPECT=Planner; STATUS=STUB; UPDATED=2025-10-17 -->
**Feature 6: Remaining Commands Integration**
- [ ] Add directive to all other slash command templates (scan, verify, constrain, etc.)
- **Location hint:** ".canary/templates/commands/*.md"
- **Dependencies:** Template integration

### Validation

<!-- CANARY: REQ=CBIN-144; FEATURE="TemplateValidator"; ASPECT=CLI; STATUS=STUB; UPDATED=2025-10-17 -->
**Template Validation Command:**
- [ ] Implement `canary validate-templates` CLI command
- **Location hint:** "cmd/validate_templates.go"
- **Dependencies:** Template loading

<!-- CANARY: REQ=CBIN-144; FEATURE="DirectiveDetection"; ASPECT=Engine; STATUS=STUB; UPDATED=2025-10-17 -->
**Directive Detection Logic:**
- [ ] Implement logic to scan templates for directive presence
- **Location hint:** "internal/templates/validator.go"
- **Dependencies:** Template validator

<!-- CANARY: REQ=CBIN-144; FEATURE="ValidationReporting"; ASPECT=CLI; STATUS=STUB; UPDATED=2025-10-17 -->
**Validation Reporting:**
- [ ] Generate reports showing which templates pass/fail directive check
- **Location hint:** "internal/templates/validator.go"
- **Dependencies:** Directive detection

### Testing Requirements

<!-- CANARY: REQ=CBIN-144; FEATURE="DirectiveTests"; ASPECT=Planner; STATUS=STUB; TEST=TestDirectivePresence; UPDATED=2025-10-17 -->
**Unit Tests:**
- [ ] Test directive presence in all templates
- **Location hint:** "internal/templates/validator_test.go"

<!-- CANARY: REQ=CBIN-144; FEATURE="ValidationTests"; ASPECT=CLI; STATUS=STUB; TEST=TestTemplateValidation; UPDATED=2025-10-17 -->
**Validation Tests:**
- [ ] Test validation command with valid and invalid templates
- **Location hint:** "cmd/validate_templates_test.go"

<!-- CANARY: REQ=CBIN-144; FEATURE="IntegrationTests"; ASPECT=CLI; STATUS=STUB; TEST=TestDirectiveInPrompts; UPDATED=2025-10-17 -->
**Integration Tests:**
- [ ] Test that directives appear in expanded slash command prompts
- **Location hint:** "test/integration/directive_test.go"

### Documentation

<!-- CANARY: REQ=CBIN-144; FEATURE="DirectiveDocs"; ASPECT=Docs; STATUS=STUB; UPDATED=2025-10-17 -->
**Directive Documentation:**
- [ ] Document the implementation quality directive and its purpose
- **Location hint:** "docs/concepts/implementation-quality.md"

<!-- CANARY: REQ=CBIN-144; FEATURE="ExceptionDocs"; ASPECT=Docs; STATUS=STUB; UPDATED=2025-10-17 -->
**Exception Guidelines:**
- [ ] Document when mock data and stubs are acceptable
- **Location hint:** "docs/guides/implementation-exceptions.md"

<!-- CANARY: REQ=CBIN-144; FEATURE="AgentGuidance"; ASPECT=Docs; STATUS=STUB; UPDATED=2025-10-17 -->
**Agent Guidance:**
- [ ] Create guide for agents on meeting implementation quality standards
- **Location hint:** ".canary/AGENT_CONTEXT.md"

---

**Agent Instructions:**

After implementing each feature:
1. Update the CANARY token in the spec from `STATUS=STUB` to `STATUS=IMPL`
2. Add the same token to your source code at the implementation location
3. Add `TEST=TestName` when tests are written
4. Run `canary implement CBIN-144` to see implementation progress

---

## CANARY Tokens Reference

**Main requirement token** (add to primary implementation file):
```
// CANARY: REQ=CBIN-144; FEATURE="NoMockDirective"; ASPECT=Planner; STATUS=IMPL; UPDATED=2025-10-17
```

**Directive text example** (in `.canary/templates/directives/implementation-quality.md`):
```markdown
# IMPLEMENTATION QUALITY DIRECTIVE

**NO MOCK DATA, NO SIMULATION, IMPLEMENT THE COMPLEX FEATURES**

When implementing features, you MUST provide complete, production-ready implementations:

**PROHIBITED:**
- Mock data, simulated responses, or fake values
- TODO, FIXME, or "implement this later" comments
- Placeholder functions that return hardcoded values
- Stub classes with empty methods
- Example.com, test@example.com, or other placeholder data
- Comments saying "this would be implemented in production"

**REQUIRED:**
- Full working implementations with real logic
- Comprehensive error handling for edge cases
- Input validation and boundary checking
- Proper data processing and transformations
- Complete integration with existing systems
- Production-quality code, not prototypes

**EXCEPTIONS:**
The following are the ONLY acceptable uses of mocks/stubs:
1. Unit tests mocking external dependencies (databases, APIs, file systems)
2. Documentation examples explicitly marked as "example only"
3. Prototypes explicitly labeled as proof-of-concept (not production code)

If unsure whether your case qualifies as an exception, ask for clarification.

**REMEMBER:** Shortcuts slow down development. Implement features completely the first time.
```

**Template integration example** (in `.canary/templates/commands/plan.md`):
```markdown
# Planning Implementation for CBIN-XXX

{{include:.canary/templates/directives/implementation-quality.md}}

## Your Task

Given the requirement CBIN-XXX, create a detailed implementation plan...
```

**Use `canary implement CBIN-144` to find:**
- Which features are implemented vs. still TODO
- Exact file locations and line numbers
- Context around each implementation point
