-- CANARY: REQ=CBIN-123; FEATURE="TokenStorage"; ASPECT=Storage; STATUS=IMPL; OWNER=canary; UPDATED=2025-10-16
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
