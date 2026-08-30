// Copyright (c) 2025 by Developer Network.
//
// For more details, see the LICENSE file in the root directory of this
// source code repository or contact Developer Network at info@devnw.com.

// CANARY: REQ=ENG-4306; FEATURE="TokenStorage"; ASPECT=Storage; STATUS=IMPL; OWNER=canary; UPDATED=2025-10-16
package storage

import (
	"database/sql"
	"errors"
	"fmt"
	"regexp"
	"sort"
	"time"

	"github.com/jmoiron/sqlx"
	_ "modernc.org/sqlite"
)

// Token represents a parsed CANARY token with extended metadata
type Token struct {
	ID          int
	ReqID       string
	Feature     string
	Aspect      string
	Status      string
	FilePath    string
	LineNumber  int
	Test        string
	Bench       string
	Owner       string
	Priority    int
	Phase       string
	Keywords    string
	SpecStatus  string
	CreatedAt   string
	UpdatedAt   string
	StartedAt   string
	CompletedAt string
	CommitHash  string
	Branch      string
	DependsOn   string
	Blocks      string
	RelatedTo   string
	RawToken    string
	IndexedAt   string

	// CANARY: REQ=ENG-4315; FEATURE="DocDatabaseSchema"; ASPECT=Storage; STATUS=IMPL; UPDATED=2025-10-16
	// Documentation tracking fields
	DocPath      string // Comma-separated doc file paths (e.g., "user:docs/user.md,api:docs/api.md")
	DocHash      string // Comma-separated SHA256 hashes (abbreviated, first 16 chars)
	DocType      string // Documentation type (user, technical, feature, api, architecture)
	DocCheckedAt string // ISO 8601 timestamp of last staleness check
	DocStatus    string // DOC_CURRENT, DOC_STALE, DOC_MISSING, DOC_UNHASHED

	// CANARY: REQ=ENG-4319; FEATURE="TokenNamespacing"; ASPECT=Storage; STATUS=IMPL; UPDATED=2025-10-18
	// Multi-project support
	ProjectID string // Project identifier for token isolation

	// ContentHash is the hex SHA-256 of the file this token was read from at
	// index time, so a row can be checked against disk without re-scanning.
	ContentHash string
}

// Checkpoint represents a state snapshot
type Checkpoint struct {
	ID           int
	Name         string
	Description  string
	CommitHash   string
	CreatedAt    string
	TotalTokens  int
	StubCount    int
	ImplCount    int
	TestedCount  int
	BenchedCount int
	SnapshotJSON string
}

// DB wraps the SQLite database connection. Construct it with OpenRW (the
// writer, which may create and migrate) or OpenRO (the reader, which never
// creates anything).
type DB struct {
	conn     *sqlx.DB
	path     string
	readOnly bool
}

// ReadOnly reports whether this handle was opened read-only.
func (db *DB) ReadOnly() bool { return db.readOnly }

// Path returns the database file this handle was opened from.
func (db *DB) Path() string { return db.path }

// ErrProjectRequired is returned when an unscoped query matches rows in more
// than one project. Guessing which project the caller meant would answer a
// question they did not ask, so the ambiguity is reported instead.
var ErrProjectRequired = errors.New("PROJECT_REQUIRED")

// ErrInvalidOrderBy is returned when a caller asks for an ordering that is
// not in the allowlist below.
var ErrInvalidOrderBy = errors.New("INVALID_ORDER_BY")

// orderSQL maps the public order keys onto fixed ORDER BY clauses. The
// mapping exists so no part of an ordering is ever built from caller input:
// `--order-by` used to be concatenated into the query verbatim, which made
// the flag an arbitrary-SQL channel.
var orderSQL = map[string]string{
	"":              "priority ASC, updated_at DESC",
	"updated_desc":  "updated_at DESC",
	"req_asc":       "req_id ASC",
	"priority_desc": "priority DESC",
}

// OrderKeys lists the order keys a caller may name, sorted, excluding the
// empty default. It is the single source for the allowlist in help text and
// in the INVALID_ORDER_BY contract.
func OrderKeys() []string {
	keys := make([]string, 0, len(orderSQL))
	for k := range orderSQL {
		if k != "" {
			keys = append(keys, k)
		}
	}
	sort.Strings(keys)
	return keys
}

