// Copyright (c) 2025 by Developer Network.
//
// For more details, see the LICENSE file in the root directory of this
// source code repository or contact Developer Network at info@devnw.com.

package deps

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"devnw.dev/canary/pkg/external"
	"devnw.dev/canary/pkg/storage"
)

// depsTestRoot lays out a temp project root with:
//   - .canary/project.yaml declaring a CBIN flatfile source and an ENG jira
//     source (default status map: TESTED/BENCHED -> "Done").
//   - .canary/specs/CBIN-147-test/spec.md whose "## Dependencies" section is
//     depLines, joined one-per-line.
//
// It chdirs into the root (restored via t.Cleanup) and returns the root.
func depsTestRoot(t *testing.T, depLines ...string) string {
	t.Helper()
	root := t.TempDir()

	projectYAML := `project:
  name: "demo"
  key: "CBIN"
sources:
  - name: core
    type: flatfile
    key: "CBIN"
  - name: eng
    type: jira
    key: "ENG"
    url: "https://example.atlassian.net/browse/{id}"
`
	require.NoError(t, os.MkdirAll(filepath.Join(root, ".canary"), 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(root, ".canary", "project.yaml"), []byte(projectYAML), 0o644))

	specDir := filepath.Join(root, ".canary", "specs", "CBIN-147-test")
	require.NoError(t, os.MkdirAll(specDir, 0o755))
	spec := "# Test Spec\n\n## Dependencies\n\n### Full Dependencies\n"
	for _, l := range depLines {
		spec += "- " + l + "\n"
	}
	spec += "\n## Features\n- TestFeature\n"
	require.NoError(t, os.WriteFile(filepath.Join(specDir, "spec.md"), []byte(spec), 0o644))

	originalDir, err := os.Getwd()
	require.NoError(t, err)
	t.Cleanup(func() { _ = os.Chdir(originalDir) })
	require.NoError(t, os.Chdir(root))

	return root
}

// TestCANARY_ENG_3960_DepsCheck_ExternalSatisfied proves `deps check` renders
// a satisfied external dependency as "✔ external <id> (<status>)" and does
// not count it as blocking.
func TestCANARY_ENG_3960_DepsCheck_ExternalSatisfied(t *testing.T) {
	root := depsTestRoot(t, "ENG-12 (upstream)")
	require.NoError(t, external.SaveCache(root, map[string]string{"ENG-12": "Done"}, mustFreshTime()))

	var buf bytes.Buffer
	cmd := createDepsCheckCommand()
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetArgs([]string{"CBIN-147"})
	err := cmd.Execute()

	output := buf.String()
	assert.Contains(t, output, "✔ external ENG-12 (Done)")
	assert.NoError(t, err, "satisfied external dependency must not fail deps check")
}

// TestCANARY_ENG_3960_DepsCheck_ExternalUnsatisfied proves an unsatisfied
// cached external dependency renders "✖ external <id> (<status>)" and blocks
// (deps check exits non-zero).
func TestCANARY_ENG_3960_DepsCheck_ExternalUnsatisfied(t *testing.T) {
	root := depsTestRoot(t, "ENG-13 (upstream)")
	require.NoError(t, external.SaveCache(root, map[string]string{"ENG-13": "In Progress"}, mustFreshTime()))

	var buf bytes.Buffer
	cmd := createDepsCheckCommand()
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetArgs([]string{"CBIN-147"})
	err := cmd.Execute()

	output := buf.String()
	assert.Contains(t, output, "✖ external ENG-13 (In Progress)")
	assert.Error(t, err, "unsatisfied external dependency must fail deps check")
}

// TestCANARY_ENG_3960_DepsCheck_ExternalUnknown proves an external
// dependency with no cached status renders "? external <id> (no cached
// ticket status)" and does NOT block (degradation is sacred).
func TestCANARY_ENG_3960_DepsCheck_ExternalUnknown(t *testing.T) {
	depsTestRoot(t, "ENG-14 (upstream)") // no cache written at all

	var buf bytes.Buffer
	cmd := createDepsCheckCommand()
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetArgs([]string{"CBIN-147"})
	err := cmd.Execute()

	output := buf.String()
	assert.Contains(t, output, "? external ENG-14 (no cached ticket status)")
	assert.NoError(t, err, "unknown external dependency must not fail deps check by default")
}

// TestCANARY_ENG_3960_DepsValidate_ExternalNotMissing proves `deps validate`
// never reports an external dependency as a missing requirement.
func TestCANARY_ENG_3960_DepsValidate_ExternalNotMissing(t *testing.T) {
	depsTestRoot(t, "ENG-12 (upstream)") // no cache: unknown, but still external

	var buf bytes.Buffer
	cmd := createDepsValidateCommand()
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetArgs([]string{})
	err := cmd.Execute()

	output := buf.String()
	assert.NotContains(t, output, "Missing requirement: ENG-12")
	assert.NoError(t, err, "external dependency must not fail validate by default")
}

