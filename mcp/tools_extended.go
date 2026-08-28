// Copyright (c) 2025 by Developer Network.
//
// For more details, see the LICENSE file in the root directory of this
// source code repository or contact Developer Network at info@devnw.com.

// This file defines extended MCP tools. The following handlers are
// currently stubs returning placeholder responses: specify, plan, index,
// BUG create (ID generation), and gap mark. Tracked in docs/GAP_ANALYSIS.md
// (GAP-0005). The scan tool is implemented in tools.go (internal/canaryscan).

package mcp

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"devnw.dev/canary/pkg/cmds/deps"
	"devnw.dev/canary/pkg/cmds/view"
	"devnw.dev/canary/pkg/specs"
	"devnw.dev/canary/pkg/storage"
)

// ========== SPECIFY TOOL ==========

// SpecifyParams defines parameters for the specify tool
type SpecifyParams struct {
	Description string `json:"description" jsonschema:"description:Feature description to create specification for,required"`
	Aspect      string `json:"aspect,omitempty" jsonschema:"description:Aspect (API CLI Engine Storage etc.)"`
}

// SpecifyResult defines the output for the specify tool
type SpecifyResult struct {
	Message     string `json:"message"`
	Description string `json:"description"`
	Aspect      string `json:"aspect"`
	SpecPath    string `json:"specPath,omitempty"`
}

// handleSpecify implements the specify tool handler
func handleSpecify(ctx context.Context, req *mcp.CallToolRequest, params *SpecifyParams) (*mcp.CallToolResult, *SpecifyResult, error) {
	if params.Description == "" {
		return nil, nil, fmt.Errorf("description is required")
	}

	aspect := params.Aspect
	if aspect == "" {
		aspect = "Engine"
	}

	// TODO: Implement actual specification generation
	// This would typically create a spec file in .canary/specs/
	result := &SpecifyResult{
		Message:     "Specification template created (full implementation pending)",
		Description: params.Description,
		Aspect:      aspect,
		SpecPath:    ".canary/specs/pending-spec.md",
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{
				Text: fmt.Sprintf("Created specification template for: %s", params.Description),
			},
		},
	}, result, nil
}

// ========== PLAN TOOL ==========

// PlanParams defines parameters for the plan tool
type PlanParams struct {
	ReqID     string `json:"reqId" jsonschema:"description:Requirement ID (e.g. CBIN-123),required"`
	TechStack string `json:"techStack,omitempty" jsonschema:"description:Technology stack description"`
}

// PlanResult defines the output for the plan tool
type PlanResult struct {
	Message   string `json:"message"`
	ReqID     string `json:"reqId"`
	TechStack string `json:"techStack,omitempty"`
	PlanPath  string `json:"planPath,omitempty"`
}

// handlePlan implements the plan tool handler
func handlePlan(ctx context.Context, req *mcp.CallToolRequest, params *PlanParams) (*mcp.CallToolResult, *PlanResult, error) {
	if params.ReqID == "" {
		return nil, nil, fmt.Errorf("reqId is required")
	}

	// TODO: Implement actual plan generation
	result := &PlanResult{
		Message:   "Implementation plan template created (full implementation pending)",
		ReqID:     params.ReqID,
		TechStack: params.TechStack,
		PlanPath:  fmt.Sprintf(".canary/specs/%s/plan.md", params.ReqID),
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{
				Text: fmt.Sprintf("Created implementation plan for %s", params.ReqID),
			},
		},
	}, result, nil
}

// ========== INDEX TOOL ==========

// IndexParams defines parameters for the index tool
type IndexParams struct {
	Root    string `json:"root,omitempty" jsonschema:"description:Root directory to index (default: current directory)"`
	Rebuild bool   `json:"rebuild,omitempty" jsonschema:"description:Force rebuild of index"`
}

// IndexResult defines the output for the index tool
type IndexResult struct {
	Message       string `json:"message"`
	TokensIndexed int    `json:"tokensIndexed"`
	FilesScanned  int    `json:"filesScanned"`
}

