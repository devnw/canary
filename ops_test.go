// Copyright (c) 2025 by Developer Network.
//
// For more details, see the LICENSE file in the root directory of this
// source code repository or contact Developer Network at info@devnw.com.

package canary

import (
	"path/filepath"
	"strings"
	"testing"

	"go.devnw.com/canary/internal/storage"
)

// CANARY: REQ=CBIN-212; FEATURE="OpsExtraction"; ASPECT=API; STATUS=TESTED; TEST=TestOpsExtraction_GrepAndGrouping; UPDATED=2025-11-02
func TestOpsExtraction_GrepAndGrouping(t *testing.T) {
	tmp := t.TempDir()
	dbPath := filepath.Join(tmp, "test.db")
	if err := storage.MigrateDB(dbPath, "all"); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	db, err := storage.Open(dbPath)
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	defer db.Close()

	toks := []*storage.Token{
		{ReqID: "CBIN-900", Feature: "AlphaFeature", Aspect: "API", Status: "IMPL", FilePath: "internal/a/alpha.go"},
		{ReqID: "CBIN-901", Feature: "BetaFeature", Aspect: "Engine", Status: "TESTED", FilePath: "internal/b/beta.go", Test: "TestBeta"},
		{ReqID: "CBIN-900", Feature: "AlphaCLI", Aspect: "CLI", Status: "IMPL", FilePath: "cmd/alpha/cli.go"},
	}
	for _, tk := range toks {
		if err := db.UpsertToken(tk); err != nil {
			t.Fatalf("upsert: %v", err)
		}
	}

	res, err := GrepTokens(db, "Alpha")
	if err != nil {
		t.Fatalf("grep: %v", err)
	}
	if len(res) != 2 {
		t.Fatalf("expected 2 alpha matches got %d", len(res))
	}

	out := FormatGrepResults(res)
	if !strings.Contains(out, "AlphaFeature") || !strings.Contains(out, "AlphaCLI") {
		t.Fatalf("missing features in output: %s", out)
	}

	grouped := FormatGrepResultsByRequirement(res)
	if !strings.Contains(grouped, "CBIN-900 (2 tokens)") {
		t.Fatalf("group output missing requirement summary: %s", grouped)
	}

	// Token table
	table := FormatTokensTable(res, "aspect")
	if !strings.Contains(table, "## API") || !strings.Contains(table, "## CLI") {
		t.Fatalf("table missing aspect groups: %s", table)
	}
}

// CANARY: REQ=CBIN-212; FEATURE="OpsExtraction"; ASPECT=API; STATUS=TESTED; TEST=TestOpsExtraction_FilesFormat; UPDATED=2025-11-02
func TestOpsExtraction_FilesFormat(t *testing.T) {
	fileGroups := map[string][]*storage.Token{
		"a.go": {{ReqID: "R1", Feature: "F1", Aspect: "API", Status: "IMPL"}},
		"b.go": {{ReqID: "R1", Feature: "F2", Aspect: "CLI", Status: "IMPL"}, {ReqID: "R1", Feature: "F3", Aspect: "CLI", Status: "IMPL"}},
	}
	out := FormatFilesList(fileGroups)
	if !strings.Contains(out, "Total: 2 files, 3 tokens") {
		t.Fatalf("summary mismatch: %s", out)
	}
	if !strings.Contains(out, "**API:**") || !strings.Contains(out, "**CLI:**") {
		t.Fatalf("missing aspect headings: %s", out)
	}
}
