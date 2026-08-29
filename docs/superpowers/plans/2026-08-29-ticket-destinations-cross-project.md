# Ticket Destinations & Cross-Project Canaries Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Make the ticket destination (which JIRA project new issues are created in) a first-class canary configuration; add cross-project / inter-dependent project canary support (requirements depending on requirements owned by other sources/projects/repos); then dogfood by moving canary's incomplete tickets from CP to ENG and re-pointing the tool at ENG.

**Architecture:** `SourceConfig` gains `project` (destination project key) and `destination` (bool — the source new requirements promote into; exactly one). `canary ticket sync` derives creation/remote-status targets from source config (`--project` remains an override) and, after fetching, persists a remote-status cache (`.canary/remote-status.json`). Dependency machinery (`pkg/specs` status checking, `canary deps check/validate`, `canary next` blocking, `canary view`) becomes source-aware: an ID resolving to a non-destination/external source with no local tokens is an EXTERNAL dependency — satisfied/unsatisfied by cached ticket status when available, "external (unknown)" otherwise, and never blocking by default (`--strict-external` opts into failure). A `peers:` config lists sibling repo roots so external IDs can also resolve from a peer's status.json read-only.

**Tech Stack:** Go 1.27; existing pkg/config, pkg/sources, pkg/ticket, pkg/specs, pkg/cmds/{deps,next,view,ticket}; Atlassian MCP for the dogfood move.

## Global Constraints

- Frozen stdout contracts unchanged (`CANARY_SCAN…`, `CANARY_VERIFY_OK/FAIL`, exit 0/2/3); new lines only for new behavior.
- Degradation is sacred: no JIRA creds and no cache → external deps report unknown, nothing errors, nothing blocks by default.
- Small-by-default output limits with `--limit` raise + truncation hints on any new listing.
- Backwards compat: configs without `project`/`destination`/`peers` behave exactly as today; full suite green after every task (`go test ./... -count=1`); strict verify gate stays `CANARY_VERIFY_OK` on this repo.
- New-feature CANARY tokens use JIRA story IDs from the **ENG** project (created during the task via MCP by the controller — implementers will be handed the IDs in their dispatch; never invent IDs).
- Conventional commits + trailer `Co-Authored-By: Claude Fable 5 <noreply@anthropic.com>`; hooks may modify files (re-stage/retry); never commit `bin/` or `/tmp` artifacts; push only at ship time (Task 7).
- Standing release rule: lowest-segment bump only (v0.3.0 → v0.3.1), no confirmation sought.

## Background (read once)

