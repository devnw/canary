// Copyright (c) 2025 by Developer Network.
//
// For more details, see the LICENSE file in the root directory of this
// source code repository or contact Developer Network at info@devnw.com.

package upgrade_test

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"devnw.dev/canary/pkg/sources"
	"devnw.dev/canary/pkg/upgrade"
)

func writeFile(t *testing.T, dir, name, content string) string {
	t.Helper()
	path := filepath.Join(dir, name)
	if err := os.MkdirAll(filepath.Dir(path), 0o750); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(path, []byte(content), 0o640); err != nil {
		t.Fatalf("write %s: %v", name, err)
	}
	return path
}

func readFile(t *testing.T, path string) string {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	return string(b)
}

// TestCANARY_CBIN_302_Rules table-drives one fixture per named rule, run in
// isolation (Options.Rules restricted to that one rule), verifying both the
// returned Change and the rewritten file content.
// CANARY: REQ=CBIN-302; FEATURE="TokenUpgrade"; ASPECT=Engine; STATUS=TESTED; TEST=TestCANARY_CBIN_302_Rules; UPDATED=2026-08-29
func TestCANARY_CBIN_302_Rules(t *testing.T) {
	cases := []struct {
		name    string
		file    string
		content string
		rule    string
		opts    func(o *upgrade.Options)
		wantNew string
		wantOld string
	}{
		{
			name: "join-multiline",
			file: "bug.go",
			content: "// CANARY: BUG=BUG-API-001; TITLE=\"Timeout\";\n" +
				"//         ASPECT=API; STATUS=OPEN;\n" +
				"//         SEVERITY=S2; PRIORITY=P1;\n" +
				"//         UPDATED=2025-01-01\n",
			rule:    "join-multiline",
			wantNew: `// CANARY: BUG=BUG-API-001; TITLE="Timeout"; ASPECT=API; STATUS=OPEN; SEVERITY=S2; PRIORITY=P1; UPDATED=2025-01-01`,
		},
		{
			name:    "md-heading",
			file:    "doc.md",
			content: `# CANARY: REQ=CBIN-101; FEATURE="X"; ASPECT=Docs; STATUS=IMPL; UPDATED=2025-01-01` + "\n",
			rule:    "md-heading",
			wantNew: `<!-- CANARY: REQ=CBIN-101; FEATURE="X"; ASPECT=Docs; STATUS=IMPL; UPDATED=2025-01-01 -->`,
		},
		{
			name:    "unicode-hyphen",
			file:    "u.go",
			content: "// CANARY: REQ=CBIN‑101; FEATURE=\"X\"; ASPECT=API; STATUS=IMPL; UPDATED=2025-01-01\n",
			rule:    "unicode-hyphen",
			wantNew: `// CANARY: REQ=CBIN-101; FEATURE="X"; ASPECT=API; STATUS=IMPL; UPDATED=2025-01-01`,
		},
		{
			name:    "bare-id",
			file:    "b.go",
			content: `// CANARY: REQ-42; FEATURE="X"; ASPECT=API; STATUS=IMPL; UPDATED=2025-01-01` + "\n",
			rule:    "bare-id",
			wantNew: `// CANARY: REQ=REQ-042; FEATURE="X"; ASPECT=API; STATUS=IMPL; UPDATED=2025-01-01`,
		},
		{
			name:    "bug-alias",
			file:    "ba.go",
			content: `// CANARY: BUG=BUG-API-001; TITLE="Timeout"; ASPECT=API; STATUS=OPEN; UPDATED=2025-01-01` + "\n",
			rule:    "bug-alias",
			wantNew: `// CANARY: BUG=BUG-API-001; TITLE="Timeout"; FEATURE="Timeout"; ASPECT=API; STATUS=OPEN; UPDATED=2025-01-01`,
		},
		{
			name:    "status-fixed",
			file:    "s.go",
			content: `// CANARY: REQ=CBIN-101; FEATURE="X"; ASPECT=API; STATUS=FIXED; UPDATED=2025-01-01` + "\n",
			rule:    "status-fixed",
			wantNew: `// CANARY: REQ=CBIN-101; FEATURE="X"; ASPECT=API; STATUS=REMOVED; UPDATED=2025-01-01`,
		},
		{
			name:    "pad-flatfile",
			file:    "p.go",
			content: `// CANARY: REQ=CBIN-7; FEATURE="X"; ASPECT=API; STATUS=IMPL; UPDATED=2025-01-01` + "\n",
			rule:    "pad-flatfile",
			opts:    func(o *upgrade.Options) { o.Registry = sources.Default() },
			wantNew: `// CANARY: REQ=CBIN-007; FEATURE="X"; ASPECT=API; STATUS=IMPL; UPDATED=2025-01-01`,
		},
		{
			name:    "add-updated",
			file:    "a.go",
			content: `// CANARY: REQ=CBIN-101; FEATURE="X"; ASPECT=API; STATUS=IMPL` + "\n",
			rule:    "add-updated",
			opts:    func(o *upgrade.Options) { o.Today = "2025-06-01" },
			wantNew: `// CANARY: REQ=CBIN-101; FEATURE="X"; ASPECT=API; STATUS=IMPL; UPDATED=2025-06-01`,
		},
		{
			name:    "remap",
			file:    "r.go",
			content: `// CANARY: REQ=CBIN-101; FEATURE="X"; ASPECT=API; STATUS=IMPL; UPDATED=2025-01-01` + "\n",
			rule:    "remap",
			opts:    func(o *upgrade.Options) { o.Map = map[string]string{"CBIN-101": "CP-12"} },
			wantNew: `// CANARY: REQ=CP-12; FEATURE="X"; ASPECT=API; STATUS=IMPL; UPDATED=2025-01-01`,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			dir := t.TempDir()
			path := writeFile(t, dir, tc.file, tc.content)

			o := upgrade.Options{Root: dir, Rules: []string{tc.rule}, Write: true}
			if tc.opts != nil {
				tc.opts(&o)
			}
			changes, err := upgrade.Run(o)
			if err != nil {
				t.Fatalf("Run: %v", err)
			}
			if len(changes) != 1 {
				t.Fatalf("expected 1 change, got %d: %+v", len(changes), changes)
			}
			if changes[0].Rule != tc.rule {
				t.Errorf("rule = %q, want %q", changes[0].Rule, tc.rule)
			}
			if changes[0].New != tc.wantNew {
				t.Errorf("New = %q, want %q", changes[0].New, tc.wantNew)
			}
			got := readFile(t, path)
			if !strings.Contains(got, tc.wantNew) {
				t.Errorf("file content = %q, want to contain %q", got, tc.wantNew)
			}
		})
	}
}

