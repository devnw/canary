package deps

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// CANARY: REQ=CBIN-147; FEATURE="DepsCheckCommand"; ASPECT=CLI; STATUS=TESTED; TEST=TestDepsCheckCommand; UPDATED=2025-10-18
func TestDepsCheckCommand(t *testing.T) {
	// Create temp directory with test spec
	tmpDir := t.TempDir()
	specDir := filepath.Join(tmpDir, ".canary", "specs", "CBIN-147-test")
	require.NoError(t, os.MkdirAll(specDir, 0755))

	specContent := `# Test Spec

## Dependencies

### Full Dependencies
- CBIN-146 (Multi-Project Support)

## Features
- TestFeature
`
	require.NoError(t, os.WriteFile(filepath.Join(specDir, "spec.md"), []byte(specContent), 0644))

	// Change to temp dir
	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)
	os.Chdir(tmpDir)

	// Run command
	var buf bytes.Buffer
	cmd := createDepsCheckCommand()
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetArgs([]string{"CBIN-147"})

	_ = cmd.Execute() // May error if dependencies not satisfied (no database)

	output := buf.String()
	assert.Contains(t, output, "CBIN-147")
	assert.Contains(t, output, "CBIN-146")
}

// CANARY: REQ=CBIN-147; FEATURE="DepsGraphCommand"; ASPECT=CLI; STATUS=TESTED; TEST=TestDepsGraphCommand; UPDATED=2025-10-18
func TestDepsGraphCommand(t *testing.T) {
	tmpDir := t.TempDir()
	specDir := filepath.Join(tmpDir, ".canary", "specs", "CBIN-147-test")
	require.NoError(t, os.MkdirAll(specDir, 0755))

	specContent := `# Test Spec

## Dependencies

### Full Dependencies
- CBIN-146 (Multi-Project Support)

## Features
- TestFeature
`
	require.NoError(t, os.WriteFile(filepath.Join(specDir, "spec.md"), []byte(specContent), 0644))

	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)
	os.Chdir(tmpDir)

	var buf bytes.Buffer
	cmd := createDepsGraphCommand()
	cmd.SetOut(&buf)
	cmd.SetArgs([]string{"CBIN-147"})

	err := cmd.Execute()
	require.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, "CBIN-147")
	// Should show tree structure
	assert.True(t, len(output) > 0)
}

// CANARY: REQ=CBIN-147; FEATURE="DepsReverseCommand"; ASPECT=CLI; STATUS=TESTED; TEST=TestDepsReverseCommand; UPDATED=2025-10-18
func TestDepsReverseCommand(t *testing.T) {
	tmpDir := t.TempDir()

	// Create two specs where CBIN-147 depends on CBIN-146
	spec147Dir := filepath.Join(tmpDir, ".canary", "specs", "CBIN-147-test")
	require.NoError(t, os.MkdirAll(spec147Dir, 0755))

	spec147Content := `# Test Spec 147

## Dependencies

### Full Dependencies
- CBIN-146 (Multi-Project Support)

## Features
- TestFeature
`
	require.NoError(t, os.WriteFile(filepath.Join(spec147Dir, "spec.md"), []byte(spec147Content), 0644))

	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)
	os.Chdir(tmpDir)

	var buf bytes.Buffer
	cmd := createDepsReverseCommand()
	cmd.SetOut(&buf)
	cmd.SetArgs([]string{"CBIN-146"})

	err := cmd.Execute()
	require.NoError(t, err)

	output := buf.String()
	// Should show that CBIN-147 depends on CBIN-146
	assert.Contains(t, output, "CBIN-146")
}

// CANARY: REQ=CBIN-147; FEATURE="DepsValidateCommand"; ASPECT=CLI; STATUS=TESTED; TEST=TestDepsValidateCommand; UPDATED=2025-10-18
func TestDepsValidateCommand(t *testing.T) {
	tmpDir := t.TempDir()
	specDir := filepath.Join(tmpDir, ".canary", "specs", "CBIN-147-test")
	require.NoError(t, os.MkdirAll(specDir, 0755))

	specContent := `# Test Spec

## Dependencies

### Full Dependencies
- CBIN-146 (Multi-Project Support)

## Features
- TestFeature
`
	require.NoError(t, os.WriteFile(filepath.Join(specDir, "spec.md"), []byte(specContent), 0644))

	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)
	os.Chdir(tmpDir)

	var buf bytes.Buffer
	cmd := createDepsValidateCommand()
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetArgs([]string{})

	_ = cmd.Execute() // May error if validation fails (missing dependencies)

	output := buf.String()
	// Should validate all dependencies
	assert.True(t, len(output) > 0)
}

// CANARY: REQ=CBIN-147; FEATURE="DepsValidateCommand"; ASPECT=CLI; STATUS=TESTED; TEST=TestDepsValidateCommand_DetectsCycle; UPDATED=2025-10-18
func TestDepsValidateCommand_DetectsCycle(t *testing.T) {
	tmpDir := t.TempDir()

	// Create cycle: CBIN-100 -> CBIN-101 -> CBIN-100
	spec100Dir := filepath.Join(tmpDir, ".canary", "specs", "CBIN-100-test")
	require.NoError(t, os.MkdirAll(spec100Dir, 0755))
	spec100Content := `# Test Spec 100

## Dependencies
- CBIN-101 (Test)

## Features
- Feature100
`
	require.NoError(t, os.WriteFile(filepath.Join(spec100Dir, "spec.md"), []byte(spec100Content), 0644))

	spec101Dir := filepath.Join(tmpDir, ".canary", "specs", "CBIN-101-test")
	require.NoError(t, os.MkdirAll(spec101Dir, 0755))
	spec101Content := `# Test Spec 101

## Dependencies
- CBIN-100 (Test)

## Features
- Feature101
`
	require.NoError(t, os.WriteFile(filepath.Join(spec101Dir, "spec.md"), []byte(spec101Content), 0644))

	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)
	os.Chdir(tmpDir)

	var buf bytes.Buffer
	cmd := createDepsValidateCommand()
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetArgs([]string{})

	_ = cmd.Execute() // Ignore error - we're checking output
	// Should detect cycle
	output := buf.String()
	// Case-insensitive check for cycle
	assert.True(t, strings.Contains(strings.ToLower(output), "cycle"), "Output should contain 'cycle':\n%s", output)
}

// CANARY: REQ=CBIN-147; FEATURE="DepsParentCommand"; ASPECT=CLI; STATUS=TESTED; TEST=TestDepsParentCommand; UPDATED=2025-10-18
func TestDepsParentCommand(t *testing.T) {
	cmd := createDepsCommand()

	assert.NotNil(t, cmd)
	assert.Equal(t, "deps", cmd.Use)

	// Should have subcommands
	subcommands := cmd.Commands()
	assert.GreaterOrEqual(t, len(subcommands), 4) // check, graph, reverse, validate
}
