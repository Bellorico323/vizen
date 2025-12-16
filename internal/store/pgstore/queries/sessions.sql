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
