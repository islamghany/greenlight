-- name: InsertSession :one
INSERT INTO sessions(
    id,
    user_id,
    refresh_token,
    user_agent,
    user_ip,
    expires_at
)
VALUES ($1,$2,$3,$4,$5,$6)
RETURNING *;

-- name: GetSession :one
select * from sessions
where id= $1 
LIMIT 1;

-- name: DeleteSession :exec
delete from sessions
where id = $1;