CREATE TABLE flight_offers (
    id BIGSERIAL PRIMARY KEY,
    offer_id VARCHAR UNIQUE,
    provider VARCHAR,
    origin VARCHAR(3),
    destination VARCHAR(3),
    departure_date DATE,
    return_date DATE NULL,
    airline VARCHAR,
    flight_number VARCHAR,
    base_price BIGINT,
    commission BIGINT,
    service_fee BIGINT,
    total_price BIGINT,
    profit BIGINT,
    currency VARCHAR(3),
    expires_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ DEFAULT now()
);

CREATE TABLE bookings (
    id BIGSERIAL PRIMARY KEY,
    booking_id VARCHAR UNIQUE,
    offer_id VARCHAR REFERENCES flight_offers(offer_id),
    status VARCHAR,
    expires_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ DEFAULT now()
);

CREATE INDEX idx_bookings_offer_id ON bookings (offer_id);

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
