# Migration Guidance, Onboarding, Drift Tracking, JIRA Source-of-Truth & Plugin Repair Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add `CANARY:MIGRATE` guidance comments that migration tooling feeds to agents; make canary able to onboard fresh non-canary codebases and upgrade every legacy token shape on disk; make drift detection and date tracking first-class across CLI/tools/plugins; codify ticket synchronization (JIRA) into the tool; repair all 47 audited plugin/prompt defects; then dogfood everything by migrating canary itself to JIRA (project CP on devnw.atlassian.net) as its requirement source of truth.

**Architecture:** A second token keyword `CANARY:MIGRATE <text>` is discriminated before KV parsing in both scanners, stored as `refs` rows (`kind='migrate'`), surfaced in status.json, `canary view`, and the new migration commands. New commands: `canary onboard` (foreign-tree analysis producing an agent-ready adoption plan), `canary upgrade` (on-disk legacy-token rewriter with `--map` ID remapping), `canary drift` (git-date vs UPDATED + staleness + doc-status rollup), `canary ticket sync` (computes create/transition/remap actions from token state vs the `sources:` config; applies via JIRA REST when creds exist, else emits a machine-readable plan for an agent to apply via MCP). Staleness threshold moves from hardcoded 30 to config. Plugin/prompt assets are reconciled to one modern lineage.

**Tech Stack:** Go 1.27, cobra, sqlite (existing `refs` table), `pkg/sources` registry, JIRA Cloud REST v3 (net/http, no new deps), Atlassian MCP for agent-applied actions.

## Global Constraints

- Frozen stdout contracts unchanged: `CANARY_SCAN tokens=N requirements=M STUB=... IMPL=... TESTED=... BENCHED=...`, `CANARY_VERIFY_OK`, `CANARY_VERIFY_FAIL count=N`, exit codes 0/2/3. New commands may add NEW one-line summaries (e.g. `CANARY_DRIFT ...`) but never alter existing ones.
- A `CANARY:MIGRATE` line must NEVER abort a scan. Both `pkg/canaryscan/parse.go` and `gate/scanner.go` currently route every `CANARY:` line into KV parsing where a keyword line becomes `bad kv segment` → exit 3 (`pkg/canaryscan/scan.go:109`, `run.go:78`). The discriminator lands BEFORE any other MIGRATE work merges.
- Small-by-default output (standing user directive): every new listing surface defaults ≤20 items with `--limit` (`-1` unlimited) and prints `… +N more (use --limit)` hints on truncation.
- Backwards compat: repos without MIGRATE notes, without `sources:`, without JIRA behave exactly as today; full suite green after every task (`go test ./... -count=1`).
- TDD per task; license headers on new files; CANARY tokens on new features using NEW JIRA-era IDs is NOT yet possible (chicken/egg) — new features in Tasks 1–10 use flatfile IDs CBIN-301..309 (`UPDATED=2026-08-29`), which Task 12's remap migrates to CP keys along with everything else.
- Conventional commits + trailer `Co-Authored-By: Claude Fable 5 <noreply@anthropic.com>`; pre-commit hooks may modify files (re-stage and retry). Push to `origin` AND `gh` only at the very end (Task 13), not per task.
- External ticket IDs never zero-padded; registry (`pkg/sources`) is the only ID authority.
- Do not fix the 5 MCP stub tools' implementations (specify/plan/index/bug-create/gap-mark) — docs must簡 label them stubs; implementing them is out of scope.
- JIRA specifics for dogfood: site `https://devnw.atlassian.net`, project key `CP` (CodePros). The canary CLI's REST client reads `JIRA_BASE_URL`, `JIRA_EMAIL`, `JIRA_API_TOKEN` env vars; when absent it must degrade to plan-only output, never error.

## Background (read once — exploration findings that tasks cite)

