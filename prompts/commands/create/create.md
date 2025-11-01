# Create Command Prompt

## Purpose
Generate a new CANARY token template.

## Task
Implement `canary create` to generate properly formatted CANARY tokens.

## Expected Behavior
```bash
canary create CBIN-105 "FeatureName" --aspect API --status IMPL
```

## Output
```
// CANARY: REQ=CBIN-105; FEATURE="FeatureName"; ASPECT=API; STATUS=IMPL; UPDATED=2025-11-01
```

## Options
- `--aspect`: API, CLI, Storage, Security, etc.
- `--status`: STUB, IMPL, TESTED, BENCHED
- `--owner`: Team or person responsible
- `--test`: Test function name
- `--bench`: Benchmark function name

## Standards
- Auto-generate UPDATED field with current date
- Validate aspect and status enums
- Format with proper spacing and semicolons
- Support copy-paste into source files
