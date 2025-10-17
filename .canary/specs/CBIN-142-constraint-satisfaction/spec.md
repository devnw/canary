<!-- CANARY: REQ=CBIN-142; FEATURE="ConstraintSatisfaction"; ASPECT=CLI; STATUS=STUB; UPDATED=2025-10-17 -->
# Feature Specification: Constraint Satisfaction System

**Requirement ID:** CBIN-142
**Status:** STUB
**Created:** 2025-10-17
**Last Updated:** 2025-10-17

## Overview

**Purpose:** Enable project maintainers to define and enforce hard constraints that AI agents must satisfy across all CANARY operations (specify, plan, implement, verify, etc.). By modeling agent behavior as a constraint satisfaction problem (CSP), the system ensures agents operate within defined boundaries for intent, context reduction, and expected outputs, leading to more predictable and correct implementations.

**Scope:** This feature adds a `constrain` subcommand and `/canary.constrain` slash command to create, manage, and enforce project-wide constraints. Constraints are injected into all slash command templates and agent prompts, ensuring every operation respects defined rules. Includes constraint validation, conflict detection, and priority management. Excludes automatic constraint inference or machine learning-based constraint generation.

## User Stories

### Primary User Stories

**US-1: Define Project Constraints**
As a project maintainer,
I want to define hard constraints that agents must follow during all operations,
So that agent behavior remains predictable and aligns with project goals and architectural decisions.

**Acceptance Criteria:**
- [ ] Maintainer can create constraints using `canary constrain add` with categories (intent, context, output, security, performance)
- [ ] Each constraint has priority (MUST, SHOULD, MAY), rationale, and validation criteria
- [ ] Constraints are persisted and versioned
- [ ] System prevents duplicate or conflicting constraints

**US-2: Automatic Constraint Injection**
As an AI agent,
I want constraints automatically included in every slash command prompt,
So that I don't violate project rules when creating specs, plans, or implementations.

**Acceptance Criteria:**
- [ ] All slash commands (`/canary.specify`, `/canary.plan`, etc.) automatically inject relevant constraints
- [ ] Constraints appear in agent context before task instructions
- [ ] Constraint injection is conditional based on operation type (e.g., security constraints for API features)
- [ ] Agent can query active constraints for current operation

**US-3: Validate Agent Outputs**
As a project maintainer,
I want to validate that agent outputs satisfy defined constraints,
So that I can catch violations early and provide feedback to improve future operations.

**Acceptance Criteria:**
- [ ] System can validate spec, plan, or implementation against constraints
- [ ] Validation reports which constraints were satisfied and which were violated
- [ ] Violations include specific evidence and suggested remediation
- [ ] Validation can be run manually or automatically via hooks

**US-4: Manage Constraint Evolution**
As a project maintainer,
I want to update, disable, or remove constraints as the project evolves,
So that constraints remain relevant and don't become technical debt.

**Acceptance Criteria:**
- [ ] Maintainer can list, view, edit, disable, and remove constraints
- [ ] Disabling a constraint keeps it in history but stops enforcement
- [ ] Constraint changes are tracked with timestamps and reasons
- [ ] System warns when removing constraints that are actively referenced

### Secondary User Stories (if applicable)

**US-5: Constraint Conflict Detection**
As a project maintainer,
I want to be warned when new constraints conflict with existing ones,
So that I don't create impossible-to-satisfy constraint sets.

**Acceptance Criteria:**
- [ ] System detects logical conflicts between constraints
- [ ] Warns when adding constraint that contradicts existing MUST constraint
- [ ] Provides recommendations for resolving conflicts

**US-6: Explain Constraint Violations**
As an AI agent,
I want clear explanations when I violate a constraint,
So that I can understand what went wrong and adjust my approach.

**Acceptance Criteria:**
- [ ] Violation messages include constraint text, why it was violated, and examples of valid outputs
- [ ] Agent can request constraint clarification before proceeding
- [ ] Explanations are added to gap analysis for learning

## Functional Requirements

### FR-1: Constraint Definition and Storage
**Priority:** High
**Description:** System must provide CLI commands to create, edit, list, and delete constraints with categories, priorities, validation rules, and metadata.
**Acceptance:** User can run `canary constrain add/edit/list/remove`; constraints are persisted in `.canary/constraints/` directory; schema supports all required fields.

