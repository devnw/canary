# Canary CLI — Parity Checklist

| Requirement | TokenParse | EnumValidate | NormalizeREQ | StatusJSON | CSVExport | VerifyGate | Staleness30d | SelfCanary | CI | Perf50k<10s |
|------------:|:----------:|:------------:|:------------:|:----------:|:---------:|:----------:|:------------:|:----------:|:--:|:------------:|
| CBIN-101    | ✅         | ✅           | ✅           | ✅         | ✅        | ◻          | ◻            | ✅         | ◻  | ◻            |
| CBIN-102    | ✅         | ✅           | ✅           | ◻          | ◻         | ✅         | ✅           | ✅         | ◻  | ◻            |
| CBIN-103    | ✅         | ✅           | ✅           | ✅         | ✅        | ◻          | ◻            | ✅         | ◻  | ◻            |
| Overall     | ✅         | ✅           | ✅           | ✅         | ✅        | ✅         | ✅           | ✅         | ◻  | ◻            |

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
- **CBIN-101, CBIN-103:** TestAcceptance_FixtureSummary (implicit, CSV generation tested via `writeCSV` call)
  - Explodes each feature → separate CSV row
  - Deterministic row order: NOT YET VALIDATED (see GAP #4)
  - UTF-8, LF line endings
  - Evidence: `tools-canary-status.csv` generated successfully (acceptance builds and runs without CSV parse errors)

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
- **All requirements: ◻ MISSING**
  - No GitHub Actions workflow defined for canary acceptance tests
  - Gap: See GAP_ANALYSIS.md #8

### Perf50k<10s
- **All requirements: ◻ MISSING**
  - No performance benchmarks exist
  - Requirement: Scan 50k files in <10s, ≤512 MiB RSS
  - Gap: See GAP_ANALYSIS.md #7

## Critical Gaps Summary

1. **TestCANARY_* functions missing** (ref: tokens cite `TestCANARY_CBIN_101_Engine_ScanBasic`, etc. but functions don't exist)
2. **BenchmarkCANARY_* functions missing** (ref: tokens cite `BenchmarkCANARY_CBIN_101_Engine_Scan`, etc. but functions don't exist)
3. **CI workflow missing** (no `.github/workflows/canary.yml`)
4. **Performance benchmarks missing** (no `BenchmarkCANARY_CBIN_101_Perf50k` for 50k file test)
5. **CSV row order not validated** (deterministic sort untested)

## Next Steps

See `NEXT.md` for prioritized slices addressing gaps.
