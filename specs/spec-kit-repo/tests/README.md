# Spec-Kit Test Suite

This directory contains integration tests for all spec-kit requirements tracked with CANARY tokens.

## Overview

The test suite validates that spec-kit features are properly implemented and tracked:

- **27 test functions** across 4 test modules
- **20/43 requirements** (46.5%) have test coverage
- **100% coverage** for high-priority categories

## Test Structure

```
tests/
├── conftest.py                          # Pytest configuration
├── spec_kit/
│   ├── __init__.py
│   ├── test_workflow_commands.py        # 9 tests for REQ-SK-101 to 108
│   ├── test_cli_features.py             # 5 tests for REQ-SK-201 to 203
│   ├── test_automation_scripts.py       # 6 tests for REQ-SK-501 to 504
│   └── test_agent_support.py            # 7 tests for REQ-SK-601 to 605
└── README.md                            # This file
```

## Running Tests

### Prerequisites

Install test dependencies:

```bash
# From spec-kit root directory
pip install -e ".[test]"

# Or install pytest directly
pip install pytest pytest-cov
```

### Run All Tests

```bash
# From spec-kit root directory
pytest tests/

# Or with coverage report
pytest --cov=src --cov-report=html tests/
```

### Run Specific Test Modules

```bash
# Core workflow commands
pytest tests/spec_kit/test_workflow_commands.py

# CLI features
pytest tests/spec_kit/test_cli_features.py

# Automation scripts
pytest tests/spec_kit/test_automation_scripts.py

# Agent support
pytest tests/spec_kit/test_agent_support.py
```

### Run Specific Tests

```bash
# Run a single test function
pytest tests/spec_kit/test_workflow_commands.py::test_constitution_command_exists

# Run all tests matching a pattern
pytest -k "command_exists"
```

### Verbose Output

```bash
# Show detailed output
pytest -v tests/

# Show stdout/stderr
pytest -s tests/

# Both verbose and stdout
pytest -vs tests/
```

## Test Coverage by Category

| Category | Requirements | Tested | Coverage | Status |
|----------|--------------|--------|----------|---------|
| Core Workflow Commands | 8 | 8 | 100% | ✅ |
| CLI Tool Features | 3 | 2 | 67% | ✅ |
| Script Automation | 4 | 4 | 100% | ✅ |
| Agent Support | 5 | 5 | 100% | ✅ |
| Template System | 6 | 0 | 0% | ⚠️ |
| Constitutional Framework | 6 | 2 | 33% | ⚠️ |
| Documentation System | 4 | 0 | 0% | ⚠️ |
| Quality Assurance | 4 | 0 | 0% | ⚠️ |
| Package Management | 3 | 0 | 0% | ⚠️ |

## Test Philosophy

Following spec-kit's Constitutional Article IX (Integration-First Testing), these tests are:

1. **Integration Tests**: Validate real file existence and structure, not mocked
2. **CANARY-Aware**: Every test verifies proper CANARY token tracking
3. **Comprehensive**: Test both positive cases and meta-coverage
4. **Descriptive**: Clear names and docstrings for all test functions

## Writing New Tests

### Test Naming Convention

```python
# Test function name must match the TEST= field in CANARY token
# Example CANARY format (not a real token):
# CANARY: REQ=REQ-XXX-000; FEATURE="ExampleFeature"; ASPECT=Example; STATUS=IMPL; TEST=test_my_feature_exists; OWNER=tests; UPDATED=YYYY-MM-DD
def test_my_feature_exists():
    """Test that my feature exists and is properly configured."""
    # Test implementation
```

### Example Test Pattern

```python
def test_feature_exists():
    """Test that feature file exists and contains CANARY token."""
    # 1. Build path to feature file
    feature_path = Path(__file__).parent.parent.parent / "path" / "to" / "feature.ext"

    # 2. Verify file exists
    assert feature_path.exists(), f"Feature file not found at {feature_path}"

    # 3. Read and validate content
    content = feature_path.read_text()
    assert "CANARY:" in content, "File should contain CANARY token"
    assert "REQ-SK-XXX" in content, "File should track REQ-SK-XXX"

    # 4. Verify feature-specific requirements
    assert "expected_content" in content.lower(), "Feature should reference expected content"
```

### Linking Tests to Requirements

After creating a test, update the implementation file's CANARY token:

```python
# Before (example only - not a real token):
# CANARY_TOKEN: REQ=REQ-SK-XXX; FEATURE="MyFeature"; ASPECT=CLI; STATUS=IMPL; OWNER=team; UPDATED=2025-10-15

# After (add TEST= field - example only):
# CANARY_TOKEN: REQ=REQ-SK-XXX; FEATURE="MyFeature"; ASPECT=CLI; STATUS=IMPL; TEST=test_my_feature_exists; OWNER=team; UPDATED=2025-10-15
```

The CANARY scanner will detect the TEST= field and automatically promote STATUS from IMPL to TESTED!

## Continuous Integration

To run tests in CI/CD:

```yaml
# Example GitHub Actions workflow
- name: Run Tests
  run: |
    pip install -e ".[test]"
    pytest tests/ -v --cov=src --cov-report=xml

- name: Verify CANARY Status
  run: |
    go run . scan --json specs/spec-kit
    # Verify TESTED count matches expectations
```

## Next Steps

See `../docs/PHASE_3_COMPLETE.md` for:
- Complete testing report
- Phase 4 roadmap (remaining 22 requirements)
- Testing best practices

## Related Documentation

- [PHASE_3_COMPLETE.md](../docs/PHASE_3_COMPLETE.md) - Phase 3 completion report
- [SPEC_KIT_GAP_ANALYSIS.md](../../../docs/SPEC_KIT_GAP_ANALYSIS.md) - Requirements tracking
- [SPEC_KIT_REQUIREMENTS.md](../../../docs/SPEC_KIT_REQUIREMENTS.md) - Full requirements catalog
