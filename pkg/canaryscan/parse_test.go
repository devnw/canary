package canaryscan

import (
	"os"
	"path/filepath"
	"testing"
)

// CANARY: REQ=CP-285; FEATURE="SQLCommentTokens"; ASPECT=Engine; STATUS=TESTED; TEST=TestCANARY_CP_285_SQLCommentTokenLineRegexMatches,TestCANARY_CP_285_SQLCommentTokenScanned; UPDATED=2026-08-29

// TestCANARY_CP_285_SQLCommentTokenLineRegexMatches is a narrow regression
// test on tokenLineRe itself: a "--" SQL line-comment prefix must be
// recognized as a valid CANARY token marker, same as "//", "#", and "/*".
func TestCANARY_CP_285_SQLCommentTokenLineRegexMatches(t *testing.T) {
	line := `-- CANARY: REQ=CP-820; FEATURE="SQLMigration"; ASPECT=Storage; STATUS=IMPL; UPDATED=2026-08-29`
	m := tokenLineRe.FindStringSubmatch(line)
	if m == nil {
		t.Fatalf("tokenLineRe did not match SQL comment token line: %q", line)
	}
	fields, err := parseKV(m[1], nil)
	if err != nil {
		t.Fatalf("parseKV: %v", err)
	}
	if fields["REQ"] != "CP-820" {
		t.Errorf("REQ = %q, want CP-820", fields["REQ"])
	}
}

// TestCANARY_CP_285_SQLCommentTokenScanned proves a CANARY token written as
// a SQL line comment ("-- CANARY: ...") in a .sql file is picked up by a
// full Scan. Before this fix, the tokenLineRe prefix group only recognized
// "//", "#", and "/*", so tokens in .sql migration files were invisible to
// the scanner.
func TestCANARY_CP_285_SQLCommentTokenScanned(t *testing.T) {
	root := t.TempDir()
	sql := "-- CANARY: REQ=CP-821; FEATURE=\"Migration\"; ASPECT=Storage; STATUS=IMPL; UPDATED=2026-08-29\n" +
		"CREATE TABLE widgets (id INTEGER PRIMARY KEY);\n"
	if err := os.WriteFile(filepath.Join(root, "001_widgets.sql"), []byte(sql), 0o600); err != nil {
		t.Fatal(err)
	}

	rep, err := Scan(root, DefaultSkipRegex(), nil, nil, nil)
	if err != nil {
		t.Fatal(err)
	}

	found := false
	for _, r := range rep.Requirements {
		if r.ID == "CP-821" {
			found = true
		}
	}
	if !found {
		t.Error("CANARY token in .sql file's -- comment was not scanned")
	}
}
