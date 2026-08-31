-- ===== 000001_initial_schema.up.sql =====
-- CANARY: REQ=ENG-4306; FEATURE="TokenStorage"; ASPECT=Storage; STATUS=IMPL; OWNER=canary; UPDATED=2025-10-16
-- Initial schema for CANARY token storage

CREATE TABLE IF NOT EXISTS tokens (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    req_id TEXT NOT NULL,
    feature TEXT NOT NULL,
    aspect TEXT NOT NULL,
    status TEXT NOT NULL,

    -- File location
    file_path TEXT NOT NULL,
    line_number INTEGER NOT NULL,

    -- Optional fields
    test TEXT,
    bench TEXT,
    owner TEXT,

    -- Extended metadata
    priority INTEGER DEFAULT 5,  -- 1 (highest) to 10 (lowest)
    phase TEXT,  -- Phase0, Phase1, Phase2, Phase3
    keywords TEXT,  -- Comma-separated tags

    -- Spec lifecycle
    spec_status TEXT DEFAULT 'draft',  -- draft, approved, in-progress, completed, archived

    -- Dates
    created_at TEXT,
    updated_at TEXT NOT NULL,
    started_at TEXT,
    completed_at TEXT,

    -- Git integration
    commit_hash TEXT,
    branch TEXT,

    -- Relationships
    depends_on TEXT,  -- Comma-separated REQ-IDs this depends on
    blocks TEXT,      -- Comma-separated REQ-IDs this blocks
    related_to TEXT,  -- Comma-separated REQ-IDs related to this

    -- Full token content for reference
    raw_token TEXT NOT NULL,

    -- Timestamps
    indexed_at TEXT NOT NULL,

    UNIQUE(req_id, feature, file_path, line_number)
);

CREATE INDEX IF NOT EXISTS idx_tokens_req_id ON tokens(req_id);
CREATE INDEX IF NOT EXISTS idx_tokens_status ON tokens(status);
CREATE INDEX IF NOT EXISTS idx_tokens_priority ON tokens(priority);
CREATE INDEX IF NOT EXISTS idx_tokens_aspect ON tokens(aspect);
CREATE INDEX IF NOT EXISTS idx_tokens_spec_status ON tokens(spec_status);
CREATE INDEX IF NOT EXISTS idx_tokens_phase ON tokens(phase);
CREATE INDEX IF NOT EXISTS idx_tokens_keywords ON tokens(keywords);

