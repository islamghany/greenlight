-- name: CreateUser :one
INSERT INTO users (
  name,
  hashed_password,
  email,
  activated
) VALUES (
  $1, $2, $3, $4
) RETURNING id, created_at, version;



-- name: GetUserByEmail :one
SELECT *
FROM users
WHERE email = $1
LIMIT 1;


-- name: GetUserByID :one
SELECT *
FROM users
WHERE id = $1
LIMIT 1;



-- name: UpdateUser :one
UPDATE users
SET
  hashed_password = COALESCE(sqlc.narg(hashed_password), hashed_password),
  password_changed_at = COALESCE(sqlc.narg(password_changed_at), password_changed_at),
  name = COALESCE(sqlc.narg(name), name),
  activated = COALESCE(sqlc.narg(activated), activated),
  version = version + 1
WHERE
  id = sqlc.arg(id) AND
  version = sqlc.arg(version)
RETURNING *;   