// Copyright (c) 2025 by Developer Network.
//
// For more details, see the LICENSE file in the root directory of this
// source code repository or contact Developer Network at info@devnw.com.

package audit

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"testing"
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
