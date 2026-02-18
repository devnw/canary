// Copyright (c) 2025 by Developer Network.
//
// For more details, see the LICENSE file in the root directory of this
// source code repository or contact Developer Network at info@devnw.com.

package mcp

import (
	"context"
	"fmt"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"go.devnw.com/canary/internal/canaryscan"
	"go.devnw.com/canary/internal/storage"
)

// ListParams defines parameters for the list tool
type ListParams struct {
	Status string `json:"status,omitempty" jsonschema:"description:Filter by status (STUB IMPL TESTED BENCHED)"`
	Aspect string `json:"aspect,omitempty" jsonschema:"description:Filter by aspect (API CLI Engine etc.)"`
	Owner  string `json:"owner,omitempty" jsonschema:"description:Filter by owner"`
	Limit  int    `json:"limit,omitempty" jsonschema:"description:Maximum number of results"`
}

// ListResult defines the output for the list tool
type ListResult struct {
	Tokens []*storage.Token `json:"tokens"`
	Count  int              `json:"count"`
}

// handleList implements the list tool handler
func handleList(ctx context.Context, req *mcp.CallToolRequest, params *ListParams) (*mcp.CallToolResult, *ListResult, error) {
	dbPath := ".canary/canary.db"

	db, err := storage.Open(dbPath)
	if err != nil {
		return nil, nil, fmt.Errorf("open database: %w", err)
	}
	defer db.Close()

	filters := make(map[string]string)
	if params.Status != "" {
		filters["status"] = params.Status
	}
	if params.Aspect != "" {
		filters["aspect"] = params.Aspect
	}
	if params.Owner != "" {
		filters["owner"] = params.Owner
	}

	limit := params.Limit
	if limit == 0 {
		limit = 100 // Default limit
	}

	tokens, err := db.ListTokens(filters, "", "", limit)
	if err != nil {
		return nil, nil, fmt.Errorf("list tokens: %w", err)
	}

	result := &ListResult{
		Tokens: tokens,
		Count:  len(tokens),
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{
				Text: fmt.Sprintf("Found %d CANARY tokens", len(tokens)),
			},
		},
	}, result, nil
}

// ShowParams defines parameters for the show tool
type ShowParams struct {
	ReqID string `json:"reqId" jsonschema:"description:Requirement ID (e.g. CBIN-123),required"`
}

// ShowResult defines the output for the show tool
type ShowResult struct {
	ReqID  string           `json:"reqId"`
	Tokens []*storage.Token `json:"tokens"`
	Count  int              `json:"count"`
}

// handleShow implements the show tool handler
func handleShow(ctx context.Context, req *mcp.CallToolRequest, params *ShowParams) (*mcp.CallToolResult, *ShowResult, error) {
	if params.ReqID == "" {
		return nil, nil, fmt.Errorf("reqId is required")
	}

	dbPath := ".canary/canary.db"
	db, err := storage.Open(dbPath)
	if err != nil {
		return nil, nil, fmt.Errorf("open database: %w", err)
	}
	defer db.Close()

	tokens, err := db.GetTokensByReqID(params.ReqID)
	if err != nil {
		return nil, nil, fmt.Errorf("get tokens: %w", err)
	}

	result := &ShowResult{
		ReqID:  params.ReqID,
		Tokens: tokens,
		Count:  len(tokens),
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{
				Text: fmt.Sprintf("Found %d tokens for %s", len(tokens), params.ReqID),
			},
		},
	}, result, nil
}

// CreateParams defines parameters for the create tool
type CreateParams struct {
	ReqID   string `json:"reqId" jsonschema:"description:Requirement ID (e.g. CBIN-CLI-105),required"`
	Feature string `json:"feature" jsonschema:"description:Feature name,required"`
	Aspect  string `json:"aspect,omitempty" jsonschema:"description:Aspect (API CLI Engine etc.)"`
	Status  string `json:"status,omitempty" jsonschema:"description:Status (STUB IMPL TESTED BENCHED)"`
	Owner   string `json:"owner,omitempty" jsonschema:"description:Owner/assignee"`
}

// CreateResult defines the output for the create tool
type CreateResult struct {
	Token   string `json:"token"`
	ReqID   string `json:"reqId"`
	Feature string `json:"feature"`
	Aspect  string `json:"aspect"`
	Status  string `json:"status"`
}

