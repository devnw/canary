// Copyright (c) 2025 by Developer Network.
//
// For more details, see the LICENSE file in the root directory of this
// source code repository or contact Developer Network at info@devnw.com.

package audit

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	sdk "github.com/modelcontextprotocol/go-sdk/mcp"

	mcpsrv "devnw.dev/canary/mcp"
	"devnw.dev/canary/pkg/storage"
)

// toolResult is one tool call's outcome, flattened so a test can assert on
// the error text whether the failure arrived as a tool error (IsError with
// the message in Content) or as a protocol error.
type toolResult struct {
	Text    string
	ErrText string
	Output  map[string]any
}

// callTool invokes one registered tool in-process over the SDK's in-memory
// transport. Nothing is bound and nothing is served: this exercises exactly
// the registration path the HTTP server uses, without a port.
func callTool(t *testing.T, root, name string, args map[string]any) toolResult {
	t.Helper()
	ctx := context.Background()

	srv := mcpsrv.NewServer("audit-test", mcpsrv.Deps{
		Root:   root,
		DBPath: dbPathIn(root),
	})
	clientT, serverT := sdk.NewInMemoryTransports()
	ss, err := srv.Connect(ctx, serverT, nil)
	if err != nil {
		t.Fatalf("connect server: %v", err)
	}
	defer func() { _ = ss.Close() }()

	client := sdk.NewClient(&sdk.Implementation{Name: "audit", Version: "test"}, nil)
	cs, err := client.Connect(ctx, clientT, nil)
	if err != nil {
		t.Fatalf("connect client: %v", err)
	}
	defer func() { _ = cs.Close() }()

	res, err := cs.CallTool(ctx, &sdk.CallToolParams{Name: name, Arguments: args})
	if err != nil {
		return toolResult{ErrText: err.Error()}
	}

	out := toolResult{}
	var text strings.Builder
	for _, c := range res.Content {
		if tc, ok := c.(*sdk.TextContent); ok {
			text.WriteString(tc.Text)
		}
	}
	if res.IsError {
		out.ErrText = text.String()
	} else {
		out.Text = text.String()
	}
	if m, ok := res.StructuredContent.(map[string]any); ok {
		out.Output = m
	}
	return out
}

// TestAuditF10 proves every tool the MCP server registers does something a
// caller can verify, that the documentation is generated from that same
// registration table rather than maintained by hand beside it, and that a
// path-taking tool cannot reach outside the root the server was started with.
func TestAuditF10(t *testing.T) {
	specs := mcpsrv.Registry()
	if len(specs) == 0 {
		t.Fatal("registry is empty")
	}

	seen := map[string]bool{}
	for _, s := range specs {
		if s.Register == nil {
			t.Fatalf("tool %s has no registration", s.Name)
		}
		if s.Name == "" || s.Description == "" {
			t.Fatalf("tool %+v is missing a name or description", s)
		}
		if seen[s.Name] {
			t.Fatalf("tool %s is registered twice", s.Name)
		}
		seen[s.Name] = true

		lower := strings.ToLower(s.Description)
		for _, weasel := range []string{"stub", "pending", "not yet implemented", "todo"} {
			if strings.Contains(lower, weasel) {
				t.Errorf("tool %s advertises itself as unimplemented (%q): unregister it instead", s.Name, weasel)
			}
		}
	}

	// The five placeholder tools are gone, not merely undocumented.
	for _, gone := range []string{"specify", "plan", "index", "gap-mark"} {
		if seen[gone] {
			t.Errorf("placeholder tool %s is still registered", gone)
		}
	}

	// The documentation is the registry, rendered.
	docPath := filepath.Join(repoRoot(), "docs", "MCP_TOOLS.md")
	doc, err := os.ReadFile(docPath) //nolint:gosec // repo-relative test fixture
	if err != nil {
		t.Fatalf("read %s: %v", docPath, err)
	}
	if string(doc) != mcpsrv.RenderDocs() {
		t.Fatalf("docs/MCP_TOOLS.md is not generated from the registry; regenerate with 'canary mcp --print-tools > docs/MCP_TOOLS.md'")
	}
	for _, s := range specs {
		if !strings.Contains(string(doc), s.Name) {
			t.Errorf("docs/MCP_TOOLS.md does not mention registered tool %s", s.Name)
		}
	}
}

