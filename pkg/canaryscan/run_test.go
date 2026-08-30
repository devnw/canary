package canaryscan

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
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

// TestCANARY_FIXWAVE1_MalformedEvidenceFailsPlainScan proves a plain scan --
// no --verify, no --strict -- refuses to finish over an evidence store it
// cannot parse. Exit 3 is canary's parse/IO code, and the reason reaches
// stderr: silently treating unreadable evidence as "nothing proven yet" would
// hide tampering behind ordinary-looking output.
func TestCANARY_FIXWAVE1_MalformedEvidenceFailsPlainScan(t *testing.T) {
	root := t.TempDir()
	tok := "package x\n" +
		"// CANARY: REQ=CBIN-060; FEATURE=\"X\"; ASPECT=API; STATUS=IMPL; UPDATED=2026-08-30\n"
	if err := os.WriteFile(filepath.Join(root, "ok.go"), []byte(tok), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(root, ".canary"), 0o750); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, ".canary", "evidence.json"), []byte("{not json"), 0o600); err != nil {
		t.Fatal(err)
	}

	outPath := filepath.Join(t.TempDir(), "status.json")
	var stdout, stderr bytes.Buffer
	code := Run(Config{Root: root, Out: outPath}, &stdout, &stderr)

	if code != 3 {
		t.Errorf("exit code = %d, want 3 (unparseable evidence is an error, not an absence); stdout=%q stderr=%q",
			code, stdout.String(), stderr.String())
	}
	if !strings.Contains(stderr.String(), "CANARY_PARSE_ERROR") {
		t.Errorf("stderr = %q, want a CANARY_PARSE_ERROR line naming the failure", stderr.String())
	}
	if !strings.Contains(stderr.String(), "evidence") {
		t.Errorf("stderr = %q, want the message to name the evidence store", stderr.String())
	}
}
