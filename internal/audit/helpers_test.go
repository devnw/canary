// Copyright (c) 2025 by Developer Network.
//
// For more details, see the LICENSE file in the root directory of this
// source code repository or contact Developer Network at info@devnw.com.

package audit

import (
	"crypto/sha256"
	"encoding/hex"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"devnw.dev/canary/pkg/canaryscan"
	"devnw.dev/canary/pkg/sources"
)

// buildCanary builds ./cmd/canary exactly once per test binary run and
// returns the path to the built executable. The binary is shared by every
// CLI-level audit test in this package: building it per-test would dominate
// the suite's runtime for no added coverage.
var (
	buildOnce sync.Once
	canaryBin string
	buildErr  error
)

func buildCanary(t *testing.T) string {
	t.Helper()
	buildOnce.Do(func() {
		dir, err := os.MkdirTemp("", "canarybin")
		if err != nil {
			buildErr = err
			return
		}
		bin := filepath.Join(dir, "canary")
		cmd := exec.Command("go", "build", "-o", bin, "devnw.dev/canary/cmd/canary")
		cmd.Dir = repoRoot()
		if out, berr := cmd.CombinedOutput(); berr != nil {
			buildErr = &buildFailure{out: string(out), err: berr}
			return
		}
		canaryBin = bin
	})
	if buildErr != nil {
		t.Fatalf("build canary: %v", buildErr)
	}
	return canaryBin
}

// buildFailure carries the compiler output alongside the exec error so a
// build break is diagnosable from the test log alone.
type buildFailure struct {
	out string
	err error
}

func (b *buildFailure) Error() string { return b.err.Error() + "\n" + b.out }

// repoRoot walks up from the test's working directory to the directory
// holding go.mod -- the module root the CLI is built from.
func repoRoot() string {
	dir, err := os.Getwd()
	if err != nil {
		return "."
	}
	for {
		if _, serr := os.Stat(filepath.Join(dir, "go.mod")); serr == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "."
		}
		dir = parent
	}
}

// runCLI runs the canary binary with root as its working directory and
// returns its combined output. HOME is redirected at a throwaway directory:
// `canary init` installs agent files into the user's home by default, and a
// test must never reach the developer's real ~/.claude or ~/.codex.
func runCLI(t *testing.T, root, bin string, args ...string) (string, error) {
	t.Helper()
	cmd := exec.Command(bin, args...)
	cmd.Dir = root
	home := t.TempDir()
	cmd.Env = append(os.Environ(), "HOME="+home, "USERPROFILE="+home)
	out, err := cmd.CombinedOutput()
	return string(out), err
}

// run executes the canary binary in root and fails the test if it exits
// non-zero.
func run(t *testing.T, root, bin string, args ...string) string {
	t.Helper()
	out, err := runCLI(t, root, bin, args...)
	if err != nil {
		t.Fatalf("canary %s: %v\n%s", strings.Join(args, " "), err, out)
	}
	return out
}

// runExpectFail executes the canary binary in root and fails the test if it
// exits zero.
func runExpectFail(t *testing.T, root, bin string, args ...string) string {
	t.Helper()
	out, err := runCLI(t, root, bin, args...)
	if err == nil {
		t.Fatalf("canary %s: expected failure, got success\n%s", strings.Join(args, " "), out)
	}
	return out
}

// sha256File returns path's content digest, so a test can assert a file was
// left byte-for-byte alone without holding its bytes.
func sha256File(t *testing.T, path string) string {
	t.Helper()
	b, err := os.ReadFile(path) //nolint:gosec // test-controlled path
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	sum := sha256.Sum256(b)
	return hex.EncodeToString(sum[:])
}

// scanDir runs a full canaryscan over root with that root's own registry and
// ignore rules -- the same inputs the CLI uses.
func scanDir(t *testing.T, root string) canaryscan.Report {
	t.Helper()
	reg, err := sources.LoadFromRoot(root)
	if err != nil {
		t.Fatalf("load sources: %v", err)
	}
	ignorePatterns, err := canaryscan.LoadCanaryIgnore(root)
	if err != nil {
		t.Fatalf("load .canaryignore: %v", err)
	}
	rep, err := canaryscan.Scan(root, canaryscan.DefaultSkipRegex(), nil, ignorePatterns, reg)
	if err != nil {
		t.Fatalf("scan %s: %v", root, err)
	}
	return rep
}

// git runs one git command in root, failing the test on error. Config that
// a developer machine may set globally (signing, hooks, default branch) is
// pinned per-invocation so these fixtures behave identically everywhere.
func git(t *testing.T, root string, args ...string) string {
	t.Helper()
	full := append([]string{
		"-C", root,
		"-c", "user.email=audit@example.com",
		"-c", "user.name=Audit",
		"-c", "commit.gpgsign=false",
		"-c", "init.defaultBranch=main",
	}, args...)
	cmd := exec.Command("git", full...)
	cmd.Env = append(os.Environ(), "GIT_CONFIG_NOSYSTEM=1")
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %s: %v\n%s", strings.Join(args, " "), err, out)
	}
	return strings.TrimSpace(string(out))
}

// initGitRepo turns root into a git repository with one (empty) commit, so
// `git rev-parse HEAD` resolves -- the commit every evidence record binds to.
func initGitRepo(t *testing.T, root string) {
	t.Helper()
	git(t, root, "init", "-q")
	git(t, root, "commit", "--allow-empty", "-q", "-m", "init")
}

// gitCommitAll stages and commits everything currently in root.
func gitCommitAll(t *testing.T, root string) {
	t.Helper()
	git(t, root, "add", "-A")
	git(t, root, "commit", "-q", "-m", "fixture")
}

// headCommit returns root's current HEAD SHA.
func headCommit(t *testing.T, root string) string {
	t.Helper()
	return git(t, root, "rev-parse", "HEAD")
}