// TestCANARY_ENG_3960_DepsValidate_ExternalCountsLine proves `deps validate`
// appends a summary line counting external deps by state.
func TestCANARY_ENG_3960_DepsValidate_ExternalCountsLine(t *testing.T) {
	root := depsTestRoot(t, "ENG-1 (a)", "ENG-2 (b)", "ENG-3 (c)")
	require.NoError(t, external.SaveCache(root, map[string]string{
		"ENG-1": "Done",        // satisfied
		"ENG-2": "In Progress", // unsatisfied
		// ENG-3 absent from cache -> unknown
	}, mustFreshTime()))

	var buf bytes.Buffer
	cmd := createDepsValidateCommand()
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetArgs([]string{})
	_ = cmd.Execute()

	output := buf.String()
	assert.Contains(t, output, "external: satisfied=1 unsatisfied=1 unknown=1")
}

// TestCANARY_ENG_3960_DepsValidate_StrictExternalFails proves
// --strict-external makes validate fail when any external dependency is
// unsatisfied or unknown.
func TestCANARY_ENG_3960_DepsValidate_StrictExternalFails(t *testing.T) {
	depsTestRoot(t, "ENG-14 (upstream)") // no cache -> unknown

	var buf bytes.Buffer
	cmd := createDepsValidateCommand()
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetArgs([]string{"--strict-external"})
	err := cmd.Execute()

	assert.Error(t, err, "--strict-external must fail validate on unknown external status")
	assert.Contains(t, buf.String(), "external: satisfied=0 unsatisfied=0 unknown=1")
}

// TestCANARY_ENG_3960_DepsGraph_MermaidExternalClass proves the `deps graph
// --format mermaid` CLI wiring passes an isExternal closure through to
// FormatMermaid so external nodes get the "external" mermaid class.
func TestCANARY_ENG_3960_DepsGraph_MermaidExternalClass(t *testing.T) {
	depsTestRoot(t, "ENG-12 (upstream)")

	var buf bytes.Buffer
	cmd := createDepsGraphCommand()
	cmd.SetOut(&buf)
	cmd.SetArgs([]string{"CBIN-147", "--format", "mermaid"})
	require.NoError(t, cmd.Execute())

	output := buf.String()
	assert.Contains(t, output, "classDef external stroke-dasharray: 5 5")
	assert.Contains(t, output, "class ENG_12 external")
	assert.NotContains(t, output, "class CBIN_147 external")
}

func mustFreshTime() time.Time { return time.Now().UTC() }

// TestCANARY_ENG_3961_DepsCheck_PeerDetailSurfacesName proves `deps check`
// surfaces the peer's name in its Detail line for a dependency resolved by
// a configured peer project's status.json -- Resolution.Detail's
// "peer:<name>" comes through unchanged via ShortDetail into the "✔
// external ..." rendering.
func TestCANARY_ENG_3961_DepsCheck_PeerDetailSurfacesName(t *testing.T) {
	root := depsTestRoot(t, "ENG-12 (upstream)")

	peerYAML := `project:
  name: "demo"
  key: "CBIN"
sources:
  - name: core
    type: flatfile
    key: "CBIN"
  - name: eng
    type: jira
    key: "ENG"
    url: "https://example.atlassian.net/browse/{id}"
peers:
  - name: upstream-repo
    root: "peer"
`
	require.NoError(t, os.WriteFile(filepath.Join(root, ".canary", "project.yaml"), []byte(peerYAML), 0o644))

	peerStatus := `{"requirements":[{"id":"ENG-12","features":[{"feature":"X","aspect":"Engine","status":"TESTED"}]}]}`
	require.NoError(t, os.MkdirAll(filepath.Join(root, "peer"), 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(root, "peer", "status.json"), []byte(peerStatus), 0o644))

	var buf bytes.Buffer
	cmd := createDepsCheckCommand()
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetArgs([]string{"CBIN-147"})
	err := cmd.Execute()

	output := buf.String()
	assert.Contains(t, output, "✔ external ENG-12 (peer:upstream-repo)")
	assert.NoError(t, err, "peer-satisfied external dependency must not fail deps check")
}