// handleIndex implements the index tool handler
func handleIndex(ctx context.Context, req *mcp.CallToolRequest, params *IndexParams) (*mcp.CallToolResult, *IndexResult, error) {
	root := params.Root
	if root == "" {
		root = "."
	}

	// TODO: Implement actual indexing
	// This would scan the codebase and populate the database
	result := &IndexResult{
		Message:       "Indexing functionality pending full implementation",
		TokensIndexed: 0,
		FilesScanned:  0,
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{
				Text: fmt.Sprintf("Would index directory: %s", root),
			},
		},
	}, result, nil
}

// ========== IMPLEMENT TOOL ==========

// ImplementParams defines parameters for the implement tool
type ImplementParams struct {
	ReqID string `json:"reqId" jsonschema:"description:Requirement ID to implement,required"`
}

// ImplementResult defines the output for the implement tool
type ImplementResult struct {
	Message      string `json:"message"`
	ReqID        string `json:"reqId"`
	Guidance     string `json:"guidance,omitempty"`
	SpecPath     string `json:"specPath,omitempty"`
	PlanPath     string `json:"planPath,omitempty"`
	HasSpec      bool   `json:"hasSpec"`
	HasPlan      bool   `json:"hasPlan"`
	TokenCount   int    `json:"tokenCount"`
	CurrentPhase string `json:"currentPhase,omitempty"`
}

// handleImplement implements the implement tool handler
func handleImplement(ctx context.Context, req *mcp.CallToolRequest, params *ImplementParams) (*mcp.CallToolResult, *ImplementResult, error) {
	if params.ReqID == "" {
		return nil, nil, fmt.Errorf("reqId is required")
	}

	dbPath := ".canary/canary.db"
	db, err := storage.Open(dbPath)
	if err != nil {
		return nil, nil, fmt.Errorf("open database: %w", err)
	}
	defer db.Close()

	// Get tokens for this requirement
	tokens, err := db.GetTokensByReqID(params.ReqID)
	if err != nil {
		return nil, nil, fmt.Errorf("get tokens: %w", err)
	}

	// Check for spec and plan files
	specPath := filepath.Join(".canary", "specs", params.ReqID, "spec.md")
	planPath := filepath.Join(".canary", "specs", params.ReqID, "plan.md")

	_, hasSpecErr := os.Stat(specPath)
	_, hasPlanErr := os.Stat(planPath)

	// Determine current phase
	phase := "Planning"
	if len(tokens) > 0 {
		allImpl := true
		anyTested := false
		for _, t := range tokens {
			if t.Status == "STUB" {
				allImpl = false
			}
			if t.Status == "TESTED" || t.Status == "BENCHED" {
				anyTested = true
			}
		}
		if anyTested {
			phase = "Testing/Validation"
		} else if allImpl {
			phase = "Implementation"
		}
	}

	result := &ImplementResult{
		Message:      "Implementation guidance ready",
		ReqID:        params.ReqID,
		Guidance:     fmt.Sprintf("Found %d tokens for %s. Current phase: %s", len(tokens), params.ReqID, phase),
		SpecPath:     specPath,
		PlanPath:     planPath,
		HasSpec:      hasSpecErr == nil,
		HasPlan:      hasPlanErr == nil,
		TokenCount:   len(tokens),
		CurrentPhase: phase,
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{
				Text: result.Guidance,
			},
		},
	}, result, nil
}

// ========== FILES TOOL ==========

// FilesParams defines parameters for the files tool
type FilesParams struct {
	ReqID string `json:"reqId" jsonschema:"description:Requirement ID to find files for,required"`
}

// FilesResult defines the output for the files tool
type FilesResult struct {
	ReqID     string   `json:"reqId"`
	Files     []string `json:"files"`
	FileCount int      `json:"fileCount"`
}