- Scanner token regex `pkg/canaryscan/parse.go:15` matches only `//`, `#`, `/*` prefixes; `gate/scanner.go:80` additionally matches `--`, `[//]: #`, `<!--`, `*` and strips trailing `-->`/`*/`. Placeholder segments containing `<`, `>`, `{{`, `}}`, `%s` are silently dropped (`parse.go:42-44`).
- `refs` table exists (migration 000006): columns `req_id, kind, file_path, line_number, context`; API `ReplaceRefs(kind, refs)`, `GetRefsByReqID` in `pkg/storage/refs.go`. `req_id` is NOT NULL — MIGRATE notes not tied to a requirement store `req_id=''`.
- Legacy shapes the upgrade rewriter must handle (verified): bare-ID segments (`REQ-42`, `TASK-7`, `REQ-GQL-4`, optional-hyphen `REQ1`) via `parse.go:41`; `BUG=BUG-API-001` aliasing (`parse.go:60-68`); unpadded flatfile numerics; unicode hyphens `‑`/`–`; `STATUS=FIXED`→`REMOVED` (`scan.go:157`); markdown `# CANARY:` headings (what `tools/convert-md-canary` fixes); multi-line bug tokens emitted by `pkg/cmds/bug/create.go:111-119` and `pkg/cmds/bug/bug.go:277-283` that the scanner cannot parse; missing `UPDATED=` (fatal to scan).
- Staleness: threshold hardcoded `30*24*time.Hour` at `pkg/canaryscan/run.go:88` and `run.go:126`; `verification.staleness_days` (`pkg/config/config.go:44`) read by NOTHING; `--update-stale` diag regex `REQ=([A-Z][A-Z0-9]*-\d+)` (`pkg/canaryscan/update.go:16`) fails entirely on v2 IDs like `CBIN-CLI-001` → silent no-op with a bogus success message (`run.go:95` prints diag count as "Updated N"); `UpdateStaleTokens` cannot add a missing `UPDATED=` (`update.go:76-80`); `pkg/cmds/bug/update.go:114` uses local-tz `time.Now()`.
- Git integration today: only `git rev-parse` in `pkg/cmds/index/index.go:49,54` and `pkg/cmds/checkpoint/checkpoint.go:51`. No `git log`/mtime anywhere.
- Migration machinery: `pkg/migrate/migrate.go` (spec-kit / legacy-canary LAYOUT migration; `readEmbeddedCandidate` at `migrate.go:18-40` reads from DISK not embed.FS — broken from installed binaries); `pkg/cmds/migrate/migrate.go` = `canary orphan detect|run` (help text still says `canary migrate`, `:59-60,:109-111,:131-138`); `pkg/cmds/legacy/migrate.go` = `canary migrate-from {spec-kit|legacy-canary}` + `canary detect`; `tools/convert-md-canary` unregistered one-off. NO onboarding path exists; NO token rewriter exists.
- `canary view` package `pkg/cmds/view` (`BuildView(dbPath, root, reqID, limit)`, `View` struct consumed by MCP `handleView`).
- Plugin audit: 47 findings, summarized inside Tasks 9–10 below with file:line.

---

### Task 1: `CANARY:MIGRATE` extraction + scanner discriminators

**Files:**
- Create: `pkg/canaryscan/migrate_notes.go`, `pkg/canaryscan/migrate_notes_test.go`
- Modify: `pkg/canaryscan/parse.go` (skip MIGRATE lines before parseKV), `pkg/canaryscan/scan.go` (collect notes during walk → `Report.MigrationNotes`), `pkg/canaryscan/types.go` (Report field), `gate/scanner.go` (skip MIGRATE lines), `gate/scanner_test.go` (regression test)

**Interfaces (produced; consumed by Tasks 2, 5, 8):**
```go
// MigrateNote is one CANARY:MIGRATE guidance comment.
type MigrateNote struct {
	File   string   `json:"file"`
	Line   int      `json:"line"` // 1-based
	Text   string   `json:"text"`
	ReqIDs []string `json:"req_ids,omitempty"` // registry-pattern matches inside Text, normalized
}
func ExtractMigrateNotes(relPath, content string, reg *sources.Registry) []MigrateNote
func ScanMigrateNotes(root string, skip *regexp.Regexp, ignorePatterns *ignore.GitIgnore, reg *sources.Registry) ([]MigrateNote, error)
// Report gains: MigrationNotes []MigrateNote `json:"migration_notes,omitempty"`
```

- [ ] **Step 1: Failing tests.** In `migrate_notes_test.go` (reuse `ticketRegistry(t)` from sources_test.go):
  - `TestCANARY_CBIN_301_ExtractMigrateNotes`: content with `// CANARY:MIGRATE auth flows move to pkg/auth; see CBIN-105`, `# CANARY:MIGRATE python-style note`, `<!-- CANARY:MIGRATE md note -->`, `-- CANARY:MIGRATE sql note`, and a plain `// CANARY: REQ=CBIN-105; ...` token line. Assert: 4 notes with correct 1-based lines; trailing `-->` stripped from md note text; the first note's `ReqIDs == []string{"CBIN-105"}`; the ordinary token line yields NO note.
  - `TestCANARY_CBIN_301_MigrateLineDoesNotAbortScan`: temp root with one valid token file AND a file containing only `// CANARY:MIGRATE free text with = signs; and semicolons`; `Scan(...)` returns nil error, report has the token AND `MigrationNotes` len 1.
  - In `gate/scanner_test.go`: `TestMigrateLineSkippedByGateScanner`: gate scanner over the same MIGRATE line produces zero tokens and zero errors.
