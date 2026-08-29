// Copyright (c) 2025 by Developer Network.
//
// For more details, see the LICENSE file in the root directory of this
// source code repository or contact Developer Network at info@devnw.com.

package storage

// CANARY: REQ=CP-272; FEATURE="DiagramRefsIndex"; ASPECT=Storage; STATUS=TESTED; TEST=TestCANARY_CBIN_206_RefsRoundTrip; UPDATED=2026-08-28

// Ref is a requirement reference found outside CANARY tokens (diagrams, docs).
type Ref struct {
	ReqID      string `db:"req_id" json:"req_id"`
	Kind       string `db:"kind" json:"kind"`
	FilePath   string `db:"file_path" json:"file_path"`
	LineNumber int    `db:"line_number" json:"line_number"`
	Context    string `db:"context" json:"context,omitempty"`
}

// ReplaceRefs atomically replaces all refs of the given kind.
func (db *DB) ReplaceRefs(kind string, refs []Ref) error {
	tx, err := db.conn.Beginx()
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()
	if _, err := tx.Exec(`DELETE FROM refs WHERE kind = ?`, kind); err != nil {
		return err
	}
	for _, r := range refs {
		if _, err := tx.Exec(
			`INSERT OR IGNORE INTO refs (req_id, kind, file_path, line_number, context) VALUES (?, ?, ?, ?, ?)`,
			r.ReqID, kind, r.FilePath, r.LineNumber, r.Context,
		); err != nil {
			return err
		}
	}
	return tx.Commit()
}

// GetRefsByReqID returns refs for one requirement, ordered by file then line.
func (db *DB) GetRefsByReqID(reqID string) ([]*Ref, error) {
	rows, err := db.conn.Queryx(
		`SELECT req_id, kind, file_path, line_number, COALESCE(context,'') AS context
		 FROM refs WHERE req_id = ? ORDER BY file_path, line_number`, reqID)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	var out []*Ref
	for rows.Next() {
		var r Ref
		if err := rows.StructScan(&r); err != nil {
			return nil, err
		}
		out = append(out, &r)
	}
	return out, rows.Err()
}

// maxRefsByKind caps GetRefsByKind results when limit<=0 is requested.
const maxRefsByKind = 100

// CANARY: REQ=ENG-4325; FEATURE="MigrateRefsIndex"; ASPECT=Storage; STATUS=TESTED; TEST=TestCANARY_CBIN_301_MigrateRefsRoundTrip,TestCANARY_CBIN_301_GetRefsByKindLimit; UPDATED=2026-08-29

// GetRefsByKind returns refs of the given kind across all requirements,
// ordered by file then line. limit<=0 defaults to a 100-row cap.
func (db *DB) GetRefsByKind(kind string, limit int) ([]*Ref, error) {
	if limit <= 0 {
		limit = maxRefsByKind
	}
	rows, err := db.conn.Queryx(
		`SELECT req_id, kind, file_path, line_number, COALESCE(context,'') AS context
		 FROM refs WHERE kind = ? ORDER BY file_path, line_number LIMIT ?`, kind, limit)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	var out []*Ref
	for rows.Next() {
		var r Ref
		if err := rows.StructScan(&r); err != nil {
			return nil, err
		}
		out = append(out, &r)
	}
	return out, rows.Err()
}

// DeleteAllTokens clears the tokens table. `canary index` treats the token
// index as fully derived state: each run rebuilds it from a whole-tree scan,
// so rows for tokens that no longer exist on disk (renamed or remapped REQ
// IDs, deleted files) must not survive a re-index.
// CANARY: REQ=CP-285; FEATURE="IndexRebuild"; ASPECT=Storage; STATUS=TESTED; TEST=TestCANARY_CP_285_IndexRebuildPrunes; UPDATED=2026-08-29
func (db *DB) DeleteAllTokens() error {
	_, err := db.conn.Exec(`DELETE FROM tokens`)
	return err
}
