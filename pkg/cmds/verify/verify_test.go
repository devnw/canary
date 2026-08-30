// Copyright (c) 2025 by Developer Network.
//
// For more details, see the LICENSE file in the root directory of this
// source code repository or contact Developer Network at info@devnw.com.

package verify

import (
	"bytes"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"devnw.dev/canary/pkg/evidence"
)

const fixtureDigest = "sha256:" + "abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789"

// gitRepo turns dir into a git repository with one commit and returns its
// HEAD SHA -- the commit evidence records must bind to.
func gitRepo(t *testing.T, dir string) string {
	t.Helper()
	run := func(args ...string) string {
		full := append([]string{
			"-C", dir,
			"-c", "user.email=test@example.com",
			"-c", "user.name=Test",
			"-c", "commit.gpgsign=false",
			"-c", "init.defaultBranch=main",
		}, args...)
		out, err := exec.Command("git", full...).CombinedOutput()
		if err != nil {
			t.Fatalf("git %s: %v\n%s", strings.Join(args, " "), err, out)
		}
		return strings.TrimSpace(string(out))
	}
	run("init", "-q")
	run("commit", "--allow-empty", "-q", "-m", "init")
	return run("rev-parse", "HEAD")
}

// fixture builds a project root holding one token and one claim, and returns
// the root and its HEAD commit.
func fixture(t *testing.T, token, claims string) (root, commit string) {
	t.Helper()
	root = t.TempDir()
	commit = gitRepo(t, root)
	if err := os.WriteFile(filepath.Join(root, "a.go"), []byte(token), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "GAP_ANALYSIS.md"), []byte(claims), 0o644); err != nil {
		t.Fatal(err)
	}
	return root, commit
}

// writeStore writes an evidence store under root.
func writeStore(t *testing.T, root string, recs []evidence.Record) string {
	t.Helper()
	path := filepath.Join(root, ".canary", "evidence.json")
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	data, err := json.Marshal(evidence.File{SchemaVersion: 1, Records: recs})
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, data, 0o644); err != nil {
		t.Fatal(err)
	}
	return path
}

// record builds a PASS record for project "default" at commit.
func record(req, feature, aspect, commit string) evidence.Record {
	return evidence.Record{
		ProjectID: "default", RequirementID: req, Feature: feature, Aspect: aspect,
		TestID: "TestX", Command: "go test ./...", Result: "PASS", CommitSHA: commit,
		ObservedAt: "2026-08-30T00:00:00Z", Runner: "local", ArtifactDigest: fixtureDigest,
	}
}

// runVerify runs one verification and returns stdout, stderr and exit code.
func runVerify(opts Options) (string, string, int) {
	var stdout, stderr bytes.Buffer
	code := Run(opts, &stdout, &stderr)
	return stdout.String(), stderr.String(), code
}

func TestVerifyRun_EvidenceMissing(t *testing.T) {
	root, _ := fixture(t,
		"// CANARY: REQ=CBIN-001; FEATURE=\"F\"; ASPECT=API; STATUS=TESTED; TEST=TestF; UPDATED=2026-01-01\n",
		"✅ CBIN-001 - F\n")

	stdout, stderr, code := runVerify(Options{Root: root, ClaimsPath: filepath.Join(root, "GAP_ANALYSIS.md")})
	if code != 1 {
		t.Fatalf("exit=%d, want 1 (stderr %q)", code, stderr)
	}
	want := `{"ok":false,"state":"UNVERIFIED","code":"EVIDENCE_MISSING","message":"no passing evidence at current commit"}` + "\n"
	if stdout != want {
		t.Fatalf("stdout=%q want %q", stdout, want)
	}
	if !strings.Contains(stderr, "UNVERIFIED CBIN-001 F/API reason=no_evidence") {
		t.Errorf("stderr must name the missing feature: %q", stderr)
	}
}

func TestVerifyRun_Verified(t *testing.T) {
	root, commit := fixture(t,
		"// CANARY: REQ=CBIN-001; FEATURE=\"F\"; ASPECT=API; STATUS=TESTED; TEST=TestF; UPDATED=2026-01-01\n",
		"✅ CBIN-001 - F\n")
	store := writeStore(t, root, []evidence.Record{record("CBIN-001", "F", "API", commit)})

	stdout, stderr, code := runVerify(Options{
		Root:         root,
		ClaimsPath:   filepath.Join(root, "GAP_ANALYSIS.md"),
		EvidencePath: store,
	})
	if code != 0 {
		t.Fatalf("exit=%d, want 0 (stderr %q)", code, stderr)
	}
	want := `{"ok":true,"state":"VERIFIED","code":"OK","message":"all claims verified"}` + "\n"
	if stdout != want {
		t.Fatalf("stdout=%q want %q", stdout, want)
	}
	if stderr != "" {
		t.Errorf("a clean verification must be silent on stderr: %q", stderr)
	}
}