- [ ] **Step 2: Run to verify FAIL** (`go test ./pkg/canaryscan/ ./gate/ -run 'Migrate' -v`).
- [ ] **Step 3: Implement.**
  - `migrate_notes.go` (license header + `// CANARY: REQ=CBIN-301; FEATURE="MigrateNotes"; ASPECT=Engine; STATUS=TESTED; TEST=TestCANARY_CBIN_301_ExtractMigrateNotes; UPDATED=2026-08-29`): line regex `^\s*(?://|#|--|<!--|/\*+|\*|\[//\]:\s*#)?\s*CANARY:MIGRATE\b[:\s]*(.*)$`; strip trailing `-->` / `*/`; `ReqIDs` = `reg.Pattern()` matches in text, `reg.Normalize`d, deduped in order. `ScanMigrateNotes` walks like `ScanDiagramRefs` but over ALL non-binary file extensions the scanner already reads (reuse the same skip/ignore checks; simplest: factor the walk or mirror `ScanDiagramRefs`'s structure without the .md-only filter; skip files >1MB).
  - `parse.go`: in the function that iterates `tokenLineRe` matches, before calling `parseKV`, `if strings.HasPrefix(strings.TrimSpace(m[1]), "MIGRATE") { continue }` (m[1] = capture after `CANARY:`). Mirror in `gate/scanner.go` where `scannerCanaryRe`'s capture is handed to its parseKV.
  - `scan.go`: after diagram-ref attachment, `notes, _ := ScanMigrateNotes(root, skip, ignorePatterns, reg)` (warn stderr on err via the same pattern used for diagram refs), sort by File,Line, assign to `rep.MigrationNotes`.
- [ ] **Step 4: PASS + full suite** (`go test ./... -count=1`; the repo's own scan gate must stay green: `go build -o bin/canary ./cmd/canary && ./bin/canary scan --root . --out status.json`).
- [ ] **Step 5: Commit** `feat(scan): CANARY:MIGRATE guidance notes extracted without breaking token scans`.

---

### Task 2: MIGRATE notes in index, view, and MCP

**Files:**
- Modify: `pkg/storage/refs.go` (+`GetRefsByKind`), `pkg/cmds/index/index.go` (store notes as refs kind='migrate', context=Text, req_id='' or per-ReqID rows), `pkg/cmds/view/view.go` (View gains `MigrateNotes []string` + `MigrateNotesTotal int`, capped, hinted), `mcp/tools_extended.go` (view summary mentions notes count), `pkg/storage/refs_test.go`, `pkg/cmds/view/view_test.go`
- Test: extend existing files.

**Interfaces:** `func (db *DB) GetRefsByKind(kind string, limit int) ([]*Ref, error)` (limit<=0 → 100 cap); View fields exactly `MigrateNotes []string \`json:"migrate_notes,omitempty"\``, `MigrateNotesTotal int \`json:"migrate_notes_total,omitempty"\`` — entries formatted `"file:line: text"`.

- [ ] Step 1: failing tests — refs round-trip for kind='migrate' with empty req_id + per-req rows; view test seeding `ReplaceRefs("migrate", ...)` rows for CBIN-105 asserting `MigrateNotes` contents and cap behavior mirroring Diagrams.
- [ ] Step 2: FAIL. Step 3: implement (index: one ref row per note per associated ReqID, plus one `req_id=''` row when none matched; `ReplaceRefs("migrate", ...)` alongside the diagram call; view: fetch `GetRefsByReqID` already returns all kinds — filter kind=='migrate' into MigrateNotes, kind=='diagram' into Diagrams — read the current view code and adjust so the two kinds don't cross-contaminate; that latent bug (view currently formats ALL refs as diagrams) must be fixed here with a regression test). Step 4: PASS + full suite. Step 5: Commit `feat(view): surface CANARY:MIGRATE notes through index, view, and MCP`.

---

### Task 3: Fix `canary bug create` unparseable multi-line tokens

**Files:** `pkg/cmds/bug/create.go:111-119`, `pkg/cmds/bug/bug.go:277-283`, tests in `pkg/cmds/bug/`.

Canary emits 4-line bug tokens whose continuation lines the scanner drops, then the required-field gate rejects the token. Emit ONE line: `// CANARY: BUG=<id>; TITLE="…"; ASPECT=…; STATUS=OPEN; SEVERITY=…; PRIORITY=…; UPDATED=YYYY-MM-DD`. Also switch `pkg/cmds/bug/update.go:114` to `time.Now().UTC()`.

- [ ] Step 1: failing test `TestCANARY_CBIN_302_BugTokenSingleLineParseable`: generate the token string via the same helper the command uses, feed it through `canaryscan` parse on a temp file, assert it scans without error and yields STATUS=OPEN with the BUG id as REQ. Step 2 FAIL → Step 3 implement → Step 4 PASS + full suite → Step 5 commit `fix(bug): emit single-line tokens the scanner can parse`.

---

### Task 4: `canary upgrade` — on-disk legacy token rewriter

**Files:**
- Create: `pkg/upgrade/upgrade.go`, `pkg/upgrade/upgrade_test.go`, `pkg/cmds/upgrade/upgrade.go`
- Modify: `cli/cmds.go` + root `cmds.go` (register), delete `tools/convert-md-canary/` (absorbed; note in commit message).

**Interfaces (produced; consumed by Tasks 8, 12):**
```go
package upgrade
type Change struct { File string; Line int; Old, New string; Rule string }
type Options struct {
	Root string; Skip *regexp.Regexp; Ignore *ignore.GitIgnore
	Registry *sources.Registry
	Map map[string]string // old REQ id -> new id (e.g. CBIN-101 -> CP-12); applied to REQ=/BUG= values AND ✅-claim lines
	Write bool            // false = dry run
	Today string          // YYYY-MM-DD for added UPDATED=; empty -> honor CANARY_TEST_TIMESTAMP then time.Now().UTC()
}
func Run(o Options) ([]Change, error)
```
Rules (each named, each independently tested): `md-heading` (`# CANARY:` → `<!-- CANARY: … -->`), `unicode-hyphen`, `pad-flatfile` (flatfile source IDs zero-padded ≥3), `bare-id` (`CANARY: REQ-42; …` → `CANARY: REQ=REQ-042; …`; same for TASK/BUG bare segments), `bug-alias` (`BUG=X` with no REQ → keep BUG= but ensure parseable single line), `status-fixed` (`STATUS=FIXED`→`STATUS=REMOVED`), `add-updated` (append `; UPDATED=<today>` when missing), `join-multiline` (continuation lines starting with comment prefix + `KEY=` folded into the token line — detect the exact shape `bug create` used to emit), `remap` (Map applied to REQ=/BUG= token values and to `^\s*✅\s+<old>` claim lines in .md files). Rules never touch lines inside code fences except the md-heading rule’s own target, and never touch `CANARY:MIGRATE`/`CANARY:START`/`CANARY:END` lines.

CLI: `canary upgrade [--root .] [--write] [--map file.json] [--rule name ...] [--limit 20]` — dry-run default printing per-change `file:line rule: old → new` (bounded, hint), summary line `CANARY_UPGRADE files=N changes=M written=<bool>`; `--map` reads a JSON object.

- [ ] Step 1: table-driven failing tests, one fixture per rule + one combined file + one `--map` remap incl. a GAP_ANALYSIS claim line + one no-op idempotency test (running twice yields zero changes second time). Step 2 FAIL → Step 3 implement → Step 4 PASS + full suite + self-check: `./bin/canary upgrade --root .` on THIS repo must report only expected changes (do not --write here; Task 11 does) → Step 5 commit `feat(upgrade): on-disk legacy CANARY token rewriter with ID remapping`.

---

### Task 5: `canary onboard` — fresh-codebase adoption analysis

**Files:**
- Create: `pkg/cmds/onboard/onboard.go`, `pkg/cmds/onboard/onboard_test.go`
- Modify: `cli/cmds.go` + root `cmds.go` (register)

`canary onboard [--root .] [--json] [--limit 20]` — read-only analysis for a repo with few/no tokens, producing the agent handoff:
- language histogram (by extension, top N), top-level directories with file counts, entry points (dirs containing `func main(` / `package.json` `"main"`/`"bin"` / `pyproject.toml` — cheap heuristics, best-effort),
- existing token count + existing `CANARY:MIGRATE` notes (via `canaryscan.ScanMigrateNotes` — the user's pre-seeded guidance is the whole point),
- configured sources + project key + the next available flatfile ID (`reqid.GenerateNextID` fallback `<KEY>-001`),
- a `next_steps` list (create `.canary` via `canary init` if absent; propose requirements per subsystem; `canary create` per requirement; `canary index && canary scan`).
Human output bounded ~30 lines; `--json` compact struct `OnboardReport{Languages, Directories, EntryPoints, TokenCount, MigrateNotes, Sources, NextID, NextSteps}`.

- [ ] Step 1: failing test on a synthetic polyglot temp tree (go main + md with MIGRATE note + python file), asserting languages/entrypoints/notes/NextID. Step 2 FAIL → 3 implement (license header, CANARY token CBIN-303, STATUS=TESTED) → 4 PASS + full suite → 5 commit `feat(onboard): adoption analysis for non-canary codebases`.

---

### Task 6: Staleness config + update-stale repairs

**Files:** `pkg/canaryscan/config.go` (+`StaleDays int`), `pkg/canaryscan/run.go` (use cfg; load `verification.staleness_days` from project.yaml when cfg.StaleDays==0; final fallback 30), `pkg/cmds/scan/scan.go` (+`--stale-days` flag, help text un-hardcode 30), `tools/canary/main.go` (flag parity), `pkg/canaryscan/update.go` (diag regex → `REQ=(\S+)`; support ADDING `; UPDATED=<date>` when the token line lacks one), tests in `pkg/canaryscan/`.

- [ ] Step 1: failing tests — `TestCANARY_CBIN_304_StaleDaysFromConfig` (project.yaml staleness_days: 7; a 10-day-old TESTED token trips strict; with 30 it wouldn't), `TestCANARY_CBIN_304_UpdateStaleV2IDs` (token `REQ=CBIN-CLI-001` stale → actually rewritten; asserts the previous silent no-op is gone), `TestCANARY_CBIN_304_UpdateStaleAddsMissingUpdated` (token line without UPDATED= gains one — note: craft the fixture via gate-scanner-tolerated shape or feed UpdateStaleTokens a synthetic diag, since canaryscan.Scan still hard-errors on missing UPDATED; read update.go first and design the test honestly). Also fix `run.go:95` to report the count of tokens actually rewritten, not diag count — assert that.
- [ ] Steps 2-4 TDD + full suite (existing `TestAcceptance_StrictStaleness` and tools acceptance must stay green — default remains 30 when nothing configured). Step 5 commit `fix(scan): configurable staleness window and a working update-stale for v2 and dateless tokens`.

---

### Task 7: `canary drift` — code-vs-token drift + rollup

**Files:**
- Create: `pkg/drift/drift.go`, `pkg/drift/drift_test.go`, `pkg/cmds/drift/drift.go`
- Modify: `cli/cmds.go` + root `cmds.go` (register), `pkg/cmds/view/view.go` (View gains `Drifted bool \`json:"drifted,omitempty"\`` + `DriftReason string \`json:"drift_reason,omitempty"\`` computed when git available — optional, cheap single-file check).

**Engine:** `func Detect(root string, rep canaryscan.Report, staleDays int, refTime time.Time) ([]Finding, error)` where `Finding{ReqID, File, Kind, Detail string}` and Kind ∈ `code-drift` (file's last git commit date > token UPDATED; batch via one `git log` per file: `git -C root log -1 --format=%cs -- <file>`, skipping gracefully when not a git repo or file untracked), `stale` (existing Stale() semantics), `doc-drift` (DB doc_status == DOC_STALE/DOC_MISSING when a DB is present; optional, skip without error when no DB). Honors `CANARY_TEST_TIMESTAMP`.

**CLI:** `canary drift [--root .] [--json] [--stale-days N] [--strict] [--limit 20]` — bounded table grouped by requirement; summary line `CANARY_DRIFT requirements=N code_drift=A stale=B doc_drift=C`; `--strict` exits 2 when any finding. Registered in both command registries.

- [ ] Step 1: failing tests using a real temp git repo (`git init`, commit file, set token UPDATED to yesterday-of-commit vs after): code-drift positive + negative cases, non-git-root soft skip, summary-line format, strict exit code (test the RunE return, not os.Exit). Step 2 FAIL → 3 implement (CBIN-305) → 4 PASS + full suite → 5 commit `feat(drift): git-aware code-vs-token drift detection with staleness and doc rollup`.

---

### Task 8: `canary ticket sync` — codified ticket-source synchronization

**Files:**
- Create: `pkg/ticket/plan.go`, `pkg/ticket/jira.go`, `pkg/ticket/plan_test.go`, `pkg/ticket/jira_test.go` (httptest), `pkg/cmds/ticket/ticket.go`
- Modify: `pkg/config/config.go` (`SourceConfig` gains `API string \`yaml:"api,omitempty"\`` and `StatusMap map[string]string \`yaml:"status_map,omitempty"\``; `pkg/sources` passes them through on `Source`), `cli/cmds.go` + root `cmds.go`.

**Plan computation (pure, fully tested):**
```go
type Action struct {
	Type    string `json:"type"` // create_issue | transition | remap
	ReqID   string `json:"req_id,omitempty"`
	Issue   string `json:"issue,omitempty"`   // e.g. CP-12 (transition target)
	To      string `json:"to,omitempty"`      // JIRA status name
	Summary string `json:"summary,omitempty"` // for create_issue
	Description string `json:"description,omitempty"`
	Source  string `json:"source"`
}
func ComputePlan(tokens []*storage.Token, reg *sources.Registry, remoteStatus map[string]string) ([]Action, error)
```
Rules: requirement whose prefix resolves to a NON-flatfile source → `transition` action when mapped canary rollup status (worst of its tokens: STUB<IMPL<TESTED<BENCHED, mapped through the source's StatusMap defaulting `STUB:"To Do" IMPL:"In Progress" TESTED:"Done" BENCHED:"Done"`) differs from `remoteStatus[id]` (caller supplies; empty map → all transitions proposed). Requirement on a FLATFILE source when at least one non-flatfile source exists → `create_issue` (summary `"<ReqID>: <primary feature>"`, description = feature/aspect/status/file list, bounded) followed by a paired `remap` action (`ReqID` old, `Issue` placeholder `""` — filled at apply time). Deterministic ordering.

**JIRA REST apply:** `type JiraClient struct{ BaseURL, Email, Token string }` with `CreateIssue(project, issueType, summary, description string) (key string, err error)` (POST `/rest/api/3/issue`, ADF paragraph description) and `TransitionIssue(key, toStatusName string) error` (GET transitions, match by target status name case-insensitively, POST transition). Tested against `httptest.Server` fixtures only — no live calls in tests. `canary ticket sync [--db …] [--plan out.json] [--apply] [--project KEY] [--issue-type Story]`: computes plan (remoteStatus fetched only under `--apply` via search `project=KEY` paged); without creds (`JIRA_BASE_URL/JIRA_EMAIL/JIRA_API_TOKEN`) `--apply` prints the plan path and `CANARY_TICKET_PLAN actions=N applied=0 reason=no_credentials` (exit 0). With creds: applies create_issue actions (recording created keys into the paired remap actions), applies transitions, writes the completed plan (with real keys) to `--plan` for `canary upgrade --map` consumption (a remap map JSON is also written next to it as `<plan>.map.json`). Summary line `CANARY_TICKET_SYNC created=X transitioned=Y remap_pending=Z`.

- [ ] Step 1: failing tests for ComputePlan (flatfile→create+remap pairing; jira-source status mismatch→transition; matching status→no action; StatusMap override honored) and JiraClient (httptest: create returns key; transition resolves id by name; auth header basic base64(email:token)). Step 2 FAIL → 3 implement (CBIN-306; config passthrough tested in pkg/sources tests) → 4 PASS + full suite → 5 commit `feat(ticket): codified ticket-source sync with JIRA REST apply and agent-consumable plans`.

---

### Task 9: Plugin repair wave A — the six agent-breaking defects

**Files:** `pkg/cmds/init/base/**`, `pkg/cmds/specify/specify.go`, `pkg/cmds/plan/plan.go`, `claude-plugin/commands/canary.list.md`.
All findings verified by audit (file:line in parentheses). Strategy for template variables: **replace, don't render** — adopt the `claude-plugin/commands` convention of literal angle-bracket placeholders (`<REQ-ID>`, `<PROJECT_KEY>-105`) in every base asset that is installed verbatim.

- [ ] 1. Purge unrendered `{{.ReqID}}`/`{{.ProjectKey}}`/stray `SECURITY_REVIEW` from `base/AGENT_CONTEXT.md` (:28,:35,:81) and all 12 affected `base/commands/*.md` + their `base/templates/commands/` mirrors (119 occurrences; doc.md 29, implement.md 17, list.md 12, specify.md 11, grep.md 10, show.md 9, update-stale.md 8, scan.md 8, bug.md 6, plan.md 4, status.md 3, files.md 2). `SECURITY_REVIEW` survives ONLY in `templates/spec-template.md` + `templates/plan-template.md` where `specify.go:79` legitimately substitutes it.
- [ ] 2. Fix the placeholder contract: `specify.go:76` / `plan.go:120` replace `"CBIN-XXX"` but the templates now contain `{{.ReqID}}-XXX`. Make both ends agree: change templates to use `CBIN-XXX`-style neutral placeholder `REQID-XXX` AND update both ReplaceAll calls to that literal; add a regression test that a generated spec.md contains the real ID and zero `{{` sequences.
- [ ] 3. `--limit` lies in three list.md copies (claude-plugin/commands/canary.list.md:38,46,78,137 + both base copies :36,44,76,135): default is 20, `0`→default, `-1` unlimited; delete every `--limit 0` "for all results" recommendation.
- [ ] 4. `base/agents/security-reviewer.md:1` — delete the stray ```` ```yaml ```` fence so frontmatter parses.
- [ ] 5. Rewrite `base/system/agent.md` from scratch (~60 lines): a real canary-agent system prompt — token format, scan/verify/view/drift command contract, small-context rules — with NO `canary log` (command doesn't exist) and no authoring scaffolding sections.
- [ ] 6. Reconcile `base/commands/bug.md` ↔ `base/templates/commands/bug.md` (the only drifted pair): canonical `bug_id_regex: '^BUG(-[A-Za-z]+)?-[0-9]{3,}$'` accepting both `BUG-001` and `BUG-API-001`; make the two files byte-identical again.
- [ ] Verify: `grep -rn '{{\.' pkg/cmds/init/base/ | grep -v 'spec-template\|plan-template\|implement-prompt-template\|next-prompt-template\|project.yaml\|constitution'` returns empty; `go test ./pkg/cmds/specify/ ./pkg/cmds/plan/ ./... -count=1` green; commit `fix(init): repair agent-breaking template corruption in shipped assets`.

---

### Task 10: Plugin repair wave B — misleads + cosmetics + new command docs

**Files:** `claude-plugin/**`, `cursor-plugin/**`, `pkg/cmds/init/base/copilot/*`, `.github/instructions/*`, `pkg/prompts/**`, `mcp/mcp.go`, `mcp/tools.go`, `docs/MCP_*.md`, `pkg/cmds/init/base/memory/constitution.md`, `base/templates/implement-prompt-template.md`, `base/agents/{docs-writer,integration-test-writer,bench-test-writer}.md`.

- [ ] 1. `claude-plugin/CLAUDE.md`: tool count 18→20 (:74) with `view`/`deps` listed FIRST and the 5 stubs explicitly labeled "(stub — not yet implemented)"; CLI Quick Reference (:54-70) gains `view` (lead position), `deps graph --format mermaid`, `drift`, `upgrade`, `onboard`, `ticket sync`, `--json` on status/files, limit semantics (default 20 / -1 unlimited), ticket-sources one-liner.
- [ ] 2. New command docs (claude-plugin/commands/): `canary.view.md`, `canary.deps.md`, `canary.search.md`, `canary.gap.md`, `canary.index.md`, `canary.onboard.md`, `canary.upgrade.md`, `canary.drift.md`, `canary.ticket.md` — follow the modern frontmatter style of existing files, SHORT (≤60 lines each), accurate flags only (verify each against the cobra definitions before writing). Mirror all nine into `base/commands/` and `base/templates/commands/` (byte-identical trios).
- [ ] 3. Skills: `canary-examples/SKILL.md` gains view-first section, mermaid deps example, `--json`, sources example, drift/upgrade/onboard recipes; `canary-workflow/SKILL.md` inserts view into the lifecycle, fixes `/canary.doc status --all` malformed slash (:~140), adds drift step; `canary-token-format/SKILL.md` required-fields table (:23) qualifies zero-padding as flatfile-only and adds a `CANARY:MIGRATE` subsection.
- [ ] 4. `pkg/prompts`: rewrite `commands/deps/deps.md` to the REAL surface (check/graph/reverse/validate + --format mermaid; the documented show/list/add are fictional); `commands/mcp/mcp.md` 20 tools incl view/deps, stubs labeled, banner block refreshed; CREATE `commands/view/view.md`; fix 8 stale `internal/…` paths (README.md:295; scan.md:90-92; list.md:36; show.md:35; db.md:56; index.md:43; migrate.md:52) → `pkg/…`; `commands/migrate/migrate.md` rewritten to document `orphan detect|run`, `upgrade`, `onboard`, `migrate-from` truthfully; `sys/requirements.md:10` zero-padding qualifier + drop invalid `MISSING` status; add a header comment to all four `sys/*.md` marking `{{VAR}}` as intentional template placeholders for downstream renderers; fix mojibake `?`-glyph lines in prompts (README.md:10-70 diagram — rebuild with ASCII `|-- ` tree; mcp.md:61-64; deps.md:26-33; specs.md:28,38; implement.md:19; migrate.md:38,39).
- [ ] 5. `mcp/mcp.go`: banner `Available MCP Tools (20 total)` (:216) with view/deps lines added and the 18 corrupted `?` glyph bullets replaced with plain `-`; `mcp/tools.go:74-79` align `list` to `capLimit` (20/100) + `Total` on ListResult (small code change + test mirroring the search cap test).
- [ ] 6. MCP docs truth pass: `docs/MCP_TOOLS_COMPLETE.md` (18→20 ×2 :5,:706; deps "Not implemented"→implemented :566; add view section; scan "placeholder"→real :205; bug-create "✅ Full"→stub returning BUG-001 :561; internal/→pkg/ :332,:517), `docs/MCP_QUICK_START.md` (7→20 tool summary incl view/deps :28-38; scan placeholder note removed :121), `docs/MCP_INTEGRATION.md` (:201 scan, :291 pkg/storage, :314 deps roadmap item checked off).
- [ ] 7. Agents: strip Geode references from `base/agents/docs-writer.md:5`, `integration-test-writer.md` (Geode/GQL/RLS/MVCC examples → generic project examples), `bench-test-writer.md`; `integration-test-writer.md:2` frontmatter uses `{{ .AgentPrefix }}`/`{{ .AgentModel }}`/`{{ .AgentColor }}` like its siblings.
- [ ] 8. `base/memory/constitution.md:19` `{{.ProjectKey}}` → `<PROJECT_KEY>`; `base/templates/implement-prompt-template.md:~60` `SECURITY_REVIEW` → `{{.Aspect}}` (this file IS rendered; confirm Aspect is in the render context of next.go:442/implement.go:257 — if absent, add it).
- [ ] 9. Copilot + instructions: `base/copilot/canary.md` + `repository.md` command enumerations gain view/deps/drift/upgrade/onboard + sources note; regenerate `.github/instructions/` via the render path (or hand-sync if not regenerable in-repo).
- [ ] 10. Cursor: `cursor-plugin/.cursor-plugin/plugin.json` version → current release; `cursor-plugin/README.md:3` broken URL → `https://github.com/devnw/canary`; `rules/canary-requirements.mdc` brought to parity with claude-plugin CLAUDE.md content (view-first, limits, sources, drift).
- [ ] Verify: `go build ./... && go test ./... -count=1` green; scan gate green; `grep -rn "18 tools\|18 comprehensive" claude-plugin docs pkg/prompts mcp` empty; commit `docs(plugins): truth pass across claude/cursor/copilot assets, prompts, and MCP docs for v0.3`.

---

### Task 11: Dogfood — upgrade canary's own tokens + backfill missing canaries

Controller-coordinated (implementer subagent does the mechanical parts):
- [ ] `./bin/canary upgrade --root . --write` (rules only; no map yet); review diff; expect md-heading/unicode/padding fixes only.
- [ ] Backfill: run `./bin/canary onboard --root .` + `canary scan`; identify major subsystems lacking tokens (known gaps: `gate/` package, `cli/`, `pkg/upgrade`, `pkg/drift`, `pkg/ticket`, `pkg/onboard` are covered by Tasks 1-8's tokens; audit `pkg/storage`, `pkg/specs`, `pkg/reqid`, `mcp/` for exported features without tokens). Add CBIN-3xx tokens (STATUS honestly reflecting TEST= evidence) for uncovered load-bearing features — target: every top-level package has ≥1 token. Re-scan; update GAP_ANALYSIS claims for anything newly TESTED.
- [ ] `canary index && canary view CBIN-301` smoke; `canary drift --root .` runs clean or explains itself; commit `chore(dogfood): upgrade own tokens and backfill missing canaries`.

---

### Task 12: Dogfood — JIRA (CP) as source of truth

Controller-coordinated with MCP (site devnw.atlassian.net, project CP):
- [ ] Add to `.canary/project.yaml` AND commit-able docs the jira source: `{name: codepros, type: jira, key: "CP", url: "https://devnw.atlassian.net/browse/{id}", api: "https://devnw.atlassian.net"}` (keep the CBIN flatfile source during transition; remove after remap verifies). Update `requirements.id_pattern` to accept both during migration, then `CP-\d+` + legacy test fixtures after.
- [ ] Generate the sync plan: `./bin/canary ticket sync --plan sync-plan.json --project CP --issue-type Story` (no creds in env → plan-only mode, exactly the codified path).
- [ ] Apply via Atlassian MCP: first create ~6 Epics grouping the requirements by aspect/subsystem (Scanner Engine; CLI & Commands; Storage & Index; MCP Server; Ticket Sources & Integrations; Docs & Plugins), then one Story per requirement from the plan's create_issue actions (summary/description from the plan; `additional_fields` parent=epic), collecting `CBIN-xxx → CP-nnn` into `sync-map.json`.
- [ ] Remap the tree: `./bin/canary upgrade --root . --write --map sync-map.json` (rewrites REQ=/BUG= values and GAP_ANALYSIS ✅ claims; TEST= function names intentionally keep their CBIN spellings). Update `.canary/project.yaml` id_pattern; re-run `canary index`, `canary scan --verify GAP_ANALYSIS.md` → must print CANARY_VERIFY_OK; `canary view CP-<n>` shows `Ticket: https://devnw.atlassian.net/browse/CP-<n>`.
- [ ] Status sync-back: for each migrated requirement compute rollup status and transition the JIRA story via MCP (`jira_get_transitions` + `jira_transition_issue`) to the mapped status (TESTED/BENCHED→Done, IMPL→In Progress, STUB→To Do) — the exact action list comes from re-running `ticket sync --plan` post-remap (now emits `transition` actions against CP keys).
- [ ] Full gates: `go test ./... -count=1`, build, scan, verify; commit `feat(dogfood): canary requirements migrated to JIRA CP as source of truth`.

---

### Task 13: Final review, ship

- [ ] Final whole-branch review (superpowers:requesting-code-review, most capable model) over the branch; fix Critical/Important in one wave.
- [ ] Version bump (minor — new commands): plugin manifests + tag `v0.3.0`; push `main` + tag to `origin` AND `gh` (`GIT_SSH_COMMAND="ssh -o ControlMaster=no -o ControlPath=none"`; `--no-verify` for gh); monitor both pipelines green per the established heal loop; confirm `proxy.golang.org` serves `devnw.dev/canary@v0.3.0`.

## Self-Review

1. **Coverage:** MIGRATE comments (T1-2 + upgrade/onboard consumption T4-5 + docs T10); migration review incl. fresh onboarding (T5) and legacy upgrades (T4, plus the bug-create emitter fix T3 and stale help-text noted in T10.4's migrate.md rewrite); drift + dates (T6-7, plugins T10); plugin 47-finding repair (T9-10, MCP list alignment folded into T10.5); dogfood schema/token migration (T11); JIRA source of truth with epics/stories + status sync + codified tooling (T8 + T12).
2. **Ordering:** discriminator (T1) precedes everything MIGRATE; upgrade (T4) precedes dogfood (T11-12); ticket sync (T8) precedes JIRA dogfood; plugin waves precede release.
3. **Contract consistency:** MigrateNote/Report fields (T1) consumed by T2/T5; upgrade.Options.Map (T4) consumed by T12 via sync-map.json; Action schema (T8) consumed by T12's MCP application; refs kind='migrate' (T2) matches T1's storage note.

## Known adaptation points (implementers: verify, don't assume)

Real walk/ignore helper shapes in `pkg/canaryscan` (T1); current view refs-filtering behavior (T2 — fix the kind cross-contamination with a test); exact multi-line emission shape in bug create (T3-4); `reqid.GenerateNextID` behavior on empty specs dir (T5); how tools/canary/main.go surfaces flags (T6); storage.Token doc_status values (T7); cobra registration patterns in both registries (T4,5,7,8).
