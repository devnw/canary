# Implementation Guidance: {{.Feature}}
<!-- CANARY: REQ=CBIN-CLI-001; FEATURE="NextTemplate"; ASPECT=Docs; STATUS=IMPL; UPDATED=2025-10-16 -->

**Requirement:** {{.ReqID}}
**Feature:** {{.Feature}}
**Aspect:** {{.Aspect}}
**Current Status:** {{.Status}}
**Priority:** {{.Priority}} (1=highest, 10=lowest)

---

## Constitutional Guidance

Before implementing, review these governing principles from `.canary/memory/constitution.md`:

{{.Constitution}}

**Key Principles for This Implementation:**
1. **Article I: Requirement-First** - This CANARY token represents a requirement that MUST be tracked
2. **Article IV: Test-First Imperative** - Write tests BEFORE implementation
3. **Article V: Simplicity** - Prefer standard library, avoid unnecessary complexity
4. **Article VII: Documentation Currency** - Update UPDATED field when modifying

---

## Specification

**Source:** {{.SpecFile}}

{{.SpecContent}}

---

## Dependencies and Context

{{if .Dependencies}}
### Prerequisites (DEPENDS_ON)
The following requirements must be completed before implementing {{.ReqID}}:
{{range .Dependencies}}
- **{{.ReqID}}**: {{.Feature}} ({{.Status}}){{if ne .Status "TESTED"}}{{if ne .Status "BENCHED"}} ⚠️ NOT COMPLETE{{end}}{{end}}
{{end}}

**Action Required:** Ensure all dependencies are at TESTED or BENCHED status before proceeding.
{{else}}
### Prerequisites
✅ No blocking dependencies - safe to proceed with implementation.
{{end}}

{{if .RelatedSpecs}}
### Related Specifications
Review these specifications for context and design consistency:
{{range .RelatedSpecs}}
- **{{.ReqID}}**: {{.Feature}} - [spec file]({{.SpecFile}})
{{end}}
{{end}}

---

## Implementation Approach

### Step 1: Write Tests First (Article IV)

Following test-first principles, create tests that verify:
{{.TestGuidance}}

**Example Test Structure:**
```go
// File: {{.SuggestedTestFile}}
package {{.PackageName}}_test

import "testing"

func TestCANARY_{{.ReqID}}_{{.Aspect}}_{{.Feature}}(t *testing.T) {
    // Setup: Create test fixtures

    // Execute: Call the function being tested

    // Verify: Assert expected outcomes match specification
}
```

### Step 2: Place CANARY Token

Add the CANARY token at the implementation site:

{{.TokenExample}}

**Important:** Update STATUS as implementation progresses:
- `STATUS=STUB` → Initial placeholder
- `STATUS=IMPL` → Implementation complete, no tests
- `STATUS=TESTED` → Tests added and passing (add `TEST=TestName`)
- `STATUS=BENCHED` → Benchmarks added (add `BENCH=BenchName`)

### Step 3: Implement Functionality

Focus on:
- ✅ Meeting specification requirements
- ✅ Maintaining simplicity (Article V)
- ✅ Using standard library when possible
- ✅ Adding clear error messages
- ✅ Following existing code patterns

**Avoid:**
- ❌ Over-engineering or premature optimization
- ❌ Adding features beyond specification scope
- ❌ Using external dependencies without justification
- ❌ Skipping error handling

### Step 4: Verify Success Criteria

From specification, verify these outcomes:
{{range .SuccessCriteria}}
- [ ] {{.}}
{{end}}

### Step 5: Update Token Status

Once tests pass, update the CANARY token:
```diff
- // CANARY: REQ={{.ReqID}}; FEATURE="{{.Feature}}"; ASPECT={{.Aspect}}; STATUS=STUB; UPDATED={{.Today}}
+ // CANARY: REQ={{.ReqID}}; FEATURE="{{.Feature}}"; ASPECT={{.Aspect}}; STATUS=TESTED; TEST=TestCANARY_{{.ReqID}}_{{.Aspect}}_{{.Feature}}; UPDATED={{.Today}}
```

---

## Suggested File Locations

Based on the ASPECT={{.Aspect}}, consider implementing in:
{{range .SuggestedFiles}}
- `{{.}}`
{{end}}

---

## Verification Checklist

Before marking this requirement as complete:

- [ ] Tests written and passing
- [ ] CANARY token placed with correct fields
- [ ] STATUS updated to TESTED (or BENCHED if benchmarks added)
- [ ] UPDATED field set to today's date
- [ ] All success criteria from specification met
- [ ] No [NEEDS CLARIFICATION] items remain unresolved
- [ ] Code follows project conventions
- [ ] Error handling is comprehensive
- [ ] Documentation is current (if public API)

---

## Next Steps After Completion

1. **Check implementation status:**
   ```bash
   canary status {{.ReqID}}    # View progress summary
   canary files {{.ReqID}}     # List implementation files
   canary show {{.ReqID}}      # Display all tokens
   ```
   **Note:** Use these commands instead of grep or sqlite3!

2. **Run full verification:**
   ```bash
   canary scan --root . --project-only
   ```

3. **Update GAP_ANALYSIS.md** (if STATUS reaches TESTED/BENCHED):
   ```markdown
   ✅ {{.ReqID}} - {{.Feature}} ({{.Aspect}}, verified)
   ```

4. **Proceed to next priority:**
   ```bash
   /canary.next
   ```

---

## Constitutional Reminder

From Article VII (Documentation Currency):
> "CANARY tokens ARE the documentation. When code changes, tokens must be updated. Stale documentation is worse than no documentation."

**Action:** Always update the UPDATED field when modifying this feature.

---

**Ready to implement {{.ReqID}}!** Follow the test-first approach and refer back to this guidance as needed.
