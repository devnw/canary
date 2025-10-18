# CANARY CLI Commands - Agent Reference

This document provides the complete CLI reference for CANARY commands, designed for AI agent execution.

## Installation

```bash
# Build from source
go build -o /usr/local/bin/canary ./cmd/canary

# The binary is self-contained with embedded templates
# No additional files needed for installation
```

## Command Overview

| Command | Purpose | Agent Usage |
|---------|---------|-------------|
| `canary init` | Initialize new project with workflow | Project setup |
| `canary constitution` | Create/view constitutional principles | Establish development rules |
| `canary specify` | Create requirement specification | Define new features |
| `canary plan` | Generate implementation plan | Plan technical approach |
| `canary implement` | Show implementation locations | Find where to code |
| `canary create` | Generate CANARY token | Create token snippets |
| `canary scan` | Scan for CANARY tokens | Generate status reports |

## Detailed Command Reference

### canary init

Initialize a new project with the full CANARY workflow structure.

```bash
canary init [project-name]
```

**Creates:**
- `.canary/` - Full workflow directory structure
- `.canary/memory/constitution.md` - Project governing principles
- `.canary/templates/` - Spec and plan templates
- `.canary/templates/commands/` - Slash command definitions
- `.canary/scripts/` - Automation scripts
- `README_CANARY.md` - Token format specification
- `GAP_ANALYSIS.md` - Requirements tracking template
- `CLAUDE.md` - AI agent integration guide

**Example:**
```bash
canary init my-project
cd my-project
```

**Agent workflow after init:**
1. Run `canary constitution` to review principles
2. Run `canary specify` to create first requirement
3. Follow the spec-driven workflow

### canary constitution

Create or view the project's constitutional principles.

```bash
canary constitution [description]
```

**Behavior:**
- If no `.canary/memory/constitution.md` exists: Creates it from template
- If already exists: Reports status

**Principles created (9 articles):**
1. Requirement-First Development
2. Specification Discipline
3. Token-Driven Planning
4. Test-First Imperative
5. Simplicity and Anti-Abstraction
6. Integration-First Testing
7. Documentation Currency
8. Continuous Improvement
9. Amendment Process

**Example:**
```bash
# Create constitution
canary constitution

# Output:
# ✅ Created constitution at: .canary/memory/constitution.md
#
# Constitutional Principles:
#   I. Requirement-First Development
#   II. Specification Discipline
#   ...
```

**Agent usage:**
- Run before starting new projects
- Reference principles when planning implementations
- Enforce Article IV (Test-First) during development

### canary specify

Create a new requirement specification from a feature description.

```bash
canary specify <feature-description>
```

**Behavior:**
- Auto-generates next requirement ID (CBIN-001, CBIN-002, etc.)
- Creates directory: `.canary/specs/CBIN-XXX-feature-name/`
- Populates `spec.md` from template
- Replaces placeholders with actual values

**Example:**
```bash
canary specify "User authentication with OAuth2 support"

# Output:
# ✅ Created specification: .canary/specs/CBIN-001-User-authentication-with-OAuth2-support/spec.md
#
# Requirement ID: CBIN-001
# Feature: User authentication with OAuth2 support
#
# Next steps:
#   1. Edit .canary/specs/CBIN-001-User-authentication-with-OAuth2-support/spec.md to complete the specification
#   2. Run: canary plan CBIN-001
```

**Agent workflow:**
1. Run `canary specify` with feature description
2. Read the generated spec.md file
3. Fill in:
   - User stories
   - Functional requirements
   - Success criteria
   - Assumptions and constraints
4. Proceed to `canary plan`

### canary plan

Generate a technical implementation plan from a requirement specification.

```bash
canary plan <CBIN-XXX> [tech-stack]
```

**Arguments:**
- `CBIN-XXX`: Required. The requirement ID
- `tech-stack`: Optional. Technology stack to use (e.g., "Go 1.21 with standard library")

**Behavior:**
- Finds spec directory matching requirement ID
- Creates `plan.md` in spec directory
- Populates from template with tech stack if provided

