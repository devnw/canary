// Copyright (c) 2025 by Developer Network.
//
// For more details, see the LICENSE file in the root directory of this
// source code repository or contact Developer Network at info@devnw.com.

package audit

import (
	"errors"
	"os"
	"os/exec"
	"testing"
)

// TestAuditF07 pins F-07: `--order-by` was concatenated straight into the
// ORDER BY clause of the tokens query, so any caller could append arbitrary
// SQL. The flag now names one of a fixed set of order keys; anything else is
// refused with a machine-readable contract on stdout and exit 2, and no query
// ever reaches SQLite.
func TestAuditF07(t *testing.T) {
	bin := buildCanary(t)
	root := initIndexedRepo(t, bin)

	cmd := exec.Command(bin, "list", "--order-by", "updated_at; DROP TABLE tokens") //nolint:gosec // test-built binary
	cmd.Dir = root
	home := t.TempDir()
	cmd.Env = append(os.Environ(), "HOME="+home, "USERPROFILE="+home)
	out, err := cmd.Output()

	var ee *exec.ExitError
	if !errors.As(err, &ee) || ee.ExitCode() != 2 {
		t.Fatalf("want exit 2: %v (stdout %q)", err, out)
	}
	want := `{"ok":false,"code":"INVALID_ORDER_BY","message":"allowed values: priority_desc,req_asc,updated_desc"}` + "\n"
	if string(out) != want {
		t.Fatalf("got %q, want %q", out, want)
	}
}

// TestAuditF07AllowedKeys proves the allowlisted keys still work, so the fix
// is a narrowing rather than a removal of ordering.
func TestAuditF07AllowedKeys(t *testing.T) {
	bin := buildCanary(t)
	root := initIndexedRepo(t, bin)

	for _, key := range []string{"updated_desc", "req_asc", "priority_desc"} {
		out := run(t, root, bin, "list", "--order-by", key)
		if out == "" {
			t.Fatalf("--order-by %s produced no output", key)
		}
	}
}
