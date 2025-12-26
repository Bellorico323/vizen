CREATE TABLE IF NOT EXISTS announcements (
  id             UUID PRIMARY KEY NOT NULL DEFAULT gen_random_uuid(),
  condominium_id UUID NOT NULL REFERENCES condominiums(id) ON DELETE CASCADE,
  author_id      UUID REFERENCES users(id) ON DELETE SET NULL,
  title          VARCHAR(150) NOT NULL,
  content        TEXT NOT NULL,
  created_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at     TIMESTAMPTZ
);

CREATE INDEX idx_announcements_condo_date ON announcements(condominium_id, created_at);

---- create above / drop below ----

DROP TABLE IF EXISTS announcements;