// resolveOrder maps an order key to its fixed clause.
func resolveOrder(orderKey string) (string, error) {
	clause, ok := orderSQL[orderKey]
	if !ok {
		return "", fmt.Errorf("%w: %q", ErrInvalidOrderBy, orderKey)
	}
	return clause, nil
}

// tokenColumns is the column list every token SELECT uses, in the order
// scanTokens reads them. Having exactly one list keeps project_id from being
// silently dropped by a query that forgot it -- which is how unscoped reads
// used to return other projects' rows with an empty ProjectID.
const tokenColumns = `id, req_id, feature, aspect, status, file_path, line_number,
		test, bench, owner, priority, phase, keywords, spec_status,
		created_at, updated_at, started_at, completed_at,
		commit_hash, branch, depends_on, blocks, related_to,
		raw_token, indexed_at,
		doc_path, doc_hash, doc_type, doc_checked_at, doc_status,
		COALESCE(project_id, '') AS project_id, COALESCE(content_hash, '') AS content_hash`

// Close closes the database connection
func (db *DB) Close() error {
	return db.conn.Close()
}

// upsertTokenQuery inserts a token, or updates the row that already occupies
// its (req_id, feature, file_path, line_number, project_id) slot.
const upsertTokenQuery = `
		INSERT INTO tokens (
			req_id, feature, aspect, status, file_path, line_number,
			test, bench, owner, priority, phase, keywords, spec_status,
			created_at, updated_at, started_at, completed_at,
			commit_hash, branch, depends_on, blocks, related_to,
			raw_token, indexed_at,
			doc_path, doc_hash, doc_type, doc_checked_at, doc_status,
			project_id, content_hash
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(req_id, feature, file_path, line_number, project_id)
		DO UPDATE SET
			aspect = excluded.aspect,
			status = excluded.status,
			test = excluded.test,
			bench = excluded.bench,
			owner = excluded.owner,
			priority = excluded.priority,
			phase = excluded.phase,
			keywords = excluded.keywords,
			spec_status = excluded.spec_status,
			updated_at = excluded.updated_at,
			started_at = excluded.started_at,
			completed_at = excluded.completed_at,
			commit_hash = excluded.commit_hash,
			branch = excluded.branch,
			depends_on = excluded.depends_on,
			blocks = excluded.blocks,
			related_to = excluded.related_to,
			raw_token = excluded.raw_token,
			indexed_at = excluded.indexed_at,
			doc_path = excluded.doc_path,
			doc_hash = excluded.doc_hash,
			doc_type = excluded.doc_type,
			doc_checked_at = excluded.doc_checked_at,
			doc_status = excluded.doc_status,
			project_id = excluded.project_id,
			content_hash = excluded.content_hash
	`

// DefaultProjectID is the identity a token carries when nothing configured
// one. Migration 000007 backfills it onto every pre-scoping row, so it is the
// value that makes an unconfigured database's rows reachable by a scoped
// query.
const DefaultProjectID = "default"

// upsertArgs flattens a token into upsertTokenQuery's bind order.
//
// A token with no project id is stored under DefaultProjectID rather than the
// empty string. project_id is what every scoped read and every scoped delete
// keys on, so a row with none would be unreachable by both -- indexable but
// never listed, never pruned.
func upsertArgs(token *Token) []any {
	if token.ProjectID == "" {
		token.ProjectID = DefaultProjectID
	}
	return []any{
		token.ReqID, token.Feature, token.Aspect, token.Status,
		token.FilePath, token.LineNumber,
		token.Test, token.Bench, token.Owner,
		token.Priority, token.Phase, token.Keywords, token.SpecStatus,
		token.CreatedAt, token.UpdatedAt, token.StartedAt, token.CompletedAt,
		token.CommitHash, token.Branch,
		token.DependsOn, token.Blocks, token.RelatedTo,
		token.RawToken, token.IndexedAt,
		token.DocPath, token.DocHash, token.DocType, token.DocCheckedAt, token.DocStatus,
		token.ProjectID, token.ContentHash,
	}
}

// UpsertToken inserts or updates a token
func (db *DB) UpsertToken(token *Token) error {
	_, err := db.conn.Exec(upsertTokenQuery, upsertArgs(token)...)
	return err
}

