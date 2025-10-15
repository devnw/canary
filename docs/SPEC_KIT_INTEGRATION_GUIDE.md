# Spec-Kit Integration Guide

This guide explains how to integrate the canary tracking system with the spec-kit submodule to achieve full feature visibility and requirements tracking.

## Overview

The spec-kit provides a comprehensive toolkit for Specification-Driven Development (SDD). By integrating canary tracking, we can:

1. **Track Implementation Status**: Monitor which spec-kit features are MISSING, STUB, IMPL, TESTED, or BENCHED
2. **Verify Claims**: Ensure documentation claims match actual implementation
3. **Detect Staleness**: Flag features that haven't been updated recently
4. **Coverage Analysis**: Measure test and benchmark coverage across all features
5. **Auto-Promotion**: Automatically promote status based on test/benchmark presence

## Integration Architecture

```
canary/
├── docs/
│   ├── SPEC_KIT_REQUIREMENTS.md       # All 46 spec-kit requirements
│   ├── SPEC_KIT_GAP_ANALYSIS.md       # Current tracking status
│   ├── CANARY_EXAMPLES_SPEC_KIT.md    # Token examples
│   └── SPEC_KIT_INTEGRATION_GUIDE.md  # This file
├── specs/
│   └── spec-kit/                       # Git submodule
│       ├── src/specify_cli/            # Add CANARY tokens here
│       ├── scripts/                    # Add CANARY tokens here
│       ├── templates/                  # Add CANARY tokens here
│       └── ...                         # Other spec-kit files
├── main.go                             # Canary scanner
├── scan.go                             # Token scanning logic
├── verify.go                           # Verification logic
└── status.json                         # Generated reports
```

## Requirements Catalog

The integration tracks **46 distinct features** across 10 categories:

### 1. Core Workflow Commands (REQ-SK-100 Series)
- Constitution, Specify, Clarify, Plan, Tasks, Implement, Analyze, Checklist commands

### 2. CLI Tool Features (REQ-SK-200 Series)
- Init, Check, Agent Detection capabilities

### 3. Template System (REQ-SK-300 Series)
- Spec, Plan, Tasks, Checklist, Constitution, Agent File templates

### 4. Constitutional Framework (REQ-SK-400 Series)
- Library-First, CLI Mandate, Test-First, Simplicity, Anti-Abstraction, Integration-First principles

### 5. Script Automation (REQ-SK-500 Series)
- Feature Creation, Plan Setup, Agent Context, Prerequisites Check scripts

### 6. Agent Support (REQ-SK-600 Series)
- Claude Code, Copilot, Gemini, Cursor, Multi-Agent support

### 7. Documentation System (REQ-SK-700 Series)
- Quickstart, Research, Data Model, API Contract generation

### 8. Quality Assurance (REQ-SK-800 Series)
- Ambiguity Detection, Consistency Validation, Coverage Analysis, Staleness Detection

### 9. Package Management (REQ-SK-900 Series)
- Release Packages, GitHub Release, Version Management

### 10. Additional Features
- Git Integration, Environment Variables, Error Handling, Extensibility

## Step-by-Step Integration

### Phase 1: Setup and Planning

1. **Review Requirements**
   ```bash
   cat docs/SPEC_KIT_REQUIREMENTS.md
   ```

2. **Check Current Status**
   ```bash
   cat docs/SPEC_KIT_GAP_ANALYSIS.md
   ```

3. **Study Token Examples**
   ```bash
   cat docs/CANARY_EXAMPLES_SPEC_KIT.md
   ```

### Phase 2: Add CANARY Tokens

#### Python Files (Specify CLI)

Add tokens to `specs/spec-kit/src/specify_cli/__init__.py`:

```python
# CANARY: REQ=REQ-SK-201; FEATURE="SpecifyCLIInit"; ASPECT=CLI; STATUS=IMPL; TEST=test_init_basic; OWNER=specify; UPDATED=2025-10-15
def init(...):
    """Bootstrap new project"""
    pass

# CANARY: REQ=REQ-SK-202; FEATURE="SpecifyCLICheck"; ASPECT=CLI; STATUS=IMPL; TEST=test_check_tools; OWNER=specify; UPDATED=2025-10-15
def check():
    """Verify installed tools"""
    pass
```

#### Bash Scripts

Add tokens to `specs/spec-kit/scripts/bash/*.sh`:

```bash
#!/usr/bin/env bash
# CANARY: REQ=REQ-SK-501; FEATURE="FeatureCreationScript"; ASPECT=Automation; STATUS=IMPL; OWNER=scripts; UPDATED=2025-10-15

create_feature() {
    # Implementation
}
```

#### Markdown Templates

Add tokens to `specs/spec-kit/templates/*.md`:

```markdown
<!-- CANARY: REQ=REQ-SK-301; FEATURE="SpecTemplate"; ASPECT=Templates; STATUS=IMPL; OWNER=templates; UPDATED=2025-10-15 -->

# Feature Specification: [FEATURE NAME]
...
```

#### Command Files

Add tokens to agent command files:

```markdown
<!-- .claude/commands/specify.md -->
<!-- CANARY: REQ=REQ-SK-102; FEATURE="SpecifyCommand"; ASPECT=CLI; STATUS=IMPL; OWNER=commands; UPDATED=2025-10-15 -->

---
description: "Create feature specification"
---
...
```

### Phase 3: Create Test Files

Add test files with CANARY markers:

```python
# tests/spec_kit/test_specify_command.py

# CANARY: REQ=REQ-SK-102; FEATURE="SpecifyCommand"; ASPECT=Testing; STATUS=TESTED; TEST=TestCANARY_REQ_SK_102_SpecifyBasic; OWNER=tests; UPDATED=2025-10-15

def test_specify_command_basic():
    """Test basic specify command execution"""
    result = run_specify_command("Build a photo album app")
    assert result.success
    assert "spec.md" in result.files_created
    assert len(result.user_stories) >= 3
```

### Phase 4: Add Benchmarks

Create benchmark files:

```python
# benchmarks/spec_kit/bench_init.py

# CANARY: REQ=REQ-SK-201; FEATURE="SpecifyCLIInit"; ASPECT=Benchmarking; STATUS=BENCHED; BENCH=BenchmarkCANARY_REQ_SK_201_Init; OWNER=benchmarks; UPDATED=2025-10-15

import time
import pytest

@pytest.mark.benchmark
def bench_init_performance(benchmark, tmpdir):
    """Benchmark project initialization"""
    result = benchmark(init_project, str(tmpdir), "claude")
    assert result.duration < 2.0  # Under 2 seconds
```

### Phase 5: Scan and Verify

1. **Run Scanner**
   ```bash
   ./canary --root ./specs/spec-kit --out spec-kit-status.json --csv spec-kit-status.csv
   ```

2. **Review Output**
   ```bash
   cat spec-kit-status.json | jq '.summary'
   ```

3. **Verify Against GAP**
   ```bash
   ./canary verify --root ./specs/spec-kit --gap docs/SPEC_KIT_GAP_ANALYSIS.md --strict
   ```

4. **Check for Staleness**
   ```bash
   ./canary --root ./specs/spec-kit --out status.json --strict
   # Exit 2 if any TESTED/BENCHED token is >30 days old
   ```

### Phase 6: Update Documentation

Update `docs/SPEC_KIT_GAP_ANALYSIS.md` as tokens are added:

```markdown
## Core Workflow Commands (REQ-SK-100 Series)

- ✅ REQ-SK-101: Constitution Command (`/speckit.constitution`)
- ✅ REQ-SK-102: Specify Command (`/speckit.specify`)
- ❌ REQ-SK-103: Clarify Command (`/speckit.clarify`)
...
```

## Token Placement Guidelines

### Where to Add Tokens

1. **Implementation Files**
   - Add to function/class definitions
   - Place at the top of the file or near key implementations
   - Use STATUS=IMPL for basic implementation

2. **Test Files**
   - Add to test file headers
   - Reference actual test function names in TEST= field
   - Use STATUS=TESTED for implementation + tests

3. **Template Files**
   - Add as comments at file header
   - Use STATUS=IMPL (templates are by nature "implemented")

4. **Script Files**
   - Add after shebang and before main logic
   - Use STATUS=IMPL or STATUS=TESTED if script tests exist

5. **Documentation Files**
   - Add as markdown comments
   - Track feature documentation completeness

### Token Fields