// TestCANARY_CBIN_302_Combined runs every rule together against one file
// combining several legacy shapes, and confirms the fully-normalized single
// line comes out the other end.
// CANARY: REQ=CBIN-302; FEATURE="TokenUpgrade"; ASPECT=Engine; STATUS=TESTED; TEST=TestCANARY_CBIN_302_Combined; UPDATED=2026-08-29
func TestCANARY_CBIN_302_Combined(t *testing.T) {
	dir := t.TempDir()
	content := "// CANARY: REQ-42; FEATURE=\"X\"; ASPECT=API; STATUS=FIXED\n"
	path := writeFile(t, dir, "combined.go", content)

	o := upgrade.Options{Root: dir, Write: true, Today: "2025-06-01"}
	changes, err := upgrade.Run(o)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if len(changes) < 3 {
		t.Fatalf("expected multiple rule applications, got %d: %+v", len(changes), changes)
	}
	seen := map[string]bool{}
	for _, c := range changes {
		seen[c.Rule] = true
	}
	for _, want := range []string{"bare-id", "status-fixed", "add-updated"} {
		if !seen[want] {
			t.Errorf("expected rule %q to have fired; changes=%+v", want, changes)
		}
	}

	got := readFile(t, path)
	want := `// CANARY: REQ=REQ-042; FEATURE="X"; ASPECT=API; STATUS=REMOVED; UPDATED=2025-06-01`
	if !strings.Contains(got, want) {
		t.Fatalf("file content = %q, want to contain %q", got, want)
	}
}

