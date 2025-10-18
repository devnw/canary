-- CANARY: REQ=CBIN-140; FEATURE="GapDatabaseSchema"; ASPECT=Storage; STATUS=IMPL; UPDATED=2025-10-17
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