// TestVerifyRun_WrongCommitIsNotVerified proves evidence recorded at another
// commit does not verify the claim at this one -- the whole point of binding
// evidence to a commit.
func TestVerifyRun_WrongCommitIsNotVerified(t *testing.T) {
	root, _ := fixture(t,
		"// CANARY: REQ=CBIN-001; FEATURE=\"F\"; ASPECT=API; STATUS=TESTED; TEST=TestF; UPDATED=2026-01-01\n",
		"✅ CBIN-001 - F\n")
	other := strings.Repeat("ab", 20)
	store := writeStore(t, root, []evidence.Record{record("CBIN-001", "F", "API", other)})

	_, stderr, code := runVerify(Options{
		Root:         root,
		ClaimsPath:   filepath.Join(root, "GAP_ANALYSIS.md"),
		EvidencePath: store,
	})
	if code != 1 {
		t.Fatalf("exit=%d, want 1", code)
	}
	if !strings.Contains(stderr, "reason=wrong_commit") {
		t.Errorf("stderr must explain the commit mismatch: %q", stderr)
	}
}

func TestVerifyRun_TextFormat(t *testing.T) {
	root, commit := fixture(t,
		"// CANARY: REQ=CBIN-001; FEATURE=\"F\"; ASPECT=API; STATUS=TESTED; TEST=TestF; UPDATED=2026-01-01\n"+
			"// CANARY: REQ=CBIN-002; FEATURE=\"G\"; ASPECT=API; STATUS=TESTED; TEST=TestG; UPDATED=2026-01-01\n",
		"✅ CBIN-001 - F\n✅ CBIN-002 - G\n")
	store := writeStore(t, root, []evidence.Record{record("CBIN-001", "F", "API", commit)})
	claims := filepath.Join(root, "GAP_ANALYSIS.md")

	stdout, _, code := runVerify(Options{Root: root, ClaimsPath: claims, EvidencePath: store, Format: FormatText})
	if code != 1 || stdout != "UNVERIFIED: 1 missing\n" {
		t.Fatalf("exit=%d stdout=%q", code, stdout)
	}

	store = writeStore(t, root, []evidence.Record{
		record("CBIN-001", "F", "API", commit),
		record("CBIN-002", "G", "API", commit),
	})
	stdout, _, code = runVerify(Options{Root: root, ClaimsPath: claims, EvidencePath: store, Format: FormatText})
	if code != 0 || stdout != "VERIFIED\n" {
		t.Fatalf("exit=%d stdout=%q", code, stdout)
	}
}

// TestVerifyRun_ScanIssueIsIncomplete proves a tree that could not be fully
// scanned yields UNKNOWN/SCAN_INCOMPLETE, never a verified verdict.
func TestVerifyRun_ScanIssueIsIncomplete(t *testing.T) {
	root, commit := fixture(t,
		"// CANARY: REQ=CBIN-001; FEATURE=\"F\"; ASPECT=API; STATUS=TESTED; TEST=TestF; UPDATED=2026-01-01\n",
		"✅ CBIN-001 - F\n")
	store := writeStore(t, root, []evidence.Record{record("CBIN-001", "F", "API", commit)})
	// A binary file is an unreadable-for-tokens file: a scan issue.
	if err := os.WriteFile(filepath.Join(root, "blob.dat"), append([]byte("CANARY: x"), 0x00, 0x01), 0o644); err != nil {
		t.Fatal(err)
	}

	stdout, stderr, code := runVerify(Options{
		Root:         root,
		ClaimsPath:   filepath.Join(root, "GAP_ANALYSIS.md"),
		EvidencePath: store,
	})
	if code != 1 {
		t.Fatalf("exit=%d, want 1", code)
	}
	want := `{"ok":false,"state":"UNKNOWN","code":"SCAN_INCOMPLETE","message":"scan incomplete"}` + "\n"
	if stdout != want {
		t.Fatalf("stdout=%q want %q", stdout, want)
	}
	if !strings.Contains(stderr, "CANARY_SCAN_ISSUE") {
		t.Errorf("stderr must list the issue: %q", stderr)
	}
}

// TestVerifyRun_UndeclaredClaim proves a claim for a requirement that no
// token declares fails, rather than passing vacuously because there is
// nothing to check.
func TestVerifyRun_UndeclaredClaim(t *testing.T) {
	root, _ := fixture(t,
		"// CANARY: REQ=CBIN-001; FEATURE=\"F\"; ASPECT=API; STATUS=TESTED; TEST=TestF; UPDATED=2026-01-01\n",
		"✅ CBIN-404 - Ghost\n")

	stdout, stderr, code := runVerify(Options{Root: root, ClaimsPath: filepath.Join(root, "GAP_ANALYSIS.md")})
	if code != 1 {
		t.Fatalf("exit=%d, want 1", code)
	}
	if !strings.Contains(stdout, "EVIDENCE_MISSING") {
		t.Fatalf("stdout=%q", stdout)
	}
	if !strings.Contains(stderr, "UNVERIFIED CBIN-404 */* reason=no_evidence") {
		t.Errorf("stderr must report the undeclared claim: %q", stderr)
	}
}

