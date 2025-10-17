// Copyright (c) 2024 by CodePros.
//
// This software is proprietary information of CodePros.
// Unauthorized use, copying, modification, distribution, and/or
// disclosure is strictly prohibited, except as provided under the terms
// of the commercial license agreement you have entered into with
// CodePros.
//
// For more details, see the LICENSE file in the root directory of this
// source code repository or contact CodePros at info@codepros.org.

// CANARY: REQ=CBIN-123; FEATURE="TokenStorage"; ASPECT=Storage; STATUS=IMPL; OWNER=canary; UPDATED=2025-10-16
package storage

import (
	"database/sql"
	"fmt"
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

	// CANARY: REQ=CBIN-136; FEATURE="DocDatabaseSchema"; ASPECT=Storage; STATUS=IMPL; UPDATED=2025-10-16
	// Documentation tracking fields
	DocPath      string // Comma-separated doc file paths (e.g., "user:docs/user.md,api:docs/api.md")
	DocHash      string // Comma-separated SHA256 hashes (abbreviated, first 16 chars)
	DocType      string // Documentation type (user, technical, feature, api, architecture)
	DocCheckedAt string // ISO 8601 timestamp of last staleness check
	DocStatus    string // DOC_CURRENT, DOC_STALE, DOC_MISSING, DOC_UNHASHED
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

// DB wraps the SQLite database connection
type DB struct {
	conn *sqlx.DB
	path string
}

// Open opens or creates the CANARY database
// Note: Migrations are handled automatically by the CLI's PersistentPreRunE
func Open(dbPath string) (*DB, error) {
	// Initialize database connection
	conn, err := InitDB(dbPath)
	if err != nil {
		return nil, fmt.Errorf("initialize database: %w", err)
	}

	// Enable foreign keys
	if _, err := conn.Exec("PRAGMA foreign_keys = ON"); err != nil {
		conn.Close()
		return nil, fmt.Errorf("enable foreign keys: %w", err)
	}

	return &DB{conn: conn, path: dbPath}, nil
}

// Close closes the database connection
func (db *DB) Close() error {
	return db.conn.Close()
}

// UpsertToken inserts or updates a token
func (db *DB) UpsertToken(token *Token) error {
	query := `
		INSERT INTO tokens (
			req_id, feature, aspect, status, file_path, line_number,
			test, bench, owner, priority, phase, keywords, spec_status,
			created_at, updated_at, started_at, completed_at,
			commit_hash, branch, depends_on, blocks, related_to,
			raw_token, indexed_at,
			doc_path, doc_hash, doc_type, doc_checked_at, doc_status
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(req_id, feature, file_path, line_number)
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
			doc_status = excluded.doc_status
	`

	_, err := db.conn.Exec(query,
		token.ReqID, token.Feature, token.Aspect, token.Status,
		token.FilePath, token.LineNumber,
		token.Test, token.Bench, token.Owner,
		token.Priority, token.Phase, token.Keywords, token.SpecStatus,
		token.CreatedAt, token.UpdatedAt, token.StartedAt, token.CompletedAt,
		token.CommitHash, token.Branch,
		token.DependsOn, token.Blocks, token.RelatedTo,
		token.RawToken, token.IndexedAt,
		token.DocPath, token.DocHash, token.DocType, token.DocCheckedAt, token.DocStatus,
	)

	return err
}

// GetTokensByReqID retrieves all tokens for a requirement
func (db *DB) GetTokensByReqID(reqID string) ([]*Token, error) {
	query := `
		SELECT id, req_id, feature, aspect, status, file_path, line_number,
			test, bench, owner, priority, phase, keywords, spec_status,
			created_at, updated_at, started_at, completed_at,
			commit_hash, branch, depends_on, blocks, related_to,
			raw_token, indexed_at,
			doc_path, doc_hash, doc_type, doc_checked_at, doc_status
		FROM tokens
		WHERE req_id = ?
		ORDER BY priority ASC, feature ASC
	`

	rows, err := db.conn.Query(query, reqID)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	return scanTokens(rows)
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

// ListTokens retrieves tokens with filters and ordering
// idPattern is a regex pattern for filtering requirement IDs (e.g., "CBIN-[1-9][0-9]{2,}")
func (db *DB) ListTokens(filters map[string]string, idPattern string, orderBy string, limit int) ([]*Token, error) {
	query := `
		SELECT id, req_id, feature, aspect, status, file_path, line_number,
			test, bench, owner, priority, phase, keywords, spec_status,
			created_at, updated_at, started_at, completed_at,
			commit_hash, branch, depends_on, blocks, related_to,
			raw_token, indexed_at,
			doc_path, doc_hash, doc_type, doc_checked_at, doc_status
		FROM tokens
		WHERE 1=1
	`
	args := []interface{}{}

	// Apply ID pattern filter using GLOB (SQLite pattern matching)
	// Convert regex pattern to GLOB pattern for common cases
	if idPattern != "" {
		// For pattern like "CBIN-[1-9][0-9]{2,}", match CBIN-100 and above
		// Use GLOB which supports ? (any char) and * (any chars)
		// Since we can't easily convert regex to GLOB, we'll use a SQL filter
		// that excludes common placeholder patterns
		query += " AND req_id NOT LIKE 'CBIN-XXX%'"
		query += " AND req_id NOT LIKE 'CBIN-###%'"
		query += " AND req_id NOT LIKE '{{%'"
		query += " AND req_id NOT LIKE 'REQ-XXX%'"
		// Match 3+ digit CBIN IDs (CBIN-100 and above)
		query += " AND req_id GLOB 'CBIN-[0-9][0-9][0-9]*'"
		query += " AND req_id NOT GLOB 'CBIN-0[0-9][0-9]*'" // Exclude CBIN-001 through CBIN-099
	}

	// Filter hidden paths by default (unless include_hidden is set)
	includeHidden, _ := filters["include_hidden"]
	if includeHidden != "true" {
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

	// Apply filters
	if v, ok := filters["status"]; ok {
		query += " AND status = ?"
		args = append(args, v)
	}
	if v, ok := filters["aspect"]; ok {
		query += " AND aspect = ?"
		args = append(args, v)
	}
	if v, ok := filters["spec_status"]; ok {
		query += " AND spec_status = ?"
		args = append(args, v)
	}
	if v, ok := filters["phase"]; ok {
		query += " AND phase = ?"
		args = append(args, v)
	}
	if v, ok := filters["owner"]; ok {
		query += " AND owner = ?"
		args = append(args, v)
	}

	// Ordering
	if orderBy == "" {
		orderBy = "priority ASC, updated_at DESC"
	}
	query += " ORDER BY " + orderBy

	// Limit
	if limit > 0 {
		query += " LIMIT ?"
		args = append(args, limit)
	}

	rows, err := db.conn.Query(query, args...)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	return scanTokens(rows)
}

// SearchTokens searches by keywords
func (db *DB) SearchTokens(keywords string) ([]*Token, error) {
	query := `
		SELECT id, req_id, feature, aspect, status, file_path, line_number,
			test, bench, owner, priority, phase, keywords, spec_status,
			created_at, updated_at, started_at, completed_at,
			commit_hash, branch, depends_on, blocks, related_to,
			raw_token, indexed_at,
			doc_path, doc_hash, doc_type, doc_checked_at, doc_status
		FROM tokens
		WHERE keywords LIKE ? OR feature LIKE ? OR req_id LIKE ?
		ORDER BY priority ASC
	`

	pattern := "%" + keywords + "%"
	rows, err := db.conn.Query(query, pattern, pattern, pattern)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	return scanTokens(rows)
}

// CANARY: REQ=CBIN-CLI-001; FEATURE="QueryAbstraction"; ASPECT=Storage; STATUS=TESTED; TEST=TestCANARY_CBIN_CLI_001_Storage_GetFilesByReqID; UPDATED=2025-10-16
// GetFilesByReqID groups tokens by file path for a requirement
func (db *DB) GetFilesByReqID(reqID string, excludeSpecs bool) (map[string][]*Token, error) {
	tokens, err := db.GetTokensByReqID(reqID)
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

// UpdatePriority updates the priority of a token
func (db *DB) UpdatePriority(reqID, feature string, priority int) error {
	query := `UPDATE tokens SET priority = ? WHERE req_id = ? AND feature = ?`
	_, err := db.conn.Exec(query, priority, reqID, feature)
	return err
}

// UpdateSpecStatus updates the spec status
func (db *DB) UpdateSpecStatus(reqID, specStatus string) error {
	query := `UPDATE tokens SET spec_status = ? WHERE req_id = ?`
	_, err := db.conn.Exec(query, specStatus, reqID)
	return err
}

// CreateCheckpoint creates a state snapshot
func (db *DB) CreateCheckpoint(name, description, commitHash, snapshotJSON string) error {
	// Get current counts
	var total, stub, impl, tested, benched int
	err := db.conn.QueryRow(`
		SELECT
			COUNT(*),
			SUM(CASE WHEN status = 'STUB' THEN 1 ELSE 0 END),
			SUM(CASE WHEN status = 'IMPL' THEN 1 ELSE 0 END),
			SUM(CASE WHEN status = 'TESTED' THEN 1 ELSE 0 END),
			SUM(CASE WHEN status = 'BENCHED' THEN 1 ELSE 0 END)
		FROM tokens
	`).Scan(&total, &stub, &impl, &tested, &benched)
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

	defer rows.Close()

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

// Helper function to scan token rows
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
		)
		if err != nil {
			return nil, err
		}
		tokens = append(tokens, t)
	}
	return tokens, rows.Err()
}
