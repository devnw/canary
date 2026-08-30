# Canary MCP Tools

<!-- GENERATED FILE. Do not edit by hand.
     Regenerate with: canary mcp --print-tools > docs/MCP_TOOLS.md -->

The Canary MCP server exposes 15 tools. Start it with `canary mcp`; it binds
127.0.0.1:8080 by default and serves the MCP endpoint at `/mcp` and a health
check at `/health`.

Every tool below has a checkable postcondition. Tools marked **mutate** write
to the index; when authentication is configured they require the read+write
token (`CANARY_MCP_TOKEN`) and are refused to a read-only token
(`CANARY_MCP_READ_TOKEN`) with 403.

## Summary

| Tool | Scope |
| --- | --- |
| `list` | read |
| `show` | read |
| `create` | read |
| `status` | read |
| `search` | read |
| `next` | read |
| `view` | read |
| `deps` | read |
| `scan` | read |
| `implement` | read |
| `files` | read |
| `grep` | read |
| `prioritize` | mutate |
| `bug-list` | read |
| `bug-create` | mutate |

## Tools

### `list` (read)

List CANARY tokens with optional filtering (default limit 20, max 100 to reduce context; Total reports the true match count)

### `show` (read)

Display all CANARY tokens for a specific requirement ID

### `create` (read)

Generate a new CANARY token template (returns the token text to paste into source; writes nothing)

### `status` (read)

Show implementation progress for a requirement

### `search` (read)

Search CANARY tokens by keywords

### `next` (read)

Identify the next highest-priority actionable requirement, applying the same dependency rule as the `canary next` CLI: a local dependency is complete only when evidence at the current commit proves it, and an external (ticket-source) dependency whose state cannot be resolved blocks selection.

### `view` (read)

Full picture of one requirement: status, files, tests, deps, spec/plan, diagrams, ticket URL. Use this FIRST instead of separate show/status/files calls.

### `deps` (read)

Dependency IDs for a requirement (forward or reverse). IDs only; follow up with view for detail.

### `scan` (read)

Scan the server root for CANARY tokens, honoring .canaryignore and the configured ticket sources. Paths outside the root are refused.

### `implement` (read)

Report implementation state for a requirement: token count, current phase, and whether its spec and plan exist

### `files` (read)

Find files containing tokens for a requirement

### `grep` (read)

Search tokens by pattern in specific fields

### `prioritize` (mutate)

Set the priority level for a requirement (writes to the index)

### `bug-list` (read)

List bug tracking tokens

### `bug-create` (mutate)

Create a bug tracking token with a transactionally reserved BUG-<ASPECT>-NNN id and persist it to the index
