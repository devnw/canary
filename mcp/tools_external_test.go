// Copyright (c) 2025 by Developer Network.
//
// For more details, see the LICENSE file in the root directory of this
// source code repository or contact Developer Network at info@devnw.com.

package mcp

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"devnw.dev/canary/pkg/external"
)

// These tests cover the MCP `next` tool's dependency behavior. It no longer
// has any dependency logic of its own: the tool delegates to
// next.SelectNext, the same selection `canary next` runs. What is asserted
// here is that the delegation is real -- that the MCP answer is the CLI
// answer, including the parts where the CLI is stricter than the replica this
// tool used to carry.

// mcpDepFixture builds a tree whose only actionable requirement (CBIN-500)
// depends on dep, with the ENG ticket source configured. Nothing is indexed:
// `next` falls back to the canonical filesystem scan, which is what the CLI
// does for a tree with no fresh index.
func mcpDepFixture(t *testing.T, dep string, extra ...string) {
	t.Helper()
	root := t.TempDir()
	t.Chdir(root)

	if err := os.MkdirAll(filepath.Join(root, ".canary"), 0o750); err != nil {
		t.Fatalf("mkdir .canary: %v", err)
	}
	projectYAML := "project:\n  name: \"demo\"\n  key: \"CBIN\"\n" +
		"sources:\n" +
		"  - name: core\n    type: flatfile\n    key: CBIN\n" +
		"  - name: eng\n    type: jira\n    key: ENG\n"
	if err := os.WriteFile(filepath.Join(root, ".canary", "project.yaml"), []byte(projectYAML), 0o600); err != nil {
		t.Fatalf("write project.yaml: %v", err)
	}

	lines := "// CANARY: REQ=CBIN-500; FEATURE=\"Consumer\"; ASPECT=API; STATUS=STUB; PRIORITY=1; DEPENDS_ON=" + dep + "; UPDATED=2026-01-01\n"
	for _, l := range extra {
		lines += l + "\n"
	}
	if err := os.WriteFile(filepath.Join(root, "consumer.go"), []byte(lines), 0o600); err != nil {
		t.Fatalf("write consumer.go: %v", err)
	}
}

// mcpNext runs the next tool against the fixture's working directory.
func mcpNext(t *testing.T) *NextResult {
	t.Helper()
	_, res, err := testDeps.handleNext(context.Background(), &mcp.CallToolRequest{}, &NextParams{})
	if err != nil {
		t.Fatalf("handleNext: %v", err)
	}
	return res
}

// TestCANARY_ENG_3960_MCP_Next_ExternalSatisfied_NotBlocking proves the tool
// does not block on an external dependency whose cached remote status is in
// the source's done-set.
func TestCANARY_ENG_3960_MCP_Next_ExternalSatisfied_NotBlocking(t *testing.T) {
	mcpDepFixture(t, "ENG-1")
	if err := external.SaveCache(".", map[string]string{"ENG-1": "Done"}, time.Now().UTC()); err != nil {
		t.Fatal(err)
	}

	if got := mcpNext(t); got.ReqID != "CBIN-500" {
		t.Errorf("satisfied external dependency must not block; got %+v", got)
	}
}

// TestCANARY_ENG_3960_MCP_Next_ExternalUnsatisfied_Blocking proves a cached
// but not-done remote status blocks selection.
func TestCANARY_ENG_3960_MCP_Next_ExternalUnsatisfied_Blocking(t *testing.T) {
	mcpDepFixture(t, "ENG-1")
	if err := external.SaveCache(".", map[string]string{"ENG-1": "In Progress"}, time.Now().UTC()); err != nil {
		t.Fatal(err)
	}

	got := mcpNext(t)
	if got.ReqID != "" {
		t.Errorf("unsatisfied external dependency must block; got %+v", got)
	}
	if got.Blocked == 0 {
		t.Error("an empty answer over blocked work must report the blocked count")
	}
}

// TestCANARY_ENG_3960_MCP_Next_ExternalUnknown_Blocking pins the behavior
// change this delegation brings: an external dependency with no cached status
// now BLOCKS, because that is what `canary next` does.
//
// The replica this tool used to carry answered the opposite -- "unknown is
// not blocking, degradation is sacred" -- for months after the CLI stopped
// agreeing. The thing being degraded was the answer to "may this work
// start?", and handing an agent a requirement whose prerequisite might not
// exist is not a degraded answer, it is a wrong one.
func TestCANARY_ENG_3960_MCP_Next_ExternalUnknown_Blocking(t *testing.T) {
	mcpDepFixture(t, "ENG-1") // no cache file written

	if got := mcpNext(t); got.ReqID != "" {
		t.Errorf("unknown external dependency must block; got %+v", got)
	}
}

// TestCANARY_ENG_3960_MCP_Next_LocalMissingDep_StillBlocking proves a
// dependency that names nothing -- neither declared here nor owned by a
// configured source -- still blocks.
func TestCANARY_ENG_3960_MCP_Next_LocalMissingDep_StillBlocking(t *testing.T) {
	mcpDepFixture(t, "CBIN-999")

	if got := mcpNext(t); got.ReqID != "" {
		t.Errorf("missing local dependency must block; got %+v", got)
	}
}

// TestCANARY_ENG_3960_MCP_Next_LocalDeclarationIsNotProof proves the tool
// applies the CLI's evidence rule to a locally declared dependency: a
// STATUS=TESTED token is a claim, and a claim with no passing evidence at
// this commit does not clear the dependency.
func TestCANARY_ENG_3960_MCP_Next_LocalDeclarationIsNotProof(t *testing.T) {
	mcpDepFixture(t, "ENG-1",
		"// CANARY: REQ=ENG-1; FEATURE=\"Upstream\"; ASPECT=API; STATUS=TESTED; UPDATED=2026-01-01")
	// Even a done-equivalent cached remote status cannot help: the id is
	// declared locally, so the local evidence rule is the one that applies.
	if err := external.SaveCache(".", map[string]string{"ENG-1": "Done"}, time.Now().UTC()); err != nil {
		t.Fatal(err)
	}

	if got := mcpNext(t); got.ReqID == "CBIN-500" {
		t.Errorf("a declared TESTED with no evidence must not satisfy a dependency; got %+v", got)
	}
}

// TestCANARY_ENG_3960_MCP_Next_UnblockedWorkIsFound proves the gate does not
// simply block everything: a requirement with no dependencies is returned,
// and the source of the answer is reported.
func TestCANARY_ENG_3960_MCP_Next_UnblockedWorkIsFound(t *testing.T) {
	root := t.TempDir()
	t.Chdir(root)
	if err := os.WriteFile(filepath.Join(root, "free.go"), []byte(
		"// CANARY: REQ=CBIN-600; FEATURE=\"Free\"; ASPECT=API; STATUS=STUB; PRIORITY=1; UPDATED=2026-01-01\n"), 0o600); err != nil {
		t.Fatalf("write free.go: %v", err)
	}

	got := mcpNext(t)
	if got.ReqID != "CBIN-600" {
		t.Fatalf("dependency-free work must be selectable; got %+v", got)
	}
	if got.Source == "" {
		t.Error("the answer must name the source it came from")
	}
}
