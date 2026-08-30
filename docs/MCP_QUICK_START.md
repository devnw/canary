# MCP Quick Start Guide

## What is MCP?

Model Context Protocol (MCP) is a standardized way for AI assistants to interact with external tools and data sources. Canary now exposes its requirement tracking functionality through MCP, allowing AI assistants to help manage your project requirements.

## Quick Start (3 steps)

### 1. Build Canary with MCP Support

```bash
cd /workspace
go build -o canary ./cmd/canary
```

### 2. Start the MCP Server

```bash
./canary mcp
```

You should see:
```
Canary MCP Server
=================
Server listening on http://127.0.0.1:8080
Authentication: none (loopback dev mode)

Endpoints:
  GET  /health         - Health check
  POST /mcp            - MCP endpoint

# ... followed by the generated tool documentation ...

Press Ctrl+C to stop
```

The tool list the server prints is generated from the tool registry, so it can
never drift from what is actually registered. The same text is checked in as
[MCP_TOOLS.md](MCP_TOOLS.md); regenerate it with:

```bash
canary mcp --print-tools > docs/MCP_TOOLS.md
```

Every registered tool does real work. The five placeholders that used to be
registered -- `specify`, `plan`, `index`, `gap-mark`, and a `bug-create` that
always answered `BUG-001` -- have been removed; `bug-create` is back as a real
tool that reserves an id transactionally and persists the row.

### Binding and authentication

The server binds `127.0.0.1` by default. To bind any other interface you must
supply both a TLS certificate/key pair and a bearer token, or it refuses to
start:

```bash
CANARY_MCP_TOKEN=... canary mcp --host 0.0.0.0 --tls-cert cert.pem --tls-key key.pem
```

| Variable | Grants |
| --- | --- |
| `CANARY_MCP_TOKEN` | read and write |
| `CANARY_MCP_READ_TOKEN` | read only (mutating tools answer 403) |

With neither set, a loopback server serves every request unauthenticated. With
either set, a `Authorization: Bearer <token>` header is required even on
loopback.

`--root` selects the tree the server answers for (default: the working
directory). Tool paths are confined to it; anything resolving outside is
refused with `ROOT_ESCAPE`.

### 3. Test the Server

In another terminal:

```bash
# Test health endpoint
curl http://localhost:8080/health

# List tokens with IMPL status
curl -X POST http://localhost:8080/mcp \
  -H 'Content-Type: application/json' \
  -d '{
    "jsonrpc": "2.0",
    "id": 1,
    "method": "tools/call",
    "params": {
      "name": "list",
      "arguments": {
        "status": "IMPL",
        "limit": 5
      }
    }
  }'

# Create a token
curl -X POST http://localhost:8080/mcp \
  -H 'Content-Type: application/json' \
  -d '{
    "jsonrpc": "2.0",
    "id": 2,
    "method": "tools/call",
    "params": {
      "name": "create",
      "arguments": {
        "reqId": "TEST-001",
        "feature": "MyFeature"
      }
    }
  }'
```

## Available Tools

### view
Full picture of one requirement in a single call: status, files, tests, deps, spec/plan, diagrams, ticket. Use this FIRST instead of chaining show/status/files.
- `reqId`: Required
- `limit`: Max entries per list section (default: 10)

### deps
Dependency IDs for a requirement (forward or reverse). IDs only; follow up with `view` for detail.
- `reqId`: Required
- `direction`: `forward` (default) or `reverse`

### list
List CANARY tokens with optional filtering.
- `status`: STUB, IMPL, TESTED, BENCHED
- `aspect`: API, CLI, Engine, etc.
- `owner`: Filter by owner
- `limit`: Max results (default: 20, max: 100); response includes `total` (true match count before capping)

### show
Display all tokens for a specific requirement.
- `reqId`: Required (e.g., "CBIN-123")

### create
Generate a new CANARY token.
- `reqId`: Required (e.g., "CBIN-200")
- `feature`: Required (e.g., "UserAuth")
- `aspect`: Optional (default: "API")
- `status`: Optional (default: "IMPL")
- `owner`: Optional

