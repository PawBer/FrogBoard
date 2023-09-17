BEGIN;
CREATE TABLE IF NOT EXISTS public.boards (
    id VARCHAR(100) PRIMARY KEY NOT NULL,
    full_name VARCHAR(255) NOT NULL,
    last_post_id INT NOT NULL,
    bump_limit INT NOT NULL
);
COMMIT;