// seedLocalToken migrates a .canary/canary.db under the current working
// directory (expected to already be a depsTestRoot) and upserts a single
// local CANARY token for reqID with the given status. Used to prove that a
// dependency id with real local tokens uses local status even when it also
// matches a configured external (ticket-source) prefix.
func seedLocalToken(t *testing.T, reqID, status string) {
	t.Helper()
	dbPath := filepath.Join(".canary", "canary.db")
	require.NoError(t, storage.MigrateDB(dbPath, "all"))
	db, err := storage.Open(dbPath)
	require.NoError(t, err)
	defer func() { _ = db.Close() }()
	require.NoError(t, db.UpsertToken(&storage.Token{
		ReqID: reqID, Feature: "Upstream", Aspect: "API", Status: status,
		FilePath: "upstream.go", LineNumber: 1,
	}))
}

// TestCANARY_ENG_3960_DepsValidate_DefaultUnsatisfiedExitsZero proves `deps
// validate` is a REPORTING gate, not a blocking gate: an unsatisfied
// external dependency shows up in the "external:" summary line but does not
// fail validate in default (non-strict) mode. Blocking on external
// dependency status is next's and deps check's job; only --strict-external
// makes validate itself fail. See the comment above this check in
// createDepsValidateCommand.
func TestCANARY_ENG_3960_DepsValidate_DefaultUnsatisfiedExitsZero(t *testing.T) {
	root := depsTestRoot(t, "ENG-13 (upstream)")
	require.NoError(t, external.SaveCache(root, map[string]string{"ENG-13": "In Progress"}, mustFreshTime()))

	var buf bytes.Buffer
	cmd := createDepsValidateCommand()
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetArgs([]string{})
	err := cmd.Execute()

	assert.NoError(t, err, "default (non-strict) validate must exit 0 even with an unsatisfied external dependency")
	assert.Contains(t, buf.String(), "external: satisfied=0 unsatisfied=1 unknown=0")
}

// TestCANARY_ENG_3960_DepsValidate_LocalTokensWinOverExternalCache proves
// countExternalDeps (deps validate's summary line) excludes a dependency id
// that has real local CANARY tokens, even when the id also matches an
// external source's key and the cache reports an unsatisfied status for it.
// The id must not be double counted as "external unsatisfied" -- it's a
// local requirement.
func TestCANARY_ENG_3960_DepsValidate_LocalTokensWinOverExternalCache(t *testing.T) {
	root := depsTestRoot(t, "ENG-12 (upstream)")
	seedLocalToken(t, "ENG-12", "IMPL")
	require.NoError(t, external.SaveCache(root, map[string]string{"ENG-12": "In Progress"}, mustFreshTime()))

	var buf bytes.Buffer
	cmd := createDepsValidateCommand()
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetArgs([]string{})
	_ = cmd.Execute()

	assert.Contains(t, buf.String(), "external: satisfied=0 unsatisfied=0 unknown=0")
}

// TestCANARY_ENG_3960_DepsGraph_MermaidExternalClass_LocalTokensWin proves
// the `deps graph --format mermaid` isExternal closure never styles a node
// as external when it has real local CANARY tokens, even though its id also
// matches the configured "eng" jira source's key.
func TestCANARY_ENG_3960_DepsGraph_MermaidExternalClass_LocalTokensWin(t *testing.T) {
	depsTestRoot(t, "ENG-12 (upstream)")
	seedLocalToken(t, "ENG-12", "TESTED")

	var buf bytes.Buffer
	cmd := createDepsGraphCommand()
	cmd.SetOut(&buf)
	cmd.SetArgs([]string{"CBIN-147", "--format", "mermaid"})
	require.NoError(t, cmd.Execute())

	output := buf.String()
	assert.NotContains(t, output, "classDef external")
	assert.NotContains(t, output, "class ENG_12 external")
}

// TestCANARY_ENG_3960_DepsValidate_ExternalWithLocalTokens_CountedMissing
// proves externalAwareSpecFinder's local-tokens gate matches
// countExternalDeps / the mermaid isExternal closure: an id whose prefix
// matches a configured external source (ENG) but that ALSO has real local
// CANARY tokens and no spec dir under .canary/specs is a genuine missing
// requirement, not silently treated as external and skipped. Before
// threading the token provider into externalAwareSpecFinder, this id was
// exempted from "missing spec" reporting purely because its prefix matched
// the "eng" source's key.
func TestCANARY_ENG_3960_DepsValidate_ExternalWithLocalTokens_CountedMissing(t *testing.T) {
	depsTestRoot(t, "ENG-12 (upstream)")
	seedLocalToken(t, "ENG-12", "IMPL")

	var buf bytes.Buffer
	cmd := createDepsValidateCommand()
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetArgs([]string{})
	err := cmd.Execute()

	output := buf.String()
	assert.Contains(t, output, "Missing requirement: ENG-12")
	assert.Error(t, err, "an external-prefix id with local tokens and no spec dir must fail validate as a real missing-spec finding")
}
