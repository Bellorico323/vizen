-- name: GetUserDeviceTokens :many
SELECT
  fcm_token
FROM user_devices
WHERE user_id = $1;

-- name: GetCondoAdminTokens :many
SELECT
  d.fcm_token
FROM user_devices d
JOIN condominium_members m ON m.user_id = d.user_id
WHERE m.condominium_id = $1 AND m.role IN ('admin', 'syndic');

-- name: SaveUserDevice :exec
INSERT INTO user_devices (user_id, fcm_token, platform)
VALUES($1, $2, $3)
ON CONFLICT (fcm_token)
DO UPDATE SET
  user_id = EXCLUDED.user_id,
  last_used_at = NOW();
