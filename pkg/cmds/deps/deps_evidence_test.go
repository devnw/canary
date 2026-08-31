// Copyright (c) 2025 by Developer Network.
//
// For more details, see the LICENSE file in the root directory of this
// source code repository or contact Developer Network at info@devnw.com.

package deps

import (
	"bytes"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"devnw.dev/canary/pkg/evidence"
)

// gitInitCommit turns root into a git repository with one empty commit so
// canaryscan.HeadCommit resolves. `deps check` binds a local dependency's
// evidence lookup to that commit, so a fixture without one is a fixture where
// no evidence can ever be current.
func gitInitCommit(t *testing.T, root string) string {
	t.Helper()
	steps := [][]string{
		{"init", "-q"},
		{"-c", "user.email=deps@example.com", "-c", "user.name=Deps", "-c", "commit.gpgsign=false", "commit", "--allow-empty", "-q", "-m", "init"},
	}
	for _, args := range steps {
		cmd := exec.Command("git", append([]string{"-C", root}, args...)...) //nolint:gosec // fixed argv
		cmd.Env = append(os.Environ(), "GIT_CONFIG_NOSYSTEM=1")
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, out)
		}
	}
	out, err := exec.Command("git", "-C", root, "rev-parse", "HEAD").Output() //nolint:gosec // fixed argv
	require.NoError(t, err)
	return strings.TrimSpace(string(out))
}

// writeDepsEvidence writes root's evidence store with the given records.
func writeDepsEvidence(t *testing.T, root string, recs ...evidence.Record) {
	t.Helper()
	path := filepath.Join(root, ".canary", "evidence.json")
	require.NoError(t, os.MkdirAll(filepath.Dir(path), 0o750))
	data, err := json.Marshal(evidence.File{SchemaVersion: 1, Records: recs})
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(path, data, 0o600))
}

// depsPassRecord builds one PASS evidence record for reqID's feature/aspect at
// commit. seedLocalToken declares Feature="Upstream", Aspect="API", so a
// record for that key is what proves the dependency.
func depsPassRecord(projectID, reqID, feature, aspect, commit string) evidence.Record {
	return evidence.Record{
		ProjectID:      projectID,
		RequirementID:  reqID,
		Feature:        feature,
		Aspect:         aspect,
		TestID:         "TestFixture",
		Command:        "go test ./...",
		Result:         "PASS",
		CommitSHA:      commit,
		ObservedAt:     "2026-01-01T00:00:00Z",
		Runner:         "local",
		ArtifactDigest: "sha256:" + strings.Repeat("ab", 32),
	}
}

// TestCANARY_F23_DepsCheck_DeclaredTestedWithoutEvidenceBlocks proves the F-23
// spec-adherence fix: a LOCAL dependency whose token declares STATUS=TESTED is
// NOT satisfied by that declaration alone. Without a current PASS record it
// blocks, exactly as `canary next` already gates it. Before this change the
// declared status was accepted as proof and the dependency counted satisfied.
func TestCANARY_F23_DepsCheck_DeclaredTestedWithoutEvidenceBlocks(t *testing.T) {
	root := depsTestRoot(t, "CBIN-146 (upstream)")
	_ = gitInitCommit(t, root)
	seedLocalToken(t, "CBIN-146", "TESTED")

	var buf bytes.Buffer
	cmd := createDepsCheckCommand()
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetArgs([]string{"CBIN-147"})
	err := cmd.Execute()

	output := buf.String()
	assert.Contains(t, output, "❌ CBIN-146", "a declared TESTED dependency with no evidence must render as blocking")
	assert.Contains(t, output, "Summary: 0 satisfied, 1 blocking")
	assert.Error(t, err, "a declared TESTED dependency with no evidence must fail deps check")
}

// TestCANARY_F23_DepsCheck_CurrentEvidenceSatisfies proves the satisfied path:
// with a PASS record for the dependency's declared feature/aspect at HEAD, the
// LOCAL dependency is satisfied and deps check exits zero.
func TestCANARY_F23_DepsCheck_CurrentEvidenceSatisfies(t *testing.T) {
	root := depsTestRoot(t, "CBIN-146 (upstream)")
	head := gitInitCommit(t, root)
	seedLocalToken(t, "CBIN-146", "TESTED")
	writeDepsEvidence(t, root, depsPassRecord("CBIN", "CBIN-146", "Upstream", "API", head))

	var buf bytes.Buffer
	cmd := createDepsCheckCommand()
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetArgs([]string{"CBIN-147", "--show-satisfied"})
	err := cmd.Execute()

	output := buf.String()
	assert.Contains(t, output, "Summary: 1 satisfied, 0 blocking")
	assert.NoError(t, err, "a dependency with current passing evidence must be satisfied")
}

// TestCANARY_F23_DepsCheck_EvidenceAtWrongCommitBlocks proves satisfaction is
// bound to HEAD: a PASS record at a different commit does not satisfy the
// dependency, so it blocks. This is the same commit-binding evidence.Complete
// enforces for verify and next.
func TestCANARY_F23_DepsCheck_EvidenceAtWrongCommitBlocks(t *testing.T) {
	root := depsTestRoot(t, "CBIN-146 (upstream)")
	_ = gitInitCommit(t, root)
	seedLocalToken(t, "CBIN-146", "TESTED")
	writeDepsEvidence(t, root, depsPassRecord("CBIN", "CBIN-146", "Upstream", "API", strings.Repeat("0", 40)))

	var buf bytes.Buffer
	cmd := createDepsCheckCommand()
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetArgs([]string{"CBIN-147"})
	err := cmd.Execute()

	output := buf.String()
	assert.Contains(t, output, "❌ CBIN-146")
	assert.Error(t, err, "evidence at a stale commit must not satisfy a dependency")
}
