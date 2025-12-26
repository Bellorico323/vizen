-- name: CreateAnnouncement :one
INSERT INTO announcements (
  condominium_id,
  author_id,
  title,
  content
) VALUES (
  $1,
  $2,
  $3,
  $4
) RETURNING id, created_at;

-- name: GetManyAnnouncementsByCondoId :many
SELECT
  a.id,
  a.title,
  a.content,
  a.created_at,
  u.name AS author_name
FROM announcements a
LEFT JOIN users u ON u.id = a.author_id
WHERE a.condominium_id = $1
ORDER BY a.created_at DESC
LIMIT $2 OFFSET $3;

-- name: DeleteAnnouncement :exec
DELETE FROM announcements
WHERE id = $1 AND condominium_id = $2;

-- name: GetAnnouncementById :one
SELECT
  *
FROM announcements
WHERE id = $1;

-- name: UpdateAnnouncement :exec
UPDATE announcements
SET title = $2,
    content = $3,
    updated_at = NOW()
WHERE id = $1;