// handleFiles implements the files tool handler
func handleFiles(ctx context.Context, req *mcp.CallToolRequest, params *FilesParams) (*mcp.CallToolResult, *FilesResult, error) {
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

	// Collect unique file paths
	fileMap := make(map[string]bool)
	for _, token := range tokens {
		if token.FilePath != "" {
			fileMap[token.FilePath] = true
		}
	}

	files := make([]string, 0, len(fileMap))
	for file := range fileMap {
		files = append(files, file)
	}

	result := &FilesResult{
		ReqID:     params.ReqID,
		Files:     files,
		FileCount: len(files),
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{
				Text: fmt.Sprintf("Found %d files containing tokens for %s", len(files), params.ReqID),
			},
		},
	}, result, nil
}

// ========== PRIORITIZE TOOL ==========

// PrioritizeParams defines parameters for the prioritize tool
type PrioritizeParams struct {
	ReqID    string `json:"reqId" jsonschema:"description:Requirement ID to prioritize,required"`
	Priority int    `json:"priority" jsonschema:"description:Priority level (lower is higher priority),required"`
}

// PrioritizeResult defines the output for the prioritize tool
type PrioritizeResult struct {
	Message  string `json:"message"`
	ReqID    string `json:"reqId"`
	Priority int    `json:"priority"`
	Updated  int    `json:"updated"`
}

// handlePrioritize implements the prioritize tool handler
func handlePrioritize(ctx context.Context, req *mcp.CallToolRequest, params *PrioritizeParams) (*mcp.CallToolResult, *PrioritizeResult, error) {
	if params.ReqID == "" {
		return nil, nil, fmt.Errorf("reqId is required")
	}

	dbPath := ".canary/canary.db"
	db, err := storage.Open(dbPath)
	if err != nil {
		return nil, nil, fmt.Errorf("open database: %w", err)
	}
	defer db.Close()

	// Get tokens first to update each one
	tokens, err := db.GetTokensByReqID(params.ReqID)
	if err != nil {
		return nil, nil, fmt.Errorf("get tokens: %w", err)
	}

	// Update priority for each token (storage API requires feature parameter)
	updated := 0
	for _, token := range tokens {
		err = db.UpdatePriority(params.ReqID, token.Feature, params.Priority)
		if err == nil {
			updated++
		}
	}

	result := &PrioritizeResult{
		Message:  fmt.Sprintf("Updated priority for %s to %d (%d tokens)", params.ReqID, params.Priority, updated),
		ReqID:    params.ReqID,
		Priority: params.Priority,
		Updated:  updated,
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{
				Text: result.Message,
			},
		},
	}, result, nil
}

// ========== GREP TOOL ==========

// GrepParams defines parameters for the grep tool
type GrepParams struct {
	Pattern string `json:"pattern" jsonschema:"description:Pattern to search for in token fields,required"`
	Field   string `json:"field,omitempty" jsonschema:"description:Field to search (req feature aspect owner or all)"`
	Limit   int    `json:"limit,omitempty" jsonschema:"description:Maximum results (default 20, max 100)"`
}

// GrepResult defines the output for the grep tool
type GrepResult struct {
	Pattern string           `json:"pattern"`
	Field   string           `json:"field"`
	Tokens  []*storage.Token `json:"tokens"`
	Count   int              `json:"count"`
	Total   int              `json:"total"`
	// TotalIsLowerBound is true when the underlying overfetch hit its ceiling
	// (maxToolLimit+1 rows came back for an "all"-field search), meaning
	// Total is a floor, not an exact count.
	TotalIsLowerBound bool `json:"total_is_lower_bound,omitempty"`
}

// grepFieldValue returns the value of the named token field for field-scoped
// grep matching. An unrecognized field yields the empty string so it never
// matches, rather than silently falling back to "all" behavior.
func grepFieldValue(tok *storage.Token, field string) string {
	switch field {
	case "req":
		return tok.ReqID
	case "feature":
		return tok.Feature
	case "aspect":
		return tok.Aspect
	case "owner":
		return tok.Owner
	default:
		return ""
	}
}

