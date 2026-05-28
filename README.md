# FlySoft Flight Integration Service

FlySoft is a Go backend microservice that simulates integration with an external flight-ticket supplier. The core business flow is:

1. Search: accept route, dates, passenger counts, and currency.
2. Price: call the provider, normalize supplier data, and calculate base price, commission, service fee, total, and profit.
3. Persist: save each priced offer with an expiry time.
4. Book: create a preliminary booking for an existing, unexpired offer.

## Architecture

The service keeps HTTP, business logic, persistence, and provider integration separate:

```text
handlers -> services -> {repository, providers}
```

Handlers only bind JSON, call services, and map results or errors. Business rules live in `internal/services`. Repositories hide GORM/PostgreSQL access. Providers hide supplier-specific payloads behind the shared `FlightProvider` interface:

```go
type FlightProvider interface {
	Search(ctx context.Context, req ProviderSearch) ([]ProviderOffer, error)
	Name() string
}
```

The current app wires one provider, `MockAvia`. A second provider plugs in by adding a package that implements `FlightProvider`, then registering it in a provider slice or aggregator used by `FlightService`. That lets the service fan out searches, merge normalized `ProviderOffer` results, and reuse the same pricing, persistence, and HTTP layers.

Money is represented as `int64` cents everywhere, including database columns. Floating point is used only by `internal/money` when rendering JSON display values like `550.00`.

All API errors use one envelope:

```json
{"success":false,"error":{"code":"VALIDATION_ERROR","message":"from is required"}}
```

Success responses use `{"success":true,"data":...}` for search and `{"success":true,"booking":...}` for bookings.

### Project Tree

```text
.
|-- cmd/app/main.go
|-- internal
|   |-- apperror
|   |-- config
|   |-- database
|   |   `-- models
|   |-- domain
|   |-- dto
|   |-- http
|   |   |-- handlers
|   |   `-- middleware
|   |-- logger
|   |-- money
|   |-- pricing
|   |-- providers
|   |   `-- mockavia
|   |-- repository
|   `-- services
|-- migrations
|-- docs
|   |-- PLAN.md
|   |-- docs.go
|   |-- swagger.json
|   `-- swagger.yaml
|-- Dockerfile
|-- docker-compose.yml
|-- Makefile
|-- .env.example
|-- ANSWERS.md
`-- README.md
```

## Tech Stack And Decisions

- Go 1.25 backend service.
- Gin for HTTP routing and middleware.
- GORM for database access.
- PostgreSQL 16 for persistence.
- golang-migrate with raw SQL migrations for schema changes.
- `log/slog` JSON logging with request ID correlation.
- swaggo for generated OpenAPI docs served by the app.
- Manual dependency wiring in `cmd/app/main.go`.
- Database columns use `origin` and `destination` instead of SQL-reserved `from` and `to`; JSON fields remain `from` and `to`.

## Run Locally

Prerequisites: Go 1.25, PostgreSQL 16, and the `migrate` CLI.

```sh
cp .env.example .env
set -a
. ./.env
set +a

createdb flysoft_flight_service

export DATABASE_URL="postgres://postgres:postgres@localhost:5432/flysoft_flight_service?sslmode=disable"
make migrate-up

go run ./cmd/app
```

The service listens on `http://localhost:8080` by default.

## Run With Docker Compose

Docker Compose starts PostgreSQL, runs migrations once, builds the app, and exposes the API:

```sh
docker compose up --build
```

Optional: create `.env` from `.env.example` first if you want to override defaults.

```sh
cp .env.example .env
docker compose up --build
```

Swagger UI is available at:

```text
http://localhost:8080/swagger/index.html
```

## Curl Examples

Health:

```sh
curl -s http://localhost:8080/health
```

Expected response:

```json
{"status":"ok"}
```

Valid search:

