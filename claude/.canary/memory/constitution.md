<!-- CANARY: REQ=CBIN-107; FEATURE="Constitution"; ASPECT=Docs; STATUS=IMPL; OWNER=canary; UPDATED=2025-10-16 -->
# CANARY Development Constitution

**Version:** 1.0
**Effective Date:** 2025-10-16
**Status:** ACTIVE

## Preamble

This constitution establishes the governing principles for development using the CANARY requirement tracking system. All specifications, plans, and implementations SHALL adhere to these principles.

## Article I: Requirement-First Development

**Section 1.1: Token Primacy**

Every feature MUST begin with a CANARY token. No implementation shall exist without a corresponding requirement identifier.

```
// CANARY: REQ=CBIN-###; FEATURE="Name"; ASPECT=API; STATUS=IMPL; UPDATED=YYYY-MM-DD
```

**Section 1.2: Evidence-Based Promotion**

Status progression follows strict evidence requirements:

- STUB → IMPL: Implementation exists
- IMPL → TESTED: TEST= field references passing test
- TESTED → BENCHED: BENCH= field references performance benchmark
- All promotions are automatic based on token evidence

**Section 1.3: Staleness Management**

TESTED and BENCHED tokens older than 30 days SHALL be considered stale. Projects SHOULD use `--update-stale` to maintain currency.

## Article II: Specification Discipline

**Section 2.1: WHAT Before HOW**

Specifications MUST focus on user needs and business value. Implementation details (languages, frameworks, databases) belong in technical plans, not specifications.

**Section 2.2: Testable Requirements**

Every functional requirement MUST be:
- Unambiguous (single interpretation)
- Testable (can be verified)
- Measurable (has clear success criteria)

**Section 2.3: Clarification Limits**

Specifications MAY include [NEEDS CLARIFICATION] markers, but:
- Maximum 3 per specification
- Only for critical decisions (scope, security, UX)
- Must provide suggested defaults

## Article III: Token-Driven Planning

**Section 3.1: Token Granularity**

Each requirement token represents:
- A single, cohesive feature
- A trackable unit of work
- A verifiable outcome

**Section 3.2: Aspect Classification**

Tokens MUST be classified by aspect:
- **API**: External interfaces
- **CLI**: Command-line tools
- **Engine**: Core logic
- **Storage**: Data persistence
- **Security**: Authentication, authorization, encryption
- **Docs**: Documentation
- Other valid aspects: Wire, Planner, Decode, Encode, RoundTrip, Bench, FrontEnd, Dist

**Section 3.3: Cross-Cutting Concerns**

Requirements spanning multiple aspects SHOULD be split into separate tokens, one per aspect.

## Article IV: Test-First Imperative

**Section 4.1: Test Before Implementation**

This is NON-NEGOTIABLE: All implementation MUST follow Test-Driven Development.

No implementation code shall be written before:
1. Test functions are written and named in token TEST= field
2. Tests are confirmed to FAIL (Red phase)
3. Implementation makes tests PASS (Green phase)

**Section 4.2: Benchmark Requirements**

Performance-critical features MUST include benchmarks:
- Baseline performance documented
- Performance regression detection
- Benchmark names in BENCH= field

## Article V: Simplicity and Anti-Abstraction

**Section 5.1: Minimal Complexity**

Features SHOULD use the simplest approach that meets requirements:
- Prefer standard library over dependencies
- Avoid premature optimization
- No speculative features

**Section 5.2: Framework Trust**

When using frameworks, trust their features directly rather than wrapping them in custom abstractions.

**Section 5.3: Complexity Justification**

Any complexity beyond simple CRUD operations MUST be documented with rationale in the technical plan.

## Article VI: Integration-First Testing

**Section 6.1: Real Environment Testing**

Tests MUST use realistic environments:
- Prefer real databases over mocks
- Use actual service instances
- Test actual file I/O operations

**Section 6.2: Contract-First Development**

APIs MUST define contracts before implementation:
- Contract defined in plan
- Contract tests written first
- Implementation satisfies contract

## Article VII: Documentation Currency

**Section 7.1: Code as Documentation**

CANARY tokens ARE the documentation. Keep them current:
- Update UPDATED field when modifying features
- Update STATUS when adding tests/benchmarks
- Add OWNER for team accountability

**Section 7.2: Gap Analysis**

Projects MUST maintain GAP_ANALYSIS.md tracking:
- ✅ Verified requirements (TESTED/BENCHED)
- Open gaps (STUB/IMPL)
- Planned features (MISSING)

**Section 7.3: Self-Verification**

Projects SHOULD run `canary scan --verify GAP_ANALYSIS.md --strict` in CI to prevent overclaims.

## Article VIII: Continuous Improvement

**Section 8.1: Metrics-Driven Development**

Track requirement metrics:
- Coverage by status (STUB vs TESTED vs BENCHED)
- Coverage by aspect (API vs CLI vs Engine, etc.)
- Staleness (tokens > 30 days old)

**Section 8.2: Regular Audits**

Projects SHOULD:
- Run `canary scan --strict` weekly
- Review stale tokens monthly
- Update or remove obsolete requirements

## Article IX: Amendment Process

**Section 9.1: Constitutional Changes**

Modifications to this constitution require:
- Explicit documentation of rationale
- Review by project maintainers
- Version increment

**Section 9.2: Project-Specific Additions**

Projects MAY add project-specific articles that do not contradict these core principles.

## Enforcement

AI agents implementing CANARY-tracked features MUST:
1. Reference this constitution before creating plans
2. Validate specifications against Article II
3. Enforce Article IV (test-first) during implementation
4. Update tokens per Article VII

Violations SHALL be documented in plan complexity tracking sections with justification.

---

**Ratified:** 2025-10-16
**Next Review:** 2026-01-16
