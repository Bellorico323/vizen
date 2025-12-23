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
