// Copyright (c) 2025 by Developer Network.
//
// For more details, see the LICENSE file in the root directory of this
// source code repository or contact Developer Network at info@devnw.com.

// This file defines the extended MCP tools. Every handler here has a
// postcondition a caller can check: the placeholder handlers that used to
// live alongside them (specify, plan, index, gap-mark, and a bug-create that
// always answered "BUG-001") are gone, not merely undocumented. The scan tool
// is implemented in tools.go.

package mcp

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"devnw.dev/canary/pkg/cmds/deps"
	"devnw.dev/canary/pkg/cmds/view"
	"devnw.dev/canary/pkg/sources"
	"devnw.dev/canary/pkg/specs"
	"devnw.dev/canary/pkg/storage"
)

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

// bugIDPattern is the shape of a bug id: BUG-<ASPECT>-NNN. Bug ids are not
// members of any configured ticket source, so the registry's pattern does not
// recognize them and they are matched separately.
var bugIDPattern = regexp.MustCompile(`^BUG-[A-Za-z0-9]+-\d+$`)

// validateReqID rejects anything that is not a requirement id this project
// could have.
//
// It exists because a requirement id is used to build filesystem paths. An id
// like "../../../etc" is not a requirement that happens to be missing, it is a
// traversal attempt, and it has to be refused before it reaches
// filepath.Join -- confinement afterwards is the second line of defence, not
// the first.
func (d Deps) validateReqID(reqID string) error {
	if strings.TrimSpace(reqID) == "" {
		return fmt.Errorf("reqId is required")
	}
	reg, err := sources.LoadFromRoot(d.root())
	if err != nil {
		return fmt.Errorf("load .canary/project.yaml: %w", err)
	}
	if p := reg.Pattern(); p != nil && p.FindString(reqID) == reqID {
		return nil
	}
	if bugIDPattern.MatchString(reqID) {
		return nil
	}
	keys := make([]string, 0, len(reg.Sources()))
	for _, s := range reg.Sources() {
		keys = append(keys, s.Key+"-<number>")
	}
	return fmt.Errorf("invalid requirement id %q: expected one of %s or BUG-<ASPECT>-<number>",
		reqID, strings.Join(keys, ", "))
}

