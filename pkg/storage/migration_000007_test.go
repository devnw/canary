package storage

import (
	"path/filepath"
	"testing"
)

// CANARY: REQ=CBIN-306; FEATURE="IndexMetadata"; ASPECT=Storage; STATUS=TESTED; TEST=TestMigration000007UpDown; UPDATED=2026-08-30
// TestMigration000007UpDown proves migration 7 applies and rolls back
// cleanly, and that rolling back does not take the token rows with it.
func TestMigration000007UpDown(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "m.db")
	if err := MigrateDB(dbPath, "all"); err != nil {
		t.Fatalf("up to latest: %v", err)
	}
	db, err := OpenRW(dbPath)
	if err != nil {
		t.Fatalf("OpenRW: %v", err)
	}
	if err := db.UpsertToken(&Token{
		ReqID: "CBIN-001", Feature: "F", Aspect: "API", Status: "STUB",
		FilePath: "a.go", LineNumber: 1, UpdatedAt: "2026-01-01",
		RawToken: "x", IndexedAt: "2026-01-01T00:00:00Z",
	}); err != nil {
		t.Fatalf("seed: %v", err)
	}
	if err := db.PutIndexMeta(IndexMeta{Root: ".", ProjectID: "default", ParserSchema: 2, ScanDigest: "d", IndexedAt: "t"}); err != nil {
		t.Fatalf("PutIndexMeta: %v", err)
	}
	if err := db.Close(); err != nil {
		t.Fatalf("close: %v", err)
	}

	// Roll back exactly one step: 7 -> 6.
	if err := TeardownDB(dbPath, "1"); err != nil {
		t.Fatalf("down one step: %v", err)
	}

	back, err := OpenRW(dbPath)
	if err != nil {
		t.Fatalf("reopen (re-migrates to latest): %v", err)
	}
	defer func() { _ = back.Close() }()

	rows, err := back.GetTokensByReqID("default", "CBIN-001")
	if err != nil || len(rows) != 1 {
		t.Fatalf("down/up round trip lost the token rows: %v %d", err, len(rows))
	}
}