// TestCANARY_CBIN_302_Remap verifies the remap rule rewrites both a
// CANARY: token's REQ= value and a GAP_ANALYSIS-style "✅ <ID>" claim line
// in a .md file when Options.Map is supplied.
// CANARY: REQ=CBIN-302; FEATURE="TokenUpgrade"; ASPECT=Engine; STATUS=TESTED; TEST=TestCANARY_CBIN_302_Remap; UPDATED=2026-08-29
func TestCANARY_CBIN_302_Remap(t *testing.T) {
	dir := t.TempDir()
	tokenPath := writeFile(t, dir, "src/token.go",
		`// CANARY: REQ=CBIN-101; FEATURE="X"; ASPECT=API; STATUS=IMPL; UPDATED=2025-01-01`+"\n")
	claimPath := writeFile(t, dir, "GAP_ANALYSIS.md",
		"✅ CBIN-101 - X: some feature\n✅ CBIN-102 - Y: unrelated\n")

	o := upgrade.Options{Root: dir, Write: true, Map: map[string]string{"CBIN-101": "CP-12"}}
	changes, err := upgrade.Run(o)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if len(changes) != 2 {
		t.Fatalf("expected 2 changes (token + claim), got %d: %+v", len(changes), changes)
	}

	gotToken := readFile(t, tokenPath)
	if !strings.Contains(gotToken, "REQ=CP-12") {
		t.Errorf("token file not remapped: %q", gotToken)
	}
	gotClaim := readFile(t, claimPath)
	if !strings.Contains(gotClaim, "✅ CP-12 - X") {
		t.Errorf("claim line not remapped: %q", gotClaim)
	}
	if !strings.Contains(gotClaim, "✅ CBIN-102 - Y") {
		t.Errorf("unrelated claim line should be untouched: %q", gotClaim)
	}
}

// TestCANARY_CP_285_RemapSQLCommentToken proves upgrade's tokenLineRe (kept
// in sync with pkg/canaryscan/parse.go's) recognizes a "--" SQL
// line-comment CANARY token, so --map remaps REQ IDs inside .sql migration
// files the same way it does for // and # comment tokens.
// CANARY: REQ=CP-285; FEATURE="SQLCommentTokens"; ASPECT=Engine; STATUS=TESTED; TEST=TestCANARY_CP_285_RemapSQLCommentToken; UPDATED=2026-08-29
func TestCANARY_CP_285_RemapSQLCommentToken(t *testing.T) {
	dir := t.TempDir()
	sqlPath := writeFile(t, dir, "migrations/001_widgets.sql",
		`-- CANARY: REQ=CBIN-101; FEATURE="X"; ASPECT=Storage; STATUS=IMPL; UPDATED=2025-01-01`+"\n"+
			"CREATE TABLE widgets (id INTEGER PRIMARY KEY);\n")

	o := upgrade.Options{Root: dir, Write: true, Map: map[string]string{"CBIN-101": "CP-12"}}
	changes, err := upgrade.Run(o)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if len(changes) != 1 {
		t.Fatalf("expected 1 change (SQL comment token), got %d: %+v", len(changes), changes)
	}

	got := readFile(t, sqlPath)
	if !strings.Contains(got, "REQ=CP-12") {
		t.Errorf("SQL comment token not remapped: %q", got)
	}
	if !strings.HasPrefix(got, "-- CANARY:") {
		t.Errorf("SQL comment prefix should be preserved: %q", got)
	}
}

// TestCANARY_CBIN_302_Idempotent proves a second run over already-upgraded
// output makes zero further changes.
// CANARY: REQ=CBIN-302; FEATURE="TokenUpgrade"; ASPECT=Engine; STATUS=TESTED; TEST=TestCANARY_CBIN_302_Idempotent; UPDATED=2026-08-29
func TestCANARY_CBIN_302_Idempotent(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "combined.go", "// CANARY: REQ-42; FEATURE=\"X\"; ASPECT=API; STATUS=FIXED\n")
	writeFile(t, dir, "bug.go",
		"// CANARY: BUG=BUG-API-001; TITLE=\"Timeout\";\n"+
			"//         ASPECT=API; STATUS=OPEN;\n"+
			"//         UPDATED=2025-01-01\n")

	o := upgrade.Options{Root: dir, Write: true, Today: "2025-06-01"}
	first, err := upgrade.Run(o)
	if err != nil {
		t.Fatalf("Run 1: %v", err)
	}
	if len(first) == 0 {
		t.Fatalf("expected changes on first run")
	}
	second, err := upgrade.Run(o)
	if err != nil {
		t.Fatalf("Run 2: %v", err)
	}
	if len(second) != 0 {
		t.Fatalf("expected 0 changes on second run, got %d: %+v", len(second), second)
	}
}