```sh
curl -s -X POST http://localhost:8080/api/v1/flights/search \
  -H 'Content-Type: application/json' \
  -d '{
    "from": "DYU",
    "to": "IST",
    "departure_date": "2099-06-15",
    "return_date": "2099-06-25",
    "adults": 1,
    "children": 1,
    "infants": 0,
    "currency": "USD"
  }'
```

Example response:

```json
{
  "success": true,
  "data": [
    {
      "offer_id": "OF-123456",
      "provider": "MockAvia",
      "from": "DYU",
      "to": "IST",
      "departure_date": "2099-06-15",
      "return_date": "2099-06-25",
      "airline": "TK",
      "flight_number": "TK255",
      "base_price": 500.00,
      "service_fee": 25.00,
      "commission": 25.00,
      "total_price": 550.00,
      "profit": 50.00,
      "currency": "USD"
    }
  ]
}
```

Validation error:

```sh
curl -s -X POST http://localhost:8080/api/v1/flights/search \
  -H 'Content-Type: application/json' \
  -d '{
    "from": "DYU",
    "to": "DYU",
    "departure_date": "2099-06-15",
    "adults": 1,
    "children": 0,
    "infants": 0,
    "currency": "USD"
  }'
```

Expected response:

```json
{"success":false,"error":{"code":"VALIDATION_ERROR","message":"from and to must be different"}}
```

Create a booking from a returned offer:

```sh
OFFER_ID=$(curl -s -X POST http://localhost:8080/api/v1/flights/search \
  -H 'Content-Type: application/json' \
  -d '{"from":"DYU","to":"IST","departure_date":"2099-06-15","adults":1,"children":0,"infants":0,"currency":"USD"}' \
  | jq -r '.data[0].offer_id')

curl -s -X POST http://localhost:8080/api/v1/bookings \
  -H 'Content-Type: application/json' \
  -d "{
    \"offer_id\": \"${OFFER_ID}\",
    \"passengers\": [
      {
        \"type\": \"adult\",
        \"first_name\": \"Alisher\",
        \"last_name\": \"Sabirov\",
        \"document_number\": \"A1234567\"
      }
    ]
  }"
```

Example response:

```json
{
  "success": true,
  "booking": {
    "booking_id": "BK-987654",
    "status": "created",
    "offer_id": "OF-123456",
    "expires_at": "2026-05-29T12:30:00Z"
  }
}
```

## Environment Variables

These variables match `.env.example`.

| Variable | Default example | Description |
|---|---:|---|
| `HTTP_PORT` | `8080` | HTTP port exposed by the app. |
| `DB_HOST` | `localhost` | PostgreSQL host for local runs. Docker Compose overrides this to `db`. |
| `DB_PORT` | `5432` | PostgreSQL port. |
| `DB_USER` | `postgres` | PostgreSQL username. |
| `DB_PASSWORD` | `postgres` | PostgreSQL password. |
| `DB_NAME` | `flysoft_flight_service` | PostgreSQL database name. |
| `DB_SSLMODE` | `disable` | PostgreSQL SSL mode. |
| `LOG_LEVEL` | `info` | `debug`, `info`, `warn`, or `error`. |
| `OFFER_TTL` | `30m` | Duration before persisted offers expire. |
| `COMMISSION_PERCENT` | `5` | Commission percentage applied to base price. |
| `SERVICE_FEE_ADULT` | `1500` | Adult service fee in cents. |
| `SERVICE_FEE_CHILD` | `1000` | Child service fee in cents. |
| `SERVICE_FEE_INFANT` | `0` | Infant service fee in cents. |

`DATABASE_URL` is used only by the Makefile migration target and is not an application setting.

## What Could Be Improved With More Time

- Provider timeout plus retry/backoff policy.
- Idempotency key support for booking creation.
- Short-lived search-result caching.
- Prometheus metrics and OpenTelemetry tracing.
- Authentication and rate limiting.
- Offer-expiry sweeper job.
- CI pipeline for lint, test, build, and Swagger freshness checks.
- Integration tests with testcontainers.
