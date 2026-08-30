// Copyright (c) 2025 by Developer Network.
//
// For more details, see the LICENSE file in the root directory of this
// source code repository or contact Developer Network at info@devnw.com.

package drift

import (
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

// commitFile writes relPath under dir with content, stages it, and commits
// it with the given committer/author date (RFC3339, e.g. "2026-08-20T12:00:00+00:00").
func commitFile(t *testing.T, dir, relPath, content, date string) {
	t.Helper()
	full := filepath.Join(dir, relPath)
	if err := os.MkdirAll(filepath.Dir(full), 0o750); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(full, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
	runGit(t, dir, nil, "add", relPath)
	env := append(os.Environ(), "GIT_AUTHOR_DATE="+date, "GIT_COMMITTER_DATE="+date)
	runGit(t, dir, env, "-c", "user.email=t@t", "-c", "user.name=t", "commit", "-q", "-m", "msg")
}

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

// --- code-drift ---------------------------------------------------------

func TestCANARY_CBIN_305_Detect_CodeDriftPositive(t *testing.T) {
	dir := initRepo(t)
	commitFile(t, dir, "foo.go", "package foo\n", "2026-08-20T12:00:00+00:00")

	rep := repWithFile("CBIN-900", "foo.go", "2026-08-01", "IMPL")
	findings, err := Detect(dir, rep, 30, time.Time{})
	if err != nil {
		t.Fatalf("Detect: %v", err)
	}

	var got *Finding
	for i := range findings {
		if findings[i].Kind == KindCodeDrift {
			got = &findings[i]
		}
	}
	if got == nil {
		t.Fatalf("expected a code-drift finding, got %+v", findings)
	}
	if got.ReqID != "CBIN-900" || got.File != "foo.go" {
		t.Errorf("finding = %+v", got)
	}
	if got.Detail != "file committed 2026-08-20, token updated 2026-08-01" {
		t.Errorf("Detail = %q", got.Detail)
	}
}

// TestCANARY_CBIN_305_Detect_CodeDriftSameDayNotDrift pins the strict
// time.Time.After semantics used by detectCodeDrift: a commit on the exact
// same day as the token's UPDATED date must NOT be reported as drift.
func TestCANARY_CBIN_305_Detect_CodeDriftSameDayNotDrift(t *testing.T) {
	dir := initRepo(t)
	commitFile(t, dir, "foo.go", "package foo\n", "2026-08-20T12:00:00+00:00")

	rep := repWithFile("CBIN-912", "foo.go", "2026-08-20", "IMPL")
	findings, err := Detect(dir, rep, 30, time.Time{})
	if err != nil {
		t.Fatalf("Detect: %v", err)
	}
	for _, f := range findings {
		if f.Kind == KindCodeDrift {
			t.Errorf("commit date == UPDATED date must not be drift: %+v", f)
		}
	}
}

func TestCANARY_CBIN_305_Detect_CodeDriftNegative(t *testing.T) {
	dir := initRepo(t)
	commitFile(t, dir, "foo.go", "package foo\n", "2026-08-01T12:00:00+00:00")

	rep := repWithFile("CBIN-901", "foo.go", "2026-08-20", "IMPL")
	findings, err := Detect(dir, rep, 30, time.Time{})
	if err != nil {
		t.Fatalf("Detect: %v", err)
	}
	for _, f := range findings {
		if f.Kind == KindCodeDrift {
			t.Errorf("unexpected code-drift finding: %+v", f)
		}
	}
}

func TestCANARY_CBIN_305_Detect_NonGitRootSoftSkip(t *testing.T) {
	dir := t.TempDir() // no `git init`
	if err := os.WriteFile(filepath.Join(dir, "foo.go"), []byte("package foo\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	rep := repWithFile("CBIN-902", "foo.go", "2020-01-01", "IMPL")
	findings, err := Detect(dir, rep, 30, time.Time{})
	if err != nil {
		t.Fatalf("Detect must not error on a non-git root: %v", err)
	}
	for _, f := range findings {
		if f.Kind == KindCodeDrift {
			t.Errorf("unexpected code-drift finding on non-git root: %+v", f)
		}
	}
}

func TestCANARY_CBIN_305_Detect_UntrackedFileSoftSkip(t *testing.T) {
	dir := initRepo(t)
	// File exists on disk but was never committed.
	if err := os.WriteFile(filepath.Join(dir, "untracked.go"), []byte("package foo\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	rep := repWithFile("CBIN-903", "untracked.go", "2020-01-01", "IMPL")
	findings, err := Detect(dir, rep, 30, time.Time{})
	if err != nil {
		t.Fatalf("Detect: %v", err)
	}
	for _, f := range findings {
		if f.Kind == KindCodeDrift {
			t.Errorf("unexpected code-drift finding for untracked file: %+v", f)
		}
	}
}

func TestCANARY_CBIN_305_Detect_CodeDriftDedupesPerFile(t *testing.T) {
	dir := initRepo(t)
	commitFile(t, dir, "foo.go", "package foo\n", "2026-08-20T12:00:00+00:00")

	rep := canaryscan.Report{
		Requirements: []canaryscan.Requirement{
			{
				ID: "CBIN-904",
				Features: []canaryscan.Feature{
					{Feature: "A", Aspect: "API", Status: "IMPL", Files: []string{"foo.go"}, Updated: "2026-08-01"},
					{Feature: "B", Aspect: "CLI", Status: "IMPL", Files: []string{"foo.go"}, Updated: "2026-08-02"},
				},
			},
		},
	}
	findings, err := Detect(dir, rep, 30, time.Time{})
	if err != nil {
		t.Fatalf("Detect: %v", err)
	}
	n := 0
	for _, f := range findings {
		if f.Kind == KindCodeDrift {
			n++
		}
	}
	if n != 1 {
		t.Errorf("expected exactly one deduped code-drift finding for the shared file, got %d: %+v", n, findings)
	}
}

// TestCANARY_CBIN_305_Detect_CachesGitLogPerFile proves lastCommitDate is
// invoked at most once per unique file, not once per (requirement, file)
// pair: two requirements sharing one file plus a second, unique file must
// drive exactly 2 gitLogFn calls even though there are 3 (req, file) pairs.
func TestCANARY_CBIN_305_Detect_CachesGitLogPerFile(t *testing.T) {
	calls := map[string]int{}
	stub := func(root, file string) string {
		calls[file]++
		return "2026-08-20"
	}

	rep := canaryscan.Report{
		Requirements: []canaryscan.Requirement{
			{
				ID: "CBIN-913",
				Features: []canaryscan.Feature{
					{Feature: "A", Aspect: "API", Status: "IMPL", Files: []string{"shared.go"}, Updated: "2026-08-01"},
				},
			},
			{
				ID: "CBIN-914",
				Features: []canaryscan.Feature{
					{Feature: "B", Aspect: "CLI", Status: "IMPL", Files: []string{"shared.go"}, Updated: "2026-08-02"},
				},
			},
			{
				ID: "CBIN-915",
				Features: []canaryscan.Feature{
					{Feature: "C", Aspect: "API", Status: "IMPL", Files: []string{"other.go"}, Updated: "2026-08-03"},
				},
			},
		},
	}

	findings := detectCodeDriftWith("/unused-root", rep, stub)

	pairCount := 0
	for _, f := range findings {
		if f.Kind == KindCodeDrift {
			pairCount++
		}
	}
	if pairCount != 3 {
		t.Fatalf("expected 3 (req, file) code-drift findings, got %d: %+v", pairCount, findings)
	}

	if len(calls) != 2 {
		t.Fatalf("expected gitLogFn called for exactly 2 unique files, got %d: %v", len(calls), calls)
	}
	for file, n := range calls {
		if n != 1 {
			t.Errorf("gitLogFn called %d times for %s, want 1 (cached per unique file, not per (req,file) pair)", n, file)
		}
	}
}

// --- stale ---------------------------------------------------------------

func TestCANARY_CBIN_305_Detect_Stale(t *testing.T) {
	dir := t.TempDir()
	refTime := time.Date(2026, 8, 29, 0, 0, 0, 0, time.UTC)

	rep := repWithFile("CBIN-905", "bar.go", "2026-06-01", "TESTED")
	findings, err := Detect(dir, rep, 30, refTime)
	if err != nil {
		t.Fatalf("Detect: %v", err)
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
	if got.ReqID != "CBIN-905" || got.File != "bar.go" {
		t.Errorf("finding = %+v", got)
	}
}

func TestCANARY_CBIN_305_Detect_StaleIgnoresNonTestedBenched(t *testing.T) {
	dir := t.TempDir()
	refTime := time.Date(2026, 8, 29, 0, 0, 0, 0, time.UTC)

	rep := repWithFile("CBIN-906", "bar.go", "2020-01-01", "IMPL")
	findings, err := Detect(dir, rep, 30, refTime)
	if err != nil {
		t.Fatalf("Detect: %v", err)
	}
	for _, f := range findings {
		if f.Kind == KindStale {
			t.Errorf("unexpected stale finding for IMPL status: %+v", f)
		}
	}
}

// --- doc-drift -------------------------------------------------------------

func TestCANARY_CBIN_305_Detect_DocDrift(t *testing.T) {
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
		{ReqID: "CBIN-907", Feature: "A", Aspect: "Docs", Status: "IMPL", FilePath: "a.go", DocStatus: "DOC_STALE", DocPath: "docs/a.md"},
		{ReqID: "CBIN-908", Feature: "B", Aspect: "Docs", Status: "IMPL", FilePath: "b.go", DocStatus: "DOC_MISSING", DocPath: "docs/b.md"},
		{ReqID: "CBIN-909", Feature: "C", Aspect: "Docs", Status: "IMPL", FilePath: "c.go", DocStatus: "DOC_CURRENT", DocPath: "docs/c.md"},
	}
	for _, tok := range toks {
		if err := db.UpsertToken(tok); err != nil {
			t.Fatal(err)
		}
	}
	_ = db.Close()

	findings, err := Detect(dir, canaryscan.Report{}, 30, time.Time{})
	if err != nil {
		t.Fatalf("Detect: %v", err)
	}
	got := map[string]bool{}
	for _, f := range findings {
		if f.Kind == KindDocDrift {
			got[f.ReqID] = true
		}
	}
	if !got["CBIN-907"] || !got["CBIN-908"] {
		t.Errorf("expected doc-drift for CBIN-907 and CBIN-908, got %v", got)
	}
	if got["CBIN-909"] {
		t.Errorf("DOC_CURRENT token must not report doc-drift")
	}
}

func TestCANARY_CBIN_305_Detect_DocDriftNoDB(t *testing.T) {
	dir := t.TempDir() // no .canary/canary.db
	findings, err := Detect(dir, canaryscan.Report{}, 30, time.Time{})
	if err != nil {
		t.Fatalf("Detect must not error when no DB is present: %v", err)
	}
	for _, f := range findings {
		if f.Kind == KindDocDrift {
			t.Errorf("unexpected doc-drift finding with no DB: %+v", f)
		}
	}
	if _, err := os.Stat(filepath.Join(dir, ".canary", "canary.db")); err == nil {
		t.Error("Detect must not create a database as a side effect")
	}
}
