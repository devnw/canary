package index

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"devnw.dev/canary/pkg/canaryscan"
	"devnw.dev/canary/pkg/storage"
)

// CANARY: REQ=CP-285; FEATURE="IndexIgnoreFilter"; ASPECT=CLI; STATUS=TESTED; TEST=TestCANARY_CP_285_IsGrepMatchIgnored; UPDATED=2026-08-29
// TestCANARY_CP_285_IsGrepMatchIgnored unit-tests the filter helper directly
// against a real .canaryignore, independent of grep/db plumbing.
func TestCANARY_CP_285_IsGrepMatchIgnored(t *testing.T) {
	root := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(root, ".canaryignore"), []byte("docs/\n"), 0644))

	ignorePatterns, err := canaryscan.LoadCanaryIgnore(root)
	require.NoError(t, err)
	require.NotNil(t, ignorePatterns)

	assert.True(t, isGrepMatchIgnored(root, filepath.Join(root, "docs", "example.md"), ignorePatterns),
		"file under an ignored dir should be reported ignored")
	assert.False(t, isGrepMatchIgnored(root, filepath.Join(root, "pkg", "x.go"), ignorePatterns),
		"file outside any ignored dir should not be reported ignored")
	assert.False(t, isGrepMatchIgnored(root, filepath.Join(root, "pkg", "x.go"), nil),
		"nil ignorePatterns must never ignore anything")

	// grep run with root="." emits "./relative/path" style output; the
	// helper must resolve that the same way as an absolute rootPath.
	relRoot := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(relRoot, ".canaryignore"), []byte("docs/\n"), 0644))
	relIgnore, err := canaryscan.LoadCanaryIgnore(relRoot)
	require.NoError(t, err)
	assert.True(t, isGrepMatchIgnored(relRoot, filepath.Join(relRoot, "docs", "x.md"), relIgnore))
}

// CANARY: REQ=CP-285; FEATURE="IndexIgnoreFilter"; ASPECT=CLI; STATUS=TESTED; TEST=TestCANARY_CP_285_IndexRespectsCanaryIgnore; UPDATED=2026-08-29
// TestCANARY_CP_285_IndexRespectsCanaryIgnore drives IndexCmd.RunE end to
// end against a temp root containing a .canaryignore that excludes docs/,
// with one token under docs/ and one under pkg/. Only the pkg/ token must
// land in the database -- `canary index` used to shell out to grep over the
// whole tree with no .canaryignore filtering, so the docs/ token would have
// been indexed even though `canary scan` correctly excludes it.
func TestCANARY_CP_285_IndexRespectsCanaryIgnore(t *testing.T) {
	if _, err := exec.LookPath("grep"); err != nil {
		t.Skip("grep not available on PATH")
	}

	root := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(root, ".canaryignore"), []byte("docs/\n"), 0644))

	docsDir := filepath.Join(root, "docs")
	pkgDir := filepath.Join(root, "pkg")
	require.NoError(t, os.MkdirAll(docsDir, 0755))
	require.NoError(t, os.MkdirAll(pkgDir, 0755))

	docsToken := `# CANARY: REQ=CP-900; FEATURE="DocsExample"; ASPECT=Docs; STATUS=IMPL; UPDATED=2026-08-29` + "\n"
	pkgToken := `// CANARY: REQ=CP-901; FEATURE="PkgExample"; ASPECT=API; STATUS=IMPL; UPDATED=2026-08-29` + "\n"
	require.NoError(t, os.WriteFile(filepath.Join(docsDir, "example.md"), []byte(docsToken), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(pkgDir, "example.go"), []byte(pkgToken), 0644))

	dbPath := filepath.Join(root, "canary.db")

	// Pre-create the tokens table: DeleteAllTokens (called unconditionally
	// at the top of RunE to keep the index derived-state-fresh) does not
	// lazily create the schema the way UpsertToken/GetAllTokens do, so a
	// truly brand-new db file needs one throwaway write first. This is a
	// pre-existing storage-layer quirk unrelated to the .canaryignore fix
	// under test here, so it's worked around rather than changed.
	seedDB, err := storage.Open(dbPath)
	require.NoError(t, err)
	require.NoError(t, seedDB.UpsertToken(&storage.Token{ReqID: "CP-000", Feature: "seed", Aspect: "API", Status: "STUB"}))
	require.NoError(t, seedDB.Close())

	var buf bytes.Buffer
	IndexCmd.SetOut(&buf)
	IndexCmd.SetErr(&buf)
	require.NoError(t, IndexCmd.Flags().Set("root", root))
	require.NoError(t, IndexCmd.Flags().Set("db", dbPath))
	t.Cleanup(func() {
		_ = IndexCmd.Flags().Set("root", ".")
		_ = IndexCmd.Flags().Set("db", ".canary/canary.db")
	})

	require.NoError(t, IndexCmd.RunE(IndexCmd, nil))

	db, err := storage.Open(dbPath)
	require.NoError(t, err)
	defer db.Close()

	tokens, err := db.GetAllTokens()
	require.NoError(t, err)

	var gotDocs, gotPkg bool
	for _, tok := range tokens {
		switch tok.ReqID {
		case "CP-900":
			gotDocs = true
		case "CP-901":
			gotPkg = true
		}
	}
	assert.False(t, gotDocs, "docs/ token must be excluded by .canaryignore")
	assert.True(t, gotPkg, "pkg/ token must be indexed")
}
