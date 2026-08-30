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

	"devnw.dev/canary/pkg/storage"
)

// TestAuditF08 pins F-08 at the storage layer: every token query used to be
// unscoped, so two projects sharing one database (or one requirement ID)
// silently answered each other's questions. Queries are now project-scoped,
// and an ambiguous unscoped lookup is refused rather than guessed.
func TestAuditF08(t *testing.T) {
	db := openTestDB(t)
	seedToken(t, db, "projA", "CBIN-001")
	seedToken(t, db, "projB", "CBIN-001")

	if _, err := db.GetTokensByReqID("", "CBIN-001"); !errors.Is(err, storage.ErrProjectRequired) {
		t.Fatalf("want ErrProjectRequired, got %v", err)
	}

	rows, err := db.GetTokensByReqID("projA", "CBIN-001")
	if err != nil || len(rows) != 1 || rows[0].ProjectID != "projA" {
		t.Fatalf("scoped query leaked foreign rows: %v %+v", err, rows)
	}
}

// TestAuditF08SingleProjectUnscoped proves the unscoped path still answers
// when there is nothing to be ambiguous about -- the fix must not force
// --project on every single-project repository.
func TestAuditF08SingleProjectUnscoped(t *testing.T) {
	db := openTestDB(t)
	seedToken(t, db, "default", "CBIN-001")

	rows, err := db.GetTokensByReqID("", "CBIN-001")
	if err != nil {
		t.Fatalf("unscoped lookup in a single-project db failed: %v", err)
	}
	if len(rows) != 1 || rows[0].ProjectID != "default" {
		t.Fatalf("unexpected rows: %+v", rows)
	}
}

// TestAuditF08CLI is the command-level half: `canary show` against a database
// holding the same requirement ID under two projects must refuse with the
// PROJECT_REQUIRED contract on stdout and exit 2, and must succeed once
// --project disambiguates.
func TestAuditF08CLI(t *testing.T) {
	bin := buildCanary(t)
	root := t.TempDir()
	db := openDBAt(t, dbPathIn(root))
	seedToken(t, db, "projA", "CBIN-001")
	seedToken(t, db, "projB", "CBIN-001")
	if err := db.Close(); err != nil {
		t.Fatalf("close db: %v", err)
	}

	cmd := exec.Command(bin, "show", "CBIN-001") //nolint:gosec // test-built binary
	cmd.Dir = root
	home := t.TempDir()
	cmd.Env = append(os.Environ(), "HOME="+home, "USERPROFILE="+home)
	out, err := cmd.Output()

	var ee *exec.ExitError
	if !errors.As(err, &ee) || ee.ExitCode() != 2 {
		t.Fatalf("want exit 2: %v (stdout %q)", err, out)
	}
	want := `{"ok":false,"code":"PROJECT_REQUIRED","message":"duplicate requirement id across projects; pass --project"}` + "\n"
	if string(out) != want {
		t.Fatalf("got %q, want %q", out, want)
	}

	scoped := run(t, root, bin, "show", "CBIN-001", "--project", "projA")
	if scoped == "" {
		t.Fatal("--project produced no output")
	}
}
