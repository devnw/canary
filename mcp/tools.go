// Copyright (c) 2025 by Developer Network.
//
// For more details, see the LICENSE file in the root directory of this
// source code repository or contact Developer Network at info@devnw.com.

package mcp

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"devnw.dev/canary/pkg/canaryscan"
	"devnw.dev/canary/pkg/external"
	"devnw.dev/canary/pkg/sources"
	"devnw.dev/canary/pkg/storage"
)

// CANARY: REQ=ENG-4323; FEATURE="ContextCaps"; ASPECT=API; STATUS=TESTED; TEST=TestCANARY_CBIN_205_SearchCapped; UPDATED=2026-08-28
const (
	defaultToolLimit = 20  // small by default to protect agent context
	maxToolLimit     = 100 // explicit ceiling even when the agent asks for more
)

// capLimit clamps a requested result count to [1, maxToolLimit], falling
// back to defaultToolLimit when the caller didn't ask for anything specific.
func capLimit(requested int) int {
	switch {
	case requested <= 0:
		return defaultToolLimit
	case requested > maxToolLimit:
		return maxToolLimit
	default:
		return requested
	}
}

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
	Total  int              `json:"total"`
	// TotalIsLowerBound is true when the underlying overfetch hit its ceiling
	// (maxToolLimit+1 rows came back), meaning Total is a floor, not an exact
	// count -- there may be more matches than reported.
	TotalIsLowerBound bool `json:"total_is_lower_bound,omitempty"`
}

// handleList implements the list tool handler
func handleList(ctx context.Context, req *mcp.CallToolRequest, params *ListParams) (*mcp.CallToolResult, *ListResult, error) {
	projectID := mcpProjectID()

	dbPath := ".canary/canary.db"

	db, err := storage.OpenRO(dbPath)
	if err != nil {
		return nil, nil, fmt.Errorf("open database: %w", err)
	}
	defer db.Close()

	filters := make(map[string]any)
	if params.Status != "" {
		filters["status"] = params.Status
	}
	if params.Aspect != "" {
		filters["aspect"] = params.Aspect
	}
	if params.Owner != "" {
		filters["owner"] = params.Owner
	}

	// Overfetch by one past maxToolLimit so Total reflects the true match
	// count (or a lower bound past the ceiling) before capping, matching
	// handleSearch's cap/Total pattern.
	all, err := db.ListTokens(projectID, filters, "", "", maxToolLimit+1)
	if err != nil {
		return nil, nil, fmt.Errorf("list tokens: %w", err)
	}

	total := len(all)
	lowerBound := total > maxToolLimit
	limit := capLimit(params.Limit)
	tokens := all
	if len(tokens) > limit {
		tokens = tokens[:limit]
	}

	result := &ListResult{
		Tokens:            tokens,
		Count:             len(tokens),
		Total:             total,
		TotalIsLowerBound: lowerBound,
	}

	// Compact summary in content so agents get the gist without parsing full result.
	text := listSummaryLine(tokens, limit)
	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: text},
		},
	}, result, nil
}

// listSummaryLine returns a short one-line summary of tokens for MCP content (reduces context).
func listSummaryLine(tokens []*storage.Token, limit int) string {
	if len(tokens) == 0 {
		return "Found 0 CANARY tokens"
	}
	return fmt.Sprintf("%d tokens (limit %d): %s", len(tokens), limit, tokensShortSummary(tokens, 5))
}

// tokensShortSummary returns up to max items as "REQ feature (status); ..." or "… +N more".
func tokensShortSummary(tokens []*storage.Token, max int) string {
	if len(tokens) == 0 {
		return ""
	}
	var sum string
	for i, t := range tokens {
		if i >= max {
			sum += fmt.Sprintf("… +%d more", len(tokens)-max)
			break
		}
		if i > 0 {
			sum += "; "
		}
		sum += fmt.Sprintf("%s %s (%s)", t.ReqID, t.Feature, t.Status)
	}
	return sum
}

// ShowParams defines parameters for the show tool
type ShowParams struct {
	ReqID string `json:"reqId" jsonschema:"description:Requirement ID (e.g. CBIN-123),required"`
	Limit int    `json:"limit,omitempty" jsonschema:"description:Maximum results (default 20, max 100)"`
}

// ShowResult defines the output for the show tool
type ShowResult struct {
	ReqID  string           `json:"reqId"`
	Tokens []*storage.Token `json:"tokens"`
	Count  int              `json:"count"`
	Total  int              `json:"total"`
}

