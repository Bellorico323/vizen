-- name: CreateResident :exec
INSERT INTO residents (
  user_id,
  apartment_id,
  type,
  is_responsible
) VALUES (
  $1,
  $2,
  $3,
  $4
);

-- name: GetResidencesByUserId :many
SELECT
  r.type as resident_type,
  r.is_responsible,
  a.id as apartment_id,
  a.block,
  a.number as apartement_number,
  c.id as condominium_id,
  c.name as condominium_name
FROM residents r
JOIN apartments a ON a.id = r.apartment_id
JOIN condominiums c ON c.id = a.condominium_id
WHERE r.user_id = $1;

-- name: GetCondoResidentsTokens :many
SELECT DISTINCT d.fcm_token
FROM user_devices d
JOIN residents r ON r.user_id = d.user_id
WHERE r.apartment_id IN (
    SELECT id FROM apartments WHERE condominium_id = $1
);

-- name: CheckUserAccessToCondo :one
SELECT EXISTS (
    SELECT 1 FROM residents r
    WHERE r.user_id = $1
    AND r.apartment_id IN (SELECT a.id FROM apartments a WHERE a.condominium_id = $2)
    UNION
    SELECT 1 FROM condominium_members m
    WHERE m.user_id = $1 AND m.condominium_id = $2
);

-- name: CheckIsResident :one
SElECT EXISTS (
  SELECT 1
  FROM residents
  WHERE user_id = $1
  AND apartment_id = $2
);
