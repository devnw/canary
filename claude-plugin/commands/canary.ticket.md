---
description: Compute (and, with credentials, apply) a plan reconciling CANARY token state against configured ticket sources (e.g. JIRA)
---

## User Input

```text
$ARGUMENTS
```

## Outline

`canary ticket sync` reads indexed tokens plus the `sources:` list in `.canary/project.yaml` and computes two kinds of actions:

- **transition** — a non-flatfile requirement (e.g. JIRA) whose rollup status no longer matches its remote issue's status.
- **create_issue** — a flatfile requirement, paired with a `remap` action, when at least one non-flatfile source is configured — promoting it out of the flatfile series.

```bash
canary ticket sync                              # plan only, written to .canary/ticket-plan.json
canary ticket sync --project KEY --apply        # apply against JIRA (needs credentials)
canary ticket sync --limit 50                   # raise the table's action cap (default 20)
```

**No-credentials degradation**: without `JIRA_BASE_URL`, `JIRA_EMAIL`, and `JIRA_API_TOKEN` all set, this command NEVER errors and NEVER touches the network — it always writes the plan and prints a plan-only summary, even with `--apply`.

**With credentials AND `--apply`**: fetches remote status for `--project`, applies `create_issue`/`transition` actions via the JIRA REST API, fills created keys into their paired remap actions, and writes the completed plan plus a remap map (`<plan>.map.json`) — feed that straight into `canary upgrade --map <plan>.map.json --write`.

## Flags

`--plan <path>` (default `.canary/ticket-plan.json`), `--apply`, `--project <key>`, `--issue-type <type>` (default `Story`), `--limit <n>` (default 20), `--db <path>`

## Guidelines

- Run without `--apply` first and review the plan file.
- Chain the output into `canary upgrade --map <plan>.map.json --write` to apply ID remaps after issue creation.