// TestVerifyRun_MissingClaimsFileIsUnknown proves an unreadable claims file
// is UNKNOWN, not "no claims".
func TestVerifyRun_MissingClaimsFileIsUnknown(t *testing.T) {
	root, _ := fixture(t,
		"// CANARY: REQ=CBIN-001; FEATURE=\"F\"; ASPECT=API; STATUS=TESTED; TEST=TestF; UPDATED=2026-01-01\n",
		"✅ CBIN-001 - F\n")

	stdout, _, code := runVerify(Options{Root: root, ClaimsPath: filepath.Join(root, "NOPE.md")})
	if code != 1 {
		t.Fatalf("exit=%d, want 1", code)
	}
	if !strings.Contains(stdout, "SCAN_INCOMPLETE") {
		t.Fatalf("stdout=%q, want SCAN_INCOMPLETE", stdout)
	}
}

// TestVerifyRun_ExternalUnknownBlocks proves an unresolvable external
// dependency downgrades an otherwise-verified verdict to UNKNOWN, and that
// --allow-unknown-external opts out of that gate.
func TestVerifyRun_ExternalUnknownBlocks(t *testing.T) {
	root := t.TempDir()
	commit := gitRepo(t, root)
	if err := os.MkdirAll(filepath.Join(root, ".canary"), 0o755); err != nil {
		t.Fatal(err)
	}
	cfg := "project:\n  name: t\n  key: CBIN\n" +
		"sources:\n" +
		"  - name: core\n    type: flatfile\n    key: CBIN\n" +
		"  - name: eng\n    type: jira\n    key: ENG\n    url: \"https://example.invalid/browse/{id}\"\n"
	if err := os.WriteFile(filepath.Join(root, ".canary", "project.yaml"), []byte(cfg), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "a.go"), []byte(
		"// CANARY: REQ=CBIN-030; FEATURE=\"F\"; ASPECT=API; STATUS=TESTED; TEST=TestF; DEPENDS_ON=ENG-1; UPDATED=2026-01-01\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	claims := filepath.Join(root, "GAP_ANALYSIS.md")
	if err := os.WriteFile(claims, []byte("✅ CBIN-030 - F\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	rec := record("CBIN-030", "F", "API", commit)
	rec.ProjectID = "CBIN"
	store := writeStore(t, root, []evidence.Record{rec})

	stdout, stderr, code := runVerify(Options{Root: root, ClaimsPath: claims, EvidencePath: store})
	if code != 1 {
		t.Fatalf("exit=%d, want 1 (stderr %q)", code, stderr)
	}
	want := `{"ok":false,"state":"UNKNOWN","code":"EXTERNAL_UNKNOWN","message":"external dependency state unknown"}` + "\n"
	if stdout != want {
		t.Fatalf("stdout=%q want %q", stdout, want)
	}
	if !strings.Contains(stderr, "EXTERNAL_UNKNOWN dep=ENG-1") {
		t.Errorf("stderr must name the unresolved dependency: %q", stderr)
	}

	stdout, stderr, code = runVerify(Options{
		Root: root, ClaimsPath: claims, EvidencePath: store, AllowUnknownExternal: true,
	})
	if code != 0 {
		t.Fatalf("exit=%d with --allow-unknown-external, want 0 (stderr %q)", code, stderr)
	}
	if !strings.Contains(stdout, `"code":"OK"`) {
		t.Fatalf("stdout=%q", stdout)
	}
}

// TestVerifyRun_ProjectOverride proves --project overrides the configured
// project key when matching records.
func TestVerifyRun_ProjectOverride(t *testing.T) {
	root, commit := fixture(t,
		"// CANARY: REQ=CBIN-001; FEATURE=\"F\"; ASPECT=API; STATUS=TESTED; TEST=TestF; UPDATED=2026-01-01\n",
		"✅ CBIN-001 - F\n")
	rec := record("CBIN-001", "F", "API", commit)
	rec.ProjectID = "OTHER"
	store := writeStore(t, root, []evidence.Record{rec})
	claims := filepath.Join(root, "GAP_ANALYSIS.md")

	if _, _, code := runVerify(Options{Root: root, ClaimsPath: claims, EvidencePath: store}); code != 1 {
		t.Fatalf("exit=%d, want 1 without the override", code)
	}
	if _, _, code := runVerify(Options{
		Root: root, ClaimsPath: claims, EvidencePath: store, ProjectID: "OTHER",
	}); code != 0 {
		t.Fatalf("exit=%d with --project OTHER, want 0", code)
	}
}
