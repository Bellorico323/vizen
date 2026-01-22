-- name: CheckBookingConflict :one
SELECT EXISTS (
  SELECT 1
  FROM bookings
  WHERE common_area_id = $1
  AND deleted_at IS NULL
  AND status IN ('confirmed', 'pending')
  AND (
    (starts_at < sqlc.arg('ends_at') AND ends_at > sqlc.arg('starts_at'))
  )
);

-- name: CreateBooking :one
INSERT INTO bookings (
  condominium_id,
  apartment_id,
  user_id,
  common_area_id,
  status,
  starts_at,
  ends_at
) VALUES (
  $1,
  $2,
  $3,
  $4,
  $5,
  $6,
  $7
) RETURNING *;

-- name: GetBookingById :one
SELECT
  b.*,
  ca.name AS common_area_name
FROM bookings b
JOIN common_areas ca ON ca.id = b.common_area_id
WHERE b.id = $1;

-- name: UpdateBookingStatus :one
UPDATE bookings
SET
  status = $1,
  updated_at = NOW()
WHERE id = $2
RETURNING *;
