---
description: Search CANARY tokens by keyword across feature names, req IDs, keyword tags, file paths, test and bench names
---

## User Input

```text
$ARGUMENTS
```

## Outline

1. **Read project key**: `.canary/project.yaml`.
2. **Run**:
   ```bash
   canary search <keywords>
   canary search <keywords> --json
   canary search <keywords> --limit -1     # unlimited (default 20)
   ```
3. Matching is case-insensitive (`LIKE`) across feature, req ID, keywords field, file path, test name, and bench name.
4. Requires an indexed database (`canary index`) — the search runs against `.canary/canary.db`, not a live scan.

## Flags

- `--limit <n>` — default 20 to protect agent context; `-1` = unlimited
- `--json` — output as JSON
- `--db <path>` — custom database path (default `.canary/canary.db`)

## Example

```bash
canary search authentication
canary search "oauth2" --limit -1 --json
```

## Guidelines

- Prefer `search` over `grep` for fuzzy, cross-field keyword matches; prefer `grep` when you know exactly which field/pattern you're matching.
- If results are truncated at the default limit, the output notes it — pass `--limit -1` to see everything.