### FR-2: Constraint Categorization
**Priority:** High
**Description:** System must support constraint categories aligned with CSP dimensions: intent (what to build), context (how much information to consider), output (format and structure), security (safety requirements), and performance (efficiency bounds).
**Acceptance:** Each constraint belongs to one or more categories; categories determine when constraints are applied; filtering by category is supported.

### FR-3: Constraint Injection into Templates
**Priority:** High
**Description:** System must automatically inject relevant constraints into all slash command templates when commands are executed by agents.
**Acceptance:** Templates in `.canary/templates/commands/` reference constraint system; constraints appear in agent prompts; injection is conditional based on constraint applicability.

### FR-4: Constraint Validation Engine
**Priority:** High
**Description:** System must validate agent outputs (specs, plans, implementations) against applicable constraints and report satisfaction/violation results.
**Acceptance:** `canary constrain validate <file>` checks output against constraints; reports pass/fail for each constraint; provides actionable feedback on violations.

### FR-5: Constraint Prioritization
**Priority:** Medium
**Description:** System must support constraint priorities (MUST, SHOULD, MAY) with different enforcement levels and violation handling.
**Acceptance:** MUST violations block operations; SHOULD violations warn but allow; MAY violations are suggestions; priority is configurable per constraint.

### FR-6: Constraint Documentation
**Priority:** Medium
**Description:** System must generate human-readable documentation of active constraints for team reference.
**Acceptance:** `canary constrain docs` generates markdown documentation; includes rationale, examples, and validation criteria; can be included in project README.

### FR-7: Constraint Versioning
**Priority:** Low
**Description:** System must track constraint history to understand how rules evolved over time.
**Acceptance:** Each constraint change creates a version entry; history includes timestamp, author, and change reason; can view constraint state at specific dates.

## Success Criteria

**Quantitative Metrics:**
- [ ] 100% of slash commands inject relevant constraints automatically
- [ ] Constraint validation completes in under 5 seconds for typical outputs
- [ ] Agents can retrieve active constraints in under 1 second
- [ ] 90% reduction in architectural violations after constraint enforcement

**Qualitative Measures:**
- [ ] Agents consistently respect project architectural decisions
- [ ] Constraint violations are caught before implementation begins
- [ ] Team members understand project rules through constraint documentation
- [ ] Constraints reduce need for repeated manual code reviews on same issues

## User Scenarios & Testing

### Scenario 1: Adding Intent Constraint
**Given:** Project requires all API features to use REST principles
**When:** Maintainer runs `canary constrain add --category intent --priority MUST --name "REST-API-Design" --rule "All API endpoints must follow RESTful principles with resource-based URLs, standard HTTP methods, and stateless operations"`
**Then:** System creates constraint; subsequent `/canary.plan` operations for API features include this constraint in agent prompt; agent designs REST-compliant APIs

### Scenario 2: Automatic Constraint Injection
**Given:** Project has 5 active constraints (2 intent, 2 security, 1 output)
**When:** Agent runs `/canary.specify` for a new authentication feature
**Then:** System identifies feature touches security domain; injects 2 security constraints + 2 intent constraints into specification prompt; agent creates spec that satisfies all 4 constraints

### Scenario 3: Validating Specification Against Constraints
**Given:** Spec created for caching feature; project has constraint "MUST: All caching mechanisms must have configurable TTL"
**When:** Maintainer runs `canary constrain validate .canary/specs/CBIN-050-caching/spec.md`
**Then:** System parses spec; checks for TTL configurability mention; reports PASS if found or FAIL with explanation if missing

### Scenario 4: Detecting Constraint Conflicts
**Given:** Existing constraint "MUST: Use SQL database for persistence"
**When:** Maintainer tries to add "MUST: Use NoSQL database for scalability"
**Then:** System detects conflict between two MUST constraints; warns that both cannot be satisfied; suggests making one SHOULD or resolving conflict

### Scenario 5: Constraint Evolution
**Given:** Project initially required synchronous APIs; now moving to async event-driven
**When:** Maintainer runs `canary constrain disable "Synchronous-API-Only"` and `canary constrain add --priority MUST --name "Event-Driven-Async" --rule "New features must use async event-driven patterns"`
**Then:** Old constraint stops enforcing; new constraint takes effect; history shows transition; agents adapt to new architecture

