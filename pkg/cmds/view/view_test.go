// Copyright (c) 2025 by Developer Network.
//
// For more details, see the LICENSE file in the root directory of this
// source code repository or contact Developer Network at info@devnw.com.

package view

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"devnw.dev/canary/pkg/storage"
)

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

// seedGitDB is like seedDB but roots the project in a real git repo with one
// committed file, so BuildView's drift check has something to compare
// against. tokenUpdated is the UPDATED value stamped on the token that
// references committedFile.
func seedGitDB(t *testing.T, committedFile, commitDate, tokenUpdated string) (dbPath, root string) {
	t.Helper()
	root = t.TempDir()
	runGit(t, root, nil, "init", "-q")

	full := filepath.Join(root, committedFile)
	if err := os.WriteFile(full, []byte("package foo\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	runGit(t, root, nil, "add", committedFile)
	env := append(os.Environ(), "GIT_AUTHOR_DATE="+commitDate, "GIT_COMMITTER_DATE="+commitDate)
	runGit(t, root, env, "-c", "user.email=t@t", "-c", "user.name=t", "commit", "-q", "-m", "msg")

	if err := os.MkdirAll(filepath.Join(root, ".canary"), 0o750); err != nil {
		t.Fatal(err)
	}
	dbPath = filepath.Join(root, ".canary", "canary.db")
	if err := storage.MigrateDB(dbPath, "all"); err != nil {
		t.Fatalf("MigrateDB: %v", err)
	}

	db, err := storage.OpenRW(dbPath)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = db.Close() }()

	if err := db.UpsertToken(&storage.Token{
		ReqID: "CBIN-950", Feature: "Foo", Aspect: "API", Status: "IMPL",
		FilePath: committedFile, LineNumber: 1, UpdatedAt: tokenUpdated,
	}); err != nil {
		t.Fatal(err)
	}
	return dbPath, root
}

func TestCANARY_CBIN_305_BuildView_Drifted(t *testing.T) {
	dbPath, root := seedGitDB(t, "foo.go", "2026-08-20T12:00:00+00:00", "2026-08-01")
	v, err := BuildView(dbPath, root, "", "CBIN-950", 10)
	if err != nil {
		t.Fatalf("BuildView: %v", err)
	}
	if !v.Drifted {
		t.Fatal("expected Drifted = true")
	}
	if v.DriftReason == "" {
		t.Error("expected a non-empty DriftReason")
	}
}

func TestCANARY_CBIN_305_BuildView_NotDrifted(t *testing.T) {
	dbPath, root := seedGitDB(t, "foo.go", "2026-08-01T12:00:00+00:00", "2026-08-20")
	v, err := BuildView(dbPath, root, "", "CBIN-950", 10)
	if err != nil {
		t.Fatalf("BuildView: %v", err)
	}
	if v.Drifted {
		t.Errorf("expected Drifted = false, got reason %q", v.DriftReason)
	}
	if v.DriftReason != "" {
		t.Errorf("DriftReason = %q, want empty", v.DriftReason)
	}
}

func TestCANARY_CBIN_305_BuildView_NonGitRootSoftSkip(t *testing.T) {
	dbPath, root := seedDB(t) // no git init
	v, err := BuildView(dbPath, root, "", "CBIN-105", 10)
	if err != nil {
		t.Fatalf("BuildView: %v", err)
	}
	if v.Drifted {
		t.Errorf("expected Drifted = false on a non-git root, got reason %q", v.DriftReason)
	}
}

func seedDB(t *testing.T) (dbPath, root string) {
	t.Helper()
	root = t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, ".canary"), 0o750); err != nil {
		t.Fatal(err)
	}
	dbPath = filepath.Join(root, ".canary", "canary.db")

	// storage.Open does not run migrations itself (see storage.go:82
	// "Migrations are handled automatically by the CLI's PersistentPreRunE");
	// tests must migrate explicitly, mirroring the convention used in
	// internal/storage/refs_test.go and cmd/canary/main.go's PersistentPreRunE.
	if err := storage.MigrateDB(dbPath, "all"); err != nil {
		t.Fatalf("MigrateDB: %v", err)
	}

	db, err := storage.OpenRW(dbPath)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = db.Close() }()

	toks := []*storage.Token{
		{ReqID: "CBIN-105", Feature: "Scanner", Aspect: "Engine", Status: "TESTED",
			FilePath: "scan.go", LineNumber: 10, Test: "TestScan", DependsOn: "CBIN-101"},
		{ReqID: "CBIN-105", Feature: "ScannerCLI", Aspect: "CLI", Status: "IMPL",
			FilePath: "cli.go", LineNumber: 5, RelatedTo: "PLAT-4521"},
	}
	for _, tok := range toks {
		if err := db.UpsertToken(tok); err != nil {
			t.Fatal(err)
		}
	}
	if err := db.ReplaceRefs("diagram", []storage.Ref{
		{ReqID: "CBIN-105", Kind: "diagram", FilePath: "docs/arch.md", LineNumber: 7},
	}); err != nil {
		t.Fatal(err)
	}
	return dbPath, root
}

