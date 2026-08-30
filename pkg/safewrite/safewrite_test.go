// Copyright (c) 2025 by Developer Network.
//
// For more details, see the LICENSE file in the root directory of this
// source code repository or contact Developer Network at info@devnw.com.

package safewrite_test

import (
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"devnw.dev/canary/pkg/safewrite"
)

// readFile returns path's bytes as a string, failing the test if it cannot.
func readFile(t *testing.T, path string) string {
	t.Helper()
	b, err := os.ReadFile(path) //nolint:gosec // test-controlled path
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	return string(b)
}

// entryNames lists dir's entries, sorted by the OS, for staging-litter checks.
func entryNames(t *testing.T, dir string) []string {
	t.Helper()
	ents, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("readdir %s: %v", dir, err)
	}
	names := make([]string, 0, len(ents))
	for _, e := range ents {
		names = append(names, e.Name())
	}
	return names
}

// TestWriteNewFile proves a first write creates the file with the requested
// mode and reports Written without Replaced.
func TestWriteNewFile(t *testing.T) {
	root := t.TempDir()
	p := filepath.Join(root, "new.md")

	res, err := safewrite.Write(p, []byte("hello\n"), 0o644, safewrite.Options{Root: root})
	if err != nil {
		t.Fatalf("Write: %v", err)
	}
	if !res.Written || res.Replaced {
		t.Fatalf("Result = %+v, want Written=true Replaced=false", res)
	}
	if got := readFile(t, p); got != "hello\n" {
		t.Fatalf("content = %q, want %q", got, "hello\n")
	}
	if runtime.GOOS != "windows" {
		info, err := os.Stat(p)
		if err != nil {
			t.Fatalf("stat: %v", err)
		}
		if perm := info.Mode().Perm(); perm != 0o644 {
			t.Fatalf("mode = %o, want 0644", perm)
		}
	}
	if names := entryNames(t, root); len(names) != 1 {
		t.Fatalf("staging litter: %v", names)
	}
}

// TestWriteIdenticalNoop proves writing the bytes already on disk is a no-op:
// no replacement, no error, and no --force needed.
func TestWriteIdenticalNoop(t *testing.T) {
	root := t.TempDir()
	p := filepath.Join(root, "same.md")
	if err := os.WriteFile(p, []byte("same\n"), 0o600); err != nil {
		t.Fatalf("seed: %v", err)
	}

	res, err := safewrite.Write(p, []byte("same\n"), 0o644, safewrite.Options{Root: root})
	if err != nil {
		t.Fatalf("Write: %v", err)
	}
	if res.Written || res.Replaced {
		t.Fatalf("Result = %+v, want Written=false Replaced=false", res)
	}
	if got := readFile(t, p); got != "same\n" {
		t.Fatalf("content = %q, want unchanged", got)
	}
}

// TestWriteRefusesWithoutForce proves a differing existing file is never
// silently replaced, and that the refusal still reports the diff so a caller
// can show the user what it declined to overwrite.
func TestWriteRefusesWithoutForce(t *testing.T) {
	root := t.TempDir()
	p := filepath.Join(root, "user.md")
	if err := os.WriteFile(p, []byte("mine\n"), 0o644); err != nil {
		t.Fatalf("seed: %v", err)
	}

	res, err := safewrite.Write(p, []byte("theirs\n"), 0o644, safewrite.Options{Root: root})
	if !errors.Is(err, safewrite.ErrWouldReplace) {
		t.Fatalf("err = %v, want ErrWouldReplace", err)
	}
	if res.Written {
		t.Fatalf("Result = %+v, want Written=false", res)
	}
	if got := readFile(t, p); got != "mine\n" {
		t.Fatalf("content = %q, want %q (refused write must not touch the file)", got, "mine\n")
	}
	if !strings.Contains(res.Diff, "-mine") || !strings.Contains(res.Diff, "+theirs") {
		t.Fatalf("Diff = %q, want both the removed and added line", res.Diff)
	}
	if names := entryNames(t, root); len(names) != 1 {
		t.Fatalf("staging litter: %v", names)
	}
}

