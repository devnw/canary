package canaryscan

import (
	"os"
	"path/filepath"
	"testing"
)

// content is built as one physical source line (embedded \n escapes) so this
// test fixture never contains a line that literally starts with "CANARY:" —
// the repo's own canary scan must not see these as real tokens.
func migrateFixtureContent() string {
	return "// CANARY:MIGRATE auth flows move to pkg/auth; see CBIN-105\n# CANARY:MIGRATE python-style note\n<!-- CANARY:MIGRATE md note -->\n-- CANARY:MIGRATE sql note\n// CANARY: REQ=CBIN-105; FEATURE=\"X\"; ASPECT=API; STATUS=IMPL; UPDATED=2026-08-29\n"
}

func TestCANARY_CBIN_301_ExtractMigrateNotes(t *testing.T) {
	reg := ticketRegistry(t)
	notes := ExtractMigrateNotes("fixture.go", migrateFixtureContent(), reg)
	if len(notes) != 4 {
		t.Fatalf("got %d notes, want 4: %+v", len(notes), notes)
	}
	wantLines := []int{1, 2, 3, 4}
	for i, n := range notes {
		if n.Line != wantLines[i] {
			t.Errorf("note %d: line = %d, want %d", i, n.Line, wantLines[i])
		}
		if n.File != "fixture.go" {
			t.Errorf("note %d: file = %q, want fixture.go", i, n.File)
		}
	}
	if notes[2].Text != "md note" {
		t.Errorf("md note text = %q, want trailing --> stripped to %q", notes[2].Text, "md note")
	}
	if len(notes[0].ReqIDs) != 1 || notes[0].ReqIDs[0] != "CBIN-105" {
		t.Errorf("notes[0].ReqIDs = %v, want [CBIN-105]", notes[0].ReqIDs)
	}
}

func TestCANARY_CBIN_301_MigrateLineDoesNotAbortScan(t *testing.T) {
	dir := t.TempDir()
	tokenFile := filepath.Join(dir, "token.go")
	tokenContent := "package x\n// CANARY: REQ=CBIN-106; FEATURE=\"Thing\"; ASPECT=API; STATUS=IMPL; UPDATED=2026-08-29\n"
	if err := os.WriteFile(tokenFile, []byte(tokenContent), 0o600); err != nil {
		t.Fatal(err)
	}
	migrateFile := filepath.Join(dir, "migrate.go")
	migrateContent := "package x\n// CANARY:MIGRATE free text with = signs; and semicolons\n"
	if err := os.WriteFile(migrateFile, []byte(migrateContent), 0o600); err != nil {
		t.Fatal(err)
	}
	rep, err := Scan(dir, nil, nil, nil, nil)
	if err != nil {
		t.Fatalf("Scan returned error, MIGRATE line must never abort a scan: %v", err)
	}
	if len(rep.Requirements) != 1 || rep.Requirements[0].ID != "CBIN-106" {
		t.Fatalf("expected the valid token to still be scanned, got %+v", rep.Requirements)
	}
	if len(rep.MigrationNotes) != 1 {
		t.Fatalf("MigrationNotes len = %d, want 1: %+v", len(rep.MigrationNotes), rep.MigrationNotes)
	}
}
