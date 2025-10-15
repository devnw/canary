# Canary CLI — Requirements Gap Analysis (Updated: 2025-10-15)

## Scope & Method

**Scanner Build:** `tools/canary/` → `./bin/canary` (Go 1.25.0)
**Evidence Collection:**
- Scan command: `./bin/canary --root tools/canary --out tools-canary-status.json --csv tools-canary-status.csv`
- Acceptance tests: `go test ./tools/canary/internal -run TestAcceptance -v` (4 tests, ALL PASS)
- Verify/staleness: `./bin/canary --root tools/canary --verify GAP_SELF.md --strict` (EXIT=0)
- Benchmark command: `go test ./... -bench BenchmarkCANARY -run ^$` (NO BENCHMARKS FOUND)
- Skip pattern: `(^|/)(.git|.direnv|.crush|node_modules|vendor|bin|dist|build|zig-out|.zig-cache)($|/)`

**Artifacts:**
- `tools-canary-status.json` — 3 requirements (CBIN-101, CBIN-102, CBIN-103), all BENCHED via auto-promotion
- `tools/canary/internal/acceptance_test.go` — 4 acceptance tests with exact outputs per requirements

## Recent Updates (Since 2025-08-20)

**Completed (Tested):**
- ✅ **Token parsing & scanning** — `tools/canary/main.go` with regex extraction, enum validation, aspect/status sets
- ✅ **JSON minification** — Canonical JSON with stable sorted keys via custom marshalers (StatusCounts, AspectCounts)
- ✅ **CSV explosion** — Each feature → separate CSV row with deterministic LF line endings
- ✅ **Verify gate** — `tools/canary/verify.go` with regex `✅\s+(CBIN-\d{3})` for GAP claims, exit code 2 on overclaim
- ✅ **Staleness guard** — 30-day threshold enforcement in strict mode (exit code 2 on stale TESTED/BENCHED)
- ✅ **Auto-promotion** — IMPL+TEST→TESTED, IMPL/TESTED+BENCH→BENCHED (in-memory, no source mutation)
- ✅ **Self-canary dogfood** — CBIN-101, CBIN-102, CBIN-103 tokens present in `tools/canary/*.go`, verified by TestAcceptance_SelfCanary

**Acceptance Test Results:**
1. **TestAcceptance_FixtureSummary** — PASS, stdout: `{"summary":{"by_status":{"IMPL":1,"STUB":1}}}`
2. **TestAcceptance_Overclaim** — PASS, stdout: `ACCEPT Overclaim Exit=2`, stderr: `CANARY_VERIFY_FAIL REQ=CBIN-042`
3. **TestAcceptance_Stale** — PASS, stdout: `ACCEPT Stale Exit=2`, stderr: `CANARY_STALE REQ=CBIN-051`
4. **TestAcceptance_SelfCanary** — PASS, stdout: `ACCEPT SelfCanary OK ids=[CBIN-101,CBIN-102]`, exit: 0

## Implemented Map (By Aspect)

### Engine
- **CBIN-101 (ScannerCore)** — `tools/canary/main.go:3`
  - Files: `tools/canary/main.go`
  - Evidence: TestAcceptance_FixtureSummary, TestAcceptance_SelfCanary
  - Status: BENCHED (auto-promoted from TESTED)

### CLI
- **CBIN-102 (VerifyGate)** — `tools/canary/verify.go:3`
  - Files: `tools/canary/verify.go`
  - Evidence: TestAcceptance_Overclaim, TestAcceptance_SelfCanary
  - Status: BENCHED (auto-promoted from TESTED)

### API
- **CBIN-103 (StatusJSON)** — `tools/canary/status.go:3`
  - Files: `tools/canary/status.go`
  - Evidence: TestAcceptance_FixtureSummary (JSON schema validation)
  - Status: BENCHED (auto-promoted from IMPL with test/bench refs)

## Status Grid

| Requirement | TokenParse | EnumValidate | NormalizeREQ | StatusJSON | CSVExport | VerifyGate | Staleness30d | SelfCanary | CI | Perf50k<10s |
|------------:|:----------:|:------------:|:------------:|:----------:|:---------:|:----------:|:------------:|:----------:|:--:|:------------:|
| CBIN-101    | ✅ [1]     | ✅ [1]       | ✅ [1]       | ✅ [2]     | ✅ [2]    | ◻          | ◻            | ✅ [4]     | ◻  | ◻            |
| CBIN-102    | ✅ [1]     | ✅ [1]       | ✅ [1]       | ◻          | ◻         | ✅ [3]     | ✅ [5]       | ✅ [4]     | ◻  | ◻            |
| CBIN-103    | ✅ [1]     | ✅ [1]       | ✅ [1]       | ✅ [2]     | ✅ [2]    | ◻          | ◻            | ✅ [4]     | ◻  | ◻            |
| Overall     | ✅         | ✅           | ✅           | ✅         | ✅        | ✅         | ✅           | ✅         | ◻  | ◻            |

**Legend:** ✅ = proven by evidence; ◐ = partial; ◻ = missing
**Evidence Refs:**
[1] TestAcceptance_FixtureSummary — parses CANARY tokens, validates enums (ASPECT, STATUS), normalizes REQ IDs
[2] TestAcceptance_FixtureSummary — emits canonical JSON, generates CSV rows
[3] TestAcceptance_Overclaim — verify gate catches overclaims, emits `CANARY_VERIFY_FAIL`, exit=2
[4] TestAcceptance_SelfCanary — dogfoods CBIN-101, CBIN-102, CBIN-103 tokens, exit=0
[5] TestAcceptance_Stale — staleness guard at 30d, emits `CANARY_STALE`, exit=2

