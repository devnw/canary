// Copyright (c) 2025 by Developer Network.
//
// For more details, see the LICENSE file in the root directory of this
// source code repository or contact Developer Network at info@devnw.com.

package audit

import (
	"path/filepath"
	"strings"
	"testing"

	"devnw.dev/canary/pkg/storage"
)

// TestStaleSchemaRefusesRead pins the CLI symptom of the missing schema-
// version check: against a pre-v7 index, `canary list` reached the query
// layer and failed with a raw "no such column: content_hash", which names
// neither the problem nor the fix. The read must now refuse with the one
// sentence that does, and must not migrate on the way -- a read path that
// rewrote the schema is the very side effect the read-only open exists to
// prevent.
func TestStaleSchemaRefusesRead(t *testing.T) {
	bin := buildCanary(t)
	root := t.TempDir()
	dbPath := filepath.Join(root, ".canary", "canary.db")

	if err := storage.MigrateDB(dbPath, "6"); err != nil {
		t.Fatalf("migrate to version 6: %v", err)
	}

	out, err := runCLI(t, root, bin, "list")
	if err == nil {
		t.Fatalf("list against a stale index succeeded:\n%s", out)
	}
	if !strings.Contains(out, "index schema is out of date; run 'canary index'") {
		t.Fatalf("stale index did not name the fix; output:\n%s", out)
	}
	if strings.Contains(out, "no such column") {
		t.Fatalf("stale index leaked a raw SQLite error:\n%s", out)
	}

	stale, version, nerr := storage.NeedsMigration(dbPath)
	if nerr != nil {
		t.Fatalf("NeedsMigration: %v", nerr)
	}
	if !stale || version != 6 {
		t.Fatalf("a read migrated the index: stale=%v version=%d", stale, version)
	}
}
