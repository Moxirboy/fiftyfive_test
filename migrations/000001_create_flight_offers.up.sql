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
