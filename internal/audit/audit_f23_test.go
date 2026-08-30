// Copyright (c) 2025 by Developer Network.
//
// For more details, see the LICENSE file in the root directory of this
// source code repository or contact Developer Network at info@devnw.com.

package audit

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// peerlessRepoWithDep builds the F-23 fixture: a git repository whose only
// actionable requirement (CBIN-002) depends on EXT-001, an id owned by a
// configured external (ticket) source for which nothing on disk says
// anything -- no remote-status cache, no peer export. EXT-001's state is
// therefore unknown, not satisfied.
func peerlessRepoWithDep(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	initGitRepo(t, root)
	if err := os.MkdirAll(filepath.Join(root, ".canary"), 0o750); err != nil {
		t.Fatal(err)
	}
	projectYAML := "project:\n  name: \"demo\"\n  key: \"CBIN\"\n" +
		"sources:\n" +
		"  - name: core\n    type: flatfile\n    key: CBIN\n" +
		"  - name: ext\n    type: jira\n    key: EXT\n"
	if err := os.WriteFile(filepath.Join(root, ".canary", "project.yaml"), []byte(projectYAML), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "a.go"), []byte(
		"// CANARY: REQ=CBIN-002; FEATURE=\"F\"; ASPECT=API; STATUS=STUB; DEPENDS_ON=EXT-001; UPDATED=2026-01-01\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	gitCommitAll(t, root)
	return root
}

// TestAuditF23 covers the `next` half of F-23: a dependency whose state
// cannot be determined from disk blocks selection by default. "Unknown"
// used to mean "carry on", which handed an agent work whose prerequisite
// might not exist. Callers who accept that risk say so explicitly with
// --allow-unknown-external.
func TestAuditF23(t *testing.T) {
	bin := buildCanary(t)
	root := peerlessRepoWithDep(t)

	got := runNextJSON(t, root, bin, "next", "--format", "json")
	if got.ReqID == "CBIN-002" {
		t.Fatalf("unknown external state did not block: %+v", got)
	}
	// Blocked work is not finished work: the empty answer must say which it
	// is. (runNextJSON already fails on any "completed" in the output.)
	if !strings.Contains(got.Message, "blocked by unmet dependencies") {
		t.Errorf("Message = %q, want the blocked-dependency explanation", got.Message)
	}

	allowed := runNextJSON(t, root, bin, "next", "--format", "json", "--allow-unknown-external")
	if allowed.ReqID != "CBIN-002" {
		t.Fatalf("--allow-unknown-external did not unblock: %+v", allowed)
	}
}

// TestAuditF23_StrictExternalFlagIsGone proves the old opt-in flag was
// removed rather than left inert: blocking on unknown state is now the
// default, so a flag that used to request it would be a lie either way it
// behaved.
func TestAuditF23_StrictExternalFlagIsGone(t *testing.T) {
	bin := buildCanary(t)
	root := peerlessRepoWithDep(t)

	out := runExpectFail(t, root, bin, "next", "--strict-external")
	if out == "" {
		t.Error("expected an unknown-flag diagnostic")
	}
}

// TestAuditF23_PeerWithoutVerifiedExportIsUnknown proves the peer half of
// F-23: a peer's status.json that carries declarations but no verification
// export cannot say a requirement is done, so it resolves unknown -- and an
// unknown dependency blocks.
func TestAuditF23_PeerWithoutVerifiedExportIsUnknown(t *testing.T) {
	bin := buildCanary(t)
	root := t.TempDir()
	initGitRepo(t, root)
	if err := os.MkdirAll(filepath.Join(root, ".canary"), 0o750); err != nil {
		t.Fatal(err)
	}
	projectYAML := "project:\n  name: \"demo\"\n  key: \"CBIN\"\n" +
		"sources:\n" +
		"  - name: core\n    type: flatfile\n    key: CBIN\n" +
		"  - name: ext\n    type: jira\n    key: EXT\n" +
		"peers:\n  - name: upstream\n    root: \"peer\"\n"
	if err := os.WriteFile(filepath.Join(root, ".canary", "project.yaml"), []byte(projectYAML), 0o644); err != nil {
		t.Fatal(err)
	}
	// A legacy peer export: TESTED declared, nothing verified.
	peerDir := filepath.Join(root, "peer")
	if err := os.MkdirAll(peerDir, 0o750); err != nil {
		t.Fatal(err)
	}
	legacy := `{"requirements":[{"id":"EXT-001","features":[{"feature":"F","aspect":"API","status":"TESTED"}]}]}`
	if err := os.WriteFile(filepath.Join(peerDir, "status.json"), []byte(legacy), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "a.go"), []byte(
		"// CANARY: REQ=CBIN-002; FEATURE=\"F\"; ASPECT=API; STATUS=STUB; DEPENDS_ON=EXT-001; UPDATED=2026-01-01\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	gitCommitAll(t, root)

	got := runNextJSON(t, root, bin, "next", "--format", "json")
	if got.ReqID == "CBIN-002" {
		t.Fatalf("a peer declaration without verification unblocked selection: %+v", got)
	}
}