// TestCANARY_CBIN_302_FenceProtection proves the legacy continuation shape
// shown as a documentation example inside a fenced code block is left
// untouched by join-multiline, mirroring real doc/template files in this
// repo (e.g. pkg/cmds/init/base/commands/bug.md).
// CANARY: REQ=CBIN-302; FEATURE="TokenUpgrade"; ASPECT=Engine; STATUS=TESTED; TEST=TestCANARY_CBIN_302_FenceProtection; UPDATED=2026-08-29
func TestCANARY_CBIN_302_FenceProtection(t *testing.T) {
	dir := t.TempDir()
	content := "Example:\n\n```go\n" +
		"// CANARY: BUG=BUG-API-NNN; TITLE=\"<Summary>\";\n" +
		"//         ASPECT=API; STATUS=OPEN;\n" +
		"//         UPDATED=<YYYY-MM-DD>\n" +
		"```\n"
	path := writeFile(t, dir, "example.md", content)

	changes, err := upgrade.Run(upgrade.Options{Root: dir, Write: true})
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if len(changes) != 0 {
		t.Fatalf("expected 0 changes inside fenced example, got %d: %+v", len(changes), changes)
	}
	got := readFile(t, path)
	if got != content {
		t.Fatalf("fenced example was modified:\n got: %q\nwant: %q", got, content)
	}
}

// TestCANARY_CBIN_302_MigrateGuard proves CANARY:MIGRATE/START/END lines
// are never touched even when they contain shapes a rule would otherwise fix.
// CANARY: REQ=CBIN-302; FEATURE="TokenUpgrade"; ASPECT=Engine; STATUS=TESTED; TEST=TestCANARY_CBIN_302_MigrateGuard; UPDATED=2026-08-29
func TestCANARY_CBIN_302_MigrateGuard(t *testing.T) {
	dir := t.TempDir()
	content := "// CANARY:MIGRATE see REQ‑999 for details, STATUS=FIXED soon\n" +
		"<!-- CANARY:START -->\nsome content\n<!-- CANARY:END -->\n"
	path := writeFile(t, dir, "guard.md", content)

	changes, err := upgrade.Run(upgrade.Options{Root: dir, Write: true})
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if len(changes) != 0 {
		t.Fatalf("expected 0 changes on guarded lines, got %d: %+v", len(changes), changes)
	}
	got := readFile(t, path)
	if got != content {
		t.Fatalf("guarded content was modified:\n got: %q\nwant: %q", got, content)
	}
}

// TestCANARY_CBIN_302_RuleFiltering proves --rule-equivalent Options.Rules
// restricts which rules run, even when multiple would otherwise fire.
// CANARY: REQ=CBIN-302; FEATURE="TokenUpgrade"; ASPECT=Engine; STATUS=TESTED; TEST=TestCANARY_CBIN_302_RuleFiltering; UPDATED=2026-08-29
func TestCANARY_CBIN_302_RuleFiltering(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "combined.go", "// CANARY: REQ-42; FEATURE=\"X\"; ASPECT=API; STATUS=FIXED; UPDATED=2025-01-01\n")

	changes, err := upgrade.Run(upgrade.Options{Root: dir, Write: false, Rules: []string{"status-fixed"}})
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if len(changes) != 1 || changes[0].Rule != "status-fixed" {
		t.Fatalf("expected exactly 1 status-fixed change, got %+v", changes)
	}
}

// TestCANARY_CBIN_302_ValidRule sanity-checks the rule-name catalog helper.
// CANARY: REQ=CBIN-302; FEATURE="TokenUpgrade"; ASPECT=Engine; STATUS=TESTED; TEST=TestCANARY_CBIN_302_ValidRule; UPDATED=2026-08-29
func TestCANARY_CBIN_302_ValidRule(t *testing.T) {
	if !upgrade.ValidRule("bare-id") {
		t.Errorf("bare-id should be a valid rule")
	}
	if upgrade.ValidRule("not-a-rule") {
		t.Errorf("not-a-rule should not be a valid rule")
	}
	if len(upgrade.AllRules) != 9 {
		t.Errorf("expected 9 named rules, got %d: %v", len(upgrade.AllRules), upgrade.AllRules)
	}
}

