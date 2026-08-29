# Ticket Sources Guide

**Feature:** Ticket-system requirement sources (JIRA, GitHub, GitLab)
**Status:** Production Ready
**Last Updated:** 2026-08-29

## Overview

By default, CANARY requirement IDs come from a local flatfile series (e.g. `CBIN-105`). A project can also declare additional **sources** so `REQ=` fields reference tickets in an external tracker — JIRA, GitHub, or GitLab — using the same token grammar. No new syntax is introduced; only the ID prefix changes.

## Configuring Sources

Declare sources in `.canary/project.yaml` under `sources:`. Every prefix used in a `REQ=` field should be declared here.

```yaml
sources:
  - name: core
    type: flatfile
    key: "CBIN"
  - name: platform
    type: jira
    key: "PLAT"
    url: "https://company.atlassian.net/browse/{id}"
  - name: app
    type: gitlab
    key: "GL"
    url: "https://gitlab.com/group/project/-/issues/{num}"
  - name: oss
    type: github
    key: "GH"
    url: "https://github.com/owner/repo/issues/{num}"
```

| Field | Description |
|-------|-------------|
| `name` | Internal label for the source. |
| `type` | `flatfile` (local series), `jira`, `github`, or `gitlab`. |
| `key` | The prefix used in `REQ=` fields for this source (e.g. `PLAT`). |
| `url` | Optional URL template. `{id}` = full ID (e.g. `PLAT-4521`), `{num}` = numeric part only (e.g. `4521`). Not needed for `flatfile`. |

If the `sources:` section is omitted entirely, a `flatfile` source is synthesized from `project.key` (default `CBIN`). Once you declare a `sources:` list, exactly the listed sources are used — include a `flatfile` entry explicitly if you still want local requirement IDs.

## Destination Project (`project:` / `destination:`)

A source can declare which ticket-system project it creates issues in, and be marked as the **destination** for promoting local (flatfile) requirements out into a ticket system:

```yaml
sources:
  - name: core
    type: flatfile
    key: "CBIN"
  - name: platform
    type: jira
    key: "PLAT"
    url: "https://company.atlassian.net/browse/{id}"
    api: "https://company.atlassian.net"
    project: "PLAT"
    destination: true
```

| Field | Description |
|-------|-------------|
| `project` | The ticket-system project key this source creates issues in and fetches remote status for. Optional; a source without it contributes no project of its own to `canary ticket sync`. |
| `destination` | Marks this source as the target for `create_issue` actions that promote a flatfile requirement. At most one source may set `destination: true` — `canary ticket sync` refuses a config with more than one. |

**Precedence** for `canary ticket sync --apply`'s creation project: the `--project` flag, when set, always wins; otherwise the destination source's `project:` — the source with `destination: true`, or, when none is marked, the first non-flatfile source in declaration order. `--project` is only required when a `create_issue` action exists and neither resolves to a project.

