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
Server listening on http://localhost:8080

Available MCP Tools:
  - list     - List CANARY tokens
  - show     - Show requirement details
  - create   - Create new token
  - status   - Show progress
  - search   - Search tokens
  - next     - Get next priority
  - scan     - Scan codebase

Press Ctrl+C to stop
```

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

### list
List CANARY tokens with optional filtering.
- `status`: STUB, IMPL, TESTED, BENCHED
- `aspect`: API, CLI, Engine, etc.
- `owner`: Filter by owner
- `limit`: Max results (default: 100)

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
Scan codebase (placeholder).
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
