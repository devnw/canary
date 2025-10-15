# Spec-Kit Integration Quick Reference

## Commands

### Scan Spec-Kit
```bash
./canary --root ./specs/spec-kit --out spec-kit-status.json --csv spec-kit-status.csv
```

### Verify Against GAP
```bash
./canary verify --root ./specs/spec-kit --gap docs/SPEC_KIT_GAP_ANALYSIS.md --strict
```

### Check Staleness
```bash
./canary --root ./specs/spec-kit --out status.json --strict
```

### View Results
```bash
cat spec-kit-status.json | jq '.summary'
cat spec-kit-status.csv | column -t -s,
```

## Token Format

### Python
```python
# CANARY: REQ=REQ-SK-###; FEATURE="Name"; ASPECT=CLI; STATUS=IMPL; OWNER=team; UPDATED=2025-10-15
```

### Bash
```bash
# CANARY: REQ=REQ-SK-###; FEATURE="Name"; ASPECT=Automation; STATUS=IMPL; OWNER=team; UPDATED=2025-10-15
```

### Markdown/HTML
```markdown
<!-- CANARY: REQ=REQ-SK-###; FEATURE="Name"; ASPECT=Templates; STATUS=IMPL; OWNER=team; UPDATED=2025-10-15 -->
```

### TOML
```toml
# CANARY: REQ=REQ-SK-###; FEATURE="Name"; ASPECT=Agent; STATUS=IMPL; OWNER=team; UPDATED=2025-10-15
```

## Required Fields

- `REQ`: Requirement ID (REQ-SK-###)
- `FEATURE`: Short name (quoted if spaces)
- `ASPECT`: Category (CLI, Testing, Templates, etc.)
- `STATUS`: MISSING/STUB/IMPL/TESTED/BENCHED/REMOVED

## Optional Fields

- `TEST`: Comma-separated test names
- `BENCH`: Comma-separated benchmark names
- `OWNER`: Team/area responsible
- `UPDATED`: Date (YYYY-MM-DD)

## Status Progression

```
MISSING → STUB → IMPL → TESTED → BENCHED
                   ↓
               REMOVED
```

## Auto-Promotion

- `IMPL` + `TEST=` → Auto-promotes to `TESTED`
- `(IMPL|TESTED)` + `BENCH=` → Auto-promotes to `BENCHED`

## Requirements Ranges

| Category | Range | Count |
|----------|-------|-------|
| Core Workflow Commands | REQ-SK-101 to 108 | 8 |
| CLI Tool Features | REQ-SK-201 to 203 | 3 |
| Template System | REQ-SK-301 to 306 | 6 |
| Constitutional Framework | REQ-SK-401 to 409 | 5 |
| Script Automation | REQ-SK-501 to 504 | 4 |
| Agent Support | REQ-SK-601 to 605 | 5 |
| Documentation System | REQ-SK-701 to 704 | 4 |
| Quality Assurance | REQ-SK-801 to 804 | 4 |
| Package Management | REQ-SK-901 to 903 | 3 |
| **Total** | | **46** |

## Common Aspects

- `CLI` - Command-line interface
- `Core` - Core functionality
- `Templates` - Template files
- `Automation` - Scripts
- `Agent` - AI agent support
- `Testing` - Test files
- `Benchmarking` - Benchmark files
- `Documentation` - Docs
- `Quality` - Quality assurance
- `Constitution` - Constitutional framework

## Documentation Files

1. **[SPEC_KIT_INTEGRATION_SUMMARY.md](SPEC_KIT_INTEGRATION_SUMMARY.md)** - Overview
2. **[SPEC_KIT_REQUIREMENTS.md](SPEC_KIT_REQUIREMENTS.md)** - All 46 requirements
3. **[SPEC_KIT_GAP_ANALYSIS.md](SPEC_KIT_GAP_ANALYSIS.md)** - Tracking
4. **[SPEC_KIT_INTEGRATION_GUIDE.md](SPEC_KIT_INTEGRATION_GUIDE.md)** - Detailed guide
5. **[CANARY_EXAMPLES_SPEC_KIT.md](CANARY_EXAMPLES_SPEC_KIT.md)** - Examples
6. **[SPEC_KIT_QUICK_REFERENCE.md](SPEC_KIT_QUICK_REFERENCE.md)** - This file

## Sample Files with Tokens

- `specs/spec-kit/src/specify_cli/__init__.py` - 3 tokens
- `specs/spec-kit/scripts/bash/create-new-feature.sh` - 1 token
- `specs/spec-kit/templates/spec-template.md` - 1 token
- `specs/spec-kit/templates/plan-template.md` - 1 token
- `specs/spec-kit/templates/commands/specify.md` - 1 token

**Total**: 7 sample tokens demonstrating the integration

## Exit Codes

- `0` - Success
- `2` - Verification or staleness failed
- `3` - Parse or IO error

## Staleness Check

With `--strict`, canary fails if any TESTED/BENCHED token has UPDATED older than 30 days.

## CI/CD Integration

```yaml
- name: Verify Spec-Kit
  run: |
    ./canary verify --root ./specs/spec-kit --gap docs/SPEC_KIT_GAP_ANALYSIS.md --strict
```

## Common Tasks

### Add New Requirement
1. Add to `SPEC_KIT_REQUIREMENTS.md`
2. Add to `SPEC_KIT_GAP_ANALYSIS.md` with ❌
3. Add CANARY token to implementation
4. Run verification

### Update Token
1. Modify code
2. Update `UPDATED=` field to current date
3. Run scan to verify

### Link Tests
1. Create test file
2. Add CANARY token with `TEST=TestName`
3. Implementation auto-promotes to TESTED

### Link Benchmarks
1. Create benchmark file
2. Add CANARY token with `BENCH=BenchName`
3. Implementation auto-promotes to BENCHED

## Best Practices

1. ✅ Add tokens close to implementation
2. ✅ Use consistent FEATURE names
3. ✅ Update UPDATED field when modifying
4. ✅ Link actual test function names
5. ✅ Use separate tokens per aspect
6. ✅ Clear OWNER identification
7. ✅ Run scanner regularly

## Need Help?

- Full guide: [SPEC_KIT_INTEGRATION_GUIDE.md](SPEC_KIT_INTEGRATION_GUIDE.md)
- Examples: [CANARY_EXAMPLES_SPEC_KIT.md](CANARY_EXAMPLES_SPEC_KIT.md)
- Requirements: [SPEC_KIT_REQUIREMENTS.md](SPEC_KIT_REQUIREMENTS.md)
