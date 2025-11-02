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
The server exposes 18 tools for AI assistants:

### Core Token Management
1. **list** - List/filter tokens
2. **show** - Show requirement details
3. **create** - Generate token template
4. **status** - Show progress stats
5. **search** - Search by keywords
6. **next** - Get next priority requirement

### Workflow
7. **scan** - Scan codebase
8. **specify** - Create specification
9. **plan** - Generate plan
10. **implement** - Get implementation guidance
11. **index** - Index tokens to database

### Query & Navigation
12. **files** - Find files for requirement
13. **grep** - Search by pattern

### Management
14. **prioritize** - Set priority

### Bug Tracking
15. **bug-list** - List bugs
16. **bug-create** - Create bug

### Gap Analysis
17. **gap-mark** - Mark gap claims

## Server Output
```
Canary MCP Server
=================
Server listening on http://localhost:8080

Available endpoints:
  GET  /health         - Health check
  POST /mcp            - MCP endpoint

Available MCP Tools (18 total):
  ? list, show, create, status, search, next
  ? scan, specify, plan, implement, index
  ? files, grep, prioritize
  ? bug-list, bug-create, gap-mark

Press Ctrl+C to stop
```

## Standards
- Use `github.com/modelcontextprotocol/go-sdk/mcp`
- Expose tools with typed parameters
- JSON schema validation
- Health check endpoint
- Graceful shutdown
- Access to `.canary/canary.db`
