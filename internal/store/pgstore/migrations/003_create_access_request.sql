-- Write your migrate up statements here

CREATE TABLE IF NOT EXISTS access_requests (
  id             UUID PRIMARY KEY NOT NULL DEFAULT gen_random_uuid(),
  user_id        UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  condominium_id UUID NOT NULL REFERENCES condominiums(id) ON DELETE CASCADE,
  apartment_id   UUID NOT NULL REFERENCES apartments(id) ON DELETE CASCADE,
  status         VARCHAR(20) NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'approved', 'rejected')),
  reviewed_by    UUID REFERENCES users(id) ON DELETE SET NULL,
  reviewed_at    TIMESTAMPTZ,
  created_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at     TIMESTAMPTZ
);

CREATE UNIQUE INDEX idx_unique_pending_request
ON access_requests(user_id, apartment_id)
WHERE status = 'pending';

CREATE INDEX idx_access_requests_condo_status
ON access_requests(condominium_id, status);

CREATE INDEX idx_access_req_users ON access_requests(user_id);
CREATE INDEX idx_access_req_apartments ON access_requests(apartment_id);

---- create above / drop below ----

-- Write your migrate down statements here. If this migration is irreversible
-- Then delete the separator line above.

DROP TABLE IF EXISTS access_requests;
