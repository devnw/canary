---
description: Full picture of one requirement in a single call -- tokens, files, tests, deps, spec, diagrams, ticket
---

## User Input

```text
$ARGUMENTS
```

## Outline

Prefer this over separate `show`/`status`/`files` calls when you need the full context for a requirement.

1. **Read project key**: `.canary/project.yaml` determines the ID prefix (e.g. `CBIN-105`).
2. **Run**:
   ```bash
   canary view <REQ-ID>
   canary view <REQ-ID> --json          # machine-readable
   canary view <REQ-ID> --limit 20      # raise per-section cap (default 10)
   ```
3. **Read the sections returned**: completion %, tokens by status, files, tests, dependency IDs (forward), spec/plan paths, diagrams, ticket URL (if a source is configured).
4. **Sections are capped** at `--limit` entries each; a truncation note tells you to raise it if you need more.

## Flags

- `--json` — compact JSON output
- `--limit <n>` — max entries per list section (default 10)

## Example Output

```
CBIN-204 — 61% complete (source: core)
Tokens:   IMPL=5 TESTED=8
Features: RequirementView (API, TESTED); ...
Files:    ./mcp/mcp.go, ./mcp/tools_extended.go, ...
Tests:    TestCANARY_CBIN_204_MCPView, ...
```

## Guidelines

- Use `view` FIRST for any "tell me about REQ-ID" request; fall back to `show`/`status`/`files`/`deps` only for a narrower question.
- If dependencies are shown, follow up with `canary deps graph <REQ-ID>` for the full tree.
