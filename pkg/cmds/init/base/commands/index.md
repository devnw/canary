---
description: Scan the codebase for CANARY tokens and store them in the SQLite database used by list/show/search/deps/view
---

## User Input

```text
$ARGUMENTS
```

## Outline

`canary index` populates `.canary/canary.db` from source. Most read commands (`list`, `search`, `deps`, `view`, `next`, `status`) query this database rather than re-scanning, so re-index after tokens change.

1. **Run**:
   ```bash
   canary index                     # scan "." into .canary/canary.db
   canary index --root <dir>        # scan a different root
   canary index --db <path>         # write to a custom database path
   ```
2. Confirm the reported token count matches expectations; a large drop usually means a skip-regex or path issue.

## Flags

- `--root <dir>` — root directory to scan (default `.`)
- `--db <path>` — path to database file (default `.canary/canary.db`)

## Example

```bash
canary index --root .
```

## Guidelines

- Run `canary index` after any batch of token edits, before relying on `list`/`search`/`view`/`deps` output.
- `canary scan` produces the human/CI-facing report (`status.json`); `canary index` is the database counterpart the other query commands depend on.
