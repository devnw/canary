// Copyright (c) 2025 by Developer Network.
//
// For more details, see the LICENSE file in the root directory of this
// source code repository or contact Developer Network at info@devnw.com.

package audit

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// TestAuditF01CLI covers the CLI half of F-01: a token that merely *names* a
// test (TEST=...) proves nothing. `canary verify` demands a passing evidence
// record at the current commit, so a STATUS=TESTED declaration with no
// evidence is UNVERIFIED, exits 1, and says which requirement is missing.
func TestAuditF01CLI(t *testing.T) {
	bin := buildCanary(t)
	root := t.TempDir()
	initGitRepo(t, root)
	if err := os.WriteFile(filepath.Join(root, "a.go"), []byte(
		"// CANARY: REQ=CBIN-001; FEATURE=\"F\"; ASPECT=API; STATUS=TESTED; TEST=TestDoesNotExist; UPDATED=2026-01-01\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "GAP_ANALYSIS.md"), []byte("✅ CBIN-001 - F\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	gitCommitAll(t, root)

	cmd := exec.Command(bin, "verify", "--root", root, "--claims", filepath.Join(root, "GAP_ANALYSIS.md"), "--format", "json")
	cmd.Dir = root
	out, err := cmd.Output()
	var ee *exec.ExitError
	if !errors.As(err, &ee) || ee.ExitCode() != 1 {
		t.Fatalf("want exit 1, got %v (stdout %q)", err, out)
	}
	want := `{"ok":false,"state":"UNVERIFIED","code":"EVIDENCE_MISSING","message":"no passing evidence at current commit"}` + "\n"
	if string(out) != want {
		t.Fatalf("stdout:\n%q\nwant:\n%q", out, want)
	}
	if !strings.Contains(string(ee.Stderr), "CBIN-001") {
		t.Fatalf("stderr must name the missing requirement: %q", ee.Stderr)
	}
}
