<!-- CANARY: REQ=CBIN-148; FEATURE="InstructionTemplates"; ASPECT=Docs; STATUS=BENCHED; TEST=TestCopilotInstructionTemplateValidity; BENCH=BenchmarkCreateCopilotInstructions; UPDATED=2025-10-19 -->

# CANARY Test-First Development Guidelines

You are working in a test directory. This project follows **strict Test-Driven Development (TDD)**.

## Article IV: Test-First Imperative (NON-NEGOTIABLE)

**No implementation code shall be written before:**

1. Test functions are written and named in token TEST= field
2. Tests are confirmed to FAIL (Red phase)
3. Implementation makes tests PASS (Green phase)

## TDD Workflow

### 1. Write the Test FIRST (Red Phase)

```go
// CANARY: REQ={{.ProjectKey}}-###; FEATURE="FeatureName"; ASPECT=API; STATUS=STUB; TEST=TestFeatureName; UPDATED=YYYY-MM-DD
func TestFeatureName(t *testing.T) {
    // Arrange
    input := "test data"

    // Act
    result, err := FeatureFunction(input)

    // Assert
    if err != nil {
        t.Fatalf("FeatureFunction failed: %v", err)
    }
    if result != expected {
        t.Errorf("got %v, want %v", result, expected)
    }
}
```

### 2. Run Test - Confirm it FAILS

```bash
go test -v -run TestFeatureName
# Expected: FAIL (function doesn't exist yet)
```

### 3. Implement Feature (Green Phase)

Write minimal code to make the test pass.

### 4. Update Token Status

```go
// CANARY: REQ={{.ProjectKey}}-###; FEATURE="FeatureName"; ASPECT=API; STATUS=TESTED; TEST=TestFeatureName; UPDATED=YYYY-MM-DD
```

## Integration Testing (Article VI)

**Prefer real environments over mocks:**

- Use real databases (not mocks)
- Use actual service instances
- Test actual file I/O operations
- Use `t.TempDir()` for filesystem tests

## Test Coverage Requirements

- Unit tests for all core functionality
- Integration tests for end-to-end flows
- Benchmarks for performance-critical code (BENCH= field)

## Complete TDD Workflow

**Read `.canary/commands/implement.md` for the complete test-first workflow**, including:
- RED-GREEN-REFACTOR cycle details
- Constitutional compliance validation
- Status progression rules
- Implementation guidance

**Read `.canary/memory/constitution.md` for Articles IV and VI** covering:
- Test-First Imperative (Article IV)
- Integration-First Testing (Article VI)

## Verification

**Use `/canary.scan` to verify test coverage:**
- Checks for TEST= fields on TESTED status tokens
- Identifies missing tests
- Validates test-first compliance

## Related Files

- `.canary/commands/implement.md` - Implementation workflow with TDD
- `.canary/commands/scan.md` - Test coverage verification
- `.canary/memory/constitution.md` - Articles IV & VI
