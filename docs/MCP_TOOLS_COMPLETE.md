# Complete MCP Tools Reference

## Overview

The Canary MCP server provides **19 tools** covering all major Canary functionality. Five are stubs returning placeholder responses today (tracked in `docs/GAP_ANALYSIS.md`, GAP-0005): `specify`, `plan`, `index`, `bug-create`, `gap-mark`.

## Tool Categories

### One-Call Hierarchical Context (2 tools)
- `view` - Full picture of a requirement: status, files, tests, deps, spec/plan, diagrams, ticket. Use FIRST instead of separate show/status/files calls.
- `deps` - Dependency IDs (forward or reverse) for a requirement; IDs only, follow up with `view` for detail.

### Core Token Management (6 tools)
- `list` - List CANARY tokens with filtering
- `show` - Show details for a specific requirement
- `create` - Create a new CANARY token
- `status` - Show implementation progress
- `search` - Search tokens by keywords
- `next` - Get next priority requirement

### Workflow & Development (5 tools)
- `scan` - Scan codebase for CANARY tokens
- `specify` - Create requirement specification (stub — not yet implemented)
- `plan` - Generate implementation plan (stub — not yet implemented)
- `implement` - Get implementation guidance
- `index` - Index codebase tokens into database (stub — not yet implemented)

### Query & Navigation (2 tools)
- `files` - Find files containing requirement tokens
- `grep` - Search tokens by pattern

### Management (1 tool)
- `prioritize` - Set requirement priority

### Bug Tracking (2 tools)
- `bug-list` - List bug tracking tokens
- `bug-create` - Create new bug token (stub — not yet implemented)

### Gap Analysis (1 tool)
- `gap-mark` - Mark gap claims as helpful/unhelpful (stub — not yet implemented)

## Detailed Tool Specifications

### 1. list
**Purpose**: List CANARY tokens with optional filtering