// CANARY: REQ=ENG-4319; FEATURE="TokenNamespacing"; ASPECT=Storage; STATUS=TESTED; TEST=TestAuditF08,TestAuditF08SingleProjectUnscoped; UPDATED=2026-08-30
// GetTokensByReqID retrieves every token for a requirement within projectID.
//
// An empty projectID means "whichever project this database holds": if the
// matching rows all belong to one project they are returned, and if they span
// more than one the call fails with ErrProjectRequired rather than mixing two
// projects' answers into one.
func (db *DB) GetTokensByReqID(projectID, reqID string) ([]*Token, error) {
	query := `SELECT ` + tokenColumns + `
		FROM tokens
		WHERE req_id = ?`
	args := []any{reqID}
	if projectID != "" {
		query += ` AND COALESCE(project_id, '') = ?`
		args = append(args, projectID)
	}
	query += ` ORDER BY priority ASC, feature ASC`

	rows, err := db.conn.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	tokens, err := scanTokens(rows)
	if err != nil {
		return nil, err
	}

	if projectID == "" && spansMultipleProjects(tokens) {
		return nil, fmt.Errorf("%w: %s", ErrProjectRequired, reqID)
	}
	return tokens, nil
}

// spansMultipleProjects reports whether tokens carry more than one project id.
func spansMultipleProjects(tokens []*Token) bool {
	if len(tokens) < 2 {
		return false
	}
	first := tokens[0].ProjectID
	for _, t := range tokens[1:] {
		if t.ProjectID != first {
			return true
		}
	}
	return false
}

// isHiddenPath determines if a token should be hidden based on its file path
// Hidden paths include test files, templates, documentation examples, and AI agent directories
func isHiddenPath(filePath string) bool {
	hiddenPatterns := []string{
		// Test files
		"_test.go",
		"Test.",
		"/tests/",
		"/test/",

		// Template directories
		".canary/templates/",
		"/templates/",
		"/base/",
		"/embedded/base/",
		"/embedded/",

		// Documentation examples
		"IMPLEMENTATION_SUMMARY",
		"FINAL_SUMMARY",
		"README_CANARY.md",
		"GAP_ANALYSIS.md",

		// AI agent directories
		".claude/",
		".cursor/",
		".github/prompts/",
		".windsurf/",
		".kilocode/",
		".roo/",
		".opencode/",
		".codex/",
		".augment/",
		".codebuddy/",
		".amazonq/",
	}

	for _, pattern := range hiddenPatterns {
		if len(filePath) >= len(pattern) && contains(filePath, pattern) {
			return true
		}
	}

	return false
}

// contains checks if a string contains a substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && indexOfSubstring(s, substr) >= 0
}

// indexOfSubstring returns the index of the first occurrence of substr in s, or -1 if not found
func indexOfSubstring(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}

