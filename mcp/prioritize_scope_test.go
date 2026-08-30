// Copyright (c) 2025 by Developer Network.
//
// For more details, see the LICENSE file in the root directory of this
// source code repository or contact Developer Network at info@devnw.com.

package mcp

import (
	"context"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"devnw.dev/canary/pkg/storage"
)

// CANARY: REQ=CBIN-308; FEATURE="ScopedWrites"; ASPECT=API; STATUS=TESTED; TEST=TestPrioritizeWritesOneProject; UPDATED=2026-08-30

// seedPriority writes one token for reqID under projectID at the given
// priority.
func seedPriority(t *testing.T, db *storage.DB, projectID, reqID string, priority int) {
	t.Helper()
	err := db.UpsertToken(&storage.Token{
		ReqID: reqID, Feature: "F", Aspect: "API", Status: "STUB",
		FilePath: projectID + "/a.go", LineNumber: 1, Priority: priority,
		UpdatedAt: "2026-01-01", RawToken: "x",
		IndexedAt: "2026-01-01T00:00:00Z", ProjectID: projectID,
	})
	if err != nil {
		t.Fatalf("seed %s/%s: %v", projectID, reqID, err)
	}
}

// TestPrioritizeWritesOneProject pins the MCP half of the unscoped-write
// finding: the prioritize tool passed the read-side "" scope into an UPDATE,
// which in a database holding two projects rewrote both of them. It must now
// resolve the same project id the CLI writes under -- the configured key, or
// "default" -- and leave every other project's rows untouched.
func TestPrioritizeWritesOneProject(t *testing.T) {
	chdirTemp(t)

	db, err := storage.OpenRW(".canary/canary.db")
	if err != nil {
		t.Fatalf("OpenRW: %v", err)
	}
	seedPriority(t, db, "default", "TEST-001", 5)
	seedPriority(t, db, "other", "TEST-001", 5)
	if cerr := db.Close(); cerr != nil {
		t.Fatalf("close: %v", cerr)
	}

	_, res, err := testDeps.handlePrioritize(context.Background(), &mcp.CallToolRequest{},
		&PrioritizeParams{ReqID: "TEST-001", Priority: 1})
	if err != nil {
		t.Fatalf("handlePrioritize: %v", err)
	}
	if res == nil || res.Updated != 1 {
		t.Fatalf("want exactly one row updated, got %+v", res)
	}

	check, err := storage.OpenRO(".canary/canary.db")
	if err != nil {
		t.Fatalf("OpenRO: %v", err)
	}
	defer func() { _ = check.Close() }()

	mine, err := check.GetTokensByReqID("default", "TEST-001")
	if err != nil || len(mine) != 1 || mine[0].Priority != 1 {
		t.Fatalf("the write did not land on its own project: %v %+v", err, mine)
	}
	theirs, err := check.GetTokensByReqID("other", "TEST-001")
	if err != nil || len(theirs) != 1 || theirs[0].Priority != 5 {
		t.Fatalf("the write reached a sibling project: %v %+v", err, theirs)
	}
}

// TestMCPWriteProjectIDIsNeverEmpty proves the writer's scope resolver never
// yields the read-side "", which storage refuses outright.
func TestMCPWriteProjectIDIsNeverEmpty(t *testing.T) {
	chdirTemp(t)

	if got := testDeps.writeProjectID(); got != "default" {
		t.Fatalf("unconfigured repository resolved to %q, want \"default\"", got)
	}
}
