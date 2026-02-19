# CANARY Development - AI Agent Guide

**Minimal context:** This project uses CANARY requirement tracking. Load only what you need.

## Commands (load only the one you use)

- **/canary.constitution** → `.canary/commands/constitution.md`
- **/canary.specify** → `.canary/commands/specify.md`
- **/canary.plan** → `.canary/commands/plan.md`
- **/canary.scan** → `.canary/commands/scan.md`
- **/canary.verify** → `.canary/commands/verify.md`
- **/canary.update-stale** → `.canary/commands/update-stale.md`

Read **only** the command file for the slash command you are running. Do not load AGENT_CONTEXT, full constitution, or GAP_ANALYSIS unless that command’s instructions tell you to.

## Scan & verify (low context)

- **Scan:** Run `canary scan --root . --out status.json`. Use the **one line** on stdout (`CANARY_SCAN tokens=N requirements=M STUB=...`) for metrics. Do not read `status.json` unless you need per-requirement detail.
- **Verify:** Run `canary scan --root . --verify GAP_ANALYSIS.md --strict`. Use stdout (`CANARY_VERIFY_OK` or `CANARY_VERIFY_FAIL count=N`) and stderr for failures. Open `GAP_ANALYSIS.md` only when fixing or updating claims.

## Token format

`// CANARY: REQ=CBIN-###; FEATURE="Name"; ASPECT=API; STATUS=IMPL; UPDATED=YYYY-MM-DD`  
Status: STUB → IMPL → TESTED → BENCHED. Aspects: API, CLI, Engine, Storage, Security, Docs, Wire, Planner, Decode, Encode, RoundTrip, Bench, FrontEnd, Dist.

## Principles (inline; load constitution only when editing principles)

1. Requirement-First: every feature has a CANARY token
2. Test-First: tests before implementation (Article IV)
3. Evidence-Based: status from TEST=/BENCH=
4. Simplicity: prefer standard library
5. Documentation Currency: keep UPDATED current