// CANARY: REQ=ENG-4318; FEATURE="PriorityFiltering"; ASPECT=Storage; STATUS=TESTED; TEST=TestAuditF07,TestAuditF07AllowedKeys; UPDATED=2026-08-30
// ListTokens retrieves tokens with filters and ordering.
//
// projectID scopes the query; "" spans every project in the database.
// idPattern is a Go regexp applied to req_id after the query runs.
// orderKey must be one of OrderKeys() or "" (the default ordering); anything
// else is refused with ErrInvalidOrderBy and no query is issued.
func (db *DB) ListTokens(projectID string, filters map[string]any, idPattern string, orderKey string, limit int) ([]*Token, error) {
	orderClause, err := resolveOrder(orderKey)
	if err != nil {
		return nil, err
	}

	query := `
		SELECT ` + tokenColumns + `
		FROM tokens
		WHERE 1=1
	`
	args := []interface{}{}

	if projectID != "" {
		query += " AND COALESCE(project_id, '') = ?"
		args = append(args, projectID)
	}

	// Exclude template/placeholder tokens regardless of project prefix.
	// idPattern (a Go regexp) is applied to req_id in Go, after the query
	// runs, so it works for any requirement-ID scheme (CBIN, ticket-source
	// prefixes like GL-, PLAT-, etc.) rather than being hardcoded to CBIN.
	query += " AND req_id NOT LIKE '%XXX%' AND req_id NOT LIKE '%###%' AND req_id NOT LIKE '{{%' AND req_id NOT LIKE '%}}%'"

	// Filter hidden paths by default (unless include_hidden is set)
	if filterString(filters, "include_hidden") != "true" {
		// Exclude test files
		query += " AND file_path NOT LIKE '%_test.go%'"
		query += " AND file_path NOT LIKE '%Test.%'"
		query += " AND file_path NOT LIKE '%/tests/%'"
		query += " AND file_path NOT LIKE '%/test/%'"

		// Exclude template directories
		query += " AND file_path NOT LIKE '%.canary/templates/%'"
		query += " AND file_path NOT LIKE '%/templates/%'"
		query += " AND file_path NOT LIKE '%/base/%'"
		query += " AND file_path NOT LIKE '%/embedded/%'"

		// Exclude documentation examples
		query += " AND file_path NOT LIKE '%IMPLEMENTATION_SUMMARY%'"
		query += " AND file_path NOT LIKE '%FINAL_SUMMARY%'"
		query += " AND file_path NOT LIKE '%README_CANARY.md%'"

		// Exclude AI agent directories
		query += " AND file_path NOT LIKE '.claude/%'"
		query += " AND file_path NOT LIKE '.cursor/%'"
		query += " AND file_path NOT LIKE '.github/prompts/%'"
		query += " AND file_path NOT LIKE '.windsurf/%'"
		query += " AND file_path NOT LIKE '.kilocode/%'"
		query += " AND file_path NOT LIKE '.roo/%'"
		query += " AND file_path NOT LIKE '.opencode/%'"
		query += " AND file_path NOT LIKE '.codex/%'"
		query += " AND file_path NOT LIKE '.augment/%'"
		query += " AND file_path NOT LIKE '.codebuddy/%'"
		query += " AND file_path NOT LIKE '.amazonq/%'"
	}

	// Apply filters. Column names come from this fixed list, never from the
	// caller: only the bound values are caller-supplied.
	for _, f := range []struct{ key, clause string }{
		{"status", " AND status = ?"},
		{"aspect", " AND aspect = ?"},
		{"spec_status", " AND spec_status = ?"},
		{"phase", " AND phase = ?"},
		{"owner", " AND owner = ?"},
		{"priority_min", " AND priority >= ?"},
		{"priority_max", " AND priority <= ?"},
	} {
		if v, ok := filters[f.key]; ok {
			query += f.clause
			args = append(args, v)
		}
	}

	// Ordering: a fixed clause chosen by key, never text from the caller.
	query += " ORDER BY " + orderClause

	// Limit
	if limit > 0 {
		query += " LIMIT ?"
		args = append(args, limit)
	}

	rows, err := db.conn.Query(query, args...)
	if err != nil {
		return nil, err
	}

	defer func() { _ = rows.Close() }()

	tokens, err := scanTokens(rows)
	if err != nil {
		return nil, err
	}

	// Apply the requirement-ID pattern in Go (post-query) so callers can use
	// arbitrary regexp syntax rather than being limited to what SQLite's
	// GLOB/LIKE operators can express. Note: when a SQL LIMIT was applied
	// above and the pattern filters rows out, results can undershoot limit;
	// that's acceptable rather than over-engineering with SQL regexp.
	if idPattern != "" {
		re, err := regexp.Compile(idPattern)
		if err != nil {
			return nil, fmt.Errorf("invalid id pattern %q: %w", idPattern, err)
		}
		filtered := tokens[:0]
		for _, tok := range tokens {
			if re.MatchString(tok.ReqID) {
				filtered = append(filtered, tok)
			}
		}
		tokens = filtered
	}

	return tokens, nil
}

// DefaultSearchLimit caps keyword searches to protect agent context.
// Deliberately small; callers raise it explicitly (--limit / limit param)
// when they need more.
const DefaultSearchLimit = 25

