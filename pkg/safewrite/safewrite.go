// Copyright (c) 2025 by Developer Network.
//
// For more details, see the LICENSE file in the root directory of this
// source code repository or contact Developer Network at info@devnw.com.

// Package safewrite is the single way this tool replaces a file on disk.
//
// Every write -- the target's new content and the .bak holding its prior
// content alike -- is staged in a temp file beside the target, fsynced,
// optionally validated, and only then renamed into place, so a crash or an
// error partway through can never leave a truncated or half-written file. The
// containing directory is fsynced after the rename where the platform supports
// it, so the rename itself is durable and not just the bytes behind it. A
// target that already exists with different bytes is refused unless the
// caller passes Force -- a tool that authors files into a user's repository
// must not silently destroy the user's edits. Paths are confined to a caller-supplied
// root with symlinks resolved, so a symlink planted in a project can never
// redirect a write outside it.
//
// CANARY: REQ=ENG-4330; FEATURE="SafeWrite"; ASPECT=Storage; STATUS=TESTED; TEST=TestWriteNewFile,TestWriteIdenticalNoop,TestWriteRefusesWithoutForce,TestWriteForceBacksUpAndPreservesMode,TestWriteFailureLeavesOriginal,TestConfineRejectsSymlinkEscape,TestConfineAcceptsPathInsideRoot; UPDATED=2026-08-30
package safewrite

