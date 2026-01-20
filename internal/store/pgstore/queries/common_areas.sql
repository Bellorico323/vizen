-- name: CreateCommonArea :one
INSERT INTO common_areas (
  condominium_id,
  name,
  capacity,
  requires_approval
) VALUES (
  $1,
  $2,
  $3,
  $4
) RETURNING *;

-- name: ListCommonAreas :many
SELECT
  *
FROM common_areas
WHERE condominium_id = $1;

-- name: GetCommonAreaIdForUpdate :one
SELECT
  *
FROM common_areas
WHERE id = $1
FOR UPDATE;
