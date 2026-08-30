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

// TestCANARY_FIXWAVE1_FeatureDoubleQuotedValuePreserved proves a FEATURE
// value that legitimately starts and ends with literal double-quote
// characters (e.g. FEATURE="\"quoted\""  ->  decoded value `"quoted"`)
// round-trips through Scan intact. decodeValue (serialize.go) already strips
// the wire-form quoting and resolves the `\"` escapes; unquote must not strip
// the resulting double quotes a second time.
func TestCANARY_FIXWAVE1_FeatureDoubleQuotedValuePreserved(t *testing.T) {
	root := t.TempDir()
	line := `// CANARY: REQ=CBIN-001; FEATURE="\"quoted\""; ASPECT=Engine; STATUS=IMPL; UPDATED=2026-08-30` + "\n"
	if err := os.WriteFile(filepath.Join(root, "x.go"), []byte("package x\n"+line), 0o600); err != nil {
		t.Fatal(err)
	}

	rep, err := Scan(root, DefaultSkipRegex(), nil, nil, nil)
	if err != nil {
		t.Fatal(err)
	}

	var got *Requirement
	for i := range rep.Requirements {
		if rep.Requirements[i].ID == "CBIN-001" {
			got = &rep.Requirements[i]
		}
	}
	if got == nil {
		t.Fatal("CBIN-001 not in report")
	}
	if len(got.Features) != 1 {
		t.Fatalf("Features = %+v, want exactly one", got.Features)
	}
	if want := `"quoted"`; got.Features[0].Feature != want {
		t.Errorf("Feature = %q, want %q (double-quote decode must not be applied twice)", got.Features[0].Feature, want)
	}
}

// TestCANARY_FIXWAVE1_UnquoteOnlyStripsSingleQuotes proves unquote (the
// legacy single-quote helper) leaves a value with literal double quotes at
// both ends untouched, since double-quote decoding already happened upstream
// in decodeValue. It still strips a legacy single-quoted pair.
func TestCANARY_FIXWAVE1_UnquoteOnlyStripsSingleQuotes(t *testing.T) {
	cases := []struct{ in, want string }{
		{in: `"quoted"`, want: `"quoted"`}, // already-decoded double quotes: not unquote's job
		{in: `'quoted'`, want: `quoted`},   // legacy single-quote form: still stripped
		{in: `plain`, want: `plain`},
	}
	for _, tc := range cases {
		if got := unquote(tc.in); got != tc.want {
			t.Errorf("unquote(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}
