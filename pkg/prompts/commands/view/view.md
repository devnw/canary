# View Command Prompt

## Purpose
Full picture of one requirement in a single, bounded call.

## Task
Implement `canary view <REQ-ID>` to aggregate everything known about a
requirement -- tokens/status, files, tests, benches, dependencies, spec/plan
paths, diagram refs, MIGRATE notes, drift, and ticket URL -- instead of
requiring separate `show`/`status`/`files`/`deps` calls.

## Expected Behavior
```bash
canary view CBIN-105
canary view CBIN-105 --json
canary view CBIN-105 --limit 20   # raise per-section cap (default 10)
```

## Output Format
```
CBIN-105 — 61% complete (source: core)
Tokens:   IMPL=5 TESTED=8
Features: UserAuth (API, TESTED); SessionStore (Storage, IMPL)
Files:    src/api/auth.go, src/storage/session.go, ...
Tests:    TestCANARY_CBIN_105_UserAuth
Depends:  CBIN-010, CBIN-020
Spec:     .canary/specs/CBIN-105-user-auth/spec.md
Plan:     .canary/specs/CBIN-105-user-auth/plan.md
Migrate:  legacy/auth.go:12: predates token system
Ticket:   https://your-org.atlassian.net/browse/PLAT-4521
```

Sections (`files`, `diagrams`, `migrate_notes`, etc.) are independently capped
at `--limit` (default 10) with a `_total` field reporting the true count and a
truncation hint when the cap was hit.

## Data Sources
- Tokens/status/files/tests/benches: `.canary/canary.db` (run `canary index` first)
- Dependencies: `.canary/specs/<REQ-ID>-*/spec.md` (forward, via the specs dependency graph)
- Diagrams and MIGRATE notes: indexed `diagram`/`migrate` refs in the database
- Drift: git history comparison against the token's `UPDATED` date
- Ticket URL: `.canary/project.yaml` `sources:` registry, when the REQ-ID matches a non-flatfile source

## Standards
- Never error on missing optional sections (no spec, no deps, no ticket source) -- omit them instead.
- Keep the default response small (protect agent context); every list section respects `--limit`.
- `--json` output must be valid, compact JSON mirroring the human-readable fields 1:1.
