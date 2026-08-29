---
description: Compute (and, with credentials, apply) a plan reconciling CANARY token state against configured ticket sources (e.g. JIRA)
---

## User Input

```text
$ARGUMENTS
```

## Outline

`canary ticket` has two subcommands: `sync` (compute/apply a reconciliation plan) and `status` (report or refresh the cached remote status, no plan involved).

### `sync`

`canary ticket sync` reads indexed tokens plus the `sources:` list in `.canary/project.yaml` and computes two kinds of actions:

- **transition** — a non-flatfile requirement (e.g. JIRA) whose rollup status no longer matches its remote issue's status.
- **create_issue** — a flatfile requirement, paired with a `remap` action, when at least one non-flatfile source is configured — promoting it out of the flatfile series.

```bash
canary ticket sync                              # plan only, written to .canary/ticket-plan.json
canary ticket sync --project KEY --apply        # apply against JIRA (needs credentials)
canary ticket sync --limit 50                   # raise the table's action cap (default 20)
```

**No-credentials degradation**: without `JIRA_BASE_URL`, `JIRA_EMAIL`, and `JIRA_API_TOKEN` all set, this command NEVER errors and NEVER touches the network — it always writes the plan and prints a plan-only summary, even with `--apply`.

**With credentials AND `--apply`**: the effective project for `create_issue` actions is `--project` when set, otherwise the configured **destination** source's `project:` field (the source with `destination: true` in `sources:`, or — when none is marked — the first non-flatfile source). Remote status is fetched and merged across every configured jira-type source that has its own `project:` set (falling back to `--project` for sources that don't), so one sync can cover multiple JIRA projects. Applies `create_issue`/`transition` actions via the JIRA REST API, fills created keys into their paired remap actions, and writes the completed plan plus a remap map (`<plan>.map.json`) — feed that straight into `canary upgrade --map <plan>.map.json --write`. On a successful fetch, also writes `.canary/remote-status.json` — the same cache `ticket status` reports.

### `status`

Report or refresh the on-disk remote-status cache without computing or applying a sync plan:

```bash
canary ticket status                       # report the cache: entry count, fetched_at, age (no network)
canary ticket status --refresh             # fetch current status from every configured jira-type source, overwrite the cache
canary ticket status --refresh --project KEY   # project fallback for sources without their own project:
```

Same no-credentials degradation as `sync`: `--refresh` without JIRA credentials never errors, never touches the network, and never touches the existing cache file.

## Flags

`sync`: `--plan <path>` (default `.canary/ticket-plan.json`), `--apply`, `--project <key>`, `--issue-type <type>` (default `Story`), `--limit <n>` (default 20), `--db <path>`

`status`: `--refresh`, `--project <key>`

## Guidelines

- Run `sync` without `--apply` first and review the plan file.
- Chain `sync`'s output into `canary upgrade --map <plan>.map.json --write` to apply ID remaps after issue creation.
- Use `ticket status` (plain) to check cache freshness before relying on external-dependency resolution in `next`/`deps`/`view`; use `--refresh` to update it without running a full sync.
