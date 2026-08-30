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
	"testing"

	"devnw.dev/canary/pkg/evidence"
)

// TestVerifyEvidencePathResolvesUnderRoot covers review finding F1: `canary
// verify --root <dir>` must resolve the default (relative) evidence store
// path under --root, not under the process's current working directory. A
// project verified from outside its own directory (e.g. from a CI job whose
// CWD is the pipeline workspace root) must find evidence recorded at
// <root>/.canary/evidence.json even though that path does not exist relative
// to CWD.
func TestVerifyEvidencePathResolvesUnderRoot(t *testing.T) {
	bin := buildCanary(t)

	root := t.TempDir()
	initGitRepo(t, root)
	if err := os.WriteFile(filepath.Join(root, "a.go"), []byte(
		"// CANARY: REQ=CBIN-001; FEATURE=\"F\"; ASPECT=API; STATUS=TESTED; TEST=TestF; UPDATED=2026-01-01\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	claims := filepath.Join(root, "GAP_ANALYSIS.md")
	if err := os.WriteFile(claims, []byte("✅ CBIN-001 - F\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	gitCommitAll(t, root)
	commit := headCommit(t, root)

	evDir := filepath.Join(root, ".canary")
	if err := os.MkdirAll(evDir, 0o755); err != nil {
		t.Fatal(err)
	}
	// Evidence lives ONLY under root/.canary/evidence.json (the default
	// relative path) -- never at any path relative to the subprocess's CWD.
	rec := evidence.Record{
		ProjectID: "default", RequirementID: "CBIN-001", Feature: "F", Aspect: "API",
		TestID: "TestF", Command: "go test ./...", Result: "PASS", CommitSHA: commit,
		ObservedAt: "2026-01-01T00:00:00Z", Runner: "local",
		ArtifactDigest: "sha256:" + "abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789",
	}
	data, err := json.Marshal(evidence.File{SchemaVersion: 1, Records: []evidence.Record{rec}})
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(evDir, "evidence.json"), data, 0o644); err != nil {
		t.Fatal(err)
	}

	// The subprocess's CWD is a directory that is NOT root and holds no
	// .canary/evidence.json of its own -- if verify still resolved the
	// evidence path relative to CWD, this would fail with EVIDENCE_MISSING.
	cwd := t.TempDir()

	cmd := exec.Command(bin, "verify", "--root", root, "--claims", claims, "--format", "json")
	cmd.Dir = cwd
	out, err := cmd.Output()
	if err != nil {
		if ee, ok := err.(*exec.ExitError); ok {
			t.Fatalf("verify failed: %v\nstdout=%s\nstderr=%s", err, out, ee.Stderr)
		}
		t.Fatalf("verify failed: %v\nstdout=%s", err, out)
	}
	want := `{"ok":true,"state":"VERIFIED","code":"OK","message":"all claims verified"}` + "\n"
	if string(out) != want {
		t.Fatalf("stdout=%q want %q", out, want)
	}
}
