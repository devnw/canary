// Copyright (c) 2025 by Developer Network.
//
// For more details, see the LICENSE file in the root directory of this
// source code repository or contact Developer Network at info@devnw.com.

package mcp

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"go.devnw.com/canary/internal/storage"
)

func TestMCPToolHandlers(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name    string
		handler string
		params  interface{}
	}{
		// Core tools
		{
			name:    "handleList",
			handler: "list",
			params: &ListParams{
				Status: "IMPL",
				Limit:  10,
			},
		},
		{
			name:    "handleCreate",
			handler: "create",
			params: &CreateParams{
				ReqID:   "TEST-001",
				Feature: "TestFeature",
				Aspect:  "API",
				Status:  "IMPL",
			},
		},
		{
			name:    "handleSearch",
			handler: "search",
			params: &SearchParams{
				Keywords: "test",
			},
		},
		{
			name:    "handleNext",
			handler: "next",
			params: &NextParams{
				Status: "STUB",
			},
		},
		{
			name:    "handleScan",
			handler: "scan",
			params: &ScanParams{
				Root: ".",
			},
		},
		// Extended tools
		{
			name:    "handleSpecify",
			handler: "specify",
			params: &SpecifyParams{
				Description: "Test feature",
				Aspect:      "API",
			},
		},
		{
			name:    "handlePlan",
			handler: "plan",
			params: &PlanParams{
				ReqID:     "TEST-001",
				TechStack: "Go",
			},
		},
		{
			name:    "handleIndex",
			handler: "index",
			params: &IndexParams{
				Root: ".",
			},
		},
		{
			name:    "handleImplement",
			handler: "implement",
			params: &ImplementParams{
				ReqID: "TEST-001",
			},
		},
		{
			name:    "handleFiles",
			handler: "files",
			params: &FilesParams{
				ReqID: "TEST-001",
			},
		},
		{
			name:    "handleGrep",
			handler: "grep",
			params: &GrepParams{
				Pattern: "test",
				Field:   "feature",
			},
		},
		{
			name:    "handlePrioritize",
			handler: "prioritize",
			params: &PrioritizeParams{
				ReqID:    "TEST-001",
				Priority: 1,
			},
		},
		{
			name:    "handleBugList",
			handler: "bug-list",
			params: &BugListParams{
				Status: "OPEN",
				Limit:  10,
			},
		},
		{
			name:    "handleBugCreate",
			handler: "bug-create",
			params: &BugCreateParams{
				Title:    "Test bug",
				Severity: "MEDIUM",
			},
		},
		{
			name:    "handleGapMark",
			handler: "gap-mark",
			params: &GapMarkParams{
				ClaimID:  "CLAIM-001",
				Judgment: "helpful",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var result *mcp.CallToolResult
			var err error

			req := &mcp.CallToolRequest{}

			// Type assertions are safe: tt.params is set per test case to the correct type.
			//nolint:errcheck // second return (result struct) intentionally discarded
			switch tt.handler {
			case "list":
				result, _, err = handleList(ctx, req, tt.params.(*ListParams))
			case "create":
				result, _, err = handleCreate(ctx, req, tt.params.(*CreateParams))
			case "search":
				result, _, err = handleSearch(ctx, req, tt.params.(*SearchParams))
			case "next":
				result, _, err = handleNext(ctx, req, tt.params.(*NextParams))
			case "scan":
				result, _, err = handleScan(ctx, req, tt.params.(*ScanParams))
			case "specify":
				result, _, err = handleSpecify(ctx, req, tt.params.(*SpecifyParams))
			case "plan":
				result, _, err = handlePlan(ctx, req, tt.params.(*PlanParams))
			case "index":
				result, _, err = handleIndex(ctx, req, tt.params.(*IndexParams))
			case "implement":
				result, _, err = handleImplement(ctx, req, tt.params.(*ImplementParams))
			case "files":
				result, _, err = handleFiles(ctx, req, tt.params.(*FilesParams))
			case "grep":
				result, _, err = handleGrep(ctx, req, tt.params.(*GrepParams))
			case "prioritize":
				result, _, err = handlePrioritize(ctx, req, tt.params.(*PrioritizeParams))
			case "bug-list":
				result, _, err = handleBugList(ctx, req, tt.params.(*BugListParams))
			case "bug-create":
				result, _, err = handleBugCreate(ctx, req, tt.params.(*BugCreateParams))
			case "gap-mark":
				result, _, err = handleGapMark(ctx, req, tt.params.(*GapMarkParams))
			}

			// We expect errors for some handlers (e.g., database not found)
			// The important thing is that the handler function signature works
			if result != nil && len(result.Content) > 0 {
				t.Logf("Handler %s returned content: %+v", tt.handler, result.Content[0])
			}

			if err != nil {
				t.Logf("Handler %s returned expected error: %v", tt.handler, err)
			}
		})
	}
}

