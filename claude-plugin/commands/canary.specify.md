---
description: Create a new CANARY requirement specification from a natural language feature description
scripts:
  sh: .canary/scripts/create-new-requirement.sh "$ARGUMENTS"
---


## User Input

```text
$ARGUMENTS
```

You **MUST** consider the user input before proceeding (if not empty).

## Outline

The text the user typed after `/canary.specify` is the feature description.

Given that feature description, do this:

1. **Determine project key**:
   - Read `.canary/project.yaml` for the project key (e.g., `PROJECT_KEY`)
   - Use the key when generating requirement IDs: `<PROJECT_KEY>-<NNN>`

2. **Generate requirement ID**:
   - Scan existing `.canary/specs/` directory for highest `<PROJECT_KEY>-###` number
   - Assign next sequential ID (e.g., if `<PROJECT_KEY>-105` exists, use `<PROJECT_KEY>-106`)
   - Format: `<PROJECT_KEY>-XXX` (zero-padded 3 digits)

3. **Generate concise feature name** (2-4 words):
   - Extract meaningful keywords from description
   - Use action-noun format (e.g., "user-authentication", "data-validation")
   - Keep technical terms (OAuth2, JWT, API, etc.)

4. **Run the script** `.canary/scripts/create-new-requirement.sh --req-id <PROJECT_KEY>-XXX --feature "name"`:
   - Creates `.canary/specs/<PROJECT_KEY>-XXX-feature-name/` directory
   - Initializes `spec.md` from template
   - Returns SPEC_FILE path

5. **Create the specification**:
   - Load `templates/spec-template.md` to understand required sections
   - Focus on WHAT users need and WHY (not HOW to implement)
   - Fill required sections:
     - Feature Overview
     - User Stories
     - Functional Requirements
     - Success Criteria (measurable, technology-agnostic)
     - Assumptions and Constraints
   - Limit [NEEDS CLARIFICATION] markers to maximum 3 critical items

6. **Generate CANARY token**:
   ```
   // CANARY: REQ=<PROJECT_KEY>-XXX; FEATURE="FeatureName"; ASPECT=API; STATUS=STUB; UPDATED=YYYY-MM-DD
   ```
   - Determine appropriate ASPECT based on feature description
   - Set STATUS=STUB (will be promoted when implemented)
   - Use current date for UPDATED field

7. **Document where to place token**:
   - Suggest file/location in codebase where token should go
   - Provide example of token placement

8. **Create requirement tracking entry**:
   - Update `.canary/requirements.md` (create if doesn't exist)
   - Add entry: `- [ ] <PROJECT_KEY>-XXX - FeatureName (STATUS=STUB)`

9. **Report completion**:
   - Requirement ID and feature name
   - Spec file path
   - CANARY token (ready to paste into code)
   - Suggested code location
   - Next steps: Use `/canary.plan` to create implementation plan

## Quality Validation

After creating the spec, validate:

- [ ] No implementation details (languages, frameworks, APIs)
- [ ] Focused on user value and business needs
- [ ] All requirements are testable and unambiguous
- [ ] Success criteria are measurable and technology-agnostic
- [ ] Maximum 3 [NEEDS CLARIFICATION] markers
- [ ] All acceptance scenarios defined
- [ ] Feature scope clearly bounded

## Example Flow

User input: "Add user authentication with email/password and OAuth2 support"

1. Generate ID: `<PROJECT_KEY>-107`
2. Feature name: "user-authentication"
3. Create: `.canary/specs/<PROJECT_KEY>-107-user-authentication/spec.md`
4. CANARY token:
   ```
   // CANARY: REQ=<PROJECT_KEY>-107; FEATURE="UserAuthentication"; ASPECT=API; STATUS=STUB; UPDATED=2025-10-16
   ```
5. Suggest placement:
   ```go
   // File: src/auth/auth.go
   // CANARY: REQ=<PROJECT_KEY>-107; FEATURE="UserAuthentication"; ASPECT=API; STATUS=STUB; UPDATED=2025-10-16
   package auth

   func Authenticate(credentials Credentials) (*Session, error) {
       // Implementation will go here
   }
   ```

## Guidelines

- **Specification Focus**: WHAT and WHY, not HOW
- **Measurable Outcomes**: "Users can login in < 3 seconds" not "JWT token generation"
- **Technology Agnostic**: Don't specify frameworks, databases, or languages
- **Token Placement**: Suggest logical location in codebase based on ASPECT
- **Status Progression**: Starts as STUB, becomes IMPL when implemented, TESTED when tests added