// CANARY: REQ=ENG-4323; FEATURE="ContextCaps"; ASPECT=Storage; STATUS=TESTED; TEST=TestCANARY_CBIN_205_SearchTokensLimit; UPDATED=2026-08-28
// SearchTokens searches by keywords across keyword tags, feature names,
// requirement IDs, file paths, test names, and bench names, bounded by
// limit (or DefaultSearchLimit when limit <= 0).
// projectID scopes the search; "" spans every project in the database.
func (db *DB) SearchTokens(projectID, keywords string, limit int) ([]*Token, error) {
	if limit <= 0 {
		limit = DefaultSearchLimit
	}

	query := `
		SELECT ` + tokenColumns + `
		FROM tokens
		WHERE (keywords LIKE ? OR feature LIKE ? OR req_id LIKE ?
			OR file_path LIKE ? OR test LIKE ? OR bench LIKE ?)`

	pattern := "%" + keywords + "%"
	args := []any{pattern, pattern, pattern, pattern, pattern, pattern}
	if projectID != "" {
		query += ` AND COALESCE(project_id, '') = ?`
		args = append(args, projectID)
	}
	query += ` ORDER BY priority ASC LIMIT ?`
	args = append(args, limit)

	rows, err := db.conn.Query(query, args...)
	if err != nil {
		return nil, err
	}

	defer func() { _ = rows.Close() }()

	return scanTokens(rows)
}

// CANARY: REQ=CBIN-CLI-001; FEATURE="QueryAbstraction"; ASPECT=Storage; STATUS=TESTED; TEST=TestCANARY_CBIN_CLI_001_Storage_GetFilesByReqID; UPDATED=2026-08-29
// GetFilesByReqID groups tokens by file path for a requirement within
// projectID ("" spans every project, subject to GetTokensByReqID's ambiguity
// rule).
func (db *DB) GetFilesByReqID(projectID, reqID string, excludeSpecs bool) (map[string][]*Token, error) {
	tokens, err := db.GetTokensByReqID(projectID, reqID)
	if err != nil {
		return nil, err
	}

	// Group by file path, filter specs/templates if requested
	fileGroups := make(map[string][]*Token)
	for _, token := range tokens {
		if excludeSpecs && shouldExcludeFile(token.FilePath) {
			continue
		}
		fileGroups[token.FilePath] = append(fileGroups[token.FilePath], token)
	}

	return fileGroups, nil
}

// shouldExcludeFile checks if file is spec/template/plan
func shouldExcludeFile(path string) bool {
	excludePatterns := []string{
		".canary/specs/",
		".canary/templates/",
		"base/",
		"/plan.md",
		"/spec.md",
	}
	for _, pattern := range excludePatterns {
		if contains(path, pattern) {
			return true
		}
	}
	return false
}

// UpdatePriority updates the priority of a token within projectID ("" spans
// every project).
func (db *DB) UpdatePriority(projectID, reqID, feature string, priority int) error {
	query := `UPDATE tokens SET priority = ? WHERE req_id = ? AND feature = ?`
	args := []any{priority, reqID, feature}
	if projectID != "" {
		query += ` AND COALESCE(project_id, '') = ?`
		args = append(args, projectID)
	}
	_, err := db.conn.Exec(query, args...)
	return err
}

// UpdateSpecStatus updates the spec status within projectID ("" spans every
// project).
func (db *DB) UpdateSpecStatus(projectID, reqID, specStatus string) error {
	query := `UPDATE tokens SET spec_status = ? WHERE req_id = ?`
	args := []any{specStatus, reqID}
	if projectID != "" {
		query += ` AND COALESCE(project_id, '') = ?`
		args = append(args, projectID)
	}
	_, err := db.conn.Exec(query, args...)
	return err
}

