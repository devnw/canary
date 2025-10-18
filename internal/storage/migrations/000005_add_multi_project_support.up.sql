-- CANARY: REQ=CBIN-146; FEATURE="MultiProjectSupport"; ASPECT=Storage; STATUS=IMPL; UPDATED=2025-10-18
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
