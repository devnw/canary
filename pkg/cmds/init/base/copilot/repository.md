<!-- CANARY: REQ=CBIN-148; FEATURE="InstructionTemplates"; ASPECT=Docs; STATUS=BENCHED; TEST=TestCopilotInstructionTemplateValidity; BENCH=BenchmarkCreateCopilotInstructions; UPDATED=2025-10-19 -->

# CANARY Development Instructions

This project uses the CANARY requirement tracking system. All command workflows are documented in `.canary/commands/`.

## CANARY Token Format

All features must include a CANARY token:

```
// CANARY: REQ=<PROJECT_KEY>-###; FEATURE="Name"; ASPECT=API; STATUS=IMPL; UPDATED=YYYY-MM-DD
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

**One-Call Context:**
- `/canary.view` → `.canary/commands/view.md` — full picture of a requirement in one call: `canary view <REQ-ID>`
- `/canary.deps` → `.canary/commands/deps.md` — check/graph/reverse/validate dependencies

**Requirement Management:**
- `/canary.specify` → `.canary/commands/specify.md`
- `/canary.plan` → `.canary/commands/plan.md`
- `/canary.scan` → `.canary/commands/scan.md`
- `/canary.verify` → `.canary/commands/verify.md`
- `/canary.update-stale` → `.canary/commands/update-stale.md`
- `/canary.drift` → `.canary/commands/drift.md` — code-vs-token drift detection (`--strict` for CI)
- `/canary.upgrade` → `.canary/commands/upgrade.md` — rewrite legacy token shapes (dry run by default)
- `/canary.onboard` → `.canary/commands/onboard.md` — fresh-codebase adoption analysis
- `/canary.ticket` → `.canary/commands/ticket.md` — reconcile tokens against configured ticket sources

**Query & Implementation:**
- `/canary.show` → `.canary/commands/show.md`
- `/canary.status` → `.canary/commands/status.md`
- `/canary.files` → `.canary/commands/files.md`
- `/canary.grep` → `.canary/commands/grep.md`
- `/canary.search` → `.canary/commands/search.md`
- `/canary.list` → `.canary/commands/list.md`
- `/canary.next` → `.canary/commands/next.md`
- `/canary.implement` → `.canary/commands/implement.md`
- `/canary.gap` → `.canary/commands/gap.md`
- `/canary.index` → `.canary/commands/index.md`

**Each command file contains step-by-step workflows, examples, and validation criteria.**

**Ticket sources** (JIRA, etc.) are configured in `.canary/project.yaml`'s `sources:` list; `canary ticket sync` stays plan-only without credentials, and `canary ticket status [--refresh]` reports/refreshes the cache on its own. An external (ticket-sourced or `peers:`-owned) dependency blocks `canary next` unless it resolves *satisfied* — an unresolvable one (no cache entry, no peer answer) blocks too, since handing an agent work whose prerequisite might not exist is a wrong answer, not a graceful degradation. `canary next --allow-unknown-external` accepts that risk explicitly. `canary deps validate` is the permissive one: unsatisfied and unknown externals are informational there until you pass `--strict-external`.

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
4. Run `/canary.plan <PROJECT_KEY>-XXX`
5. Follow test-first development (Article IV)
6. Place CANARY tokens as you implement
7. Run `/canary.scan` to verify

**For any question about commands, read the corresponding file in `.canary/commands/` first.**
