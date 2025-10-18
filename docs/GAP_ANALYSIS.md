# Canary CLI — Requirements Gap Analysis (Updated: 2025-09-20)

## Scope & Method

Scanned repository root `.` using ad-hoc legacy scanner (`main.go + scan.go + verify.go`) built with:

```bash
go build -o ./bin_canary ./scan.go ./verify.go ./main.go
./bin_canary --root . --out status.json --csv status.csv
```

Skip set (hard-coded): `.git, node_modules, vendor, bin, zig-out, .zig-cache, .crush, data, certs`.

Evidence artifacts: `./status.json` (pretty JSON) and `./status.csv` (row explosion). No acceptance test harness or benches present. Verification mode not exercised against `CBIN-###` because tokens absent.

## Recent Updates

None in this codebase matching the canonical spec. Implementation predates new `CBIN-###` prefix and 30-day staleness rule. No self-canary tokens (CBIN-101..103). JSON not minified, missing stable zero-count buckets. No GitHub Actions workflow for self-verify.

## Implemented Map (by ASPECT)

### API

- Legacy token parse for `REQ-GQL-###` lines (no normalization to `CBIN-###`).

### CLI

- Flags: `-root`, `-out`, `-csv`, `-verify`, `-strict` (but `-strict` uses 60-day window, spec wants 30-day). Missing `--skip` flag and canonical naming (`--root`, etc. acceptable but binary name mismatch: expected `canaryscan`).

### Engine

- Streaming scan via `filepath.WalkDir`, basic binary file skip heuristic. Lacks RE2 skip regex, lacks performance metrics, no parallelism or benchmarks.

### Docs

- README documents legacy token format (`REQ-GQL-###`). GAP / NEXT / CHECKLIST placeholders replaced in this update.

## Status Grid (Checklist Projection)

See `CHECKLIST.md` for canonical table. Summary: all requirements CBIN-101..103 are currently ◻ (missing) because tokens do not exist yet; partial legacy parsing & CSV gives Overall ◐ in select columns (see evidence notes below).

## Cross-Cutting Gaps

1. Prefix mismatch: `REQ-GQL-###` vs required `CBIN-###`.
2. Missing required self-canary tokens CBIN-101..103 (ScannerCore, VerifyGate, StatusJSON).
3. JSON schema drift: pretty-printed, unordered status buckets (missing zeroed statuses) vs required minified deterministic order.
4. CSV determinism: no explicit stable sort ordering beyond iteration order; needs guaranteed ordering.
5. Verify semantics: regex differs (implementedRe broad), needs strict `^✅ CBIN-###` and fail on lack of TESTED/BENCHED tokens.
6. Staleness window: 60 days; requirement = 30 days.
7. No staleness diagnostics format (`CANARY_STALE REQ=... updated=... age_days=...`). Current code logs aggregated message.
8. No `--skip` regex flag (hard-coded directories reduce flexibility).
9. No acceptance tests / benchmarks; no CI workflow gating.
10. Performance unknown; no evidence meeting 50k files <10s requirement.

## Milestones

### Short (next 1-2 slices)

- Implement new `tools/canary/` scanner with canonical CLI name `./bin/canary` (or `canaryscan` per spec) supporting `--skip` regex, minified JSON, deterministic CSV, strict enums.
- Seed self-canary tokens CBIN-101..103 in `main.go`, `verify.go`, `status.go` (new file) with STATUS=TESTED/IMPL as per spec and acceptance test harness.
- Implement 30-day staleness check with per-token diagnostics `CANARY_STALE ...`.
- Add strict verify regex and exit code 2 on over-claim with diagnostic `CANARY_VERIFY_FAIL REQ=...`.

### Mid (after baseline passes)

- Add acceptance tests (fixture summary, overclaim, stale, self-verify) under `tools/canary/internal/...` harness.
- Add benchmarks `BenchmarkCANARY_CBIN_101_Engine_Scan` etc. capturing ns/op, allocs/op; set initial guardrails.
- Introduce performance measurement harness for 10k mock files scaling.

### Long

- Parallel scanning & streaming writer for large repos; memory profiling to ensure <512 MiB.
- Enhanced output: optional markdown summary, diff vs previous status.json.
- Token auto-fix subcommand suggestions & stale remediation command.

## Evidence Notes

- `status.json` shows legacy IDs and only counts actual discovered statuses (STUB, TESTED) without zero buckets for other statuses.
- No canonical CBIN IDs; thus all target requirements considered unmet.
- Lack of self-canary tokens blocks dogfood acceptance case.
- Staleness threshold mismatch (60d vs 30d) and aggregated error output prevents required per-line diagnostics.

---
Prepared 2025-09-20 based on current repository state before canonical implementation migration.
