package canaryscan

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"regexp"
	"testing"
)

// TestCANARY_FIXWAVE1_UpdateStaleReadErrorFailsStrict proves that a read
// error discovered only during the --update-stale rewrite walk (never seen
// by Scan itself) still fails a --strict run, the same way a read error
// discovered by Scan does.
//
// The repro exploits a real behavioral difference in how the skip regex is
// applied: Scan (and the sub-walks it calls, ScanDiagramRefs and
// ScanMigrateNotes) match the skip regex against the path as built during
// filepath.WalkDir(root, ...) -- which is root-prefixed -- while
// UpdateStaleTokens (update.go) matches it against the root-relative path.
// A skip regex anchored to the (absolute) root prefix therefore excludes a
// directory from every Scan-side walk while leaving UpdateStaleTokens free
// to walk into it. An unreadable file placed there is thus invisible to
// Scan but gets opened (and fails) only by the update walk.
func TestCANARY_FIXWAVE1_UpdateStaleReadErrorFailsStrict(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("running as root: permission bits do not deny reads")
	}
	t.Setenv("CANARY_TEST_TIMESTAMP", "2026-08-30T00:00:00Z")

	root := t.TempDir()

	// A stale TESTED token so --update-stale actually runs its rewrite walk.
	stale := "package x\n" +
		"// CANARY: REQ=CBIN-050; FEATURE=\"X\"; ASPECT=API; STATUS=TESTED; TEST=TestX; UPDATED=2020-01-01\n"
	if err := os.WriteFile(filepath.Join(root, "ok.go"), []byte(stale), 0o600); err != nil {
		t.Fatal(err)
	}

	// An unreadable fixture in a directory excluded from Scan's walks (see
	// skipExpr below) but not from UpdateStaleTokens'.
	hiddenDir := filepath.Join(root, "hidden")
	if err := os.MkdirAll(hiddenDir, 0o750); err != nil {
		t.Fatal(err)
	}
	secret := filepath.Join(hiddenDir, "secret.dat")
	if err := os.WriteFile(secret, []byte("whatever\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(secret, 0o000); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chmod(secret, 0o600) })

	// Anchored to the absolute root prefix: matches the root-joined "path"
	// Scan's walks see, but not the root-relative "rel" UpdateStaleTokens
	// matches against.
	skip, err := regexp.Compile(regexp.QuoteMeta(root) + `/hidden(/|$)`)
	if err != nil {
		t.Fatal(err)
	}

	outPath := filepath.Join(t.TempDir(), "status.json")
	cfg := Config{
		Root:        root,
		Out:         outPath,
		Strict:      true,
		UpdateStale: true,
		SkipRegex:   skip,
	}
	var stdout, stderr bytes.Buffer
	exitCode := Run(cfg, &stdout, &stderr)

	if exitCode != 2 {
		t.Errorf("exit code = %d, want 2 (an unreadable file during --update-stale must fail --strict); stdout=%q stderr=%q",
			exitCode, stdout.String(), stderr.String())
	}

	data, err := os.ReadFile(outPath)
	if err != nil {
		t.Fatalf("reading %s: %v", outPath, err)
	}
	var rep Report
	if err := json.Unmarshal(data, &rep); err != nil {
		t.Fatalf("unmarshal status.json: %v", err)
	}
	if !hasIssue(rep.Issues, "secret.dat", IssueReadError) {
		t.Errorf("rep.Issues = %+v, want a read_error issue for %s (updateIssues must merge into the final report)", rep.Issues, secret)
	}
}
