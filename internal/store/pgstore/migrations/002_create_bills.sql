CREATE TABLE IF NOT EXISTS bills (
  id              UUID PRIMARY KEY NOT NULL DEFAULT gen_random_uuid(),
  condominium_id  UUID NOT NULL REFERENCES condominiums(id) ON DELETE CASCADE,
  apartment_id    UUID NOT NULL REFERENCES apartments(id) ON DELETE CASCADE,
  bill_type       VARCHAR(50) NOT NULL CHECK(bill_type IN ('rent', 'condominium_fee', 'water', 'electricity', 'gas', 'fine')),
  value_in_cents  BIGINT NOT NULL CHECK(value_in_cents >= 0),
  due_date        DATE,
  paid_at         TIMESTAMPTZ,
  created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at      TIMESTAMPTZ
);

CREATE INDEX idx_bills_condominium_id ON bills(condominium_id);
CREATE INDEX idx_bills_apartment_id ON bills(apartment_id);
CREATE INDEX idx_bills_apartment_unpaid
ON bills(apartment_id)
WHERE paid_at IS NULL;

---- create above / drop below ----

DROP TABLE IF EXISTS bills;
