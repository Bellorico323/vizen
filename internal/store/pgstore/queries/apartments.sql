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
