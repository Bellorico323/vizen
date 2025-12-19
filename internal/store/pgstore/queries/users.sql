-- name: GetUserByEmail :one
SELECT *
FROM users
WHERE email = $1;

-- name: CreateUser :one
INSERT INTO users (
  name,
  avatar_url,
  email
)
VALUES (
  $1,
  $2,
  $3
) RETURNING id;

-- name: GetUserByID :one
SELECT *
FROM users
WHERE id = $1;
