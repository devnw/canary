// Copyright (c) 2025 by Developer Network.
//
// For more details, see the LICENSE file in the root directory of this
// source code repository or contact Developer Network at info@devnw.com.

package audit

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestAuditF03 covers F-03: `canary init` used to rewrite every file it
// authors on every run, silently destroying whatever the user had put in
// GAP_ANALYSIS.md. Re-running init without --force must keep the customized
// bytes, say so, and still succeed.
func TestAuditF03(t *testing.T) {
	bin := buildCanary(t)
	root := t.TempDir()

	run(t, root, bin, "init", "--key", "CBIN")

	gap := filepath.Join(root, "GAP_ANALYSIS.md")
	custom := []byte("# my customized gap analysis\n")
	if err := os.WriteFile(gap, custom, 0o644); err != nil {
		t.Fatalf("write custom GAP_ANALYSIS.md: %v", err)
	}

	out := run(t, root, bin, "init", "--key", "CBIN")

	got, err := os.ReadFile(gap) //nolint:gosec // test-controlled path
	if err != nil {
		t.Fatalf("read GAP_ANALYSIS.md: %v", err)
	}
	if !bytes.Equal(got, custom) {
		t.Fatalf("re-init clobbered user file without --force:\n%s", got)
	}
	if !strings.Contains(out, "kept existing") {
		t.Fatalf("re-init skipped a user file silently; output:\n%s", out)
	}
}

// TestAuditF03Force proves --force is the documented escape hatch: it does
// overwrite the customized file, and keeps the prior bytes in a .bak.
func TestAuditF03Force(t *testing.T) {
	bin := buildCanary(t)
	root := t.TempDir()

	run(t, root, bin, "init", "--key", "CBIN")

	gap := filepath.Join(root, "GAP_ANALYSIS.md")
	custom := []byte("# my customized gap analysis\n")
	if err := os.WriteFile(gap, custom, 0o644); err != nil {
		t.Fatalf("write custom GAP_ANALYSIS.md: %v", err)
	}

	run(t, root, bin, "init", "--key", "CBIN", "--force")

	got, err := os.ReadFile(gap) //nolint:gosec // test-controlled path
	if err != nil {
		t.Fatalf("read GAP_ANALYSIS.md: %v", err)
	}
	if bytes.Equal(got, custom) {
		t.Fatal("--force did not overwrite the customized file")
	}
	bak, err := os.ReadFile(gap + ".bak") //nolint:gosec // test-controlled path
	if err != nil {
		t.Fatalf("read GAP_ANALYSIS.md.bak: %v", err)
	}
	if !bytes.Equal(bak, custom) {
		t.Fatalf("backup = %q, want the pre-overwrite bytes", bak)
	}
}