// handleCreate implements the create tool handler
func handleCreate(ctx context.Context, req *mcp.CallToolRequest, params *CreateParams) (*mcp.CallToolResult, *CreateResult, error) {
	if params.ReqID == "" {
		return nil, nil, fmt.Errorf("reqId is required")
	}
	if params.Feature == "" {
		return nil, nil, fmt.Errorf("feature is required")
	}

	aspect := params.Aspect
	if aspect == "" {
		aspect = "API"
	}

	status := params.Status
	if status == "" {
		status = "IMPL"
	}

	today := time.Now().UTC().Format("2006-01-02")
	token := fmt.Sprintf("// CANARY: REQ=%s; FEATURE=\"%s\"; ASPECT=%s; STATUS=%s",
		params.ReqID, params.Feature, aspect, status)

	if params.Owner != "" {
		token += fmt.Sprintf("; OWNER=%s", params.Owner)
	}

	token += fmt.Sprintf("; UPDATED=%s", today)

	result := &CreateResult{
		Token:   token,
		ReqID:   params.ReqID,
		Feature: params.Feature,
		Aspect:  aspect,
		Status:  status,
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{
				Text: fmt.Sprintf("Created CANARY token for %s - %s", params.ReqID, params.Feature),
			},
		},
	}, result, nil
}

// StatusParams defines parameters for the status tool
type StatusParams struct {
	ReqID string `json:"reqId" jsonschema:"description:Requirement ID (e.g. CBIN-123),required"`
}

// StatusResult defines the output for the status tool
type StatusResult struct {
	ReqID         string           `json:"reqId"`
	Stats         StatusStats      `json:"stats"`
	CompletionPct int              `json:"completionPct"`
	Tokens        []*storage.Token `json:"tokens"`
}

// StatusStats holds progress statistics
type StatusStats struct {
	Total     int `json:"total"`
	Stub      int `json:"stub"`
	Impl      int `json:"impl"`
	Tested    int `json:"tested"`
	Benched   int `json:"benched"`
	Completed int `json:"completed"`
}

// handleStatus implements the status tool handler
func handleStatus(ctx context.Context, req *mcp.CallToolRequest, params *StatusParams) (*mcp.CallToolResult, *StatusResult, error) {
	if params.ReqID == "" {
		return nil, nil, fmt.Errorf("reqId is required")
	}

	dbPath := ".canary/canary.db"
	db, err := storage.Open(dbPath)
	if err != nil {
		return nil, nil, fmt.Errorf("open database: %w", err)
	}
	defer db.Close()

	tokens, err := db.GetTokensByReqID(params.ReqID)
	if err != nil {
		return nil, nil, fmt.Errorf("get tokens: %w", err)
	}

	stats := StatusStats{
		Total: len(tokens),
	}

	for _, token := range tokens {
		switch token.Status {
		case "STUB":
			stats.Stub++
		case "IMPL":
			stats.Impl++
		case "TESTED":
			stats.Tested++
			stats.Completed++
		case "BENCHED":
			stats.Benched++
			stats.Completed++
		}
	}

	completionPct := 0
	if stats.Total > 0 {
		completionPct = (stats.Completed * 100) / stats.Total
	}

	result := &StatusResult{
		ReqID:         params.ReqID,
		Stats:         stats,
		CompletionPct: completionPct,
		Tokens:        tokens,
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{
				Text: fmt.Sprintf("Status for %s: %d%% complete (%d/%d tokens)",
					params.ReqID, completionPct, stats.Completed, stats.Total),
			},
		},
	}, result, nil
}

// SearchParams defines parameters for the search tool
type SearchParams struct {
	Keywords string `json:"keywords" jsonschema:"description:Search keywords,required"`
}

// SearchResult defines the output for the search tool
type SearchResult struct {
	Keywords string           `json:"keywords"`
	Tokens   []*storage.Token `json:"tokens"`
	Count    int              `json:"count"`
}

