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
   `--format mermaid` renders a flowchart with click-through ticket links where a source is configured; `--status` marks each dependency ✅/❌ satisfied.
3. **reverse `<REQ-ID>`** — What would be blocked if this requirement changes?
   ```bash
   canary deps reverse CBIN-146
   ```
4. **validate** — Check the whole graph for cycles, self-deps, and missing requirements.
   ```bash
   canary deps validate
   ```

## Examples

```bash
canary deps graph CBIN-147 --format mermaid
canary deps check CBIN-147 --show-satisfied
canary deps validate
```

## Guidelines

- `check`/`graph` build the graph from `.canary/specs/`; requirements without specs won't appear.
- For a quick "what does X depend on" answer without the tree rendering, prefer the MCP `deps` tool (IDs only) or `canary view <REQ-ID>` (includes forward deps).
