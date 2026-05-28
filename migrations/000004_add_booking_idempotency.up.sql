ALTER TABLE bookings ADD COLUMN idempotency_key VARCHAR;
ALTER TABLE bookings ADD COLUMN request_hash VARCHAR;

-- Partial unique index: at most one booking per idempotency key, while still
-- allowing many bookings created without a key (NULL).
CREATE UNIQUE INDEX idx_bookings_idempotency_key
    ON bookings (idempotency_key)
    WHERE idempotency_key IS NOT NULL;
