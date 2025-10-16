# CANARY - AI Agent Integration Guide

**For Claude Code, Cursor, and other AI coding agents**

## Overview

CANARY is a requirement tracking system with full CLI support, designed for AI-agent-driven development. All functionality is available through standalone CLI commands - no shell scripts or external dependencies required.

## Quick Start

```bash
# Install (single binary, no dependencies)
go build -o /usr/local/bin/canary ./cmd/canary

# Initialize project
canary init my-project
cd my-project

# Agent workflow
canary constitution              # Review principles
canary specify "feature desc"    # Create requirement
canary plan CBIN-001             # Create implementation plan
canary create CBIN-001 "Name"    # Generate token for source
canary scan                      # Verify implementation
```

## Commands for Agents

### Project Setup

**Command:** `canary init <project-name>`

**Purpose:** Initialize new project with full workflow structure

**Creates:**
- `.canary/memory/constitution.md` - 9 articles of governing principles
- `.canary/templates/` - Spec and plan templates
- `.canary/scripts/` - Automation tools
- `README_CANARY.md` - Token format reference
- `GAP_ANALYSIS.md` - Requirements tracking
- `CLAUDE.md` - Slash command reference

**Agent usage:**
```bash
canary init my-api
cd my-api
```

### Constitutional Principles

**Command:** `canary constitution`

**Purpose:** Create/view project governing principles

**Output:** Creates `.canary/memory/constitution.md` with 9 articles:
1. Requirement-First Development
2. Specification Discipline
3. Token-Driven Planning
4. **Test-First Imperative** (NON-NEGOTIABLE)
5. Simplicity and Anti-Abstraction
6. Integration-First Testing
7. Documentation Currency
8. Continuous Improvement
9. Amendment Process

**Agent usage:**
```bash
# Review before starting any implementation
canary constitution
```

**Key principle for agents:**
- **Article IV is NON-NEGOTIABLE**: Tests MUST be written before implementation
- No implementation code without failing tests first (Red phase)
- Implementation makes tests pass (Green phase)

### Create Requirement Specification

**Command:** `canary specify <feature-description>`

**Purpose:** Create structured requirement from natural language description

**Behavior:**
- Auto-generates next requirement ID (CBIN-001, CBIN-002, ...)
- Creates `.canary/specs/CBIN-XXX-feature-name/spec.md`
- Populates template with feature description and date

**Agent usage:**
```bash
canary specify "Add user authentication with JWT tokens"
# Creates: .canary/specs/CBIN-001-Add-user-authentication-with-JWT-tokens/spec.md
```

**Agent workflow after specify:**
1. Read generated `spec.md`
2. Fill in:
   - User stories with acceptance criteria
   - Functional requirements (testable, measurable)
   - Success criteria
   - Assumptions and constraints
3. Ensure ≤ 3 [NEEDS CLARIFICATION] markers
4. Proceed to `canary plan`

### Create Implementation Plan

**Command:** `canary plan <CBIN-XXX> [tech-stack]`

**Purpose:** Generate technical implementation plan

**Arguments:**
- `CBIN-XXX`: Requirement ID
- `tech-stack`: Optional. Technology choices (e.g., "Go 1.21 with stdlib")

**Agent usage:**
```bash
canary plan CBIN-001 "Go standard library with golang-jwt/jwt for tokens"
# Creates: .canary/specs/CBIN-001-Add-user-authentication-with-JWT-tokens/plan.md
```

**Agent workflow after plan:**
1. Read generated `plan.md`
2. Fill in:
   - Tech stack rationale
   - CANARY token placement specification
   - Implementation phases:
     - **Phase 0**: Pre-implementation gates (constitutional compliance)
     - **Phase 1**: Test creation (Red phase) - REQUIRED
     - **Phase 2**: Implementation (Green phase)
     - **Phase 3**: Benchmarking (performance-critical features)
3. **Execute Phase 1 BEFORE Phase 2** (Article IV)

### Generate CANARY Token

**Command:** `canary create <req-id> <feature-name> [flags]`

**Purpose:** Generate properly formatted CANARY token for source code

**Flags:**
- `--aspect`: API, CLI, Engine, Storage, Security, Docs, etc. (default: API)
- `--status`: STUB, IMPL, TESTED, BENCHED (default: IMPL)
- `--owner`: Team/person responsible
- `--test`: Test function name (promotes to TESTED)
- `--bench`: Benchmark function name (promotes to BENCHED)

