<!-- CANARY: REQ=CBIN-148; FEATURE="InstructionTemplates"; ASPECT=Docs; STATUS=TESTED; TEST=TestCopilotInstructionTemplateValidity; UPDATED=2025-10-19 -->

# CANARY Development Instructions

This project uses the CANARY requirement tracking system.

## CANARY Token Format

All features must include a CANARY token:

```
// CANARY: REQ={{.ProjectKey}}-###; FEATURE="Name"; ASPECT=API; STATUS=IMPL; UPDATED=YYYY-MM-DD
```

## Required Fields

- **REQ**: Requirement ID (format: {{.ProjectKey}}-###)
- **FEATURE**: Short feature name (PascalCase)
- **ASPECT**: Category (API, CLI, Engine, Storage, Security, Docs, etc.)
- **STATUS**: Implementation state (STUB, IMPL, TESTED, BENCHED)
- **UPDATED**: Last update date (YYYY-MM-DD format)

## Status Progression

- STUB → IMPL: Implementation exists
- IMPL → TESTED: TEST= field added with passing test
- TESTED → BENCHED: BENCH= field added with benchmark

## Test-First Development (NON-NEGOTIABLE)

Per Article IV of the CANARY Constitution:

1. Write test function FIRST (red phase)
2. Add TEST=FunctionName to CANARY token
3. Implement feature to make test pass (green phase)
4. Update STATUS from IMPL to TESTED

## Available Slash Commands

- `/canary.specify` - Create new requirement specification
- `/canary.plan` - Generate implementation plan
- `/canary.scan` - Scan codebase for tokens
- `/canary.verify` - Verify GAP_ANALYSIS.md claims
- `/canary.implement` - Generate implementation guidance

## Constitutional Principles

Follow these principles from `.canary/memory/constitution.md`:

1. **Requirement-First** (Article I): Every feature starts with a token
2. **Test-First** (Article IV): Tests before implementation, always
3. **Simplicity** (Article V): Prefer standard library, avoid complexity
4. **Evidence-Based** (Article I.2): Status promoted only with TEST=/BENCH= evidence

For complete details, see `.canary/memory/constitution.md`.
