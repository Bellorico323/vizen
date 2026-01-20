-- name: CreateInvite :one
INSERT INTO invites (
  condominium_id,
  apartment_id,
  issued_by,
  guest_name,
  guest_type,
  starts_at,
  ends_at
) VALUES (
  $1,
  $2,
  $3,
  $4,
  $5,
  $6,
  $7
) RETURNING *;

-- name: GetInviteByToken :one
SELECT
  i.*,
  a.block,
  a.number AS apartment_number,
  u.name AS resident_name
FROM invites i
JOIN apartments a ON a.id = i.apartment_id
JOIN users u ON u.id = i.issued_by
WHERE i.token = $1;

-- name: GetInviteById :one
SELECT
  *
FROM invites
WHERE id = $1;

-- name: ListInvites :many
SELECT
    i.*,
    a.block,
    a.number as apartment_number
FROM invites i
JOIN apartments a ON a.id = i.apartment_id
WHERE i.condominium_id = $1
  AND (sqlc.narg('apartment_id')::uuid IS NULL OR i.apartment_id = sqlc.narg('apartment_id'))
  AND (
      sqlc.narg('only_active')::boolean IS NULL
      OR (sqlc.narg('only_active')::boolean = TRUE AND i.ends_at > NOW() AND i.revoked_at IS NULL)
  )
ORDER BY i.created_at DESC
LIMIT $2 OFFSET $3;

-- name: RevokeInvite :exec
UPDATE invites
SET revoked_at = NOW()
WHERE id = $1 AND issued_by = $2;
