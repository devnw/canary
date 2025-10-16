# canary

Scan, update, create, verify and manage **CANARY** tokens across
repositories, emit `status.json` / `status.csv`, and **verify** GAP claims.

**Now with full spec-kit integration!** Track all 46 spec-kit features with comprehensive requirements documentation. See [`docs/SPEC_KIT_INTEGRATION_SUMMARY.md`](docs/SPEC_KIT_INTEGRATION_SUMMARY.md) for details.

## Build

```bash
go build -o ./bin/canary ./cmd/canary
```

## CLI Commands

Canary provides spec-kit-inspired commands for managing CANARY tokens:

### Initialize a New Project

```bash
canary init <project-name>
# Creates README_CANARY.md and GAP_ANALYSIS.md templates
```

### Create a New Requirement Token

```bash
canary create CBIN-105 "FeatureName" --aspect API --status IMPL --owner team
# Outputs a properly formatted CANARY token ready to paste
```

### Scan for Tokens

```bash
canary scan --root . --out status.json --csv status.csv
canary scan --root . --verify GAP_ANALYSIS.md --strict
canary scan --root . --update-stale  # Auto-update stale TESTED/BENCHED tokens
```

### Exit Codes

- **Exit 0**: OK
- **Exit 2**: Verification/staleness failed
- **Exit 3**: Parse or IO error

### Legacy Usage

The standalone scanner is still available at `tools/canary`:

```bash
go run ./tools/canary --root . --out status.json
```

**Token format**

```text
Example template (replace with actual values):
CANARY: REQ=CBIN-101; FEATURE="MyFeature"; ASPECT=API; STATUS=IMPL; TEST=TestCANARY_CBIN_101_API_MyFeature; BENCH=BenchmarkCANARY_CBIN_101_API_MyFeature; OWNER=team; UPDATED=2025-10-15

Valid ASPECT values: API, CLI, Engine, Planner, Storage, Wire, Security, Docs, Encode, Decode, RoundTrip, Bench, FrontEnd, Dist
Valid STATUS values: MISSING, STUB, IMPL, TESTED, BENCHED, REMOVED
```

**Supported comment styles**: `//`, `#`, `--`, `<!--` (Python, Go, Bash, SQL, Markdown, etc.)

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

## Spec-Kit Integration

This project includes comprehensive tracking for the **spec-kit** submodule:

- **46 Requirements** across 10 categories (Core Workflows, CLI Tools, Templates, etc.)
- **Complete Documentation**: Requirements catalog, gap analysis, integration guide, examples
- **Sample Implementation**: 7 tokens demonstrating the integration pattern
- **Enhanced Scanner**: Support for HTML-style comments in Markdown files

### Quick Start

```bash
# Scan spec-kit
./canary --root ./specs/spec-kit-repo --out spec-kit-status.json --csv spec-kit-status.csv

# Verify against gap analysis
./canary verify --root ./specs/spec-kit-repo --gap docs/SPEC_KIT_GAP_ANALYSIS.md --strict
```

### Documentation

- [Integration Summary](docs/SPEC_KIT_INTEGRATION_SUMMARY.md) - Overview and status
- [Requirements Catalog](docs/SPEC_KIT_REQUIREMENTS.md) - All 46 requirements
- [Gap Analysis](docs/SPEC_KIT_GAP_ANALYSIS.md) - Tracking document
- [Integration Guide](docs/SPEC_KIT_INTEGRATION_GUIDE.md) - Step-by-step instructions
- [CANARY Examples](docs/CANARY_EXAMPLES_SPEC_KIT.md) - Token patterns and examples