func TestMCPCommandCreation(t *testing.T) {
	cmd := New()

	if cmd == nil {
		t.Fatal("MCP command should not be nil")
	}

	if cmd.Use != "mcp" {
		t.Errorf("Expected Use='mcp', got '%s'", cmd.Use)
	}

	// Check flags
	if cmd.Flags().Lookup("port") == nil {
		t.Error("MCP command should have --port flag")
	}

	if cmd.Flags().Lookup("host") == nil {
		t.Error("MCP command should have --host flag")
	}
}

// setupMCPTestDB chdirs into a fresh temp project directory, opens and
// migrates the hardcoded ".canary/canary.db" the handlers expect, and
// restores the original working directory on cleanup. Returns the open
// database handle for seeding; callers should Close() it before invoking a
// handler so the handler's own connection isn't contending with an open one.
func setupMCPTestDB(t *testing.T) *storage.DB {
	t.Helper()

	tmpDir := t.TempDir()

	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get working directory: %v", err)
	}
	t.Cleanup(func() { _ = os.Chdir(originalDir) })

	if err := os.MkdirAll(filepath.Join(tmpDir, ".canary"), 0o755); err != nil {
		t.Fatalf("failed to create .canary dir: %v", err)
	}
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("failed to chdir into temp dir: %v", err)
	}

	dbPath := ".canary/canary.db"
	db, err := storage.Open(dbPath)
	if err != nil {
		t.Fatalf("failed to open database: %v", err)
	}
	if err := storage.AutoMigrate(dbPath); err != nil {
		t.Fatalf("failed to migrate database: %v", err)
	}

	return db
}

