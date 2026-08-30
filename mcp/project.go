// Copyright (c) 2025 by Developer Network.
//
// For more details, see the LICENSE file in the root directory of this
// source code repository or contact Developer Network at info@devnw.com.

package mcp

import "devnw.dev/canary/pkg/config"

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

// mcpWriteProjectID resolves the project an MCP tool *writes* rows under.
//
// The read-side rule above does not carry over to a writer. An unscoped
// UPDATE is not a broader question, it is an edit to every project sharing
// the database, so the storage layer refuses "" outright. This mirrors the
// CLI's WriteProjectID exactly: the configured project.key, falling back to
// config's "default" -- the same value migration 000007 backfills onto
// pre-scoping rows.
func mcpWriteProjectID() string {
	cfg, err := config.Load(".")
	if err != nil {
		// An unreadable project.yaml is not a reason to widen the write.
		// ProjectID's nil receiver yields config's own "default", which is
		// still a scope -- unlike "", which is every project at once.
		cfg = nil
	}
	return cfg.ProjectID()
}
