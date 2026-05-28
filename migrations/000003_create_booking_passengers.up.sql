CREATE TABLE booking_passengers (
    id BIGSERIAL PRIMARY KEY,
    booking_id BIGINT REFERENCES bookings(id) ON DELETE CASCADE,
    type VARCHAR,
    first_name VARCHAR,
    last_name VARCHAR,
    document_number VARCHAR,
    created_at TIMESTAMPTZ DEFAULT now()
);

CREATE INDEX idx_booking_passengers_booking_id ON booking_passengers (booking_id);
