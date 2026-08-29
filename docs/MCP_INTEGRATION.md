# MCP (Model Context Protocol) Integration

## Overview

Canary now supports the Model Context Protocol (MCP), allowing AI assistants to interact with Canary requirement tracking through a standardized interface.

## Usage

### Starting the MCP Server

```bash
# Start with default settings (localhost:8080)
canary mcp

# Start on custom host/port
canary mcp --host 0.0.0.0 --port 9000

# Start on specific interface
canary mcp --host 127.0.0.1 --port 3000
```

### Available Endpoints

- **POST /mcp** - Main MCP endpoint for tool calls
- **GET /health** - Health check endpoint

## Available MCP Tools

### 1. list - List CANARY Tokens

List tokens with optional filtering.

**Parameters:**
- `status` (string, optional): Filter by status (STUB, IMPL, TESTED, BENCHED)
- `aspect` (string, optional): Filter by aspect (API, CLI, Engine, etc.)
- `owner` (string, optional): Filter by owner
- `limit` (number, optional): Maximum number of results (default: 100)

**Example Request:**
```json
{
  "jsonrpc": "2.0",
  "id": 1,
  "method": "tools/call",
  "params": {
    "name": "list",
    "arguments": {
      "status": "IMPL",
      "aspect": "CLI",
      "limit": 10
    }
  }
}
```

**Example Response:**
```json
{
  "tokens": [...],
  "count": 5
}
```

### 2. show - Show Requirement Details

Display all tokens for a specific requirement ID.

**Parameters:**
- `reqId` (string, required): Requirement ID (e.g., CBIN-123)

**Example Request:**
```json
{
  "jsonrpc": "2.0",
  "id": 2,
  "method": "tools/call",
  "params": {
    "name": "show",
    "arguments": {
      "reqId": "CBIN-133"
    }
  }
}
```

### 3. create - Create CANARY Token

Generate a new CANARY token template.

**Parameters:**
- `reqId` (string, required): Requirement ID (e.g., CBIN-CLI-105)
- `feature` (string, required): Feature name
- `aspect` (string, optional): Aspect (default: API)
- `status` (string, optional): Status (default: IMPL)
- `owner` (string, optional): Owner/assignee

**Example Request:**
```json
{
  "jsonrpc": "2.0",
  "id": 3,
  "method": "tools/call",
  "params": {
    "name": "create",
    "arguments": {
      "reqId": "CBIN-200",
      "feature": "UserProfile",
      "aspect": "API",
      "status": "IMPL"
    }
  }
}
```

### 4. status - Show Implementation Status

Show implementation progress for a requirement.

**Parameters:**
- `reqId` (string, required): Requirement ID

**Example Request:**
```json
{
  "jsonrpc": "2.0",
  "id": 4,
  "method": "tools/call",
  "params": {
    "name": "status",
    "arguments": {
      "reqId": "CBIN-133"
    }
  }
}
```

**Example Response:**
```json
{
  "reqId": "CBIN-133",
  "stats": {
    "total": 10,
    "stub": 0,
    "impl": 3,
    "tested": 5,
    "benched": 2,
    "completed": 7
  },
  "completionPct": 70,
  "tokens": [...]
}
```

### 5. search - Search Tokens

Search CANARY tokens by keywords.

**Parameters:**
- `keywords` (string, required): Search keywords

**Example Request:**
```json
{
  "jsonrpc": "2.0",
  "id": 5,
  "method": "tools/call",
  "params": {
    "name": "search",
    "arguments": {
      "keywords": "authentication"
    }
  }
}
```

### 6. next - Get Next Priority

Identify the next highest priority unimplemented requirement.

**Parameters:**
- `status` (string, optional): Filter by status (default: STUB,IMPL)
- `aspect` (string, optional): Filter by aspect

**Example Request:**
```json
{
  "jsonrpc": "2.0",
  "id": 6,
  "method": "tools/call",
  "params": {
    "name": "next",
    "arguments": {
      "status": "STUB"
    }
  }
}
```

### 7. scan - Scan Codebase

Scan codebase for CANARY tokens (fully implemented — calls the real `canaryscan` scanner, the same engine the `canary scan` CLI command uses).

**Parameters:**
- `root` (string, optional): Root directory to scan (default: .)
- `projectOnly` (boolean, optional): Filter by project ID pattern

## Testing the MCP Server

### Using curl

```bash
# Start the server in one terminal
canary mcp

# In another terminal, test the health endpoint
curl http://localhost:8080/health

# Call the list tool
curl -H 'Content-Type: application/json' \
     -d '{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"list","arguments":{"status":"IMPL"}}}' \
     http://localhost:8080/mcp

# Create a token
curl -H 'Content-Type: application/json' \
     -d '{"jsonrpc":"2.0","id":2,"method":"tools/call","params":{"name":"create","arguments":{"reqId":"TEST-001","feature":"TestFeature"}}}' \
     http://localhost:8080/mcp

# Get status
curl -H 'Content-Type: application/json' \
     -d '{"jsonrpc":"2.0","id":3,"method":"tools/call","params":{"name":"status","arguments":{"reqId":"CBIN-133"}}}' \
     http://localhost:8080/mcp
```

