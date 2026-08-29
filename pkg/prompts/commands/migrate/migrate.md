# Migrate Command Prompt

## Purpose
Document the family of commands used to migrate a codebase onto (or between
versions of) CANARY: orphan spec generation, legacy token-shape upgrades,
fresh-codebase onboarding, and spec-kit/legacy-canary system migration.

Note: `canary migrate <steps>` (no subcommand) is a separate, narrower
command that applies *database schema* migrations (`--db`, steps as an
integer or `all`) -- it is not covered here.

## Task
Implement the orphan/upgrade/onboard/migrate-from surface truthfully.

## orphan detect | run
Requirements with CANARY tokens but no `.canary/specs/` spec file.
```bash
canary orphan detect [--show-features] [--root .] [--db <path>]
canary orphan run <REQ-ID> [--dry-run]
canary orphan run --all [--dry-run]
```
`run` generates `spec.md` (from existing token metadata) and `plan.md`
(reflecting current implementation state) under
`.canary/specs/<REQ-ID>-<name>/`.

## upgrade
Rewrites legacy on-disk token shapes (markdown headings, unicode hyphens,
unpadded IDs, bare IDs, bug tokens missing `FEATURE=`, `STATUS=FIXED`, missing
`UPDATED=`, old multi-line bug-create continuations) into the current
parseable form. Dry run by default.
```bash
canary upgrade --root . [--write] [--rule <name>] [--map <file>] [--limit 20]
```
With `--map {"OLD-ID":"NEW-ID",...}`, also remaps requirement IDs across
tokens and `GAP_ANALYSIS.md` "✅ &lt;ID&gt;" claim lines.

## onboard
Fresh-codebase adoption analysis for a repo with few or no tokens yet:
language histogram, top directories, entry points, existing token count
(heuristic), pre-seeded `CANARY:MIGRATE` notes, configured sources, and a
next-steps checklist. Read-only.
```bash
canary onboard --root . [--json] [--limit 20]
```

## migrate-from
Migrate an existing spec-kit or legacy-canary project onto the unified
system: creates `.canary/`, copies/merges templates and configuration,
preserves existing tokens and docs, and creates missing files (constitution,
slash commands, etc.).
```bash
canary migrate-from <spec-kit|legacy-canary> [directory] [--dry-run] [--force]
```

## Standards
- `orphan run` and `upgrade --write` must be safe to re-run (idempotent on already-migrated content).
- `upgrade` is dry-run by default; never write without `--write`.
- `migrate-from` should preserve everything it doesn't understand rather than deleting it.
