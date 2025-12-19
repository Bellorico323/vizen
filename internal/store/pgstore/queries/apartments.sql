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
