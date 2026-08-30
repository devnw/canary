// Copyright (c) 2025 by Developer Network.
//
// For more details, see the LICENSE file in the root directory of this
// source code repository or contact Developer Network at info@devnw.com.

package mcp

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"devnw.dev/canary/pkg/cmds/view"
	"devnw.dev/canary/pkg/config"
	"devnw.dev/canary/pkg/safewrite"
)

// CANARY: REQ=CP-281; FEATURE="MCPToolRegistry"; ASPECT=API; STATUS=TESTED; TEST=TestAuditF10,TestRegistryHasNoStubs,TestRenderDocsCoversRegistry; UPDATED=2026-08-30

// DefaultDBPath is the index every Canary command defaults to, relative to
// the server root.
const DefaultDBPath = ".canary/canary.db"

// Deps are the resolved inputs every tool handler runs against: the tree the
// server was started on and the index inside it.
//
// They exist so a tool answers questions about the server's root rather than
// about whatever directory the process happens to be in. A path a caller
// supplies is resolved against Root and confined to it (see confine), which
// is what keeps `scan` from being pointed at an unrelated tree.
type Deps struct {
	// Root is the tree this server answers for. It is resolved to an
	// absolute path at startup; an empty value means the working directory.
	Root string
	// DBPath is the token index. Empty means DefaultDBPath under Root.
	DBPath string
}

// resolve fills in the defaults, so a zero Deps is usable and every handler
// sees the same values.
func (d Deps) resolve() Deps {
	if strings.TrimSpace(d.Root) == "" {
		d.Root = "."
	}
	if strings.TrimSpace(d.DBPath) == "" {
		d.DBPath = filepath.Join(d.Root, filepath.FromSlash(DefaultDBPath))
	}
	return d
}

// db is the index path handlers open.
func (d Deps) db() string { return d.resolve().DBPath }

// root is the tree handlers scan and confine paths to.
func (d Deps) root() string { return d.resolve().Root }

// projectID is the project a READ scopes to: unscoped, so storage answers
// from whatever the database holds and refuses when that would be ambiguous.
// It mirrors the CLI's read-side rule exactly.
func (d Deps) projectID() string { return "" }

// writeProjectID is the project a WRITE lands in.
//
// The read rule above does not carry over: an unscoped UPDATE is not a
// broader question, it is an edit to every project sharing the database, and
// storage refuses "" outright. This is the CLI's WriteProjectID -- the
// configured project.key, falling back to config's "default", the same value
// migration 000007 backfills onto pre-scoping rows -- read from the server
// root rather than the process's working directory.
func (d Deps) writeProjectID() string {
	cfg, err := config.Load(d.root())
	if err != nil {
		// An unreadable project.yaml is not a reason to widen the write.
		cfg = nil
	}
	return cfg.ProjectID()
}

// rootEscape is the message a tool returns when a supplied path resolves
// outside the server root. It is matched by contract, so it is a constant.
const rootEscape = "ROOT_ESCAPE: path resolves outside configured root"

// confine resolves a caller-supplied path against the server root and refuses
// anything that lands outside it.
//
// A relative path is joined to Root, never to the process working directory:
// the server's answer must not depend on where it was launched from, and
// "../.." must mean "above the configured root" (refused) rather than "above
// wherever this process happens to be".
func (d Deps) confine(path string) (string, error) {
	root := d.root()
	if strings.TrimSpace(path) == "" {
		path = root
	}
	if !filepath.IsAbs(path) {
		path = filepath.Join(root, path)
	}
	resolved, err := safewrite.Confine(root, path)
	if err != nil {
		return "", fmt.Errorf("%s: %w", rootEscape, err)
	}
	return resolved, nil
}

