---
description: Create or update project governing principles and CANARY development guidelines
---

<!-- CANARY: REQ=CBIN-109; FEATURE="ConstitutionCmd"; ASPECT=CLI; STATUS=IMPL; OWNER=canary; UPDATED=2025-10-16 -->

## User Input

```text
$ARGUMENTS
```

## Outline

Create or update the project's constitutional principles that will govern all CANARY requirement tracking and development.

1. **Read existing constitution** (if `.canary/memory/constitution.md` exists) to understand current principles

2. **Analyze user request** from the arguments:
   - If arguments are empty: Create initial constitution using the template
   - If arguments provided: Update specific principles based on user guidance

3. **Create/update constitution** at `.canary/memory/constitution.md`:
   - Include core CANARY principles (requirement-first, test-driven, evidence-based)
   - Add project-specific principles from user input
   - Ensure consistency with Article I-IX framework
   - Document rationale for any custom additions

4. **Validate constitution** against quality criteria:
   - All principles are actionable
   - No contradictions between articles
   - Clear enforcement mechanisms
   - Measurable compliance criteria

5. **Report completion**:
   - Path to constitution file
   - Summary of principles established
   - Next steps: Use `/canary.specify` to create your first requirement

## Guidelines

- **Core Principles**: Always include Article I (Requirement-First), Article IV (Test-First), Article VII (Documentation Currency)
- **Project Additions**: Add project-specific articles (X, XI, etc.) for domain-specific needs
- **Enforcement**: Constitution principles should be referenced in `/canary.plan` validation gates
- **Evolution**: Document amendments with version numbers and effective dates

## Example Custom Additions

If user specifies security requirements:
```markdown
## Article X: Security Standards

**Section 10.1: Authentication**
All API endpoints MUST implement OAuth2 authentication.

**Section 10.2: Data Encryption**
All sensitive data MUST be encrypted at rest and in transit.
```

If user specifies performance requirements:
```markdown
## Article X: Performance Baselines

**Section 10.1: Response Time**
API endpoints MUST respond within 200ms for 95th percentile requests.

**Section 10.2: Throughput**
System MUST handle 10,000 concurrent users.
```
