---
name: canary-examples
description: >
  Use when the user asks for canary examples, common patterns, or CLI usage
  recipes. Provides quick-reference patterns grouped by workflow.
---

# CANARY CLI Examples

## Getting Started

```bash
# Initialize CANARY in a new project
canary init --agents claude

# Set up project principles
canary constitution

# Create your first requirement
canary specify "Add user authentication with email/password"
# Creates: .canary/specs/PROJ-101-user-authentication/spec.md
```

## View First (One-Call Context)

```bash
# Full picture of a requirement: status, files, tests, deps, spec, ticket
canary view PROJ-105

# Machine-readable
canary view PROJ-105 --json

# Raise the per-section cap (default 10)
canary view PROJ-105 --limit 20
```

Prefer `view` over separate `show`/`status`/`files` calls when you need the full context.

## Daily Development

```bash
# What should I work on next?
canary next --prompt

# Get implementation guidance for a specific requirement
canary implement PROJ-105 --prompt

# Create a new requirement
canary specify "Add rate limiting to API endpoints"

# Plan implementation
canary plan PROJ-106
```

## Querying Tokens

```bash
# List top priority requirements
canary list --limit 10

# List requirements needing tests
canary list --status IMPL --limit 10

# List STUB requirements for a specific aspect
canary list --aspect CLI --status STUB

# Machine-readable list (agent-friendly)
canary list --status STUB --json

# Show all tokens for a requirement
canary show PROJ-105

# Check progress on a requirement
canary status PROJ-105

# Search tokens by keyword (grep = exact field/pattern; search = fuzzy across fields)
canary grep "Auth"
canary grep "src/api"
canary search "authentication" --limit -1   # -1 = unlimited (default 20)

# List files containing tokens for a requirement
canary files PROJ-105
```

## Scanning & Verification

```bash
# Quick scan — one-line metrics on stdout
canary scan --root .

# Full scan with reports
canary scan --root . --out status.json --csv status.csv

# Strict scan (catches stale tokens >30 days)
canary scan --root . --strict

# Verify GAP_ANALYSIS.md claims
canary scan --root . --verify GAP_ANALYSIS.md --strict
# Success: CANARY_VERIFY_OK
# Failure: CANARY_VERIFY_FAIL count=N

# Project-only scan (excludes test/example IDs)
canary scan --root . --project-only
```

## Maintenance

```bash
# Update stale tokens (TESTED/BENCHED older than 30 days)
canary scan --root . --update-stale

# Rebuild token database
canary index

# Check documentation freshness
canary doc status --all

# Update documentation hashes after editing docs
canary doc update --all --stale-only

# Generate documentation coverage report
canary doc report --show-undocumented
```

## Bug Tracking

```bash
# Create a bug report
canary bug "Login fails on first attempt after signup"

# List open bugs
canary bug list --status OPEN

# Search bugs
canary grep "BUG-API"
```

## Dependency Management

```bash
# Check dependency status
canary deps check PROJ-105

# Visualize dependency graph (ascii tree)
canary deps graph PROJ-105

# Same, as a mermaid flowchart (click-through ticket links if a source is configured)
canary deps graph PROJ-105 --format mermaid

# Validate no circular dependencies
canary deps validate

# Find what depends on a requirement
canary deps reverse PROJ-105
```

## Onboarding an Existing Codebase

```bash
# Fresh-codebase adoption analysis: languages, entry points, existing tokens, MIGRATE notes
canary onboard --root .
canary onboard --json
```

## Drift Detection

```bash
# Compare tokens against reality: code-drift, staleness, doc-drift
canary drift

# Fail (exit 2) if any drift finding exists -- good for CI
canary drift --strict

# Override the staleness window (days)
canary drift --stale-days 45
```

## Upgrading Legacy Token Shapes

```bash
# Preview fixes for legacy shapes (markdown headings, unicode hyphens, unpadded IDs, etc.)
canary upgrade --root .

# Apply
canary upgrade --root . --write

# Only one rule
canary upgrade --rule status-fixed --write
```

## Ticket Sources (JIRA, etc.)

Configure sources in `.canary/project.yaml`:

```yaml
sources:
  - type: jira
    project: PROJ
    base_url: https://your-org.atlassian.net
```

```bash
# Compute a sync plan (always plan-only without JIRA_BASE_URL/JIRA_EMAIL/JIRA_API_TOKEN set)
canary ticket sync --project PROJ

# Apply against JIRA (requires credentials)
canary ticket sync --project PROJ --apply

# Chain a JIRA-driven remap into upgrade
canary upgrade --map .canary/ticket-plan.json.map.json --write
```

## MCP Server (Optional)

```bash
# Start MCP server for IDE integration
canary mcp
# Serves on http://localhost:8080/mcp

# Health check
curl http://localhost:8080/health
```

## Token Lifecycle Example

```bash
# 1. Specify requirement
canary specify "Add caching layer for API responses"
# Output: Created PROJ-110

# 2. Plan implementation
canary plan PROJ-110

# 3. Place STUB token in code
# // CANARY: REQ=PROJ-110; FEATURE="ResponseCache"; ASPECT=API; STATUS=STUB; UPDATED=2025-10-18

# 4. Implement and update to IMPL
# // CANARY: REQ=PROJ-110; FEATURE="ResponseCache"; ASPECT=API; STATUS=IMPL; UPDATED=2025-10-18

# 5. Add tests and update to TESTED
# // CANARY: REQ=PROJ-110; FEATURE="ResponseCache"; ASPECT=API; STATUS=TESTED; TEST=TestResponseCache; UPDATED=2025-10-18

# 6. Add benchmarks and update to BENCHED
# // CANARY: REQ=PROJ-110; FEATURE="ResponseCache"; ASPECT=API; STATUS=BENCHED; TEST=TestResponseCache; BENCH=BenchResponseCache; UPDATED=2025-10-18

# 7. Verify
canary scan --root .
# CANARY_SCAN tokens=1 requirements=1 STUB=0 IMPL=0 TESTED=0 BENCHED=1
```

## Filtering Patterns

```bash
# By status
canary list --status STUB          # Not started
canary list --status IMPL          # Needs tests
canary list --status TESTED        # Needs benchmarks
canary list --status BENCHED       # Complete

# By aspect
canary list --aspect API
canary list --aspect Security
canary list --aspect Engine

# By owner
canary list --owner backend-team

# By phase
canary list --phase Phase1

# Combined
canary list --status STUB --aspect API --limit 5

# Custom ordering
canary list --order-by "updated_at ASC" --limit 20  # Oldest first
```
