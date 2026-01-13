-- name: CreateApartment :one
INSERT INTO apartments (
  condominium_id,
  block,
  number
) VALUES (
  $1,
  $2,
  $3
) RETURNING id;

-- name: GetApartmentById :one
SELECT
  *
FROM apartments
WHERE id = $1;

-- name: GetApartmentsByUserId :many
SELECT
  a.id as apartment_id,
  a.block,
  a.number,
  r.type,
  r.is_responsible,
  c.name as condominium_name,
  c.id as condominium_id
FROM apartments a
JOIN residents r ON r.apartment_id = a.id
JOIN condominiums c ON c.id = a.condominium_id
WHERE (sqlc.narg('condominium_id')::uuid IS NULL OR a.condominium_id = sqlc.narg('condominium_id'))
  AND r.user_id = @user_id
ORDER BY c.name, a.block, a.number;