**Example:**
```bash
canary plan CBIN-001 "Go standard library with bcrypt for password hashing"

# Output:
# ✅ Created implementation plan: .canary/specs/CBIN-001-User-authentication-with-OAuth2-support/plan.md
#
# Requirement: CBIN-001
#
# Next steps:
#   1. Edit .canary/specs/CBIN-001-User-authentication-with-OAuth2-support/plan.md to complete the plan
#   2. Implement following TDD (test-first)
#   3. Add CANARY tokens to source code
#   4. Run: canary scan
```

**Agent workflow:**
1. Run `canary plan CBIN-XXX` with optional tech stack
2. Read generated plan.md
3. Fill in:
   - Tech stack rationale
   - CANARY token placement
   - Implementation phases (Phase 0-3)
   - Test strategy
   - Constitutional compliance notes
4. Implement following TDD (tests first!)
5. Add CANARY tokens to source code
6. Run `canary scan` to verify

### canary create

Generate a formatted CANARY token ready to paste into source code.

```bash
canary create <req-id> <feature-name> [flags]
```

**Flags:**
- `--aspect string`: Requirement aspect/category (default "API")
- `--status string`: Implementation status (default "IMPL")
- `--owner string`: Team/person responsible
- `--test string`: Test function name
- `--bench string`: Benchmark function name

**Example:**
```bash
canary create CBIN-105 "UserProfile" --aspect API --status IMPL --owner backend --test TestUserProfile

# Output:
# // CANARY: REQ=CBIN-105; FEATURE="UserProfile"; ASPECT=API; STATUS=IMPL; TEST=TestUserProfile; OWNER=backend; UPDATED=2025-10-16
#
# // Paste this above your implementation:
# // func UserProfile() { ... }
```

**Agent usage:**
- Use when adding CANARY tokens to source code
- Ensures correct format and auto-fills UPDATED date
- Copy output directly into source files

### canary implement

Show implementation points and exact locations for a specific requirement.

```bash
canary implement <CBIN-XXX> [flags]
```

**Flags:**
- `--status string`: Filter by status (STUB, IMPL, TESTED, BENCHED)
- `--aspect string`: Filter by aspect (API, CLI, Engine, etc.)
- `--feature string`: Filter by feature name (partial match)
- `--context`: Show code context around each token
- `--context-lines int`: Number of context lines (default 3)

**Behavior:**
- Scans entire codebase for CANARY tokens matching the requirement ID
- Shows file paths and line numbers for each implementation point
- Displays sub-features from both spec and source code
- Provides progress tracking (% implemented)
- Optionally shows code context for precise navigation

**Example:**
```bash
canary implement CBIN-001

# Output:
# Implementation points for CBIN-001:
#
# 1. JWTGeneration (API, STUB)
#    Location: .canary/specs/CBIN-001-User-authentication/spec.md:175
#
# 2. JWTValidation (API, IMPL)
#    Location: src/auth.go:45
#    Test: TestJWTValidation
#
# 3. UserLogin (API, TESTED)
#    Location: src/handlers/auth.go:23
#    Test: TestUserLogin
#
# Summary:
#   STUB: 1
#   IMPL: 1
#   TESTED: 1
#   Total: 3 implementation points
#
# Progress: 67% (2/3)
```

**With context:**
```bash
canary implement CBIN-001 --context --context-lines 2

# Shows:
#    Context:
#       44: func ValidateJWT(token string) (*Claims, error) {
#    >> 45: // CANARY: REQ=CBIN-001; FEATURE="JWTValidation"; ASPECT=API; STATUS=IMPL; TEST=TestJWTValidation; UPDATED=2025-10-16
#       46:     claims := &Claims{}
#       47:     parsedToken, err := jwt.ParseWithClaims(token, claims, keyFunc)
```

**Filter by status:**
```bash
# Show only unimplemented features
canary implement CBIN-001 --status STUB

# Show only completed features
canary implement CBIN-001 --status TESTED
```

**Filter by aspect:**
```bash
# Show only API-related implementation points
canary implement CBIN-001 --aspect API

# Show only tests
canary implement CBIN-001 --aspect API --feature Test
```

**Agent usage:**
- Run before starting implementation to see what's needed
- Run during implementation to find exact locations
- Run after implementation to verify completeness
- Use `--status STUB` to find remaining work
- Use `--context` to get code snippets for LLM context

**Reduces agent context by:**
1. Showing exact file:line locations (no need to search)
2. Providing code context (no need to read entire files)
3. Filtering to relevant features only
4. Displaying progress at a glance

### canary scan

