-- name: CreateBill :one
INSERT INTO bills (
  condominium_id,
  apartment_id,
  bill_type,
  value_in_cents,
  due_date,
  digitable_line,
  pix_code,
  status
) VALUES (
  $1,
  $2,
  $3,
  $4,
  $5,
  $6,
  $7,
  'pending'
) RETURNING *;

-- name: GetBillById :one
SELECT
  *
FROM bills
WHERE id = $1
  AND condominium_id = $2
  AND status <> 'cancelled'
LIMIT 1;

-- name: ListBillsByApartmentId :many
SELECT
  *
FROM bills
WHERE apartment_id = $1
  AND condominium_id = $2
  AND status <> 'cancelled'
ORDER BY due_date ASC;

-- name: ListBillsByCondominiumId :many
SELECT
  *
FROM bills
WHERE condominium_id = $1
  AND (sqlc.narg('status')::varchar IS NOT NULL OR status = sqlc.narg('status'))
ORDER BY due_date DESC
LIMIT $2 OFFSET $3;

-- name: UpdateBillStatus :one
UPDATE bills
SET
  status = sqlc.arg('status'),
  paid_at = CASE WHEN sqlc.arg('status') = 'paid' THEN NOW() ELSE NULL END,
  updated_at = NOW()
WHERE id = $1 AND condominium_id = $2
RETURNING *;
