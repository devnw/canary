// Copyright (c) 2025 by Developer Network.
//
// For more details, see the LICENSE file in the root directory of this
// source code repository or contact Developer Network at info@devnw.com.

package audit

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

// TestAuditF06 pins F-06: a read-only command must never create the token
// database. The CLI used to run storage.AutoMigrate from the root command's
// PersistentPreRunE, so *every* invocation -- `canary list` on a repo with no
// index, `canary scan`, even `canary --help` -- materialised
// .canary/canary.db and printed a creation banner on stdout.
func TestAuditF06(t *testing.T) {
	bin := buildCanary(t)
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "a.go"), []byte(
		"// CANARY: REQ=CBIN-001; FEATURE=\"F\"; ASPECT=API; STATUS=STUB; UPDATED=2026-01-01\n"), 0o600); err != nil {
		t.Fatalf("write a.go: %v", err)
	}

	for _, args := range [][]string{{"list"}, {"scan", "--root", "."}, {"show", "CBIN-001"}} {
		cmd := exec.Command(bin, args...) //nolint:gosec // test-built binary
		cmd.Dir = root
		home := t.TempDir()
		cmd.Env = append(os.Environ(), "HOME="+home, "USERPROFILE="+home)
		_ = cmd.Run() // exit code irrelevant here
		if _, err := os.Stat(filepath.Join(root, ".canary", "canary.db")); err == nil {
			t.Fatalf("%v created a database", args)
		}
	}
}