// TestCANARY_CBIN_302_AtomicWrite proves Run's write path never truncates
// the original file: a successful rewrite lands fully-correct content, and
// the original file's mode (0600, distinct from writeFile's default 0640)
// survives the stage-then-rename sequence. Also confirms no staging file is
// left behind.
// CANARY: REQ=CBIN-302; FEATURE="TokenUpgrade"; ASPECT=Engine; STATUS=TESTED; TEST=TestCANARY_CBIN_302_AtomicWrite; UPDATED=2026-08-29
func TestCANARY_CBIN_302_AtomicWrite(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "mode.go")
	content := `// CANARY: REQ=CBIN-101; FEATURE="X"; ASPECT=API; STATUS=IMPL` + "\n"
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("write fixture: %v", err)
	}

	changes, err := upgrade.Run(upgrade.Options{Root: dir, Write: true, Today: "2025-06-01"})
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if len(changes) != 1 {
		t.Fatalf("expected 1 change, got %d: %+v", len(changes), changes)
	}

	want := `// CANARY: REQ=CBIN-101; FEATURE="X"; ASPECT=API; STATUS=IMPL; UPDATED=2025-06-01` + "\n"
	got := readFile(t, path)
	if got != want {
		t.Fatalf("content after atomic write = %q, want %q", got, want)
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("readdir: %v", err)
	}
	for _, e := range entries {
		if strings.HasPrefix(e.Name(), ".canary-stage-") {
			t.Errorf("leftover staging file after successful write: %s", e.Name())
		}
	}

	if runtime.GOOS == "windows" {
		return // chmod is unreliable on Windows; content correctness above is the load-bearing assertion there.
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("stat: %v", err)
	}
	if perm := info.Mode().Perm(); perm != 0o600 {
		t.Errorf("mode after atomic write = %o, want 0600 (original mode must survive the rewrite)", perm)
	}
}

// TestCANARY_CBIN_302_CRLF proves add-updated's suffix/whitespace handling
// operates on the \r-stripped body so a CRLF line's trailing \r stays at
// the true end of the line instead of getting embedded mid-line (the
// pre-fix bug produced "...STATUS=IMPL\r; UPDATED=..."), and that a CRLF
// file no rule touches comes out byte-identical.
// CANARY: REQ=CBIN-302; FEATURE="TokenUpgrade"; ASPECT=Engine; STATUS=TESTED; TEST=TestCANARY_CBIN_302_CRLF; UPDATED=2026-08-29
func TestCANARY_CBIN_302_CRLF(t *testing.T) {
	t.Run("MissingUpdated", func(t *testing.T) {
		dir := t.TempDir()
		content := "// CANARY: REQ=CBIN-101; FEATURE=\"X\"; ASPECT=API; STATUS=IMPL\r\n"
		path := writeFile(t, dir, "crlf.go", content)

		changes, err := upgrade.Run(upgrade.Options{Root: dir, Write: true, Rules: []string{"add-updated"}, Today: "2025-06-01"})
		if err != nil {
			t.Fatalf("Run: %v", err)
		}
		if len(changes) != 1 {
			t.Fatalf("expected 1 change, got %d: %+v", len(changes), changes)
		}
		want := "// CANARY: REQ=CBIN-101; FEATURE=\"X\"; ASPECT=API; STATUS=IMPL; UPDATED=2025-06-01\r\n"
		got := readFile(t, path)
		if got != want {
			t.Fatalf("CRLF rewrite = %q, want %q", got, want)
		}
		if strings.Contains(got, "\r;") {
			t.Fatalf("trailing \\r embedded mid-line: %q", got)
		}
	})

	t.Run("BlockCommentSuffix", func(t *testing.T) {
		dir := t.TempDir()
		content := "/* CANARY: REQ=CBIN-101; FEATURE=\"X\"; ASPECT=API; STATUS=IMPL */\r\n"
		path := writeFile(t, dir, "crlf_block.go", content)

		changes, err := upgrade.Run(upgrade.Options{Root: dir, Write: true, Rules: []string{"add-updated"}, Today: "2025-06-01"})
		if err != nil {
			t.Fatalf("Run: %v", err)
		}
		if len(changes) != 1 {
			t.Fatalf("expected 1 change, got %d: %+v", len(changes), changes)
		}
		want := "/* CANARY: REQ=CBIN-101; FEATURE=\"X\"; ASPECT=API; STATUS=IMPL; UPDATED=2025-06-01 */\r\n"
		got := readFile(t, path)
		if got != want {
			t.Fatalf("CRLF block-comment rewrite = %q, want %q", got, want)
		}
	})

	t.Run("UntouchedByteIdentical", func(t *testing.T) {
		dir := t.TempDir()
		content := "// CANARY: REQ=CBIN-101; FEATURE=\"X\"; ASPECT=API; STATUS=IMPL; UPDATED=2025-01-01\r\n" +
			"other content\r\nmore content\r\n"
		path := writeFile(t, dir, "crlf_untouched.go", content)

		changes, err := upgrade.Run(upgrade.Options{Root: dir, Write: true})
		if err != nil {
			t.Fatalf("Run: %v", err)
		}
		if len(changes) != 0 {
			t.Fatalf("expected 0 changes, got %d: %+v", len(changes), changes)
		}
		got := readFile(t, path)
		if got != content {
			t.Fatalf("untouched CRLF file changed:\n got: %q\nwant: %q", got, content)
		}
	})
}

