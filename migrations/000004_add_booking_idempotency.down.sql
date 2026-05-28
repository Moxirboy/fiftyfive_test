DROP INDEX IF EXISTS idx_bookings_idempotency_key;
ALTER TABLE bookings DROP COLUMN IF EXISTS request_hash;
ALTER TABLE bookings DROP COLUMN IF EXISTS idempotency_key;
