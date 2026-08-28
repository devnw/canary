// Copyright (c) 2025 by Developer Network.
//
// For more details, see the LICENSE file in the root directory of this
// source code repository or contact Developer Network at info@devnw.com.

package storage

// CANARY: REQ=CBIN-206; FEATURE="DiagramRefsIndex"; ASPECT=Storage; STATUS=TESTED; TEST=TestCANARY_CBIN_206_RefsRoundTrip; UPDATED=2026-08-28

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
