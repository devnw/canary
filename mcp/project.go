// Copyright (c) 2025 by Developer Network.
//
// For more details, see the LICENSE file in the root directory of this
// source code repository or contact Developer Network at info@devnw.com.

package mcp

// CANARY: REQ=CBIN-307; FEATURE="ProjectScopeCLI"; ASPECT=API; STATUS=IMPL; UPDATED=2026-08-30

// mcpProjectID resolves the project every MCP tool scopes its queries to.
//
// It mirrors the CLI's read-side rule: unscoped, so the storage layer answers
// from whatever the database actually holds and refuses when that would be
// ambiguous. The MCP surface has no per-request project selector yet -- that
// is Task 8's job -- and inventing a scope here would hide rows rather than
// disambiguate them.
func mcpProjectID() string {
	return ""
}
