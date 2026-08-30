// Copyright (c) 2025 by Developer Network.
//
// For more details, see the LICENSE file in the root directory of this
// source code repository or contact Developer Network at info@devnw.com.

package mcp

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"devnw.dev/canary/pkg/safewrite"
	"devnw.dev/canary/pkg/storage"
)

// TestRegistryHasNoStubs proves nothing advertises itself as unimplemented,
// and that the five placeholder tools are gone rather than merely hidden from
// the banner.
func TestRegistryHasNoStubs(t *testing.T) {
	names := map[string]bool{}
	for _, s := range Registry() {
		if s.Name == "" || s.Description == "" || s.Register == nil {
			t.Fatalf("incomplete tool spec: %+v", s)
		}
		if names[s.Name] {
			t.Fatalf("duplicate tool %s", s.Name)
		}
		names[s.Name] = true

		lower := strings.ToLower(s.Description)
		for _, weasel := range []string{"stub", "pending", "not yet implemented"} {
			if strings.Contains(lower, weasel) {
				t.Errorf("tool %s advertises itself as unimplemented: %q", s.Name, s.Description)
			}
		}
	}
	for _, gone := range []string{"specify", "plan", "index", "gap-mark"} {
		if names[gone] {
			t.Errorf("placeholder tool %s is still registered", gone)
		}
	}
	for _, required := range []string{"list", "show", "create", "status", "search", "next",
		"view", "deps", "scan", "implement", "files", "grep", "prioritize", "bug-list", "bug-create"} {
		if !names[required] {
			t.Errorf("tool %s is missing from the registry", required)
		}
	}
}

// TestMutatingToolsMatchRegistry proves the authorization scope set is
// derived from the registry, not maintained beside it.
func TestMutatingToolsMatchRegistry(t *testing.T) {
	mutating := mutatingTools()
	for _, s := range Registry() {
		if s.Mutates != mutating[s.Name] {
			t.Errorf("tool %s: Mutates=%v but scope set says %v", s.Name, s.Mutates, mutating[s.Name])
		}
	}
	for _, want := range []string{"prioritize", "bug-create"} {
		if !mutating[want] {
			t.Errorf("%s must be scoped as mutating", want)
		}
	}
}

// TestRenderDocsCoversRegistry proves the generated documentation names every
// registered tool and marks the mutating ones.
func TestRenderDocsCoversRegistry(t *testing.T) {
	docs := RenderDocs()
	if !strings.HasPrefix(docs, "# Canary MCP Tools\n") {
		t.Fatalf("docs do not start with the expected title: %.40q", docs)
	}
	for _, s := range Registry() {
		if !strings.Contains(docs, "### `"+s.Name+"`") {
			t.Errorf("docs do not document tool %s", s.Name)
		}
		if !strings.Contains(docs, s.Description) {
			t.Errorf("docs do not carry %s's description", s.Name)
		}
	}
	if !strings.Contains(docs, "`prioritize` | mutate") {
		t.Error("docs do not mark prioritize as mutating")
	}
	// Deterministic: two renders of the same registry are byte-identical, or
	// the generated file would churn on every regeneration.
	if RenderDocs() != docs {
		t.Error("RenderDocs is not deterministic")
	}
}

// TestDepsResolveDefaults proves a zero Deps names the working directory and
// the conventional index.
func TestDepsResolveDefaults(t *testing.T) {
	d := Deps{}.resolve()
	if d.Root != "." {
		t.Errorf("Root = %q, want \".\"", d.Root)
	}
	if d.DBPath != filepath.Join(".", filepath.FromSlash(DefaultDBPath)) {
		t.Errorf("DBPath = %q, want the conventional index", d.DBPath)
	}
}

