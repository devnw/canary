---
name: canary-workflow
description: >
  Use when the user starts a new feature, asks "how do I use canary", asks about the
  development workflow, or needs guidance on the CANARY requirement lifecycle.
  Covers the full specify -> plan -> implement -> scan -> verify pipeline.
---

# CANARY Development Workflow

## Overview

CANARY tracks requirements by embedding structured tokens directly in source code. Every feature follows this lifecycle:

1. **Specify** — Define what to build
2. **Plan** — Design how to build it
3. **Implement** — Build it (test-first)
4. **View** — Check the full picture in one call (status, files, tests, deps)
5. **Scan** — Verify token status
6. **Verify** — Confirm claims match reality
7. **Drift** — Catch tokens that fell out of sync with the code (ongoing)

## Step-by-Step Workflow

### 1. Initialize (First Time Only)

```bash
canary init --agents claude
```

This creates the `.canary/` directory with project configuration, templates, and scripts.

### 2. Establish Principles (First Time Only)

```
/canary.constitution
```

Creates `.canary/memory/constitution.md` with governing principles (test-first, requirement-first, etc.).

### 3. Specify a Requirement

```
/canary.specify Add user authentication with OAuth2 support
```

This:
- Generates a requirement ID (e.g., `PROJ-105`)
- Creates `.canary/specs/PROJ-105-user-authentication/spec.md`
- Produces a CANARY token template ready to paste into code

### 4. Plan Implementation

```
/canary.plan PROJ-105
```

This:
- Loads the specification
- Generates a technical implementation plan at `.canary/specs/PROJ-105-user-authentication/plan.md`
- Includes test-first phases, architecture, and CANARY token placement

### 5. Implement (Test-First)

```
/canary.implement PROJ-105
```

This generates comprehensive implementation guidance. Follow TDD:

1. Write failing tests first
2. Implement minimal code to pass tests
3. Place CANARY tokens above implementations:
   ```go
   // CANARY: REQ=PROJ-105; FEATURE="UserAuth"; ASPECT=API; STATUS=IMPL; UPDATED=2025-10-16
   ```
4. Add `TEST=TestFunctionName` to auto-promote to TESTED
5. Add `BENCH=BenchFunctionName` to auto-promote to BENCHED

### 6. View the Full Picture

```
/canary.view PROJ-105
```

One call for status, files, tests, dependency IDs, spec/plan paths, and ticket URL — use this instead of chaining `show`/`status`/`files` when you need the whole context for a requirement.

### 7. Scan for Status

```
/canary.scan
```

Runs `canary scan` and reports:
- Token counts by status (STUB/IMPL/TESTED/BENCHED)
- Coverage metrics
- Stale tokens needing updates
- Action items

### 8. Verify Claims

```
/canary.verify
```

Checks that `GAP_ANALYSIS.md` claims match actual token status. Catches overclaims (claiming a feature is tested when it's only implemented).

### 9. Check Drift (Ongoing)

```
/canary.drift --strict
```

Rescans the repo and reports tokens whose file was committed after their `UPDATED` date (code-drift), tokens stale past the staleness window, and (with an indexed database) doc-drift. `--strict` exits 2 on any finding — good for CI.

Dependencies on a ticket-source (JIRA, etc.) or `peers:`-owned requirement resolve via a peer's `status.json` or the cached remote status (`canary ticket status --refresh`). For `canary next`, only *satisfied* clears a dependency: unsatisfied blocks, and unresolvable (no cache entry, no peer answer) blocks too — run `canary ticket status --refresh`, or pass `canary next --allow-unknown-external` to accept the risk. `canary deps validate` reports unsatisfied/unknown externals on its `external:` summary line without failing, until `--strict-external`.

## Quick Priority Workflow

Don't know what to work on next?

```
/canary.next
```

Automatically selects the highest-priority unfinished requirement and generates full implementation guidance.

## Maintenance

- `/canary.update-stale` — Report evidence currency for stale tokens (>30 days); rewrites nothing
- `/canary.list --status IMPL` — Find implementations needing tests
- `/canary.doc` with arguments `status --all` — Check documentation freshness (invokes `canary doc status --all`)
- `/canary.drift --strict` — Check for code/token drift (exit 2 on any finding)

## Key Principle: Minimal Context

- Only load the command file you need
- Use stdout one-liners from `canary scan` for metrics
- Don't read `status.json` or `GAP_ANALYSIS.md` unless editing them
- Reference constitution only when creating/editing principles
