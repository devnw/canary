-- CANARY: REQ=CBIN-136; FEATURE="DocDatabaseSchema"; ASPECT=Storage; STATUS=IMPL; UPDATED=2025-10-16
-- Add documentation tracking fields to tokens table

-- Add documentation tracking columns
ALTER TABLE tokens ADD COLUMN doc_path TEXT DEFAULT '';
ALTER TABLE tokens ADD COLUMN doc_hash TEXT DEFAULT '';
ALTER TABLE tokens ADD COLUMN doc_type TEXT DEFAULT '';
ALTER TABLE tokens ADD COLUMN doc_checked_at TEXT DEFAULT '';
ALTER TABLE tokens ADD COLUMN doc_status TEXT DEFAULT '';

-- Create index for efficient documentation staleness queries
CREATE INDEX IF NOT EXISTS idx_tokens_doc_status ON tokens(doc_status);
CREATE INDEX IF NOT EXISTS idx_tokens_doc_checked_at ON tokens(doc_checked_at);

-- Comments:
-- doc_path: Comma-separated doc file paths (e.g., "user:docs/user.md,api:docs/api.md")
-- doc_hash: Comma-separated SHA256 hashes (abbreviated, first 16 chars)
-- doc_type: Documentation type (user, technical, feature, api, architecture)
-- doc_checked_at: ISO 8601 timestamp of last staleness check
-- doc_status: DOC_CURRENT, DOC_STALE, DOC_MISSING, DOC_UNHASHED
