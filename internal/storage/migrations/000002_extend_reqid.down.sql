-- CANARY: REQ=CBIN-139; FEATURE="AspectBasedReqIDSchema"; ASPECT=Storage; STATUS=IMPL; UPDATED=2025-10-16
-- Rollback aspect-based req_id extension

-- No actual schema changes to rollback since TEXT columns were already unlimited
-- This migration is informational only

-- Warning: Rolling back this migration means the application will no longer
-- properly support aspect-based requirement IDs (CBIN-<ASPECT>-XXX format)
-- Any tokens with new format IDs will need to be migrated back manually