// handleImplement reports what has been done for a requirement so far:
// how many tokens carry it, which phase those tokens put it in, and whether
// its spec and plan exist.
func (d Deps) handleImplement(ctx context.Context, req *mcp.CallToolRequest, params *ImplementParams) (*mcp.CallToolResult, *ImplementResult, error) {
	// Shape first: params.ReqID becomes a path two statements from here.
	if err := d.validateReqID(params.ReqID); err != nil {
		return nil, nil, err
	}

	projectID := d.projectID()

	dbPath := d.db()
	db, err := storage.OpenRO(dbPath)
	if err != nil {
		return nil, nil, fmt.Errorf("open database: %w", err)
	}
	defer db.Close()

	// Get tokens for this requirement
	tokens, err := db.GetTokensByReqID(projectID, params.ReqID)
	if err != nil {
		return nil, nil, toolErr("get tokens", err)
	}

	// Spec and plan live under the server root, and are confined to it: a
	// validated id cannot escape, and this proves it rather than assuming it.
	specPath, err := d.confine(filepath.Join(".canary", "specs", params.ReqID, "spec.md"))
	if err != nil {
		return nil, nil, err
	}
	planPath, err := d.confine(filepath.Join(".canary", "specs", params.ReqID, "plan.md"))
	if err != nil {
		return nil, nil, err
	}

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
		SpecPath:     relativeToRoot(d.root(), specPath),
		PlanPath:     relativeToRoot(d.root(), planPath),
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
func (d Deps) handleFiles(ctx context.Context, req *mcp.CallToolRequest, params *FilesParams) (*mcp.CallToolResult, *FilesResult, error) {
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
func (d Deps) handlePrioritize(ctx context.Context, req *mcp.CallToolRequest, params *PrioritizeParams) (*mcp.CallToolResult, *PrioritizeResult, error) {
	if params.ReqID == "" {
		return nil, nil, fmt.Errorf("reqId is required")
	}

	// A mutating tool resolves a real project id rather than the read-side
	// "": an unscoped UPDATE would rewrite every project sharing the
	// database, and storage refuses it outright.
	projectID := d.writeProjectID()

	// A mutating tool: OpenRW may create and migrate.
	db, err := storage.OpenRW(d.db())
	if err != nil {
		return nil, nil, fmt.Errorf("open database: %w", err)
	}
	defer db.Close()

	// Get tokens first to update each one
	tokens, err := db.GetTokensByReqID(projectID, params.ReqID)
	if err != nil {
		return nil, nil, toolErr("get tokens", err)
	}

	// Update priority for each token (storage API requires feature parameter).
	// A failure is reported rather than counted as a no-op: a tool that says
	// "updated 0 tokens" when the write itself failed is indistinguishable
	// from one that found nothing to update.
	updated := 0
	for _, token := range tokens {
		if err := db.UpdatePriority(projectID, params.ReqID, token.Feature, params.Priority); err != nil {
			return nil, nil, toolErr("update priority for "+params.ReqID+" ("+token.Feature+")", err)
		}
		updated++
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
func (d Deps) handleGrep(ctx context.Context, req *mcp.CallToolRequest, params *GrepParams) (*mcp.CallToolResult, *GrepResult, error) {
	if params.Pattern == "" {
		return nil, nil, fmt.Errorf("pattern is required")
	}

	field := params.Field
	if field == "" {
		field = "all"
	}

	projectID := d.projectID()

	dbPath := d.db()
	db, err := storage.OpenRO(dbPath)
	if err != nil {
		return nil, nil, fmt.Errorf("open database: %w", err)
	}
	defer db.Close()

	var all []*storage.Token
	if field == "all" {
		all, err = db.SearchTokens(projectID, params.Pattern, maxToolLimit+1)
		if err != nil {
			return nil, nil, toolErr("search tokens", err)
		}
	} else {
		// SearchTokens only matches keywords/feature/req_id/file_path/test/
		// bench, so it can't be reused for owner/aspect scoping. Fetch the
		// candidate set and filter in Go on the exact named field instead.
		candidates, err := db.ListTokens(projectID, nil, "", "", 0)
		if err != nil {
			return nil, nil, toolErr("list tokens", err)
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
func (d Deps) handleBugList(ctx context.Context, req *mcp.CallToolRequest, params *BugListParams) (*mcp.CallToolResult, *BugListResult, error) {
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

	// ListTokens has no "req_id_prefix" filter key; a bogus filter entry is
	// simply ignored, which previously made this return ALL tokens instead
	// of just bugs. Fetch the candidate set and filter for the BUG- prefix
	// in Go instead.
	tokens, err := db.ListTokens(projectID, filters, "", "updated_desc", 0)
	if err != nil {
		return nil, nil, toolErr("list bugs", err)
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
	Title       string `json:"title" jsonschema:"description:Bug title,required"`
	Aspect      string `json:"aspect,omitempty" jsonschema:"description:Aspect the bug belongs to (API CLI Engine Storage etc.) -- becomes the middle segment of the id"`
	Severity    string `json:"severity,omitempty" jsonschema:"description:Severity level (S1 S2 S3 S4)"`
	Priority    string `json:"priority,omitempty" jsonschema:"description:Priority level (P0 P1 P2 P3)"`
	File        string `json:"file,omitempty" jsonschema:"description:File the bug lives in relative to the server root, optionally with :line"`
	Owner       string `json:"owner,omitempty" jsonschema:"description:Bug owner/assignee"`
	Description string `json:"description,omitempty" jsonschema:"description:Detailed description, recorded on the token"`
}

// BugCreateResult defines the output for the bug create tool
type BugCreateResult struct {
	// Token is the CANARY comment to paste into the source file. The row is
	// rebuilt from source on the next `canary index`, so the comment is what
	// makes the bug survive a re-index.
	Token    string `json:"token"`
	BugID    string `json:"bugId"`
	Title    string `json:"title"`
	Aspect   string `json:"aspect"`
	Severity string `json:"severity"`
	Priority string `json:"priority"`
	FilePath string `json:"filePath"`
	Line     int    `json:"line"`
}

// aspectPattern is the shape a bug aspect may have. It becomes part of an id,
// so it is restricted to what an id segment can hold rather than accepting
// arbitrary caller text.
var aspectPattern = regexp.MustCompile(`^[A-Za-z0-9]+$`)

// CANARY: REQ=CP-282; FEATURE="TransactionalIDs"; ASPECT=API; STATUS=TESTED; TEST=TestAuditF10MutatingPostconditions,TestBugCreatePersistsReservedID; UPDATED=2026-08-30

// handleBugCreate reserves a bug id and persists the bug.
//
// The previous handler returned the literal string "BUG-001" every time and
// wrote nothing: two calls produced the same id, and neither produced a row.
// The id now comes from storage.ReserveID -- one immediate transaction, with
// the reservation table's primary key rejecting a duplicate -- and the token
// is upserted, so the id the caller is handed names a row that exists.
func (d Deps) handleBugCreate(ctx context.Context, req *mcp.CallToolRequest, params *BugCreateParams) (*mcp.CallToolResult, *BugCreateResult, error) {
	title := strings.TrimSpace(params.Title)
	if title == "" {
		return nil, nil, fmt.Errorf("title is required")
	}

	aspect := strings.ToUpper(strings.TrimSpace(params.Aspect))
	if aspect == "" {
		aspect = "API"
	}
	if !aspectPattern.MatchString(aspect) {
		return nil, nil, fmt.Errorf("invalid aspect %q: expected letters and digits only", params.Aspect)
	}

	severity := strings.ToUpper(strings.TrimSpace(params.Severity))
	if severity == "" {
		severity = "S3"
	}
	priority := strings.ToUpper(strings.TrimSpace(params.Priority))
	if priority == "" {
		priority = "P2"
	}

	// The file the bug is about is a caller-supplied path, so it is confined
	// to the server root like any other.
	filePath, line, err := d.bugLocation(params.File)
	if err != nil {
		return nil, nil, err
	}

	projectID := d.writeProjectID()
	db, err := storage.OpenRW(d.db())
	if err != nil {
		return nil, nil, fmt.Errorf("open database: %w", err)
	}
	defer db.Close()

	bugID, err := db.ReserveID(projectID, "BUG-"+aspect)
	if err != nil {
		return nil, nil, toolErr("reserve bug id", err)
	}

	updated := time.Now().UTC().Format("2006-01-02")
	token := buildBugToken(bugID, title, aspect, severity, priority, updated)

	row := &storage.Token{
		ReqID:      bugID,
		Feature:    title,
		Aspect:     aspect,
		Status:     "OPEN",
		FilePath:   filePath,
		LineNumber: line,
		Owner:      strings.TrimSpace(params.Owner),
		Priority:   bugPriorityValue(priority),
		Keywords:   fmt.Sprintf("SEVERITY=%s;PRIORITY=%s", severity, priority),
		Phase:      strings.TrimSpace(params.Description),
		UpdatedAt:  updated,
		IndexedAt:  time.Now().UTC().Format(time.RFC3339),
		RawToken:   token,
		ProjectID:  projectID,
	}
	if err := db.UpsertToken(row); err != nil {
		return nil, nil, toolErr("save bug "+bugID, err)
	}

	result := &BugCreateResult{
		Token:    token,
		BugID:    bugID,
		Title:    title,
		Aspect:   aspect,
		Severity: severity,
		Priority: priority,
		FilePath: filePath,
		Line:     line,
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{
				Text: fmt.Sprintf("Created bug %s: %s (%s/%s at %s:%d). Add the returned CANARY comment to the source file so the row survives the next 'canary index'.",
					bugID, title, severity, priority, filePath, line),
			},
		},
	}, result, nil
}

// bugLocation parses an optional "path" or "path:line" and confines the path
// to the server root. An unspecified location is recorded as the root itself
// rather than invented: a bug that names a file it is not in is worse than a
// bug that names none.
func (d Deps) bugLocation(file string) (string, int, error) {
	file = strings.TrimSpace(file)
	if file == "" {
		return ".", 0, nil
	}
	path := file
	line := 0
	if idx := strings.LastIndex(file, ":"); idx > 0 {
		if n, err := strconv.Atoi(file[idx+1:]); err == nil {
			path = file[:idx]
			line = n
		}
	}
	resolved, err := d.confine(path)
	if err != nil {
		return "", 0, err
	}
	return relativeToRoot(d.root(), resolved), line, nil
}

// buildBugToken renders the single-line CANARY comment for a bug, in the same
// grammar `canary bug create` emits so the scanner reads both identically.
func buildBugToken(bugID, title, aspect, severity, priority, updated string) string {
	return fmt.Sprintf("// CANARY: BUG=%s; TITLE=%q; FEATURE=%q; ASPECT=%s; STATUS=OPEN; SEVERITY=%s; PRIORITY=%s; UPDATED=%s",
		bugID, title, title, aspect, severity, priority, updated)
}

// bugPriorityValue maps P0..P3 onto the numeric priority the index sorts by,
// matching pkg/cmds/bug's parsePriorityValue.
func bugPriorityValue(priority string) int {
	switch priority {
	case "P0":
		return 0
	case "P1":
		return 1
	case "P2":
		return 2
	case "P3":
		return 3
	default:
		return 2
	}
}

// ========== VIEW TOOL ==========

// ViewParams identifies the requirement to aggregate.
type ViewParams struct {
	ReqID string `json:"reqId" jsonschema:"description:requirement ID e.g. CBIN-105 or PLAT-4521,required"`
	Limit int    `json:"limit,omitempty" jsonschema:"description:max entries per list section (default 10)"`
}

// handleView returns the full bounded picture of one requirement: status,
// files, tests, deps, spec/plan, diagrams, and ticket URL, in one call.
// CANARY: REQ=CP-270; FEATURE="RequirementView"; ASPECT=API; STATUS=TESTED; TEST=TestCANARY_CBIN_204_MCPView,TestCANARY_CBIN_204_MCPViewUnknown,TestCANARY_CBIN_204_MCPViewEmptyReqID; UPDATED=2026-08-28
// CANARY: REQ=ENG-4325; FEATURE="MigrateNotesView"; ASPECT=API; STATUS=TESTED; TEST=TestCANARY_CBIN_301_MCPViewMigrateNotes; UPDATED=2026-08-29
func (d Deps) handleView(ctx context.Context, req *mcp.CallToolRequest, params *ViewParams) (*mcp.CallToolResult, *view.View, error) {
	if params.ReqID == "" {
		return nil, nil, fmt.Errorf("reqId is required")
	}

	limit := params.Limit
	if limit > maxToolLimit {
		limit = maxToolLimit
	}
	v, err := view.BuildView(d.db(), d.root(), d.projectID(), params.ReqID, limit)
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
	if v.MigrateNotesTotal > 0 {
		summary += fmt.Sprintf(", %d migration notes", v.MigrateNotesTotal)
	}
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
// CANARY: REQ=CP-270; FEATURE="RequirementDeps"; ASPECT=API; STATUS=TESTED; TEST=TestCANARY_CBIN_204_MCPDepsForward,TestCANARY_CBIN_204_MCPDepsReverse,TestCANARY_CBIN_204_MCPDepsInvalidDirection; UPDATED=2026-08-28
func (d Deps) handleDeps(ctx context.Context, req *mcp.CallToolRequest, params *DepsParams) (*mcp.CallToolResult, *DepsResult, error) {
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

	graph, err := deps.BuildGraphIn(d.root())
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
