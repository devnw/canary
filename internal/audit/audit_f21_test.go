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

// TestAuditF21 covers F-21: `canary upgrade --write` rewrites token lines with
// its own regexes, so a rule bug could silently delete a requirement. Every
// written file is now checked token-for-token against its pre-rewrite state,
// with each rule's declared field edits normalized away -- the padded ID here
// is an intended change, the token itself must survive.
func TestAuditF21(t *testing.T) {
	bin := buildCanary(t)
	root := t.TempDir()

	tok := "// CANARY: REQ=CBIN-7; FEATURE=\"F\"; ASPECT=API; STATUS=IMPL; UPDATED=2026-01-01\n"
	if err := os.WriteFile(filepath.Join(root, "a.go"), []byte(tok), 0o644); err != nil {
		t.Fatalf("seed a.go: %v", err)
	}

	run(t, root, bin, "upgrade", "--write", "--root", root)

	rep := scanDir(t, root)
	if len(rep.Requirements) != 1 || rep.Requirements[0].ID != "CBIN-007" {
		t.Fatalf("token lost or not normalized: %+v", rep.Requirements)
	}
}

// TestAuditF21PreservesUntouchedFields proves a rewrite carries through the
// fields the selected rule has no business changing: only STATUS moves, and
// the OWNER the rule may not touch is still there afterwards.
func TestAuditF21PreservesUntouchedFields(t *testing.T) {
	bin := buildCanary(t)
	root := t.TempDir()

	// status-fixed alone may change STATUS and nothing else. The extra
	// OWNER field is outside every enabled rule's remit, so it pins the
	// token's identity across the rewrite.
	src := "// CANARY: REQ=CBIN-8; FEATURE=\"G\"; ASPECT=API; STATUS=FIXED; OWNER=team; UPDATED=2026-01-01\n"
	path := filepath.Join(root, "b.go")
	if err := os.WriteFile(path, []byte(src), 0o644); err != nil {
		t.Fatalf("seed b.go: %v", err)
	}

	run(t, root, bin, "upgrade", "--write", "--rule", "status-fixed", "--root", root)

	got, err := os.ReadFile(path) //nolint:gosec // test-controlled path
	if err != nil {
		t.Fatalf("read b.go: %v", err)
	}
	if !strings.Contains(string(got), "STATUS=REMOVED") {
		t.Fatalf("status-fixed rule did not apply: %s", got)
	}
	if !strings.Contains(string(got), "OWNER=team") {
		t.Fatalf("rewrite dropped a field the rule may not touch: %s", got)
	}
}
