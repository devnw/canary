# Implementation Guidance: {{.FeatureName}}
<!-- CANARY: REQ=CBIN-CLI-001; FEATURE="ImplementTemplate"; ASPECT=Docs; STATUS=TESTED; UPDATED=2025-10-16 -->

**Requirement ID:** {{.ReqID}}

---

## 📋 Specification

{{.SpecContent}}

---

## 🗺️ Implementation Plan

{{.PlanContent}}

---

## 📝 Implementation Checklist

{{.Checklist}}

---

## 📊 Current Progress

- **Total Features:** {{.Progress.Total}}
- **Tested:** {{.Progress.Tested}}
- **Implemented (needs tests):** {{.Progress.Impl}}
- **Stub (not started):** {{.Progress.Stub}}

---

## ⚖️ Constitutional Principles

{{.Constitution}}

---

## 🎯 Implementation Instructions

### Test-First Development (Article IV)

1. **Write tests BEFORE implementation**
   - Create test file: `*_test.go` or equivalent
   - Test function name format: `TestCANARY_{{.ReqID}}_<ASPECT>_<Description>`
   - All tests must fail initially (TDD Red phase)

2. **Implement to pass tests**
   - Write minimal code to make tests pass (TDD Green phase)
   - Refactor while keeping tests green

3. **Update CANARY tokens**
   - Start with: `STATUS=STUB`
   - After implementation: `STATUS=IMPL`
   - After tests pass: `STATUS=TESTED; TEST=TestFunctionName`
   - After benchmarks: `STATUS=BENCHED; BENCH=BenchFunctionName`

### CANARY Token Placement

Every implementation point must have a CANARY token:

```
// CANARY: REQ={{.ReqID}}; FEATURE="FeatureName"; ASPECT=API; STATUS=IMPL; UPDATED=YYYY-MM-DD
```

**Token Fields:**
- `REQ` - Requirement ID ({{.ReqID}})
- `FEATURE` - Short descriptive name (quoted)
- `ASPECT` - Category (API, CLI, Engine, Storage, Security, Docs, etc.)
- `STATUS` - Implementation state (STUB → IMPL → TESTED → BENCHED)
- `TEST` - Test function name (when STATUS=TESTED)
- `BENCH` - Benchmark function name (when STATUS=BENCHED)
- `OWNER` - Team/person responsible (optional)
- `UPDATED` - Last update date (YYYY-MM-DD format)

### File Organization

Follow the structure in the Implementation Plan. Key principles:

- Place tokens above the primary function/class/component
- Group related features in the same file
- Update `UPDATED` field whenever code changes
- Keep tokens synchronized with actual implementation state

### Success Criteria

Before marking complete:

✅ All tests written and passing
✅ All CANARY tokens updated to TESTED
✅ Code follows constitutional principles (simplicity, no unnecessary abstraction)
✅ Documentation updated if needed
✅ `UPDATED` field current on all tokens

---

## 🚀 Get Started

1. Review the specification and plan above
2. Create test files first (TDD Red)
3. Implement features to pass tests (TDD Green)
4. Add/update CANARY tokens as you go
5. Query tokens using proper commands:
   - `canary show {{.ReqID}}` - Display all tokens
   - `canary files {{.ReqID}}` - List implementation files
   - `canary status {{.ReqID}}` - Show progress summary
   - **DO NOT use grep or sqlite3 directly!**
6. Run `canary scan` to verify token status
7. Mark spec checklist items complete as you finish

---

**Remember:** Test-first, token-driven, evidence-based development!
