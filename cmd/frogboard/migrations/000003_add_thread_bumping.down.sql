BEGIN;
ALTER TABLE threads
DROP COLUMN last_bump;
ALTER TABLE threads
DROP COLUMN post_count;
ALTER TABLE boards
DROP COLUMN bump_limit;
COMMIT;