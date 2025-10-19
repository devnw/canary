<!-- CANARY: REQ=CBIN-148; FEATURE="InstructionTemplates"; ASPECT=Docs; STATUS=TESTED; TEST=TestCopilotInstructionTemplateValidity; UPDATED=2025-10-19 -->

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

## Required Specification Sections

1. **Overview** - Purpose and scope
2. **User Stories** - Who wants what and why
3. **Functional Requirements** - Testable, unambiguous requirements
4. **Success Criteria** - Measurable, technology-agnostic outcomes
5. **User Scenarios** - Given/When/Then acceptance tests

## Technology-Agnostic Writing

Describe requirements without mentioning:
- Programming languages
- Frameworks or libraries
- Specific databases
- API protocols
- Infrastructure details

Implementation details belong in the **plan.md** file, not the specification.

## Measurable Success Criteria

Every success criterion must be:
- **Measurable**: Include specific numbers/percentages
- **User-Focused**: Describe outcomes from user perspective
- **Verifiable**: Can be tested without knowing implementation

## Related Commands

- `/canary.specify` - Create new specification
- `/canary.plan` - Create implementation plan (where HOW goes)
- `/canary.scan` - Check spec completeness