**Required Fields**:
- `REQ`: Requirement ID (REQ-SK-###)
- `FEATURE`: Short descriptive name (quoted if spaces)
- `ASPECT`: Category (CLI, Testing, Templates, etc.)
- `STATUS`: Current status (MISSING/STUB/IMPL/TESTED/BENCHED/REMOVED)

**Optional Fields**:
- `TEST`: Comma-separated list of test function names
- `BENCH`: Comma-separated list of benchmark function names
- `OWNER`: Team/area responsible (specify, scripts, tests, etc.)
- `UPDATED`: Last update date (YYYY-MM-DD format)

### Status Progression

```
MISSING → STUB → IMPL → TESTED → BENCHED
                    ↓
                REMOVED
```

**Auto-Promotion Rules**:
1. IMPL + TEST= field → Auto-promotes to TESTED
2. (IMPL or TESTED) + BENCH= field → Auto-promotes to BENCHED

## File Organization

### Recommended Structure

```
canary/
├── docs/                               # Integration documentation
│   ├── SPEC_KIT_REQUIREMENTS.md
│   ├── SPEC_KIT_GAP_ANALYSIS.md
│   ├── CANARY_EXAMPLES_SPEC_KIT.md
│   └── SPEC_KIT_INTEGRATION_GUIDE.md
│
├── specs/spec-kit/                     # Spec-kit submodule
│   ├── src/specify_cli/                # Add tokens to Python files
│   ├── scripts/bash/                   # Add tokens to bash scripts
│   ├── scripts/powershell/             # Add tokens to PS scripts
│   ├── templates/                      # Add tokens to templates
│   └── ...
│
├── tests/spec_kit/                     # Spec-kit tests
│   ├── test_constitution.py            # Constitution tests
│   ├── test_specify.py                 # Specify command tests
│   ├── test_plan.py                    # Plan command tests
│   └── ...
│
└── benchmarks/spec_kit/                # Spec-kit benchmarks
    ├── bench_init.py
    ├── bench_specify.py
    └── ...
```

## Verification Workflow

### CI/CD Integration

Add to `.github/workflows/verify-spec-kit.yml`:

```yaml
name: Verify Spec-Kit Tracking

on: [push, pull_request]

jobs:
  verify:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
        with:
          submodules: true  # Important: fetch spec-kit submodule

      - name: Build Canary
        run: go build -o canary .

      - name: Scan Spec-Kit
        run: |
          ./canary --root ./specs/spec-kit --out spec-kit-status.json --csv spec-kit-status.csv

      - name: Verify GAP Analysis
        run: |
          ./canary verify --root ./specs/spec-kit --gap docs/SPEC_KIT_GAP_ANALYSIS.md --strict

      - name: Check Staleness
        run: |
          ./canary --root ./specs/spec-kit --out status.json --strict

      - name: Upload Report
        uses: actions/upload-artifact@v3
        with:
          name: spec-kit-status
          path: |
            spec-kit-status.json
            spec-kit-status.csv
```

### Local Development Workflow

```bash
# 1. Make changes to spec-kit files
cd specs/spec-kit
# ... edit files, add CANARY tokens ...

# 2. Run local scan
cd ../..
./canary --root ./specs/spec-kit --out status.json

# 3. Check results
cat status.json | jq '.summary'

# 4. Verify against GAP
./canary verify --root ./specs/spec-kit --gap docs/SPEC_KIT_GAP_ANALYSIS.md

# 5. Update GAP analysis if needed
vi docs/SPEC_KIT_GAP_ANALYSIS.md

# 6. Commit changes
git add specs/spec-kit docs/
git commit -m "feat: add CANARY tracking for REQ-SK-###"
```

## Reporting and Metrics

### Status Report

The scanner generates comprehensive reports:

```json
{
  "generated_at": "2025-10-15T10:30:00Z",
  "requirements": [
    {
      "id": "REQ-SK-101",
      "features": [
        {
          "feature": "ConstitutionCommand",
          "aspect": "CLI",
          "status": "TESTED",
          "files": ["src/specify_cli/commands/constitution.py"],
          "tests": ["test_constitution_basic", "test_constitution_validation"],
          "benches": [],
          "owner": "specify",
          "updated": "2025-10-15"
        }
      ]
    }
  ],
  "summary": {
    "by_status": {
      "TESTED": 15,
      "IMPL": 20,
      "STUB": 8,
      "MISSING": 3
    },
    "by_aspect": {
      "CLI": 12,
      "Testing": 10,
      "Templates": 8,
      "Automation": 6,
      "Quality": 5
    }
  }
}
```

### CSV Export

```csv
req,feature,aspect,status,file,test,bench,owner,updated
REQ-SK-101,ConstitutionCommand,CLI,TESTED,src/specify_cli/commands/constitution.py,test_constitution_basic,,specify,2025-10-15
REQ-SK-101,ConstitutionCommand,CLI,TESTED,src/specify_cli/commands/constitution.py,test_constitution_validation,,specify,2025-10-15
REQ-SK-102,SpecifyCommand,CLI,TESTED,src/specify_cli/commands/specify.py,test_specify_basic,,specify,2025-10-15
```

### Coverage Metrics

Track progress over time:

```bash
# Total requirements
TOTAL=46

# Count tracked requirements
TRACKED=$(grep -c "✅" docs/SPEC_KIT_GAP_ANALYSIS.md)

# Calculate coverage
COVERAGE=$((TRACKED * 100 / TOTAL))

echo "Coverage: $COVERAGE% ($TRACKED/$TOTAL)"
```

## Maintenance

### Regular Updates

1. **Weekly**: Scan for staleness
   ```bash
   ./canary --root ./specs/spec-kit --strict
   ```

2. **Monthly**: Review coverage
   ```bash
   ./canary --root ./specs/spec-kit --out report.json
   cat report.json | jq '.summary'
   ```

3. **Per Release**: Update all UPDATED timestamps
   ```bash
   # Update all CANARY tokens with current date
   find specs/spec-kit -type f -exec sed -i 's/UPDATED=20[0-9][0-9]-[0-9][0-9]-[0-9][0-9]/UPDATED=2025-10-15/g' {} \;
   ```

### Adding New Requirements

When adding new spec-kit features:

1. **Add to SPEC_KIT_REQUIREMENTS.md**
   ```markdown
   ### REQ-SK-###: New Feature Name
   **Feature**: Feature description
   **Aspect**: Category
   ...
   ```

2. **Add to SPEC_KIT_GAP_ANALYSIS.md**
   ```markdown
   - ❌ REQ-SK-###: New Feature Name
   ```

3. **Add CANARY tokens to implementation**
   ```python
   # CANARY: REQ=REQ-SK-###; FEATURE="NewFeature"; ASPECT=CLI; STATUS=IMPL; OWNER=team; UPDATED=2025-10-15
   ```

4. **Run verification**
   ```bash
   ./canary verify --root ./specs/spec-kit --gap docs/SPEC_KIT_GAP_ANALYSIS.md
   ```

## Best Practices

1. **Token Placement**: Add tokens as close to implementation as possible
2. **Consistent Naming**: Use consistent FEATURE names across files
3. **Update Timestamps**: Always update UPDATED field when modifying code
4. **Link Tests**: Always reference actual test function names in TEST= field
5. **Track Aspects**: Use separate tokens for different aspects (CLI, Testing, etc.)
6. **Owner Clarity**: Use clear OWNER field to identify responsible area
7. **Regular Scans**: Run scanner regularly to catch staleness
8. **Document Changes**: Update GAP analysis when adding/removing tokens

## Troubleshooting

### Token Not Detected

**Problem**: Added CANARY token but not showing in scan results

**Solutions**:
1. Check token format matches regex: `^\s*(?://|#|--)\s*CANARY:\s*(.*)$`
2. Ensure file is not in skip dirs (node_modules, .git, bin, etc.)
3. Verify file is not detected as binary
4. Check for proper key=value format with semicolon separators

### Verification Failure

**Problem**: `canary verify` fails with missing requirements

**Solutions**:
1. Check GAP analysis has all requirements listed
2. Ensure requirement IDs match exactly (REQ-SK-### format)
3. Verify at least one CANARY token exists per requirement
4. Check for typos in requirement IDs

### Staleness Failure

**Problem**: `--strict` mode fails with stale UPDATED dates

**Solutions**:
1. Update UPDATED field in affected CANARY tokens
2. Use bulk update command if many tokens are stale
3. Review if feature is truly maintained (consider STATUS=REMOVED)
4. Adjust staleness threshold if needed

## Summary

This integration provides:

✅ **Complete Feature Tracking**: All 46 spec-kit features tracked with CANARY tokens
✅ **Test Coverage Visibility**: Automatic promotion based on test presence
✅ **Staleness Detection**: Flag features not updated in 30+ days
✅ **Verification**: Ensure GAP analysis matches actual implementation
✅ **CI/CD Integration**: Automated checks in build pipeline
✅ **Comprehensive Reporting**: JSON and CSV output for analysis

Next steps:
1. Add CANARY tokens to all spec-kit source files
2. Create test files with CANARY markers
3. Add benchmarks where applicable
4. Set up CI/CD verification
5. Monitor coverage and staleness regularly
