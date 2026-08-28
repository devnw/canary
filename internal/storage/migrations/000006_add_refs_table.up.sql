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