Scan codebase for CANARY tokens and generate reports.

```bash
canary scan [flags]
```

**Note:** Currently wraps the `tools/canary` scanner. Passes all flags through.

**Common usage:**
```bash
# Generate JSON report
canary scan --root . --out status.json

# Generate both JSON and CSV
canary scan --root . --out status.json --csv status.csv

# Verify claims in GAP_ANALYSIS.md
canary scan --root . --verify GAP_ANALYSIS.md

# Check for stale tokens (>30 days)
canary scan --root . --strict

# Auto-update stale tokens
canary scan --root . --update-stale
```

**Exit codes:**
- 0: Success
- 2: Verification/staleness failed
- 3: Parse or IO error

**Agent usage:**
- Run after implementation to verify status
- Use `--verify GAP_ANALYSIS.md` to check claims
- Use `--strict` in CI/CD pipelines
- Parse `status.json` for requirement coverage metrics

## Complete Workflow Example

```bash
# 1. Initialize project
canary init my-api

# 2. Create constitution
cd my-api
canary constitution

# 3. Specify requirement
canary specify "Add user authentication with JWT"

# 4. Create implementation plan
canary plan CBIN-001 "Go 1.21, golang-jwt/jwt library"

# 5. Check what needs to be implemented
canary implement CBIN-001 --status STUB
# Shows all unimplemented features with location hints

# 6. Implement first feature (following TDD)
canary implement CBIN-001 --feature JWTGeneration --context
# Shows exact location to implement

# Write tests first
canary create CBIN-001 "JWTGeneration" --aspect API --status IMPL --test TestJWTGeneration
# Add token and implement feature

# 7. Check progress
canary implement CBIN-001
# Shows: Progress: 33% (1/3)

# 8. Continue implementing remaining features
canary implement CBIN-001 --status STUB
# Repeat for each feature

# 9. Scan and verify
canary scan --root . --out status.json
canary scan --root . --verify GAP_ANALYSIS.md --strict

# 10. Update GAP_ANALYSIS.md
echo "✅ CBIN-001 - User authentication fully tested" >> GAP_ANALYSIS.md
```

## Agent Integration Notes

**For Claude Code, Cursor, and similar AI agents:**

1. **Before any implementation:**
   - Run `canary constitution` to review principles
   - Run `canary specify` to create structured requirements
   - Run `canary plan` to create implementation plan

2. **During implementation:**
   - Follow Article IV: Test-First Imperative
   - Use `canary create` to generate properly formatted tokens
   - Add tokens at the function/module level

3. **After implementation:**
   - Run `canary scan` to verify status
   - Update GAP_ANALYSIS.md with completed requirements
   - Use `--verify` to ensure no overclaiming

4. **Constitutional compliance:**
   - Article I: Every feature MUST start with a requirement (use `specify`)
   - Article IV: Tests MUST be written before implementation (non-negotiable)
   - Article VII: Keep tokens current (use `--update-stale`)

## Binary Deployment

The canary binary is self-contained:

```bash
# Build
go build -o canary ./cmd/canary

# Install system-wide
sudo cp canary /usr/local/bin/

# No configuration files needed
# Templates are embedded in the binary
```

**Verification:**
```bash
# Should show help
canary --help

# Should show version
canary version
```

## Status Values

| Status | Meaning | Next Step |
|--------|---------|-----------|
| STUB | Planned but not implemented | Implement it |
| IMPL | Implemented | Add tests (TEST=) |
| TESTED | Implemented with tests | Add benchmarks (BENCH=) |
| BENCHED | Tested with benchmarks | Maintain currency |
| REMOVED | Deprecated/removed | Archive |

**Auto-promotion:** Scanner promotes IMPL→TESTED when TEST= present, TESTED→BENCHED when BENCH= present.

## Valid Aspects

API, CLI, Engine, Storage, Security, Docs, Wire, Planner, Decode, Encode, RoundTrip, Bench, FrontEnd, Dist

## Required Fields

- `REQ`: Requirement ID (CBIN-###)
- `FEATURE`: Short feature name
- `ASPECT`: Category
- `STATUS`: Implementation state
- `UPDATED`: Last update date (YYYY-MM-DD)

## Optional Fields

- `TEST`: Test function name
- `BENCH`: Benchmark function name
- `OWNER`: Team/person responsible
