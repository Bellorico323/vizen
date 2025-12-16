-- name: CreateAccountWithCredentials :exec
INSERT INTO accounts (
  user_id,
  provider_account_id,
  provider_id,
  password_hash
) VALUES (
  $1,
  $2,
  $3,
  $4
);

-- name: GetAccountByUserId :one
SELECT
  *
FROM accounts
WHERE user_id = $1;