Remote status is fetched and merged across **every** configured jira-type source that has its own `project:` set (falling back to `--project` for sources that don't) — a single `ticket sync` run can cover multiple JIRA projects at once.

## `canary ticket status`

Report or refresh the on-disk remote-status cache (`.canary/remote-status.json`) without computing or applying a sync plan:

```bash
canary ticket status                       # report the cache: entry count, fetched_at, age
canary ticket status --refresh             # fetch current status from every configured jira-type source and overwrite the cache
canary ticket status --refresh --project KEY   # project fallback for sources without their own project:
```

Plain `canary ticket status` never touches the network — a missing cache is reported (`CANARY_TICKET_STATUS cached=0 reason=no_cache`), not an error. `--refresh` degrades the same way `ticket sync --apply` does: without JIRA credentials (`JIRA_BASE_URL`, `JIRA_EMAIL`, `JIRA_API_TOKEN` all set, or a source `api:` field as the `JIRA_BASE_URL` fallback), it never errors, never touches the network, and never touches the existing cache file — it prints `CANARY_TICKET_STATUS cached=0 reason=no_credentials`.

The cache file is also written as a side effect of `canary ticket sync --apply` on a successful fetch, so a normal sync run keeps it fresh without a separate `ticket status --refresh` call.

## External Dependencies

A requirement's `DEPENDS_ON=`/spec dependency can point at an ID owned by a configured ticket source (or an unconfigured prefix a peer project owns — see below) instead of a local flatfile ID. `canary next`, `canary deps check`/`validate`, `canary view`, and the mermaid graph all resolve such an ID against the cached remote status into one of three states:

| State | Meaning |
|-------|---------|
| `satisfied` | The cached remote status is in the source's "done" set (its `TESTED`/`BENCHED` `status_map` entries, default `Done`). |
| `unsatisfied` | The cached remote status is present but not in the done set. |
| `unknown` | No cache file, or no cached entry for that ID — run `canary ticket status --refresh`. |

**Local tokens always win**: an ID with at least one local CANARY token in the index is treated as a local requirement, never as external, even when its prefix also matches a configured ticket source's `key`.

**Unknown never blocks by default**: `unsatisfied` external dependencies block `canary next`'s selection and `canary deps check`'s exit code the same as a local unmet dependency; `unknown` never blocks either by default — degradation is sacred, since a missing/stale cache must never stall an agent. `canary deps validate` is more permissive still: by default it fails on neither `unsatisfied` nor `unknown` external counts (they're informational on the `external:` summary line only). Two flags opt into stricter behavior:

```bash
canary next --strict-external            # unknown external deps also block selection
canary deps validate --strict-external   # fail (non-zero exit) when any external dep is unsatisfied or unknown
```

`canary deps check <REQ-ID>` prints each external dependency's resolution on its own line:

```
✔ external ENG-12 (Done)
✖ external ENG-13 (In Progress)
? external ENG-14 (no cached ticket status)
```

`canary deps validate` reports aggregate counts on a separate summary line, and never treats an external ID as a "missing spec" error:

```
external: satisfied=3 unsatisfied=1 unknown=2
```

`canary view <REQ-ID>` annotates each external entry in `Depends:` the same way, e.g. `ENG-12 (external: Done)`, and `canary deps graph --format mermaid` styles external nodes with a dashed `classDef external` border.

## Peer Projects (`peers:`)

For inter-dependent repos, declare sibling projects at the top level of `.canary/project.yaml` — each peer's own `canary scan --out status.json` is read (read-only, never written to) to resolve requirement IDs that peer owns, including IDs under a prefix this project's own `sources:` list doesn't recognize at all:

```yaml
peers:
  - name: platform-service
    root: "../platform-service"
```

| Field | Description |
|-------|-------------|
| `name` | Label shown in resolution details, e.g. `peer:platform-service`. |
| `root` | The peer project's root directory; relative paths resolve against this project's own root. Its `<root>/status.json` is read. |

**Resolution order**: for an ID that isn't a known local flatfile ID, every configured peer is consulted, in declaration order, **before** the ticket-status cache — the first peer with an entry for that ID wins. A requirement found with at least one `TESTED`/`BENCHED` feature is `satisfied` (Detail `peer:<name>`); found but not is `unsatisfied` (Detail `peer:<name> (<worst status>)`). A missing or malformed peer `status.json` is a soft skip — Resolve falls through to the next peer, then the ticket cache, never erroring.

## Referencing Tickets in Tokens

Use the source's `key` as the `REQ=` prefix — the token format is unchanged:

```go
// CANARY: REQ=PLAT-4521; FEATURE="SSOLogin"; ASPECT=API; STATUS=IMPL; UPDATED=2026-08-28
```

External IDs are written **verbatim, never zero-padded**. `PLAT-4521` stays `PLAT-4521` (not `PLAT-04521`); `GH-7` stays `GH-7` (not `GH-007`).

## What Changes in Output

- **`status.json`**: each requirement entry gains a `source` field (the matching source `name`) and, when a `url` template is configured, a `ticket_url` field with the resolved link.
- **`canary view <REQ-ID>`**: prints a `Ticket:` line with the resolved URL when the requirement's source has one configured.
- **`canary scan --verify GAP_ANALYSIS.md`**: claim lines work the same for ticket-sourced IDs, e.g.:

  ```
  ✅ PLAT-4521 - SSOLogin fully tested
  ```

  This is verified exactly like a `CBIN-###` claim — status must be `TESTED` or `BENCHED`.

## Mermaid Diagram References

IDs found inside fenced ` ```mermaid ` code blocks and standalone `.mmd` files are indexed as diagram references for the matching requirement, the same way source-file tokens are indexed. This works for any configured source's prefix, not just `flatfile`.

`canary deps graph --format mermaid` emits a dependency graph where nodes for ticket-sourced requirements are clickable, linking out to the resolved `ticket_url` (JIRA/GitHub/GitLab) instead of only to internal spec/plan paths.

## Summary

- Declare each ticket-system prefix once in `.canary/project.yaml` `sources:`.
- Reference tickets with `REQ=<KEY>-<ID>` — same grammar, no padding.
- `status.json`, `canary view`, `GAP_ANALYSIS.md` verification, and mermaid diagrams all understand ticket-sourced requirements alongside the local flatfile series.
- Mark one source `destination: true` (or rely on the first non-flatfile default) to control where `ticket sync` creates new issues; `canary ticket status [--refresh]` reports or refreshes the remote-status cache independently of a sync run.
- External dependencies (ticket-sourced or peer-owned) resolve to satisfied/unsatisfied/unknown; `unknown` never blocks `next`/`deps check` by default, and `deps validate` doesn't block on either without `--strict-external`. `peers:` lets sibling repos resolve each other's requirement IDs read-only.