### Scenario 6: Agent Queries Constraints
**Given:** Agent is planning implementation for rate limiting feature
**When:** Agent executes internal query for constraints matching "performance" and "API" categories
**Then:** System returns 3 relevant constraints about API response times, rate limiting standards, and caching requirements; agent incorporates into plan

## Key Entities (if data-driven feature)

### Entity 1: Constraint
**Attributes:**
- constraint_id: Unique identifier (e.g., CONST-001)
- name: Short descriptive name
- category: One or more of [intent, context, output, security, performance]
- priority: MUST, SHOULD, or MAY
- rule_text: Natural language description of the constraint
- validation_criteria: How to verify satisfaction (may be manual or automated)
- rationale: Why this constraint exists
- examples_valid: Examples of outputs that satisfy constraint
- examples_invalid: Examples of outputs that violate constraint
- applies_to: Which operations this constraint affects (specify, plan, implement, all)
- created_at: Timestamp
- created_by: Author
- status: active, disabled, deprecated
- version: Version number for tracking changes

**Relationships:**
- May conflict with other constraints
- References gap analysis entries when violations occur
- Linked to CANARY requirements that depend on constraint

### Entity 2: ConstraintViolation
**Attributes:**
- violation_id: Unique identifier
- constraint_id: Which constraint was violated
- artifact_path: File or output that violated constraint
- violation_timestamp: When violation was detected
- evidence: Specific text or code that violated constraint
- severity: Based on constraint priority
- resolved: Whether violation was fixed
- resolution_note: How violation was addressed

**Relationships:**
- Links to Constraint
- May link to GapAnalysisEntry for learning

### Entity 3: ConstraintSet
**Attributes:**
- set_id: Identifier for a coherent group of constraints
- name: Descriptive name (e.g., "API Design Constraints")
- description: Purpose of this constraint set
- constraints: List of constraint IDs in this set

**Relationships:**
- Contains multiple Constraints
- Can be applied as a group to operations

## Assumptions

- Constraints are primarily enforced through agent prompt injection and validation, not runtime code enforcement
- Most constraint validation is manual or semi-automated (human review with checklist)
- Constraints are expressed in natural language understandable by AI agents
- Project maintainers have authority to define and modify constraints
- Constraints evolve slowly (not changed daily)

## Constraints

**Technical Constraints:**
- Must integrate with existing slash command template system
- Constraint files must be human-readable and version-control-friendly
- Validation engine must handle both structured and unstructured outputs

**Business Constraints:**
- Setup time for constraint system should be under 30 minutes for new projects
- Constraints should not significantly slow down agent operations
- Documentation must be accessible to non-technical stakeholders

**Regulatory Constraints:**
- May need to support compliance-related constraints (GDPR, security standards)

## Out of Scope

- Automatic generation of constraints from existing code
- Machine learning-based constraint inference
- Runtime enforcement of constraints in compiled code
- Formal verification or proof of constraint satisfaction
- Constraint synthesis from natural language project descriptions
- Integration with external compliance frameworks
- Real-time constraint violation detection during coding

## Dependencies

- Slash command system (CBIN-110, `/canary.specify`, `/canary.plan`, etc.)
- Template system in `.canary/templates/commands/`
- Storage system for persisting constraints
- (Optional) Gap analysis system (CBIN-140) for tracking violation patterns

## Risks & Mitigation

| Risk | Impact | Probability | Mitigation |
|------|--------|-------------|------------|
| Constraints become too numerous and overwhelming | High | Medium | Limit to 20 active constraints; require justification for new ones |
| Conflicting constraints create impossible situations | High | Medium | Implement conflict detection; warn on additions |
| Agents ignore constraints in prompts | High | Low | Make constraints prominent; validate outputs; track violations |
| Constraint language is ambiguous | Medium | High | Require examples for each constraint; validate clarity |
| Maintenance burden of keeping constraints current | Medium | Medium | Regular constraint review cycles; deprecation process |
| Over-constraining stifles creativity | Medium | Medium | Use SHOULD/MAY for guidance; reserve MUST for critical rules |

## Clarifications Needed

[NEEDS CLARIFICATION: Should constraint validation be automated where possible, or remain primarily manual?]
**Options:** A) Fully manual validation with checklists, B) Automated validation for structured constraints (e.g., API format), manual for complex ones, C) Fully automated with extensible validation rules
**Impact:** Determines implementation complexity and maintenance burden; affects validation speed and accuracy

