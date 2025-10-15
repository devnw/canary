# CANARY Token Examples for Spec-Kit Integration

This document provides examples of how to add CANARY tokens throughout the spec-kit codebase to track feature implementation.

## Token Format

```
CANARY: REQ=REQ-SK-###; FEATURE="Name"; ASPECT=...; STATUS=MISSING|STUB|IMPL|TESTED|BENCHED|REMOVED; TEST=...; BENCH=...; OWNER=...; UPDATED=YYYY-MM-DD
```

## Status Definitions

- **MISSING**: Feature not yet implemented
- **STUB**: Placeholder/stub implementation exists
- **IMPL**: Full implementation without tests
- **TESTED**: Implementation with passing tests (auto-promoted from IMPL if TEST= present)
- **BENCHED**: Implementation with benchmarks (auto-promoted if BENCH= present)
- **REMOVED**: Feature removed/deprecated

## Python Examples (Specify CLI)

### Example 1: Constitution Command

```python
# In src/specify_cli/__init__.py or command file

# CANARY: REQ=REQ-SK-101; FEATURE="ConstitutionCommand"; ASPECT=CLI; STATUS=IMPL; TEST=test_constitution_command; OWNER=specify; UPDATED=2025-10-15
def constitution_command(args: str):
    """
    Create or update project governing principles.
    Implements /speckit.constitution command.
    """
    # Implementation here
    pass
```

### Example 2: Specify Command with Test

```python
# In src/specify_cli/commands/specify.py

# CANARY: REQ=REQ-SK-102; FEATURE="SpecifyCommand"; ASPECT=CLI; STATUS=TESTED; TEST=test_specify_command_basic,test_specify_command_clarification; OWNER=specify; UPDATED=2025-10-15
def specify_command(description: str):
    """
    Define feature requirements from user description.
    Implements /speckit.specify command.
    """
    # Implementation here
    pass
```

### Example 3: CLI Init with Benchmark

```python
# In src/specify_cli/__init__.py

# CANARY: REQ=REQ-SK-201; FEATURE="SpecifyCLIInit"; ASPECT=CLI; STATUS=BENCHED; TEST=test_init_new_project,test_init_existing_dir; BENCH=bench_init_performance; OWNER=specify; UPDATED=2025-10-15
def init(project_name: str, ai_assistant: str = None):
    """
    Bootstrap new project with spec-kit framework.
    """
    # Implementation here
    pass
```

## Bash Script Examples

### Example 4: Feature Creation Script

```bash
#!/usr/bin/env bash
# scripts/bash/create-new-feature.sh

# CANARY: REQ=REQ-SK-501; FEATURE="FeatureCreationScript"; ASPECT=Automation; STATUS=TESTED; TEST=test_create_feature_script; OWNER=scripts; UPDATED=2025-10-15

set -e

create_feature() {
    # Implementation here
    :
}
```

### Example 5: Agent Context Update

```bash
#!/usr/bin/env bash
# scripts/bash/update-agent-context.sh

# CANARY: REQ=REQ-SK-503; FEATURE="AgentContextUpdate"; ASPECT=Automation; STATUS=IMPL; OWNER=scripts; UPDATED=2025-10-15

update_agent_file() {
    # Implementation here
    :
}
```

## Markdown Template Examples

### Example 6: Spec Template

```markdown
<!-- templates/spec-template.md -->

<!-- CANARY: REQ=REQ-SK-301; FEATURE="SpecTemplate"; ASPECT=Templates; STATUS=IMPL; OWNER=templates; UPDATED=2025-10-15 -->

# Feature Specification: [FEATURE NAME]

...
```

### Example 7: Plan Template

```markdown
<!-- templates/plan-template.md -->

<!-- CANARY: REQ=REQ-SK-302; FEATURE="PlanTemplate"; ASPECT=Templates; STATUS=IMPL; OWNER=templates; UPDATED=2025-10-15 -->

# Implementation Plan: [FEATURE]

...
```

## TOML Examples (Gemini Commands)

### Example 8: Gemini Specify Command

```toml
# .gemini/commands/specify.toml

# CANARY: REQ=REQ-SK-602; FEATURE="GeminiSupport"; ASPECT=Agent; STATUS=IMPL; OWNER=agents; UPDATED=2025-10-15

description = "Create feature specification"

prompt = """
...
"""
```

## Multi-File Feature Tracking

### Example 9: Constitutional Framework

```python
# In constitution validator file

# CANARY: REQ=REQ-SK-401; FEATURE="LibraryFirstPrinciple"; ASPECT=Constitution; STATUS=TESTED; TEST=test_library_first_validation; OWNER=constitution; UPDATED=2025-10-15
def validate_library_first(plan):
    """Enforce Article I: Library-First Principle"""
    pass

# CANARY: REQ=REQ-SK-402; FEATURE="CLIInterfaceMandate"; ASPECT=Constitution; STATUS=TESTED; TEST=test_cli_interface_validation; OWNER=constitution; UPDATED=2025-10-15
def validate_cli_interface(library):
    """Enforce Article II: CLI Interface Mandate"""
    pass

# CANARY: REQ=REQ-SK-403; FEATURE="TestFirstImperative"; ASPECT=Constitution; STATUS=TESTED; TEST=test_tdd_enforcement; OWNER=constitution; UPDATED=2025-10-15
def validate_test_first(implementation):
    """Enforce Article III: Test-First Imperative"""
    pass
```

## Test File Examples

### Example 10: Constitution Command Tests