### Using MCP Client SDKs

AI assistants with MCP support can connect to the server and use these tools directly through their native MCP client implementations.

## Architecture

### MCP Package Structure

```
mcp/
|-- mcp.go            # Server initialization and HTTP handler
|-- tools.go          # Core MCP tool implementations
|-- tools_extended.go # Extended MCP tool implementations (incl. view, deps)
|-- mcp_test.go       # Unit tests for MCP tools
```

### Tool Implementation Pattern

Each MCP tool follows this pattern:

```go
// 1. Define input parameters struct
type ToolParams struct {
    Field string `json:"field" jsonschema:"description=Field description,required"`
}

// 2. Define output result struct
type ToolResult struct {
    Data string `json:"data"`
}

// 3. Implement handler function
func handleTool(
    ctx context.Context,
    req *mcp.CallToolRequest,
    params *ToolParams,
) (*mcp.CallToolResult, *ToolResult, error) {
    // Validate params
    // Execute logic
    // Return result
    return &mcp.CallToolResult{
        Content: []mcp.Content{
            &mcp.TextContent{Text: "Success message"},
        },
    }, &ToolResult{Data: "..."}, nil
}

// 4. Register with server
mcp.AddTool(server, &mcp.Tool{
    Name:        "tool-name",
    Description: "Tool description",
}, handleTool)
```

## Integration with Existing CLI

The MCP server leverages existing Canary infrastructure:

- **Database Access**: Uses `pkg/storage` package directly
- **Token Management**: Reuses storage.Token structures
- **Filtering**: Applies same filter logic as CLI commands
- **Business Logic**: Calls existing internal packages

### Benefits of This Approach:

1. **No Code Duplication**: MCP tools reuse existing logic
2. **Consistency**: Same behavior as CLI commands
3. **Maintainability**: Single source of truth for business logic
4. **Testing**: Existing tests cover the underlying functionality

## Future Enhancements

### Planned Tools:

- [x] `view` - Full requirement context in one call (status, files, tests, deps, spec, ticket)
- [x] `deps` - Dependency IDs, forward or reverse
- [ ] `specify` - Create requirement specifications (stub exists; real generation pending)
- [ ] `plan` - Generate implementation plans (stub exists; real generation pending)
- [x] `implement` - Generate implementation guidance
- [ ] `gap-mark` - Record gap analysis entries (stub exists; database integration pending)
- [ ] `gap-query` - Query gap analysis
- [ ] `bug-create` - Create bug tracking tokens (stub exists; collision-safe ID generation pending)
- [x] `bug-list` - List bug tokens

### Advanced Features:

- [ ] **Streaming Responses**: Long-running operations (scan, implement)
- [ ] **Progress Updates**: Real-time progress for scans
- [ ] **File Operations**: Read/write spec and plan files
- [ ] **Authentication**: API key or token-based auth
- [ ] **Rate Limiting**: Prevent abuse
- [ ] **Caching**: Cache frequently accessed data
- [ ] **WebSocket Support**: For real-time updates

## Security Considerations

### Current Implementation:

- Server binds to localhost by default
- No authentication required (localhost-only access)
- Read-only operations are safe
- Write operations modify local database

### Production Recommendations:

1. **Use Authentication**: Add API key validation
2. **Enable HTTPS**: Use TLS certificates
3. **Rate Limiting**: Prevent abuse
4. **Access Control**: Restrict which tools can be called
5. **Input Validation**: Validate all parameters strictly
6. **Audit Logging**: Log all tool calls

### Example with Authentication:

```go
// Add middleware for API key validation
func authMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        apiKey := r.Header.Get("X-API-Key")
        if apiKey != expectedKey {
            http.Error(w, "Unauthorized", http.StatusUnauthorized)
            return
        }
        next.ServeHTTP(w, r)
    })
}

mux.Handle("/mcp", authMiddleware(handler))
```

## Troubleshooting

### Server Won't Start

Check if port is already in use:
```bash
lsof -i :8080
```

Use a different port:
```bash
canary mcp --port 9000
```

### Database Not Found

The MCP server requires an indexed database. Run:
```bash
canary index
```

### Connection Refused

Ensure the server is running and listening on the correct host/port:
```bash
curl http://localhost:8080/health
```

## Development

### Adding New Tools

1. **Define parameter and result structs** in `tools.go`
2. **Implement handler function** following the pattern
3. **Register tool** in `mcp.go` using `mcp.AddTool()`
4. **Add tests** in `mcp_test.go`
5. **Update documentation** in this file

### Testing

```bash
# Run MCP tests
go test ./mcp/...

# Run all tests
go test ./...

# Test with race detector
go test -race ./mcp/...
```

## References

- [Model Context Protocol Specification](https://spec.modelcontextprotocol.io/)
- [MCP Go SDK](https://github.com/modelcontextprotocol/go-sdk)
- [Canary Documentation](../README.md)
