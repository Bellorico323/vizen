-- name: CheckBookingConflict :one
SELECT EXISTS (
  SELECT 1
  FROM bookings
  WHERE common_area_id = $1
  AND deleted_at IS NULL
  AND (
    (starts_at < $3 AND ends_at > $2)
  )
);

-- name: CreateBooking :one
INSERT INTO bookings (
  condominium_id,
  apartment_id,
  user_id,
  common_area_id,
  starts_at,
  ends_at
) VALUES (
  $1,
  $2,
  $3,
  $4,
  $5,
  $6
) RETURNING *;
