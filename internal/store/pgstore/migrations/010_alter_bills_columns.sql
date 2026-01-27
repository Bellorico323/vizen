ALTER TABLE bills
ALTER COLUMN due_date SET NOT NULL,
ADD COLUMN digitable_line VARCHAR(60),
ADD COLUMN pix_code       TEXT,
ADD COLUMN status         VARCHAR(50) NOT NULL DEFAULT 'pending'
CHECK(status IN ('pending', 'paid', 'overdue', 'cancelled'));

---- create above / drop below ----

ALTER TABLE bills
DROP COLUMN digitable_line,
DROP COLUMN pix_code,
DROP COLUMN status,
ALTER COLUMN due_date DROP NOT NULL;
