-- name: CreatePackage :one
INSERT INTO packages (
  condominium_id,
  apartment_id,
  received_by,
  recipient_name,
  photo_url,
  status
) VALUES (
  $1,
  $2,
  $3,
  $4,
  $5,
  'pending'
) RETURNING *;

-- name: UpdatePackageToWithdrawn :one
UPDATE packages
SET
  status = 'withdrawn',
  withdrawn_at = NOW(),
  withdrawn_by = sqlc.narg('withdrawn_by')
WHERE id = $1
RETURNING *;

-- name: GetPackageById :one
SELECT
  p.*,
  a.block,
  a.number AS apartment_number,
  u.name AS received_by_name
FROM packages p
JOIN apartments a ON a.id = p.apartment_id
JOIN users u ON u.id = p.received_by
WHERE p.id = $1;

-- name: ListPackagesByCondominium :many
SELECT
  p.id,
  p.recipient_name,
  p.received_at,
  p.status,
  a.block,
  a.number as apartment_number,
  p.withdrawn_at,
  p.withdrawn_by
FROM packages p
JOIN apartments a ON a.id = p.apartment_id
WHERE p.condominium_id = $1
  AND (sqlc.narg('status')::text IS NULL OR p.status = sqlc.narg('status')::text)
ORDER BY p.received_at DESC;

-- name: ListPackagesByApartment :many
SELECT
  p.id,
  p.recipient_name,
  p.received_at,
  p.status,
  p.photo_url,
  u.name as received_by_name,
  p.withdrawn_at,
  p.withdrawn_by
FROM packages p
JOIN users u ON u.id = p.received_by
WHERE p.apartment_id = $1
  AND (sqlc.narg('status')::text IS NULL OR p.status = sqlc.narg('status')::text)
ORDER BY p.received_at DESC;
