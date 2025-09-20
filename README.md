# canary

Scan, update, create, verify and manage **CANARY** tokens across
repositories, emit `status.json` / `status.csv`, and **verify** GAP claims.

## Build

```bash
go build -o ./bin/canary ./cmd/canary
```

## Usage

```bash
./bin/canary --root . --out status.json --csv status.csv
./bin/canary --root . --verify GAP_ANALYSIS.md --strict
./bin/canary verify --root . --gap docs/GAP_ANALYSIS.md --strict --skip '(^|/)(.git|.direnv|node_modules|vendor|bin|dist|build|zig-out|.zig-cache)(/|$)'; echo EXIT:$?
```

- **Exit 0**: OK
- **Exit 2**: Verification/staleness failed
- **Exit 3**: Parse or IO error

**Token format**

```text
CANARY: REQ=REQ-GQL-###; FEATURE="Name"; ASPECT=...; STATUS=MISSING|STUB|IMPL|TESTED|BENCHED|REMOVED; TEST=...; BENCH=...; OWNER=...; UPDATED=YYYY-MM-DD
```

## Status Auto-Promotion

The scanner auto-promotes statuses based on evidence references:

| From        | Evidence Condition    | To      |
| ----------- | --------------------- | ------- |
| IMPL        | ≥1 test (TEST=)       | TESTED  |
| IMPL/TESTED | ≥1 benchmark (BENCH=) | BENCHED |

Notes:

- Promotion is applied in-memory; original source comments remain unchanged.
- BENCHED dominates TESTED in summary counts.
- `--strict` still validates staleness on TESTED/BENCHED after promotion.
- A future `--no-promote` flag may allow raw status reporting.

Example: if a feature is marked `STATUS=IMPL` and has a `TEST=TestCANARY_REQ_GQL_030_TxnCommit`, the report will show it as `TESTED`.

## Testing

```bash
cd tools/canary
go test -v
```

## CANARY at a glance

Policy excerpt (see `docs/CANARY_POLICY.md`). Example tokens:

`CANARY: REQ=CBIN-101; FEATURE="ScannerCore"; ASPECT=Engine; STATUS=TESTED; TEST=TestCANARY_CBIN_101_Engine_ScanBasic; BENCH=BenchmarkCANARY_CBIN_101_Engine_Scan; OWNER=canary; UPDATED=2025-09-20`

`CANARY: REQ=CBIN-102; FEATURE="VerifyGate"; ASPECT=CLI; STATUS=TESTED; TEST=TestCANARY_CBIN_102_CLI_Verify; BENCH=BenchmarkCANARY_CBIN_102_CLI_Verify; OWNER=canary; UPDATED=2025-09-20`