// CreateCheckpoint creates a state snapshot of projectID's tokens ("" counts
// every project).
func (db *DB) CreateCheckpoint(projectID, name, description, commitHash, snapshotJSON string) error {
	// Get current counts. COALESCE keeps an empty table reporting zeroes
	// rather than a NULL scan error.
	countQuery := `
		SELECT
			COUNT(*),
			COALESCE(SUM(CASE WHEN status = 'STUB' THEN 1 ELSE 0 END), 0),
			COALESCE(SUM(CASE WHEN status = 'IMPL' THEN 1 ELSE 0 END), 0),
			COALESCE(SUM(CASE WHEN status = 'TESTED' THEN 1 ELSE 0 END), 0),
			COALESCE(SUM(CASE WHEN status = 'BENCHED' THEN 1 ELSE 0 END), 0)
		FROM tokens
		WHERE 1=1`
	var countArgs []any
	if projectID != "" {
		countQuery += ` AND COALESCE(project_id, '') = ?`
		countArgs = append(countArgs, projectID)
	}

	var total, stub, impl, tested, benched int
	err := db.conn.QueryRow(countQuery, countArgs...).Scan(&total, &stub, &impl, &tested, &benched)
	if err != nil {
		return err
	}

	query := `
		INSERT INTO checkpoints (name, description, commit_hash, created_at,
			total_tokens, stub_count, impl_count, tested_count, benched_count,
			snapshot_json)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err = db.conn.Exec(query, name, description, commitHash, time.Now().UTC().Format(time.RFC3339),
		total, stub, impl, tested, benched, snapshotJSON)
	return err
}

// GetCheckpoints retrieves all checkpoints
func (db *DB) GetCheckpoints() ([]*Checkpoint, error) {
	query := `
		SELECT id, name, description, commit_hash, created_at,
			total_tokens, stub_count, impl_count, tested_count, benched_count,
			snapshot_json
		FROM checkpoints
		ORDER BY created_at DESC
	`

	rows, err := db.conn.Query(query)
	if err != nil {
		return nil, err
	}

	defer func() { _ = rows.Close() }()

	var checkpoints []*Checkpoint
	for rows.Next() {
		cp := &Checkpoint{}
		err := rows.Scan(&cp.ID, &cp.Name, &cp.Description, &cp.CommitHash, &cp.CreatedAt,
			&cp.TotalTokens, &cp.StubCount, &cp.ImplCount, &cp.TestedCount, &cp.BenchedCount,
			&cp.SnapshotJSON)
		if err != nil {
			return nil, err
		}
		checkpoints = append(checkpoints, cp)
	}

	return checkpoints, rows.Err()
}

// scanTokens reads rows selected with tokenColumns. There is exactly one
// scanner because there is exactly one column list: a second one drifted from
// the first is how project_id came to be missing from most reads.
func scanTokens(rows *sql.Rows) ([]*Token, error) {
	var tokens []*Token
	for rows.Next() {
		t := &Token{}
		err := rows.Scan(
			&t.ID, &t.ReqID, &t.Feature, &t.Aspect, &t.Status,
			&t.FilePath, &t.LineNumber,
			&t.Test, &t.Bench, &t.Owner,
			&t.Priority, &t.Phase, &t.Keywords, &t.SpecStatus,
			&t.CreatedAt, &t.UpdatedAt, &t.StartedAt, &t.CompletedAt,
			&t.CommitHash, &t.Branch,
			&t.DependsOn, &t.Blocks, &t.RelatedTo,
			&t.RawToken, &t.IndexedAt,
			&t.DocPath, &t.DocHash, &t.DocType, &t.DocCheckedAt, &t.DocStatus,
			&t.ProjectID, &t.ContentHash,
		)
		if err != nil {
			return nil, err
		}
		tokens = append(tokens, t)
	}
	return tokens, rows.Err()
}

// filterString reads key from filters as a string. Filter values arrive as
// `any` so numeric bounds can be bound as integers; only the handful of
// values this package inspects itself need the string view.
func filterString(filters map[string]any, key string) string {
	v, ok := filters[key]
	if !ok {
		return ""
	}
	if s, ok := v.(string); ok {
		return s
	}
	return fmt.Sprintf("%v", v)
}

// CANARY: REQ=ENG-4319; FEATURE="TokenNamespacing"; ASPECT=Storage; STATUS=IMPL; UPDATED=2026-08-30
// GetTokensByProject retrieves all tokens for a specific project.
func (db *DB) GetTokensByProject(projectID string) ([]*Token, error) {
	query := `SELECT ` + tokenColumns + `
		FROM tokens
		WHERE COALESCE(project_id, '') = ?
		ORDER BY priority ASC, feature ASC
	`

	rows, err := db.conn.Query(query, projectID)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	return scanTokens(rows)
}

// GetAllTokens retrieves all tokens across all projects. It is deliberately
// unscoped: its callers (drift detection) reason about the whole database.
func (db *DB) GetAllTokens() ([]*Token, error) {
	query := `SELECT ` + tokenColumns + `
		FROM tokens
		ORDER BY priority ASC, updated_at DESC
	`

	rows, err := db.conn.Query(query)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	return scanTokens(rows)
}
