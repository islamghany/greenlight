CREATE TABLE IF NOT EXISTS users (
    id bigserial PRIMARY KEY,
    created_at timestamp(0) with time zone NOT NULL DEFAULT NOW(),
    name text NOT NULL,
    email text UNIQUE NOT NULL,
    hashed_password bytea NOT NULL,
    activated bool NOT NULL,
    password_changed_at timestamptz NOT NULL DEFAULT('0001-01-01 00:00:00Z'),
    version integer NOT NULL DEFAULT 1
);