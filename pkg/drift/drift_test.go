// Copyright (c) 2025 by Developer Network.
//
// For more details, see the LICENSE file in the root directory of this
// source code repository or contact Developer Network at info@devnw.com.

package drift

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"devnw.dev/canary/pkg/canaryscan"
	"devnw.dev/canary/pkg/storage"
)

// --- git test helpers -------------------------------------------------

func runGit(t *testing.T, dir string, env []string, args ...string) {
	t.Helper()
	full := append([]string{"-C", dir}, args...)
	cmd := exec.Command("git", full...)
	if env != nil {
		cmd.Env = env
	}
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git %v: %v\n%s", full, err, out)
	}
}

func initRepo(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	runGit(t, dir, nil, "init", "-q")
	return dir
}

// commitFile writes relPath under dir with content, stages it, and commits it.
func commitFile(t *testing.T, dir, relPath, content string) {
	t.Helper()
	writeFile(t, dir, relPath, content)
	runGit(t, dir, nil, "add", relPath)
	runGit(t, dir, nil, "-c", "user.email=t@t", "-c", "user.name=t", "commit", "-q", "-m", "msg")
}

func writeFile(t *testing.T, dir, relPath, content string) {
	t.Helper()
	full := filepath.Join(dir, relPath)
	if err := os.MkdirAll(filepath.Dir(full), 0o750); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(full, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
}

func hashOf(t *testing.T, path string) string {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}

// openIndexed writes a token index at dir/.canary/canary.db carrying the given
// tokens plus an index_meta row, then returns it opened read-only — the exact
// shape Check reads. Callers control each token's ContentHash so a test can
// pin the matching, mismatching, and missing-baseline cases precisely.
func openIndexed(t *testing.T, dir string, tokens []*storage.Token, withMeta bool) *storage.DB {
	t.Helper()
	dbPath := filepath.Join(dir, ".canary", "canary.db")
	if err := os.MkdirAll(filepath.Dir(dbPath), 0o750); err != nil {
		t.Fatal(err)
	}
	if err := storage.MigrateDB(dbPath, "all"); err != nil {
		t.Fatalf("MigrateDB: %v", err)
	}
	rw, err := storage.OpenRW(dbPath)
	if err != nil {
		t.Fatal(err)
	}
	for _, tok := range tokens {
		if err := rw.UpsertToken(tok); err != nil {
			t.Fatal(err)
		}
	}
	if withMeta {
		if err := rw.PutIndexMeta(storage.IndexMeta{
			Root:         dir,
			ProjectID:    "default",
			CommitSHA:    "0000000000000000000000000000000000000000",
			ParserSchema: canaryscan.ParserSchemaVersion,
			ScanDigest:   "digest",
			IndexedAt:    "2026-01-01T00:00:00Z",
		}); err != nil {
			t.Fatal(err)
		}
	}
	if err := rw.Close(); err != nil {
		t.Fatal(err)
	}

	ro, err := storage.OpenRO(dbPath)
	if err != nil {
		t.Fatalf("OpenRO: %v", err)
	}
	t.Cleanup(func() { _ = ro.Close() })
	return ro
}

// fileToken builds a token for rel under dir with its ContentHash set to the
// file's current digest — the baseline `canary index` would have recorded.
func fileToken(t *testing.T, dir, reqID, rel string) *storage.Token {
	t.Helper()
	return &storage.Token{
		ReqID:       reqID,
		Feature:     "F",
		Aspect:      "API",
		Status:      "IMPL",
		FilePath:    rel,
		ContentHash: hashOf(t, filepath.Join(dir, rel)),
		ProjectID:   "default",
		UpdatedAt:   "2026-08-01",
	}
}

func stateOf(states []ReqState, reqID string) (ReqState, bool) {
	for _, s := range states {
		if s.RequirementID == reqID {
			return s, true
		}
	}
	return ReqState{}, false
}

// --- Check: the hash/commit drift verdict --------------------------------

// TestCANARY_CP_278_Check_CleanRepoCurrent: an unchanged, committed, indexed
// file is CURRENT — the "clean repo = no drift" intent, now hash-based.
func TestCANARY_CP_278_Check_CleanRepoCurrent(t *testing.T) {
	dir := initRepo(t)
	commitFile(t, dir, "foo.go", "package foo\n")
	db := openIndexed(t, dir, []*storage.Token{fileToken(t, dir, "CBIN-900", "foo.go")}, true)

	states, err := Check(dir, db, "default")
	if err != nil {
		t.Fatalf("Check: %v", err)
	}
	s, ok := stateOf(states, "CBIN-900")
	if !ok || s.State != StateCurrent {
		t.Fatalf("want CBIN-900 CURRENT, got %+v", states)
	}
}

// TestCANARY_CP_278_Check_ChangedFileDrifted: a file whose bytes changed after
// indexing is DRIFTED — the "detects a changed file" intent — and it fires on
// the hash alone, regardless of dates or same-day commits.
func TestCANARY_CP_278_Check_ChangedFileDrifted(t *testing.T) {
	dir := initRepo(t)
	commitFile(t, dir, "foo.go", "package foo\n")
	db := openIndexed(t, dir, []*storage.Token{fileToken(t, dir, "CBIN-901", "foo.go")}, true)

	// Change the file on disk after the baseline was captured.
	writeFile(t, dir, "foo.go", "package foo\n// changed\n")

	states, err := Check(dir, db, "default")
	if err != nil {
		t.Fatalf("Check: %v", err)
	}
	s, ok := stateOf(states, "CBIN-901")
	if !ok || s.State != StateDrifted {
		t.Fatalf("want CBIN-901 DRIFTED, got %+v", states)
	}
}

// TestCANARY_CP_278_Check_GitFailureUnknown: hashes match but git cannot
// answer (no repository), so the verdict is UNKNOWN, never CURRENT.
func TestCANARY_CP_278_Check_GitFailureUnknown(t *testing.T) {
	dir := t.TempDir() // no `git init`
	writeFile(t, dir, "foo.go", "package foo\n")
	db := openIndexed(t, dir, []*storage.Token{fileToken(t, dir, "CBIN-902", "foo.go")}, true)

	states, err := Check(dir, db, "default")
	if err != nil {
		t.Fatalf("Check: %v", err)
	}
	s, ok := stateOf(states, "CBIN-902")
	if !ok || s.State != StateUnknown {
		t.Fatalf("want CBIN-902 UNKNOWN on git failure, got %+v", states)
	}
}

// TestCANARY_CP_278_Check_MissingBaselineUnknown: a token with no recorded
// content hash cannot be compared, so its requirement is UNKNOWN.
func TestCANARY_CP_278_Check_MissingBaselineUnknown(t *testing.T) {
	dir := initRepo(t)
	commitFile(t, dir, "foo.go", "package foo\n")
	tok := fileToken(t, dir, "CBIN-903", "foo.go")
	tok.ContentHash = "" // baseline never captured
	db := openIndexed(t, dir, []*storage.Token{tok}, true)

	states, err := Check(dir, db, "default")
	if err != nil {
		t.Fatalf("Check: %v", err)
	}
	s, ok := stateOf(states, "CBIN-903")
	if !ok || s.State != StateUnknown {
		t.Fatalf("want CBIN-903 UNKNOWN on missing baseline, got %+v", states)
	}
}

// TestCANARY_CP_278_Check_UnreadableFileUnknown: a baseline exists but the file
// is gone from disk, so it cannot be hashed — UNKNOWN.
func TestCANARY_CP_278_Check_UnreadableFileUnknown(t *testing.T) {
	dir := initRepo(t)
	commitFile(t, dir, "foo.go", "package foo\n")
	tok := fileToken(t, dir, "CBIN-904", "foo.go")
	db := openIndexed(t, dir, []*storage.Token{tok}, true)

	if err := os.Remove(filepath.Join(dir, "foo.go")); err != nil {
		t.Fatal(err)
	}

	states, err := Check(dir, db, "default")
	if err != nil {
		t.Fatalf("Check: %v", err)
	}
	s, ok := stateOf(states, "CBIN-904")
	if !ok || s.State != StateUnknown {
		t.Fatalf("want CBIN-904 UNKNOWN on unreadable file, got %+v", states)
	}
}

// TestCANARY_CP_278_Check_Precedence: a requirement whose files include one
// DRIFTED and one UNKNOWN is DRIFTED (DRIFTED > UNKNOWN > CURRENT).
func TestCANARY_CP_278_Check_Precedence(t *testing.T) {
	dir := initRepo(t)
	commitFile(t, dir, "drifted.go", "package foo\n")
	commitFile(t, dir, "gone.go", "package foo\n")
	toks := []*storage.Token{
		fileToken(t, dir, "CBIN-905", "drifted.go"),
		fileToken(t, dir, "CBIN-905", "gone.go"),
	}
	db := openIndexed(t, dir, toks, true)

	// drifted.go changes on disk; gone.go becomes unreadable.
	writeFile(t, dir, "drifted.go", "package foo\n// changed\n")
	if err := os.Remove(filepath.Join(dir, "gone.go")); err != nil {
		t.Fatal(err)
	}

	states, err := Check(dir, db, "default")
	if err != nil {
		t.Fatalf("Check: %v", err)
	}
	s, ok := stateOf(states, "CBIN-905")
	if !ok || s.State != StateDrifted {
		t.Fatalf("want CBIN-905 DRIFTED (DRIFTED beats UNKNOWN), got %+v", states)
	}
}

// TestCANARY_CP_278_Check_NoIndex: a database that was never indexed (no
// index_meta row) is a clear error, not everything-UNKNOWN.
func TestCANARY_CP_278_Check_NoIndex(t *testing.T) {
	dir := initRepo(t)
	commitFile(t, dir, "foo.go", "package foo\n")
	db := openIndexed(t, dir, []*storage.Token{fileToken(t, dir, "CBIN-906", "foo.go")}, false)

	_, err := Check(dir, db, "default")
	if !errors.Is(err, ErrNoIndex) {
		t.Fatalf("want ErrNoIndex, got %v", err)
	}
}

// --- advisories: stale + doc-drift, kept separate from the State -----------

func repWithFile(reqID, file, updated, status string) canaryscan.Report {
	return canaryscan.Report{
		Requirements: []canaryscan.Requirement{
			{
				ID: reqID,
				Features: []canaryscan.Feature{
					{Feature: "Foo", Aspect: "API", Status: status, Files: []string{file}, Updated: updated},
				},
			},
		},
	}
}

func TestCANARY_CP_278_Advisories_Stale(t *testing.T) {
	dir := t.TempDir()
	refTime := time.Date(2026, 8, 29, 0, 0, 0, 0, time.UTC)

	rep := repWithFile("CBIN-907", "bar.go", "2026-06-01", "TESTED")
	findings, err := Advisories(dir, rep, 30, refTime)
	if err != nil {
		t.Fatalf("Advisories: %v", err)
	}
	var got *Finding
	for i := range findings {
		if findings[i].Kind == KindStale {
			got = &findings[i]
		}
	}
	if got == nil {
		t.Fatalf("expected a stale finding, got %+v", findings)
	}
	if got.ReqID != "CBIN-907" || got.File != "bar.go" {
		t.Errorf("finding = %+v", got)
	}
}

func TestCANARY_CP_278_Advisories_StaleIgnoresNonTestedBenched(t *testing.T) {
	dir := t.TempDir()
	refTime := time.Date(2026, 8, 29, 0, 0, 0, 0, time.UTC)

	rep := repWithFile("CBIN-908", "bar.go", "2020-01-01", "IMPL")
	findings, err := Advisories(dir, rep, 30, refTime)
	if err != nil {
		t.Fatalf("Advisories: %v", err)
	}
	for _, f := range findings {
		if f.Kind == KindStale {
			t.Errorf("unexpected stale finding for IMPL status: %+v", f)
		}
	}
}

func TestCANARY_CP_278_Advisories_DocDrift(t *testing.T) {
	dir := t.TempDir()
	dbDir := filepath.Join(dir, ".canary")
	if err := os.MkdirAll(dbDir, 0o750); err != nil {
		t.Fatal(err)
	}
	dbPath := filepath.Join(dbDir, "canary.db")
	if err := storage.MigrateDB(dbPath, "all"); err != nil {
		t.Fatalf("MigrateDB: %v", err)
	}
	db, err := storage.OpenRW(dbPath)
	if err != nil {
		t.Fatal(err)
	}
	toks := []*storage.Token{
		{ReqID: "CBIN-909", Feature: "A", Aspect: "Docs", Status: "IMPL", FilePath: "a.go", DocStatus: "DOC_STALE", DocPath: "docs/a.md"},
		{ReqID: "CBIN-910", Feature: "B", Aspect: "Docs", Status: "IMPL", FilePath: "b.go", DocStatus: "DOC_MISSING", DocPath: "docs/b.md"},
		{ReqID: "CBIN-911", Feature: "C", Aspect: "Docs", Status: "IMPL", FilePath: "c.go", DocStatus: "DOC_CURRENT", DocPath: "docs/c.md"},
	}
	for _, tok := range toks {
		if err := db.UpsertToken(tok); err != nil {
			t.Fatal(err)
		}
	}
	_ = db.Close()

	findings, err := Advisories(dir, canaryscan.Report{}, 30, time.Time{})
	if err != nil {
		t.Fatalf("Advisories: %v", err)
	}
	got := map[string]bool{}
	for _, f := range findings {
		if f.Kind == KindDocDrift {
			got[f.ReqID] = true
		}
	}
	if !got["CBIN-909"] || !got["CBIN-910"] {
		t.Errorf("expected doc-drift for CBIN-909 and CBIN-910, got %v", got)
	}
	if got["CBIN-911"] {
		t.Errorf("DOC_CURRENT token must not report doc-drift")
	}
}

func TestCANARY_CP_278_Advisories_NoDB(t *testing.T) {
	dir := t.TempDir() // no .canary/canary.db
	findings, err := Advisories(dir, canaryscan.Report{}, 30, time.Time{})
	if err != nil {
		t.Fatalf("Advisories must not error when no DB is present: %v", err)
	}
	for _, f := range findings {
		if f.Kind == KindDocDrift {
			t.Errorf("unexpected doc-drift finding with no DB: %+v", f)
		}
	}
	if _, err := os.Stat(filepath.Join(dir, ".canary", "canary.db")); err == nil {
		t.Error("Advisories must not create a database as a side effect")
	}
}
