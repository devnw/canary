-- Reserved requirement/bug identifiers.
--
-- Allocation used to be "read the maximum id, add one, write it", which is
-- two statements with a gap in the middle: two callers reading the same
-- maximum both wrote the same id. The primary key below is what makes the
-- allocation atomic -- a duplicate number cannot be inserted, so a racing
-- writer is rejected and retries against the new maximum instead of
-- overwriting the winner's row.
CREATE TABLE IF NOT EXISTS id_reservations (
    project_id  TEXT    NOT NULL,
    prefix      TEXT    NOT NULL,
    num         INTEGER NOT NULL,
    reserved_at TEXT    NOT NULL,
    PRIMARY KEY (project_id, prefix, num)
);
