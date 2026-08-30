// Copyright (c) 2025 by Developer Network.
//
// For more details, see the LICENSE file in the root directory of this
// source code repository or contact Developer Network at info@devnw.com.

package storage

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"testing"
)

// CANARY: REQ=CBIN-305; FEATURE="ExplicitDatabaseOpens"; ASPECT=Storage; STATUS=TESTED; TEST=TestOpenROMissingCreatesNothing,TestOpenRORejectsWrites,TestOpenRWCreatesAndMigrates,TestOrderKeyAllowlist,TestReplaceIndexRollsBack; UPDATED=2026-08-30

// TestOpenROMissingCreatesNothing proves the read-only open never brings a
// database into existence -- not the file, not its parent directory.
func TestOpenROMissingCreatesNothing(t *testing.T) {
	root := t.TempDir()
	dbPath := filepath.Join(root, ".canary", "canary.db")

	db, err := OpenRO(dbPath)
	if err == nil {
		_ = db.Close()
		t.Fatal("OpenRO on a missing database must fail")
	}
	if !errors.Is(err, fs.ErrNotExist) {
		t.Fatalf("want fs.ErrNotExist, got %v", err)
	}
	if _, serr := os.Stat(filepath.Join(root, ".canary")); serr == nil {
		t.Fatal("OpenRO created the .canary directory")
	}
}

// TestOpenRORejectsWrites proves the read-only handle is enforced by SQLite,
// not by caller discipline.
func TestOpenRORejectsWrites(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), ".canary", "canary.db")

	rw, err := OpenRW(dbPath)
	if err != nil {
		t.Fatalf("OpenRW: %v", err)
	}
	if err := rw.UpsertToken(&Token{
		ReqID: "CBIN-001", Feature: "F", Aspect: "API", Status: "STUB",
		FilePath: "a.go", LineNumber: 1, UpdatedAt: "2026-01-01",
		RawToken: "x", IndexedAt: "2026-01-01T00:00:00Z", ProjectID: "default",
	}); err != nil {
		t.Fatalf("seed: %v", err)
	}
	if err := rw.Close(); err != nil {
		t.Fatalf("close: %v", err)
	}

	ro, err := OpenRO(dbPath)
	if err != nil {
		t.Fatalf("OpenRO: %v", err)
	}
	defer func() { _ = ro.Close() }()

	if !ro.ReadOnly() {
		t.Fatal("OpenRO handle does not report itself read-only")
	}
	rows, err := ro.GetTokensByReqID("default", "CBIN-001")
	if err != nil || len(rows) != 1 {
		t.Fatalf("read through OpenRO failed: %v %d", err, len(rows))
	}
	if err := ro.UpsertToken(&Token{
		ReqID: "CBIN-002", Feature: "G", Aspect: "API", Status: "STUB",
		FilePath: "b.go", LineNumber: 1, UpdatedAt: "2026-01-01",
		RawToken: "x", IndexedAt: "2026-01-01T00:00:00Z", ProjectID: "default",
	}); err == nil {
		t.Fatal("a write through OpenRO succeeded")
	}
}

// TestOpenRWCreatesAndMigrates proves the writer creates a database at the
// current schema version, with the tables migrations define.
func TestOpenRWCreatesAndMigrates(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), ".canary", "canary.db")

	db, err := OpenRW(dbPath)
	if err != nil {
		t.Fatalf("OpenRW: %v", err)
	}
	defer func() { _ = db.Close() }()

	var version int
	if err := db.conn.Get(&version, "SELECT COALESCE(MAX(version), 0) FROM schema_migrations WHERE dirty = 0"); err != nil {
		t.Fatalf("read schema version: %v", err)
	}
	if version != LatestVersion {
		t.Fatalf("schema version %d, want %d", version, LatestVersion)
	}

	// index_meta arrives with migration 000007 and starts empty.
	meta, err := db.GetIndexMeta()
	if err != nil {
		t.Fatalf("GetIndexMeta: %v", err)
	}
	if meta != nil {
		t.Fatalf("a fresh database reported index metadata: %+v", meta)
	}
}

// TestOrderKeyAllowlist proves ordering is chosen from a fixed set and that
// anything else is refused before a query is built.
func TestOrderKeyAllowlist(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), ".canary", "canary.db")
	db, err := OpenRW(dbPath)
	if err != nil {
		t.Fatalf("OpenRW: %v", err)
	}
	defer func() { _ = db.Close() }()

	for _, key := range append([]string{""}, OrderKeys()...) {
		if _, err := db.ListTokens("", nil, "", key, 0); err != nil {
			t.Fatalf("order key %q rejected: %v", key, err)
		}
	}

	for _, key := range []string{"updated_at; DROP TABLE tokens", "req_id DESC", "priority ASC, updated_at DESC"} {
		_, err := db.ListTokens("", nil, "", key, 0)
		if !errors.Is(err, ErrInvalidOrderBy) {
			t.Fatalf("order key %q: want ErrInvalidOrderBy, got %v", key, err)
		}
	}

	// The injection attempt must not have reached SQLite.
	var name string
	if err := db.conn.Get(&name, "SELECT name FROM sqlite_master WHERE type='table' AND name='tokens'"); err != nil {
		t.Fatalf("tokens table is gone: %v", err)
	}
}

