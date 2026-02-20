# CANARY Token Specification

## Format

CANARY tokens track requirements directly in source code:

```
// CANARY: REQ=CBIN-###; FEATURE="Name"; ASPECT=API; STATUS=IMPL; [TEST=TestName]; [BENCH=BenchName]; [OWNER=team]; UPDATED=YYYY-MM-DD
```

## Required Fields

- **REQ**: Requirement ID (format: CBIN-###)
- **FEATURE**: Short feature name
- **ASPECT**: Category (API, CLI, Engine, Storage, etc.)
- **STATUS**: Implementation state
- **UPDATED**: Last update date (YYYY-MM-DD)

## Status Values

- **MISSING**: Planned but not implemented
- **STUB**: Placeholder implementation
- **IMPL**: Implemented
- **TESTED**: Implemented with tests (auto-promoted from IMPL+TEST)
- **BENCHED**: Tested with benchmarks (auto-promoted from TESTED+BENCH)
- **REMOVED**: Deprecated/removed

## Optional Fields

- **TEST**: Test function name (promotes IMPL → TESTED)
- **BENCH**: Benchmark function name (promotes TESTED → BENCHED)
- **OWNER**: Team/person responsible

## Example

```go
// CANARY: REQ=CBIN-001; FEATURE="UserAuth"; ASPECT=API; STATUS=TESTED; TEST=TestUserAuth; OWNER=backend; UPDATED=2025-10-16
func AuthenticateUser(credentials *Credentials) (*Session, error) {
    // implementation
}
```

## Usage

```bash
# Scan for tokens and generate reports
canary scan --root . --out status.json --csv status.csv

# Verify GAP_ANALYSIS.md claims
canary scan --root . --verify GAP_ANALYSIS.md

# Check for stale tokens (30-day threshold)
canary scan --root . --strict

# Auto-update stale TESTED/BENCHED tokens
canary scan --root . --update-stale
```
