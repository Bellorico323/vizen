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