import (
	"bytes"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

var (
	// ErrWouldReplace reports that the target exists with different content
	// and Force was not set. Callers typically print a notice and skip.
	ErrWouldReplace = errors.New("target exists with different content; pass force")
	// ErrRootEscape reports that the target, with symlinks resolved, lies
	// outside the confinement root.
	ErrRootEscape = errors.New("path resolves outside root")
)

// stagePrefix names every staging file this package creates. It is a fixed,
// dotted prefix so a caller can recognize (and a test can assert the absence
// of) leftover staging files.
const stagePrefix = ".canary-stage-"

// Options configures Write.
type Options struct {
	// Root confines the write. Required: an unset root is an error rather
	// than an implicit "anywhere".
	Root string
	// Force allows replacing an existing file whose bytes differ.
	Force bool
	// Backup writes the prior bytes to <path>.bak before replacing.
	Backup bool
	// Validate, when set, inspects the staged bytes before the rename. A
	// non-nil error aborts the write with the original file untouched.
	Validate func([]byte) error
}

// Result describes what Write did.
type Result struct {
	// Written is false when the target already held exactly these bytes.
	Written bool
	// Replaced is true when an existing file's content was replaced.
	Replaced bool
	// BackupPath is the .bak file written, when Options.Backup asked for one.
	BackupPath string
	// Diff summarizes old-to-new line changes. It is set whenever an
	// existing file differs -- including on the ErrWouldReplace refusal, so
	// a caller can show what it declined to overwrite.
	Diff string
}

// Write replaces path's content with data, atomically and confined to
// Options.Root.
//
// The target is left byte-identical on every failure path. When the target
// already exists its own permissions are preserved and mode is ignored; mode
// applies to newly created files only.
func Write(path string, data []byte, mode fs.FileMode, o Options) (Result, error) {
	var res Result

	real, err := Confine(o.Root, path)
	if err != nil {
		return res, err
	}

	old, oldMode, exists, err := inspect(real)
	if err != nil {
		return res, err
	}

	if exists {
		if bytes.Equal(old, data) {
			return res, nil
		}
		res.Diff = diffLines(old, data)
		if !o.Force {
			return res, fmt.Errorf("%w: %s", ErrWouldReplace, path)
		}
		mode = oldMode
	}

	dir := filepath.Dir(real)
	if err := os.MkdirAll(dir, 0o750); err != nil {
		return res, fmt.Errorf("create %s: %w", dir, err)
	}
	tmpPath, err := stage(dir, data, mode)
	if err != nil {
		return res, err
	}
	defer func() { _ = os.Remove(tmpPath) }() // no-op once the rename succeeds

	if o.Validate != nil {
		if err := o.Validate(data); err != nil {
			return res, fmt.Errorf("validate %s: %w", path, err)
		}
	}

	if exists && o.Backup {
		// The backup is staged and renamed like the target itself: it is the
		// only surviving copy of the bytes about to be replaced, so a crash
		// mid-backup must not be able to leave it truncated.
		bak := real + ".bak"
		bakTmp, berr := stage(dir, old, oldMode)
		if berr != nil {
			return res, fmt.Errorf("stage backup %s: %w", bak, berr)
		}
		if rerr := os.Rename(bakTmp, bak); rerr != nil {
			_ = os.Remove(bakTmp)
			return res, fmt.Errorf("write backup %s: %w", bak, rerr)
		}
		res.BackupPath = bak
	}

	if err := os.Rename(tmpPath, real); err != nil {
		return res, fmt.Errorf("rename into %s: %w", path, err)
	}
	syncDir(dir)

	res.Written = true
	res.Replaced = exists
	return res, nil
}

// inspect reads the target's current bytes and permissions. A missing target
// is reported through exists=false rather than an error; anything that is not
// a regular file (a directory, a device, a dangling symlink) is refused,
// because renaming over it is not a content replacement.
func inspect(real string) (old []byte, mode fs.FileMode, exists bool, err error) {
	info, serr := os.Lstat(real)
	if serr != nil {
		if errors.Is(serr, fs.ErrNotExist) {
			return nil, 0, false, nil
		}
		return nil, 0, false, serr
	}
	if !info.Mode().IsRegular() {
		return nil, 0, false, fmt.Errorf("safewrite: %s is not a regular file", real)
	}
	old, err = os.ReadFile(real) //nolint:gosec // real is the confined target
	if err != nil {
		return nil, 0, false, err
	}
	return old, info.Mode().Perm(), true, nil
}

// stage writes data to a fresh temp file in dir, fsyncs it, and returns its
// path. The temp file is removed on any error so a failed write never leaves
// staging litter behind.
func stage(dir string, data []byte, mode fs.FileMode) (path string, err error) {
	tmp, err := os.CreateTemp(dir, stagePrefix+"*")
	if err != nil {
		return "", fmt.Errorf("stage in %s: %w", dir, err)
	}
	path = tmp.Name()
	defer func() {
		if err != nil {
			_ = tmp.Close()
			_ = os.Remove(path)
			path = ""
		}
	}()

	if _, err = tmp.Write(data); err != nil {
		return path, fmt.Errorf("write staging file: %w", err)
	}
	if err = tmp.Chmod(mode); err != nil {
		return path, fmt.Errorf("chmod staging file: %w", err)
	}
	// fsync before the rename: without it the rename can be durable while
	// the content behind it is not, which is exactly the truncated-file
	// outcome this package exists to prevent.
	if err = tmp.Sync(); err != nil {
		return path, fmt.Errorf("sync staging file: %w", err)
	}
	if err = tmp.Close(); err != nil {
		return path, fmt.Errorf("close staging file: %w", err)
	}
	return path, nil
}

// syncDir fsyncs a directory so the renames just performed inside it are
// themselves durable, not merely the bytes they point at: without it a crash
// can lose the rename while keeping the staging file.
//
// This is best effort by design. Opening a directory as a file, and fsyncing
// one, are not portable -- Windows refuses both -- and there is nothing useful
// a caller could do about a failure here: the content is already fsynced and
// the rename already succeeded, so the write is correct either way. Errors are
// therefore ignored rather than surfaced.
func syncDir(dir string) {
	f, err := os.Open(dir) //nolint:gosec // dir is the confined target's parent
	if err != nil {
		return
	}
	_ = f.Sync()
	_ = f.Close()
}

// Confine resolves symlinks in path and its parents and returns the resolved
// path, or ErrRootEscape when it does not lie within the resolved root.
// Components that do not exist yet are appended verbatim: a file about to be
// created cannot be resolved, but every existing directory leading to it can.
//
// Note the ordering: filepath.Abs cleans the path lexically -- collapsing
// "." and ".." textually -- before any symlink is resolved. So "root/link/../x"
// is evaluated as "root/x", not as "<link's target>/../x". That is strictly
// more restrictive than resolving ".." against the link's real parent: it can
// reject a path that a symlink-first walk would have allowed inside the root,
// but it can never let one out of it, which is the only direction that matters
// for confinement.
func Confine(root, path string) (string, error) {
	if strings.TrimSpace(root) == "" {
		return "", errors.New("safewrite: confinement root is required")
	}
	absRoot, err := filepath.Abs(root)
	if err != nil {
		return "", err
	}
	realRoot, err := resolve(absRoot)
	if err != nil {
		return "", fmt.Errorf("resolve root %s: %w", root, err)
	}
	absPath, err := filepath.Abs(path)
	if err != nil {
		return "", err
	}
	realPath, err := resolve(absPath)
	if err != nil {
		return "", fmt.Errorf("resolve %s: %w", path, err)
	}
	if realPath != realRoot && !strings.HasPrefix(realPath, realRoot+string(filepath.Separator)) {
		return "", fmt.Errorf("%w: %s is not under %s", ErrRootEscape, realPath, realRoot)
	}
	return realPath, nil
}

// resolve evaluates symlinks over the longest existing prefix of p and
// re-appends the components that do not exist yet.
func resolve(p string) (string, error) {
	cur := filepath.Clean(p)
	rest := ""
	for {
		resolved, err := filepath.EvalSymlinks(cur)
		if err == nil {
			if rest == "" {
				return resolved, nil
			}
			return filepath.Join(resolved, rest), nil
		}
		if !errors.Is(err, fs.ErrNotExist) {
			return "", err
		}
		parent := filepath.Dir(cur)
		if parent == cur {
			return "", err
		}
		rest = filepath.Join(filepath.Base(cur), rest)
		cur = parent
	}
}

// diffLines renders a minimal, deterministic line diff of old against
// updated: every line the replacement drops as "-<line>", every line it
// introduces as "+<line>". Removals come before additions and each group is
// sorted, so the same pair of files always produces the same text -- this is a
// notice for a human, not a patch to apply, and stable output keeps it
// diffable in logs. Repeated lines are counted, so dropping one of three
// identical lines shows up as exactly one removal.
//
// The rendering reaches callers through Result.Diff; nothing outside this
// package needs to diff arbitrary byte slices, so this stays unexported.
func diffLines(old, updated []byte) string {
	removed := surplus(lineCounts(old), lineCounts(updated))
	added := surplus(lineCounts(updated), lineCounts(old))
	sort.Strings(removed)
	sort.Strings(added)

	out := make([]string, 0, len(removed)+len(added))
	for _, l := range removed {
		out = append(out, "-"+l)
	}
	for _, l := range added {
		out = append(out, "+"+l)
	}
	return strings.Join(out, "\n")
}

// lineCounts counts each distinct line in b. A single trailing newline is not
// a line of its own.
func lineCounts(b []byte) map[string]int {
	s := strings.TrimSuffix(string(b), "\n")
	counts := map[string]int{}
	if s == "" {
		return counts
	}
	for _, l := range strings.Split(s, "\n") {
		counts[l]++
	}
	return counts
}

// surplus expands, for each line, the number of occurrences a has beyond b.
func surplus(a, b map[string]int) []string {
	var out []string
	for line, n := range a {
		for i := b[line]; i < n; i++ {
			out = append(out, line)
		}
	}
	return out
}
