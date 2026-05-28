CREATE TABLE bookings (
    id BIGSERIAL PRIMARY KEY,
    booking_id VARCHAR UNIQUE,
    offer_id VARCHAR REFERENCES flight_offers(offer_id),
    status VARCHAR,
    expires_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ DEFAULT now()
);

CREATE INDEX idx_bookings_offer_id ON bookings (offer_id);