**Parameters**:
- `status` (string, optional): Filter by status (STUB, IMPL, TESTED, BENCHED)
- `aspect` (string, optional): Filter by aspect (API, CLI, Engine, etc.)
- `owner` (string, optional): Filter by owner
- `limit` (number, optional): Maximum results (default: 20, max: 100 — mirrors `search`'s cap)

**Returns**:
- `tokens`: Array of matching tokens
- `count`: Number of tokens in this response (after capping)
- `total`: True match count before capping

**Example**:
```json
{
  "name": "list",
  "arguments": {
    "status": "IMPL",
    "aspect": "CLI",
    "limit": 10
  }
}
```

### 2. show
**Purpose**: Display all CANARY tokens for a specific requirement ID

**Parameters**:
- `reqId` (string, required): Requirement ID (e.g., "CBIN-123")

**Returns**:
- `reqId`: The requested requirement ID
- `tokens`: Array of all tokens for this requirement
- `count`: Number of tokens

**Example**:
```json
{
  "name": "show",
  "arguments": {
    "reqId": "CBIN-133"
  }
}
```

### 3. create
**Purpose**: Generate a new CANARY token template

**Parameters**:
- `reqId` (string, required): Requirement ID
- `feature` (string, required): Feature name
- `aspect` (string, optional): Aspect (default: "API")
- `status` (string, optional): Status (default: "IMPL")
- `owner` (string, optional): Owner/assignee

**Returns**:
- `token`: Formatted CANARY token string
- `reqId`, `feature`, `aspect`, `status`: Echo of parameters

**Example**:
```json
{
  "name": "create",
  "arguments": {
    "reqId": "PROJ-001",
    "feature": "UserAuthentication",
    "aspect": "API",
    "status": "STUB"
  }
}
```

### 4. status
**Purpose**: Show implementation progress for a requirement

**Parameters**:
- `reqId` (string, required): Requirement ID

**Returns**:
- `reqId`: The requirement ID
- `stats`: Object with counts (total, stub, impl, tested, benched, completed)
- `completionPct`: Percentage complete (0-100)
- `tokens`: Array of tokens

**Example**:
```json
{
  "name": "status",
  "arguments": {
    "reqId": "CBIN-133"
  }
}
```

**Response**:
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
  "completionPct": 70
}
```

### 5. search
**Purpose**: Search CANARY tokens by keywords

**Parameters**:
- `keywords` (string, required): Search keywords

**Returns**:
- `keywords`: The search term
- `tokens`: Array of matching tokens
- `count`: Number of matches

**Example**:
```json
{
  "name": "search",
  "arguments": {
    "keywords": "authentication"
  }
}
```

### 6. next
**Purpose**: Identify next highest priority unimplemented requirement

**Parameters**:
- `status` (string, optional): Filter by status (default: "STUB,IMPL")
- `aspect` (string, optional): Filter by aspect

**Returns**:
- `token`: The next priority token (or null if none)
- `reqId`, `feature`, `aspect`, `status`, `priority`: Token details
- `message`: Status message

**Example**:
```json
{
  "name": "next",
  "arguments": {
    "status": "STUB"
  }
}
```

### 7. scan
**Purpose**: Scan codebase for CANARY tokens

**Parameters**:
- `root` (string, optional): Root directory (default: ".")
- `projectOnly` (boolean, optional): Filter by project ID pattern

**Returns**:
- `message`: Status message
- `root`: Scanned directory

**Note**: Fully implemented — calls the real `canaryscan` scanner (`pkg/canaryscan`), the same engine the `canary scan` CLI command uses.

### 8. specify
**Purpose**: Create a requirement specification

**Parameters**:
- `description` (string, required): Feature description
- `aspect` (string, optional): Aspect (default: "Engine")

**Returns**:
- `message`: Status message
- `description`: Echo of description
- `aspect`: The aspect
- `specPath`: Path to spec file

**Example**:
```json
{
  "name": "specify",
  "arguments": {
    "description": "User authentication with OAuth2 support",
    "aspect": "API"
  }
}
```

### 9. plan
**Purpose**: Generate implementation plan for a requirement

**Parameters**:
- `reqId` (string, required): Requirement ID
- `techStack` (string, optional): Technology stack description

**Returns**:
- `message`: Status message
- `reqId`: The requirement ID
- `techStack`: Technology stack
- `planPath`: Path to plan file

**Example**:
```json
{
  "name": "plan",
  "arguments": {
    "reqId": "CBIN-200",
    "techStack": "Go with sqlx for database"
  }
}
```

### 10. implement
**Purpose**: Get implementation guidance for a requirement

**Parameters**:
- `reqId` (string, required): Requirement ID to implement

**Returns**:
- `message`: Status message
- `reqId`: The requirement ID
- `guidance`: Implementation guidance text
- `specPath`, `planPath`: Paths to related files
- `hasSpec`, `hasPlan`: Boolean flags
- `tokenCount`: Number of tokens
- `currentPhase`: Current development phase

**Example**:
```json
{
  "name": "implement",
  "arguments": {
    "reqId": "CBIN-200"
  }
}
```

**Response**:
```json
{
  "reqId": "CBIN-200",
  "guidance": "Found 5 tokens for CBIN-200. Current phase: Implementation",
  "hasSpec": true,
  "hasPlan": true,
  "tokenCount": 5,
  "currentPhase": "Implementation"
}
```

### 11. index
**Purpose**: Index codebase tokens into database

**Parameters**:
- `root` (string, optional): Root directory (default: ".")
- `rebuild` (boolean, optional): Force rebuild of index

**Returns**:
- `message`: Status message
- `tokensIndexed`: Number of tokens indexed
- `filesScanned`: Number of files scanned

**Note**: Currently placeholder - requires integration with index command.

### 12. files
**Purpose**: Find files containing tokens for a requirement

**Parameters**:
- `reqId` (string, required): Requirement ID

**Returns**:
- `reqId`: The requirement ID
- `files`: Array of file paths
- `fileCount`: Number of files

**Example**:
```json
{
  "name": "files",
  "arguments": {
    "reqId": "CBIN-133"
  }
}
```

**Response**:
```json
{
  "reqId": "CBIN-133",
  "files": [
    "pkg/storage/storage.go",
    "cli/list/list.go",
    "cli/show/show.go"
  ],
  "fileCount": 3
}
```

### 13. grep
**Purpose**: Search tokens by pattern in specific fields

**Parameters**:
- `pattern` (string, required): Pattern to search for
- `field` (string, optional): Field to search (default: "all")

**Returns**:
- `pattern`: The search pattern
- `field`: Field searched
- `tokens`: Matching tokens
- `count`: Number of matches

**Example**:
```json
{
  "name": "grep",
  "arguments": {
    "pattern": "database",
    "field": "feature"
  }
}
```

### 14. prioritize
**Purpose**: Set priority level for a requirement

**Parameters**:
- `reqId` (string, required): Requirement ID
- `priority` (number, required): Priority level (lower is higher priority)

**Returns**:
- `message`: Status message
- `reqId`: The requirement ID
- `priority`: New priority level
- `updated`: Number of tokens updated

**Example**:
```json
{
  "name": "prioritize",
  "arguments": {
    "reqId": "CBIN-200",
    "priority": 1
  }
}
```

### 15. bug-list
**Purpose**: List bug tracking tokens

**Parameters**:
- `status` (string, optional): Filter by status (OPEN, INVESTIGATING, FIXED, WONTFIX)
- `severity` (string, optional): Filter by severity (CRITICAL, HIGH, MEDIUM, LOW)
- `limit` (number, optional): Maximum results (default: 100)

**Returns**:
- `bugs`: Array of bug tokens
- `count`: Number of bugs

**Example**:
```json
{
  "name": "bug-list",
  "arguments": {
    "status": "OPEN",
    "severity": "HIGH"
  }
}
```

### 16. bug-create
**Purpose**: Create a new bug tracking token

**Parameters**:
- `title` (string, required): Bug title/description
- `severity` (string, optional): Severity (default: "MEDIUM")
- `component` (string, optional): Affected component
- `description` (string, optional): Detailed description

**Returns**:
- `token`: Formatted bug token
- `bugId`: Generated bug ID
- `title`: Bug title
- `severity`: Severity level

**Example**:
```json
{
  "name": "bug-create",
  "arguments": {
    "title": "Database connection leak in production",
    "severity": "CRITICAL",
    "component": "Storage"
  }
}
```

### 17. gap-mark
**Purpose**: Mark gap analysis claim as helpful or unhelpful

**Parameters**:
- `claimId` (string, required): Gap analysis claim ID
- `judgment` (string, required): Judgment (helpful, unhelpful, unclear)
- `reason` (string, optional): Reason for judgment

**Returns**:
- `message`: Status message
- `claimId`: The claim ID
- `judgment`: The judgment

**Example**:
```json
{
  "name": "gap-mark",
  "arguments": {
    "claimId": "GAP-CLAIM-42",
    "judgment": "helpful",
    "reason": "Correctly identified missing error handling"
  }
}
```

### 18. view
**Purpose**: Full picture of one requirement in a single, bounded call -- status, files, tests, deps, spec/plan, diagrams, ticket URL. Use this FIRST instead of separate show/status/files calls.

**Parameters**:
- `reqId` (string, required): Requirement ID (e.g., "CBIN-105")
- `limit` (number, optional): Max entries per list section (default: 10)

**Returns**: the full `view.View` struct -- statuses, completion %, features, files (+ total), tests, benches, depends_on, blocks, related_to, spec_path, plan_path, diagrams (+ total), migrate_notes (+ total), drifted/drift_reason, ticket_url.

**Example**:
```json
{
  "name": "view",
  "arguments": {
    "reqId": "CBIN-204",
    "limit": 20
  }
}
```

### 19. deps
**Purpose**: Dependency IDs for a requirement, forward (what it depends on) or reverse (what depends on it). IDs only -- follow up with `view` for detail on any returned ID.

**Parameters**:
- `reqId` (string, required): Requirement ID
- `direction` (string, optional): `forward` (default) or `reverse`

**Returns**:
- `reqId`, `direction`: Echo of parameters
- `dependencies`: Array of requirement IDs
- `count`: Number of dependencies

**Example**:
```json
{
  "name": "deps",
  "arguments": {
    "reqId": "CBIN-147",
    "direction": "reverse"
  }
}
```

## Implementation Status

### Fully Functional (12 tools)
- view, deps
- list, show, create, status, search, next
- scan
- files, grep, prioritize

### Functional with Database Access (2 tools)
- bug-list (requires indexed database)
- implement (reads from database and filesystem)

### Stub Implementation (5 tools)
- specify (needs specification generation)
- plan (needs plan generation)
- index (needs indexing logic)
- bug-create (needs collision-safe ID generation; currently always returns BUG-001)
- gap-mark (needs gap database integration)

## Usage Examples

### AI Assistant Workflow

```
User: "What should I work on next?"

