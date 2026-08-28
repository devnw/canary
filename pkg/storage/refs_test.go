// Copyright (c) 2025 by Developer Network.
//
// For more details, see the LICENSE file in the root directory of this
// source code repository or contact Developer Network at info@devnw.com.

package storage

import (
	"path/filepath"
	"testing"
)

// CANARY: REQ=CBIN-206; FEATURE="DiagramRefsIndex"; ASPECT=Storage; STATUS=TESTED; TEST=TestCANARY_CBIN_206_RefsRoundTrip; UPDATED=2026-08-28
func TestCANARY_CBIN_206_RefsRoundTrip(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	if err := MigrateDB(dbPath, "all"); err != nil {
		t.Fatalf("Failed to migrate database: %v", err)
	}

	db, err := Open(dbPath)
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer func() { _ = db.Close() }()

	refs := []Ref{
		{ReqID: "CBIN-105", Kind: "diagram", FilePath: "docs/arch.md", LineNumber: 12, Context: "flowchart"},
		{ReqID: "CBIN-105", Kind: "diagram", FilePath: "docs/flow.mmd", LineNumber: 3},
		{ReqID: "PLAT-4521", Kind: "diagram", FilePath: "docs/arch.md", LineNumber: 14},
	}
	if err := db.ReplaceRefs("diagram", refs); err != nil {
		t.Fatalf("ReplaceRefs: %v", err)
	}
	got, err := db.GetRefsByReqID("CBIN-105")
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 {
		t.Fatalf("got %d refs, want 2", len(got))
	}
	if got[0].FilePath != "docs/arch.md" || got[0].LineNumber != 12 {
		t.Errorf("ordering/content wrong: %+v", got[0])
	}

	// Replace is idempotent and clears old rows of the same kind.
	if err := db.ReplaceRefs("diagram", refs[:1]); err != nil {
		t.Fatal(err)
	}
	got, _ = db.GetRefsByReqID("CBIN-105")
	if len(got) != 1 {
		t.Errorf("after replace: got %d refs, want 1", len(got))
	}
	got, _ = db.GetRefsByReqID("PLAT-4521")
	if len(got) != 0 {
		t.Errorf("PLAT refs should be cleared by replace, got %d", len(got))
	}
}
