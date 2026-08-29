---
description: Compare CANARY tokens against reality -- commits after UPDATED, staleness, and doc rollup drift
---

## User Input

```text
$ARGUMENTS
```

## Outline

`canary drift` rescans the repo fresh and reports three kinds of findings:

- **code-drift** — a token's file was committed (per git history) after its `UPDATED` date; the code moved on but the token wasn't refreshed.
- **stale** — `TESTED`/`BENCHED` tokens older than the staleness window (same rule as `canary scan --strict`).
- **doc-drift** — when `.canary/canary.db` exists, tokens whose tracked documentation is `DOC_STALE` or `DOC_MISSING`.

```bash
canary drift                          # table, default limit 20
canary drift --json                   # {findings:[...], summary:{...}}
canary drift --stale-days 45          # override the staleness window
canary drift --strict                 # exit 2 if any finding exists
canary drift --limit 50               # raise the table's requirement-group cap
```

## Flags

- `--root <dir>` (default `.`), `--json`, `--stale-days <n>` (0 = use `.canary/project.yaml` `verification.staleness_days`, else 30), `--strict`, `--limit <n>` (default 20)

## Guidelines

- Use `--strict` in CI to fail the build on any drift finding.
- Run `canary index` first if you want `doc-drift` findings included — without a database, only `code-drift` and `stale` are computed.
- Follow up a `code-drift` finding by updating the token's `UPDATED` field (or, if the shape is legacy, `canary upgrade --write`).
