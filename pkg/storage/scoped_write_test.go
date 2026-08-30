// Copyright (c) 2025 by Developer Network.
//
// For more details, see the LICENSE file in the root directory of this
// source code repository or contact Developer Network at info@devnw.com.

package storage

import (
	"errors"
	"path/filepath"
	"testing"
)

// CANARY: REQ=CBIN-308; FEATURE="ScopedWrites"; ASPECT=Storage; STATUS=TESTED; TEST=TestUpdatePriorityRejectsUnscoped,TestUpdateSpecStatusRejectsUnscoped,TestUpdatePriorityScopedStillWrites; UPDATED=2026-08-30
// CANARY: REQ=CBIN-309; FEATURE="StaleSchemaGuard"; ASPECT=Storage; STATUS=TESTED; TEST=TestOpenRORefusesStaleSchema,TestStaleSchemaRefusesRead; UPDATED=2026-08-30

// seedScoped writes one token for reqID under projectID with the given
// priority, so a test can prove a write reached exactly one project's rows.
func seedScoped(t *testing.T, db *DB, projectID, reqID string, priority int) {
	t.Helper()
	err := db.UpsertToken(&Token{
		ReqID: reqID, Feature: "F", Aspect: "API", Status: "STUB",
		FilePath: projectID + "/a.go", LineNumber: 1, Priority: priority,
		UpdatedAt: "2026-01-01", RawToken: "x",
		IndexedAt: "2026-01-01T00:00:00Z", ProjectID: projectID,
	})
	if err != nil {
		t.Fatalf("seed %s/%s: %v", projectID, reqID, err)
	}
}

// TestUpdatePriorityRejectsUnscoped proves an unscoped priority write is
// refused outright rather than quietly rewriting every project's rows.
func TestUpdatePriorityRejectsUnscoped(t *testing.T) {
	db, err := OpenRW(filepath.Join(t.TempDir(), ".canary", "canary.db"))
	if err != nil {
		t.Fatalf("OpenRW: %v", err)
	}
	defer func() { _ = db.Close() }()

	seedScoped(t, db, "projA", "CBIN-001", 5)
	seedScoped(t, db, "projB", "CBIN-001", 5)

	if err := db.UpdatePriority("", "CBIN-001", "F", 1); err == nil {
		t.Fatal("an unscoped UpdatePriority succeeded")
	}

	for _, project := range []string{"projA", "projB"} {
		rows, err := db.GetTokensByReqID(project, "CBIN-001")
		if err != nil || len(rows) != 1 {
			t.Fatalf("lookup %s: %v %d", project, err, len(rows))
		}
		if rows[0].Priority != 5 {
			t.Fatalf("%s priority changed to %d", project, rows[0].Priority)
		}
	}
}

// TestUpdateSpecStatusRejectsUnscoped is the same proof for the spec-status
// writer, which had the identical unscoped fallthrough.
func TestUpdateSpecStatusRejectsUnscoped(t *testing.T) {
	db, err := OpenRW(filepath.Join(t.TempDir(), ".canary", "canary.db"))
	if err != nil {
		t.Fatalf("OpenRW: %v", err)
	}
	defer func() { _ = db.Close() }()

	seedScoped(t, db, "projA", "CBIN-001", 5)
	seedScoped(t, db, "projB", "CBIN-001", 5)

	if err := db.UpdateSpecStatus("", "CBIN-001", "APPROVED"); err == nil {
		t.Fatal("an unscoped UpdateSpecStatus succeeded")
	}

	for _, project := range []string{"projA", "projB"} {
		rows, err := db.GetTokensByReqID(project, "CBIN-001")
		if err != nil || len(rows) != 1 {
			t.Fatalf("lookup %s: %v %d", project, err, len(rows))
		}
		if rows[0].SpecStatus != "" {
			t.Fatalf("%s spec status changed to %q", project, rows[0].SpecStatus)
		}
	}
}

// TestUpdatePriorityScopedStillWrites proves the refusal did not cost the
// legitimate scoped write, and that it touches only its own project.
func TestUpdatePriorityScopedStillWrites(t *testing.T) {
	db, err := OpenRW(filepath.Join(t.TempDir(), ".canary", "canary.db"))
	if err != nil {
		t.Fatalf("OpenRW: %v", err)
	}
	defer func() { _ = db.Close() }()

	seedScoped(t, db, "projA", "CBIN-001", 5)
	seedScoped(t, db, "projB", "CBIN-001", 5)

	if err := db.UpdatePriority("projA", "CBIN-001", "F", 1); err != nil {
		t.Fatalf("scoped UpdatePriority: %v", err)
	}

	mine, err := db.GetTokensByReqID("projA", "CBIN-001")
	if err != nil || len(mine) != 1 || mine[0].Priority != 1 {
		t.Fatalf("scoped write did not land: %v %+v", err, mine)
	}
	theirs, err := db.GetTokensByReqID("projB", "CBIN-001")
	if err != nil || len(theirs) != 1 || theirs[0].Priority != 5 {
		t.Fatalf("scoped write reached a sibling project: %v %+v", err, theirs)
	}
}

// TestOpenRORefusesStaleSchema proves a read against a pre-v7 database is
// refused with a typed error naming the fix, rather than handing the caller a
// raw "no such column" from the first query. A read path must never migrate.
func TestOpenRORefusesStaleSchema(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), ".canary", "canary.db")

	// Migrate to exactly version 6: the last schema before content_hash, the
	// column `canary list` selects.
	if err := MigrateDB(dbPath, "6"); err != nil {
		t.Fatalf("migrate to version 6: %v", err)
	}

	db, err := OpenRO(dbPath)
	if err == nil {
		_ = db.Close()
		t.Fatal("OpenRO accepted a stale-schema database")
	}
	if !errors.Is(err, ErrSchemaOutOfDate) {
		t.Fatalf("want ErrSchemaOutOfDate, got %v", err)
	}

	// The refusal must not have migrated the database behind the caller's
	// back -- a read command that silently rewrote the schema would be the
	// very side effect OpenRO exists to prevent.
	stale, version, nerr := NeedsMigration(dbPath)
	if nerr != nil {
		t.Fatalf("NeedsMigration: %v", nerr)
	}
	if !stale || version != 6 {
		t.Fatalf("OpenRO migrated the database: stale=%v version=%d", stale, version)
	}
}
