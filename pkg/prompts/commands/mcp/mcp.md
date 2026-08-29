# MCP Command Prompt

## Purpose
Start Model Context Protocol server for AI assistant integration.

## Task
Implement `canary mcp` to run an MCP server exposing CANARY functionality.

## Expected Behavior
```bash
# Start MCP server (default localhost:8080)
canary mcp

# Custom port and host
canary mcp --port 9000 --host 0.0.0.0
```

## MCP Tools Exposed
The server exposes 19 tools for AI assistants. Five are stubs returning placeholder
responses today (tracked in docs/GAP_ANALYSIS.md, GAP-0005): **specify**, **plan**,
**index**, **bug-create**, **gap-mark**.

### One-Call Hierarchical Context
1. **view** - Full picture of a requirement: status, files, tests, deps, spec/plan, diagrams, ticket. Use FIRST instead of separate show/status/files calls.
2. **deps** - Dependency IDs (forward or reverse) for a requirement; IDs only, follow up with view for detail.

### Core Token Management
3. **list** - List/filter tokens (default limit 20, max 100; `Total` reports the true match count)
4. **show** - Show requirement details
5. **create** - Generate token template
6. **status** - Show progress stats
7. **search** - Search by keywords (default limit 20, max 100; `-1` unlimited via the CLI)
8. **next** - Get next priority requirement

### Workflow
9. **scan** - Scan codebase (real: calls the canaryscan scanner)
10. **specify** - Create specification (stub — not yet implemented)
11. **plan** - Generate plan (stub — not yet implemented)
12. **implement** - Get implementation guidance
13. **index** - Index tokens to database (stub — not yet implemented)

### Query & Navigation
14. **files** - Find files for requirement
15. **grep** - Search by pattern

### Management
16. **prioritize** - Set priority

### Bug Tracking
17. **bug-list** - List bugs
18. **bug-create** - Create bug (stub — not yet implemented)

### Gap Analysis
19. **gap-mark** - Mark gap claims (stub — not yet implemented)

## Server Output
```
Canary MCP Server
=================
Server listening on http://localhost:8080

Available endpoints:
  GET  /health         - Health check
  POST /mcp            - MCP endpoint

Available MCP Tools (19 total):
  - view, deps
  - list, show, create, status, search, next
  - scan, specify, plan, implement, index
  - files, grep, prioritize
  - bug-list, bug-create, gap-mark

Press Ctrl+C to stop
```

## Standards
- Use `github.com/modelcontextprotocol/go-sdk/mcp`
- Expose tools with typed parameters
- JSON schema validation
- Health check endpoint
- Graceful shutdown
- Access to `.canary/canary.db`
