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
# View first — one call for the full picture
canary view <REQ-ID>                           # status, files, tests, deps, spec, ticket
canary view <REQ-ID> --json                    # machine-readable
canary view <REQ-ID> --limit 20                # raise section limits (default 10)

# Workflow
canary scan --root . --out status.json        # Scan for tokens
canary scan --root . --verify GAP_ANALYSIS.md --strict  # Verify claims (evidence-backed)
canary verify --root . --claims GAP_ANALYSIS.md         # Verdict as one JSON line (exit 1 if unverified)
canary evidence from-go-test --commit "$(git rev-parse HEAD)" < gotest.json > evidence.json
canary evidence ingest --in evidence.json               # Merge into .canary/evidence.json
canary create <ID> "Name" --aspect API --status STUB    # Create token
canary next --prompt                           # Next priority with guidance
canary implement <query> --prompt              # Implementation guidance

# Query
canary list --status STUB --limit 10 --json    # List by status (add --json for tooling)
canary show <REQ-ID>                           # Show tokens for requirement
canary status <REQ-ID> --json                  # Progress summary
canary grep <pattern>                          # Search tokens
canary search <keywords>                       # Search across features/files/tests
canary files <REQ-ID> --json                   # Implementation files
canary deps graph <REQ-ID> --format mermaid    # Dependency tree (ascii or mermaid)
canary deps check <REQ-ID>                     # Are dependencies satisfied?
canary gap query --req-id <REQ-ID>             # Past implementation gaps

# Drift, upgrade, onboarding
canary drift --strict                          # Token-vs-reality drift (exit 2 on findings)
canary upgrade --root . --write                # Rewrite legacy token shapes
canary onboard --root .                        # Fresh-codebase adoption analysis
canary ticket sync --project <KEY>              # Reconcile tokens against JIRA (plan-only without creds)
canary ticket status [--refresh]                # Report or refresh the cached remote-status snapshot

# Management
canary scan --root . --update-stale            # Report evidence currency for stale tokens
canary index                                   # Rebuild token database
canary doc status --all                        # Check documentation freshness
```

List-shaped commands (`list`, `search`, `drift`, `view`, `onboard`, `ticket sync`, `upgrade`) default to a small `--limit` (20; `view` defaults to 10 per section) to protect agent context. `list` and `search` accept `--limit -1` for unlimited; the rest just take a larger `--limit`. Truncated output includes a hint to raise it.

Ticket sources (JIRA, etc.) are configured in `.canary/project.yaml`'s `sources:` list; without credentials, `ticket sync` still computes and writes a plan, it just never touches the network. A `destination: true` source (or, by default, the first non-flatfile source) is where new issues are created. External dependencies (ticket-sourced or `peers:`-owned) resolve to satisfied/unsatisfied/unknown from the cache; `unknown` never blocks by default, and `deps validate` treats `unsatisfied` the same way — pass `--strict-external` (on `next`/`deps validate`) to make either fail.

## MCP Server (Optional)

If `canary mcp` is running, MCP tools are available as supplements to CLI commands:

```bash
canary mcp  # Starts HTTP server on localhost:8080/mcp
```

MCP provides 19 tools. Lead with `view` and `deps` for hierarchical context in one call:
- **One-call context:** view, deps
- **Core:** list, show, create, status, search, next
- **Workflow:** scan, specify (stub — not yet implemented), plan (stub — not yet implemented), implement, index (stub — not yet implemented)
- **Query:** files, grep
- **Management:** prioritize
- **Bug tracking:** bug-list, bug-create (stub — not yet implemented)
- **Gap analysis:** gap-mark (stub — not yet implemented)

## Development Workflow

1. `/canary.constitution` — Establish project principles
2. `/canary.specify <description>` — Create requirement spec
3. `/canary.plan <REQ-ID>` — Generate implementation plan
4. `/canary.implement <REQ-ID>` — Get implementation guidance (test-first)
5. `/canary.scan` — Verify token status
6. `/canary.verify` — Confirm GAP_ANALYSIS claims
