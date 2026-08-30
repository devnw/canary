// Copyright (c) 2025 by Developer Network.
//
// For more details, see the LICENSE file in the root directory of this
// source code repository or contact Developer Network at info@devnw.com.

package audit

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"devnw.dev/canary/pkg/safewrite"
)

// TestAuditF05 covers F-05: a failure partway through a file rewrite used to
// leave the target truncated or half-migrated. Every write now stages,
// validates, and only then renames, so an injected failure leaves the tree
// exactly as it was -- same bytes, same digest, no staging litter.
func TestAuditF05(t *testing.T) {
	root := t.TempDir()
	p := filepath.Join(root, "f.md")
	if err := os.WriteFile(p, []byte("old"), 0o644); err != nil {
		t.Fatalf("seed: %v", err)
	}
	before := sha256File(t, p)

	_, err := safewrite.Write(p, []byte("new"), 0o644, safewrite.Options{
		Root:     root,
		Force:    true,
		Validate: func([]byte) error { return errors.New("boom") },
	})
	if err == nil {
		t.Fatal("want validate failure")
	}
	if sha256File(t, p) != before {
		t.Fatal("failed write mutated target")
	}
	ents, err := os.ReadDir(root)
	if err != nil {
		t.Fatalf("readdir: %v", err)
	}
	if len(ents) != 1 {
		t.Fatalf("staging litter: %v", ents)
	}
}

// TestAuditF05Confinement proves the same write path refuses to follow a
// symlink out of its root, so a "migration" can never reach a file the
// caller did not confine it to.
func TestAuditF05Confinement(t *testing.T) {
	root := t.TempDir()
	outside := t.TempDir()
	victim := filepath.Join(outside, "victim.md")
	if err := os.WriteFile(victim, []byte("untouched"), 0o644); err != nil {
		t.Fatalf("seed: %v", err)
	}
	if err := os.Symlink(victim, filepath.Join(root, "link.md")); err != nil {
		t.Skipf("symlinks unavailable: %v", err)
	}

	_, err := safewrite.Write(filepath.Join(root, "link.md"), []byte("pwned"), 0o644,
		safewrite.Options{Root: root, Force: true})
	if !errors.Is(err, safewrite.ErrRootEscape) {
		t.Fatalf("err = %v, want ErrRootEscape", err)
	}
	b, err := os.ReadFile(victim) //nolint:gosec // test-controlled path
	if err != nil {
		t.Fatalf("read victim: %v", err)
	}
	if string(b) != "untouched" {
		t.Fatalf("victim = %q, want untouched", b)
	}
}
