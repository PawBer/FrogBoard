BEGIN;
CREATE TABLE IF NOT EXISTS public.boards (
    id VARCHAR(100) NOT NULL,
    full_name VARCHAR(255),
    last_post_id INT,
    bump_limit INT
);

INSERT INTO public.boards (id, full_name, last_post_id) VALUES ('b', 'Random', '0', 5) ON CONFLICT DO NOTHING;
COMMIT;