Assistant -> MCP:
{
  "name": "next",
  "arguments": {}
}

Response:
{
  "reqId": "CBIN-142",
  "feature": "Database Migration",
  "priority": 1,
  "status": "STUB"
}
User gets: "You should work on CBIN-142: Database Migration (Priority 1)"

---

User: "Show me all files that need changes for CBIN-142"

Assistant → MCP:
{
  "name": "files",
  "arguments": {
    "reqId": "CBIN-142"
  }
}

Response:
{
  "files": ["pkg/storage/migrations.go", "cmd/canary/main.go"],
  "fileCount": 2
}

---

User: "Create a token for new feature: user profile API"

Assistant → MCP:
{
  "name": "create",
  "arguments": {
    "reqId": "PROJ-050",
    "feature": "UserProfileAPI",
    "aspect": "API",
    "status": "STUB"
  }
}

Response:
{
  "token": "// CANARY: REQ=PROJ-050; FEATURE=\"UserProfileAPI\"; ASPECT=API; STATUS=STUB; UPDATED=2025-11-01"
}
```

## Coverage Matrix

| CLI Command | MCP Tool(s) | Status |
|-------------|-------------|--------|
| view | view | ✅ Full |
| deps | deps | ✅ Full |
| list | list | ✅ Full |
| show | show | ✅ Full |
| create | create | ✅ Full |
| status | status | ✅ Full |
| search | search | ✅ Full |
| next | next | ✅ Full |
| scan | scan | ✅ Full |
| specify | specify | 🚧 Stub |
| plan | plan | 🚧 Stub |
| implement | implement | ⚠️ Partial |
| index | index | 🚧 Stub |
| files | files | ✅ Full |
| grep | grep | ✅ Full |
| prioritize | prioritize | ✅ Full |
| bug (list) | bug-list | ⚠️ DB-dependent |
| bug (create) | bug-create | 🚧 Stub (always returns BUG-001) |
| gap (mark) | gap-mark | 🚧 Stub |
| checkpoint | - | Not implemented |
| constitution | - | Not implemented |
| doc | - | Not implemented |
| drift | - | Not implemented |
| upgrade | - | Not implemented |
| onboard | - | Not implemented |
| ticket | - | Not implemented |
| migrate | - | Not implemented |
| project | - | Not implemented |

**Legend**:
- ✅ Full: Complete implementation, ready for use
- ⚠️ Partial/DB-dependent: Works with database setup
- 🚧 Stub: Placeholder implementation, needs full integration
- Not implemented: No MCP tool yet

## Server Information

### Starting the Server
```bash
# Default (localhost:8080)
canary mcp