**Agent usage:**
```bash
# Generate token for implementation
canary create CBIN-001 "UserAuth" \
  --aspect API \
  --status IMPL \
  --test TestUserAuth \
  --owner backend

# Output (copy into source):
# // CANARY: REQ=CBIN-001; FEATURE="UserAuth"; ASPECT=API; STATUS=IMPL; TEST=TestUserAuth; OWNER=backend; UPDATED=2025-10-16
```

**Where to place tokens:**
- At package, function, or struct level
- Above the main implementation
- One token per logical feature unit

### Scan and Verify

**Command:** `canary scan [flags]`

**Purpose:** Scan codebase for tokens, generate reports, verify claims

**Common usage:**
```bash
# Generate status report
canary scan --root . --out status.json --csv status.csv

# Verify GAP_ANALYSIS.md claims
canary scan --root . --verify GAP_ANALYSIS.md

# Check for stale tokens (>30 days)
canary scan --root . --strict

# Auto-update stale tokens
canary scan --root . --update-stale
```

**Agent usage:**
```bash
# After implementation
canary scan --root . --out status.json

# Parse status.json for:
# - Coverage by status (STUB/IMPL/TESTED/BENCHED)
# - Coverage by aspect (API/CLI/Engine/etc)
# - Stale tokens needing updates

# Before claiming completion
canary scan --root . --verify GAP_ANALYSIS.md --strict
```

## Complete Agent Workflow

```bash
# 1. Initialize
canary init my-service
cd my-service

# 2. Review principles
canary constitution
# Key takeaway: Article IV - Test-First is NON-NEGOTIABLE

# 3. Create requirement
canary specify "Add health check endpoint"
# Creates: CBIN-001

# 4. Edit spec (agent fills in details)
# Edit .canary/specs/CBIN-001-Add-health-check-endpoint/spec.md
# Fill in user stories, requirements, success criteria

# 5. Create plan
canary plan CBIN-001 "Go 1.21 net/http"
# Edit .canary/specs/CBIN-001-Add-health-check-endpoint/plan.md
# Fill in tech rationale, phases, test strategy

# 6. Implement - PHASE 1: TESTS FIRST (Article IV)
# Write test:
cat > health_test.go <<'EOF'
package main

import "testing"

// Test MUST exist before implementation
func TestHealthCheck(t *testing.T) {
    // This will FAIL until implementation exists (Red phase)
    resp := HealthCheck()
    if resp != "OK" {
        t.Errorf("expected OK, got %s", resp)
    }
}
EOF

# Run test - confirm it FAILS
go test ./...

# 7. Implement - PHASE 2: MAKE TESTS PASS (Green phase)
cat > main.go <<'EOF'
package main

// CANARY: REQ=CBIN-001; FEATURE="HealthCheck"; ASPECT=API; STATUS=IMPL; TEST=TestHealthCheck; OWNER=backend; UPDATED=2025-10-16
func HealthCheck() string {
    return "OK"
}
EOF

# Run test - confirm it PASSES
go test ./...

# 8. Scan and verify
canary scan --root . --out status.json

# 9. Update GAP_ANALYSIS.md
echo "✅ CBIN-001 - Health check endpoint fully tested" >> GAP_ANALYSIS.md

# 10. Verify claims
canary scan --root . --verify GAP_ANALYSIS.md --strict
```

## Constitutional Compliance for Agents

### Article I: Requirement-First Development
**Agent action:** Always run `canary specify` before coding
**Verification:** Every feature has a CBIN-XXX requirement ID

### Article IV: Test-First Imperative
**Agent action:** MUST write tests before implementation
**Workflow:**
1. Run `canary plan` and fill in Phase 1 (test creation)
2. Write tests - confirm they FAIL (Red phase)
3. Write implementation - confirm tests PASS (Green phase)
**Verification:** Every CANARY token with STATUS=IMPL or TESTED must have TEST= field

### Article VII: Documentation Currency
**Agent action:** Keep tokens up-to-date
**Verification:** Run `canary scan --strict` to check for stale tokens (>30 days)
**Maintenance:** Run `canary scan --update-stale` to auto-update UPDATED field

## Status Progression

