// Copyright (c) 2025 by Developer Network.
//
// For more details, see the LICENSE file in the root directory of this
// source code repository or contact Developer Network at info@devnw.com.

package audit

import (
	"bytes"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// reqStateJSON mirrors the fields of drift.ReqState the audit needs to read
// back out of `canary drift --format json`. It is decoded from the top-level
// JSON array the command emits on stdout.
type reqStateJSON struct {
	RequirementID string `json:"requirement_id"`
	State         string `json:"state"`
}

// appendLine appends line (plus a newline) to the file at path, changing its
// bytes — and therefore its content hash — without disturbing any CANARY token
// already in it.
func appendLine(t *testing.T, path, line string) {
	t.Helper()
	f, err := os.OpenFile(path, os.O_APPEND|os.O_WRONLY, 0o600) //nolint:gosec // test-controlled path
	if err != nil {
		t.Fatalf("open %s: %v", path, err)
	}
	defer func() { _ = f.Close() }()
	if _, err := f.WriteString(line + "\n"); err != nil {
		t.Fatalf("append to %s: %v", path, err)
	}
}

// runJSON2 runs the canary binary in root, captures stdout separately from
// stderr, and decodes stdout as the drift JSON array. Keeping stdout clean is
// the point: `--format json` must emit only the ReqState array, so a decoder
// can read it without tripping over a prose banner.
func runJSON2(t *testing.T, root, bin string, args ...string) []reqStateJSON {
	t.Helper()
	cmd := exec.Command(bin, args...)
	cmd.Dir = root
	home := t.TempDir()
	cmd.Env = append(os.Environ(), "HOME="+home, "USERPROFILE="+home)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		t.Fatalf("canary %s: %v\nstderr: %s", strings.Join(args, " "), err, stderr.String())
	}
	var out []reqStateJSON
	if err := json.Unmarshal(bytes.TrimSpace(stdout.Bytes()), &out); err != nil {
		t.Fatalf("decode drift json: %v\nstdout: %q", err, stdout.String())
	}
	return out
}

const (
	f20AContent = "package foo\n\n" +
		"// CANARY: REQ=CBIN-950; FEATURE=\"Foo\"; ASPECT=API; STATUS=IMPL; UPDATED=2026-08-01\n" +
		"func Foo() {}\n"
	f20BContent = "package foo\n\n" +
		"// CANARY: REQ=CBIN-950; FEATURE=\"Bar\"; ASPECT=CLI; STATUS=IMPL; UPDATED=2026-08-01\n" +
		"func Bar() {}\n"
)

// gitRepoWithTwoFileToken builds a git repo where requirement CBIN-950 is
// carried by tokens in TWO files (a.go and b.go), committed and ready to
// index. Two files per requirement is what lets the test prove per-file hash
// comparison: changing only b.go must drift the requirement while a.go stays
// current.
func gitRepoWithTwoFileToken(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	initGitRepo(t, root)
	if err := os.WriteFile(filepath.Join(root, "a.go"), []byte(f20AContent), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "b.go"), []byte(f20BContent), 0o600); err != nil {
		t.Fatal(err)
	}
	gitCommitAll(t, root)
	return root
}

// TestAuditF20 pins the F-20 remediation: drift is CURRENT|DRIFTED|UNKNOWN
// decided by content hash and git availability, never by the UPDATED= date.
//
//   - A same-day change to b.go (its bytes differ from the indexed baseline)
//     is DRIFTED even though nothing about any UPDATED date moved and the
//     commit lands on the same day — the hash catches it.
//   - With b.go restored to its indexed bytes (hashes match again) and the
//     repo's .git removed, only the git leg is undecidable, so the verdict is
//     UNKNOWN, not CURRENT: a git failure never reads as current.
func TestAuditF20(t *testing.T) {
	bin := buildCanary(t)
	root := gitRepoWithTwoFileToken(t)

	run(t, root, bin, "index", "--root", ".")

	// same-day second-file change -> DRIFTED (by hash, not date)
	appendLine(t, filepath.Join(root, "b.go"), "// changed")
	gitCommitAll(t, root)
	out := runJSON2(t, root, bin, "drift", "--format", "json")
	if len(out) == 0 || out[0].State != "DRIFTED" {
		t.Fatalf("same-day change not DRIFTED: %+v", out)
	}

	// git failure -> UNKNOWN. Restore b.go to its indexed bytes so the hash
	// leg is decidable-and-matching, then remove .git so only the git leg is
	// undecidable.
	if err := os.WriteFile(filepath.Join(root, "b.go"), []byte(f20BContent), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.RemoveAll(filepath.Join(root, ".git")); err != nil {
		t.Fatal(err)
	}
	out2 := runJSON2(t, root, bin, "drift", "--format", "json")
	if len(out2) == 0 || out2[0].State != "UNKNOWN" {
		t.Fatalf("git failure not UNKNOWN: %+v", out2)
	}
}
