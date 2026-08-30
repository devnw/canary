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

// TestAuditF11 covers F-11: a managed file whose CANARY markers are malformed
// (here, a START with no END) used to be "repaired" by appending a second,
// duplicate section. Marker structure is now validated before any rewrite, so
// init reports the bad file and leaves it byte-identical.
func TestAuditF11(t *testing.T) {
	bin := buildCanary(t)
	root := t.TempDir()

	bad := []byte("<!-- CANARY:guide:START -->\nno end marker\n")
	claude := filepath.Join(root, "CLAUDE.md")
	if err := os.WriteFile(claude, bad, 0o644); err != nil {
		t.Fatalf("seed CLAUDE.md: %v", err)
	}

	out := runExpectFail(t, root, bin, "init", "--key", "CBIN")

	got, err := os.ReadFile(claude) //nolint:gosec // test-controlled path
	if err != nil {
		t.Fatalf("read CLAUDE.md: %v", err)
	}
	if !bytes.Equal(got, bad) {
		t.Fatalf("malformed-marker file was mutated; output: %s", out)
	}
	if !strings.Contains(out, "CLAUDE.md") {
		t.Fatalf("failure did not name the offending file; output: %s", out)
	}
	// The other managed files still get written: one bad file must not
	// abort the whole bootstrap.
	if _, err := os.Stat(filepath.Join(root, "AGENTS.md")); err != nil {
		t.Fatalf("AGENTS.md was not written despite an unrelated file failing: %v", err)
	}
}

// TestAuditF11NoDuplicateSection proves the append-a-duplicate behavior is
// gone in the healthy case too: a file whose markers are intact is updated in
// place, never grown a second section.
func TestAuditF11NoDuplicateSection(t *testing.T) {
	bin := buildCanary(t)
	root := t.TempDir()

	run(t, root, bin, "init", "--key", "CBIN")
	first, err := os.ReadFile(filepath.Join(root, "CLAUDE.md")) //nolint:gosec // test-controlled path
	if err != nil {
		t.Fatalf("read CLAUDE.md: %v", err)
	}
	run(t, root, bin, "init", "--key", "CBIN")
	second, err := os.ReadFile(filepath.Join(root, "CLAUDE.md")) //nolint:gosec // test-controlled path
	if err != nil {
		t.Fatalf("read CLAUDE.md: %v", err)
	}

	if !bytes.Equal(first, second) {
		t.Fatalf("re-init changed CLAUDE.md:\nfirst:\n%s\nsecond:\n%s", first, second)
	}
	if n := bytes.Count(second, []byte(":START -->")); n != 1 {
		t.Fatalf("CLAUDE.md has %d start markers, want 1:\n%s", n, second)
	}
}