// TestCANARY_CBIN_302_PlaceholderGuard proves every rule skips a token
// line whose content looks like a template/example placeholder (Go
// template braces, angle-bracket placeholders, printf verbs) rather than a
// real value, mirroring canaryscan.parseKV's own escape hatch — even when
// the line would otherwise clearly match a rule (missing UPDATED=, a
// BUG=-only line missing FEATURE=).
// CANARY: REQ=CBIN-302; FEATURE="TokenUpgrade"; ASPECT=Engine; STATUS=TESTED; TEST=TestCANARY_CBIN_302_PlaceholderGuard; UPDATED=2026-08-29
func TestCANARY_CBIN_302_PlaceholderGuard(t *testing.T) {
	cases := []struct {
		name    string
		content string
	}{
		{
			name:    "GoTemplateBraces",
			content: `// CANARY: REQ={{.ReqID}}-SECURITY_REVIEW-105; FEATURE="X"; ASPECT=API; STATUS=IMPL` + "\n",
		},
		{
			name:    "AngleBracketPlaceholder",
			content: `// CANARY: REQ=CBIN-101; FEATURE="<ShortProblemSummary>"; ASPECT=API; STATUS=IMPL` + "\n",
		},
		{
			name:    "PrintfVerb",
			content: `// CANARY: BUG=%s; TITLE="X"; ASPECT=API; STATUS=OPEN; UPDATED=2025-01-01` + "\n",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			dir := t.TempDir()
			path := writeFile(t, dir, "placeholder.go", tc.content)

			changes, err := upgrade.Run(upgrade.Options{Root: dir, Write: true, Today: "2025-06-01"})
			if err != nil {
				t.Fatalf("Run: %v", err)
			}
			if len(changes) != 0 {
				t.Fatalf("expected 0 changes on placeholder content, got %d: %+v", len(changes), changes)
			}
			got := readFile(t, path)
			if got != tc.content {
				t.Fatalf("placeholder content was modified:\n got: %q\nwant: %q", got, tc.content)
			}
		})
	}
}

// TestCANARY_CBIN_302_UnicodeHyphenProse proves unicode-hyphen only
// normalizes REQ=/BUG=/TASK=/PARENT= values and bare ID segments: a
// legitimate en-dash inside FEATURE= prose is left untouched while a
// unicode hyphen inside a REQ= id on the same line is normalized.
// CANARY: REQ=CBIN-302; FEATURE="TokenUpgrade"; ASPECT=Engine; STATUS=TESTED; TEST=TestCANARY_CBIN_302_UnicodeHyphenProse; UPDATED=2026-08-29
func TestCANARY_CBIN_302_UnicodeHyphenProse(t *testing.T) {
	dir := t.TempDir()
	content := "// CANARY: REQ=CBIN‑101; FEATURE=\"Retry–Backoff\"; ASPECT=API; STATUS=IMPL; UPDATED=2025-01-01\n"
	path := writeFile(t, dir, "prose.go", content)

	changes, err := upgrade.Run(upgrade.Options{Root: dir, Write: true, Rules: []string{"unicode-hyphen"}})
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if len(changes) != 1 {
		t.Fatalf("expected 1 change, got %d: %+v", len(changes), changes)
	}

	got := readFile(t, path)
	want := `// CANARY: REQ=CBIN-101; FEATURE="Retry–Backoff"; ASPECT=API; STATUS=IMPL; UPDATED=2025-01-01`
	if !strings.Contains(got, want) {
		t.Fatalf("file content = %q, want to contain %q", got, want)
	}
	if strings.Contains(got, `FEATURE="Retry-Backoff"`) {
		t.Fatalf("prose en-dash inside FEATURE= must not be rewritten: %q", got)
	}
}

