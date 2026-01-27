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

-- name: ListBookings :many
SELECT
  b.*,
  ca.name as common_area_name,
  u.name as user_name,
  a.number as apartment_number,
  a.block as apartment_block
FROM bookings b
JOIN common_areas ca ON ca.id = b.common_area_id
JOIN users u ON u.id = b.user_id
JOIN apartments a ON a.id = b.apartment_id
WHERE b.condominium_id = $1
  AND (sqlc.narg('user_id')::uuid IS NULL OR b.user_id = sqlc.narg('user_id')::uuid)
  AND (sqlc.narg('common_area_id')::uuid IS NULL OR b.common_area_id = sqlc.narg('common_area_id'))
  AND (sqlc.narg('from_date')::timestamptz IS NULL OR b.starts_at >= sqlc.narg('from_date'))
  AND (sqlc.narg('to_date')::timestamptz IS NULL OR b.ends_at <= sqlc.narg('to_date'))
ORDER BY b.starts_at DESC;

-- name: GetAreaAvailability :many
SELECT
  starts_at,
  ends_at,
  status
FROM bookings
WHERE common_area_id = $1
  AND status IN ('confirmed', 'pending')
  AND deleted_at IS NULL
  AND starts_at >= sqlc.arg('starts_at')
  AND ends_at <= sqlc.arg('ends_at')
ORDER BY starts_at ASC;