[NEEDS CLARIFICATION: How should constraints be scoped - global to all operations, or contextual to specific operation types?]
**Options:** A) Global constraints apply everywhere, B) Constraints tagged with applicable operations (specify/plan/implement), C) Hierarchical with global and operation-specific overrides
**Impact:** Affects constraint organization and injection logic; determines flexibility vs. simplicity

[NEEDS CLARIFICATION: Should the constraint satisfaction framework be documented as a formal CSP model?]
**Options:** A) Document informally as agent guidelines, B) Use CSP terminology and concepts in documentation, C) Implement full CSP solver with formal constraint language
**Impact:** Affects documentation complexity, user understanding, and potential for formal analysis

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

<!-- CANARY: REQ=CBIN-142; FEATURE="ConstrainCmd"; ASPECT=CLI; STATUS=STUB; UPDATED=2025-10-17 -->
**Feature 1: Constrain CLI Command**
- [ ] Implement `canary constrain` subcommand with add/edit/list/remove/validate/docs actions
- **Location hint:** "cmd/constrain.go"
- **Dependencies:** Constraint storage

<!-- CANARY: REQ=CBIN-142; FEATURE="ConstraintStorage"; ASPECT=Storage; STATUS=STUB; UPDATED=2025-10-17 -->
**Feature 2: Constraint Persistence**
- [ ] Implement constraint storage in `.canary/constraints/` with YAML/JSON format
- **Location hint:** "internal/constraints/storage.go"
- **Dependencies:** File system access

<!-- CANARY: REQ=CBIN-142; FEATURE="ConstraintInjection"; ASPECT=Planner; STATUS=STUB; UPDATED=2025-10-17 -->
**Feature 3: Template Injection System**
- [ ] Modify slash command templates to inject relevant constraints into agent prompts
- **Location hint:** ".canary/templates/commands/*.md", "internal/constraints/injector.go"
- **Dependencies:** Template system, constraint retrieval

<!-- CANARY: REQ=CBIN-142; FEATURE="ConstraintValidation"; ASPECT=Engine; STATUS=STUB; UPDATED=2025-10-17 -->
**Feature 4: Validation Engine**
- [ ] Implement constraint validation logic for checking outputs against constraints
- **Location hint:** "internal/constraints/validator.go"
- **Dependencies:** Constraint loading, parsing of outputs

<!-- CANARY: REQ=CBIN-142; FEATURE="ConflictDetection"; ASPECT=Engine; STATUS=STUB; UPDATED=2025-10-17 -->
**Feature 5: Conflict Detection**
- [ ] Implement logic to detect conflicting constraints when adding/editing
- **Location hint:** "internal/constraints/conflict.go"
- **Dependencies:** Constraint storage, conflict resolution rules

<!-- CANARY: REQ=CBIN-142; FEATURE="ConstraintDocs"; ASPECT=Docs; STATUS=STUB; UPDATED=2025-10-17 -->
**Feature 6: Documentation Generation**
- [ ] Implement `canary constrain docs` to generate markdown documentation of active constraints
- **Location hint:** "internal/constraints/docgen.go"
- **Dependencies:** Constraint loading, markdown templating

### Slash Command

<!-- CANARY: REQ=CBIN-142; FEATURE="ConstrainSlashCmd"; ASPECT=CLI; STATUS=STUB; UPDATED=2025-10-17 -->
**Slash Command Creation:**
- [ ] Create `/canary.constrain` slash command with template
- **Location hint:** ".canary/templates/commands/constrain.md"
- **Dependencies:** Constraint CLI command

### Data Layer

<!-- CANARY: REQ=CBIN-142; FEATURE="ConstraintSchema"; ASPECT=Storage; STATUS=STUB; UPDATED=2025-10-17 -->
**Constraint Schema:**
- [ ] Define constraint data schema with all required fields
- **Location hint:** "internal/constraints/types.go" or schema documentation
- **Dependencies:** None

<!-- CANARY: REQ=CBIN-142; FEATURE="ConstraintQuery"; ASPECT=Storage; STATUS=STUB; UPDATED=2025-10-17 -->
**Constraint Query System:**
- [ ] Implement querying constraints by category, priority, applies_to, status
- **Location hint:** "internal/constraints/query.go"
- **Dependencies:** Constraint storage

### Integration

