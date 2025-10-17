# Requirements Gap Analysis

## Claimed Requirements

List requirements that are fully implemented and verified:

✅ CBIN-101 - ScannerCore engine (BENCHED)
✅ CBIN-102 - VerifyGate CLI (BENCHED)

## Gaps

List requirements that are planned or in progress:

- [ ] CBIN-103 - Status schema and JSON output

## Verification

Run verification with:

```bash
canary scan --root . --verify GAP_ANALYSIS.md
```

This will:
- ✅ Verify claimed requirements are TESTED or BENCHED
- ❌ Fail with exit code 2 if claims are overclaimed
