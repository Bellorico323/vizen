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

-- name: ListCondominiunsByUserId :many
SELECT
  c.id,
  c.name,
  c.cnpj,
  c.address,
  m.role::text as role_name,
  1 as priority
FROM condominium_members m
JOIN condominiums c ON c.id = m.condominium_id
WHERE m.user_id = $1

UNION ALL

SELECT
  c.id,
  c.name,
  c.cnpj,
  c.address,
  r.type as role_name,
  2 as priority
FROM residents r
JOIN apartments a ON a.id = r.apartment_id
JOIN condominiums c ON c.id = a.condominium_id
WHERE r.user_id = $1;