func TestCANARY_CBIN_204_BuildView(t *testing.T) {
	dbPath, root := seedDB(t)
	v, err := BuildView(dbPath, root, "", "CBIN-105", 10)
	if err != nil {
		t.Fatalf("BuildView: %v", err)
	}
	if v.ReqID != "CBIN-105" {
		t.Errorf("ReqID = %q", v.ReqID)
	}
	if v.Statuses["TESTED"] != 1 || v.Statuses["IMPL"] != 1 {
		t.Errorf("Statuses = %v", v.Statuses)
	}
	if v.Completion != 50 {
		t.Errorf("Completion = %d, want 50", v.Completion)
	}
	if len(v.Files) != 2 || v.FilesTotal != 2 {
		t.Errorf("Files = %v (total %d)", v.Files, v.FilesTotal)
	}
	if len(v.Tests) != 1 || v.Tests[0] != "TestScan" {
		t.Errorf("Tests = %v", v.Tests)
	}
	if len(v.DependsOn) != 1 || v.DependsOn[0] != "CBIN-101" {
		t.Errorf("DependsOn = %v", v.DependsOn)
	}
	if len(v.RelatedTo) != 1 || v.RelatedTo[0] != "PLAT-4521" {
		t.Errorf("RelatedTo = %v", v.RelatedTo)
	}
	if len(v.Diagrams) != 1 || v.Diagrams[0] != "docs/arch.md:7" {
		t.Errorf("Diagrams = %v", v.Diagrams)
	}
}

func TestCANARY_CBIN_204_BuildView_FileCap(t *testing.T) {
	dbPath, root := seedDB(t)
	v, err := BuildView(dbPath, root, "", "CBIN-105", 1)
	if err != nil {
		t.Fatal(err)
	}
	if len(v.Files) != 1 || v.FilesTotal != 2 {
		t.Errorf("cap not applied: files=%v total=%d", v.Files, v.FilesTotal)
	}
}

func TestCANARY_CBIN_204_BuildView_DiagramCap(t *testing.T) {
	dbPath, root := seedDB(t)

	// Add extra diagram refs to test the cap
	db, err := storage.OpenRW(dbPath)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = db.Close() }()

	if err := db.ReplaceRefs("diagram", []storage.Ref{
		{ReqID: "CBIN-105", Kind: "diagram", FilePath: "docs/arch.md", LineNumber: 7},
		{ReqID: "CBIN-105", Kind: "diagram", FilePath: "docs/design.md", LineNumber: 12},
		{ReqID: "CBIN-105", Kind: "diagram", FilePath: "docs/flow.md", LineNumber: 5},
	}); err != nil {
		t.Fatal(err)
	}

	v, err := BuildView(dbPath, root, "", "CBIN-105", 1)
	if err != nil {
		t.Fatal(err)
	}
	if len(v.Diagrams) != 1 || v.DiagramsTotal != 3 {
		t.Errorf("cap not applied: diagrams=%v total=%d", v.Diagrams, v.DiagramsTotal)
	}
}

func TestCANARY_CBIN_204_BuildView_NotFound(t *testing.T) {
	dbPath, root := seedDB(t)
	if _, err := BuildView(dbPath, root, "", "CBIN-999", 10); err == nil {
		t.Error("unknown requirement must return an error")
	}
}

// CANARY: REQ=CBIN-301; FEATURE="MigrateNotesView"; ASPECT=CLI; STATUS=TESTED; TEST=TestCANARY_CBIN_301_BuildView_MigrateNotes,TestCANARY_CBIN_301_BuildView_MigrateNotesCap; UPDATED=2026-08-29
func TestCANARY_CBIN_301_BuildView_MigrateNotes(t *testing.T) {
	dbPath, root := seedDB(t)

	db, err := storage.OpenRW(dbPath)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = db.Close() }()

	if err := db.ReplaceRefs("migrate", []storage.Ref{
		{ReqID: "CBIN-105", Kind: "migrate", FilePath: "old/legacy.go", LineNumber: 20, Context: "move to new client"},
	}); err != nil {
		t.Fatal(err)
	}

	v, err := BuildView(dbPath, root, "", "CBIN-105", 10)
	if err != nil {
		t.Fatalf("BuildView: %v", err)
	}

	// Regression: diagram and migrate refs must not cross-contaminate.
	if len(v.Diagrams) != 1 || v.Diagrams[0] != "docs/arch.md:7" {
		t.Errorf("Diagrams = %v, want only the diagram ref", v.Diagrams)
	}
	if v.DiagramsTotal != 1 {
		t.Errorf("DiagramsTotal = %d, want 1", v.DiagramsTotal)
	}
	if len(v.MigrateNotes) != 1 || v.MigrateNotes[0] != "old/legacy.go:20: move to new client" {
		t.Errorf("MigrateNotes = %v", v.MigrateNotes)
	}
	if v.MigrateNotesTotal != 1 {
		t.Errorf("MigrateNotesTotal = %d, want 1", v.MigrateNotesTotal)
	}
}

func TestCANARY_CBIN_301_BuildView_MigrateNotesCap(t *testing.T) {
	dbPath, root := seedDB(t)

	db, err := storage.OpenRW(dbPath)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = db.Close() }()

	if err := db.ReplaceRefs("migrate", []storage.Ref{
		{ReqID: "CBIN-105", Kind: "migrate", FilePath: "a.go", LineNumber: 1, Context: "note a"},
		{ReqID: "CBIN-105", Kind: "migrate", FilePath: "b.go", LineNumber: 2, Context: "note b"},
		{ReqID: "CBIN-105", Kind: "migrate", FilePath: "c.go", LineNumber: 3, Context: "note c"},
	}); err != nil {
		t.Fatal(err)
	}

	v, err := BuildView(dbPath, root, "", "CBIN-105", 1)
	if err != nil {
		t.Fatal(err)
	}
	if len(v.MigrateNotes) != 1 || v.MigrateNotesTotal != 3 {
		t.Errorf("cap not applied: notes=%v total=%d", v.MigrateNotes, v.MigrateNotesTotal)
	}
}