// handleGrep implements the grep tool handler
func handleGrep(ctx context.Context, req *mcp.CallToolRequest, params *GrepParams) (*mcp.CallToolResult, *GrepResult, error) {
	if params.Pattern == "" {
		return nil, nil, fmt.Errorf("pattern is required")
	}

	field := params.Field
	if field == "" {
		field = "all"
	}

	dbPath := ".canary/canary.db"
	db, err := storage.Open(dbPath)
	if err != nil {
		return nil, nil, fmt.Errorf("open database: %w", err)
	}
	defer db.Close()

	var all []*storage.Token
	if field == "all" {
		all, err = db.SearchTokens(params.Pattern, maxToolLimit+1)
		if err != nil {
			return nil, nil, fmt.Errorf("search tokens: %w", err)
		}
	} else {
		// SearchTokens only matches keywords/feature/req_id/file_path/test/
		// bench, so it can't be reused for owner/aspect scoping. Fetch the
		// candidate set and filter in Go on the exact named field instead.
		candidates, err := db.ListTokens(nil, "", "", 0)
		if err != nil {
			return nil, nil, fmt.Errorf("list tokens: %w", err)
		}
		for _, tok := range candidates {
			if strings.Contains(grepFieldValue(tok, field), params.Pattern) {
				all = append(all, tok)
			}
		}
	}

	total := len(all)
	// Only the "all"-field branch overfetches with a capped query
	// (maxToolLimit+1); the field-scoped branch lists everything with no
	// cap, so its Total is always exact.
	lowerBound := field == "all" && total > maxToolLimit
	limit := capLimit(params.Limit)
	tokens := all
	if len(tokens) > limit {
		tokens = tokens[:limit]
	}

	result := &GrepResult{
		Pattern:           params.Pattern,
		Field:             field,
		Tokens:            tokens,
		Count:             len(tokens),
		Total:             total,
		TotalIsLowerBound: lowerBound,
	}

	totalText := fmt.Sprintf("%d matches", total)
	if lowerBound {
		totalText = fmt.Sprintf("%d+ matches", maxToolLimit)
	}
	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{
				Text: fmt.Sprintf("Found %s for pattern '%s' (showing %d)", totalText, params.Pattern, len(tokens)),
			},
		},
	}, result, nil
}

// ========== BUG TOOLS ==========

// BugListParams defines parameters for the bug list tool
type BugListParams struct {
	Status   string `json:"status,omitempty" jsonschema:"description:Filter by status (OPEN INVESTIGATING FIXED WONTFIX)"`
	Severity string `json:"severity,omitempty" jsonschema:"description:Filter by severity (CRITICAL HIGH MEDIUM LOW)"`
	Limit    int    `json:"limit,omitempty" jsonschema:"description:Maximum results (default 20, max 100)"`
}

// BugListResult defines the output for the bug list tool
type BugListResult struct {
	Bugs  []*storage.Token `json:"bugs"`
	Count int              `json:"count"`
	Total int              `json:"total"`
}

// handleBugList implements the bug list tool handler
func handleBugList(ctx context.Context, req *mcp.CallToolRequest, params *BugListParams) (*mcp.CallToolResult, *BugListResult, error) {
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

	// ListTokens has no "req_id_prefix" filter key; a bogus filter entry is
	// simply ignored, which previously made this return ALL tokens instead
	// of just bugs. Fetch the candidate set and filter for the BUG- prefix
	// in Go instead.
	tokens, err := db.ListTokens(filters, "", "updated_at DESC", 0)
	if err != nil {
		return nil, nil, fmt.Errorf("list bugs: %w", err)
	}

	var bugs []*storage.Token
	for _, tok := range tokens {
		if strings.HasPrefix(tok.ReqID, "BUG-") {
			bugs = append(bugs, tok)
		}
	}

	total := len(bugs)
	limit := capLimit(params.Limit)
	if len(bugs) > limit {
		bugs = bugs[:limit]
	}

	result := &BugListResult{
		Bugs:  bugs,
		Count: len(bugs),
		Total: total,
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{
				Text: fmt.Sprintf("Found %d bugs (showing %d)", total, len(bugs)),
			},
		},
	}, result, nil
}