// TestConfineRefusesEscape proves a caller-supplied path is resolved against
// the server root -- not the process working directory -- and refused when it
// lands outside.
func TestConfineRefusesEscape(t *testing.T) {
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, "sub"), 0o750); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	d := Deps{Root: root}

	got, err := d.confine("sub")
	if err != nil {
		t.Fatalf("confine(sub): %v", err)
	}
	if !strings.HasSuffix(got, "sub") {
		t.Errorf("confine(sub) = %q", got)
	}

	// An empty path means the root itself.
	if _, err := d.confine(""); err != nil {
		t.Errorf("confine(\"\"): %v", err)
	}

	for _, escape := range []string{"../../..", "..", "/etc", filepath.Join(root, "..", "elsewhere")} {
		_, err := d.confine(escape)
		if err == nil {
			t.Errorf("confine(%q) was allowed", escape)
			continue
		}
		if !strings.Contains(err.Error(), "ROOT_ESCAPE") {
			t.Errorf("confine(%q) error does not carry ROOT_ESCAPE: %v", escape, err)
		}
		if !strings.Contains(err.Error(), safewrite.ErrRootEscape.Error()) {
			t.Errorf("confine(%q) does not wrap safewrite.ErrRootEscape: %v", escape, err)
		}
	}
}

// TestBugCreatePersistsReservedID proves the bug-create tool writes the row
// it reports. The handler it replaces returned the literal "BUG-001" every
// time and persisted nothing, so two calls collided and neither produced a
// bug.
func TestBugCreatePersistsReservedID(t *testing.T) {
	root := t.TempDir()
	d := Deps{Root: root}
	ctx := context.Background()

	_, first, err := d.handleBugCreate(ctx, &mcp.CallToolRequest{}, &BugCreateParams{
		Title: "first bug", Aspect: "api", Severity: "S1", Priority: "P0",
	})
	if err != nil {
		t.Fatalf("handleBugCreate: %v", err)
	}
	if first.BugID != "BUG-API-001" {
		t.Fatalf("BugID = %q, want BUG-API-001 (aspect is normalized, id is reserved)", first.BugID)
	}

	_, second, err := d.handleBugCreate(ctx, &mcp.CallToolRequest{}, &BugCreateParams{
		Title: "second bug", Aspect: "API",
	})
	if err != nil {
		t.Fatalf("handleBugCreate: %v", err)
	}
	if second.BugID == first.BugID {
		t.Fatalf("two bugs share id %s", first.BugID)
	}

	db, err := storage.OpenRO(d.db())
	if err != nil {
		t.Fatalf("OpenRO: %v", err)
	}
	defer func() { _ = db.Close() }()

	for _, want := range []struct{ id, title string }{
		{first.BugID, "first bug"},
		{second.BugID, "second bug"},
	} {
		rows, err := db.GetTokensByReqID(storage.DefaultProjectID, want.id)
		if err != nil {
			t.Fatalf("read %s: %v", want.id, err)
		}
		if len(rows) != 1 {
			t.Fatalf("%s: got %d rows, want 1", want.id, len(rows))
		}
		if rows[0].Feature != want.title || rows[0].Status != "OPEN" {
			t.Errorf("%s persisted as %+v", want.id, rows[0])
		}
	}

	// A bug file path is confined like any other caller-supplied path.
	if _, _, err := d.handleBugCreate(ctx, &mcp.CallToolRequest{}, &BugCreateParams{
		Title: "escaping bug", File: "../../../etc/passwd:1",
	}); err == nil || !strings.Contains(err.Error(), "ROOT_ESCAPE") {
		t.Fatalf("bug file outside the root was accepted: %v", err)
	}
}

// TestToolErrExplainsProjectRequired proves the storage refusal reaches an
// agent as an actionable message that still carries the contract token.
func TestToolErrExplainsProjectRequired(t *testing.T) {
	err := toolErr("get tokens", storage.ErrProjectRequired)
	if !strings.Contains(err.Error(), "PROJECT_REQUIRED") {
		t.Fatalf("error lost the contract token: %v", err)
	}
	if !strings.Contains(err.Error(), "--project") {
		t.Fatalf("error does not say what to do: %v", err)
	}
}
