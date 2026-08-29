# Scan Command Prompt

## Purpose
Scan the codebase for CANARY tokens and generate reports.

## Task
You are implementing the `canary scan` command. This command:
- Recursively scans directories for CANARY tokens
- Parses token fields (REQ, FEATURE, ASPECT, STATUS, TEST, BENCH, OWNER, UPDATED)
- Validates enum values and formats
- Generates JSON and CSV reports
- Supports verification mode to check gap analysis claims
- Enforces staleness checks in strict mode

## Expected Behavior

### Basic Scan
```bash
canary scan --root . --out status.json --csv status.csv
```

Outputs:
- `status.json` - Structured token database
- `status.csv` - Flattened CSV for analysis

### Verification Mode
```bash
canary scan --root . --verify GAP_ANALYSIS.md --strict
```

- Checks if claimed requirements match actual implementation status
- Fails (exit 2) if GAP_ANALYSIS.md claims a requirement is implemented but tokens show STATUS=STUB
- Fails (exit 2) if STATUS=TESTED/BENCHED tokens have UPDATED > 30 days old (when --strict)

### Exit Codes
- 0: Success
- 1: General error
- 2: Verification or staleness failure
- 3: Parse/IO error

## Token Format
```
CANARY: REQ=CBIN-###; FEATURE="Name"; ASPECT=API; STATUS=IMPL; TEST=TestName; BENCH=BenchName; OWNER=team; UPDATED=YYYY-MM-DD
```

## Output Schema

### status.json
```json
{
  "generated_at": "2025-11-01T12:00:00Z",
  "requirements": [
    {
      "id": "CBIN-001",
      "features": [
        {
          "feature": "UserAuth",
          "aspect": "API",
          "status": "TESTED",
          "files": ["src/auth.go"],
          "tests": ["TestCANARY_CBIN_001_UserAuth"],
          "benches": [],
          "owner": "backend",
          "updated": "2025-10-15"
        }
      ]
    }
  ],
  "summary": {
    "by_status": {"STUB": 5, "IMPL": 10, "TESTED": 8, "BENCHED": 2},
    "by_aspect": {"API": 12, "CLI": 8, "Storage": 5}
  }
}
```

### status.csv
```csv
req,feature,aspect,status,file,test,bench,owner,updated
CBIN-001,UserAuth,API,TESTED,src/auth.go,TestCANARY_CBIN_001_UserAuth,,backend,2025-10-15
```

## Standards
- Use streaming file processing for performance
- Handle malformed tokens gracefully (warn, don't fail)
- Normalize requirement IDs (CBIN-001, CBIN-1 -> CBIN-001)
- Stable sort for deterministic output
- No secrets in output

## Implementation Notes
- See `pkg/storage` for token database schema
- Use `pkg/reqid` for requirement ID normalization
- Use `pkg/docs` for GAP_ANALYSIS.md parsing
- Performance target: <10s for 50k files
