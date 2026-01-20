-- name: LogAccessEntry :one
INSERT INTO access_logs (
  invite_id,
  condominium_id,
  authorized_by
) VALUES (
  $1,
  $2,
  $3
) RETURNING *;
