// Copyright (c) 2025 by Developer Network.
//
// For more details, see the LICENSE file in the root directory of this
// source code repository or contact Developer Network at info@devnw.com.

package audit

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// nextOut is the subset of `canary next --format json` these audits read.
// Its shape is the CLI's contract: which requirement was chosen, and which
// source that answer came from.
type nextOut struct {
	ReqID   string `json:"req_id"`
	Feature string `json:"feature"`
	Status  string `json:"status"`
	Source  string `json:"source"`
	Message string `json:"message"`
}

// runNextJSON runs `canary next` in root capturing stdout ONLY (stderr is
// where notes and diagnostics belong) and decodes it. A stdout that is not
// pure JSON fails the test: the whole point of --format json is that a
// program can read it without filtering.
func runNextJSON(t *testing.T, root, bin string, args ...string) nextOut {
	t.Helper()
	cmd := exec.Command(bin, args...) //nolint:gosec // test-built binary
	cmd.Dir = root
	home := t.TempDir()
	cmd.Env = append(os.Environ(), "HOME="+home, "USERPROFILE="+home)
	var stderr strings.Builder
	cmd.Stderr = &stderr
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("canary %s: %v\nstdout: %s\nstderr: %s", strings.Join(args, " "), err, out, stderr.String())
	}
	var got nextOut
	if jerr := json.Unmarshal(out, &got); jerr != nil {
		t.Fatalf("stdout not pure JSON: %q (stderr: %s)", out, stderr.String())
	}
	if strings.Contains(string(out), "completed") {
		t.Fatalf("next claimed completion: %q", out)
	}
	return got
}

// TestAuditF13 covers F-13: `canary next` with no index must answer from a
// canonical filesystem scan, say so, and never claim the work is finished.
//
// The old fallback shelled out to grep with a hardcoded extension list and
// its own field extraction, so it could disagree with `canary scan` about
// what the tree holds -- and, when its grep found nothing, `next` announced
// "🎉 All requirements completed!" over a repository it had never
// successfully read.
func TestAuditF13(t *testing.T) {
	bin := buildCanary(t)
	root := t.TempDir()
	initGitRepo(t, root)
	if err := os.WriteFile(filepath.Join(root, "a.go"), []byte(
		"// CANARY: REQ=CBIN-001; FEATURE=\"F\"; ASPECT=API; STATUS=STUB; UPDATED=2026-01-01\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	got := runNextJSON(t, root, bin, "next", "--format", "json")
	if got.ReqID != "CBIN-001" {
		t.Errorf("ReqID = %q, want CBIN-001", got.ReqID)
	}
	if got.Source != "filesystem" {
		t.Errorf("Source = %q, want filesystem", got.Source)
	}

	// A read must not have a write side effect: no index is created to
	// answer this question.
	if _, err := os.Stat(filepath.Join(root, ".canary", "canary.db")); err == nil {
		t.Error("next created an index; a read must never write one")
	}

	// --format json is a machine contract and outranks --dry-run, which used
	// to short-circuit ahead of it and print prose to a caller parsing JSON.
	dry := runNextJSON(t, root, bin, "next", "--format=json", "--dry-run")
	if dry.ReqID != "CBIN-001" || dry.Source != "filesystem" {
		t.Errorf("--format=json --dry-run = %+v, want the JSON answer", dry)
	}

	// The old --json spelling still works as an alias.
	alias := runNextJSON(t, root, bin, "next", "--json")
	if alias.ReqID != "CBIN-001" {
		t.Errorf("--json alias = %+v, want CBIN-001", alias)
	}
}

// TestAuditF13_EmptyTreeNeverClaimsCompletion proves the second half of
// F-13: with no index and nothing to do, `next` reports that it found no
// actionable requirement *from the filesystem*, rather than asserting every
// requirement is complete -- a claim it has no index to support.
func TestAuditF13_EmptyTreeNeverClaimsCompletion(t *testing.T) {
	bin := buildCanary(t)
	root := t.TempDir()
	initGitRepo(t, root)

	got := runNextJSON(t, root, bin, "next", "--format", "json")
	if got.ReqID != "" {
		t.Errorf("ReqID = %q, want empty (nothing to select)", got.ReqID)
	}
	if got.Source != "filesystem" {
		t.Errorf("Source = %q, want filesystem", got.Source)
	}

	// Text mode says the same thing in words.
	text, err := runCLI(t, root, bin, "next")
	if err != nil {
		t.Fatalf("canary next: %v\n%s", err, text)
	}
	if !strings.Contains(text, "no actionable requirements found (source=filesystem)") {
		t.Errorf("text output = %q, want the filesystem no-work line", text)
	}
	if strings.Contains(text, "All requirements completed") {
		t.Errorf("next claimed completion without an index: %q", text)
	}
}

// TestAuditF13_StaleIndexFallsBackToFilesystem proves the freshness half of
// the source decision: an index whose recorded commit no longer matches HEAD
// describes a tree that has since moved, so its answer is not usable and the
// filesystem is scanned instead.
func TestAuditF13_StaleIndexFallsBackToFilesystem(t *testing.T) {
	bin := buildCanary(t)
	root := t.TempDir()
	initGitRepo(t, root)
	if err := os.WriteFile(filepath.Join(root, "a.go"), []byte(
		"// CANARY: REQ=CBIN-001; FEATURE=\"F\"; ASPECT=API; STATUS=STUB; UPDATED=2026-01-01\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	gitCommitAll(t, root)
	run(t, root, bin, "index")

	// A fresh index answers from the database.
	fresh := runNextJSON(t, root, bin, "next", "--format", "json")
	if fresh.Source != "database" {
		t.Fatalf("Source = %q, want database for a freshly built index", fresh.Source)
	}

	// Move HEAD: the index now describes a commit that is not this one.
	if err := os.WriteFile(filepath.Join(root, "b.go"), []byte(
		"// CANARY: REQ=CBIN-002; FEATURE=\"G\"; ASPECT=API; STATUS=STUB; UPDATED=2026-01-01\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	gitCommitAll(t, root)

	stale := runNextJSON(t, root, bin, "next", "--format", "json")
	if stale.Source != "filesystem" {
		t.Errorf("Source = %q, want filesystem for an index built at another commit", stale.Source)
	}
}
