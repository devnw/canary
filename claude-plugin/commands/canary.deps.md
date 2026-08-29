---
description: Check, visualize, and validate requirement dependencies declared in specs
---

## User Input

```text
$ARGUMENTS
```

## Outline

`canary deps` has four subcommands; there is no `show`/`list`/`add` -- dependencies are declared in spec files, not added via CLI.

1. **check `<REQ-ID>`** — Is this requirement's dependencies satisfied? Only `TESTED`/`BENCHED` count; `IMPL` is insufficient.
   ```bash
   canary deps check CBIN-147 [--show-satisfied]
   ```
2. **graph `<REQ-ID>`** — Visual dependency tree (direct + transitive).
   ```bash
   canary deps graph CBIN-147 [--status] [--format ascii|mermaid]
   ```
   `--format mermaid` renders a flowchart with click-through ticket links where a source is configured, and a dashed `classDef external` border on nodes resolving to an external (ticket-source or peer-owned) dependency; `--status` marks each dependency ✅/❌ satisfied.
3. **reverse `<REQ-ID>`** — What would be blocked if this requirement changes?
   ```bash
   canary deps reverse CBIN-146
   ```
4. **validate** — Check the whole graph for cycles, self-deps, and missing requirements.
   ```bash
   canary deps validate [--strict-external]
   ```
   External IDs (ticket-source or peer-owned) are never "missing spec" errors — they're counted on a separate `external: satisfied=N unsatisfied=M unknown=K` line. `--strict-external` fails validation when any external dependency is unsatisfied or has unknown (uncached) status; by default `validate` reports counts only and never fails on either (unlike `next`/`deps check`, where unsatisfied always blocks).

## External Dependencies

A dependency ID with zero local CANARY tokens is resolved against `.canary/remote-status.json` (see `canary ticket status`) or a configured peer project, never against a stale/unrelated local match — **local tokens always win**. `deps check` prints each external dependency's resolution on its own line:

```
✔ external ENG-12 (Done)
✖ external ENG-13 (In Progress)
? external ENG-14 (no cached ticket status)
```

## Examples

```bash
canary deps graph CBIN-147 --format mermaid
canary deps check CBIN-147 --show-satisfied
canary deps validate
canary deps validate --strict-external
```

## Guidelines

- `check`/`graph` build the graph from `.canary/specs/`; requirements without specs won't appear.
- For a quick "what does X depend on" answer without the tree rendering, prefer the MCP `deps` tool (IDs only) or `canary view <REQ-ID>` (includes forward deps, with external ones annotated).
