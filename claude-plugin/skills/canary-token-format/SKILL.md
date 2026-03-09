---
name: canary-token-format
description: >
  Use when placing or editing CANARY tokens in source code, asking about token
  format, fields, aspects, status progression, or auto-promotion rules.
---

# CANARY Token Format Reference

## Standard Token

```
// CANARY: REQ=<ID>; FEATURE="<Name>"; ASPECT=<Aspect>; STATUS=<Status>; UPDATED=YYYY-MM-DD
```

## Full Token (All Fields)

```
// CANARY: REQ=PROJ-105; FEATURE="UserAuth"; ASPECT=API; STATUS=TESTED; TEST=TestUserAuth; BENCH=BenchUserAuth; OWNER=team; PRIORITY=1; DOC=user:docs/auth.md; DOC_HASH=a3f5b8c2; UPDATED=2025-10-18
```

## Required Fields

| Field | Format | Description |
|-------|--------|-------------|
| `REQ` | `<KEY>-<NNN>` | Requirement ID (zero-padded 3+ digits) |
| `FEATURE` | `"QuotedName"` | Feature name in quotes |
| `ASPECT` | See list below | Architectural aspect |
| `STATUS` | `STUB\|IMPL\|TESTED\|BENCHED` | Implementation status |
| `UPDATED` | `YYYY-MM-DD` | Last update date |

## Optional Fields

| Field | Format | Description |
|-------|--------|-------------|
| `TEST` | `TestFunctionName` | Test function name (auto-promotes to TESTED) |
| `BENCH` | `BenchFunctionName` | Benchmark name (auto-promotes to BENCHED) |
| `OWNER` | `team-name` | Responsible team/person |
| `PRIORITY` | `1-10` | Priority (1=highest) |
| `PHASE` | `Phase0-Phase3` | Implementation phase |
| `KEYWORDS` | `comma,separated` | Searchable keywords |
| `DOC` | `type:path` | Documentation file with type prefix |
| `DOC_HASH` | `hex16chars` | SHA256 hash (first 16 chars) of doc file |
| `DEPENDS_ON` | `REQ-ID,REQ-ID` | Dependencies |
| `BLOCKS` | `REQ-ID,REQ-ID` | What this blocks |

## Valid Aspects

`API`, `CLI`, `Engine`, `Storage`, `Security`, `Docs`, `Wire`, `Planner`, `Decode`, `Encode`, `RoundTrip`, `Bench`, `FrontEnd`, `Dist`

## Status Progression

```
STUB → IMPL → TESTED → BENCHED
```

- **STUB** — Placeholder, not yet implemented. Token marks where code will go.
- **IMPL** — Implementation exists. Token is in working code.
- **TESTED** — Has passing tests. Auto-promoted when `TEST=` field is added.
- **BENCHED** — Has performance benchmarks. Auto-promoted when `BENCH=` field is added.

## Auto-Promotion Rules

Adding `TEST=TestFunctionName` automatically sets `STATUS=TESTED`:
```go
// CANARY: REQ=PROJ-105; FEATURE="UserAuth"; ASPECT=API; STATUS=TESTED; TEST=TestUserAuth; UPDATED=2025-10-18
```

Adding `BENCH=BenchFunctionName` automatically sets `STATUS=BENCHED`:
```go
// CANARY: REQ=PROJ-105; FEATURE="UserAuth"; ASPECT=API; STATUS=BENCHED; TEST=TestUserAuth; BENCH=BenchUserAuth; UPDATED=2025-10-18
```

## Placement Guidelines

Place tokens **above** the implementation they track:

```go
// CANARY: REQ=PROJ-105; FEATURE="UserAuth"; ASPECT=API; STATUS=IMPL; UPDATED=2025-10-18
func Authenticate(creds Credentials) (*Session, error) {
    // implementation
}
```

A single requirement can have multiple tokens across different files and aspects:

```go
// In src/api/auth.go
// CANARY: REQ=PROJ-105; FEATURE="UserAuth"; ASPECT=API; STATUS=TESTED; TEST=TestUserAuth; UPDATED=2025-10-18

// In internal/storage/session.go
// CANARY: REQ=PROJ-105; FEATURE="SessionStore"; ASPECT=Storage; STATUS=IMPL; UPDATED=2025-10-18

// In cmd/app/auth.go
// CANARY: REQ=PROJ-105; FEATURE="AuthCLI"; ASPECT=CLI; STATUS=STUB; UPDATED=2025-10-18
```

## Documentation Linking

Link documentation with type prefixes:

```go
// Single doc:
// DOC=user:docs/auth.md; DOC_HASH=a3f5b8c2e1d4a6f9

// Multiple docs:
// DOC=user:docs/auth.md,api:docs/api/auth.md; DOC_HASH=a3f5b8c2,b4e6d8f0
```

Documentation types: `user`, `api`, `technical`, `feature`, `architecture`

## Bug Tokens

Bug tokens use a different format:

```go
// CANARY: BUG=BUG-API-123; TITLE="Login fails on first attempt"; ASPECT=API; STATUS=OPEN; SEVERITY=S2; PRIORITY=P1; REPRO=3/5; UPDATED=2025-10-18
```

## Creating Tokens via CLI

```bash
canary create <REQ-ID> "FeatureName" --aspect API --status STUB
```

## Language Support

Tokens work in any language that supports line comments:

```python
# CANARY: REQ=PROJ-105; FEATURE="UserAuth"; ASPECT=API; STATUS=IMPL; UPDATED=2025-10-18
```

```rust
// CANARY: REQ=PROJ-105; FEATURE="UserAuth"; ASPECT=API; STATUS=IMPL; UPDATED=2025-10-18
```

```yaml
# CANARY: REQ=PROJ-105; FEATURE="Config"; ASPECT=Docs; STATUS=IMPL; UPDATED=2025-10-18
```
