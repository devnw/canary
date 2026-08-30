---
description: Scan codebase for CANARY tokens and generate status reports
---


## User Input

```text
$ARGUMENTS
```

## Outline

Scan the codebase for CANARY requirement tokens and generate comprehensive status reports.

1. **Determine project key**:
   - Read `.canary/project.yaml` for the project key (e.g., `PROJECT_KEY`)
   - Use the key when referencing requirement IDs: `<PROJECT_KEY>-<NNN>`

2. **Determine scan scope**:
   - Default: Current project root
   - If arguments provided: Parse for --root, --out, --csv options
   - Apply default skip pattern: `.git, node_modules, vendor, bin, dist, build`

3. **Run canary scanner**:

   ```bash
   canary scan --root . --out status.json --csv status.csv
   ```

   Optional flags from arguments:
   - `--strict`: Enforce 30-day staleness check
   - `--skip`: Custom regex pattern for excluded paths
   - `--verify`: Path to GAP_ANALYSIS.md for claim verification

4. **Use scanner stdout (minimal context)**:
   - The scanner prints a **single line** to stdout: `CANARY_SCAN tokens=N requirements=M STUB=a IMPL=b TESTED=c BENCHED=d`
   - Use this line for metrics; **do not read status.json** unless you need per-requirement or per-file detail.
   - If `--verify` was used: stdout also has `CANARY_VERIFY_OK` (pass) or `CANARY_VERIFY_FAIL count=N` (fail); stderr lists failing REQs.

5. **Generate summary report**:

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

6. **Identify action items**:
   - Stale tokens needing updates
   - STUB/IMPL requirements needing tests
   - TESTED requirements that could use benchmarks
   - Missing OWNER assignments

7. **Suggest next steps**:
   - If stale tokens found: "Run `canary scan --update-stale` to report which still have current evidence"
   - If STUB tokens found: "Use `/canary.plan` to plan implementation for `<PROJECT_KEY>-<NNN>`"
   - If IMPL tokens without tests: "Add TEST= field and create test functions"

## Example Output

```markdown
## CANARY Token Scan Results

**Scan Date:** 2025-10-16
**Total Requirements:** 10

### Status Distribution

- BENCHED: 3 (30%)
- TESTED: 4 (40%)
- IMPL: 2 (20%)
- STUB: 1 (10%)

### Aspect Coverage

- API: 4 tokens
- CLI: 3 tokens
- Engine: 2 tokens
- Storage: 1 token

### Quality Metrics

- Test Coverage: 70% (TESTED+BENCHED)
- Benchmark Coverage: 30% (BENCHED)
- Stale Tokens: 2 (<PROJECT_KEY>-001, <PROJECT_KEY>-004)

**Reports Generated:**

- [status.json](./status.json) - Detailed JSON report
- [status.csv](./status.csv) - Spreadsheet format

### Action Items

1. **Check Stale Tokens**: Run `canary scan --update-stale` (reports evidence currency; rewrites nothing)
   - <PROJECT_KEY>-001: UserAuth (updated 2024-01-01, age 288 days)
   - <PROJECT_KEY>-004: Cache (updated 2024-01-01, age 288 days)

2. **Add Tests**: 2 IMPL requirements need tests
   - <PROJECT_KEY>-003: DataValidation
   - <PROJECT_KEY>-007: ReportGenerator

3. **Add Benchmarks**: 4 TESTED requirements could use performance benchmarks
   - <PROJECT_KEY>-002: TokenParser
   - <PROJECT_KEY>-005: Serializer
```

## Guidelines

- **Automatic Execution**: Run scanner without prompting user
- **Clear Visualization**: Use tables, percentages, and status indicators
- **Actionable Output**: Provide specific commands/next steps
- **Links**: Link to generated report files for easy access
- **Trend Analysis**: If running repeatedly, show improvement over time
