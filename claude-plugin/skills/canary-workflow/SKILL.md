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
4. **Scan** — Verify token status
5. **Verify** — Confirm claims match reality

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

### 6. Scan for Status

```
/canary.scan
```

Runs `canary scan` and reports:
- Token counts by status (STUB/IMPL/TESTED/BENCHED)
- Coverage metrics
- Stale tokens needing updates
- Action items

### 7. Verify Claims

```
/canary.verify
```

Checks that `GAP_ANALYSIS.md` claims match actual token status. Catches overclaims (claiming a feature is tested when it's only implemented).

## Quick Priority Workflow

Don't know what to work on next?

```
/canary.next
```

Automatically selects the highest-priority unfinished requirement and generates full implementation guidance.

## Maintenance

- `/canary.update-stale` — Refresh UPDATED dates on stale tokens (>30 days)
- `/canary.list --status IMPL` — Find implementations needing tests
- `/canary.doc status --all` — Check documentation freshness

## Key Principle: Minimal Context

- Only load the command file you need
- Use stdout one-liners from `canary scan` for metrics
- Don't read `status.json` or `GAP_ANALYSIS.md` unless editing them
- Reference constitution only when creating/editing principles
