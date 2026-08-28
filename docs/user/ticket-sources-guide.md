# Ticket Sources Guide

**Feature:** Ticket-system requirement sources (JIRA, GitHub, GitLab)
**Status:** Production Ready
**Last Updated:** 2026-08-28

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
