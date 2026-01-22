CREATE TABLE IF NOT EXISTS common_areas (
  id                UUID PRIMARY KEY NOT NULL DEFAULT gen_random_uuid(),
  condominium_id    UUID NOT NULL REFERENCES condominiums(id) ON DELETE CASCADE,
  name              VARCHAR(100) NOT NULL,
  capacity          INT,
  requires_approval BOOLEAN NOT NULL DEFAULT false,

  UNIQUE(condominium_id, name)
);

CREATE TABLE IF NOT EXISTS bookings (
  id             UUID PRIMARY KEY NOT NULL DEFAULT gen_random_uuid(),
  condominium_id UUID NOT NULL REFERENCES condominiums(id) ON DELETE CASCADE,
  apartment_id   UUID NOT NULL REFERENCES apartments(id) ON DELETE CASCADE,
  user_id        UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  common_area_id UUID NOT NULL REFERENCES common_areas(id) ON DELETE CASCADE,
  status         VARCHAR(100) NOT NULL DEFAULT 'confirmed' CHECK(status IN ('pending', 'confirmed', 'cancelled', 'denied')),
  starts_at      TIMESTAMPTZ NOT NULL,
  ends_at        TIMESTAMPTZ NOT NULL,
  created_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at     TIMESTAMPTZ,
  deleted_at     TIMESTAMPTZ,

  CONSTRAINT valid_time_range CHECK (ends_at > starts_at)
);

CREATE INDEX idx_bookings_conflict ON bookings(common_area_id, starts_at, ends_at)
WHERE deleted_at IS NULL AND status IN ('confirmed', 'pending');

---- create above / drop below ----

DROP TABLE IF EXISTS bookings;
DROP TABLE IF EXISTS common_areas;
