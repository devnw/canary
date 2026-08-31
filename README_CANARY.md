# CANARY Token Specification

## Format

CANARY tokens track requirements directly in source code:

```
// CANARY: REQ=CBIN-###; FEATURE="Name"; ASPECT=API; STATUS=IMPL; [TEST=TestName]; [BENCH=BenchName]; [OWNER=team]; UPDATED=<YYYY-MM-DD>
```

## Required Fields

- **REQ**: Requirement ID. The prefix is your project's configured source key
  (set via `canary init --key`, e.g. `CBIN`, `PROJ`, `ENG`), followed by a
  number — e.g. `CBIN-123`. Ticket-source keys (e.g. a JIRA project key) are
  configured per project; `CBIN-###` is only the default example.
- **FEATURE**: Short feature name
- **ASPECT**: Category (API, CLI, Engine, Storage, etc.)
- **STATUS**: Implementation state
- **UPDATED**: Last update date (YYYY-MM-DD)

## Status Values

- **MISSING**: Planned but not implemented
- **STUB**: Placeholder implementation
- **IMPL**: Implemented
- **TESTED**: Declared as implemented with tests
- **BENCHED**: Declared as tested with benchmarks
- **REMOVED**: Deprecated/removed

> STATUS is a **declaration**, written by the author. CANARY never promotes it
> automatically. Whether a requirement is actually done is answered by
> *verification* against evidence (passing TEST=/BENCH= at the scanned commit),
> exported as the `verified` set — see
> [docs/canary-evidence.schema.json](docs/canary-evidence.schema.json).

## Optional Fields

- **TEST**: Test function name (recorded as evidence for verification; does not change STATUS)
- **BENCH**: Benchmark function name (recorded as evidence for verification; does not change STATUS)
- **OWNER**: Team/person responsible

## Example

```go
// CANARY: REQ=CBIN-001; FEATURE="UserAuth"; ASPECT=API; STATUS=TESTED; TEST=TestUserAuth; OWNER=backend; UPDATED=2026-08-29
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

# Report which stale claims still have current PASS evidence (mutates nothing)
canary scan --root . --update-stale
```
