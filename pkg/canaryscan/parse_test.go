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

// TestCANARY_FIXWAVE1_ScanCarriesDeclaredPriority proves a token's PRIORITY
// declaration survives aggregation into the report. It is a pointer, so a
// token that declares nothing is distinguishable from one that declares 0:
// consumers (canary next's filesystem path) must be able to tell "the author
// asked for this first" from "the author said nothing".
func TestCANARY_FIXWAVE1_ScanCarriesDeclaredPriority(t *testing.T) {
	root := t.TempDir()
	files := map[string]string{
		"urgent.go":  "// CANARY: REQ=CBIN-001; FEATURE=\"Urgent\"; ASPECT=API; STATUS=STUB; PRIORITY=1; UPDATED=2026-08-30\n",
		"later.go":   "// CANARY: REQ=CBIN-002; FEATURE=\"Later\"; ASPECT=API; STATUS=STUB; PRIORITY=9; UPDATED=2026-08-30\n",
		"silent.go":  "// CANARY: REQ=CBIN-003; FEATURE=\"Silent\"; ASPECT=API; STATUS=STUB; UPDATED=2026-08-30\n",
		"zero.go":    "// CANARY: REQ=CBIN-004; FEATURE=\"Zero\"; ASPECT=API; STATUS=STUB; PRIORITY=0; UPDATED=2026-08-30\n",
		"garbage.go": "// CANARY: REQ=CBIN-005; FEATURE=\"Garbage\"; ASPECT=API; STATUS=STUB; PRIORITY=soon; UPDATED=2026-08-30\n",
	}
	for name, body := range files {
		if err := os.WriteFile(filepath.Join(root, name), []byte("package x\n"+body), 0o600); err != nil {
			t.Fatal(err)
		}
	}

	rep, err := Scan(root, DefaultSkipRegex(), nil, nil, nil)
	if err != nil {
		t.Fatal(err)
	}

	got := map[string]*int{}
	for _, r := range rep.Requirements {
		if len(r.Features) != 1 {
			t.Fatalf("%s: Features = %+v, want exactly one", r.ID, r.Features)
		}
		got[r.ID] = r.Features[0].Priority
	}

	for id, want := range map[string]*int{
		"CBIN-001": intPtr(1),
		"CBIN-002": intPtr(9),
		"CBIN-003": nil, // declared none
		"CBIN-004": intPtr(0),
		"CBIN-005": nil, // unparsable PRIORITY is ignored, as `canary index` ignores it
	} {
		p, ok := got[id]
		if !ok {
			t.Errorf("%s missing from report", id)
			continue
		}
		switch {
		case want == nil && p != nil:
			t.Errorf("%s: Priority = %d, want absent", id, *p)
		case want != nil && p == nil:
			t.Errorf("%s: Priority absent, want %d", id, *want)
		case want != nil && p != nil && *p != *want:
			t.Errorf("%s: Priority = %d, want %d", id, *p, *want)
		}
	}
}

// TestCANARY_FIXWAVE1_ScanPriorityKeepsMostUrgent proves that when several
// tokens fold into one aggregate feature and disagree about PRIORITY, the
// most urgent (lowest) declaration is the one the report carries.
func TestCANARY_FIXWAVE1_ScanPriorityKeepsMostUrgent(t *testing.T) {
	root := t.TempDir()
	a := "// CANARY: REQ=CBIN-010; FEATURE=\"Split\"; ASPECT=API; STATUS=STUB; PRIORITY=7; UPDATED=2026-08-30\n"
	b := "// CANARY: REQ=CBIN-010; FEATURE=\"Split\"; ASPECT=API; STATUS=STUB; PRIORITY=2; UPDATED=2026-08-30\n"
	if err := os.WriteFile(filepath.Join(root, "a.go"), []byte("package x\n"+a), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "b.go"), []byte("package x\n"+b), 0o600); err != nil {
		t.Fatal(err)
	}

	rep, err := Scan(root, DefaultSkipRegex(), nil, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(rep.Requirements) != 1 || len(rep.Requirements[0].Features) != 1 {
		t.Fatalf("report = %+v, want one requirement with one feature", rep.Requirements)
	}
	p := rep.Requirements[0].Features[0].Priority
	if p == nil || *p != 2 {
		t.Errorf("Priority = %v, want 2 (the most urgent declaration wins)", p)
	}
}

// intPtr is a local helper for the pointer-valued Priority assertions.
func intPtr(v int) *int { return &v }

// TestCANARY_FIXWAVE1_FeatureSortIsTupleOrdered proves features within a
// requirement are ordered by (Feature, Aspect) as a tuple rather than by the
// concatenation of the two, which lets a longer feature name sort ahead of a
// shorter one it starts with.
func TestCANARY_FIXWAVE1_FeatureSortIsTupleOrdered(t *testing.T) {
	root := t.TempDir()
	body := "package x\n" +
		"// CANARY: REQ=CBIN-020; FEATURE=\"A\"; ASPECT=Engine; STATUS=STUB; UPDATED=2026-08-30\n" +
		"// CANARY: REQ=CBIN-020; FEATURE=\"AB\"; ASPECT=API; STATUS=STUB; UPDATED=2026-08-30\n"
	if err := os.WriteFile(filepath.Join(root, "x.go"), []byte(body), 0o600); err != nil {
		t.Fatal(err)
	}

	rep, err := Scan(root, DefaultSkipRegex(), nil, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(rep.Requirements) != 1 {
		t.Fatalf("Requirements = %+v, want one", rep.Requirements)
	}
	feats := rep.Requirements[0].Features
	if len(feats) != 2 {
		t.Fatalf("Features = %+v, want two", feats)
	}
	// "A" < "AB" as names; concatenation would rank "ABAPI" before "AEngine".
	if feats[0].Feature != "A" || feats[1].Feature != "AB" {
		t.Errorf("feature order = [%s/%s, %s/%s], want A/Engine before AB/API",
			feats[0].Feature, feats[0].Aspect, feats[1].Feature, feats[1].Aspect)
	}
}
