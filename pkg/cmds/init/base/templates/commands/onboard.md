---
description: Fresh-codebase adoption analysis -- languages, entry points, existing tokens, MIGRATE notes, next steps
---

## User Input

```text
$ARGUMENTS
```

## Outline

Run this first when adopting CANARY on a codebase that doesn't use it yet (or only partially does).

1. **Run**:
   ```bash
   canary onboard --root .
   canary onboard --json              # machine-readable
   canary onboard --limit 40          # raise per-section cap (default 20)
   ```
2. Read the report: detected languages (by extension), top directories, likely entry points, existing CANARY token count (heuristic — run `canary scan` for the exact figure), and any `CANARY:MIGRATE <free text>` notes already left in the codebase.
3. Use the "next steps" the report suggests (typically: run `canary scan`, add tokens to hot paths, `canary index`).

## Flags

- `--root <dir>` — root directory to analyze (default `.`)
- `--json` — compact JSON output
- `--limit <n>` — max entries per list section (default 20)

## Guidelines

- `onboard` is read-only and heuristic — it does not write tokens or a database. Follow it with `canary scan` and `canary index` once you start adding tokens.
- If the codebase already has ad-hoc migration notes, they'll show up as `CANARY:MIGRATE` findings; fold them into real tokens as you go.
