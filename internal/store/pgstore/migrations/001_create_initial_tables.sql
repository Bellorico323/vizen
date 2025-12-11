CREATE EXTENSION IF NOT EXISTS "pgcrypto";

CREATE TABLE IF NOT EXISTS condominiums (
    id          UUID PRIMARY KEY NOT NULL DEFAULT gen_random_uuid(),
    name        VARCHAR(100) NOT NULL,
    cnpj        VARCHAR(25) NOT NULL UNIQUE,
    address     VARCHAR(255) NOT NULL,
    plan_type   VARCHAR(20) NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ
);

CREATE TABLE IF NOT EXISTS apartments (
    id              UUID PRIMARY KEY NOT NULL DEFAULT gen_random_uuid(),
    condominium_id  UUID NOT NULL REFERENCES condominiums(id) ON DELETE CASCADE,
    block           VARCHAR(20),
    number          VARCHAR(10) NOT NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ
);

CREATE TABLE IF NOT EXISTS users (
    id              UUID PRIMARY KEY NOT NULL DEFAULT gen_random_uuid(),
    role            VARCHAR(25) NOT NULL CHECK (role IN ('admin', 'resident', 'staff')),
    name            VARCHAR(100) NOT NULL,
    avatar_url      TEXT,
    email           VARCHAR(100) NOT NULL UNIQUE,
    email_verified  TIMESTAMPTZ,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ
);

CREATE TABLE IF NOT EXISTS residents (
  id              UUID PRIMARY KEY NOT NULL DEFAULT gen_random_uuid(),
  user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  apartment_id UUID NOT NULL REFERENCES apartments(id) ON DELETE CASCADE,
  type VARCHAR(25) NOT NULL CHECK(type IN ('owner', 'tenant', 'dependent')),
  is_responsible BOOLEAN NOT NULL DEFAULT false,
  created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),

  UNIQUE(user_id, apartment_id)
);

CREATE TABLE IF NOT EXISTS accounts (
  id                        UUID PRIMARY KEY NOT NULL DEFAULT gen_random_uuid(),
  user_id                   UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  provider_account_id       TEXT NOT NULL,
  provider_id               TEXT NOT NULL,
  password_hash             BYTEA,
  access_token              TEXT,
  refresh_token             TEXT,
  access_token_expires_at   BIGINT,
  refresh_token_expires_at  BIGINT,
  scope                     TEXT,
  id_token                  TEXT,
  created_at                TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at                TIMESTAMPTZ,

  UNIQUE(provider_id, provider_account_id)
);

CREATE TABLE IF NOT EXISTS sessions (
  id         UUID PRIMARY KEY NOT NULL DEFAULT gen_random_uuid(),
  user_id    UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  token      VARCHAR(255) UNIQUE NOT NULL,
  expires_at TIMESTAMPTZ NOT NULL,
  ip_address INET,
  user_agent TEXT,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ
);

CREATE TABLE IF NOT EXISTS verifications (
  id         UUID PRIMARY KEY NOT NULL DEFAULT gen_random_uuid(),
  identifier TEXT NOT NULL,
  value      TEXT NOT NULL,
  expires_at TIMESTAMPTZ NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ
);

CREATE INDEX idx_apartments_condominium_id ON apartments(condominium_id);
CREATE INDEX idx_users_email ON users(email);

---- create above / drop below ----
DROP TABLE IF EXISTS verifications;
DROP TABLE IF EXISTS sessions;
DROP TABLE IF EXISTS accounts;
DROP TABLE IF EXISTS residents;
DROP TABLE IF EXISTS users;
DROP TABLE IF EXISTS apartments;
DROP TABLE IF EXISTS condominiums;
