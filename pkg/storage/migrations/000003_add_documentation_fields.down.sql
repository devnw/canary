-- CANARY: REQ=CBIN-136; FEATURE="DocDatabaseSchema"; ASPECT=Storage; STATUS=IMPL; UPDATED=2025-10-16
-- Rollback documentation tracking fields from tokens table

-- Drop indices first
DROP INDEX IF EXISTS idx_tokens_doc_status;
DROP INDEX IF EXISTS idx_tokens_doc_checked_at;

-- Note: SQLite doesn't support DROP COLUMN in older versions
-- For production, this would require recreating the table
-- For development, recommend dropping and recreating the database

-- This is a placeholder for SQLite compatibility
-- In practice, run: rm .canary/canary.db && canary index
