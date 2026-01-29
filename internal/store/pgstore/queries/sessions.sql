-- name: CreateSession :exec
INSERT INTO sessions (
  user_id,
  token,
  expires_at,
  ip_address,
  user_agent
) VALUES (
  $1,
  $2,
  $3,
  $4,
  $5
);

-- name: GetSessionByToken :one
SELECT *
FROM sessions
WHERE token = $1;

-- name: UpdateRefreshToken :exec
UPDATE sessions
SET token = $1,
    expires_at = $2,
    updated_at = NOW()
WHERE id = $3;

-- name: DeleteSession :exec
DELETE FROM sessions
WHERE token = $1;
