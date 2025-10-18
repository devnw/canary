-- CANARY: REQ=CBIN-140; FEATURE="GapDatabaseSchema"; ASPECT=Storage; STATUS=IMPL; UPDATED=2025-10-17
-- Rollback gap analysis tracking tables

-- Drop indices first
DROP INDEX IF EXISTS idx_gap_entries_composite;
DROP INDEX IF EXISTS idx_gap_entries_created;
DROP INDEX IF EXISTS idx_gap_entries_helpful;
DROP INDEX IF EXISTS idx_gap_entries_category;
DROP INDEX IF EXISTS idx_gap_entries_feature;
DROP INDEX IF EXISTS idx_gap_entries_req_id;

-- Drop tables in reverse order of dependencies
DROP TABLE IF EXISTS gap_config;
DROP TABLE IF EXISTS gap_entries;
DROP TABLE IF EXISTS gap_categories;