// ToolSpec is one MCP tool: its identity, whether calling it changes stored
// state, and how it is registered.
//
// Registry is the only place tools are declared. The server, the
// authorization scope check and docs/MCP_TOOLS.md are all derived from it, so
// a tool cannot be exposed without also being documented and scoped -- the
// failure mode of the previous design, where nineteen imperative AddTool
// calls sat beside a hand-maintained forty-line banner that had drifted from
// them.
type ToolSpec struct {
	Name        string
	Description string
	// Mutates reports whether the tool writes. A read-scoped bearer token is
	// refused on these; everything else is readable with either token.
	Mutates  bool
	Register func(s *mcp.Server, deps Deps)
}

// spec builds a ToolSpec whose registration uses the spec's own name and
// description, so the documented tool and the registered tool cannot differ.
func spec[In, Out any](name, description string, mutates bool, bind func(Deps) mcp.ToolHandlerFor[In, Out]) ToolSpec {
	return ToolSpec{
		Name:        name,
		Description: description,
		Mutates:     mutates,
		Register: func(s *mcp.Server, deps Deps) {
			mcp.AddTool(s, &mcp.Tool{Name: name, Description: description}, bind(deps))
		},
	}
}

// Registry returns every tool this server exposes, in presentation order.
//
// Every entry here has a postcondition a caller can check. The five
// placeholder tools that used to be registered -- specify, plan, index,
// bug-create and gap-mark -- returned a success message and changed nothing;
// an agent could not tell them from tools that worked. They are unregistered,
// and bug-create is back as a real tool that persists a row.
func Registry() []ToolSpec {
	return []ToolSpec{
		// Core token management.
		spec("list",
			"List CANARY tokens with optional filtering (default limit 20, max 100 to reduce context; Total reports the true match count)",
			false,
			func(d Deps) mcp.ToolHandlerFor[*ListParams, *ListResult] { return d.handleList }),
		spec("show",
			"Display all CANARY tokens for a specific requirement ID",
			false,
			func(d Deps) mcp.ToolHandlerFor[*ShowParams, *ShowResult] { return d.handleShow }),
		spec("create",
			"Generate a new CANARY token template (returns the token text to paste into source; writes nothing)",
			false,
			func(d Deps) mcp.ToolHandlerFor[*CreateParams, *CreateResult] { return d.handleCreate }),
		spec("status",
			"Show implementation progress for a requirement",
			false,
			func(d Deps) mcp.ToolHandlerFor[*StatusParams, *StatusResult] { return d.handleStatus }),
		spec("search",
			"Search CANARY tokens by keywords",
			false,
			func(d Deps) mcp.ToolHandlerFor[*SearchParams, *SearchResult] { return d.handleSearch }),
		spec("next",
			"Identify the next highest-priority actionable requirement, applying the same dependency rule as the `canary next` CLI: a local dependency is complete only when evidence at the current commit proves it, and an external (ticket-source) dependency whose state cannot be resolved blocks selection.",
			false,
			func(d Deps) mcp.ToolHandlerFor[*NextParams, *NextResult] { return d.handleNext }),

		// One-call hierarchical context.
		spec("view",
			"Full picture of one requirement: status, files, tests, deps, spec/plan, diagrams, ticket URL. Use this FIRST instead of separate show/status/files calls.",
			false,
			func(d Deps) mcp.ToolHandlerFor[*ViewParams, *view.View] { return d.handleView }),
		spec("deps",
			"Dependency IDs for a requirement (forward or reverse). IDs only; follow up with view for detail.",
			false,
			func(d Deps) mcp.ToolHandlerFor[*DepsParams, *DepsResult] { return d.handleDeps }),

		// Workflow.
		spec("scan",
			"Scan the server root for CANARY tokens, honoring .canaryignore and the configured ticket sources. Paths outside the root are refused.",
			false,
			func(d Deps) mcp.ToolHandlerFor[*ScanParams, *ScanResult] { return d.handleScan }),
		spec("implement",
			"Report implementation state for a requirement: token count, current phase, and whether its spec and plan exist",
			false,
			func(d Deps) mcp.ToolHandlerFor[*ImplementParams, *ImplementResult] { return d.handleImplement }),

		// Query and navigation.
		spec("files",
			"Find files containing tokens for a requirement",
			false,
			func(d Deps) mcp.ToolHandlerFor[*FilesParams, *FilesResult] { return d.handleFiles }),
		spec("grep",
			"Search tokens by pattern in specific fields",
			false,
			func(d Deps) mcp.ToolHandlerFor[*GrepParams, *GrepResult] { return d.handleGrep }),

		// Management (mutating).
		spec("prioritize",
			"Set the priority level for a requirement (writes to the index)",
			true,
			func(d Deps) mcp.ToolHandlerFor[*PrioritizeParams, *PrioritizeResult] { return d.handlePrioritize }),

		// Bug tracking.
		spec("bug-list",
			"List bug tracking tokens",
			false,
			func(d Deps) mcp.ToolHandlerFor[*BugListParams, *BugListResult] { return d.handleBugList }),
		spec("bug-create",
			"Create a bug tracking token with a transactionally reserved BUG-<ASPECT>-NNN id and persist it to the index",
			true,
			func(d Deps) mcp.ToolHandlerFor[*BugCreateParams, *BugCreateResult] { return d.handleBugCreate }),
	}
}

