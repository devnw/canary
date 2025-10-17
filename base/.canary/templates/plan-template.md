<!-- CANARY: REQ=CBIN-Docs-116; FEATURE="PlanTemplate"; ASPECT=Docs; STATUS=IMPL; OWNER=canary; UPDATED=2025-10-16 -->
# Implementation Plan: {{.ReqID}}-<ASPECT>-XXX [FEATURE NAME]

**Requirement:** {{.ReqID}}-<ASPECT>-XXX (e.g., {{.ReqID}}-CLI-001, {{.ReqID}}-API-042)
**Specification:** [Link to spec.md]
**Status:** STUB → IMPL
**Created:** YYYY-MM-DD
**Updated:** YYYY-MM-DD

## Tech Stack Decision

### Primary Technologies
- **Language:** [Go/Python/JavaScript/etc.] [version]
- **Framework:** [Standard library/framework name] [version]
- **Database:** [PostgreSQL/MySQL/etc.] (if needed) [version]
- **Testing:** [test framework] [version]

### Rationale
[Why these technologies were chosen]
- [Reason 1: e.g., Performance requirements]
- [Reason 2: e.g., Team expertise]
- [Reason 3: e.g., Existing stack consistency]

## CANARY Token Placement

### Token Definition
```go
// File: [path/to/file.go]
// CANARY: REQ={{.ReqID}}-<ASPECT>-XXX; FEATURE="FeatureName"; ASPECT=API; STATUS=IMPL; OWNER=team; UPDATED=YYYY-MM-DD
// Example: REQ={{.ReqID}}-API-105; FEATURE="UserAuth"; ASPECT=API; STATUS=IMPL; OWNER=team; UPDATED=2025-10-16

package [package]

// [Function/Type that implements this requirement]
```

### File Structure
```
project/
├── [package]/
│   ├── [feature].go          # Main implementation
│   ├── [feature]_test.go     # Unit tests
│   └── [feature]_bench_test.go # Benchmarks (if perf-critical)
```

## Architecture Overview

### Component Diagram
```
[High-level component structure]
Example:
- UserInput → Validator → Processor → Storage
- External API → Handler → Business Logic → Data Layer
```

### Key Components

**Component 1: [Name]**
- **Responsibility:** [What it does]
- **Interfaces:** [Exported functions/types]
- **Dependencies:** [What it depends on]

**Component 2: [Name]**
- **Responsibility:** [What it does]
- **Interfaces:** [Exported functions/types]
- **Dependencies:** [What it depends on]

## Implementation Phases

### Phase 0: Pre-Implementation Gates

**Simplicity Gate (Constitution Article V):**
- [ ] Using standard library where possible
- [ ] Minimal dependencies
- [ ] No premature optimization
- [ ] No speculative features

**Anti-Abstraction Gate (Constitution Article V):**
- [ ] Using frameworks directly (no wrapper layers)
- [ ] Single representation of concepts
- [ ] No unnecessary interfaces

**Test-First Gate (Constitution Article IV):**
- [ ] Test strategy defined
- [ ] Test functions named
- [ ] Tests will be written before implementation

**Integration-First Gate (Constitution Article VI):**
- [ ] Real environment testing planned
- [ ] Contract tests defined (if API)
- [ ] No mock-heavy testing

### Phase 1: Test Creation (Red Phase)

**Step 1.1: Create test file**
```go
// File: [package]/[feature]_test.go
package [package]

import "testing"

func TestCANARY_REQ_[Aspect]_XXX_[FeatureName](t *testing.T) {
    // Example: TestCANARY_REQ_API_105_UserAuth(t *testing.T)
    // Test implementation
    // Expected to FAIL initially
}
```

**Step 1.2: Update CANARY token**
```
// Add TEST= field
TEST=TestCANARY_REQ_[Aspect]_XXX_[FeatureName]
// Example: TEST=TestCANARY_REQ_API_105_UserAuth
```

**Step 1.3: Verify test fails**
- [ ] Run `go test ./[package]`
- [ ] Confirm test fails with expected error
- [ ] Document expected failure message

### Phase 2: Implementation (Green Phase)

**Step 2.1: Implement minimum to pass tests**
```go
// File: [package]/[feature].go
package [package]

// [Implementation that makes tests pass]
```

**Step 2.2: Update CANARY token status**
```
// Update STATUS field
STATUS=TESTED
```

**Step 2.3: Verify tests pass**
- [ ] Run `go test ./[package]`
- [ ] All tests pass
- [ ] No regressions in other tests

### Phase 3: Benchmarking (if performance-critical)

**Step 3.1: Create benchmark**
```go
// File: [package]/[feature]_bench_test.go
package [package]

import "testing"

func BenchmarkCANARY_REQ_[Aspect]_XXX_[FeatureName](b *testing.B) {
    // Example: BenchmarkCANARY_REQ_API_105_UserAuth(b *testing.B)
    // Benchmark implementation
}
```

