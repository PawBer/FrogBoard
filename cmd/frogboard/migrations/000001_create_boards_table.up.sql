BEGIN;
CREATE TABLE IF NOT EXISTS public.boards (
    id VARCHAR(100) NOT NULL,
    full_name VARCHAR(255),
    last_post_id INT
);

INSERT INTO public.boards (id, full_name, last_post_id) VALUES ('b', 'Random', '0') ON CONFLICT DO NOTHING;
COMMIT;
