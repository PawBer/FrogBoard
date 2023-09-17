BEGIN;
CREATE TABLE IF NOT EXISTS public.users (
    username VARCHAR(255) PRIMARY KEY,
    password_hash VARCHAR(255) NOT NULL,
    display_name VARCHAR(255) NOT NULL,
    permission VARCHAR(255) NOT NULL
);
COMMIT;