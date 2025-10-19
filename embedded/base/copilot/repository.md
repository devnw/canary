<!-- CANARY: REQ=CBIN-148; FEATURE="InstructionTemplates"; ASPECT=Docs; STATUS=BENCHED; TEST=TestCopilotInstructionTemplateValidity; BENCH=BenchmarkCreateCopilotInstructions; UPDATED=2025-10-19 -->

# CANARY Development Instructions

This project uses the CANARY requirement tracking system. All command workflows are documented in `.canary/commands/`.

## CANARY Token Format

All features must include a CANARY token:

```
// CANARY: REQ={{.ProjectKey}}-###; FEATURE="Name"; ASPECT=API; STATUS=IMPL; UPDATED=YYYY-MM-DD
```

**Status Progression**: STUB → IMPL → TESTED → BENCHED

See `.canary/commands/specify.md` for complete token format details.

## Test-First Development (NON-NEGOTIABLE)

Per Article IV of `.canary/memory/constitution.md`:

1. Write test function FIRST (red phase)
2. Add TEST=FunctionName to CANARY token
3. Implement feature to make test pass (green phase)
4. Update STATUS from IMPL to TESTED

## Available Slash Commands

**All command details are in `.canary/commands/` - read these files for complete workflows:**

**Requirement Management:**
- `/canary.specify` → `.canary/commands/specify.md`
- `/canary.plan` → `.canary/commands/plan.md`
- `/canary.scan` → `.canary/commands/scan.md`
- `/canary.verify` → `.canary/commands/verify.md`
- `/canary.update-stale` → `.canary/commands/update-stale.md`

**Query & Implementation:**
- `/canary.show` → `.canary/commands/show.md`
- `/canary.status` → `.canary/commands/status.md`
- `/canary.files` → `.canary/commands/files.md`
- `/canary.grep` → `.canary/commands/grep.md`
- `/canary.list` → `.canary/commands/list.md`
- `/canary.next` → `.canary/commands/next.md`
- `/canary.implement` → `.canary/commands/implement.md`

**Each command file contains step-by-step workflows, examples, and validation criteria.**

## Constitutional Principles

See `.canary/memory/constitution.md` for complete governing principles:

- **Article I**: Requirement-First Development
- **Article IV**: Test-First Imperative (non-negotiable)
- **Article V**: Simplicity and Anti-Abstraction
- **Article VI**: Integration-First Testing
- **Article VII**: Documentation Currency

## Quick Workflow

**Starting a new feature:**
1. Read `.canary/commands/specify.md` for the workflow
2. Run `/canary.specify <description>`
3. Read `.canary/commands/plan.md` for planning
4. Run `/canary.plan {{.ProjectKey}}-XXX`
5. Follow test-first development (Article IV)
6. Place CANARY tokens as you implement
7. Run `/canary.scan` to verify

**For any question about commands, read the corresponding file in `.canary/commands/` first.**
