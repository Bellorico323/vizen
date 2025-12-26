CREATE TABLE IF NOT EXISTS user_devices (
  id            UUID PRIMARY KEY NOT NULL DEFAULT gen_random_uuid(),
  user_id       UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  fcm_token     TEXT NOT NULL,
  platform      VARCHAR(20),
  last_used_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),

  UNIQUE(fcm_token)
);

CREATE INDEX idx_user_devices_user ON user_devices(user_id);

---- create above / drop below ----

DROP TABLE IF EXISTS user_devices;
