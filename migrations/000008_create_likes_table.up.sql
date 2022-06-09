CREATE TABLE IF NOT EXISTS likes (
    id bigserial PRIMARY KEY,
    movie_id bigint NOT NULL REFERENCES movies ON DELETE CASCADE,
    user_id bigint NOT NULL REFERENCES users ON DELETE CASCADE
);