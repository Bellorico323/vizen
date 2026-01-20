CREATE TABLE IF NOT EXISTS invites (
  id             UUID PRIMARY KEY NOT NULL DEFAULT gen_random_uuid(),
  condominium_id UUID NOT NULL REFERENCES condominiums(id) ON DELETE CASCADE,
  apartment_id   UUID NOT NULL REFERENCES apartments(id) ON DELETE CASCADE,
  issued_by      UUID NOT NULL REFERENCES users(id),
  guest_name     VARCHAR(255) NOT NULL,
  guest_type     VARCHAR(50) NOT NULL DEFAULT 'guest' CHECK (guest_type IN ('guest', 'service', 'delivery')),
  token          UUID NOT NULL DEFAULT gen_random_uuid(),
  starts_at      TIMESTAMPTZ NOT NULL,
  ends_at        TIMESTAMPTZ NOT NULL,
  revoked_at     TIMESTAMPTZ,
  created_at     TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_invites_token ON invites(token);
CREATE INDEX idx_invites_validity ON invites(condominium_id, starts_at, ends_at);

CREATE TABLE IF NOT EXISTS access_logs (
  id             UUID PRIMARY KEY NOT NULL DEFAULT gen_random_uuid(),
  invite_id      UUID NOT NULL REFERENCES invites(id),
  condominium_id UUID NOT NULL REFERENCES condominiums(id),
  entered_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  authorized_by  UUID REFERENCES users(id)
);

---- create above / drop below ----

DROP TABLE IF EXISTS access_logs;
DROP TABLE IF EXISTS invites;
