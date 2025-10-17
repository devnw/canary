---
description: Generate technical implementation plan from requirement specification
---

## User Input

```text
$ARGUMENTS
```

## Outline

Create a detailed technical implementation plan for a CANARY requirement.

1. **Identify requirement**:
   - If argument is REQ ID (CBIN-XXX): Load that spec
   - If empty: Find most recent STUB requirement
   - Error if spec not found

2. **Load and validate specification**:
   - Read `.canary/specs/CBIN-XXX-name/spec.md`
   - Verify no [NEEDS CLARIFICATION] markers remain
   - Confirm ready for planning

3. **Review constitution**:
   - Load `.canary/memory/constitution.md`
   - Note applicable principles (test-first, simplicity, etc.)
   - Will enforce these in plan validation gates

4. **Analyze user's technical requirements** (from arguments):
   - Tech stack preferences (languages, frameworks)
   - Architectural constraints
   - Performance requirements
   - Security requirements

5. **Create implementation plan** at `.canary/specs/CBIN-XXX-name/plan.md`:

   **Required sections:**
   - **Tech Stack Decision**: Chosen technologies with rationale
   - **Architecture Overview**: High-level component structure
   - **CANARY Token Placement**: Where to add the requirement token in code
   - **Implementation Phases**: Ordered steps (test-first approach)
   - **Testing Strategy**: Unit, integration, acceptance tests
   - **Constitutional Gates**: Validation against constitution principles

   **Example structure:**
   ```markdown
   # Implementation Plan: CBIN-XXX FeatureName

   ## Tech Stack Decision
   - Language: Go 1.25
   - Framework: Standard library
   - Database: PostgreSQL (if needed)
   - Rationale: [Why these choices]

   ## CANARY Token Placement
   ```go
   // File: internal/feature/feature.go
   // CANARY: REQ=CBIN-XXX; FEATURE="FeatureName"; ASPECT=API; STATUS=IMPL; OWNER=team; UPDATED=2025-10-16
   package feature

   func ExecuteFeature() error {
       // implementation
   }
   ```

   ## Implementation Phases

   ### Phase 0: Pre-Implementation Gates
   - [ ] Constitution compliance (Article I, IV, V)
   - [ ] Test-first approach planned
   - [ ] Simplicity gate passed

   ### Phase 1: Test Creation
   - [ ] Write TestExecuteFeature (red phase)
   - [ ] Update token: TEST=TestExecuteFeature
   - [ ] Verify test fails

   ### Phase 2: Implementation
   - [ ] Implement ExecuteFeature to pass tests
   - [ ] Update token STATUS to TESTED
   - [ ] Verify all tests pass

   ### Phase 3: Benchmarking (if performance-critical)
   - [ ] Write BenchmarkExecuteFeature
   - [ ] Update token: BENCH=BenchmarkExecuteFeature
   - [ ] Update token STATUS to BENCHED
   - [ ] Document baseline performance

   ## Testing Strategy
   - **Unit Tests**: TestExecuteFeature
   - **Integration Tests**: [If needed]
   - **Acceptance Tests**: [Based on spec success criteria]
   - **Benchmarks**: BenchmarkExecuteFeature (if perf-critical)

   ## Constitutional Compliance
   - ✅ Article I: CANARY token created
   - ✅ Article IV: Test-first approach
   - ✅ Article V: Simplicity (using standard library)
   - ✅ Article VI: Integration tests planned
   ```

6. **Validate plan** against constitution:
   - **Article I Gate**: CANARY token properly formatted
   - **Article IV Gate**: Tests planned before implementation
   - **Article V Gate**: Simplicity maintained (no unnecessary complexity)
   - **Article VI Gate**: Integration testing strategy defined

7. **Update requirement tracking**:
   - Update `.canary/requirements.md`:
     ```markdown
     - [ ] CBIN-XXX - FeatureName (STATUS=STUB → ready for implementation)
     ```

8. **Report completion**:
   - Plan file path
   - Key decisions summary
   - Constitutional compliance status
   - Next steps: Use `/canary.tasks` to break down into actionable tasks

## Plan Quality Checklist

After creating plan, validate:

- [ ] Tech stack decisions have documented rationale
- [ ] CANARY token placement clearly specified
- [ ] Test-first approach explicitly outlined
- [ ] Implementation phases respect dependencies
- [ ] All constitutional gates addressed
- [ ] Performance considerations documented (if applicable)
- [ ] Security considerations documented (if applicable)

## Guidelines

- **Constitution First**: Reference and enforce constitutional principles
- **Test-First Mandatory**: Phase 1 must always be test creation
- **Token Evolution**: Show how token progresses from STUB → IMPL → TESTED → BENCHED
- **Simplicity**: Justify any complexity against Article V
- **Realistic**: Plan should be implementable by AI agent in `/canary.implement`
