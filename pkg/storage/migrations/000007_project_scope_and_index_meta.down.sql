-- Reverses 000007.
--
-- index_meta and tokens.content_hash are both removed, so the up migration
-- can be applied again afterwards -- a down migration that leaves the column
-- behind makes `ADD COLUMN content_hash` fail on the next up, which turns a
-- rollback into a one-way door.
--
-- The tokens.project_id backfill is deliberately NOT reversed. Its
-- pre-migration value was the empty string, which carries no less information
-- than 'default' and which the up migration would immediately rewrite again;
-- restoring it would only re-create the ambiguity the migration removed.
DROP TABLE IF EXISTS index_meta;

ALTER TABLE tokens DROP COLUMN content_hash;