// mutatingTools is the set of tool names a read-scoped caller may not invoke,
// derived from the registry so a new mutating tool is scoped the moment it is
// declared.
func mutatingTools() map[string]bool {
	m := make(map[string]bool)
	for _, s := range Registry() {
		if s.Mutates {
			m[s.Name] = true
		}
	}
	return m
}

// NewServer builds the MCP server with every registered tool bound to deps.
func NewServer(version string, deps Deps) *mcp.Server {
	if strings.TrimSpace(version) == "" {
		version = "dev"
	}
	deps = deps.resolve()
	server := mcp.NewServer(&mcp.Implementation{
		Name:    "canary-server",
		Version: version,
	}, nil)
	for _, s := range Registry() {
		s.Register(server, deps)
	}
	return server
}

// RenderDocs returns docs/MCP_TOOLS.md, generated from Registry.
//
// The file on disk is asserted to equal this string, so the documentation
// cannot describe a tool set the server does not have.
func RenderDocs() string {
	specs := Registry()

	var b strings.Builder
	b.WriteString("# Canary MCP Tools\n\n")
	b.WriteString("<!-- GENERATED FILE. Do not edit by hand.\n")
	b.WriteString("     Regenerate with: canary mcp --print-tools > docs/MCP_TOOLS.md -->\n\n")
	fmt.Fprintf(&b, "The Canary MCP server exposes %d tools. Start it with `canary mcp`; it binds\n", len(specs))
	b.WriteString("127.0.0.1:8080 by default and serves the MCP endpoint at `/mcp` and a health\n")
	b.WriteString("check at `/health`.\n\n")
	b.WriteString("Every tool below has a checkable postcondition. Tools marked **mutate** write\n")
	b.WriteString("to the index; when authentication is configured they require the read+write\n")
	b.WriteString("token (`CANARY_MCP_TOKEN`) and are refused to a read-only token\n")
	b.WriteString("(`CANARY_MCP_READ_TOKEN`) with 403.\n\n")

	b.WriteString("## Summary\n\n")
	b.WriteString("| Tool | Scope |\n")
	b.WriteString("| --- | --- |\n")
	for _, s := range specs {
		fmt.Fprintf(&b, "| `%s` | %s |\n", s.Name, scopeLabel(s.Mutates))
	}
	b.WriteString("\n## Tools\n\n")
	for _, s := range specs {
		fmt.Fprintf(&b, "### `%s` (%s)\n\n%s\n\n", s.Name, scopeLabel(s.Mutates), s.Description)
	}
	// Exactly one trailing newline. The generated file is committed, and the
	// repository's end-of-file hook rewrites anything else -- which would
	// break the byte-equality check that keeps the file honest.
	return strings.TrimRight(b.String(), "\n") + "\n"
}

func scopeLabel(mutates bool) string {
	if mutates {
		return "mutate"
	}
	return "read"
}
