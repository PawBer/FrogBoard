BEGIN;
CREATE TABLE IF NOT EXISTS public.citations (
    id SERIAL NOT NULL PRIMARY KEY,
    board_id VARCHAR(100) NOT NULL,
    post_id INT NOT NULL,
    cites INT NOT NULL
);

CREATE TABLE IF NOT EXISTS public.file_infos (
    id VARCHAR(255) NOT NULL PRIMARY KEY,
    file_name VARCHAR(255) NOT NULL,
    content_type VARCHAR(255) NOT NULL
);

CREATE TABLE IF NOT EXISTS public.post_files (
    id SERIAL NOT NULL PRIMARY KEY,
    board_id VARCHAR(100) NOT NULL,
    post_id INT NOT NULL,
    file_id VARCHAR(255) NOT NULL
);

CREATE TABLE IF NOT EXISTS public.threads (
    board_id TEXT NOT NULL,
    id INT NOT NULL,
    created_at TIMESTAMP NOT NULL,
    title VARCHAR(255) NOT NULL,
    content TEXT NOT NULL,
    PRIMARY KEY(board_id, id)
);

CREATE TABLE IF NOT EXISTS public.replies (
    board_id TEXT NOT NULL,
    id INT NOT NULL,
    created_at TIMESTAMP NOT NULL,
    thread_id INT NOT NULL,
    content TEXT NOT NULL,
    PRIMARY KEY(board_id, id)
);
COMMIT;