- `pkg/config.SourceConfig{Name, Type, Key, URL, API, StatusMap}`; `pkg/sources.Source` mirrors it; `Registry.Resolve/TicketURL/Normalize/Pattern/Sources`; `FromProjectConfig` fallback chain (never breaks a scan).
- `pkg/cmds/ticket/ticket.go` `runTicketSync`: id_pattern-filtered `ListTokens`, `ticket.ComputePlan(tokens, reg, remoteStatus)`, creds env `JIRA_*` (BaseURL falls back to the matching source's `API`), `--project` required under `--apply` when mutable actions exist, `FetchRemoteStatus(client, project)` paged, partial-progress reporting, plan + `<plan>.map.json` outputs.
- `pkg/ticket.ComputePlan`: flatfile requirement + ≥1 non-flatfile source → `create_issue`+`remap` pair; non-flatfile requirement with rollup-status ≠ remoteStatus[id] → `transition`; worst-of rollup STUB<IMPL<TESTED<BENCHED; default status map overridable per-source.
- Dependencies: spec `## Dependencies` lines (any `[A-Z][A-Z0-9]*-\d+` prefix) → `pkg/specs` graph (`deps check/graph/reverse/validate`); token `DEPENDS_ON=` consumed by `pkg/cmds/next` `hasUnresolvedDependencies` (dep satisfied iff its tokens roll up TESTED/BENCHED; **a dep with zero local tokens = blocked**, which is exactly wrong for external ticket IDs) and mirrored in `mcp/tools.go` next; `view` lists DependsOn/Blocks/RelatedTo raw.
- `canary view` BuildView(dbPath, root, reqID, limit); mermaid renderer takes urlFor.
- JIRA: devnw.atlassian.net; CP = completed history (epics CP-229..234); ENG = destination for open/future work. MCP `jira_move_issue` exists (deferred tool).

---

### Task 1: Destination configuration (`project` + `destination` on sources)

**Files:** `pkg/config/config.go`, `pkg/sources/sources.go` (+tests), `pkg/ticket/plan.go`, `pkg/cmds/ticket/ticket.go` (+tests), docs touched in Task 5.

**Interfaces (produced):**
- `SourceConfig`/`Source` gain `Project string yaml:"project,omitempty"` and `Destination bool yaml:"destination,omitempty"`.
- `Registry.DestinationSource() (Source, bool)` — the source with Destination=true; when none is marked, the FIRST non-flatfile source; false when only flatfile sources exist. `NewRegistry` errors on >1 Destination=true.
- `ticket.Action` gains `Project string json:"project,omitempty"` — create_issue actions carry the destination source's Project.
- `runTicketSync`: effective creation project = `--project` flag > destination source's Project; effective remote-status fetch covers EVERY jira source that has a Project (map merged, keys disjoint by issue prefix); `--project` no longer strictly required under `--apply` when the destination source's Project is set (the guard now checks "no resolvable project for a mutable action").

- [ ] Step 1: failing tests — registry (DestinationSource default-first/marked/none/duplicate-error), ComputePlan stamping Project on create actions, CLI test: apply path with destination-configured source needs no `--project` flag (httptest); multi-source remote-status merge test.
- [ ] Steps 2-4: TDD; full suite; strict gate. Step 5: commit `feat(ticket): configurable destination project per source`.

---

### Task 2: Remote-status cache + external-dependency resolution engine

**Files:** Create `pkg/external/external.go` (+test); modify `pkg/cmds/ticket/ticket.go` (write cache after fetch), `pkg/ticket/plan.go` if needed.

**Interfaces (produced; consumed by Tasks 3-4):**
- Cache file `.canary/remote-status.json`: `{"fetched_at":"RFC3339","statuses":{"CP-240":"Done",...}}`. `ticket sync --apply` (and a new `canary ticket status --refresh` subcommand that ONLY fetches+caches) write it.
- `package external`: `type Resolution struct { ID, State, Detail string }` with State ∈ `satisfied | unsatisfied | unknown`; `func Resolve(id string, reg *sources.Registry, root string) Resolution`:
  1. id resolves to a flatfile source or no source → State unknown, Detail "not external" (callers treat as local).
  2. ticket-source id: read cache; status in the source's DoneSet (statuses mapped to Done via reversed StatusMap, default {"Done"}) → satisfied; present-but-not-done → unsatisfied with Detail the remote status; absent/no cache → unknown with Detail "no cached ticket status (run canary ticket status --refresh)".
  3. Peer resolution (Task 4) layers in front of ticket status.
- Never performs network I/O (cache-only) — degradation sacred.

- [ ] TDD: cache round-trip; Resolve satisfied/unsatisfied/unknown; stale-cache tolerated (fetched_at surfaced in Detail when older than 24h). `canary ticket status --refresh` httptest test + no-creds degradation (`CANARY_TICKET_STATUS cached=0 reason=no_credentials`, exit 0). Commit `feat(external): cached ticket-status resolution for external dependencies`.

---

### Task 3: Source-aware dependencies in deps / next / view / MCP

**Files:** `pkg/specs/` (status checker adapter), `pkg/cmds/deps/deps.go`, `pkg/cmds/next/next.go`, `mcp/tools.go` (next parity), `pkg/cmds/view/view.go`, tests alongside each.

**Behavior:**
- `hasUnresolvedDependencies` (next, both CLI and MCP): for a dep ID with zero local tokens, consult `external.Resolve`: satisfied → not blocking; unsatisfied → blocking; unknown → NOT blocking (default) with a one-line stderr note the first time; new `--strict-external` on `next` makes unknown block.
- `canary deps check <id>`: external deps listed with their Resolution state (`✔ external ENG-12 (Done)` / `✖ external ENG-13 (In Progress)` / `? external ENG-14 (no cached status)`); `deps validate`: external IDs are never "missing spec" errors; new counts line `external: satisfied=N unsatisfied=M unknown=K`.
- `view`: DependsOn entries annotated `ENG-12 (external: Done)` when resolvable; `Drifted` unaffected.
- Mermaid graph: external nodes get `classDef external` styling + click URLs (already URL-capable).

- [ ] TDD: seeded cache + specs fixtures covering all three states in next-blocking, deps check output, validate counts, view annotation, mermaid class. Keep MCP next in lockstep (test). Commit `feat(deps): external ticket-source dependencies resolve via cached status`.

---

### Task 4: Peer projects (inter-dependent repos)

**Files:** `pkg/config/config.go` (`Peers []PeerConfig{Name, Root string}` yaml `peers:`), `pkg/external/external.go` (peer layer), `pkg/cmds/view/view.go` (peer hit annotation), tests.

**Behavior:** for an external or unknown-prefix ID, before ticket-cache lookup, scan configured peers: read `<peer.Root>/status.json` (canaryscan Report shape, read-only, missing file → skip); requirement found with any TESTED/BENCHED feature → satisfied (Detail `peer:<name>`); found but not → unsatisfied; not found → fall through to ticket cache. Peer roots resolved relative to the repo root; never written to. `canary deps check` Detail shows the peer name.

- [ ] TDD: temp peer dir with a status.json fixture; precedence test (peer beats ticket cache); missing/malformed peer file soft-skip. Commit `feat(external): peer-project resolution for inter-dependent repos`.

---

### Task 5: Docs & plugins for destinations, external deps, peers

**Files:** `docs/user/ticket-sources-guide.md` (destination/project/peers sections), `claude-plugin/commands/canary.ticket.md` + `canary.deps.md` + base trios (byte-identical), `claude-plugin/CLAUDE.md` + skills (one-line updates), `pkg/cmds/init/base/project.yaml` template (commented `project:`/`destination:`/`peers:` examples), cursor/copilot parity lines.

- [ ] Verify every documented flag against `--help` output; base commands ↔ templates dirs stay byte-identical; commit `docs(plugins): ticket destinations, external dependencies, and peer projects`.

---

### Task 6: Dogfood — ENG as destination, move incomplete CP tickets (controller-coordinated)

- [ ] `.canary/project.yaml`: add ENG jira source `{name: eng, type: jira, key: "ENG", url: https://devnw.atlassian.net/browse/{id}, api: https://devnw.atlassian.net, project: "ENG", destination: true, status_map like CP}`; CP source loses destination (history only); id_pattern gains `ENG-[0-9]+`.
- [ ] Via MCP: move every NOT-completed CP canary story to ENG — the 27 "In Progress" stories + backlog follow-ups CP-286..292 (`jira_move_issue`; collect old→new key map; verify status preserved, re-transition if the move resets it). Done stories (20 + CP-285) stay in CP.
- [ ] `canary upgrade --root . --map <move-map> --write` remaps moved REQ ids in tokens/claims; `canary index`; strict verify green; `canary ticket status --refresh`-equivalent via MCP-written cache (controller may write `.canary/remote-status.json` from MCP search output — same schema) so `canary deps`/`next` demonstrate external resolution.
- [ ] Full gates + commit `chore(dogfood): ENG destination, incomplete tickets moved CP->ENG`.

---

### Task 7: Ship v0.3.1

- [ ] Final whole-branch review (most capable model); one consolidated fix wave for Critical/Important.
- [ ] Patch bump (standing rule, no confirmation): plugin manifests 0.3.0→0.3.1, merge to main, push origin+gh, tag v0.3.1, monitor pipelines green (heal loop), confirm proxy serves v0.3.1.

## Self-Review

Destination config (T1) covers "configuration of the actual ticket destination"; T2-4 cover cross-project/dependent/inter-dependent canaries (ticket-status external deps + peer repos); T6 executes the CP→ENG move for incomplete tickets; the standing patch-bump rule is encoded in T7 and Global Constraints. Interfaces consistent: `Source.Project/Destination` (T1) consumed by T2 cache writer and T6 config; `external.Resolve` (T2) consumed by T3/T4; peer config (T4) documented in T5.

## Known adaptation points

Real shapes of `hasUnresolvedDependencies` in next.go and its MCP mirror; deps check/validate output formats; whether jira_move_issue preserves status (T6 verifies live); ComputePlan's Action ordering tests when adding Project field.
