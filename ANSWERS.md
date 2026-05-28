# Post-task Answers

## 1. Adding A Second Provider

Create a new package under `internal/providers` that implements `FlightProvider`. The adapter owns supplier-specific request/response mapping and returns normalized `ProviderOffer` values. Wire it through a provider registry or slice, let the flight service fan out searches in parallel, then merge the results before pricing and persistence. Handlers, pricing, repositories, and booking logic do not need provider-specific changes. Each provider is already wrapped by the retry/timeout decorator (`providers.NewRetry`), so a new adapter inherits per-attempt timeouts and retries with no extra code.

## 2. Caching Search Results

Use a normalized key such as `from|to|departure_date|return_date|adults|children|infants|currency`. Store results in Redis for a short TTL, for example 1-5 minutes, because fares are volatile. Cache normalized provider offers before final pricing so configured fees still apply at response time. Add singleflight or a lock per key to prevent cache stampedes.

## 3. Preventing Duplicate Bookings

This is implemented. `POST /api/v1/bookings` accepts an `Idempotency-Key` header; the first request stores the key together with a SHA-256 hash of the request body, and a partial unique index (`idx_bookings_idempotency_key`) guarantees one booking per key even under concurrent retries (a lost race re-reads and replays the winner). Repeating the same key with the same body returns the original booking; reusing a key with a different body returns `409 IDEMPOTENCY_CONFLICT`. To further guarantee one active booking per offer, add a unique constraint or transactional check on `offer_id`.

## 4. Logs And Monitoring

Keep structured `slog` JSON logs on stdout with `request_id` on each request and downstream service log. Add Prometheus RED metrics: request rate, error count, and duration, plus provider latency and booking counters. Add OpenTelemetry traces across handler, service, provider, and database boundaries. Alert on elevated 5xx rate, provider failures, high latency, and failed booking creation.

## 5. Future Microservice Splits

Split only when operational pressure justifies it. Natural boundaries are a Search/Provider Integration service for provider adapters and normalization, a Pricing service for money rules, a Booking service for offer and booking lifecycle, and an API gateway/BFF for external clients. Shared contracts should remain small and versioned so provider churn does not leak into booking or pricing.
