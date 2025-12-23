-- name: CreateAccessRequest :one
INSERT INTO access_requests (
  user_id,
  condominium_id,
  apartment_id,
  type
) VALUES (
  $1,
  $2,
  $3,
  $4
) RETURNING id;

-- name: ListPendingRequestsByCondo :many
SELECT
  ar.*,
  u.name,
  u.email,
  a.block,
  a.number
FROM access_requests ar
JOIN users u ON u.id = ar.user_id
JOIN apartments a ON a.id = ar.apartment_id
WHERE ar.condominium_id = $1 AND ar.status = 'pending';

-- name: UpdateAccessRequestStatus :exec
UPDATE access_requests
SET status = $2,
    reviewed_by = $3,
    reviewed_at = NOW(),
    updated_at = NOW()
WHERE id = $1;

-- name: GetAccessRequestById :one
SELECT
  *
FROM access_requests
WHERE id = $1;
