---
description: Rewrite legacy on-disk CANARY token shapes into the current parseable form; dry-run by default
---

## User Input

```text
$ARGUMENTS
```

## Outline

Fixes tokens written in older/incompatible shapes: markdown `# CANARY:` headings, unicode hyphens inside IDs, unpadded flatfile IDs, bare legacy ID segments, bug tokens missing `FEATURE=`, `STATUS=FIXED` (not a valid status), missing `UPDATED=`, and the old multi-line bug-create continuation shape. With `--map`, also remaps old requirement IDs to new ones across tokens and `GAP_ANALYSIS.md` "✅ &lt;ID&gt;" claim lines.

1. **Preview** (dry run, default — nothing written):
   ```bash
   canary upgrade --root .
   ```
2. **Apply**:
   ```bash
   canary upgrade --root . --write
   ```
3. **Scope to one rule** (repeatable):
   ```bash
   canary upgrade --rule status-fixed --write
   ```
4. **Remap IDs** (e.g. after merging into a new project's numbering):
   ```bash
   canary upgrade --map id-map.json --write   # {"OLD-ID":"NEW-ID",...}
   ```

## Flags

- `--root <dir>`, `--write` (default: dry run), `--map <file>`, `--rule <name>` (repeatable), `--limit <n>` (default 20, bounds printed changes)

## Rules

`join-multiline`, `md-heading`, `unicode-hyphen`, `bare-id`, `bug-alias`, `status-fixed`, `pad-flatfile`, `add-updated`, `remap`

## Guidelines

- Always run without `--write` first and review the diff-like preview before applying.
- Pair `--map` with `canary ticket sync`'s `.map.json` output to apply a JIRA-driven remap.
