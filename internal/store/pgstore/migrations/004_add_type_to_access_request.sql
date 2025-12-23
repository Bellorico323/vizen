ALTER TABLE access_requests
ADD COLUMN type VARCHAR(25) NOT NULL DEFAULT 'tenant'
CHECK(type IN ('owner', 'tenant', 'dependent'));

---- create above / drop below ----

ALTER TABLE access_requests DROP COLUMN type;
