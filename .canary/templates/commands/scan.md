---
description: Scan codebase for CANARY tokens and generate status reports
---

# CANARY: REQ=CBIN-111; FEATURE="ScanCmd"; ASPECT=CLI; STATUS=IMPL; OWNER=canary; UPDATED=2025-10-16

## User Input

```text
$ARGUMENTS
```

## Outline

Scan the codebase for CANARY requirement tokens and generate comprehensive status reports.

1. **Determine scan scope**:
   - Default: Current project root
   - If arguments provided: Parse for --root, --out, --csv options
   - Apply default skip pattern: `.git, node_modules, vendor, bin, dist, build`

2. **Run canary scanner**:
   ```bash
   canary scan --root . --out status.json --csv status.csv
   ```

   Optional flags from arguments:
   - `--strict`: Enforce 30-day staleness check
   - `--skip`: Custom regex pattern for excluded paths
   - `--verify`: Path to GAP_ANALYSIS.md for claim verification

3. **Parse scanner output**:
   - Read `status.json` for structured data
   - Extract key metrics:
     - Total tokens by status (STUB, IMPL, TESTED, BENCHED)
     - Coverage by aspect (API, CLI, Engine, etc.)
     - Stale tokens (if --strict used)
     - Unique requirements

4. **Generate summary report**:
   ```markdown
   ## CANARY Token Scan Results

   **Scan Date:** YYYY-MM-DD
   **Total Requirements:** N

   ### Status Distribution
   - BENCHED: X (Y%)
   - TESTED: X (Y%)
   - IMPL: X (Y%)
   - STUB: X (Y%)
   - MISSING: X (Y%)

   ### Aspect Coverage
   - API: X tokens
   - CLI: X tokens
   - Engine: X tokens
   - Storage: X tokens

   ### Quality Metrics
   - Test Coverage: X% (TESTED+BENCHED / total)
   - Benchmark Coverage: X% (BENCHED / total)
   - Stale Tokens: X (if --strict used)

   **Reports Generated:**
   - status.json (detailed JSON)
   - status.csv (spreadsheet format)
   ```

5. **Identify action items**:
   - Stale tokens needing updates
   - STUB/IMPL requirements needing tests
   - TESTED requirements that could use benchmarks
   - Missing OWNER assignments

6. **Suggest next steps**:
   - If stale tokens found: "Run `canary scan --update-stale` to auto-update"
   - If STUB tokens found: "Use `/canary.plan` to plan implementation for CBIN-XXX"
   - If IMPL tokens without tests: "Add TEST= field and create test functions"

## Example Output

```markdown
## CANARY Token Scan Results

**Scan Date:** 2025-10-16
**Total Requirements:** 10

### Status Distribution
- BENCHED: 3 (30%) ✅
- TESTED: 4 (40%) ✅
- IMPL: 2 (20%) ⚠️
- STUB: 1 (10%) ⚠️

### Aspect Coverage
- API: 4 tokens
- CLI: 3 tokens
- Engine: 2 tokens
- Storage: 1 token

### Quality Metrics
- Test Coverage: 70% (TESTED+BENCHED)
- Benchmark Coverage: 30% (BENCHED)
- Stale Tokens: 2 (CBIN-001, CBIN-004)

**Reports Generated:**
- [status.json](./status.json) - Detailed JSON report
- [status.csv](./status.csv) - Spreadsheet format

### Action Items

1. **Update Stale Tokens**: Run `canary scan --update-stale`
   - CBIN-001: UserAuth (updated 2024-01-01, age 288 days)
   - CBIN-004: Cache (updated 2024-01-01, age 288 days)

2. **Add Tests**: 2 IMPL requirements need tests
   - CBIN-003: DataValidation
   - CBIN-007: ReportGenerator

3. **Add Benchmarks**: 4 TESTED requirements could use performance benchmarks
   - CBIN-002: TokenParser
   - CBIN-005: Serializer
```

## Guidelines

- **Automatic Execution**: Run scanner without prompting user
- **Clear Visualization**: Use tables, percentages, and status indicators
- **Actionable Output**: Provide specific commands/next steps
- **Links**: Link to generated report files for easy access
- **Trend Analysis**: If running repeatedly, show improvement over time