// handleSearch implements the search tool handler
func handleSearch(ctx context.Context, req *mcp.CallToolRequest, params *SearchParams) (*mcp.CallToolResult, *SearchResult, error) {
	if params.Keywords == "" {
		return nil, nil, fmt.Errorf("keywords is required")
	}

	dbPath := ".canary/canary.db"
	db, err := storage.Open(dbPath)
	if err != nil {
		return nil, nil, fmt.Errorf("open database: %w", err)
	}
	defer db.Close()

	tokens, err := db.SearchTokens(params.Keywords)
	if err != nil {
		return nil, nil, fmt.Errorf("search tokens: %w", err)
	}

	result := &SearchResult{
		Keywords: params.Keywords,
		Tokens:   tokens,
		Count:    len(tokens),
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{
				Text: fmt.Sprintf("Found %d tokens matching '%s'", len(tokens), params.Keywords),
			},
		},
	}, result, nil
}

// NextParams defines parameters for the next tool
type NextParams struct {
	Status string `json:"status,omitempty" jsonschema:"description:Filter by status (STUB or IMPL)"`
	Aspect string `json:"aspect,omitempty" jsonschema:"description:Filter by aspect"`
}

// NextResult defines the output for the next tool
type NextResult struct {
	Token    *storage.Token `json:"token,omitempty"`
	ReqID    string         `json:"reqId,omitempty"`
	Feature  string         `json:"feature,omitempty"`
	Aspect   string         `json:"aspect,omitempty"`
	Status   string         `json:"status,omitempty"`
	Priority int            `json:"priority,omitempty"`
	Message  string         `json:"message,omitempty"`
}

// handleNext implements the next tool handler
func handleNext(ctx context.Context, req *mcp.CallToolRequest, params *NextParams) (*mcp.CallToolResult, *NextResult, error) {
	dbPath := ".canary/canary.db"
	db, err := storage.Open(dbPath)
	if err != nil {
		return nil, nil, fmt.Errorf("open database: %w", err)
	}
	defer db.Close()

	filters := make(map[string]string)
	if params.Status != "" {
		filters["status"] = params.Status
	} else {
		// Default to incomplete statuses
		filters["status"] = "STUB,IMPL"
	}

	if params.Aspect != "" {
		filters["aspect"] = params.Aspect
	}

	// Get highest priority token
	tokens, err := db.ListTokens(filters, "", "priority ASC, updated_at DESC", 1)
	if err != nil {
		return nil, nil, fmt.Errorf("query tokens: %w", err)
	}

	if len(tokens) == 0 {
		result := &NextResult{
			Message: "No unimplemented requirements found",
		}
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{
					Text: "All requirements completed!",
				},
			},
		}, result, nil
	}

	token := tokens[0]
	result := &NextResult{
		Token:    token,
		ReqID:    token.ReqID,
		Feature:  token.Feature,
		Aspect:   token.Aspect,
		Status:   token.Status,
		Priority: token.Priority,
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{
				Text: fmt.Sprintf("Next priority: %s - %s (Priority: %d, Status: %s)",
					token.ReqID, token.Feature, token.Priority, token.Status),
			},
		},
	}, result, nil
}

// ScanParams defines parameters for the scan tool
type ScanParams struct {
	Root        string `json:"root,omitempty" jsonschema:"description:Root directory to scan"`
	ProjectOnly bool   `json:"projectOnly,omitempty" jsonschema:"description:Filter by project requirement ID pattern"`
}

// ScanResult defines the output for the scan tool
type ScanResult struct {
	Message string `json:"message"`
	Root    string `json:"root"`
	Tokens  int    `json:"total_tokens"`
}

// handleScan implements the scan tool handler by calling the canary scanner.
func handleScan(ctx context.Context, req *mcp.CallToolRequest, params *ScanParams) (*mcp.CallToolResult, *ScanResult, error) {
	root := params.Root
	if root == "" {
		root = "."
	}
	skipRegex := canaryscan.DefaultSkipRegex()
	rep, err := canaryscan.Scan(root, skipRegex, nil, nil)
	if err != nil {
		return nil, nil, fmt.Errorf("scan %s: %w", root, err)
	}
	tokens := rep.Summary.TotalTokens
	unique := rep.Summary.UniqueRequirements
	msg := fmt.Sprintf("Scanned %s: %d tokens, %d unique requirements", root, tokens, unique)
	result := &ScanResult{
		Message: msg,
		Root:    root,
		Tokens:  rep.Summary.TotalTokens,
	}
	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: msg},
		},
	}, result, nil
}
