// Copyright (c) 2025 by Developer Network.
//
// For more details, see the LICENSE file in the root directory of this
// source code repository or contact Developer Network at info@devnw.com.

package audit

import (
	"os"
	"path/filepath"
	"testing"

	"devnw.dev/canary/pkg/storage"
)

// dbPathIn returns the conventional index location inside root -- the same
// path every command defaults to.
func dbPathIn(root string) string {
	return filepath.Join(root, ".canary", "canary.db")
}

// openDBAt opens (creating and migrating) the index at path. It is the
// writer-side open, so it is the one a fixture uses to seed rows.
func openDBAt(t *testing.T, path string) *storage.DB {
	t.Helper()
	db, err := storage.OpenRW(path)
	if err != nil {
		t.Fatalf("open %s: %v", path, err)
	}
	return db
}

// openTestDB opens a throwaway index in a temp directory, closed for the
// caller when the test ends.
func openTestDB(t *testing.T) *storage.DB {
	t.Helper()
	db := openDBAt(t, filepath.Join(t.TempDir(), ".canary", "canary.db"))
	t.Cleanup(func() { _ = db.Close() })
	return db
}

// seedToken writes one minimal token for reqID under projectID.
func seedToken(t *testing.T, db *storage.DB, projectID, reqID string) {
	t.Helper()
	err := db.UpsertToken(&storage.Token{
		ReqID:      reqID,
		Feature:    "F-" + projectID,
		Aspect:     "API",
		Status:     "STUB",
		FilePath:   projectID + "/a.go",
		LineNumber: 1,
		Priority:   5,
		UpdatedAt:  "2026-01-01",
		RawToken:   "REQ=" + reqID,
		IndexedAt:  "2026-01-01T00:00:00Z",
		ProjectID:  projectID,
	})
	if err != nil {
		t.Fatalf("seed %s/%s: %v", projectID, reqID, err)
	}
}

// initIndexedRepo builds a git repository holding one CANARY token and runs
// `canary index` in it, so the caller starts from a real, committed index
// rather than a hand-built database.
func initIndexedRepo(t *testing.T, bin string) string {
	t.Helper()
	root := t.TempDir()
	initGitRepo(t, root)
	if err := os.WriteFile(filepath.Join(root, "a.go"), []byte(
		"// CANARY: REQ=CBIN-001; FEATURE=\"F\"; ASPECT=API; STATUS=STUB; UPDATED=2026-01-01\n"), 0o600); err != nil {
		t.Fatalf("write a.go: %v", err)
	}
	gitCommitAll(t, root)
	run(t, root, bin, "index", "--root", ".")
	if _, err := os.Stat(dbPathIn(root)); err != nil {
		t.Fatalf("index did not create %s: %v", dbPathIn(root), err)
	}
	return root
}