### status
Show implementation progress.
- `reqId`: Required

Returns stats like:
- Total tokens
- Completion percentage
- Breakdown by status

### search
Search tokens by keywords.
- `keywords`: Required search term

### next
Get next priority requirement.
- `status`: Optional filter (default: "STUB,IMPL")
- `aspect`: Optional filter

### scan
Scan codebase for CANARY tokens (calls the real `canaryscan` scanner).
- `root`: Directory to scan (default: ".")
- `projectOnly`: Boolean filter

## Usage with AI Assistants

Once the MCP server is running, AI assistants with MCP support can automatically discover and use these tools. You can ask natural language questions like:

- "What's the status of CBIN-133?"
- "Show me all incomplete CLI requirements"
- "What should I work on next?"
- "Create a token for the new authentication feature"

The AI will translate your request into the appropriate MCP tool call and present the results in a natural way.

## Configuration

### Custom Port
```bash
./canary mcp --port 9000
```

### Listen on All Interfaces
```bash
./canary mcp --host 0.0.0.0 --port 8080
```

### Help
```bash
./canary mcp --help
```

## Troubleshooting

### Port Already in Use
```bash
# Find what's using port 8080
lsof -i :8080

# Use a different port
./canary mcp --port 9000
```

### Database Not Found
Make sure your Canary database is indexed:
```bash
./canary index
```

### Connection Issues
Check that the server is running:
```bash
curl http://localhost:8080/health
```

Should return:
```json
{"status":"healthy","time":"2025-11-01T..."}
```

## Next Steps

- Read the full documentation: [MCP_INTEGRATION.md](MCP_INTEGRATION.md)
- See implementation details: [MCP_IMPLEMENTATION_SUMMARY.md](MCP_IMPLEMENTATION_SUMMARY.md)
- Explore the MCP package: `mcp/mcp.go` and `mcp/tools.go`

## Example: Complete Workflow

```bash
# Terminal 1: Start server
./canary mcp

# Terminal 2: Use the tools
# 1. Create a new requirement
curl -X POST http://localhost:8080/mcp \
  -H 'Content-Type: application/json' \
  -d '{
    "jsonrpc": "2.0",
    "id": 1,
    "method": "tools/call",
    "params": {
      "name": "create",
      "arguments": {
        "reqId": "MYPROJ-001",
        "feature": "User Authentication",
        "aspect": "API",
        "status": "STUB"
      }
    }
  }'

# 2. List all requirements
curl -X POST http://localhost:8080/mcp \
  -H 'Content-Type: application/json' \
  -d '{
    "jsonrpc": "2.0",
    "id": 2,
    "method": "tools/call",
    "params": {
      "name": "list",
      "arguments": {
        "limit": 10
      }
    }
  }'

# 3. Get next priority item
curl -X POST http://localhost:8080/mcp \
  -H 'Content-Type: application/json' \
  -d '{
    "jsonrpc": "2.0",
    "id": 3,
    "method": "tools/call",
    "params": {
      "name": "next",
      "arguments": {}
    }
  }'

# 4. Check status of a requirement
curl -X POST http://localhost:8080/mcp \
  -H 'Content-Type: application/json' \
  -d '{
    "jsonrpc": "2.0",
    "id": 4,
    "method": "tools/call",
    "params": {
      "name": "status",
      "arguments": {
        "reqId": "MYPROJ-001"
      }
    }
  }'
```

## Tips

1. **Keep server running**: The MCP server should run continuously while you work
2. **Use health checks**: Monitor server health at `/health`
3. **Set appropriate limits**: Use the `limit` parameter to control result sizes
4. **Filter effectively**: Combine status and aspect filters for precise queries
5. **Check logs**: Server logs to stdout for debugging

## Security Note

By default, the MCP server:
- Listens on localhost only (not accessible from network)
- Has no authentication (safe for local development)
- Reads/writes to local Canary database

For production use, consider:
- Adding authentication (API keys)
- Using HTTPS
- Implementing rate limiting
- Running behind a reverse proxy

See [MCP_INTEGRATION.md](MCP_INTEGRATION.md#security-considerations) for details.