// TestReplaceIndexRollsBack proves a rebuild that fails partway leaves the
// previous index exactly as it was, rather than emptied or half-written.
func TestReplaceIndexRollsBack(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), ".canary", "canary.db")
	db, err := OpenRW(dbPath)
	if err != nil {
		t.Fatalf("OpenRW: %v", err)
	}
	defer func() { _ = db.Close() }()

	good := func(id string) *Token {
		return &Token{
			ReqID: id, Feature: "F", Aspect: "API", Status: "STUB",
			FilePath: id + ".go", LineNumber: 1, UpdatedAt: "2026-01-01",
			RawToken: "x", IndexedAt: "2026-01-01T00:00:00Z", ProjectID: "default",
		}
	}
	meta := IndexMeta{Root: ".", ProjectID: "default", ParserSchema: 2, ScanDigest: "d", IndexedAt: "2026-01-01T00:00:00Z"}

	if err := db.ReplaceIndex("default", []*Token{good("CBIN-001")}, nil, meta); err != nil {
		t.Fatalf("first rebuild: %v", err)
	}

	// Two tokens differing only in fields outside the uniqueness key collapse
	// into one row, so the post-insert count check must reject the rebuild.
	dup := good("CBIN-002")
	dupAgain := good("CBIN-002")
	dupAgain.Status = "IMPL"
	if err := db.ReplaceIndex("default", []*Token{dup, dupAgain}, nil, meta); err == nil {
		t.Fatal("a rebuild that collapsed two tokens into one row committed")
	}

	rows, err := db.GetTokensByReqID("default", "CBIN-001")
	if err != nil || len(rows) != 1 {
		t.Fatalf("failed rebuild destroyed the previous index: %v %d", err, len(rows))
	}
	if gone, err := db.GetTokensByReqID("default", "CBIN-002"); err != nil || len(gone) != 0 {
		t.Fatalf("failed rebuild left rows behind: %v %+v", err, gone)
	}
}

// TestReplaceIndexIsProjectScoped proves one project's rebuild leaves a
// sibling project's rows in a shared database alone.
func TestReplaceIndexIsProjectScoped(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), ".canary", "canary.db")
	db, err := OpenRW(dbPath)
	if err != nil {
		t.Fatalf("OpenRW: %v", err)
	}
	defer func() { _ = db.Close() }()

	other := &Token{
		ReqID: "CBIN-900", Feature: "Other", Aspect: "API", Status: "STUB",
		FilePath: "o.go", LineNumber: 1, UpdatedAt: "2026-01-01",
		RawToken: "x", IndexedAt: "2026-01-01T00:00:00Z", ProjectID: "sibling",
	}
	if err := db.UpsertToken(other); err != nil {
		t.Fatalf("seed sibling: %v", err)
	}

	meta := IndexMeta{Root: ".", ProjectID: "mine", ParserSchema: 2, ScanDigest: "d", IndexedAt: "2026-01-01T00:00:00Z"}
	mine := &Token{
		ReqID: "CBIN-001", Feature: "F", Aspect: "API", Status: "STUB",
		FilePath: "a.go", LineNumber: 1, UpdatedAt: "2026-01-01",
		RawToken: "x", IndexedAt: "2026-01-01T00:00:00Z", ProjectID: "mine",
	}
	if err := db.ReplaceIndex("mine", []*Token{mine}, nil, meta); err != nil {
		t.Fatalf("rebuild: %v", err)
	}

	kept, err := db.GetTokensByProject("sibling")
	if err != nil || len(kept) != 1 {
		t.Fatalf("sibling project's rows were destroyed: %v %d", err, len(kept))
	}
}

// CANARY: REQ=CBIN-306; FEATURE="IndexMetadata"; ASPECT=Storage; STATUS=TESTED; TEST=TestCANARY_CBIN_306_IndexMetaRoundTrip; UPDATED=2026-08-30
// TestCANARY_CBIN_306_IndexMetaRoundTrip proves metadata survives a write and
// that the single row is replaced rather than duplicated.
func TestCANARY_CBIN_306_IndexMetaRoundTrip(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), ".canary", "canary.db")
	db, err := OpenRW(dbPath)
	if err != nil {
		t.Fatalf("OpenRW: %v", err)
	}
	defer func() { _ = db.Close() }()

	first := IndexMeta{Root: "/a", ProjectID: "p", CommitSHA: "abc", ParserSchema: 2, ScanDigest: "d1", IndexedAt: "2026-01-01T00:00:00Z"}
	if err := db.PutIndexMeta(first); err != nil {
		t.Fatalf("PutIndexMeta: %v", err)
	}
	second := IndexMeta{Root: "/b", ProjectID: "p", CommitSHA: "", ParserSchema: 3, ScanDigest: "d2", IndexedAt: "2026-01-02T00:00:00Z"}
	if err := db.PutIndexMeta(second); err != nil {
		t.Fatalf("PutIndexMeta (replace): %v", err)
	}

	got, err := db.GetIndexMeta()
	if err != nil {
		t.Fatalf("GetIndexMeta: %v", err)
	}
	if got == nil || *got != second {
		t.Fatalf("got %+v, want %+v", got, second)
	}

	var rows int
	if err := db.conn.Get(&rows, "SELECT COUNT(*) FROM index_meta"); err != nil {
		t.Fatalf("count: %v", err)
	}
	if rows != 1 {
		t.Fatalf("index_meta holds %d rows, want exactly 1", rows)
	}
}
