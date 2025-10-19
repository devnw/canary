<!-- CANARY: REQ=CBIN-148; FEATURE="InstructionTemplates"; ASPECT=Docs; STATUS=BENCHED; TEST=TestCopilotInstructionTemplateValidity; BENCH=BenchmarkCreateCopilotInstructions; UPDATED=2025-10-19 -->

# CANARY Specification Writing Guidelines

You are working in the `.canary/specs/` directory where requirement specifications are stored.

## Specification Focus: WHAT and WHY, Not HOW

**Key Principle:** Specifications describe WHAT users need and WHY, NOT HOW to implement it.

### ✅ Good Specification Language

- "Users can authenticate within 3 seconds"
- "System handles 10,000 concurrent requests"
- "Data validation provides clear error messages"
- "Search results appear in under 500ms"

### ❌ Avoid Implementation Details

- ~~"Use JWT tokens with RS256 encryption"~~
- ~~"Implement using Redis cache"~~
- ~~"Store in PostgreSQL database"~~
- ~~"Use React hooks for state management"~~

## Technology-Agnostic Writing

Describe requirements without mentioning:
- Programming languages
- Frameworks or libraries
- Specific databases
- API protocols
- Infrastructure details

Implementation details belong in the **plan.md** file, not the specification.

## Required Specification Sections

1. **Overview** - Purpose and scope
2. **User Stories** - Who wants what and why
3. **Functional Requirements** - Testable, unambiguous requirements
4. **Success Criteria** - Measurable, technology-agnostic outcomes
5. **User Scenarios** - Given/When/Then acceptance tests

## Measurable Success Criteria

Every success criterion must be:
- **Measurable**: Include specific numbers/percentages
- **User-Focused**: Describe outcomes from user perspective
- **Verifiable**: Can be tested without knowing implementation

## Creating Specifications

**Read `.canary/commands/specify.md` for the complete workflow**, including:
- How to generate requirement IDs
- Required sections and quality standards
- Validation criteria
- Example specifications

## Planning Implementation

**Read `.canary/commands/plan.md` for creating technical plans** (where HOW goes):
- Architecture decisions
- Technology choices
- Implementation phases
- Testing strategy

## Related Files

- `.canary/commands/specify.md` - Complete specification workflow
- `.canary/commands/plan.md` - Implementation planning workflow
- `.canary/templates/spec-template.md` - Specification template
- `.canary/memory/constitution.md` - Governing principles