<!-- CANARY: REQ=CBIN-142; FEATURE="SpecifyIntegration"; ASPECT=Planner; STATUS=STUB; UPDATED=2025-10-17 -->
**Specify Command Integration:**
- [ ] Update `/canary.specify` template to inject constraints
- **Location hint:** ".canary/templates/commands/specify.md"
- **Dependencies:** Constraint injection system

<!-- CANARY: REQ=CBIN-142; FEATURE="PlanIntegration"; ASPECT=Planner; STATUS=STUB; UPDATED=2025-10-17 -->
**Plan Command Integration:**
- [ ] Update `/canary.plan` template to inject constraints
- **Location hint:** ".canary/templates/commands/plan.md"
- **Dependencies:** Constraint injection system

<!-- CANARY: REQ=CBIN-142; FEATURE="OtherCmdIntegration"; ASPECT=Planner; STATUS=STUB; UPDATED=2025-10-17 -->
**Other Command Integration:**
- [ ] Update remaining slash commands (verify, scan, etc.) to inject constraints
- **Location hint:** ".canary/templates/commands/*.md"
- **Dependencies:** Constraint injection system

### Testing Requirements

<!-- CANARY: REQ=CBIN-142; FEATURE="ConstraintUnitTests"; ASPECT=CLI; STATUS=STUB; TEST=TestConstraintManagement; UPDATED=2025-10-17 -->
**Unit Tests:**
- [ ] Test constraint CRUD operations, validation, conflict detection
- **Location hint:** "internal/constraints/*_test.go"

<!-- CANARY: REQ=CBIN-142; FEATURE="InjectionTests"; ASPECT=Planner; STATUS=STUB; TEST=TestConstraintInjection; UPDATED=2025-10-17 -->
**Injection Tests:**
- [ ] Test that constraints are correctly injected into all slash command prompts
- **Location hint:** "test/integration/constraint_injection_test.go"

<!-- CANARY: REQ=CBIN-142; FEATURE="ValidationTests"; ASPECT=Engine; STATUS=STUB; TEST=TestConstraintValidation; UPDATED=2025-10-17 -->
**Validation Tests:**
- [ ] Test validation engine with various constraint types and outputs
- **Location hint:** "internal/constraints/validator_test.go"

### Documentation

<!-- CANARY: REQ=CBIN-142; FEATURE="CSPConceptDocs"; ASPECT=Docs; STATUS=STUB; UPDATED=2025-10-17 -->
**CSP Framework Documentation:**
- [ ] Document CANARY as constraint satisfaction problem for agent management
- **Location hint:** "docs/concepts/constraint-satisfaction.md"

<!-- CANARY: REQ=CBIN-142; FEATURE="ConstraintUsageGuide"; ASPECT=Docs; STATUS=STUB; UPDATED=2025-10-17 -->
**Usage Guide:**
- [ ] Create guide for defining effective constraints with examples
- **Location hint:** "docs/guides/defining-constraints.md", ".canary/AGENT_CONTEXT.md"

<!-- CANARY: REQ=CBIN-142; FEATURE="ConstraintCLIDocs"; ASPECT=Docs; STATUS=STUB; UPDATED=2025-10-17 -->
**CLI Documentation:**
- [ ] Document all `canary constrain` subcommands and flags
- **Location hint:** "docs/cli/constrain.md"

---

**Agent Instructions:**

After implementing each feature:
1. Update the CANARY token in the spec from `STATUS=STUB` to `STATUS=IMPL`
2. Add the same token to your source code at the implementation location
3. Add `TEST=TestName` when tests are written
4. Run `canary implement CBIN-142` to see implementation progress

---

## CANARY Tokens Reference

**Main requirement token** (add to primary implementation file):
```
// CANARY: REQ=CBIN-142; FEATURE="ConstraintSatisfaction"; ASPECT=CLI; STATUS=IMPL; UPDATED=2025-10-17
```

**Sub-feature tokens** (use the specific feature names from Implementation Checklist):
```
// CANARY: REQ=CBIN-142; FEATURE="ConstrainCmd"; ASPECT=CLI; STATUS=IMPL; TEST=TestConstrainCmd; UPDATED=2025-10-17
```

**Use `canary implement CBIN-142` to find:**
- Which features are implemented vs. still TODO
- Exact file locations and line numbers
- Context around each implementation point
