package canaryscan

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

// TestCANARY_FIXWAVE1_ScanReadErrorFailsStrict proves a file the scan could
// not read fails a --strict run: a partial scan must never be reported as a
// clean one.
//
// (This replaces the former --update-stale rewrite-walk variant of this test.
// --update-stale no longer walks the tree to rewrite tokens, so the only walk
// that can discover a read error is Scan's own.)
func TestCANARY_FIXWAVE1_ScanReadErrorFailsStrict(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("running as root: permission bits do not deny reads")
	}
	t.Setenv("CANARY_TEST_TIMESTAMP", "2026-08-30T00:00:00Z")

	root := t.TempDir()
	ok := "package x\n" +
		"// CANARY: REQ=CBIN-050; FEATURE=\"X\"; ASPECT=API; STATUS=IMPL; UPDATED=2026-08-30\n"
	if err := os.WriteFile(filepath.Join(root, "ok.go"), []byte(ok), 0o600); err != nil {
		t.Fatal(err)
	}
	secret := filepath.Join(root, "secret.dat")
	if err := os.WriteFile(secret, []byte("whatever\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(secret, 0o000); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chmod(secret, 0o600) })

	outPath := filepath.Join(t.TempDir(), "status.json")
	var stdout, stderr bytes.Buffer
	exitCode := Run(Config{Root: root, Out: outPath, Strict: true}, &stdout, &stderr)

	if exitCode != 2 {
		t.Errorf("exit code = %d, want 2 (an unreadable file must fail --strict); stdout=%q stderr=%q",
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
		t.Errorf("rep.Issues = %+v, want a read_error issue for %s", rep.Issues, secret)
	}
}