**Step 3.2: Update CANARY token**
```
// Add BENCH= field
BENCH=BenchmarkCANARY_REQ_[Aspect]_XXX_[FeatureName]
// Example: BENCH=BenchmarkCANARY_REQ_API_105_UserAuth
// Update STATUS
STATUS=BENCHED
```

**Step 3.3: Document baseline**
- [ ] Run `go test -bench=. ./[package]`
- [ ] Document baseline performance (ns/op, B/op, allocs/op)
- [ ] Set performance regression threshold

## Testing Strategy

### Unit Tests
**Test:** `TestCANARY_REQ_[Aspect]_XXX_[FeatureName]`
**Example:** `TestCANARY_REQ_API_105_UserAuth`
**Coverage:** [What functionality is tested]
**Test Cases:**
- [ ] Happy path
- [ ] Edge cases
- [ ] Error conditions

### Integration Tests (if applicable)
**Test:** `TestIntegration_[Feature]`
**Coverage:** [End-to-end scenario]
**Environment:** [Real database/service/file system]

### Acceptance Tests
**Based on spec success criteria:**
- [ ] [Criterion 1 from spec]
- [ ] [Criterion 2 from spec]
- [ ] [Criterion 3 from spec]

### Performance Benchmarks (if applicable)
**Benchmark:** `BenchmarkCANARY_REQ_[Aspect]_XXX_[FeatureName]`
**Example:** `BenchmarkCANARY_REQ_API_105_UserAuth`
**Baseline:** [target performance]
**Metrics:** [ns/op, memory, allocations]

## Constitutional Compliance

### Article I: Requirement-First Development
- ✅ CANARY token defined
- ✅ Token placed in appropriate file
- ✅ Token includes all required fields

### Article IV: Test-First Imperative
- ✅ Tests written before implementation
- ✅ Tests fail initially (red phase)
- ✅ Implementation makes tests pass (green phase)

### Article V: Simplicity and Anti-Abstraction
- ✅ Using standard library/minimal dependencies
- ✅ No unnecessary abstractions
- ✅ Framework features used directly

### Article VI: Integration-First Testing
- ✅ Real environment testing
- ✅ Contract tests for APIs
- ✅ Minimal mocking

### Article VII: Documentation Currency
- ✅ CANARY token includes OWNER
- ✅ UPDATED field will be maintained
- ✅ Status progresses with evidence (STUB→IMPL→TESTED→BENCHED)

## Complexity Tracking

### Justified Complexity
[Document any complexity that violates simplicity principles]

**Exception:** [Description of complex aspect]
**Justification:** [Why complexity is necessary]
**Constitutional Article:** [Which article this relates to]
**Mitigation:** [How complexity is managed]

### Dependencies Added
- [Dependency name] [version] - [Justification]
- [Dependency name] [version] - [Justification]

## Implementation Checklist

- [ ] Phase 0 gates all passed
- [ ] Test file created
- [ ] Tests fail initially (red)
- [ ] Implementation created
- [ ] Tests pass (green)
- [ ] CANARY token updated with TEST= field
- [ ] Token STATUS updated to TESTED
- [ ] Benchmark created (if perf-critical)
- [ ] Token updated with BENCH= field
- [ ] Token STATUS updated to BENCHED
- [ ] All acceptance criteria met
- [ ] Constitutional compliance verified
- [ ] Ready for code review

---

## Next Steps

1. Review this plan for accuracy and completeness
2. Use `/canary.implement` to execute implementation
3. Run `/canary.scan` after implementation to verify token status
4. Run `/canary.verify` to confirm claim in GAP_ANALYSIS.md

---

## CLI Commands for This Implementation

**Start implementation:**
```bash
canary implement {{.ReqID}} --prompt
```

**During development:**
```bash
# Check progress
canary status {{.ReqID}}
canary show {{.ReqID}} --group-by status

# List implementation files
canary files {{.ReqID}}

# Create CANARY tokens as you implement
canary create {{.ReqID}} "FeatureName" --aspect <ASPECT> --status IMPL
```

**Track mistakes (gap analysis):**
```bash
canary gap mark {{.ReqID}} <feature> \
  --category logic_error|test_failure|performance|edge_case \
  --description "what went wrong" \
  --action "corrective action taken"
```

**Verify implementation:**
```bash
# Scan for all tokens
canary scan --root . --out status.json --csv status.csv

# Verify with strict staleness checks
canary scan --strict

# Verify GAP_ANALYSIS claims
canary scan --verify GAP_ANALYSIS.md

# Update stale tokens
canary scan --update-stale
```

**Documentation:**
```bash
# Create documentation from template
canary doc create {{.ReqID}} --type technical --output docs/<feature>.md

# Update documentation hash after editing
canary doc update {{.ReqID}}

# Check documentation status
canary doc status {{.ReqID}}
```

**After completion:**
```bash
# Create checkpoint
canary checkpoint "{{.ReqID}}-complete" "Completed {{.ReqID}} implementation"

# Find next work
canary next --prompt
canary list --status STUB --limit 10
```