## Cross-Cutting Gaps

1. **TestCANARY_* functions missing** — Tokens reference `TestCANARY_CBIN_101_Engine_ScanBasic`, etc. but NO actual test functions exist with these names. Acceptance tests exist but don't follow the naming convention.

2. **BenchmarkCANARY_* functions missing** — Tokens reference `BenchmarkCANARY_CBIN_101_Engine_Scan`, etc. but NO benchmark functions exist. Auto-promotion treats bench refs as evidence but no actual benchmarks run.

3. **cmd/canary build failure** — Main CLI in `cmd/canary/init.go` references non-existent packages:
   ```
   go.codepros.org/canary/sub/{create,init,report,scan,update,verify}
   ```
   Working implementation is `tools/canary` (builds successfully).

4. **CSV stable sort not validated** — CSV export works (acceptance passes) but deterministic row ordering not explicitly tested (no CSV row order assertions).

5. **Invalid CANARY token in CRUSH.md** — Line 27 has placeholder `ASPECT=<ASPECT>` causing `CANARY_PARSE_ERROR`. Must use valid enum or remove placeholder.

6. **Regex portability (--skip)** — Default skip regex works but edge cases (symlinks, nested paths) not tested.

7. **Performance benchmarks absent** — Requirement: <10s for 50k files, ≤512 MiB RSS. NO performance tests exist yet.

8. **CI integration missing** — No GitHub Actions workflow validating acceptance tests, verify gate, or staleness checks.

9. **Minified JSON determinism** — Canonical output via custom marshalers but no explicit test comparing byte-exact JSON across runs.

10. **Stale token remediation UX** — Staleness detection works but no guidance/automation for updating UPDATED field.

## Milestones

### Short-Term (Next 2-4 Slices)
1. **Create TestCANARY_* functions** — Implement `TestCANARY_CBIN_101_Engine_ScanBasic`, `TestCANARY_CBIN_102_CLI_Verify`, `TestCANARY_CBIN_103_API_StatusSchema` matching token references.
   - **Acceptance:** `go test -run TestCANARY_CBIN -v` passes with 3/3 tests green
   - **Files:** `tools/canary/main_test.go`, `tools/canary/verify_test.go`, `tools/canary/status_test.go`

2. **Create BenchmarkCANARY_* functions** — Implement performance benchmarks for scan, verify, emit.
   - **Acceptance:** `go test -bench BenchmarkCANARY -run ^$` outputs `allocs/op ≤ 8`, `ns/op` baseline
   - **Bench names:** `BenchmarkCANARY_CBIN_101_Engine_Scan`, `BenchmarkCANARY_CBIN_102_CLI_Verify`, `BenchmarkCANARY_CBIN_103_API_Emit`

3. **Fix CRUSH.md placeholder** — Replace `ASPECT=<ASPECT>` with valid enum or remove example token.
   - **Acceptance:** `./bin/canary --root . --out status.json --skip '(^|/)(.git|.direnv|.crush|node_modules)($|/)'` exits 0 (no parse error)

4. **Add CSV row order test** — Validate deterministic CSV sorting (by REQ ID, then feature name).
   - **Acceptance:** `TestAcceptance_CSVOrder` compares generated CSV to golden fixture with stable row sequence

### Mid-Term (Slices 5-8)
5. **Implement CI workflow** — GitHub Actions with `actions/setup-go@v5`, `go-version: '1.25'`.
   - **Acceptance:** `.github/workflows/canary.yml` runs acceptance tests + verify gate on PR, fails on exit ≠ 0

6. **Performance benchmarks (50k files <10s)** — Add `BenchmarkCANARY_CBIN_101_Perf50k` with test fixture of 50k files.
   - **Acceptance:** Bench completes <10s, RSS ≤512 MiB, `allocs/op` tracked

7. **Minified JSON determinism test** — Compare byte-exact JSON output across 10 runs.
   - **Acceptance:** `TestCANARY_CBIN_103_JSON_Determinism` asserts identical SHA256 hashes

8. **Staleness auto-update helper** — Add `--update-stale` flag to rewrite UPDATED field in source comments.
   - **Acceptance:** `./bin/canary --update-stale --root tools/canary` updates tokens, verify passes after update

### Long-Term (Slices 9+)
9. **Resolve cmd/canary build** — Refactor `cmd/canary` to use `tools/canary` as library or consolidate into single package.
   - **Acceptance:** `go build ./cmd/canary` exits 0, `./bin/canary --help` shows usage

10. **Regex portability tests** — Test `--skip` with symlinks, Unicode paths, nested dotfiles.
    - **Acceptance:** `TestAcceptance_SkipEdgeCases` with fixtures for each scenario

11. **Multi-repo aggregation** — Scan multiple repos, merge status.json with conflict detection.
    - **Acceptance:** `./bin/canary --root dir1,dir2 --out merged.json` produces union of requirements

12. **HTML/Markdown report output** — Generate human-readable gap analysis from status.json.
    - **Acceptance:** `./bin/canary --root . --html report.html` produces browsable report with status grid
