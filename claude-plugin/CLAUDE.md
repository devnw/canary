# CANARY Requirement Tracking

This project uses CANARY requirement tracking. Load only what you need.

## Token Format

```
// CANARY: REQ=<ID>; FEATURE="Name"; ASPECT=<Aspect>; STATUS=<Status>; [TEST=TestName]; [BENCH=BenchName]; [OWNER=team]; [DOC=type:path]; [DOC_HASH=hash]; UPDATED=YYYY-MM-DD
```

## Status Progression

- **STUB** — Planned but not implemented
- **IMPL** — Implemented (token placed in code)
- **TESTED** — Implemented with tests (auto-promoted when `TEST=` field added)
- **BENCHED** — Tested with benchmarks (auto-promoted when `BENCH=` field added)

## Valid Aspects

API, CLI, Engine, Storage, Security, Docs, Wire, Planner, Decode, Encode, RoundTrip, Bench, FrontEnd, Dist

## Constitutional Principles

1. **Requirement-First** — Every feature starts with a CANARY token
2. **Test-First** — Tests written before implementation (Article IV)
3. **Evidence-Based** — Status promoted based on `TEST=`/`BENCH=` fields
4. **Simplicity** — Minimal complexity, standard library preferred
5. **Documentation Currency** — Tokens kept current with `UPDATED` field

## Minimal Context Philosophy

- For a given slash command, read only that command's instructions
- **Scan:** Use the one-line stdout (`CANARY_SCAN tokens=N requirements=M STUB=a IMPL=b TESTED=c BENCHED=d`) for metrics; do not read `status.json` unless detail is needed
- **Verify:** Use stdout (`CANARY_VERIFY_OK` or `CANARY_VERIFY_FAIL count=N`); open `GAP_ANALYSIS.md` only when fixing claims
- Reference `.canary/memory/constitution.md` only when creating or editing principles

## Project Configuration

Read `.canary/project.yaml` for the project key, name, and scanner settings. The project key determines requirement ID prefixes (e.g., `CBIN-105`).

## CLI Quick Reference

```bash
# Workflow
canary scan --root . --out status.json        # Scan for tokens
canary scan --root . --verify GAP_ANALYSIS.md --strict  # Verify claims
canary create <ID> "Name" --aspect API --status STUB    # Create token
canary next --prompt                           # Next priority with guidance
canary implement <query> --prompt              # Implementation guidance

# Query
canary list --status STUB --limit 10           # List by status
canary show <REQ-ID>                           # Show tokens for requirement
canary status <REQ-ID>                         # Progress summary
canary grep <pattern>                          # Search tokens
canary files <REQ-ID>                          # Implementation files

# Management
canary scan --root . --update-stale            # Update stale tokens
canary index                                   # Rebuild token database
canary doc status --all                        # Check documentation freshness
```

## MCP Server (Optional)

If `canary mcp` is running, MCP tools are available as supplements to CLI commands:

```bash
canary mcp  # Starts HTTP server on localhost:8080/mcp
```

MCP provides 18 tools: list, show, create, scan, specify, plan, implement, status, search, next, files, grep, index, prioritize, bug-list, bug-create, gap-mark, and more.

## Development Workflow

1. `/canary.constitution` — Establish project principles
2. `/canary.specify <description>` — Create requirement spec
3. `/canary.plan <REQ-ID>` — Generate implementation plan
4. `/canary.implement <REQ-ID>` — Get implementation guidance (test-first)
5. `/canary.scan` — Verify token status
6. `/canary.verify` — Confirm GAP_ANALYSIS claims