```python
# tests/test_constitution_command.py

# CANARY: REQ=REQ-SK-101; FEATURE="ConstitutionCommand"; ASPECT=Testing; STATUS=TESTED; TEST=TestCANARY_REQ_SK_101_ConstitutionBasic; OWNER=tests; UPDATED=2025-10-15

import pytest

def test_constitution_command_basic():
    """Test basic constitution command execution"""
    result = constitution_command("Create principles for code quality")
    assert result.success
    assert "constitution.md" in result.files_created
```

### Example 11: Agent Detection Tests

```python
# tests/test_agent_detection.py

# CANARY: REQ=REQ-SK-203; FEATURE="AgentDetection"; ASPECT=Testing; STATUS=TESTED; TEST=TestCANARY_REQ_SK_203_AgentDetection; OWNER=tests; UPDATED=2025-10-15

def test_agent_detection_claude():
    """Test Claude Code agent detection"""
    result = detect_agent(".claude/")
    assert result.agent_type == "claude"
    assert result.detected == True

def test_agent_detection_multi():
    """Test multiple agent detection"""
    result = detect_all_agents(project_root)
    assert "claude" in result.agents
    assert "gemini" in result.agents
```

## Benchmark Examples

### Example 12: Init Command Benchmark

```python
# benchmarks/bench_init.py

# CANARY: REQ=REQ-SK-201; FEATURE="SpecifyCLIInit"; ASPECT=Benchmarking; STATUS=BENCHED; BENCH=BenchmarkCANARY_REQ_SK_201_InitPerformance; OWNER=benchmarks; UPDATED=2025-10-15

import time

def bench_init_performance(tmpdir):
    """Benchmark project initialization performance"""
    start = time.time()
    init(str(tmpdir / "test-project"), ai_assistant="claude")
    duration = time.time() - start
    assert duration < 2.0  # Should complete in under 2 seconds
```

## Integration Examples

### Example 13: Multi-Aspect Feature

```python
# In various files tracking the same requirement

# File: src/specify_cli/commands/plan.py
# CANARY: REQ=REQ-SK-104; FEATURE="PlanCommand"; ASPECT=CLI; STATUS=IMPL; OWNER=commands; UPDATED=2025-10-15

# File: templates/plan-template.md
# CANARY: REQ=REQ-SK-104; FEATURE="PlanCommand"; ASPECT=Templates; STATUS=IMPL; OWNER=templates; UPDATED=2025-10-15

# File: tests/test_plan_command.py
# CANARY: REQ=REQ-SK-104; FEATURE="PlanCommand"; ASPECT=Testing; STATUS=TESTED; TEST=test_plan_generation; OWNER=tests; UPDATED=2025-10-15

# File: docs/plan-command.md
# CANARY: REQ=REQ-SK-104; FEATURE="PlanCommand"; ASPECT=Documentation; STATUS=IMPL; OWNER=docs; UPDATED=2025-10-15
```

This creates multiple tracking points for the same requirement across different aspects:
- CLI implementation
- Template definition
- Test coverage
- Documentation

## Quality Assurance Examples

### Example 14: Ambiguity Detection

```python
# In quality module

# CANARY: REQ=REQ-SK-801; FEATURE="AmbiguityDetection"; ASPECT=Quality; STATUS=TESTED; TEST=test_ambiguity_detection; BENCH=bench_ambiguity_scan; OWNER=quality; UPDATED=2025-10-15
def detect_ambiguities(spec_text: str) -> List[Ambiguity]:
    """
    Scan specification for ambiguous requirements.
    Returns list of detected ambiguities with suggestions.
    """
    pass
```

### Example 15: Coverage Analysis

```python
# In coverage module

# CANARY: REQ=REQ-SK-803; FEATURE="CoverageAnalysis"; ASPECT=Quality; STATUS=TESTED; TEST=test_coverage_analysis; OWNER=quality; UPDATED=2025-10-15
def analyze_feature_coverage(spec_dir: Path) -> CoverageReport:
    """
    Analyze test and implementation coverage for feature.
    """
    pass
```

## Best Practices

1. **Place tokens near the implementation**: Add CANARY tokens as close to the actual implementation as possible
2. **One token per aspect**: Use separate tokens for different aspects (CLI, Testing, Benchmarking, etc.)
3. **Update timestamps**: Always update UPDATED when modifying tracked code
4. **Link tests explicitly**: Use TEST= field to reference actual test function names
5. **Group by requirement**: Keep all aspects of the same REQ-SK-### requirement visible in reports
6. **Owner tracking**: Use OWNER field to identify responsible team/area

## Scanning Examples

### Scan spec-kit directory

```bash
# From canary project root
./canary --root ./specs/spec-kit --out spec-kit-status.json --csv spec-kit-status.csv
```

### Verify against gap analysis

```bash
./canary verify --root ./specs/spec-kit --gap docs/SPEC_KIT_GAP_ANALYSIS.md --strict
```

### Check for stale markers

```bash
# Fail if any TESTED/BENCHED tokens have UPDATED older than 30 days
./canary --root ./specs/spec-kit --out status.json --strict
```

## Integration Workflow

1. **Add tokens**: Insert CANARY tokens throughout spec-kit codebase
2. **Run scan**: Execute canary scanner to generate status.json
3. **Update GAP**: Mark requirements as ✅ in gap analysis
4. **Verify**: Run `canary verify` to ensure all claims are tracked
5. **Monitor**: Set up CI/CD to fail on staleness or missing tokens

## Summary

These examples demonstrate:
- Token placement in various file types (Python, Bash, Markdown, TOML)
- Different status levels (IMPL, TESTED, BENCHED)
- Multi-aspect tracking for comprehensive requirements
- Test and benchmark linking
- Owner and timestamp tracking
- Integration with canary verification system

Apply these patterns throughout the spec-kit submodule to achieve full feature tracking.
