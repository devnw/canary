package index

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"devnw.dev/canary/pkg/storage"
)

// runIndex drives IndexCmd.RunE against root, restoring the shared command's
// flags afterwards. The command object is a package-level singleton, so a
// test that leaves --root or --db set would poison its neighbours.
func runIndex(t *testing.T, root, dbPath string) error {
	t.Helper()
	var buf bytes.Buffer
	IndexCmd.SetOut(&buf)
	IndexCmd.SetErr(&buf)
	require.NoError(t, IndexCmd.Flags().Set("root", root))
	require.NoError(t, IndexCmd.Flags().Set("db", dbPath))
	t.Cleanup(func() {
		_ = IndexCmd.Flags().Set("root", ".")
		_ = IndexCmd.Flags().Set("db", ".canary/canary.db")
	})
	return IndexCmd.RunE(IndexCmd, nil)
}

// CANARY: REQ=CP-285; FEATURE="IndexIgnoreFilter"; ASPECT=CLI; STATUS=TESTED; TEST=TestCANARY_CP_285_IndexRespectsCanaryIgnore; UPDATED=2026-08-30
// TestCANARY_CP_285_IndexRespectsCanaryIgnore drives IndexCmd.RunE end to
// end against a temp root containing a .canaryignore that excludes docs/,
// with one token under docs/ and one under pkg/. Only the pkg/ token must
// land in the database -- `canary index` used to shell out to grep over the
// whole tree with no .canaryignore filtering, so the docs/ token would have
// been indexed even though `canary scan` correctly excludes it. The index now
// walks with canaryscan, the same walk `canary scan` uses, so the two cannot
// drift apart again.
func TestCANARY_CP_285_IndexRespectsCanaryIgnore(t *testing.T) {
	root := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(root, ".canaryignore"), []byte("docs/\n"), 0600))

	docsDir := filepath.Join(root, "docs")
	pkgDir := filepath.Join(root, "pkg")
	require.NoError(t, os.MkdirAll(docsDir, 0750))
	require.NoError(t, os.MkdirAll(pkgDir, 0750))

	docsToken := `# CANARY: REQ=CP-900; FEATURE="DocsExample"; ASPECT=Docs; STATUS=IMPL; UPDATED=2026-08-29` + "\n"
	pkgToken := `// CANARY: REQ=CP-901; FEATURE="PkgExample"; ASPECT=API; STATUS=IMPL; UPDATED=2026-08-29` + "\n"
	require.NoError(t, os.WriteFile(filepath.Join(docsDir, "example.md"), []byte(docsToken), 0600))
	require.NoError(t, os.WriteFile(filepath.Join(pkgDir, "example.go"), []byte(pkgToken), 0600))

	dbPath := filepath.Join(root, "canary.db")
	require.NoError(t, runIndex(t, root, dbPath))

	db, err := storage.OpenRW(dbPath)
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

// CANARY: REQ=CP-285; FEATURE="IndexRebuild"; ASPECT=CLI; STATUS=TESTED; TEST=TestCANARY_CP_285_IndexRebuildPrunes; UPDATED=2026-08-30
// TestCANARY_CP_285_IndexRebuildPrunes proves the index is derived state: a
// row whose token no longer exists on disk must not survive a re-index.
func TestCANARY_CP_285_IndexRebuildPrunes(t *testing.T) {
	root := t.TempDir()
	tokenFile := filepath.Join(root, "a.go")
	require.NoError(t, os.WriteFile(tokenFile,
		[]byte(`// CANARY: REQ=CP-902; FEATURE="Gone"; ASPECT=API; STATUS=IMPL; UPDATED=2026-08-29`+"\n"), 0600))

	dbPath := filepath.Join(root, "canary.db")
	require.NoError(t, runIndex(t, root, dbPath))

	require.NoError(t, os.Remove(tokenFile))
	require.NoError(t, runIndex(t, root, dbPath))

	db, err := storage.OpenRW(dbPath)
	require.NoError(t, err)
	defer db.Close()

	tokens, err := db.GetAllTokens()
	require.NoError(t, err)
	for _, tok := range tokens {
		assert.NotEqual(t, "CP-902", tok.ReqID, "row for a deleted token survived the re-index")
	}
}

// CANARY: REQ=ENG-4307; FEATURE="IndexCmd"; ASPECT=CLI; STATUS=TESTED; TEST=TestIndexStoresLineNumbersAndDeclaredUpdated; UPDATED=2026-08-30
// TestIndexStoresLineNumbersAndDeclaredUpdated pins two facts the grep-based
// indexer got wrong: the row must point at the token's real line, and the
// UPDATED it stores must be the one the author wrote.
func TestIndexStoresLineNumbersAndDeclaredUpdated(t *testing.T) {
	root := t.TempDir()
	body := "package x\n\n// leading comment\n" +
		`// CANARY: REQ=CP-903; FEATURE="Located"; ASPECT=API; STATUS=IMPL; UPDATED=2026-08-29` + "\n"
	require.NoError(t, os.WriteFile(filepath.Join(root, "a.go"), []byte(body), 0600))

	dbPath := filepath.Join(root, "canary.db")
	require.NoError(t, runIndex(t, root, dbPath))

	db, err := storage.OpenRW(dbPath)
	require.NoError(t, err)
	defer db.Close()

	tokens, err := db.GetTokensByReqID("", "CP-903")
	require.NoError(t, err)
	require.Len(t, tokens, 1)
	assert.Equal(t, 4, tokens[0].LineNumber, "token line number must be the real line")
	assert.Equal(t, "2026-08-29", tokens[0].UpdatedAt)
	assert.NotEmpty(t, tokens[0].ContentHash, "content hash must be recorded")
	assert.Equal(t, "default", tokens[0].ProjectID)
}

// CANARY: REQ=ENG-4307; FEATURE="IndexCmd"; ASPECT=CLI; STATUS=TESTED; TEST=TestIndexRefusesUnparseableToken; UPDATED=2026-08-30
// TestIndexRefusesUnparseableToken proves a token the canonical parser
// rejects fails the run instead of being skipped with a warning.
func TestIndexRefusesUnparseableToken(t *testing.T) {
	root := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(root, "bad.go"),
		[]byte(`// CANARY: REQ=CP-904; FEATURE="Bad"; ASPECT=NOPE; STATUS=IMPL; UPDATED=2026-08-29`+"\n"), 0600))

	dbPath := filepath.Join(root, "canary.db")
	err := runIndex(t, root, dbPath)
	require.Error(t, err, "index must refuse a tree it cannot scan cleanly")
}
