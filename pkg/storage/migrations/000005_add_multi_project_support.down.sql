-- CANARY: REQ=CBIN-146; FEATURE="MultiProjectSupport"; ASPECT=Storage; STATUS=IMPL; UPDATED=2025-10-18
-- Rollback multi-project support

-- Drop project-related index
DROP INDEX IF EXISTS idx_tokens_project_id;

-- SQLite doesn't support DROP COLUMN directly
-- To remove project_id, we would need to:
-- 1. Create a new table without project_id
-- 2. Copy data
-- 3. Drop old table
-- 4. Rename new table
-- This is complex and risky, so we leave the column in place but unused

-- Drop projects table and related indexes
DROP INDEX IF EXISTS idx_projects_active;
DROP INDEX IF EXISTS idx_projects_path;
DROP TABLE IF EXISTS projects;
