# Gap Analysis

**Source of truth for requirements and gaps:** [GitLab issues](https://gitlab.com/devnw/codepros/oss/canary/-/issues)  
**Last updated:** 2026-02-18  
**Branch:** refactor

**Resolved/closed items** are in [GAP_ANALYSIS_RESOLVED.md](./GAP_ANALYSIS_RESOLVED.md) (GAP-0001, 0002, 0010, 0011, 0012). **Closed this cycle:** GAP-0003, 0004, 0005, 0006, 0007, 0008, 0009 (see resolution notes below and in RESOLVED). **Won't do** (CI/CD, Docker, pre-commit): see [AGENT_GAP_POLICY.md](./AGENT_GAP_POLICY.md).

---

## Gaps (OPEN)

### GAP-0014: Project structure does not follow Go best practices for importability

**Status:** **Blocked** (deferred; structural refactor)

- **GitLab issue:** _(to be created)_
- **Category:** Code Quality
- **Severity:** High
- **Requirement source:** User / project structure; [Go internal packages](https://go.dev/doc/go1.4#internalpackages)
- **Current state (evidence):**
  - All library-style code lives under `internal/`: `internal/canaryscan`, `internal/storage`, `internal/specs`, `internal/gap`, `internal/migrate`, `internal/reqid`, etc. The `gate` package is at repo root but scanner entry used by CLI is `internal/canaryscan`.
  - Go `internal` packages are **not importable by other modules**. External tools cannot `import "go.devnw.com/canary/internal/canaryscan"` (forbidden by the compiler).
  - No public, versioned API package exists for scanner or other functionality that other tools might want to reuse.
- **Why this is a gap:** Other tools (e.g. CI plugins, IDE integrations, alternative CLIs) cannot import canary’s scanner or helpers without forking or copying code. This conflicts with Go best practice of exposing a small set of importable packages for reuse.
- **Impact/Risk:** No ecosystem reuse; duplication; harder to maintain alternative front-ends (MCP, other CLIs).
- **Recommended fix:**
  1. Expose one or more **public API packages** at module root (not under `internal/`). For example:
     - **Scanner:** Move `internal/canaryscan` to a top-level package `scan/` (so `import "go.devnw.com/canary/scan"`) with a stable public API: `scan.Config`, `scan.Run`, `scan.Scan`, `scan.Report`, `scan.DefaultSkipRegex`, etc. Keep package minimal and document it as the supported entry point for scanning.
     - Optionally expose a thin `storage` or `canary` package if other tools need to query the token DB; otherwise keep storage in `internal`.
  2. Document the public API in the repo and in godoc. Reserve `internal/` for implementation details that are not part of the supported contract.
- **Acceptance criteria:**
  - At least one package (e.g. `go.devnw.com/canary/scan`) is importable by external modules and provides the scanner API.
  - README or docs describe which packages are public and how to import them.
  - Existing `cmd/canary` and `tools/canary` use the public package(s) where appropriate.
- **Unblock steps:** Create a top-level `scan/` (or similar) package with a minimal public API; move `internal/canaryscan` logic behind it or re-export; update `cmd/canary` and `tools/canary` to use the new package; document in README.

---

### GAP-0015: Sub-commands are implemented in internal/cmds/\* instead of cli/ package

**Status:** **Blocked** (deferred; structural refactor)

- **GitLab issue:** _(to be created)_
- **Category:** Code Quality
- **Severity:** Medium
- **Requirement source:** User; standard layout for CLI-heavy projects
- **Current state (evidence):**
  - `cli/cmds.go` only **aggregates** commands: it imports `internal/cmds/scan`, `internal/cmds/show`, etc. and returns `scan.ScanCmd`, `show.ShowCmd`, etc.
  - All command **definitions** (cobra.Command, RunE, flags, init) live under `internal/cmds/*`: e.g. `internal/cmds/scan/scan.go`, `internal/cmds/show/show.go`, `internal/cmds/next/next.go`. The `cli/` package does not contain any sub-command implementation.
  - `cmd/canary/main.go` imports `cli` and registers `cli.Commands()`; flag wiring for init, db, next, etc. is done in main.go against commands that live in internal/cmds.
- **Why this is a gap:** The CLI layer is split: the “CLI” package only aggregates, while the actual sub-commands live in internal. Best practice for a CLI-focused repo is to implement sub-commands in the `cli/` package so that the CLI is the single place that owns command structure, flags, and wiring; it then calls into internal or public packages for business logic.
- **Impact/Risk:** Harder to see the full CLI surface in one place; contributors look in internal/cmds for behavior; main.go has to know about internal/cmds for flag wiring.
- **Recommended fix:**
  1. **Move sub-command implementation into `cli/`.** For each command (scan, show, list, init, next, …), implement the cobra command in the `cli` package, e.g. `cli/scan.go`, `cli/show.go`, `cli/next.go`, or `cli/commands/scan.go`, etc. Each file defines the `*cobra.Command` and its `RunE` (and flags) in the same package.
  2. Keep **business logic** in `internal/` (or new public packages): storage, scanner, specs, migrate, etc. The cli layer only parses flags, calls into those packages, and formats output.
  3. **Remove or deprecate `internal/cmds/*`** as the home for command definitions: either delete after moving logic into cli, or reduce internal/cmds to pure library helpers called by cli, with no cobra types.
  4. **Centralize flag registration** in cli (or in the command’s init in cli) so `cmd/canary/main.go` stays minimal (root command, global flags, add `cli.Commands()`).
- **Acceptance criteria:**
  - All top-level sub-commands (scan, show, list, init, next, etc.) are **defined and implemented** in the `cli/` package (same module path as current `cli`).
  - `cli.Commands()` returns commands that are defined in `cli/`, not in `internal/cmds`.
  - No dependency from `cmd/canary` to `internal/cmds` for command or flag wiring; only `cli` (and optionally public packages) are used.
  - Existing behavior and tests preserved (or updated to match).
- **Unblock steps:** Move each sub-command from `internal/cmds/*` into `cli/` (e.g. `cli/scan.go`, `cli/next.go`); keep business logic in `internal/`; remove or reduce `internal/cmds` to helpers only; centralize flag registration in `cli`.

---

### GAP-0003: --prompt flag stubbed across CLI commands

**Status:** **Closed — Implemented**

- **GitLab issue:** [#10](https://gitlab.com/devnw/codepros/oss/canary/-/issues/10)
- **Category:** Feature
- **Severity:** Medium
- **Requirement source:** docs/GAP_ANALYSIS.md; README/docs (agent/custom prompt workflows)
- **Current state (evidence):**
  - internal/cmds/show/show.go:41 — `// TODO: Implement --prompt flag to load custom prompts`
  - internal/cmds/grep/grep.go:38, files/files.go:31, status/status.go:34, specify/specify.go:25, search/search.go:22, prioritize/prioritize.go:20, plan/plan.go:27, next/next.go:49, list/list.go:31, index/index.go:25, implement/implement.go:47, create/create.go:26, checkpoint/checkpoint.go:28, bug/list.go:30, bug/create.go:32, bug/update.go:29, bug/show.go:28 — same TODO pattern
- **Why this is a gap:** Flag exists but behavior not implemented; prompt not loaded or used.
- **Impact/Risk:** Users and agents cannot use custom prompts; flag is dead.
- **Recommended fix:** Implement prompt loading (file + embedded name) and wire to commands, or remove flag and document.
- **Acceptance criteria:** `--prompt` either loads file/embedded prompt or is removed with doc note.
- **Resolution:** All commands that expose `--prompt` or `--prompt-arg` now call `utils.LoadPrompt(prompt)` when the flag is set; invalid or missing prompt returns an error. Key files: `internal/cmds/internal/utils/prompt.go`, and command files under `internal/cmds/` (show, files, grep, status, specify, search, prioritize, plan, list, index, create, checkpoint, bug/\*, next, implement).

---

### GAP-0004: Embedded prompt loading and listing not implemented

**Status:** **Closed — Implemented**

- **GitLab issue:** [#11](https://gitlab.com/devnw/codepros/oss/canary/-/issues/11)
- **Category:** Feature
- **Severity:** Medium
- **Requirement source:** docs/GAP_ANALYSIS.md; internal/cmds/internal/utils/prompt.go
- **Current state (evidence):**
  - internal/cmds/internal/utils/prompt.go:54 — `// TODO: Implement embedded prompt loading`
  - internal/cmds/internal/utils/prompt.go:90 — `// TODO: Implement listing of embedded prompts`
- **Why this is a gap:** Embedded prompt names and listing are unimplemented.
- **Impact/Risk:** Embedded prompts cannot be used; no discovery of available prompts.
- **Recommended fix:** Implement loading from embedded FS (prompts/sys, prompts/commands, .canary/templates) and listing of names.
- **Acceptance criteria:** loadEmbeddedPrompt(name) returns content for valid names; GetAvailablePrompts() returns list of embedded prompt names.
- **Resolution:** `loadEmbeddedPrompt` uses `internal/prompts`: command prompts via `prompts.GetCommand(name)`, system prompts via `prompts.All()[name]`. `GetAvailablePrompts()` uses `prompts.ListCommands()` and keys of `prompts.All()`, deduped and sorted. Tests: `internal/cmds/internal/utils/prompt_test.go`.

---

### GAP-0005: MCP tools are stubs (scan, specify, plan, index, gap, BUG ID)

**Status:** **Closed — Implemented**

- **GitLab issue:** [#12](https://gitlab.com/devnw/codepros/oss/canary/-/issues/12)
- **Category:** Feature
- **Severity:** Medium
- **Requirement source:** docs/GAP_ANALYSIS.md; MCP integration (README/docs)
- **Current state (evidence):**
  - mcp/tools.go:408-409 — `// TODO: Implement actual scan functionality`
  - mcp/tools_extended.go:45, 85, 124, 488, 538 — TODOs for spec, plan, index, BUG ID, gap marking
- **Why this is a gap:** MCP tools return placeholders; no real scan/spec/plan/index/gap/ID.
- **Impact/Risk:** MCP consumers get placeholder data.
- **Recommended fix:** Wire MCP tools to existing CLI/library implementations or document as experimental/stub.
- **Acceptance criteria:** MCP scan returns real results when backend available; other tools implemented or explicitly documented as stub.
- **Resolution:** Scan tool wired to `internal/canaryscan.Scan()` in `mcp/tools.go`; returns real token counts and unique requirements. specify, plan, index, BUG create (ID), and gap mark remain stubs; documented in `mcp/tools_extended.go` (file comment) and tracked for future work.

---

### GAP-0006: BUG filesystem-based search not implemented

**Status:** **Closed — Implemented**

- **GitLab issue:** [#13](https://gitlab.com/devnw/codepros/oss/canary/-/issues/13)
- **Category:** Feature
- **Severity:** Low
- **Requirement source:** docs/GAP_ANALYSIS.md; internal/cmds/bug/bug.go:208
- **Current state (evidence):** internal/cmds/bug/bug.go:208 — `// TODO: Implement filesystem-based search for BUG tokens`
- **Why this is a gap:** BUG tokens cannot be discovered via filesystem search.
- **Impact/Risk:** BUG discovery feature incomplete.
- **Recommended fix:** Implement filesystem scan for BUG token pattern (or use existing scanner) and wire to bug commands.
- **Acceptance criteria:** BUG tokens discoverable via filesystem search where documented.
- **Resolution:** `listBugsFromFilesystem` in `internal/cmds/bug/bug.go` now calls `canaryscan.Scan(root, ...)`, filters requirements by BUG ID pattern, converts to `[]*storage.Token`, applies severity/priority filter and formatBugList. Parser accepts `BUG=` as REQ in `internal/canaryscan/parse.go`; BUG statuses (IN_PROGRESS, VERIFIED, etc.) added to statusSet.

---

### GAP-0007: Doc command output has placeholder sections

**Status:** **Closed — Implemented**

- **GitLab issue:** [#14](https://gitlab.com/devnw/codepros/oss/canary/-/issues/14)
- **Category:** Docs
- **Severity:** Low
- **Requirement source:** docs/GAP_ANALYSIS.md; internal/cmds/doc/doc_commands.go:96, 100, 104, 108
- **Why this is a gap:** Generated doc content is placeholder text only.
- **Impact/Risk:** Doc command output not usable as-is.
- **Recommended fix:** Replace with real content from specs/tokens or mark as auto-generated stub.
- **Acceptance criteria:** Doc command output has real overview/usage/examples or clearly marked stub.
- **Resolution:** Fallback template in `internal/cmds/doc/doc_commands.go` (when template file missing) now uses clearly marked stubs: "_(Auto-generated stub. Replace with ...)_" instead of "TODO: ...".

---

### GAP-0008: Implement listing and next JSON output not implemented

**Status:** **Closed — Implemented**

- **GitLab issue:** [#15](https://gitlab.com/devnw/codepros/oss/canary/-/issues/15)
- **Category:** Feature
- **Severity:** Low
- **Requirement source:** internal/cmds/implement/implement.go:364; internal/cmds/next/next.go:98, 581
- **Why this is a gap:** Implement listing and next --json are stubbed or placeholder.
- **Recommended fix:** Implement listing in implement command; add JSON output for next; ensure generated snippet valid or document as template.
- **Acceptance criteria:** implement listing works if documented; next --json returns valid JSON.
- **Resolution:** `listUnimplemented()` in `internal/cmds/implement/implement.go` lists `.canary/specs` dirs and prints `CBIN-XXX\tfeature-name`. `next --json` in `internal/cmds/next/next.go` outputs `nextJSONOutput` (req_id, feature, aspect, status, priority, file_path, updated) as JSON.

---

### GAP-0009: Orphan test skipped due to database insertion issue

**Status:** **Closed — Implemented**

- **GitLab issue:** [#16](https://gitlab.com/devnw/codepros/oss/canary/-/issues/16)
- **Category:** Test
- **Severity:** Medium
- **Requirement source:** internal/migrate/orphan_test.go:149
- **Why this is a gap:** Test not run; underlying insertion issue not resolved.
- **Recommended fix:** Fix database insertion and re-enable test, or document skip and open separate bug.
- **Acceptance criteria:** Test runs in CI with skip removed and passing, or documented deferral.
- **Resolution:** Skip removed. Test fixed by (1) setting `RawToken` and `IndexedAt` on the inserted token (required by schema), (2) using a non-test file path (`pkg/api/handler.go`), (3) calling `ListTokens` with `include_hidden: "true"` so the token is returned. `internal/migrate/orphan_test.go`.

---

### GAP-0013: Label "gap" for gap-analysis issues missing

- **GitLab issue:** [#20](https://gitlab.com/devnw/codepros/oss/canary/-/issues/20)
- **Category:** Other
- **Severity:** Low
- **Status:** **Blocked**
- **Requirement source:** Gap analysis workflow (label taxonomy)
- **Current state (evidence):** Label creation via `glab api POST projects/77200275/labels` failed (response parse error); no `gap` label in project.
- **Why blocked:** Cannot create label programmatically with current glab/API response handling.
- **Unblock steps:** Create the label "gap" in GitLab via **Settings → Labels → New label** (name: `gap`, color e.g. `#d73a4a`). Then apply it to issues #8–#20. Alternatively fix glab or use a different API client that handles the label-create response.
- **Acceptance criteria:** Label "gap" exists and is applied to gap-analysis-created issues.
