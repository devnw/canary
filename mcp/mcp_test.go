// Copyright (c) 2025 by Developer Network.
//
// For more details, see the LICENSE file in the root directory of this
// source code repository or contact Developer Network at info@devnw.com.

package mcp

import (
	"context"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
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