| Status | Meaning | Agent Workflow |
|--------|---------|----------------|
| STUB | Planned, not implemented | Create from `canary specify`, add to plan |
| IMPL | Implemented | Add token with `canary create`, implement feature |
| TESTED | Implemented + tests | Add TEST= field to token (auto-promoted by scanner) |
| BENCHED | Tested + benchmarks | Add BENCH= field to token (auto-promoted) |

**Auto-promotion rules:**
- IMPL → TESTED when TEST= field present
- TESTED → BENCHED when BENCH= field present
- Scanner performs promotion automatically

## Output Parsing

**status.json structure:**
```json
{
  "generated_at": "2025-10-16T00:00:00Z",
  "requirements": [
    {
      "id": "CBIN-001",
      "features": [
        {
          "feature": "HealthCheck",
          "aspect": "API",
          "status": "TESTED",
          "files": ["main.go"],
          "tests": ["TestHealthCheck"],
          "benches": [],
          "owner": "backend",
          "updated": "2025-10-16"
        }
      ]
    }
  ],
  "summary": {
    "by_status": {"TESTED": 1, "IMPL": 0, "STUB": 0},
    "by_aspect": {"API": 1},
    "total_tokens": 1,
    "unique_requirements": 1
  }
}
```

**Agent usage:**
- Parse `summary.by_status` for completion percentage
- Check `summary.by_aspect` for coverage across aspects
- Identify stale tokens from `features[].updated` dates
- Verify all required features are TESTED or BENCHED

## Exit Codes

| Code | Meaning | Agent Response |
|------|---------|----------------|
| 0 | Success | Proceed |
| 2 | Verification/staleness failed | Fix issues, update tokens |
| 3 | Parse/IO error | Check file permissions, token format |

## Binary Installation

```bash
# Build
go build -o canary ./cmd/canary

# Install system-wide
sudo cp canary /usr/local/bin/

# Verify
canary --help
```

**No external dependencies:**
- All templates embedded in binary
- No configuration files needed
- Portable across systems

## Agent Best Practices

1. **Always start with constitution:**
   ```bash
   canary constitution
   ```

2. **Use spec-driven workflow:**
   ```bash
   canary specify "feature" → edit spec → canary plan "CBIN-XXX" → edit plan → implement
   ```

3. **Enforce test-first (Article IV):**
   - Write tests
   - Confirm tests FAIL
   - Implement feature
   - Confirm tests PASS

4. **Add tokens during implementation:**
   ```bash
   canary create CBIN-XXX "FeatureName" --test TestName
   ```

5. **Verify before claiming completion:**
   ```bash
   canary scan --verify GAP_ANALYSIS.md --strict
   ```

6. **Maintain currency:**
   ```bash
   canary scan --update-stale
   ```

## Example: Multi-Feature Project

```bash
# Initialize
canary init auth-service
cd auth-service

# Feature 1: JWT tokens
canary specify "JWT token generation and validation"
canary plan CBIN-001 "golang-jwt/jwt v5"
# ... implement with tests ...

# Feature 2: OAuth2
canary specify "OAuth2 integration with Google and GitHub"
canary plan CBIN-002 "golang.org/x/oauth2"
# ... implement with tests ...

# Feature 3: Rate limiting
canary specify "Rate limiting per user and IP"
canary plan CBIN-003 "golang.org/x/time/rate"
# ... implement with tests ...

# Scan all features
canary scan --root . --out status.json --csv status.csv

# Verify completion
canary scan --verify GAP_ANALYSIS.md --strict

# Expected status.json summary:
# {
#   "summary": {
#     "by_status": {"TESTED": 3},
#     "by_aspect": {"API": 2, "Security": 1},
#     "total_tokens": 3,
#     "unique_requirements": 3
#   }
# }
```

## Troubleshooting

**Issue:** "specification not found for CBIN-XXX"
**Solution:** Run `canary specify` first to create the spec

**Issue:** "plan already exists"
**Solution:** Edit existing plan or delete it to regenerate

**Issue:** Scan fails with parse error
**Solution:** Check CANARY token format, ensure all required fields present

**Issue:** Verification fails with "overclaim"
**Solution:** Update GAP_ANALYSIS.md to match actual status (don't claim TESTED unless token has TEST=)

## Reference

- [CLI_COMMANDS.md](./CLI_COMMANDS.md) - Complete command reference
- [README.md](./README.md) - Project overview
- `.canary/memory/constitution.md` - Constitutional principles (after init)
- `.canary/templates/` - Spec and plan templates (after init)