// BugCreateParams defines parameters for the bug create tool
type BugCreateParams struct {
	Title       string `json:"title" jsonschema:"description:Bug title/description,required"`
	Severity    string `json:"severity,omitempty" jsonschema:"description:Severity level (CRITICAL HIGH MEDIUM LOW)"`
	Component   string `json:"component,omitempty" jsonschema:"description:Affected component"`
	Description string `json:"description,omitempty" jsonschema:"description:Detailed description"`
}

// BugCreateResult defines the output for the bug create tool
type BugCreateResult struct {
	Token    string `json:"token"`
	BugID    string `json:"bugId"`
	Title    string `json:"title"`
	Severity string `json:"severity"`
}

// handleBugCreate implements the bug create tool handler
func handleBugCreate(ctx context.Context, req *mcp.CallToolRequest, params *BugCreateParams) (*mcp.CallToolResult, *BugCreateResult, error) {
	if params.Title == "" {
		return nil, nil, fmt.Errorf("title is required")
	}

	severity := params.Severity
	if severity == "" {
		severity = "MEDIUM"
	}

	// Generate bug ID (simplified - actual implementation would check for uniqueness)
	bugID := fmt.Sprintf("BUG-%03d", 1) // TODO: Generate proper unique ID

	token := fmt.Sprintf("// CANARY: REQ=%s; FEATURE=\"%s\"; STATUS=OPEN; SEVERITY=%s",
		bugID, params.Title, severity)

	if params.Component != "" {
		token += fmt.Sprintf("; COMPONENT=%s", params.Component)
	}

	result := &BugCreateResult{
		Token:    token,
		BugID:    bugID,
		Title:    params.Title,
		Severity: severity,
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{
				Text: fmt.Sprintf("Created bug %s: %s", bugID, params.Title),
			},
		},
	}, result, nil
}

// ========== VIEW TOOL ==========

// ViewParams identifies the requirement to aggregate.
type ViewParams struct {
	ReqID string `json:"reqId" jsonschema:"description:requirement ID e.g. CBIN-105 or PLAT-4521,required"`
	Limit int    `json:"limit,omitempty" jsonschema:"description:max entries per list section (default 10)"`
}

// handleView returns the full bounded picture of one requirement: status,
// files, tests, deps, spec/plan, diagrams, and ticket URL, in one call.
// CANARY: REQ=CBIN-204; FEATURE="RequirementView"; ASPECT=API; STATUS=TESTED; TEST=TestCANARY_CBIN_204_MCPView,TestCANARY_CBIN_204_MCPViewUnknown,TestCANARY_CBIN_204_MCPViewEmptyReqID; UPDATED=2026-08-28
func handleView(ctx context.Context, req *mcp.CallToolRequest, params *ViewParams) (*mcp.CallToolResult, *view.View, error) {
	if params.ReqID == "" {
		return nil, nil, fmt.Errorf("reqId is required")
	}

	v, err := view.BuildView(".canary/canary.db", ".", params.ReqID, params.Limit)
	if err != nil {
		return nil, nil, err
	}

	// v.DiagramsTotal is the full diagram count (set from the un-truncated
	// list in BuildView), while v.Diagrams is capped at the request limit --
	// same relationship as FilesTotal/Files. Prefer DiagramsTotal so the
	// summary reports the true count even when the list section is capped;
	// it's only 0 when there are truly no diagrams (or none were queried),
	// in which case len(v.Diagrams) is also 0, so the fallback is safe.
	diagramsTotal := v.DiagramsTotal
	if diagramsTotal == 0 {
		diagramsTotal = len(v.Diagrams)
	}
	summary := fmt.Sprintf("%s: %d%% complete, %d files, %d tests, %d diagrams",
		v.ReqID, v.Completion, v.FilesTotal, len(v.Tests), diagramsTotal)
	if len(v.DependsOn) > 0 {
		summary += ", depends on " + strings.Join(v.DependsOn, ",")
	}
	if v.TicketURL != "" {
		summary += ", ticket " + v.TicketURL
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: summary}},
	}, v, nil
}

