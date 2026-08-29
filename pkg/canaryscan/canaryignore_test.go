package canaryscan

import (
	"os"
	"path/filepath"
	"testing"
)

// CANARY: REQ=CP-285; FEATURE="CanaryIgnoreAnchoring"; ASPECT=Engine; STATUS=TESTED; TEST=TestCANARY_CP_285_NestedDocsDirsScanned; UPDATED=2026-08-29
// TestCANARY_CP_285_NestedDocsDirsScanned is a regression test for the
// nested-dir footgun in gitignore-style .canaryignore patterns: a bare
// "docs/" (unanchored) matches ANY directory named docs at any depth, so a
// nested pkg/docs/ was silently excluded from scans right alongside the
// intended root-level docs/. Anchoring the pattern as "/docs/" scopes it to
// the root only, so pkg/docs/ is scanned while the root docs/ stays
// excluded.
func TestCANARY_CP_285_NestedDocsDirsScanned(t *testing.T) {
	root := t.TempDir()

	if err := os.WriteFile(filepath.Join(root, ".canaryignore"), []byte("/docs/\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	// Root-level docs/ - must stay excluded.
	rootDocs := filepath.Join(root, "docs")
	if err := os.MkdirAll(rootDocs, 0o750); err != nil {
		t.Fatal(err)
	}
	rootDocToken := "# CANARY: REQ=CP-810; FEATURE=\"RootDoc\"; ASPECT=Docs; STATUS=IMPL; UPDATED=2026-08-29\n"
	if err := os.WriteFile(filepath.Join(rootDocs, "example.md"), []byte(rootDocToken), 0o600); err != nil {
		t.Fatal(err)
	}

	// Nested pkg/docs/ - must now be scanned.
	pkgDocs := filepath.Join(root, "pkg", "docs")
	if err := os.MkdirAll(pkgDocs, 0o750); err != nil {
		t.Fatal(err)
	}
	pkgDocToken := "// CANARY: REQ=CP-811; FEATURE=\"PkgDoc\"; ASPECT=API; STATUS=IMPL; UPDATED=2026-08-29\n"
	if err := os.WriteFile(filepath.Join(pkgDocs, "x.go"), []byte(pkgDocToken), 0o600); err != nil {
		t.Fatal(err)
	}

	ignorePatterns, err := LoadCanaryIgnore(root)
	if err != nil {
		t.Fatal(err)
	}
	if ignorePatterns == nil {
		t.Fatal("expected non-nil ignorePatterns")
	}

	rep, err := Scan(root, DefaultSkipRegex(), nil, ignorePatterns)
	if err != nil {
		t.Fatal(err)
	}

	found := map[string]bool{}
	for _, r := range rep.Requirements {
		found[r.ID] = true
	}

	if found["CP-810"] {
		t.Error("root docs/ token CP-810 should be excluded by /docs/ but was scanned")
	}
	if !found["CP-811"] {
		t.Error("nested pkg/docs/ token CP-811 should be scanned but was excluded")
	}
}
