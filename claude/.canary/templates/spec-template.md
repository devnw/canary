# CANARY: REQ=CBIN-115; FEATURE="SpecTemplate"; ASPECT=Docs; STATUS=IMPL; OWNER=canary; UPDATED=2025-10-16
# Feature Specification: [FEATURE NAME]

**Requirement ID:** CBIN-XXX
**Status:** STUB
**Created:** YYYY-MM-DD
**Last Updated:** YYYY-MM-DD

## Overview

**Purpose:** [What problem does this feature solve?]

**Scope:** [What is included and excluded from this feature]

## User Stories

### Primary User Stories

**US-1: [Story Title]**
As a [user type],
I want to [action],
So that [benefit/value].

**Acceptance Criteria:**
- [ ] [Specific, testable criterion]
- [ ] [Specific, testable criterion]
- [ ] [Specific, testable criterion]

**US-2: [Story Title]**
As a [user type],
I want to [action],
So that [benefit/value].

**Acceptance Criteria:**
- [ ] [Specific, testable criterion]
- [ ] [Specific, testable criterion]

### Secondary User Stories (if applicable)

[Additional user stories that support the primary stories]

## Functional Requirements

### FR-1: [Requirement Name]
**Priority:** High/Medium/Low
**Description:** [What the system must do]
**Acceptance:** [How to verify this requirement]

### FR-2: [Requirement Name]
**Priority:** High/Medium/Low
**Description:** [What the system must do]
**Acceptance:** [How to verify this requirement]

## Success Criteria

**Quantitative Metrics:**
- [ ] [Measurable outcome, e.g., "Users complete task in < 3 minutes"]
- [ ] [Performance metric, e.g., "Handles 10,000 concurrent users"]
- [ ] [Accuracy metric, e.g., "95% success rate"]

**Qualitative Measures:**
- [ ] [User satisfaction indicator]
- [ ] [Task completion quality]
- [ ] [System reliability measure]

**Important:** All success criteria must be:
- **Measurable**: Include specific numbers/percentages
- **Technology-Agnostic**: No mention of implementation details
- **User-Focused**: Describe outcomes from user perspective
- **Verifiable**: Can be tested without knowing implementation

## User Scenarios & Testing

### Scenario 1: [Happy Path]
**Given:** [Initial condition]
**When:** [Action taken]
**Then:** [Expected outcome]

### Scenario 2: [Edge Case]
**Given:** [Initial condition]
**When:** [Action taken]
**Then:** [Expected outcome]

### Scenario 3: [Error Case]
**Given:** [Initial condition]
**When:** [Action taken]
**Then:** [Expected outcome]

## Key Entities (if data-driven feature)

### Entity 1: [Entity Name]
**Attributes:**
- [attribute]: [description]
- [attribute]: [description]

**Relationships:**
- [relationship to other entities]

### Entity 2: [Entity Name]
[Same structure as above]

## Assumptions

- [Assumption about environment, users, or constraints]
- [Assumption about external systems or data]
- [Assumption about user behavior or preferences]

## Constraints

**Technical Constraints:**
- [Any technical limitations to be aware of]

**Business Constraints:**
- [Budget, timeline, or resource constraints]

**Regulatory Constraints:**
- [Compliance requirements, if applicable]

## Out of Scope

- [Explicitly state what is NOT included]
- [Features that might be confused as part of this]
- [Future enhancements to be added later]

## Dependencies

- [Other requirements this depends on]
- [External systems or services required]
- [Team dependencies or prerequisites]

## Risks & Mitigation

| Risk | Impact | Probability | Mitigation |
|------|--------|-------------|------------|
| [Risk description] | High/Med/Low | High/Med/Low | [How to mitigate] |

## Clarifications Needed

[NEEDS CLARIFICATION: Specific question about scope/security/UX]
**Options:** A) [option], B) [option], C) [option]
**Impact:** [How this decision affects the feature]

[Maximum 3 clarifications - use only for critical decisions]

## Review & Acceptance Checklist

**Content Quality:**
- [ ] No implementation details (languages, frameworks, APIs)
- [ ] Focused on user value and business needs
- [ ] Written for non-technical stakeholders
- [ ] All mandatory sections completed