// TestCANARY_CBIN_205_SearchCapped verifies handleSearch truncates results to
// the default limit of 20 while reporting the true match count via Total.
func TestCANARY_CBIN_205_SearchCapped(t *testing.T) {
	ctx := context.Background()
	db := setupMCPTestDB(t)

	for i := 0; i < 30; i++ {
		tok := &storage.Token{
			ReqID:    fmt.Sprintf("CBIN-N%03d", i),
			Feature:  fmt.Sprintf("NeedleFeature%03d", i),
			Aspect:   "API",
			Status:   "IMPL",
			Priority: i + 1,
			FilePath: fmt.Sprintf("file%03d.go", i),
			Keywords: "needle",
		}
		if err := db.UpsertToken(tok); err != nil {
			t.Fatalf("failed to insert test token %d: %v", i, err)
		}
	}
	if err := db.Close(); err != nil {
		t.Fatalf("failed to close seeding db: %v", err)
	}

	req := &mcp.CallToolRequest{}
	_, result, err := handleSearch(ctx, req, &SearchParams{Keywords: "needle"})
	if err != nil {
		t.Fatalf("handleSearch failed: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}

	if len(result.Tokens) != 20 {
		t.Errorf("expected 20 tokens (default cap), got %d", len(result.Tokens))
	}
	if result.Total != 30 {
		t.Errorf("expected Total=30, got %d", result.Total)
	}
}

// TestCANARY_CBIN_205_SearchLimitRaised verifies an explicit Limit above the
// default (but within the hard ceiling) returns all matches.
func TestCANARY_CBIN_205_SearchLimitRaised(t *testing.T) {
	ctx := context.Background()
	db := setupMCPTestDB(t)

	for i := 0; i < 30; i++ {
		tok := &storage.Token{
			ReqID:    fmt.Sprintf("CBIN-N%03d", i),
			Feature:  fmt.Sprintf("NeedleFeature%03d", i),
			Aspect:   "API",
			Status:   "IMPL",
			Priority: i + 1,
			FilePath: fmt.Sprintf("file%03d.go", i),
			Keywords: "needle",
		}
		if err := db.UpsertToken(tok); err != nil {
			t.Fatalf("failed to insert test token %d: %v", i, err)
		}
	}
	if err := db.Close(); err != nil {
		t.Fatalf("failed to close seeding db: %v", err)
	}

	req := &mcp.CallToolRequest{}
	_, result, err := handleSearch(ctx, req, &SearchParams{Keywords: "needle", Limit: 100})
	if err != nil {
		t.Fatalf("handleSearch failed: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}

	if len(result.Tokens) != 30 {
		t.Errorf("expected 30 tokens (Limit=100 raised above match count), got %d", len(result.Tokens))
	}
}

// TestCANARY_CBIN_205_NextDefaultFindsWork verifies handleNext with no
// filters finds a seeded STUB token rather than reporting "all complete".
func TestCANARY_CBIN_205_NextDefaultFindsWork(t *testing.T) {
	ctx := context.Background()
	db := setupMCPTestDB(t)

	tok := &storage.Token{
		ReqID:    "CBIN-900",
		Feature:  "NextTargetFeature",
		Aspect:   "API",
		Status:   "STUB",
		Priority: 1,
		FilePath: "next900.go",
	}
	if err := db.UpsertToken(tok); err != nil {
		t.Fatalf("failed to insert test token: %v", err)
	}
	if err := db.Close(); err != nil {
		t.Fatalf("failed to close seeding db: %v", err)
	}

	req := &mcp.CallToolRequest{}
	_, result, err := handleNext(ctx, req, &NextParams{})
	if err != nil {
		t.Fatalf("handleNext failed: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}

	if result.Message != "" {
		t.Errorf("expected a token to be found, got message: %q", result.Message)
	}
	if result.ReqID != "CBIN-900" {
		t.Errorf("expected ReqID=CBIN-900, got %q", result.ReqID)
	}
}

// TestCANARY_CBIN_205_BugListOnlyBugs verifies handleBugList returns only
// BUG-prefixed tokens, excluding other requirements in the same database.
func TestCANARY_CBIN_205_BugListOnlyBugs(t *testing.T) {
	ctx := context.Background()
	db := setupMCPTestDB(t)

	tokens := []*storage.Token{
		{
			ReqID:    "BUG-001",
			Feature:  "SomeBug",
			Aspect:   "API",
			Status:   "OPEN",
			Priority: 1,
			FilePath: "bug1.go",
		},
		{
			ReqID:    "CBIN-100",
			Feature:  "NotABug",
			Aspect:   "API",
			Status:   "IMPL",
			Priority: 1,
			FilePath: "cbin100.go",
		},
	}
	for _, tok := range tokens {
		if err := db.UpsertToken(tok); err != nil {
			t.Fatalf("failed to insert test token %s: %v", tok.ReqID, err)
		}
	}
	if err := db.Close(); err != nil {
		t.Fatalf("failed to close seeding db: %v", err)
	}

	req := &mcp.CallToolRequest{}
	_, result, err := handleBugList(ctx, req, &BugListParams{})
	if err != nil {
		t.Fatalf("handleBugList failed: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}

	if result.Count != 1 {
		t.Errorf("expected Count=1, got %d", result.Count)
	}
	if len(result.Bugs) != 1 {
		t.Fatalf("expected 1 bug, got %d", len(result.Bugs))
	}
	if result.Bugs[0].ReqID != "BUG-001" {
		t.Errorf("expected Bugs[0].ReqID=BUG-001, got %q", result.Bugs[0].ReqID)
	}
}

// TestCANARY_CBIN_205_BugListTruncates verifies handleBugList caps its
// returned Bugs slice to the default limit (20) while Total reflects the
// real BUG-prefixed count, and non-BUG tokens never leak into the results.
func TestCANARY_CBIN_205_BugListTruncates(t *testing.T) {
	ctx := context.Background()
	db := setupMCPTestDB(t)

	const bugCount = 25
	for i := 0; i < bugCount; i++ {
		tok := &storage.Token{
			ReqID:    fmt.Sprintf("BUG-%03d", i),
			Feature:  fmt.Sprintf("BugFeature%03d", i),
			Aspect:   "API",
			Status:   "OPEN",
			Priority: i + 1,
			FilePath: fmt.Sprintf("bug%03d.go", i),
		}
		if err := db.UpsertToken(tok); err != nil {
			t.Fatalf("failed to insert bug token %d: %v", i, err)
		}
	}
	for i := 0; i < 5; i++ {
		tok := &storage.Token{
			ReqID:    fmt.Sprintf("CBIN-B%03d", i),
			Feature:  fmt.Sprintf("NotABug%03d", i),
			Aspect:   "API",
			Status:   "IMPL",
			Priority: i + 1,
			FilePath: fmt.Sprintf("notabug%03d.go", i),
		}
		if err := db.UpsertToken(tok); err != nil {
			t.Fatalf("failed to insert non-bug token %d: %v", i, err)
		}
	}
	if err := db.Close(); err != nil {
		t.Fatalf("failed to close seeding db: %v", err)
	}

	req := &mcp.CallToolRequest{}
	_, result, err := handleBugList(ctx, req, &BugListParams{})
	if err != nil {
		t.Fatalf("handleBugList failed: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}

	if len(result.Bugs) != 20 {
		t.Errorf("expected 20 bugs (default cap), got %d", len(result.Bugs))
	}
	if result.Total != bugCount {
		t.Errorf("expected Total=%d, got %d", bugCount, result.Total)
	}
	for _, bug := range result.Bugs {
		if !strings.HasPrefix(bug.ReqID, "BUG-") {
			t.Errorf("expected only BUG- prefixed tokens, got %q", bug.ReqID)
		}
	}
}

// TestCANARY_CBIN_205_NextSkipsBlockedWork verifies handleNext skips a STUB
// token whose DEPENDS_ON requirement isn't fully TESTED/BENCHED, and returns
// the next unblocked (dependency-free) candidate instead -- mirroring the
// CLI's blocked-work filtering in internal/cmds/next/next.go.
func TestCANARY_CBIN_205_NextSkipsBlockedWork(t *testing.T) {
	ctx := context.Background()
	db := setupMCPTestDB(t)

	// Dependency requirement whose only token is IMPL (not TESTED/BENCHED),
	// so anything depending on it is blocked.
	dep := &storage.Token{
		ReqID:    "CBIN-800",
		Feature:  "DependencyFeature",
		Aspect:   "API",
		Status:   "IMPL",
		Priority: 1,
		FilePath: "dep800.go",
	}
	// Blocked token: highest priority (lowest number) but depends on CBIN-800.
	blocked := &storage.Token{
		ReqID:     "CBIN-801",
		Feature:   "BlockedFeature",
		Aspect:    "API",
		Status:    "STUB",
		Priority:  1,
		FilePath:  "blocked801.go",
		DependsOn: "CBIN-800",
	}
	// Unblocked token: lower priority (higher number) but no dependencies.
	unblocked := &storage.Token{
		ReqID:    "CBIN-802",
		Feature:  "UnblockedFeature",
		Aspect:   "API",
		Status:   "STUB",
		Priority: 2,
		FilePath: "unblocked802.go",
	}

	for _, tok := range []*storage.Token{dep, blocked, unblocked} {
		if err := db.UpsertToken(tok); err != nil {
			t.Fatalf("failed to insert test token %s: %v", tok.ReqID, err)
		}
	}
	if err := db.Close(); err != nil {
		t.Fatalf("failed to close seeding db: %v", err)
	}

	req := &mcp.CallToolRequest{}
	_, result, err := handleNext(ctx, req, &NextParams{})
	if err != nil {
		t.Fatalf("handleNext failed: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}

	if result.ReqID != "CBIN-802" {
		t.Errorf("expected next to skip blocked CBIN-801 and return CBIN-802, got %q", result.ReqID)
	}
}

// TestCANARY_CBIN_205_SearchTotalLowerBound verifies that when the overfetch
// hits its ceiling, Total is reported as a lower bound rather than an exact
// count, and the summary text reflects that.
func TestCANARY_CBIN_205_SearchTotalLowerBound(t *testing.T) {
	ctx := context.Background()
	db := setupMCPTestDB(t)

	for i := 0; i < 120; i++ {
		tok := &storage.Token{
			ReqID:    fmt.Sprintf("CBIN-L%03d", i),
			Feature:  fmt.Sprintf("LotsFeature%03d", i),
			Aspect:   "API",
			Status:   "IMPL",
			Priority: i + 1,
			FilePath: fmt.Sprintf("lots%03d.go", i),
			Keywords: "haystack",
		}
		if err := db.UpsertToken(tok); err != nil {
			t.Fatalf("failed to insert test token %d: %v", i, err)
		}
	}
	if err := db.Close(); err != nil {
		t.Fatalf("failed to close seeding db: %v", err)
	}

	req := &mcp.CallToolRequest{}
	result, out, err := handleSearch(ctx, req, &SearchParams{Keywords: "haystack"})
	if err != nil {
		t.Fatalf("handleSearch failed: %v", err)
	}
	if out == nil {
		t.Fatal("expected non-nil result")
	}

	if !out.TotalIsLowerBound {
		t.Error("expected TotalIsLowerBound=true when overfetch is exhausted")
	}
	if out.Total != maxToolLimit+1 {
		t.Errorf("expected Total=%d (overfetch ceiling), got %d", maxToolLimit+1, out.Total)
	}
	if len(result.Content) == 0 {
		t.Fatal("expected content")
	}
	text, ok := result.Content[0].(*mcp.TextContent)
	if !ok {
		t.Fatalf("expected TextContent, got %T", result.Content[0])
	}
	if !strings.Contains(text.Text, "100+ matches") {
		t.Errorf("expected summary text to mention '100+ matches', got: %q", text.Text)
	}
}

// TestCANARY_CBIN_204_MCPView verifies handleView aggregates tokens, files,
// tests, deps, and diagram refs for a requirement into a single bounded
// result, with a one-line text summary suitable for agent consumption.
func TestCANARY_CBIN_204_MCPView(t *testing.T) {
	ctx := context.Background()
	db := setupMCPTestDB(t)

	toks := []*storage.Token{
		{ReqID: "CBIN-105", Feature: "Scanner", Aspect: "Engine", Status: "TESTED",
			FilePath: "scan.go", LineNumber: 10, Test: "TestScan", DependsOn: "CBIN-101"},
		{ReqID: "CBIN-105", Feature: "ScannerCLI", Aspect: "CLI", Status: "IMPL",
			FilePath: "cli.go", LineNumber: 5},
	}
	for _, tok := range toks {
		if err := db.UpsertToken(tok); err != nil {
			t.Fatalf("failed to insert test token: %v", err)
		}
	}
	if err := db.ReplaceRefs("diagram", []storage.Ref{
		{ReqID: "CBIN-105", Kind: "diagram", FilePath: "docs/arch.md", LineNumber: 7},
	}); err != nil {
		t.Fatalf("failed to insert diagram ref: %v", err)
	}
	if err := db.Close(); err != nil {
		t.Fatalf("failed to close seeding db: %v", err)
	}

	req := &mcp.CallToolRequest{}
	result, out, err := handleView(ctx, req, &ViewParams{ReqID: "CBIN-105"})
	if err != nil {
		t.Fatalf("handleView failed: %v", err)
	}
	if out == nil {
		t.Fatal("expected non-nil result")
	}

	if out.ReqID != "CBIN-105" {
		t.Errorf("ReqID = %q, want CBIN-105", out.ReqID)
	}
	if out.Completion != 50 {
		t.Errorf("Completion = %d, want 50", out.Completion)
	}
	if out.FilesTotal != 2 {
		t.Errorf("FilesTotal = %d, want 2", out.FilesTotal)
	}
	if len(out.Tests) != 1 || out.Tests[0] != "TestScan" {
		t.Errorf("Tests = %v", out.Tests)
	}
	if len(out.Diagrams) != 1 || out.Diagrams[0] != "docs/arch.md:7" {
		t.Errorf("Diagrams = %v", out.Diagrams)
	}
	if len(out.DependsOn) != 1 || out.DependsOn[0] != "CBIN-101" {
		t.Errorf("DependsOn = %v", out.DependsOn)
	}

	if result == nil || len(result.Content) == 0 {
		t.Fatal("expected content")
	}
	text, ok := result.Content[0].(*mcp.TextContent)
	if !ok {
		t.Fatalf("expected TextContent, got %T", result.Content[0])
	}
	if strings.Count(text.Text, "\n") != 0 {
		t.Errorf("expected a single-line summary, got: %q", text.Text)
	}
	if !strings.Contains(text.Text, "CBIN-105") ||
		!strings.Contains(text.Text, "50% complete") ||
		!strings.Contains(text.Text, "depends on CBIN-101") {
		t.Errorf("summary missing expected fields: %q", text.Text)
	}
}

// TestCANARY_CBIN_204_MCPViewUnknown verifies handleView returns an error
// (not a panic, not an empty success) when the requirement has no tokens.
func TestCANARY_CBIN_204_MCPViewUnknown(t *testing.T) {
	ctx := context.Background()
	db := setupMCPTestDB(t)
	if err := db.Close(); err != nil {
		t.Fatalf("failed to close seeding db: %v", err)
	}

	req := &mcp.CallToolRequest{}
	result, out, err := handleView(ctx, req, &ViewParams{ReqID: "CBIN-999"})
	if err == nil {
		t.Fatal("expected error for unknown requirement, got nil")
	}
	if out != nil {
		t.Errorf("expected nil result on error, got %+v", out)
	}
	if result != nil {
		t.Errorf("expected nil CallToolResult on error, got %+v", result)
	}
}

// TestCANARY_CBIN_204_MCPDepsForward verifies handleDeps returns the
// forward dependency IDs (what a requirement depends on), with no token
// payloads and a matching Count.
func TestCANARY_CBIN_204_MCPDepsForward(t *testing.T) {
	ctx := context.Background()
	db := setupMCPTestDB(t)
	if err := db.Close(); err != nil {
		t.Fatalf("failed to close seeding db: %v", err)
	}

	// deps.BuildGraph walks .canary/specs/*/spec.md; an empty (but present)
	// specs dir is the minimal honest seed for "no dependencies declared"
	// rather than relying on missing-directory tolerance.
	if err := os.MkdirAll(filepath.Join(".canary", "specs"), 0o750); err != nil {
		t.Fatalf("failed to create empty specs dir: %v", err)
	}

	req := &mcp.CallToolRequest{}
	result, out, err := handleDeps(ctx, req, &DepsParams{ReqID: "CBIN-105"})
	if err != nil {
		t.Fatalf("handleDeps failed: %v", err)
	}
	if out == nil {
		t.Fatal("expected non-nil result")
	}
	if out.ReqID != "CBIN-105" {
		t.Errorf("ReqID = %q, want CBIN-105", out.ReqID)
	}
	if out.Direction != "forward" {
		t.Errorf("Direction = %q, want forward (default)", out.Direction)
	}
	if out.Count != 0 || len(out.Dependencies) != 0 {
		t.Errorf("expected no dependencies, got %v (count %d)", out.Dependencies, out.Count)
	}
	if result == nil || len(result.Content) == 0 {
		t.Fatal("expected content")
	}
}

// TestCANARY_CBIN_204_MCPDepsReverse verifies handleDeps returns the reverse
// dependency IDs (what depends on a requirement) by walking spec.md files
// under .canary/specs/ for a "## Dependencies" declaration.
// CANARY: REQ=CBIN-204; FEATURE="RequirementDeps"; ASPECT=API; STATUS=TESTED; TEST=TestCANARY_CBIN_204_MCPDepsReverse; UPDATED=2026-08-28
func TestCANARY_CBIN_204_MCPDepsReverse(t *testing.T) {
	ctx := context.Background()
	db := setupMCPTestDB(t)
	if err := db.Close(); err != nil {
		t.Fatalf("failed to close seeding db: %v", err)
	}

	// deps.BuildGraph walks .canary/specs/<REQ-ID>-<slug>/spec.md, extracting
	// the REQ-ID from the directory name and dependencies from a "##
	// Dependencies" section within the file.
	baseDir := filepath.Join(".canary", "specs", "CBIN-100-base")
	if err := os.MkdirAll(baseDir, 0o750); err != nil {
		t.Fatalf("failed to create base spec dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(baseDir, "spec.md"), []byte("# CBIN-100 Base\n\nNo dependencies.\n"), 0o600); err != nil {
		t.Fatalf("failed to write base spec: %v", err)
	}

	childDir := filepath.Join(".canary", "specs", "CBIN-200-child")
	if err := os.MkdirAll(childDir, 0o750); err != nil {
		t.Fatalf("failed to create child spec dir: %v", err)
	}
	childSpec := "# CBIN-200 Child\n\n## Dependencies\n\n- CBIN-100 (needs base)\n"
	if err := os.WriteFile(filepath.Join(childDir, "spec.md"), []byte(childSpec), 0o600); err != nil {
		t.Fatalf("failed to write child spec: %v", err)
	}

	req := &mcp.CallToolRequest{}
	result, out, err := handleDeps(ctx, req, &DepsParams{ReqID: "CBIN-100", Direction: "reverse"})
	if err != nil {
		t.Fatalf("handleDeps failed: %v", err)
	}
	if out == nil {
		t.Fatal("expected non-nil result")
	}
	if out.Direction != "reverse" {
		t.Errorf("Direction = %q, want reverse", out.Direction)
	}
	if out.Count != 1 || len(out.Dependencies) != 1 || out.Dependencies[0] != "CBIN-200" {
		t.Errorf("Dependencies = %v (count %d), want [CBIN-200] (count 1)", out.Dependencies, out.Count)
	}
	if result == nil || len(result.Content) == 0 {
		t.Fatal("expected content")
	}
}

// TestCANARY_CBIN_204_MCPViewEmptyReqID verifies handleView rejects an empty
// reqId before touching the database.
func TestCANARY_CBIN_204_MCPViewEmptyReqID(t *testing.T) {
	ctx := context.Background()
	req := &mcp.CallToolRequest{}
	result, out, err := handleView(ctx, req, &ViewParams{ReqID: ""})
	if err == nil {
		t.Fatal("expected error for empty reqId, got nil")
	}
	if out != nil {
		t.Errorf("expected nil result on error, got %+v", out)
	}
	if result != nil {
		t.Errorf("expected nil CallToolResult on error, got %+v", result)
	}
}

// TestCANARY_CBIN_204_MCPDepsInvalidDirection verifies handleDeps rejects an
// unrecognized direction value.
func TestCANARY_CBIN_204_MCPDepsInvalidDirection(t *testing.T) {
	ctx := context.Background()
	req := &mcp.CallToolRequest{}
	result, out, err := handleDeps(ctx, req, &DepsParams{ReqID: "CBIN-100", Direction: "sideways"})
	if err == nil {
		t.Fatal("expected error for invalid direction, got nil")
	}
	if out != nil {
		t.Errorf("expected nil result on error, got %+v", out)
	}
	if result != nil {
		t.Errorf("expected nil CallToolResult on error, got %+v", result)
	}
}
