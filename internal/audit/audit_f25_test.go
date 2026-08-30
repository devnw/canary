// Copyright (c) 2025 by Developer Network.
//
// For more details, see the LICENSE file in the root directory of this
// source code repository or contact Developer Network at info@devnw.com.

package audit

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// repoPath returns the absolute path to name relative to the module root, so
// a static-file audit can read a checked-in file (Makefile, .gitlab-ci.yml)
// regardless of the test's working directory.
func repoPath(t *testing.T, name string) string {
	t.Helper()
	return filepath.Join(repoRoot(), name)
}

// TestAuditF25 covers F-25: the Makefile must carry the real target set, be
// fail-closed (no `|| true` masking), carry none of the foreign pcap project's
// targets, and pin every downloaded tool by version. A green `make lint` that
// swallows failures, or a `bench:` body that runs the bare word `bench`,
// proves nothing -- so this asserts the static shape of a correct build file.
func TestAuditF25(t *testing.T) {
	b, err := os.ReadFile(repoPath(t, "Makefile"))
	if err != nil {
		t.Fatalf("read Makefile: %v", err)
	}
	s := string(b)
	for _, tgt := range []string{
		"build:", "test:", "test-race:", "lint:", "security:",
		"verify:", "bench:", "fuzz:", "clean:",
	} {
		if !strings.Contains(s, "\n"+tgt) && !strings.HasPrefix(s, tgt) {
			t.Fatalf("missing target %s", tgt)
		}
	}
	if strings.Contains(s, "|| true") {
		t.Fatal("fail-open construct in Makefile")
	}
	if strings.Contains(s, "pcap") {
		t.Fatal("foreign project targets still present")
	}
	for _, pin := range []string{
		"GOLANGCI_LINT_VERSION :=", "GOSEC_VERSION :=", "GOVULNCHECK_VERSION :=",
	} {
		if !strings.Contains(s, pin) {
			t.Fatalf("tool not pinned: %s", pin)
		}
	}
}