// TestCANARY_CBIN_302_RemapCollision proves remap does exact-ID matching,
// not a prefix match: a map entry for "CBIN-1" must not also rewrite
// "CBIN-10" tokens or "✅ CBIN-10" claim lines.
// CANARY: REQ=CBIN-302; FEATURE="TokenUpgrade"; ASPECT=Engine; STATUS=TESTED; TEST=TestCANARY_CBIN_302_RemapCollision; UPDATED=2026-08-29
func TestCANARY_CBIN_302_RemapCollision(t *testing.T) {
	dir := t.TempDir()
	tokenPath := writeFile(t, dir, "src/token.go",
		"// CANARY: REQ=CBIN-1; FEATURE=\"One\"; ASPECT=API; STATUS=IMPL; UPDATED=2025-01-01\n"+
			"// CANARY: REQ=CBIN-10; FEATURE=\"Ten\"; ASPECT=API; STATUS=IMPL; UPDATED=2025-01-01\n")
	claimPath := writeFile(t, dir, "GAP_ANALYSIS.md",
		"✅ CBIN-1 - One: first feature\n✅ CBIN-10 - Ten: tenth feature\n")

	o := upgrade.Options{Root: dir, Write: true, Rules: []string{"remap"}, Map: map[string]string{"CBIN-1": "CP-9"}}
	changes, err := upgrade.Run(o)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if len(changes) != 2 {
		t.Fatalf("expected 2 changes (one token + one claim), got %d: %+v", len(changes), changes)
	}

	gotToken := readFile(t, tokenPath)
	if !strings.Contains(gotToken, "REQ=CP-9;") {
		t.Errorf("CBIN-1 token not remapped: %q", gotToken)
	}
	if !strings.Contains(gotToken, "REQ=CBIN-10;") {
		t.Errorf("CBIN-10 token must not be touched by a CBIN-1 map entry: %q", gotToken)
	}

	gotClaim := readFile(t, claimPath)
	if !strings.Contains(gotClaim, "✅ CP-9 - One") {
		t.Errorf("CBIN-1 claim not remapped: %q", gotClaim)
	}
	if !strings.Contains(gotClaim, "✅ CBIN-10 - Ten") {
		t.Errorf("CBIN-10 claim must not be touched by a CBIN-1 map entry: %q", gotClaim)
	}
}

// TestCANARY_CBIN_302_MDHeadingFenceProtection proves a `# CANARY:` heading
// shown as a documentation example inside a fenced code block (e.g. a
// skill doc illustrating the pre-upgrade shape) is left untouched by
// md-heading, same as every other rule's fenced content — the previous
// "md-heading is exempt from the fence mask" behavior was itself the bug.
// CANARY: REQ=CBIN-302; FEATURE="TokenUpgrade"; ASPECT=Engine; STATUS=TESTED; TEST=TestCANARY_CBIN_302_MDHeadingFenceProtection; UPDATED=2026-08-29
func TestCANARY_CBIN_302_MDHeadingFenceProtection(t *testing.T) {
	dir := t.TempDir()
	content := "Example of the old heading shape:\n\n```markdown\n" +
		`# CANARY: REQ=CBIN-101; FEATURE="X"; ASPECT=Docs; STATUS=IMPL; UPDATED=2025-01-01` + "\n" +
		"```\n"
	path := writeFile(t, dir, "heading-example.md", content)

	changes, err := upgrade.Run(upgrade.Options{Root: dir, Write: true, Rules: []string{"md-heading"}})
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if len(changes) != 0 {
		t.Fatalf("expected 0 changes inside fenced example, got %d: %+v", len(changes), changes)
	}
	got := readFile(t, path)
	if got != content {
		t.Fatalf("fenced md-heading example was modified:\n got: %q\nwant: %q", got, content)
	}
}
