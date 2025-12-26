-- name: CreateCondominiumMember :exec
INSERT INTO condominium_members  (
  condominium_id,
  user_id,
  role
) VALUES (
  $1,
  $2,
  $3
);

-- name: GetCondominiumMemberRole :one
SELECT
  role
FROM condominium_members
WHERE condominium_id = $1
AND user_id = $2;

-- name: GetUserMemberships :many
SELECT
  m.role,
  c.id as condominium_id,
  c.name as condominium_name
FROM condominium_members m
JOIN condominiums c ON c.id = m.condominium_id
WHERE user_id = $1;
