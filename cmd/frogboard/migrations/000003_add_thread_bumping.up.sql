BEGIN;
ALTER TABLE threads
ADD last_bump TIMESTAMP;
ALTER TABLE threads
ADD post_count INT;
ALTER TABLE boards
ADD bump_limit INT;
COMMIT;