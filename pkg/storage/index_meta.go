// Copyright (c) 2025 by Developer Network.
//
// For more details, see the LICENSE file in the root directory of this
// source code repository or contact Developer Network at info@devnw.com.

package storage

import (
	"database/sql"
	"errors"
	"fmt"
)

// CANARY: REQ=CBIN-306; FEATURE="IndexMetadata"; ASPECT=Storage; STATUS=TESTED; TEST=TestAuditF12,TestAuditF12CleanRunCommits,TestCANARY_CBIN_306_IndexMetaRoundTrip; UPDATED=2026-08-30

// IndexMeta records what an index was built from. Without it a reader can
// only see rows, not whether those rows still describe the tree in front of
// it -- which is how `canary next` came to announce "all requirements
// completed" over a database that was empty because it had never been built.
type IndexMeta struct {
	// Root is the directory the index was built from.
	Root string
	// ProjectID is the project the indexed tokens were written under.
	ProjectID string
	// CommitSHA is the git HEAD at index time, or "" when the tree is not a
	// git repository. It is never fabricated.
	CommitSHA string
	// ParserSchema is canaryscan.ParserSchemaVersion at index time, so an
	// index built by an older grammar can be recognised as stale.
	ParserSchema int
	// ScanDigest is the digest of the scanned file set (see index command).
	ScanDigest string
	// IndexedAt is the RFC3339 UTC timestamp of the run.
	IndexedAt string
}

// PutIndexMeta writes the single index_meta row, replacing whatever was
// there. It is exported for callers that record metadata outside a rebuild;
// `canary index` uses the transactional form so metadata and rows commit
// together or not at all.
func (db *DB) PutIndexMeta(m IndexMeta) error {
	_, err := db.conn.Exec(putIndexMetaQuery, putIndexMetaArgs(m)...)
	return err
}

// GetIndexMeta returns the recorded index metadata, or (nil, nil) when the
// index has never been built. A missing row is an answer, not an error.
func (db *DB) GetIndexMeta() (*IndexMeta, error) {
	var m IndexMeta
	err := db.conn.QueryRow(`
		SELECT root, project_id, commit_sha, parser_schema, scan_digest, indexed_at
		FROM index_meta WHERE id = 1
	`).Scan(&m.Root, &m.ProjectID, &m.CommitSHA, &m.ParserSchema, &m.ScanDigest, &m.IndexedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("read index metadata: %w", err)
	}
	return &m, nil
}

const putIndexMetaQuery = `
	INSERT INTO index_meta (id, root, project_id, commit_sha, parser_schema, scan_digest, indexed_at)
	VALUES (1, ?, ?, ?, ?, ?, ?)
	ON CONFLICT(id) DO UPDATE SET
		root = excluded.root,
		project_id = excluded.project_id,
		commit_sha = excluded.commit_sha,
		parser_schema = excluded.parser_schema,
		scan_digest = excluded.scan_digest,
		indexed_at = excluded.indexed_at
`

func putIndexMetaArgs(m IndexMeta) []any {
	return []any{m.Root, m.ProjectID, m.CommitSHA, m.ParserSchema, m.ScanDigest, m.IndexedAt}
}

// ReplaceIndex rebuilds one project's slice of the index in a single
// transaction: the project's existing token rows are deleted, every supplied
// token is inserted, the supplied reference kinds are replaced, and the
// metadata row is written. Any failure rolls the whole thing back, so a run
// that dies partway leaves the previous index exactly as it was rather than
// an emptied or half-populated table.
//
// refs is keyed by ref kind; a kind present with an empty slice clears that
// kind. Kinds absent from the map are left untouched.
func (db *DB) ReplaceIndex(projectID string, tokens []*Token, refs map[string][]Ref, meta IndexMeta) error {
	if db.readOnly {
		return errors.New("replace index: database opened read-only")
	}
	if projectID == "" {
		return errors.New("replace index: project id is required")
	}

	// The connection is opened with _txlock=immediate, so BEGIN takes the
	// write lock here rather than discovering a competing writer partway
	// through the rebuild.
	tx, err := db.conn.Beginx()
	if err != nil {
		return fmt.Errorf("begin index transaction: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	if _, err := tx.Exec(`DELETE FROM tokens WHERE COALESCE(project_id, '') = ?`, projectID); err != nil {
		return fmt.Errorf("clear token index for project %s: %w", projectID, err)
	}

	stmt, err := tx.Preparex(upsertTokenQuery)
	if err != nil {
		return fmt.Errorf("prepare token insert: %w", err)
	}
	defer func() { _ = stmt.Close() }()

	for _, t := range tokens {
		if _, err := stmt.Exec(upsertArgs(t)...); err != nil {
			return fmt.Errorf("index token %s/%s at %s:%d: %w", t.ReqID, t.Feature, t.FilePath, t.LineNumber, err)
		}
	}

	for kind, list := range refs {
		if _, err := tx.Exec(`DELETE FROM refs WHERE kind = ?`, kind); err != nil {
			return fmt.Errorf("clear %s refs: %w", kind, err)
		}
		for _, r := range list {
			if _, err := tx.Exec(
				`INSERT OR IGNORE INTO refs (req_id, kind, file_path, line_number, context) VALUES (?, ?, ?, ?, ?)`,
				r.ReqID, kind, r.FilePath, r.LineNumber, r.Context,
			); err != nil {
				return fmt.Errorf("index %s ref %s: %w", kind, r.ReqID, err)
			}
		}
	}

	if _, err := tx.Exec(putIndexMetaQuery, putIndexMetaArgs(meta)...); err != nil {
		return fmt.Errorf("record index metadata: %w", err)
	}

	// Validation before commit: the row count must match what was handed in.
	// An UPSERT that silently collapsed two scanned tokens into one row would
	// otherwise commit a smaller index than the tree actually holds.
	var stored int
	if err := tx.Get(&stored, `SELECT COUNT(*) FROM tokens WHERE COALESCE(project_id, '') = ?`, projectID); err != nil {
		return fmt.Errorf("verify index row count: %w", err)
	}
	if want := len(tokens); stored != want {
		return fmt.Errorf("index verification failed: stored %d rows, scanned %d tokens", stored, want)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit index: %w", err)
	}
	return nil
}