**Requirement Completeness:**
- [ ] No [NEEDS CLARIFICATION] markers remain
- [ ] Requirements are testable and unambiguous
- [ ] Success criteria are measurable and technology-agnostic
- [ ] All acceptance scenarios defined
- [ ] Edge cases identified
- [ ] Scope clearly bounded
- [ ] Dependencies and assumptions identified

**Readiness:**
- [ ] All functional requirements have clear acceptance criteria
- [ ] User scenarios cover primary flows
- [ ] Ready for technical planning (`/canary.plan`)

---

## Implementation Checklist

Break down this requirement into specific implementation points. Each point gets its own CANARY token to help agents locate where to implement changes.

### Core Features

<!-- CANARY: REQ=CBIN-XXX; FEATURE="CoreFeature1"; ASPECT=API; STATUS=STUB; UPDATED=YYYY-MM-DD -->
**Feature 1: [Component Name]**
- [ ] Implement [specific functionality]
- **Location hint:** [e.g., "auth.go", "handlers/", "services/auth/"]
- **Dependencies:** [other features this depends on]

<!-- CANARY: REQ=CBIN-XXX; FEATURE="CoreFeature2"; ASPECT=API; STATUS=STUB; UPDATED=YYYY-MM-DD -->
**Feature 2: [Component Name]**
- [ ] Implement [specific functionality]
- **Location hint:** [e.g., "validation.go", "middleware/"]
- **Dependencies:** [other features this depends on]

### Data Layer (if applicable)

<!-- CANARY: REQ=CBIN-XXX; FEATURE="DataModel"; ASPECT=Storage; STATUS=STUB; UPDATED=YYYY-MM-DD -->
**Data Model:**
- [ ] Define schema/structure
- **Location hint:** [e.g., "models/", "schema/"]

<!-- CANARY: REQ=CBIN-XXX; FEATURE="DataAccess"; ASPECT=Storage; STATUS=STUB; UPDATED=YYYY-MM-DD -->
**Data Access:**
- [ ] Implement CRUD operations
- **Location hint:** [e.g., "repository/", "dao/"]

### Testing Requirements

<!-- CANARY: REQ=CBIN-XXX; FEATURE="UnitTests"; ASPECT=API; STATUS=STUB; TEST=TestCBINXXX; UPDATED=YYYY-MM-DD -->
**Unit Tests:**
- [ ] Test core functionality
- **Location hint:** [e.g., "*_test.go files"]

<!-- CANARY: REQ=CBIN-XXX; FEATURE="IntegrationTests"; ASPECT=API; STATUS=STUB; TEST=TestCBINXXXIntegration; UPDATED=YYYY-MM-DD -->
**Integration Tests:**
- [ ] Test end-to-end flows
- **Location hint:** [e.g., "integration/", "*_integration_test.go"]

### Documentation

<!-- CANARY: REQ=CBIN-XXX; FEATURE="APIDocs"; ASPECT=Docs; STATUS=STUB; UPDATED=YYYY-MM-DD -->
**API Documentation:**
- [ ] Document public interfaces
- **Location hint:** [e.g., "docs/api/", "README.md"]

---

**Agent Instructions:**

After implementing each feature:
1. Update the CANARY token in the spec from `STATUS=STUB` to `STATUS=IMPL`
2. Add the same token to your source code at the implementation location
3. Add `TEST=TestName` when tests are written
4. Run `canary implement CBIN-XXX` to see implementation progress

---

## CANARY Tokens Reference

**Main requirement token** (add to primary implementation file):
```
// CANARY: REQ=CBIN-XXX; FEATURE="FeatureName"; ASPECT=API; STATUS=IMPL; UPDATED=YYYY-MM-DD
```

**Sub-feature tokens** (use the specific feature names from Implementation Checklist):
```
// CANARY: REQ=CBIN-XXX; FEATURE="CoreFeature1"; ASPECT=API; STATUS=IMPL; TEST=TestCoreFeature1; UPDATED=YYYY-MM-DD
```

**Use `canary implement CBIN-XXX` to find:**
- Which features are implemented vs. still TODO
- Exact file locations and line numbers
- Context around each implementation point
