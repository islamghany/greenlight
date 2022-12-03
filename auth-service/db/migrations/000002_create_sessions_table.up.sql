Create table if not exists sessions(
    id uuid primary key,
    user_id bigint not null,
    refresh_token text not null,
    user_agent text not null,
    user_ip text not null,
    expires_at timestamptz not null,
    created_at timestamptz not null default (now())
);

Alter table "sessions"
Add CONSTRAINT "sessions_user_fk"
FOREIGN key ("user_id")
REFERENCES "users" ("id")
On DELETE CASCADE;