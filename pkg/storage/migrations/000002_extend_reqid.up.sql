-- CANARY: REQ=CBIN-139; FEATURE="AspectBasedReqIDSchema"; ASPECT=Storage; STATUS=IMPL; UPDATED=2025-10-16
-- Extend req_id to support aspect-based format: CBIN-SECURITY_REVIEW-XXX

-- SQLite TEXT type already supports variable length strings
-- This migration documents the schema extension for aspect-based IDs
-- Old format: CBIN-XXX (8-9 chars)
-- New format: CBIN-SECURITY_REVIEW-XXX (up to ~25 chars, e.g., CBIN-RoundTrip-001)

-- Add a comment to the schema (SQLite doesn't support ALTER TABLE COMMENT)
-- No actual ALTER needed since TEXT columns are already unlimited

-- Note: Both old (CBIN-XXX) and new (CBIN-SECURITY_REVIEW-XXX) formats are now valid
-- The application layer handles backward compatibility via internal/reqid package
