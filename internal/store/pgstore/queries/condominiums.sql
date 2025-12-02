-- name: CreateCondominium :exec
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
);
