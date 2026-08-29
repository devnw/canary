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
