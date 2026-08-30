// Copyright (c) 2025 by Developer Network.
//
// For more details, see the LICENSE file in the root directory of this
// source code repository or contact Developer Network at info@devnw.com.

package mcp

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"devnw.dev/canary/pkg/canaryscan"
	"devnw.dev/canary/pkg/cmds/next"
	"devnw.dev/canary/pkg/config"
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
func (d Deps) handleList(ctx context.Context, req *mcp.CallToolRequest, params *ListParams) (*mcp.CallToolResult, *ListResult, error) {
	projectID := d.projectID()

	dbPath := d.db()

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
		return nil, nil, toolErr("list tokens", err)
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
func (d Deps) handleShow(ctx context.Context, req *mcp.CallToolRequest, params *ShowParams) (*mcp.CallToolResult, *ShowResult, error) {
	if params.ReqID == "" {
		return nil, nil, fmt.Errorf("reqId is required")
	}

	projectID := d.projectID()

	dbPath := d.db()
	db, err := storage.OpenRO(dbPath)
	if err != nil {
		return nil, nil, fmt.Errorf("open database: %w", err)
	}
	defer db.Close()

	tokens, err := db.GetTokensByReqID(projectID, params.ReqID)
	if err != nil {
		return nil, nil, toolErr("get tokens", err)
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
func (d Deps) handleCreate(ctx context.Context, req *mcp.CallToolRequest, params *CreateParams) (*mcp.CallToolResult, *CreateResult, error) {
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
func (d Deps) handleStatus(ctx context.Context, req *mcp.CallToolRequest, params *StatusParams) (*mcp.CallToolResult, *StatusResult, error) {
	if params.ReqID == "" {
		return nil, nil, fmt.Errorf("reqId is required")
	}

	projectID := d.projectID()

	dbPath := d.db()
	db, err := storage.OpenRO(dbPath)
	if err != nil {
		return nil, nil, fmt.Errorf("open database: %w", err)
	}
	defer db.Close()

	tokens, err := db.GetTokensByReqID(projectID, params.ReqID)
	if err != nil {
		return nil, nil, toolErr("get tokens", err)
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
func (d Deps) handleSearch(ctx context.Context, req *mcp.CallToolRequest, params *SearchParams) (*mcp.CallToolResult, *SearchResult, error) {
	if params.Keywords == "" {
		return nil, nil, fmt.Errorf("keywords is required")
	}

	projectID := d.projectID()

	dbPath := d.db()
	db, err := storage.OpenRO(dbPath)
	if err != nil {
		return nil, nil, fmt.Errorf("open database: %w", err)
	}
	defer db.Close()

	all, err := db.SearchTokens(projectID, params.Keywords, maxToolLimit+1)
	if err != nil {
		return nil, nil, toolErr("search tokens", err)
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
	// Source names where the answer came from: "database" (a fresh token
	// index) or "filesystem" (the index was missing or stale, so the tree was
	// scanned). An agent that gets no work back can tell an empty tree from
	// an unbuilt index.
	Source string `json:"source,omitempty"`
	// Blocked counts candidates passed over because a dependency was not
	// complete -- what separates "nothing left to do" from "everything left
	// is waiting on something".
	Blocked int `json:"blocked,omitempty"`
}

// handleNext answers the next-requirement question by delegating to the CLI's
// own selection.
//
// This tool used to carry a hand-maintained replica of the dependency rule,
// and the replica drifted: it accepted a declared STATUS=TESTED as proof of
// completion and let an external dependency with no cached status pass, long
// after `canary next` required evidence for the first and blocked on the
// second. Two implementations of "may this work start?" is one too many, so
// the replica is gone and next.SelectNext -- the same code path the CLI runs,
// including its stale-index fallback to a filesystem scan -- answers here.
//
// CANARY: REQ=ENG-3960; FEATURE="ExternalDeps"; ASPECT=API; STATUS=TESTED; TEST=TestCANARY_ENG_3960_MCP_Next_ExternalSatisfied_NotBlocking,TestCANARY_ENG_3960_MCP_Next_ExternalUnsatisfied_Blocking,TestCANARY_ENG_3960_MCP_Next_ExternalUnknown_Blocking,TestCANARY_ENG_3960_MCP_Next_LocalMissingDep_StillBlocking; UPDATED=2026-08-30
func (d Deps) handleNext(ctx context.Context, req *mcp.CallToolRequest, params *NextParams) (*mcp.CallToolResult, *NextResult, error) {
	filters := map[string]string{}
	if params.Status != "" {
		filters["status"] = params.Status
	}
	if params.Aspect != "" {
		filters["aspect"] = params.Aspect
	}

	// The notes about unresolvable dependencies go to the server's stderr,
	// not into the tool result: they are operator diagnostics, and folding
	// them into an agent-visible payload would make every answer noisier
	// without making any of them more actionable.
	token, source, blocked, err := next.SelectNext(d.db(), d.root(), d.projectID(), filters, false, os.Stderr)
	if err != nil {
		return nil, nil, toolErr("select next requirement", err)
	}

	if token == nil {
		msg := "No actionable requirements found"
		if blocked > 0 {
			msg = fmt.Sprintf("No actionable requirements: %d candidate(s) are blocked on incomplete dependencies", blocked)
		}
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: msg}},
		}, &NextResult{Message: msg, Source: source, Blocked: blocked}, nil
	}

	result := &NextResult{
		Token:    token,
		ReqID:    token.ReqID,
		Feature:  token.Feature,
		Aspect:   token.Aspect,
		Status:   token.Status,
		Priority: token.Priority,
		Source:   source,
		Blocked:  blocked,
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{
				Text: fmt.Sprintf("Next priority: %s - %s (Priority: %d, Status: %s, source: %s)",
					token.ReqID, token.Feature, token.Priority, token.Status, source),
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
	// Requirements is the unique requirement count, which the message has
	// always reported but the structured result did not.
	Requirements int `json:"unique_requirements"`
}

// handleScan scans a directory under the server root for CANARY tokens.
//
// It is the tool with a path parameter, so it is the tool that could be
// pointed anywhere: the previous version passed params.Root straight to the
// scanner, which walked whatever tree the server process could reach. Now the
// path is resolved against the server root and refused if it lands outside.
//
// It also runs the same inputs `canary scan` does -- the configured ticket
// sources and the tree's .canaryignore -- instead of nil for both, which is
// what made the tool report a different token count than the CLI on the same
// tree.
func (d Deps) handleScan(ctx context.Context, req *mcp.CallToolRequest, params *ScanParams) (*mcp.CallToolResult, *ScanResult, error) {
	root, err := d.confine(params.Root)
	if err != nil {
		return nil, nil, err
	}

	cfg, err := config.Load(d.root())
	if err != nil {
		return nil, nil, fmt.Errorf("load .canary/project.yaml: %w", err)
	}
	reg, err := sources.FromProjectConfig(cfg)
	if err != nil {
		return nil, nil, fmt.Errorf("load .canary/project.yaml: %w", err)
	}

	// projectOnly restricts the report to this project's own requirement ids.
	// The parameter was declared and then never read, so asking for it
	// silently returned every requirement in the tree.
	var projectFilter *regexp.Regexp
	if params.ProjectOnly {
		if cfg.Requirements.IDPattern == "" {
			return nil, nil, fmt.Errorf("projectOnly requires requirements.id_pattern in .canary/project.yaml")
		}
		projectFilter, err = regexp.Compile(cfg.Requirements.IDPattern)
		if err != nil {
			return nil, nil, fmt.Errorf("compile requirements.id_pattern %q: %w", cfg.Requirements.IDPattern, err)
		}
	}

	ignorePatterns, err := canaryscan.LoadCanaryIgnore(d.root())
	if err != nil {
		return nil, nil, fmt.Errorf("load .canaryignore: %w", err)
	}

	rep, err := canaryscan.Scan(root, canaryscan.DefaultSkipRegex(), projectFilter, ignorePatterns, reg)
	if err != nil {
		return nil, nil, fmt.Errorf("scan %s: %w", root, err)
	}

	rel := relativeToRoot(d.root(), root)
	msg := fmt.Sprintf("Scanned %s: %d tokens, %d unique requirements",
		rel, rep.Summary.TotalTokens, rep.Summary.UniqueRequirements)
	result := &ScanResult{
		Message:      msg,
		Root:         rel,
		Tokens:       rep.Summary.TotalTokens,
		Requirements: rep.Summary.UniqueRequirements,
	}
	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: msg},
		},
	}, result, nil
}

// relativeToRoot reports path as the caller asked about it -- relative to the
// server root -- so a result never discloses where on the filesystem the
// server happens to be rooted.
func relativeToRoot(root, path string) string {
	rel, err := filepath.Rel(root, path)
	if err != nil || rel == "" {
		return "."
	}
	return rel
}

// toolErr wraps a storage failure into a tool error an agent can act on.
//
// storage.ErrProjectRequired is the one refusal with a remedy the caller can
// apply, and its bare message ("PROJECT_REQUIRED") does not say what to do,
// so the remedy is spelled out here. Everything else is passed through
// unchanged.
func toolErr(op string, err error) error {
	if errors.Is(err, storage.ErrProjectRequired) {
		return fmt.Errorf("%s: PROJECT_REQUIRED: this index holds more than one project, so an unscoped answer would mix them; run 'canary index' for a single project or query the CLI with --project", op)
	}
	return fmt.Errorf("%s: %w", op, err)
}
