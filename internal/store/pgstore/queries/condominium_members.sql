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
