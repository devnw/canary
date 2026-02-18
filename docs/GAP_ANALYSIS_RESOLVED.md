# Resolved Gaps

**Source of truth:** GitLab issues (closed with evidence).  
When a gap is closed with verification, it is moved here from [GAP_ANALYSIS.md](./GAP_ANALYSIS.md) and the corresponding GitLab issue is closed.

---

## Resolved items

### GAP-0001: Scan subcommand shells out to go run (Closed — Implemented)

- **GitLab issue:** [#8](https://gitlab.com/devnw/codepros/oss/canary/-/issues/8)
- **Resolution:** Scanner logic moved into `internal/canaryscan`; `internal/cmds/scan/scan.go` now calls `canaryscan.Run(cfg, ...)` with no exec. `canary scan --root <dir>` works from any CWD (e.g. from `/tmp` with `--root` pointing at repo). `tools/canary/main.go` is a thin wrapper that calls `canaryscan.RunFromArgs`.
- **Key files:** `internal/canaryscan/` (config, types, parse, scan, verify, update, write, run, env), `internal/cmds/scan/scan.go`, `tools/canary/main.go`
- **Closed:** 2026-02-18

### GAP-0002: Build/env docs out of sync (Closed — Implemented)

- **GitLab issue:** [#9](https://gitlab.com/devnw/codepros/oss/canary/-/issues/9)
- **Resolution:** CONTRIBUTING.md updated to remove nix/direnv requirement and point to DEVELOPMENT.md. DEVELOPMENT.md populated with Go build, test, lint, and self-canary instructions.
- **Key files:** `docs/CONTRIBUTING.md`, `docs/DEVELOPMENT.md`
- **Closed:** 2026-02-18

### GAP-0010: Acceptance tests SelfCanary and UpdateStale failing (Closed — Implemented)

- **GitLab issue:** [#17](https://gitlab.com/devnw/codepros/oss/canary/-/issues/17)
- **Resolution:** Staleness and update-stale now respect `CANARY_TEST_TIMESTAMP` env var. `internal/canaryscan`: `Stale(..., refTime)`, `RefTimeFromEnv()`, and update-stale use ref time so tests with fixed timestamp pass. All `go test ./tools/canary/...` pass.
- **Key files:** `internal/canaryscan/verify.go`, `internal/canaryscan/update.go`, `internal/canaryscan/run.go`, `internal/canaryscan/env.go`
- **Closed:** 2026-02-18

### GAP-0011: No GitLab CI (Closed — Won't do)

- **GitLab issue:** [#18](https://gitlab.com/devnw/codepros/oss/canary/-/issues/18)
- **Resolution:** **Won't do** for this repo. CI/CD, Docker tag pinning, and pre-commit hooks are handled via different controls. Documented for future agents in `docs/AGENT_GAP_POLICY.md`.
- **Key files:** `docs/AGENT_GAP_POLICY.md`
- **Closed:** 2026-02-18

### GAP-0012: Dual scanner implementations (Closed — Implemented)

- **GitLab issue:** [#19](https://gitlab.com/devnw/codepros/oss/canary/-/issues/19)
- **Resolution:** Single scanner path: all scan behavior lives in `internal/canaryscan`. `cmd/canary` scan subcommand and `tools/canary` binary both use it. Root `main.go` remains `//go:build ignore` (legacy, not built by default).
- **Key files:** Same as GAP-0001
- **Closed:** 2026-02-18

### GAP-0003: --prompt flag stubbed across CLI commands (Closed — Implemented)

- **GitLab issue:** [#10](https://gitlab.com/devnw/codepros/oss/canary/-/issues/10)
- **Resolution:** All commands with `--prompt` or `--prompt-arg` now call `utils.LoadPrompt(prompt)` when the flag is set; invalid or missing prompt returns an error.
- **Key files:** `internal/cmds/internal/utils/prompt.go`, command files under `internal/cmds/`
- **Closed:** 2026-02-18

### GAP-0004: Embedded prompt loading and listing not implemented (Closed — Implemented)

- **GitLab issue:** [#11](https://gitlab.com/devnw/codepros/oss/canary/-/issues/11)
- **Resolution:** `loadEmbeddedPrompt` uses `internal/prompts` (command + system prompts); `GetAvailablePrompts()` returns deduped, sorted list. Tests in `internal/cmds/internal/utils/prompt_test.go`.
- **Key files:** `internal/cmds/internal/utils/prompt.go`, `prompt_test.go`, `internal/prompts/`
- **Closed:** 2026-02-18

### GAP-0005: MCP tools stubs (Closed — Implemented)

- **GitLab issue:** [#12](https://gitlab.com/devnw/codepros/oss/canary/-/issues/12)
- **Resolution:** Scan tool wired to `canaryscan.Scan()` in `mcp/tools.go`; returns real token counts. specify, plan, index, BUG ID, gap mark documented as stubs in `mcp/tools_extended.go`.
- **Key files:** `mcp/tools.go`, `mcp/tools_extended.go`
- **Closed:** 2026-02-18

### GAP-0006: BUG filesystem-based search not implemented (Closed — Implemented)

- **GitLab issue:** [#13](https://gitlab.com/devnw/codepros/oss/canary/-/issues/13)
- **Resolution:** `listBugsFromFilesystem` uses `canaryscan.Scan()`, filters BUG IDs, converts to tokens, applies filters and formatBugList. Parser accepts `BUG=` as REQ; BUG statuses added to statusSet.
- **Key files:** `internal/cmds/bug/bug.go`, `internal/canaryscan/parse.go`
- **Closed:** 2026-02-18

### GAP-0007: Doc command placeholder sections (Closed — Implemented)

- **GitLab issue:** [#14](https://gitlab.com/devnw/codepros/oss/canary/-/issues/14)
- **Resolution:** Fallback doc template uses clearly marked stubs ("_(Auto-generated stub. Replace with ...)_") instead of "TODO:".
- **Key files:** `internal/cmds/doc/doc_commands.go`
- **Closed:** 2026-02-18

### GAP-0008: Implement listing and next JSON output (Closed — Implemented)

- **GitLab issue:** [#15](https://gitlab.com/devnw/codepros/oss/canary/-/issues/15)
- **Resolution:** `listUnimplemented()` lists `.canary/specs` dirs; `next --json` outputs `nextJSONOutput` as JSON.
- **Key files:** `internal/cmds/implement/implement.go`, `internal/cmds/next/next.go`
- **Closed:** 2026-02-18

### GAP-0009: Orphan test skipped (Closed — Implemented)

- **GitLab issue:** [#16](https://gitlab.com/devnw/codepros/oss/canary/-/issues/16)
- **Resolution:** Skip removed. Test fixed by setting RawToken/IndexedAt on token, using non-test path, and ListTokens with `include_hidden: "true"`.
- **Key files:** `internal/migrate/orphan_test.go`
- **Closed:** 2026-02-18
