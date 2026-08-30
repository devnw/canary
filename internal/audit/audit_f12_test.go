// Copyright (c) 2025 by Developer Network.
//
// For more details, see the LICENSE file in the root directory of this
// source code repository or contact Developer Network at info@devnw.com.

package audit

import (
	"os"
	"path/filepath"
	"testing"
)

// TestAuditF12 pins F-12: `canary index` truncated the token table and then
// inserted row by row with no transaction, warning-and-continuing on every
// failure. A run that died partway -- or that hit a token the parser rejects
// -- left the index emptied or half-written with no way to tell. The whole
// run is now one transaction: either the new index replaces the old one, or
// the old one survives byte for byte.
//
// What this case proves specifically is PRE-FLIGHT REFUSAL: the injected bad
// token is rejected by the scan before OpenRW is ever reached, so the
// database is never opened for writing and the digest comparison below is
// satisfied trivially. That is worth pinning on its own -- a run that
// refuses must not so much as touch the file -- but it is not evidence about
// the transaction.
//
// The rollback evidence lives at the storage layer, in
// pkg/storage.TestReplaceIndexRollsBack: it drives ReplaceIndex to a failure
// raised *inside* the transaction (the post-insert row-count check, which
// runs after the token inserts, the ref rewrite, and the metadata write) and
// proves the previous index survives intact and the half-written rows are
// gone. Duplicating that at the CLI level would need an injected fault the
// binary has no seam for, and would prove nothing the storage test does not.
func TestAuditF12(t *testing.T) {
	bin := buildCanary(t)
	root := initIndexedRepo(t, bin)
	dbPath := dbPathIn(root)
	before := sha256File(t, dbPath)

	// A token whose ASPECT is not in the canonical set is a parse error the
	// scanner reports as an issue; index must refuse the whole run.
	if err := os.WriteFile(filepath.Join(root, "bad.go"), []byte(
		"// CANARY: REQ=CBIN-009; FEATURE=\"X\"; ASPECT=NOPE; STATUS=STUB; UPDATED=2026-01-01\n"), 0o600); err != nil {
		t.Fatalf("write bad.go: %v", err)
	}

	out := runExpectFail(t, root, bin, "index", "--root", ".")
	if out == "" {
		t.Fatal("failed index said nothing about why")
	}

	if got := sha256File(t, dbPath); got != before {
		t.Fatal("failed index mutated the database")
	}
}

// TestAuditF12CleanRunCommits is the positive half: a run with no scan issues
// must actually replace the index and record its metadata.
func TestAuditF12CleanRunCommits(t *testing.T) {
	bin := buildCanary(t)
	root := initIndexedRepo(t, bin)

	if err := os.WriteFile(filepath.Join(root, "b.go"), []byte(
		"// CANARY: REQ=CBIN-002; FEATURE=\"G\"; ASPECT=CLI; STATUS=IMPL; UPDATED=2026-01-02\n"), 0o600); err != nil {
		t.Fatalf("write b.go: %v", err)
	}

	run(t, root, bin, "index", "--root", ".")

	db := openDBAt(t, dbPathIn(root))
	defer func() { _ = db.Close() }()

	rows, err := db.GetTokensByReqID("", "CBIN-002")
	if err != nil {
		t.Fatalf("lookup CBIN-002: %v", err)
	}
	if len(rows) != 1 {
		t.Fatalf("want 1 row for CBIN-002, got %d", len(rows))
	}

	meta, err := db.GetIndexMeta()
	if err != nil {
		t.Fatalf("GetIndexMeta: %v", err)
	}
	if meta == nil {
		t.Fatal("index recorded no metadata")
	}
	if meta.CommitSHA == "" {
		t.Fatal("index metadata has no commit sha in a git repository")
	}
	if meta.ScanDigest == "" {
		t.Fatal("index metadata has no scan digest")
	}
}