// ========== DEPS TOOL ==========

// DepsParams selects a requirement and traversal direction.
type DepsParams struct {
	ReqID     string `json:"reqId" jsonschema:"description:requirement ID,required"`
	Direction string `json:"direction,omitempty" jsonschema:"description:forward (what it depends on, default) or reverse (what depends on it)"`
}

// DepsResult carries dependency IDs only -- deliberately no token payloads.
type DepsResult struct {
	ReqID        string   `json:"reqId"`
	Direction    string   `json:"direction"`
	Dependencies []string `json:"dependencies"`
	Count        int      `json:"count"`
}

// handleDeps returns dependency IDs for a requirement, forward (what it
// depends on) or reverse (what depends on it). IDs only; callers use the
// view tool for detail on any returned ID.
// CANARY: REQ=CBIN-204; FEATURE="RequirementDeps"; ASPECT=API; STATUS=TESTED; TEST=TestCANARY_CBIN_204_MCPDepsForward,TestCANARY_CBIN_204_MCPDepsReverse,TestCANARY_CBIN_204_MCPDepsInvalidDirection; UPDATED=2026-08-28
func handleDeps(ctx context.Context, req *mcp.CallToolRequest, params *DepsParams) (*mcp.CallToolResult, *DepsResult, error) {
	if params.ReqID == "" {
		return nil, nil, fmt.Errorf("reqId is required")
	}

	dir := params.Direction
	if dir == "" {
		dir = "forward"
	}
	if dir != "forward" && dir != "reverse" {
		return nil, nil, fmt.Errorf("invalid direction %q: must be \"forward\" or \"reverse\"", dir)
	}

	graph, err := deps.BuildGraph()
	if err != nil {
		return nil, nil, err
	}

	var ids []string
	if dir == "reverse" {
		for _, d := range graph.GetReverseDependencies(params.ReqID) {
			ids = append(ids, d.Source)
		}
	} else {
		gg := specs.NewGraphGenerator(nil)
		ids = gg.GetTransitiveDependencies(graph, params.ReqID)
	}

	result := &DepsResult{ReqID: params.ReqID, Direction: dir, Dependencies: ids, Count: len(ids)}
	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("%s %s deps: %d", params.ReqID, dir, len(ids))}},
	}, result, nil
}

// ========== GAP ANALYSIS TOOLS ==========

// GapMarkParams defines parameters for the gap mark tool
type GapMarkParams struct {
	ClaimID  string `json:"claimId" jsonschema:"description:Gap analysis claim ID,required"`
	Judgment string `json:"judgment" jsonschema:"description:Judgment (helpful unhelpful unclear),required"`
	Reason   string `json:"reason,omitempty" jsonschema:"description:Reason for judgment"`
}

// GapMarkResult defines the output for the gap mark tool
type GapMarkResult struct {
	Message  string `json:"message"`
	ClaimID  string `json:"claimId"`
	Judgment string `json:"judgment"`
}

// handleGapMark implements the gap mark tool handler
func handleGapMark(ctx context.Context, req *mcp.CallToolRequest, params *GapMarkParams) (*mcp.CallToolResult, *GapMarkResult, error) {
	if params.ClaimID == "" {
		return nil, nil, fmt.Errorf("claimId is required")
	}
	if params.Judgment == "" {
		return nil, nil, fmt.Errorf("judgment is required")
	}

	// TODO: Implement actual gap marking in database
	result := &GapMarkResult{
		Message:  fmt.Sprintf("Marked claim %s as %s", params.ClaimID, params.Judgment),
		ClaimID:  params.ClaimID,
		Judgment: params.Judgment,
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{
				Text: result.Message,
			},
		},
	}, result, nil
}
