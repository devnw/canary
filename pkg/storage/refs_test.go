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

// CANARY: REQ=CBIN-301; FEATURE="MigrateRefsIndex"; ASPECT=Storage; STATUS=TESTED; TEST=TestCANARY_CBIN_301_MigrateRefsRoundTrip; UPDATED=2026-08-29
func TestCANARY_CBIN_301_MigrateRefsRoundTrip(t *testing.T) {
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
		{ReqID: "CBIN-105", Kind: "migrate", FilePath: "old/legacy.go", LineNumber: 20, Context: "move to new client"},
		{ReqID: "", Kind: "migrate", FilePath: "old/orphan.go", LineNumber: 4, Context: "unowned note, no ReqID matched"},
	}
	if err := db.ReplaceRefs("migrate", refs); err != nil {
		t.Fatalf("ReplaceRefs: %v", err)
	}

	got, err := db.GetRefsByReqID("CBIN-105")
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || got[0].Kind != "migrate" || got[0].Context != "move to new client" {
		t.Errorf("GetRefsByReqID(CBIN-105) = %+v", got)
	}

	all, err := db.GetRefsByKind("migrate", 0)
	if err != nil {
		t.Fatal(err)
	}
	if len(all) != 2 {
		t.Fatalf("GetRefsByKind(migrate) len = %d, want 2", len(all))
	}
	if all[0].FilePath != "old/legacy.go" || all[1].FilePath != "old/orphan.go" {
		t.Errorf("GetRefsByKind ordering wrong: %+v", all)
	}
	// The empty req_id row round-trips.
	foundEmpty := false
	for _, r := range all {
		if r.ReqID == "" && r.FilePath == "old/orphan.go" {
			foundEmpty = true
		}
	}
	if !foundEmpty {
		t.Errorf("expected empty-req_id row to round-trip, got %+v", all)
	}

	// diagram kind is untouched by a migrate ReplaceRefs call.
	if err := db.ReplaceRefs("diagram", []Ref{
		{ReqID: "CBIN-105", Kind: "diagram", FilePath: "docs/arch.md", LineNumber: 1},
	}); err != nil {
		t.Fatal(err)
	}
	byReq, err := db.GetRefsByReqID("CBIN-105")
	if err != nil {
		t.Fatal(err)
	}
	if len(byReq) != 2 {
		t.Fatalf("expected diagram+migrate refs for CBIN-105, got %d: %+v", len(byReq), byReq)
	}
}

// CANARY: REQ=CBIN-301; FEATURE="MigrateRefsIndex"; ASPECT=Storage; STATUS=TESTED; TEST=TestCANARY_CBIN_301_GetRefsByKindLimit; UPDATED=2026-08-29
func TestCANARY_CBIN_301_GetRefsByKindLimit(t *testing.T) {
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

	var refs []Ref
	for i := 0; i < 150; i++ {
		refs = append(refs, Ref{ReqID: "CBIN-105", Kind: "migrate", FilePath: "f.go", LineNumber: i + 1})
	}
	if err := db.ReplaceRefs("migrate", refs); err != nil {
		t.Fatal(err)
	}

	got, err := db.GetRefsByKind("migrate", 0)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 100 {
		t.Errorf("GetRefsByKind default cap = %d, want 100", len(got))
	}

	got, err = db.GetRefsByKind("migrate", 5)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 5 {
		t.Errorf("GetRefsByKind(limit=5) = %d, want 5", len(got))
	}
}
