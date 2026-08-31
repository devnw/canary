// Copyright (c) 2025 by Developer Network.
//
// For more details, see the LICENSE file in the root directory of this
// source code repository or contact Developer Network at info@devnw.com.

package canaryscan

import (
	"os"
	"path/filepath"
	"testing"
)

// TestScanMultiLanguageTokens proves the scanner recognises CANARY tokens in
// the line- and block-comment styles of several languages. Ported from the
// former root-package scan_test.go so the single canaryscan parser remains the
// place these forms are proven.
func TestScanMultiLanguageTokens(t *testing.T) {
	dir := t.TempDir()
	cases := map[string]string{
		"example.zig":   "// CANARY: REQ=CBIN-201; FEATURE=\"ZigCore\"; ASPECT=API; STATUS=IMPL; OWNER=zig; UPDATED=2025-11-02",
		"lib.rs":        "// CANARY: REQ=CBIN-202; FEATURE=\"RustCore\"; ASPECT=Engine; STATUS=IMPL; OWNER=rust; UPDATED=2025-11-02",
		"component.tsx": "// CANARY: REQ=CBIN-203; FEATURE=\"TSXComp\"; ASPECT=FrontEnd; STATUS=STUB; OWNER=web; UPDATED=2025-11-02",
		"view.jsx":      "/* CANARY: REQ=CBIN-204; FEATURE=\"JSXView\"; ASPECT=FrontEnd; STATUS=IMPL; OWNER=web; UPDATED=2025-11-02 */",
	}
	for name, line := range cases {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(line+"\n"), 0o600); err != nil {
			t.Fatalf("write %s: %v", name, err)
		}
	}
	rep, err := Scan(dir, DefaultSkipRegex(), nil, nil, nil)
	if err != nil {
		t.Fatalf("scan: %v", err)
	}
	want := map[string]bool{"CBIN-201": true, "CBIN-202": true, "CBIN-203": true, "CBIN-204": true}
	for _, r := range rep.Requirements {
		delete(want, r.ID)
	}
	if len(want) != 0 {
		t.Fatalf("missing requirements parsed: %v", want)
	}
}

// TestScanBlockCommentToken proves the scanner recognises a CANARY token in a
// single-line C-style block comment (/* ... */). Ported from the former
// root-package scan_test.go. Note: multi-line JSDoc (/**\n * CANARY) forms are
// deliberately NOT asserted here -- the production canaryscan parser matches on
// a token line, and the star-continuation JSDoc shapes were only ever handled
// by the now-deleted gate scanner, never by the CLI.
func TestScanBlockCommentToken(t *testing.T) {
	dir := t.TempDir()
	content := `/* CANARY: REQ=CBIN-208; FEATURE="PlainBlock"; ASPECT=API; STATUS=IMPL; OWNER=docs; UPDATED=2025-11-02 */
`
	if err := os.WriteFile(filepath.Join(dir, "doc.js"), []byte(content), 0o600); err != nil {
		t.Fatalf("write doc.js: %v", err)
	}
	rep, err := Scan(dir, DefaultSkipRegex(), nil, nil, nil)
	if err != nil {
		t.Fatalf("scan: %v", err)
	}
	found := false
	for _, r := range rep.Requirements {
		if r.ID == "CBIN-208" {
			found = true
		}
	}
	if !found {
		t.Fatal("CBIN-208 in /* ... */ block comment was not scanned")
	}
}
