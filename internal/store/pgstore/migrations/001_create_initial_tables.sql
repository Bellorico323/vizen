CREATE EXTENSION IF NOT EXISTS "pgcrypto";

CREATE TABLE IF NOT EXISTS condominiums (
    id          UUID PRIMARY KEY NOT NULL DEFAULT gen_random_uuid(),
    name        VARCHAR(100) NOT NULL,
    cnpj        VARCHAR(25) NOT NULL UNIQUE,
    address     VARCHAR(255) NOT NULL,
    plan_type   VARCHAR(20) NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at   TIMESTAMPTZ
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
    condominium_id  UUID NOT NULL REFERENCES condominiums(id) ON DELETE CASCADE,
    role            VARCHAR(25) NOT NULL CHECK (role IN ('admin', 'resident', 'staff')),
    name            VARCHAR(100) NOT NULL,
    avatar_url      TEXT,
    email           VARCHAR(50) NOT NULL UNIQUE,
    password_hash   BYTEA NOT NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ
);

CREATE INDEX idx_apartments_condominium_id ON apartments(condominium_id);
CREATE INDEX idx_users_condominuim_id ON users(condominium_id);
CREATE INDEX idx_users_email ON users(email);

---- create above / drop below ----

DROP TABLE IF EXISTS users;
DROP TABLE IF EXISTS apartments;
DROP TABLE IF EXISTS condominiums;