# Custom port
canary mcp --port 9000

# Custom host and port
canary mcp --host 0.0.0.0 --port 8080
```

### Health Check
```bash
curl http://localhost:8080/health
```

Response:
```json
{
  "status": "healthy",
  "time": "2025-11-01T19:30:00Z"
}
```

### Tool Discovery
AI assistants automatically discover all 19 tools through the MCP protocol. No manual configuration needed.

## Testing

### Unit Tests
```bash
go test ./mcp/... -v
```

All 19 tool handlers have basic tests covering:
- Parameter validation
- Handler function signatures
- Basic error handling
- Response format

### Integration Testing
```bash
# Start server
canary mcp --port 9000 &

# Test a tool
curl -X POST http://localhost:9000/mcp \
  -H 'Content-Type: application/json' \
  -d '{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"create","arguments":{"reqId":"TEST-001","feature":"Test"}}}'

# Cleanup
pkill -f "canary mcp"
```

## Future Enhancements

### Phase 2: Complete Implementation
- [ ] Full scan integration with codebase scanner
- [ ] Specification generation with AI
- [ ] Plan generation with templates
- [ ] Indexing integration
- [ ] Gap analysis database integration

### Phase 3: Advanced Features
- [ ] Checkpoint management tools
- [ ] Constitution tools
- [ ] Documentation tracking tools
- [ ] Dependency analysis tools
- [ ] Migration tools
- [ ] Project management tools

### Phase 4: Enhanced Capabilities
- [ ] Streaming responses for long operations
- [ ] Progress updates via WebSocket
- [ ] Batch operations
- [ ] Transaction support
- [ ] Undo/redo capabilities

## Architecture

### Tool Design Pattern
Each tool follows this structure:

1. **Parameter Struct**: Defines input with JSON schema tags
2. **Result Struct**: Defines output structure
3. **Handler Function**: Implements business logic
4. **Registration**: Added to server in `mcp.go`

### Example:
```go
// 1. Parameters
type MyToolParams struct {
    Field string `json:"field" jsonschema:"description:Field description,required"`
}

// 2. Result
type MyToolResult struct {
    Data string `json:"data"`
}

// 3. Handler
func handleMyTool(ctx context.Context, req *mcp.CallToolRequest, 
                  params *MyToolParams) (*mcp.CallToolResult, *MyToolResult, error) {
    // Implementation
    return &mcp.CallToolResult{
        Content: []mcp.Content{&mcp.TextContent{Text: "Success"}},
    }, &MyToolResult{Data: "result"}, nil
}

// 4. Registration (in mcp.go)
mcp.AddTool(server, &mcp.Tool{
    Name: "my-tool",
    Description: "Does something",
}, handleMyTool)
```

## Files

```
mcp/
|-- mcp.go              # Server setup and tool registration
|-- tools.go            # Core tools (list, show, create, status, search, next, scan)
|-- tools_extended.go   # Extended tools (specify, plan, index, implement, files, grep,
|                       #   prioritize, bug-list, bug-create, gap-mark, view, deps)
|-- mcp_test.go         # Unit tests for all tools
```

## Summary

The Canary MCP implementation provides **19 tools** covering:
- ✅ 12 fully functional tools ready for immediate use
- ⚠️ 2 database-dependent tools (work with indexed database)
- 🚧 5 stub tools (need full integration)

This provides AI assistants with extensive access to Canary's requirement tracking system through a standardized MCP interface.
