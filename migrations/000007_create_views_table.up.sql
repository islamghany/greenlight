CREATE TABLE IF NOT EXISTS views(
    count integer NOT NULL DEFAULT 0,
    movie_id bigint NOT NULL REFERENCES movies ON DELETE CASCADE
);