// TestAuditF10RootEscape proves a tool-supplied path is confined to the root
// the server was started with. Before this, `scan` walked wherever it was
// pointed: an agent could read requirement tokens out of any tree the server
// process could reach.
func TestAuditF10RootEscape(t *testing.T) {
	root := t.TempDir()

	res := callTool(t, root, "scan", map[string]any{"root": "../../.."})
	if !strings.Contains(res.ErrText, "ROOT_ESCAPE") {
		t.Fatalf("escape not refused: %+v", res)
	}

	res = callTool(t, root, "implement", map[string]any{"reqId": "../../../etc/passwd"})
	if res.ErrText == "" {
		t.Fatalf("implement accepted a path-shaped requirement id: %+v", res)
	}
}

// TestAuditF10MutatingPostconditions proves the two tools that claim to
// change state actually change it. A tool whose only evidence of success is
// its own success message is indistinguishable from a stub.
func TestAuditF10MutatingPostconditions(t *testing.T) {
	root := t.TempDir()
	dbPath := dbPathIn(root)

	db := openDBAt(t, dbPath)
	seedToken(t, db, storage.DefaultProjectID, "CBIN-700")
	if err := db.Close(); err != nil {
		t.Fatalf("close: %v", err)
	}

	// prioritize: the row's priority actually moves.
	res := callTool(t, root, "prioritize", map[string]any{"reqId": "CBIN-700", "priority": 3})
	if res.ErrText != "" {
		t.Fatalf("prioritize: %s", res.ErrText)
	}
	after := openDBAt(t, dbPath)
	toks, err := after.GetTokensByReqID(storage.DefaultProjectID, "CBIN-700")
	if err != nil {
		t.Fatalf("read back: %v", err)
	}
	if len(toks) == 0 {
		t.Fatal("token vanished")
	}
	for _, tok := range toks {
		if tok.Priority != 3 {
			t.Fatalf("priority = %d, want 3 (prioritize did not persist)", tok.Priority)
		}
	}
	if err := after.Close(); err != nil {
		t.Fatalf("close: %v", err)
	}

	// bug-create: the id it returns names a row that exists.
	res = callTool(t, root, "bug-create", map[string]any{"title": "audit bug", "aspect": "API"})
	if res.ErrText != "" {
		t.Fatalf("bug-create: %s", res.ErrText)
	}
	bugID, _ := res.Output["bugId"].(string)
	if bugID == "" {
		t.Fatalf("bug-create returned no id: %+v", res)
	}
	if bugID == "BUG-001" {
		t.Fatal("bug-create still returns the hardcoded placeholder id")
	}

	final := openDBAt(t, dbPath)
	defer func() { _ = final.Close() }()
	rows, err := final.GetTokensByReqID(storage.DefaultProjectID, bugID)
	if err != nil {
		t.Fatalf("read back %s: %v", bugID, err)
	}
	if len(rows) == 0 {
		t.Fatalf("bug-create returned %s but persisted no row", bugID)
	}
	if rows[0].Feature != "audit bug" {
		t.Fatalf("persisted feature = %q, want %q", rows[0].Feature, "audit bug")
	}

	// A second creation must not reuse the id.
	res = callTool(t, root, "bug-create", map[string]any{"title": "second bug", "aspect": "API"})
	if res.ErrText != "" {
		t.Fatalf("bug-create (second): %s", res.ErrText)
	}
	if second, _ := res.Output["bugId"].(string); second == bugID {
		t.Fatalf("two bugs share id %s", bugID)
	}
}
