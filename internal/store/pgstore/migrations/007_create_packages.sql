CREATE TABLE IF NOT EXISTS packages (
  id             UUID PRIMARY KEY NOT NULL DEFAULT gen_random_uuid(),
  condominium_id UUID NOT NULL REFERENCES condominiums(id) ON DELETE CASCADE,
  apartment_id   UUID NOT NULL REFERENCES apartments(id) ON DELETE CASCADE,
  received_by    UUID NOT NULL REFERENCES users(id),
  received_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  recipient_name VARCHAR(255),
  photo_url      TEXT,
  status         VARCHAR(25) NOT NULL CHECK(status IN ('pending', 'withdrawn')),
  withdrawn_at   TIMESTAMPTZ,
  withdrawn_by   UUID REFERENCES users(id)
);

CREATE INDEX idx_packages_condo_status ON packages(condominium_id, status);
CREATE INDEX idx_packages_apt_status ON packages(apartment_id, status);
---- create above / drop below ----
DROP TABLE IF EXISTS packages;
