# Canary CLI — Parity Checklist

| Requirement | TokenParse | EnumValidate | NormalizeREQ | StatusJSON | CSVExport | VerifyGate | Staleness30d | SelfCanary | CI | Perf50k<10s |
|------------:|:----------:|:------------:|:------------:|:----------:|:---------:|:----------:|:------------:|:----------:|:--:|:------------:|
| CBIN-101    | ✅         | ✅           | ✅           | ✅         | ✅        | ◻          | ◻            | ✅         | ✅ | ✅            |
| CBIN-102    | ✅         | ✅           | ✅           | ◻          | ◻         | ✅         | ✅           | ✅         | ✅ | ✅            |
| CBIN-103    | ✅         | ✅           | ✅           | ✅         | ✅        | ◻          | ◻            | ✅         | ✅ | ✅            |
| Overall     | ✅         | ✅           | ✅           | ✅         | ✅        | ✅         | ✅           | ✅         | ✅ | ✅            |

**Legend:** ✅ = proven by tests/evidence; ◐ = partial; ◻ = missing

## Evidence Links (By Column)

### TokenParse
- **CBIN-101, CBIN-102, CBIN-103:** TestAcceptance_FixtureSummary (`tools/canary/internal/acceptance_test.go:52`)
  - Parses `CANARY: REQ=...; FEATURE="..."; ASPECT=...; STATUS=...; ...` from fixture files
  - Regex: `tokenLineRe` in `tools/canary/main.go:53`
  - Diagnostic: None (implicit pass via JSON output)

### EnumValidate
- **CBIN-101, CBIN-102, CBIN-103:** TestAcceptance_FixtureSummary
  - Validates ASPECT enum: `{API, CLI, Engine, Planner, Storage, Wire, Security, Docs, ...}` (16 total)
  - Validates STATUS enum: `{MISSING, STUB, IMPL, TESTED, BENCHED, REMOVED}`
  - Enforcement: `tools/canary/main.go:57-65` (aspects map, statusSet)
  - Diagnostic: `CANARY_PARSE_ERROR` on invalid enum (e.g., `ASPECT=<ASPECT>` in CRUSH.md causes parse failure)

### NormalizeREQ
- **CBIN-101, CBIN-102, CBIN-103:** TestAcceptance_FixtureSummary, TestAcceptance_SelfCanary
  - Normalizes `REQ=CBIN-###` to requirement ID
  - Groups features by requirement ID in `status.json` → `requirements[].id`
  - Evidence: `tools-canary-status.json` shows 3 unique IDs: CBIN-101, CBIN-102, CBIN-103

### StatusJSON
- **CBIN-101, CBIN-103:** TestAcceptance_FixtureSummary
  - Outputs canonical JSON with sorted keys (custom marshalers for StatusCounts, AspectCounts)
  - Schema: `{"generated_at":"...", "requirements":[...], "summary":{"by_status":{...}, "by_aspect":{...}}}`
  - Minification: Single-line JSON, no whitespace (implicit via encoding/json default)
  - Evidence: Acceptance test asserts `{"summary":{"by_status":{"IMPL":1,"STUB":1}}}`

### CSVExport
- **CBIN-101, CBIN-103:** TestAcceptance_FixtureSummary, TestAcceptance_CSVOrder (`tools/canary/internal/acceptance_test.go:136`)
  - Explodes each feature → separate CSV row
  - Deterministic row order: ✅ VALIDATED (Slice 9, 2025-10-15)
  - UTF-8, LF line endings
  - Evidence: `tools-canary-status.csv` generated successfully, TestAcceptance_CSVOrder verifies byte-for-byte determinism

### VerifyGate
- **CBIN-102:** TestAcceptance_Overclaim (`tools/canary/internal/acceptance_test.go:79`)
  - Parses GAP_ANALYSIS.md for `✅\s+(CBIN-\d{3})` claims
  - Compares claimed requirements vs. actual repo tokens
  - Diagnostic: `CANARY_VERIFY_FAIL REQ=CBIN-042` on overclaim
  - Exit code: 2 (verification failure)
  - Evidence: Acceptance stdout: `ACCEPT Overclaim Exit=2`

