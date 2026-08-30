-- Project scoping and index metadata.
--
-- index_meta records what the token index was built from, so a reader can
-- tell a fresh index from a stale one instead of trusting whatever rows
-- happen to be present. It holds exactly one row (id = 1).
--
-- The project_id backfill gives every pre-existing row the same identity the
-- CLI resolves for an unconfigured repository ("default"), so a database
-- written before project scoping stays queryable after it.
--
-- content_hash records the digest of the file each token was read from, so a
-- token row can be checked against the file on disk without re-scanning. The
-- down migration drops it again, so this ADD COLUMN is safe to re-apply.
CREATE TABLE IF NOT EXISTS index_meta (
    id INTEGER PRIMARY KEY CHECK (id = 1),
    root TEXT NOT NULL,
    project_id TEXT NOT NULL,
    commit_sha TEXT NOT NULL DEFAULT '',
    parser_schema INTEGER NOT NULL,
    scan_digest TEXT NOT NULL,
    indexed_at TEXT NOT NULL
);

UPDATE tokens SET project_id = 'default' WHERE project_id = '' OR project_id IS NULL;

ALTER TABLE tokens ADD COLUMN content_hash TEXT NOT NULL DEFAULT '';
