-- name: CreateCondominium :one
INSERT INTO condominiums (
  name,
  cnpj,
  address,
  plan_type
)
VALUES (
  $1,
  $2,
  $3,
  $4
) RETURNING id;

-- name: GetCondominiumById :one
SELECT
*
FROM condominiums
WHERE id = $1;

-- name: GetCondominiumByAddress :one
SELECT
  *
FROM condominiums
WHERE address = $1;