// TestWriteForceBacksUpAndPreservesMode proves that a forced replacement keeps
// the target's original permissions (not the mode argument) and preserves the
// prior bytes in <path>.bak.
func TestWriteForceBacksUpAndPreservesMode(t *testing.T) {
	root := t.TempDir()
	p := filepath.Join(root, "keep.md")
	if err := os.WriteFile(p, []byte("old\n"), 0o600); err != nil {
		t.Fatalf("seed: %v", err)
	}

	res, err := safewrite.Write(p, []byte("new\n"), 0o644, safewrite.Options{
		Root: root, Force: true, Backup: true,
	})
	if err != nil {
		t.Fatalf("Write: %v", err)
	}
	if !res.Written || !res.Replaced {
		t.Fatalf("Result = %+v, want Written=true Replaced=true", res)
	}
	if got := readFile(t, p); got != "new\n" {
		t.Fatalf("content = %q, want %q", got, "new\n")
	}
	if res.BackupPath != p+".bak" {
		t.Fatalf("BackupPath = %q, want %q", res.BackupPath, p+".bak")
	}
	if got := readFile(t, res.BackupPath); got != "old\n" {
		t.Fatalf("backup content = %q, want %q", got, "old\n")
	}
	if runtime.GOOS == "windows" {
		return // chmod is unreliable there; the content assertions above carry the test.
	}
	info, err := os.Stat(p)
	if err != nil {
		t.Fatalf("stat: %v", err)
	}
	if perm := info.Mode().Perm(); perm != 0o600 {
		t.Fatalf("mode = %o, want 0600 (an existing file keeps its own mode)", perm)
	}
}

// TestWriteFailureLeavesOriginal proves a validation failure is fully
// transactional: the target keeps its original bytes and no staging file
// survives.
func TestWriteFailureLeavesOriginal(t *testing.T) {
	root := t.TempDir()
	p := filepath.Join(root, "tx.md")
	if err := os.WriteFile(p, []byte("original\n"), 0o644); err != nil {
		t.Fatalf("seed: %v", err)
	}

	boom := errors.New("boom")
	_, err := safewrite.Write(p, []byte("replacement\n"), 0o644, safewrite.Options{
		Root:     root,
		Force:    true,
		Validate: func([]byte) error { return boom },
	})
	if !errors.Is(err, boom) {
		t.Fatalf("err = %v, want the validate error", err)
	}
	if got := readFile(t, p); got != "original\n" {
		t.Fatalf("content = %q, want %q", got, "original\n")
	}
	if names := entryNames(t, root); len(names) != 1 {
		t.Fatalf("staging litter: %v", names)
	}
}

// TestConfineRejectsSymlinkEscape proves a symlink inside the root that points
// outside it cannot be used to write beyond the root.
func TestConfineRejectsSymlinkEscape(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("symlink creation requires elevation on Windows")
	}
	root := t.TempDir()
	outside := t.TempDir()
	link := filepath.Join(root, "escape")
	if err := os.Symlink(outside, link); err != nil {
		t.Fatalf("symlink: %v", err)
	}

	target := filepath.Join(link, "pwned.md")
	if _, err := safewrite.Confine(root, target); !errors.Is(err, safewrite.ErrRootEscape) {
		t.Fatalf("Confine err = %v, want ErrRootEscape", err)
	}
	if _, err := safewrite.Write(target, []byte("x"), 0o644, safewrite.Options{Root: root, Force: true}); !errors.Is(err, safewrite.ErrRootEscape) {
		t.Fatalf("Write err = %v, want ErrRootEscape", err)
	}
	if names := entryNames(t, outside); len(names) != 0 {
		t.Fatalf("wrote outside root: %v", names)
	}
}

// TestConfineAcceptsPathInsideRoot proves the confinement check is not
// vacuous: an ordinary path under the root resolves rather than erroring.
func TestConfineAcceptsPathInsideRoot(t *testing.T) {
	root := t.TempDir()
	got, err := safewrite.Confine(root, filepath.Join(root, "sub", "f.md"))
	if err != nil {
		t.Fatalf("Confine: %v", err)
	}
	if filepath.Base(got) != "f.md" {
		t.Fatalf("Confine = %q, want a path ending in f.md", got)
	}
}
