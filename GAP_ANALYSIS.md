# Requirements Gap Analysis

## Claimed Requirements

List requirements that are fully implemented and verified:

✅ CBIN-101 - ScannerCore Engine fully benchmarked
✅ CBIN-102 - VerifyGate CLI fully benchmarked

## Gaps

List requirements that are planned or in progress:

- [ ] CBIN-103 - StatusJSON (STATUS=BENCHED, needs more coverage)
- [ ] CBIN-132 - NextCmd (STATUS=BENCHED, needs more coverage)

## Verification

Run verification with:

```bash
canary scan --root . --verify GAP_ANALYSIS.md
```

This will:
- ✅ Verify claimed requirements are TESTED or BENCHED
- ❌ Fail with exit code 2 if claims are overclaimed
