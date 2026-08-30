// Copyright (c) 2025 by Developer Network.
//
// For more details, see the LICENSE file in the root directory of this
// source code repository or contact Developer Network at info@devnw.com.

package storage

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
)

// ErrReadOnlyReservation is returned when a reservation is attempted through
// a read-only handle. Allocating an id is a write; asking for one from a
// reader is a caller mistake, not a transient failure.
var ErrReadOnlyReservation = errors.New("cannot reserve an id through a read-only database handle")

// reserveAttempts bounds the retry loop. Each attempt loses only to a racing
// writer that took the same number first, and every loss raises the maximum
// the next attempt reads, so the loop terminates in practice long before this
// ceiling. The ceiling exists so a pathological database (a corrupt
// reservation row, a clock-stopped writer) fails loudly instead of spinning.
const reserveAttempts = 100

// idNumberWidth is the zero-padded width of a reserved number: BUG-API-007,
// not BUG-API-7. It matches the format `canary bug create` has always
// printed, so ids minted before and after this table are the same shape.
const idNumberWidth = 3

// CANARY: REQ=CP-282; FEATURE="TransactionalIDs"; ASPECT=Storage; STATUS=TESTED; TEST=TestAuditF18,TestAuditF18SeedsPastExistingTokens,TestAuditF18RefusesUnscopedReservation,TestReserveIDConcurrent; UPDATED=2026-08-30

// ReserveID allocates the next identifier in a series and records the
// allocation, returning e.g. "BUG-API-007".
//
// The allocation is a single immediate transaction: the write lock is taken
// at BEGIN (OpenRW sets _txlock=immediate), the highest number in the series
// is read, and the successor is INSERTed. The table's primary key is the
// actual guarantee -- if a concurrent writer commits the same number first,
// this transaction's INSERT is rejected and the whole read-then-write is
// retried against the new maximum. No caller can be handed an id that another
// caller already holds.
//
// The series is seeded from the token index the first time it is used, so a
// database that predates this table does not start re-issuing ids its own
// rows already carry.
//
// projectID is required. Numbering is per project, so an unscoped reservation
// would allocate out of whatever mixture of projects the database happens to
// hold and hand two projects the same id.
func (db *DB) ReserveID(projectID, prefix string) (string, error) {
	if db.readOnly {
		return "", ErrReadOnlyReservation
	}
	projectID = strings.TrimSpace(projectID)
	if projectID == "" {
		return "", fmt.Errorf("%w: reserving an id requires a project", ErrProjectRequired)
	}
	prefix = strings.TrimSpace(prefix)
	if prefix == "" {
		return "", errors.New("reserving an id requires a prefix")
	}

	var lastErr error
	for attempt := 0; attempt < reserveAttempts; attempt++ {
		id, err := db.reserveOnce(projectID, prefix)
		if err == nil {
			return id, nil
		}
		if !retryableReservation(err) {
			return "", err
		}
		lastErr = err
	}
	return "", fmt.Errorf("reserve %s id after %d attempts: %w", prefix, reserveAttempts, lastErr)
}

// reserveOnce performs one read-then-insert inside a single immediate
// transaction. Its error is returned verbatim so ReserveID can tell a lost
// race (retry) from a real failure (report).
func (db *DB) reserveOnce(projectID, prefix string) (id string, err error) {
	tx, err := db.conn.Beginx()
	if err != nil {
		return "", err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	next, err := nextNumber(tx, projectID, prefix)
	if err != nil {
		return "", err
	}

	_, err = tx.Exec(
		`INSERT INTO id_reservations (project_id, prefix, num, reserved_at) VALUES (?, ?, ?, ?)`,
		projectID, prefix, next, time.Now().UTC().Format(time.RFC3339),
	)
	if err != nil {
		return "", err
	}

	if err = tx.Commit(); err != nil {
		return "", err
	}
	return fmt.Sprintf("%s-%0*d", prefix, idNumberWidth, next), nil
}

// nextNumber is the successor of the highest number this series has ever
// used, counting both recorded reservations and ids already carried by
// indexed tokens.
//
// The token scan is what keeps the first reservation in an existing database
// from colliding with history. It is not merely a fallback: an id can also
// enter the index from source (a token written by hand, or one indexed from
// another checkout) without ever passing through this table.
func nextNumber(tx *sqlx.Tx, projectID, prefix string) (int, error) {
	var reserved int
	if err := tx.Get(&reserved,
		`SELECT COALESCE(MAX(num), 0) FROM id_reservations WHERE project_id = ? AND prefix = ?`,
		projectID, prefix,
	); err != nil {
		return 0, fmt.Errorf("read reserved ids: %w", err)
	}

	var ids []string
	if err := tx.Select(&ids,
		`SELECT req_id FROM tokens WHERE project_id = ? AND req_id GLOB ?`,
		projectID, prefix+"-[0-9]*",
	); err != nil {
		return 0, fmt.Errorf("read indexed ids: %w", err)
	}

	max := reserved
	for _, id := range ids {
		if n, ok := idNumber(id, prefix); ok && n > max {
			max = n
		}
	}
	return max + 1, nil
}

// idNumber extracts the numeric tail of "PREFIX-007". A tail that is not all
// digits belongs to a different series that merely shares the prefix
// (BUG-API-1-alt), and is ignored rather than guessed at.
func idNumber(id, prefix string) (int, bool) {
	tail, ok := strings.CutPrefix(id, prefix+"-")
	if !ok || tail == "" {
		return 0, false
	}
	n, err := strconv.Atoi(tail)
	if err != nil || n < 0 {
		return 0, false
	}
	return n, true
}

// retryableReservation reports whether an error means "someone else got there
// first, read again" rather than "this will not work".
//
// modernc's driver reports both conditions as plain messages, so they are
// matched by text. Matching too broadly would turn a genuine failure into a
// hundred silent retries and then a confusing timeout, so only the two known
// contention conditions qualify.
func retryableReservation(err error) bool {
	if err == nil {
		return false
	}
	msg := strings.ToLower(err.Error())
	switch {
	case strings.Contains(msg, "unique constraint failed"),
		strings.Contains(msg, "constraint failed: id_reservations"),
		strings.Contains(msg, "database is locked"),
		strings.Contains(msg, "database table is locked"),
		strings.Contains(msg, "sqlite_busy"):
		return true
	default:
		return false
	}
}