### Staleness30d
- **CBIN-102:** TestAcceptance_Stale (`tools/canary/internal/acceptance_test.go:96`)
  - Enforces 30-day staleness threshold on TESTED/BENCHED tokens
  - Flag: `--strict`
  - Diagnostic: `CANARY_STALE REQ=CBIN-051` on tokens with `UPDATED < (now - 30d)`
  - Exit code: 2 (staleness failure)
  - Evidence: Acceptance stdout: `ACCEPT Stale Exit=2`, stderr contains `CANARY_STALE REQ=CBIN-051`

### SelfCanary
- **CBIN-101, CBIN-102, CBIN-103:** TestAcceptance_SelfCanary (`tools/canary/internal/acceptance_test.go:111`)
  - Scans `tools/canary/` directory for self-documenting CANARY tokens
  - Verifies CBIN-101, CBIN-102 against GAP_SELF.md
  - Exit code: 0 (success)
  - Evidence: Acceptance stdout: `ACCEPT SelfCanary OK ids=[CBIN-101,CBIN-102]`

### CI
- **All requirements: ✅ VALIDATED (Slice 8, 2025-10-15)**
  - GitHub Actions workflow: `.github/workflows/canary.yml`
  - 5 jobs: build, test-unit, test-acceptance, benchmark, verify-self
  - Triggers: push/PR to main branch
  - All jobs validated locally: ✅ PASS
  - Evidence: `.github/workflows/canary.yml:1`

### Perf50k<10s
- **All requirements: ✅ VALIDATED (Slice 10, 2025-10-15)**
  - 50k file benchmark: BenchmarkCANARY_CBIN_101_Engine_Scan50k
  - **Actual performance: 1.85s** (81.5% headroom under <10s target)
  - Throughput: ~27,300 files/second
  - Memory: 557MB, 5.5M allocs
  - Evidence: `tools/canary/main_test.go:102`
  - Baselines established:
    - BenchmarkCANARY_CBIN_101_Engine_Scan: 3.3ms/100 files, 1.1MB, 11353 allocs
    - BenchmarkCANARY_CBIN_101_Engine_Scan50k: 1.85s/50k files, 557MB, 5.5M allocs
    - BenchmarkCANARY_CBIN_102_CLI_Verify: 36µs/50 claims, 5.2KB, 13 allocs
    - BenchmarkCANARY_CBIN_103_API_Emit: 0.9ms/300 tokens, 36KB, 2119 allocs

## Critical Gaps Summary

1. ~~**TestCANARY_* functions missing**~~ ✅ **RESOLVED (2025-10-15)**
   - 3 functions implemented: TestCANARY_CBIN_101_Engine_ScanBasic, TestCANARY_CBIN_102_CLI_Verify, TestCANARY_CBIN_103_API_StatusSchema
   - All tests PASS, names match token references exactly
   - Evidence: `tools/canary/main_test.go:16`, `tools/canary/verify_test.go:11`, `tools/canary/status_test.go:12`

2. ~~**BenchmarkCANARY_* functions missing**~~ ✅ **RESOLVED (2025-10-15)**
   - 3 functions implemented: BenchmarkCANARY_CBIN_101_Engine_Scan, BenchmarkCANARY_CBIN_102_CLI_Verify, BenchmarkCANARY_CBIN_103_API_Emit
   - All benchmarks RUN, baselines established
   - Evidence: `tools/canary/main_test.go:86`, `tools/canary/verify_test.go:123`, `tools/canary/status_test.go:167`

3. ~~**CI workflow missing**~~ ✅ **RESOLVED (2025-10-15 Slice 8)**
   - Workflow created: `.github/workflows/canary.yml` with 5 jobs
   - Evidence: `.github/workflows/canary.yml:1`

4. ~~**Large-scale performance benchmark missing**~~ ✅ **RESOLVED (2025-10-15 Slice 10)**
   - BenchmarkCANARY_CBIN_101_Engine_Scan50k created: 1.85s for 50k files (81.5% under target)
   - Evidence: `tools/canary/main_test.go:102`

5. ~~**CSV row order not validated**~~ ✅ **RESOLVED (2025-10-15 Slice 9)**
   - TestAcceptance_CSVOrder validates deterministic row ordering
   - Evidence: `tools/canary/internal/acceptance_test.go:136`

## Next Steps

See `NEXT.md` for prioritized slices addressing gaps.
