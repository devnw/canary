-- CANARY: REQ=CBIN-123; FEATURE="TokenStorage"; ASPECT=Storage; STATUS=IMPL; OWNER=canary; UPDATED=2025-10-16
-- Rollback initial schema

DROP TABLE IF EXISTS search_history;
DROP TABLE IF EXISTS checkpoints;
DROP TABLE IF EXISTS tokens;