-- Checkpoints for state snapshots
CREATE TABLE IF NOT EXISTS checkpoints (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL UNIQUE,
    description TEXT,
    commit_hash TEXT,
    created_at TEXT NOT NULL,

    -- Summary stats at checkpoint time
    total_tokens INTEGER,
    stub_count INTEGER,
    impl_count INTEGER,
    tested_count INTEGER,
    benched_count INTEGER,

    -- Snapshot data (JSON)
    snapshot_json TEXT NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_checkpoints_created_at ON checkpoints(created_at);

-- Search history for analytics
CREATE TABLE IF NOT EXISTS search_history (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    query TEXT NOT NULL,
    filters TEXT,  -- JSON of filters used
    results_count INTEGER,
    searched_at TEXT NOT NULL
);

-- ===== 000002_extend_reqid.up.sql =====
-- CANARY: REQ=ENG-4316; FEATURE="AspectBasedReqIDSchema"; ASPECT=Storage; STATUS=IMPL; UPDATED=2025-10-16
-- Extend req_id to support aspect-based format: CBIN-SECURITY_REVIEW-XXX

-- SQLite TEXT type already supports variable length strings
-- This migration documents the schema extension for aspect-based IDs
-- Old format: CBIN-XXX (8-9 chars)
-- New format: CBIN-SECURITY_REVIEW-XXX (up to ~25 chars, e.g., CBIN-RoundTrip-001)

-- Add a comment to the schema (SQLite doesn't support ALTER TABLE COMMENT)
-- No actual ALTER needed since TEXT columns are already unlimited

-- Note: Both old (CBIN-XXX) and new (CBIN-SECURITY_REVIEW-XXX) formats are now valid
-- The application layer handles backward compatibility via internal/reqid package

-- ===== 000003_add_documentation_fields.up.sql =====
-- CANARY: REQ=ENG-4315; FEATURE="DocDatabaseSchema"; ASPECT=Storage; STATUS=IMPL; UPDATED=2025-10-16
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

-- ===== 000004_create_gap_tables.up.sql =====
-- CANARY: REQ=ENG-4317; FEATURE="GapDatabaseSchema"; ASPECT=Storage; STATUS=IMPL; UPDATED=2025-10-17
-- Create gap analysis tracking tables

-- Gap categories (predefined categories for classification)
CREATE TABLE IF NOT EXISTS gap_categories (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL UNIQUE,
    description TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Gap entries (individual gap analysis records)
CREATE TABLE IF NOT EXISTS gap_entries (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    gap_id TEXT NOT NULL UNIQUE,
    req_id TEXT NOT NULL,
    feature TEXT NOT NULL,
    aspect TEXT,
    category_id INTEGER NOT NULL,
    description TEXT NOT NULL,
    corrective_action TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    created_by TEXT DEFAULT 'unknown',
    helpful_count INTEGER DEFAULT 0,
    unhelpful_count INTEGER DEFAULT 0,
    FOREIGN KEY (category_id) REFERENCES gap_categories(id)
);

-- Gap configuration (settings for gap analysis behavior)
CREATE TABLE IF NOT EXISTS gap_config (
    id INTEGER PRIMARY KEY CHECK (id = 1),
    max_gap_injection INTEGER DEFAULT 10,
    min_helpful_threshold INTEGER DEFAULT 1,
    ranking_strategy TEXT DEFAULT 'helpful_desc',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Create indices for efficient queries
CREATE INDEX IF NOT EXISTS idx_gap_entries_req_id ON gap_entries(req_id);
CREATE INDEX IF NOT EXISTS idx_gap_entries_feature ON gap_entries(feature);
CREATE INDEX IF NOT EXISTS idx_gap_entries_category ON gap_entries(category_id);
CREATE INDEX IF NOT EXISTS idx_gap_entries_helpful ON gap_entries(helpful_count DESC);
CREATE INDEX IF NOT EXISTS idx_gap_entries_created ON gap_entries(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_gap_entries_composite ON gap_entries(helpful_count DESC, created_at DESC);

-- Insert default categories
INSERT INTO gap_categories (name, description) VALUES
    ('logic_error', 'Incorrect business logic or algorithm implementation'),
    ('test_failure', 'Tests incorrectly written or missing critical test cases'),
    ('performance', 'Performance issues or inefficient implementation'),
    ('security', 'Security vulnerabilities or insecure practices'),
    ('edge_case', 'Unhandled edge cases or boundary conditions'),
    ('integration', 'Integration issues with existing systems'),
    ('documentation', 'Incorrect or misleading documentation'),
    ('other', 'Other types of implementation gaps');

-- Insert default configuration
INSERT INTO gap_config (id, max_gap_injection, min_helpful_threshold, ranking_strategy)
VALUES (1, 10, 1, 'helpful_desc');

-- Comments:
-- gap_id: Unique identifier like "GAP-CBIN-140-001"
-- req_id: CANARY requirement ID (e.g., "CBIN-140")
-- feature: Feature name from CANARY token
-- aspect: Implementation aspect (API, CLI, Storage, etc.)
-- category_id: Foreign key to gap_categories
-- description: What went wrong
-- corrective_action: How it was fixed
-- helpful_count: Number of times marked as helpful
-- unhelpful_count: Number of times marked as unhelpful
-- ranking_strategy: Options: helpful_desc, recency_desc, weighted

-- ===== 000005_add_multi_project_support.up.sql =====
-- CANARY: REQ=ENG-4319; FEATURE="MultiProjectSupport"; ASPECT=Storage; STATUS=IMPL; UPDATED=2025-10-18
-- Add multi-project support to CANARY database

-- Create projects table for project registry
CREATE TABLE IF NOT EXISTS projects (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    path TEXT NOT NULL UNIQUE,
    active BOOLEAN DEFAULT FALSE,
    created_at TEXT NOT NULL,
    metadata TEXT
);

CREATE INDEX IF NOT EXISTS idx_projects_path ON projects(path);
CREATE INDEX IF NOT EXISTS idx_projects_active ON projects(active);

-- Recreate tokens table with project_id and updated unique constraint
-- SQLite doesn't support modifying UNIQUE constraints, so we need to recreate the table

-- Step 1: Rename existing tokens table
ALTER TABLE tokens RENAME TO tokens_old;

-- Step 2: Create new tokens table with project_id and updated unique constraint
CREATE TABLE tokens (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    req_id TEXT NOT NULL,
    feature TEXT NOT NULL,
    aspect TEXT NOT NULL,
    status TEXT NOT NULL,

    -- File location
    file_path TEXT NOT NULL,
    line_number INTEGER NOT NULL,

    -- Optional fields
    test TEXT,
    bench TEXT,
    owner TEXT,

    -- Extended metadata
    priority INTEGER DEFAULT 5,
    phase TEXT,
    keywords TEXT,

    -- Spec lifecycle
    spec_status TEXT DEFAULT 'draft',

    -- Dates
    created_at TEXT,
    updated_at TEXT NOT NULL,
    started_at TEXT,
    completed_at TEXT,

    -- Git integration
    commit_hash TEXT,
    branch TEXT,

    -- Relationships
    depends_on TEXT,
    blocks TEXT,
    related_to TEXT,

    -- Full token content for reference
    raw_token TEXT NOT NULL,

    -- Timestamps
    indexed_at TEXT NOT NULL,

    -- Documentation fields
    doc_path TEXT,
    doc_hash TEXT,
    doc_type TEXT,
    doc_checked_at TEXT,
    doc_status TEXT,

    -- Multi-project support (NEW)
    project_id TEXT DEFAULT '',

    -- Updated unique constraint including project_id
    UNIQUE(req_id, feature, file_path, line_number, project_id)
);

-- Step 3: Copy data from old table to new table (with default project_id = '')
INSERT INTO tokens (
    id, req_id, feature, aspect, status,
    file_path, line_number,
    test, bench, owner,
    priority, phase, keywords,
    spec_status,
    created_at, updated_at, started_at, completed_at,
    commit_hash, branch,
    depends_on, blocks, related_to,
    raw_token, indexed_at,
    doc_path, doc_hash, doc_type, doc_checked_at, doc_status,
    project_id
)
SELECT
    id, req_id, feature, aspect, status,
    file_path, line_number,
    test, bench, owner,
    priority, phase, keywords,
    spec_status,
    created_at, updated_at, started_at, completed_at,
    commit_hash, branch,
    depends_on, blocks, related_to,
    raw_token, indexed_at,
    doc_path, doc_hash, doc_type, doc_checked_at, doc_status,
    '' -- Default project_id for existing tokens
FROM tokens_old;

-- Step 4: Drop old tokens table
DROP TABLE tokens_old;

-- Step 5: Recreate all indexes
CREATE INDEX IF NOT EXISTS idx_tokens_req_id ON tokens(req_id);
CREATE INDEX IF NOT EXISTS idx_tokens_status ON tokens(status);
CREATE INDEX IF NOT EXISTS idx_tokens_priority ON tokens(priority);
CREATE INDEX IF NOT EXISTS idx_tokens_aspect ON tokens(aspect);
CREATE INDEX IF NOT EXISTS idx_tokens_spec_status ON tokens(spec_status);
CREATE INDEX IF NOT EXISTS idx_tokens_phase ON tokens(phase);
CREATE INDEX IF NOT EXISTS idx_tokens_keywords ON tokens(keywords);
CREATE INDEX IF NOT EXISTS idx_tokens_project_id ON tokens(project_id);

-- ===== 000006_add_refs_table.up.sql =====
-- Requirement references found outside CANARY tokens (e.g. mermaid diagrams).
-- kind: 'diagram' today; future kinds may include 'doc', 'adr'.
CREATE TABLE IF NOT EXISTS refs (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    req_id TEXT NOT NULL,
    kind TEXT NOT NULL DEFAULT 'diagram',
    file_path TEXT NOT NULL,
    line_number INTEGER NOT NULL DEFAULT 0,
    context TEXT DEFAULT '',
    indexed_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(req_id, kind, file_path, line_number)
);

CREATE INDEX IF NOT EXISTS idx_refs_req_id ON refs(req_id);
CREATE INDEX IF NOT EXISTS idx_refs_kind ON refs(kind);

-- ===== 000007_project_scope_and_index_meta.up.sql =====
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

-- ===== 000008_id_reservations.up.sql =====
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
