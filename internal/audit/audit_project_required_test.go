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

	"devnw.dev/canary/pkg/storage"
)

// wantProjectRequired is the exact contract line both commands must emit --
// byte for byte, alone on stdout, because the caller is a program parsing it.
const wantProjectRequired = `{"ok":false,"code":"PROJECT_REQUIRED","message":"duplicate requirement id across projects; pass --project"}` + "\n"

// runContract runs the binary in root and asserts it refused with the
// PROJECT_REQUIRED contract: exit 2, that one line on stdout, nothing else.
func runContract(t *testing.T, root, bin string, args ...string) {
	t.Helper()
	cmd := exec.Command(bin, args...) //nolint:gosec // test-built binary
	cmd.Dir = root
	home := t.TempDir()
	cmd.Env = append(os.Environ(), "HOME="+home, "USERPROFILE="+home)
	out, err := cmd.Output()

	var ee *exec.ExitError
	if !errors.As(err, &ee) {
		t.Fatalf("canary %v: want exit 2, got %v (stdout %q)", args, err, out)
	}
	if ee.ExitCode() != 2 {
		t.Fatalf("canary %v: want exit 2, got %d (stdout %q, stderr %q)", args, ee.ExitCode(), out, ee.Stderr)
	}
	if string(out) != wantProjectRequired {
		t.Fatalf("canary %v stdout: got %q, want %q", args, out, wantProjectRequired)
	}
	// The refusal returns an error up through cobra rather than exiting
	// inside RunE, so it must be silenced on the way out: neither the
	// sentinel's text nor the command's usage block belongs in the output.
	if stderr := string(ee.Stderr); strings.Contains(stderr, "Usage:") || strings.Contains(stderr, "Error:") {
		t.Fatalf("canary %v leaked cobra output on stderr: %q", args, stderr)
	}
}

// seedAmbiguousDep builds a repository whose index holds CBIN-100 (the work
// item, in one project) depending on CBIN-200, which exists under two
// projects -- so resolving that dependency is exactly the ambiguity
// PROJECT_REQUIRED exists to refuse.
func seedAmbiguousDep(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	db := openDBAt(t, dbPathIn(root))

	err := db.UpsertToken(&storage.Token{
		ReqID: "CBIN-100", Feature: "Work", Aspect: "API", Status: "STUB",
		FilePath: "work.go", LineNumber: 1, Priority: 1,
		UpdatedAt: "2026-01-01", RawToken: "x",
		IndexedAt: "2026-01-01T00:00:00Z", ProjectID: "projA",
		DependsOn: "CBIN-200",
	})
	if err != nil {
		t.Fatalf("seed CBIN-100: %v", err)
	}
	seedToken(t, db, "projA", "CBIN-200")
	seedToken(t, db, "projB", "CBIN-200")

	if cerr := db.Close(); cerr != nil {
		t.Fatalf("close db: %v", cerr)
	}
	return root
}

// TestProjectRequiredNext pins the `next` half of the swallowed-contract
// finding: dependency resolution called GetTokensByReqID without inspecting
// its error and treated any failure as "this dependency is unresolved". A
// PROJECT_REQUIRED refusal therefore blocked every candidate, and next
// reported "all requirements completed" and exited 0 -- the worst possible
// answer, because it is indistinguishable from success. The refusal must
// reach the surface as the contract instead.
func TestProjectRequiredNext(t *testing.T) {
	bin := buildCanary(t)
	root := seedAmbiguousDep(t)

	runContract(t, root, bin, "next", "--dry-run")
}

// TestProjectRequiredNextScopedStillAnswers proves the propagation did not
// turn a legitimately blocked dependency into an error: with --project
// naming the scope, there is no ambiguity and the command answers normally.
func TestProjectRequiredNextScopedStillAnswers(t *testing.T) {
	bin := buildCanary(t)
	root := seedAmbiguousDep(t)

	// CBIN-200 is STUB in projA, so CBIN-100 is genuinely blocked and next
	// reports no available work -- but it exits 0 and says so, rather than
	// refusing.
	out := run(t, root, bin, "next", "--dry-run", "--project", "projA")
	if out == "" {
		t.Fatal("scoped next produced no output")
	}
}

// TestProjectRequiredDepsCheck pins the `deps check` half: the database
// token provider returned an empty slice on ANY error, so a PROJECT_REQUIRED
// refusal was reported as "dependency missing" -- inventing an answer out of
// a question canary had declined to answer.
func TestProjectRequiredDepsCheck(t *testing.T) {
	bin := buildCanary(t)
	root := seedAmbiguousDep(t)

	specDir := filepath.Join(root, ".canary", "specs", "CBIN-100-work")
	if err := os.MkdirAll(specDir, 0o750); err != nil {
		t.Fatalf("mkdir spec dir: %v", err)
	}
	spec := "# CBIN-100\n\n## Dependencies\n\n- CBIN-200 (the ambiguous one)\n\n## Notes\n"
	if err := os.WriteFile(filepath.Join(specDir, "spec.md"), []byte(spec), 0o600); err != nil {
		t.Fatalf("write spec.md: %v", err)
	}

	runContract(t, root, bin, "deps", "check", "CBIN-100")
}