// handleShow implements the show tool handler
func handleShow(ctx context.Context, req *mcp.CallToolRequest, params *ShowParams) (*mcp.CallToolResult, *ShowResult, error) {
	if params.ReqID == "" {
		return nil, nil, fmt.Errorf("reqId is required")
	}

	projectID := mcpProjectID()

	dbPath := ".canary/canary.db"
	db, err := storage.OpenRO(dbPath)
	if err != nil {
		return nil, nil, fmt.Errorf("open database: %w", err)
	}
	defer db.Close()

	tokens, err := db.GetTokensByReqID(projectID, params.ReqID)
	if err != nil {
		return nil, nil, fmt.Errorf("get tokens: %w", err)
	}

	total := len(tokens)
	limit := capLimit(params.Limit)
	if len(tokens) > limit {
		tokens = tokens[:limit]
	}

	result := &ShowResult{
		ReqID:  params.ReqID,
		Tokens: tokens,
		Count:  len(tokens),
		Total:  total,
	}

	text := showSummaryLine(params.ReqID, tokens)
	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: text},
		},
	}, result, nil
}

func showSummaryLine(reqID string, tokens []*storage.Token) string {
	if len(tokens) == 0 {
		return fmt.Sprintf("Found 0 tokens for %s", reqID)
	}
	const maxInSummary = 5
	sum := fmt.Sprintf("%d tokens for %s: ", len(tokens), reqID)
	for i, t := range tokens {
		if i >= maxInSummary {
			sum += fmt.Sprintf("… +%d more", len(tokens)-maxInSummary)
			break
		}
		if i > 0 {
			sum += ", "
		}
		sum += fmt.Sprintf("%s (%s)", t.Feature, t.Status)
	}
	return sum
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
	Total         int              `json:"total"`
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

	projectID := mcpProjectID()

	dbPath := ".canary/canary.db"
	db, err := storage.OpenRO(dbPath)
	if err != nil {
		return nil, nil, fmt.Errorf("open database: %w", err)
	}
	defer db.Close()

	tokens, err := db.GetTokensByReqID(projectID, params.ReqID)
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

	// Stats and completion are computed over ALL tokens above; only the
	// embedded token array is truncated to keep the response small.
	total := len(tokens)
	limit := capLimit(0)
	if len(tokens) > limit {
		tokens = tokens[:limit]
	}

	result := &StatusResult{
		ReqID:         params.ReqID,
		Stats:         stats,
		CompletionPct: completionPct,
		Tokens:        tokens,
		Total:         total,
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
	Limit    int    `json:"limit,omitempty" jsonschema:"description:Maximum results (default 20, max 100)"`
}

// SearchResult defines the output for the search tool
type SearchResult struct {
	Keywords string           `json:"keywords"`
	Tokens   []*storage.Token `json:"tokens"`
	Count    int              `json:"count"`
	Total    int              `json:"total"`
	// TotalIsLowerBound is true when the underlying overfetch hit its ceiling
	// (maxToolLimit+1 rows came back), meaning Total is a floor, not an exact
	// count -- there may be more matches than reported.
	TotalIsLowerBound bool `json:"total_is_lower_bound,omitempty"`
}

// handleSearch implements the search tool handler
func handleSearch(ctx context.Context, req *mcp.CallToolRequest, params *SearchParams) (*mcp.CallToolResult, *SearchResult, error) {
	if params.Keywords == "" {
		return nil, nil, fmt.Errorf("keywords is required")
	}

	projectID := mcpProjectID()

	dbPath := ".canary/canary.db"
	db, err := storage.OpenRO(dbPath)
	if err != nil {
		return nil, nil, fmt.Errorf("open database: %w", err)
	}
	defer db.Close()

	all, err := db.SearchTokens(projectID, params.Keywords, maxToolLimit+1)
	if err != nil {
		return nil, nil, fmt.Errorf("search tokens: %w", err)
	}

	total := len(all)
	lowerBound := total > maxToolLimit
	limit := capLimit(params.Limit)
	tokens := all
	if len(tokens) > limit {
		tokens = tokens[:limit]
	}

	result := &SearchResult{
		Keywords:          params.Keywords,
		Tokens:            tokens,
		Count:             len(tokens),
		Total:             total,
		TotalIsLowerBound: lowerBound,
	}

	totalText := fmt.Sprintf("%d matches", total)
	if lowerBound {
		totalText = fmt.Sprintf("%d+ matches", maxToolLimit)
	}
	text := fmt.Sprintf("Found %s for %q (showing %d): %s", totalText, params.Keywords, len(tokens), tokensShortSummary(tokens, 5))
	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: text},
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

// nextCandidateFetchLimit mirrors the CLI's next command
// (selectFromDatabase in internal/cmds/next/next.go), which fetches up to 50
// priority-ordered candidates per status so it can skip blocked tokens
// without missing unblocked work further down the list.
const nextCandidateFetchLimit = 50

// handleNext implements the next tool handler
func handleNext(ctx context.Context, req *mcp.CallToolRequest, params *NextParams) (*mcp.CallToolResult, *NextResult, error) {
	projectID := mcpProjectID()

	dbPath := ".canary/canary.db"
	db, err := storage.OpenRO(dbPath)
	if err != nil {
		return nil, nil, fmt.Errorf("open database: %w", err)
	}
	defer db.Close()

	// Source registry + per-call dedup for the "no cached status" stderr
	// note, shared across every hasUnresolvedDependencies call this request.
	reg, err := sources.LoadFromRoot(".")
	if err != nil {
		return nil, nil, fmt.Errorf("load .canary/project.yaml: %w", err)
	}
	warned := map[string]bool{}

	var token *storage.Token
	if params.Status != "" {
		filters := make(map[string]any)
		filters["status"] = params.Status
		if params.Aspect != "" {
			filters["aspect"] = params.Aspect
		}

		candidates, err := db.ListTokens(projectID, filters, "", "", nextCandidateFetchLimit)
		if err != nil {
			return nil, nil, fmt.Errorf("query tokens: %w", err)
		}
		token = firstUnblocked(db, projectID, candidates, reg, warned)
	} else {
		// No status filter: query STUB first, then IMPL if none found,
		// mirroring the CLI's next command (internal/cmds/next/next.go).
		// A single filters["status"] = "STUB,IMPL" never matches anything
		// since ListTokens does an exact equality comparison.
		for _, status := range []string{"STUB", "IMPL"} {
			filters := make(map[string]any)
			filters["status"] = status
			if params.Aspect != "" {
				filters["aspect"] = params.Aspect
			}

			candidates, err := db.ListTokens(projectID, filters, "", "", nextCandidateFetchLimit)
			if err != nil {
				return nil, nil, fmt.Errorf("query tokens: %w", err)
			}
			token = firstUnblocked(db, projectID, candidates, reg, warned)
			if token != nil {
				break
			}
		}
	}

	if token == nil {
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

// firstUnblocked returns the first candidate (already priority-ordered) whose
// dependencies are all resolved, or nil if every candidate is blocked.
func firstUnblocked(db *storage.DB, projectID string, candidates []*storage.Token, reg *sources.Registry, warned map[string]bool) *storage.Token {
	for _, tok := range candidates {
		if !hasUnresolvedDependencies(db, projectID, tok, reg, warned) {
			return tok
		}
	}
	return nil
}

// hasUnresolvedDependencies reports whether tok names a DEPENDS_ON
// requirement that blocks selection. This is a minimal replica of the CLI's
// unexported helper of the same name in pkg/cmds/next/next.go
// (hasUnresolvedDependencies) -- kept in sync manually since it can't be
// imported directly. Unlike the CLI, MCP has no --strict-external flag: an
// external (ticket-source) dependency with zero local tokens always uses the
// non-strict default (satisfied -> not blocking, unsatisfied -> blocking,
// unknown/no-cached-status -> not blocking, degradation is sacred).
// CANARY: REQ=ENG-3960; FEATURE="ExternalDeps"; ASPECT=API; STATUS=TESTED; TEST=TestCANARY_ENG_3960_MCP_Next_ExternalSatisfied_NotBlocking,TestCANARY_ENG_3960_MCP_Next_ExternalUnsatisfied_Blocking,TestCANARY_ENG_3960_MCP_Next_ExternalUnknown_NotBlocking,TestCANARY_ENG_3960_MCP_Next_LocalMissingDep_StillBlocking; UPDATED=2026-08-29
func hasUnresolvedDependencies(db *storage.DB, projectID string, tok *storage.Token, reg *sources.Registry, warned map[string]bool) bool {
	if tok.DependsOn == "" {
		return false
	}

	deps := strings.Split(tok.DependsOn, ",")
	for _, dep := range deps {
		dep = strings.TrimSpace(dep)
		if dep == "" {
			continue
		}

		depTokens, err := db.GetTokensByReqID(projectID, dep)
		if err != nil || len(depTokens) == 0 {
			res := external.Resolve(dep, reg, ".")
			if !res.IsExternal() {
				return true // Local dependency not found = blocking
			}
			switch res.State {
			case external.StateSatisfied:
				continue
			case external.StateUnsatisfied:
				return true
			default: // unknown: not blocking by default (degradation is sacred)
				if warned != nil && !warned[dep] {
					warned[dep] = true
					fmt.Fprintf(os.Stderr, "note: external dependency %s has no cached status (canary ticket status --refresh)\n", dep)
				}
				continue
			}
		}

		allComplete := true
		for _, depToken := range depTokens {
			if depToken.Status != "TESTED" && depToken.Status != "BENCHED" {
				allComplete = false
				break
			}
		}
		if !allComplete {
			return true // Dependency incomplete = blocking
		}
	}

	return false
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
	rep, err := canaryscan.Scan(root, skipRegex, nil, nil, nil)
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
