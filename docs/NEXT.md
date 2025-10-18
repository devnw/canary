# NEXT — Canary CLI (Updated: 2025-09-20)

## Completed (this slice)

None — migration to canonical CBIN spec not yet started.

## Up Next (small, verifiable slices)

### Slice 1: Canonical Scanner Skeleton

Scope: Introduce `tools/canary/` with `main.go`, `status.go`, `verify.go` implementing CLI flags, minified JSON (ordered keys), deterministic CSV, strict enums, `--skip` regex.

Acceptance:

```bash
go build -o ./bin/canary ./tools/canary && ./bin/canary --root tools/canary --out status.json --csv status.csv
```

Expected stdout contains: `CANARY_OK wrote status.json`.

CANARY: add tokens CBIN-101 (Engine ScanBasic) status TESTED placeholder test name.

### Slice 2: Self-Canary Tokens & Verify Gate

Scope: Add tokens CBIN-101..103 (ScannerCore, VerifyGate, StatusJSON) and implement strict verify regex `^✅ CBIN-\d{3}`; exit 2 with `CANARY_VERIFY_FAIL` diagnostics.

Acceptance:

```bash
./bin/canary --root tools/canary --verify GAP_ANALYSIS.md --strict || echo EXIT=$?
```

With GAP containing only ✅ CBIN-101,102 should exit 0; adding an unimplemented ✅ CBIN-999 line should produce `CANARY_VERIFY_FAIL REQ=CBIN-999` and exit 2.

Benchmark: `BenchmarkCANARY_CBIN_101_Engine_Scan` initial run ns/op captured.

CANARY: Test names per spec for 101,102,103.

### Slice 3: Staleness 30d & Diagnostics

Scope: Implement 30-day window (config via flag constant), emit one line per stale token `CANARY_STALE REQ=<id> updated=<date> age_days=<N> threshold=30`.

Acceptance: Create a token with UPDATED=2024-01-01 STATUS=TESTED; run strict; expect exit 2 and diagnostic line.

Benchmark guard unchanged.

CANARY: Add `TestCANARY_CBIN_102_CLI_Verify` referencing stale scenario fixture.

### Slice 4: Acceptance Test Harness

Scope: Add tests under `tools/canary/internal/...` implementing four acceptance cases (Fixture summary, Overclaim, Stale, Self-verify) printing required sentinel lines.

Acceptance: `go test ./tools/canary/... -run TestAcceptance -v` prints all sentinel lines and exits 0.

Benchmark: Add baseline bench assertions (skip if testing race). Guard: `allocs/op <= 20` for scan bench.

CANARY: Ensure test names match token declarations.

### Slice 5: CI Workflow Gate

Scope: Add `.github/workflows/canary.yml` building, scanning, self-verifying, uploading artifacts.

Acceptance: GitHub Actions run shows successful job `scan` and uploaded artifacts `canary-status`.

CANARY: Add bench token for CBIN-101 BENCH field.

### Slice 6: Performance Fixture (Optional Stretch)

Scope: Generate 10k synthetic small text files and measure scan time; ensure <2s local baseline to extrapolate <10s for 50k.

Acceptance: Local bench (document in GAP_ANALYSIS) shows `BenchmarkCANARY_CBIN_101_Engine_Scan` with ns/op ≤ target threshold.

Guard: `ns/op ≤ 2.0e6`.

CANARY: Add or update bench token for CBIN-101 once stable.

---
Prioritization Rationale: Establish canonical correctness (Slices 1–3) before test harness & CI (Slices 4–5); performance tuning deferred until correctness baseline exists.
