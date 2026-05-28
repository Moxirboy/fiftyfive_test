# FlySoft Flight Integration Service — Implementation Plan

> Source spec: `Тестовое_задание_для_Backend_Developer_GoLang.pdf` (Backend Developer GoLang test task).
> This plan is the single source of truth for the build. The actual implementation is delegated to
> the **Codex CLI** — see [`CODEX_PROMPTS.md`](./CODEX_PROMPTS.md) for the ready-to-run, per-phase prompts.

---

## 1. Goal

Build a small Go backend microservice — **Flight Integration Service** — that simulates integration with an
external flight-ticket supplier. It accepts a search request, calls a **mock provider**, normalizes the
provider's data into a unified internal format, computes the final price (base + commission + service fee +
profit), persists offers, returns them to the client, and lets the client create a **preliminary booking**
against a chosen offer.

The task grades: sequential thinking, backend process logic, **mathematical precision in money calculations**,
clean code, microservice architecture understanding, working with external data providers, and designing an
**extensible** service (new providers added without rewriting core logic).

> The task explicitly says: *quality of architecture, logic, calculations, and integration approach matters more
> than volume of code.* This plan optimizes for that.

---

## 2. Locked decisions (with rationale)

| Decision | Choice | Why |
|---|---|---|
| Language | **Go 1.25** | Required. Installed: `go1.25.6`. |
| HTTP router | **Gin** (`github.com/gin-gonic/gin`) | Chosen. Popular, fast, good middleware ergonomics. |
| DB | **PostgreSQL 16** | Required (Postgres or MySQL). |
| Data access | **GORM** (`gorm.io/gorm` + `gorm.io/driver/postgres`) | Chosen. Productive ORM. Used for **queries only**. |
| Schema migrations | **golang-migrate** (raw SQL `.up/.down`) | The task lists *migrations* as a plus; explicit SQL is cleaner and more production-like than `AutoMigrate`. |
| Money | **`int64` minor units (cents)** via a `Money` type | Spec **forbids** raw `float64` for money. All math is integer cents; float only at the JSON display boundary. |
| Logging | **`log/slog`** (stdlib JSON handler) | Structured, zero extra deps, carries `request_id`. |
| Config | env vars + `.env` (`.env.example` committed) | 12-factor, Docker-friendly. |
| DI | **manual constructor wiring** in `main.go` | "DI without unnecessary magic" — no `wire`/`fx`. |
| API docs | **swaggo/swag** + `gin-swagger` | Generates OpenAPI from handler annotations; served at `/swagger`. |
| Tests | Go stdlib `testing` (table-driven) | Unit tests for pricing + validation (the highest-signal areas). |

---

## 3. Scope — **Core + high-value bonuses**

**IN (mandatory):** Go · REST API · PostgreSQL · Docker + Docker Compose · README with run instructions · clean project structure.

**IN (high-value bonuses):**
- Gin router (idiomatic structure, business logic out of handlers)
- GORM data layer + golang-migrate migrations
- Swagger / OpenAPI documentation
- Unit tests (pricing calculation + request validation)
- Structured logging (`slog`, JSON)
- Graceful shutdown (SIGINT/SIGTERM, in-flight drain)
- `request_id` middleware + in every log line
- `GET /health` health-check endpoint
- Docker Compose for service **and** database (+ one-shot migration step)
- Unified error format across all endpoints

**OUT (deferred to "future improvements" + answered in §15 Q&A):**
- Provider call timeout config, retry/backoff mechanic
- Idempotency key for booking creation
- Caching of search results
- Auth / rate limiting / metrics

> Note: the provider interface takes `context.Context`, so context cancellation is respected, but no
> configurable timeout/retry layer is built (that was the "Everything" tier). This boundary is intentional
> and called out in the README's "what could be improved" section.

---

## 4. Project structure

Repo root: `flysoft-flight-service/` (a new dir inside `task_provider/`; the PDF stays in the parent as reference).

```
flysoft-flight-service/
├── cmd/
│   └── app/
│       └── main.go                  # entrypoint: config → logger → db → repos → provider → services → handlers → server; graceful shutdown
├── internal/
│   ├── config/
│   │   └── config.go                # env loading + validation
│   ├── logger/
│   │   └── logger.go                # slog setup; ctx<->logger helpers; request_id field
│   ├── apperror/
│   │   └── apperror.go              # AppError{Code,HTTPStatus,Message}; constructors; error codes
│   ├── money/
│   │   └── money.go                 # Money int64 (cents); FromCents; MarshalJSON "%.2f"; arithmetic
│   ├── pricing/
│   │   ├── pricing.go               # Calculator: commission, service_fee, total, profit
│   │   └── pricing_test.go          # table-driven, includes the spec's worked example
│   ├── domain/
│   │   ├── flight/flight.go         # Offer entity, SearchCriteria value object
│   │   └── booking/booking.go       # Booking, Passenger entities, PassengerType
│   ├── dto/
│   │   ├── flight_dto.go            # SearchRequest, OfferResponse, search validation
│   │   ├── booking_dto.go           # BookingRequest, PassengerDTO, BookingResponse, booking validation
│   │   ├── response.go              # success envelope {success,data}/{success,booking}
│   │   └── validate_test.go         # table-driven validation tests
│   ├── providers/
│   │   ├── provider.go              # FlightProvider interface + ProviderOffer + ProviderSearch types
│   │   └── mockavia/
│   │       ├── mockavia.go          # implements FlightProvider; canned routes in supplier format → normalized
│   │       ├── mockavia_test.go
│   │       └── fixtures.go          # supplier-format sample payload
│   ├── repository/
│   │   ├── repository.go            # FlightOfferRepository, BookingRepository interfaces
│   │   ├── flight_repository.go     # GORM impl
│   │   └── booking_repository.go    # GORM impl (offer+passengers in a tx)
│   ├── database/
│   │   ├── database.go              # GORM connect + pool config
│   │   └── models/                  # GORM table models
│   │       ├── flight_offer.go
│   │       ├── booking.go
│   │       └── booking_passenger.go
│   ├── services/
│   │   ├── flight_service.go        # search usecase: validate→provider→normalize→price→persist→DTO
│   │   ├── booking_service.go       # booking usecase: validate→load offer→check expiry→persist→DTO
│   │   └── service_test.go          # with mocked repo + provider
│   └── http/
│       ├── router.go                # gin engine + route registration + swagger route
│       ├── middleware/
│       │   ├── request_id.go        # X-Request-ID (generate if absent) → ctx
│       │   ├── logger.go            # per-request structured access log
│       │   └── recovery.go          # panic → 500 unified error
│       └── handlers/
│           ├── flight_handler.go    # POST /api/v1/flights/search
│           ├── booking_handler.go   # POST /api/v1/bookings
│           └── health_handler.go    # GET /health
├── migrations/
│   ├── 000001_init.up.sql
│   └── 000001_init.down.sql
├── docs/                            # swag-generated (swagger.json/yaml/docs.go); + copy of this PLAN + spec PDF
├── Dockerfile                       # multi-stage build
├── docker-compose.yml               # app + postgres + one-shot migrate
├── Makefile                         # run, test, migrate, swag, lint, docker targets
├── .env.example
├── .gitignore
├── AGENTS.md                        # Codex grounding doc (conventions + spec digest)
├── go.mod / go.sum
└── README.md
```

**Layering rule (must hold):** `handlers → services → {repository, providers}`. Handlers do HTTP only
(bind, call service, map result/error). All business logic lives in `services`. `domain`, `money`,
`pricing`, `apperror` are dependency-free leaves.

---

## 5. Data model & migrations

All money columns are **`BIGINT` (cents)**. Dates are `DATE`; timestamps are `TIMESTAMPTZ`.

> **Reserved-word note:** `from` and `to` are SQL reserved words. DB columns are named **`origin`** and
> **`destination`**; the JSON/DTO fields stay `from`/`to`. This avoids quoting reserved identifiers everywhere
> and is a deliberate, documented deviation from the spec's literal column names.

### `flight_offers`
`id` BIGSERIAL PK · `offer_id` VARCHAR UNIQUE · `provider` VARCHAR · `origin` VARCHAR(3) · `destination` VARCHAR(3) ·
`departure_date` DATE · `return_date` DATE NULL · `airline` VARCHAR · `flight_number` VARCHAR ·
`base_price` BIGINT · `commission` BIGINT · `service_fee` BIGINT · `total_price` BIGINT · `profit` BIGINT ·
`currency` VARCHAR(3) · `expires_at` TIMESTAMPTZ · `created_at` TIMESTAMPTZ DEFAULT now().
Index: unique(`offer_id`).

### `bookings`
`id` BIGSERIAL PK · `booking_id` VARCHAR UNIQUE · `offer_id` VARCHAR (FK→`flight_offers.offer_id`) ·
`status` VARCHAR · `expires_at` TIMESTAMPTZ · `created_at` TIMESTAMPTZ DEFAULT now().
Index: unique(`booking_id`), index(`offer_id`).

### `booking_passengers`
`id` BIGSERIAL PK · `booking_id` BIGINT (FK→`bookings.id`, ON DELETE CASCADE) · `type` VARCHAR ·
`first_name` VARCHAR · `last_name` VARCHAR · `document_number` VARCHAR · `created_at` TIMESTAMPTZ DEFAULT now().
Index: index(`booking_id`).

> `booking_passengers.booking_id` is the **internal** `bookings.id` FK (for referential integrity). The public
> `BK-…` business key lives on `bookings.booking_id`.

---

## 6. Money & pricing logic (the highest-signal part)

Everything is `int64` cents. The provider already sends cents (`price_adult: 30000` = $300.00).

```
base_cents       = adults*price_adult_cents + children*price_child_cents + infants*price_infant_cents
commission_cents = round_half_up(base_cents * COMMISSION_PERCENT / 100)   // COMMISSION_PERCENT = 5
service_fee_cents= adults*1500 + children*1000 + infants*0                // adult=$15, child=$10, infant=$0
total_cents      = base_cents + commission_cents + service_fee_cents
profit_cents     = commission_cents + service_fee_cents
```

**Rounding rule (predictable, documented):** round half up to the nearest cent.
For positive integers: `commission = (base*percent + 50) / 100` (integer division).

**Reference implementation (give this to Codex verbatim):**
```go
// commission returns commission in cents, rounded half-up to the nearest cent.
func commission(baseCents int64, percent int64) int64 {
    return (baseCents*percent + 50) / 100
}
```

**JSON output:** `Money` marshals as a 2-decimal JSON number via `fmt.Sprintf("%.2f", float64(c)/100)`.
This is the only place a float appears, and only for display — never in arithmetic. Output looks like `466.00`.

**Worked example from the spec (MUST be a passing unit test):**
- 1 adult base $300.00 (30000c) + 1 child base $200.00 (20000c) → base = 50000c
- commission 5% = 2500c ($25.00)
- service_fee = 1500 + 1000 = 2500c ($25.00)
- total = 50000 + 2500 + 2500 = 55000c ($550.00)
- profit = 2500 + 2500 = 5000c ($50.00)

`COMMISSION_PERCENT` and the per-type service fees are config-injected (defaults 5 / 15 / 10 / 0) so a reviewer
sees they're not hard-coded magic numbers.

---

## 7. API contracts

All responses use a **unified envelope**. Success: `{"success": true, "data": ...}` (or `"booking"` for bookings).
Error (any failure): `{"success": false, "error": {"code": "...", "message": "..."}}`.

### 7.1 `POST /api/v1/flights/search`
Request:
```json
{ "from":"DYU","to":"IST","departure_date":"2026-06-15","return_date":"2026-06-25",
  "adults":1,"children":1,"infants":0,"currency":"USD" }
```
**Validation** (→ `VALIDATION_ERROR`, 400):
- `from` required; `to` required; `from != to`
- `departure_date` parseable `YYYY-MM-DD` and **not in the past** (date-only compare vs today)
- total passengers (`adults+children+infants`) > 0
- `infants <= adults`
- `currency` required

Flow: validate → `provider.Search(ctx, criteria)` → normalize → price each offer (per requested pax mix) →
persist each offer (`OF-…`, `expires_at = now + OFFER_TTL`, default 30m) → return `data: [OfferResponse...]`.

`OfferResponse`: `offer_id, provider, from, to, departure_date, return_date, airline, flight_number,
base_price, service_fee, commission, total_price, profit, currency`.

### 7.2 `POST /api/v1/bookings`
Request:
```json
{ "offer_id":"OF-123456",
  "passengers":[ {"type":"adult","first_name":"Alisher","last_name":"Sabirov","document_number":"A1234567"} ] }
```
**Validation:**
- `offer_id` required (else `VALIDATION_ERROR` 400)
- offer must exist (else `OFFER_NOT_FOUND` 404)
- offer must not be expired (else `OFFER_EXPIRED` 409)
- passengers non-empty; each passenger has `type` ∈ {adult,child,infant}, `first_name`, `last_name`, `document_number` (else `VALIDATION_ERROR` 400)

Flow: validate → load offer → check expiry → create booking (`BK-…`, status `created`,
`expires_at`) + passengers in **one transaction** → return `booking`.

Response:
```json
{ "success": true, "booking": {"booking_id":"BK-987654","status":"created","offer_id":"OF-123456","expires_at":"2026-05-28T15:30:00Z"} }
```

### 7.3 `GET /health` → `{"status":"ok"}` (200)

### 7.4 Error codes → HTTP status
| code | status | when |
|---|---|---|
| `BAD_REQUEST` | 400 | malformed JSON / unbindable body |
| `VALIDATION_ERROR` | 400 | field validation failed |
| `OFFER_NOT_FOUND` | 404 | booking references unknown offer |
| `OFFER_EXPIRED` | 409 | booking references expired offer |
| `PROVIDER_UNAVAILABLE` | 502 | provider call failed |
| `PROVIDER_EMPTY` | 404 | provider returned zero offers |
| `INTERNAL_ERROR` | 500 | unexpected / DB error |

A central error mapper (middleware or handler helper) turns any `*AppError` into the unified body; non-`AppError`
errors become `INTERNAL_ERROR` 500 and are logged with the `request_id`.

---

## 8. Provider layer & normalization

Interface (extensibility cornerstone — §9 of the spec):
```go
type FlightProvider interface {
    Search(ctx context.Context, req ProviderSearch) ([]ProviderOffer, error)
}
```

**MockAvia** implements it, returning canned routes in the *supplier's own format*, then normalizing.
Supplier format (per spec §7):
```json
{ "supplier":"MockAvia","routes":[
  {"id":"SUP-001","dep":"DYU","arr":"IST","carrier":"TK","flt":"255",
   "price_adult":30000,"price_child":20000,"price_infant":0,"currency":"USD"} ]}
```

**Normalization mapping** (supplier → internal `ProviderOffer`):
| supplier field | internal | note |
|---|---|---|
| `dep` | `From` | |
| `arr` | `To` | |
| `carrier` | `Airline` | e.g. `TK` |
| `carrier`+`flt` | `FlightNumber` | `"TK"+"255"` = `"TK255"` (matches response example) |
| `price_adult/child/infant` | per-type base cents | already cents |
| `currency` | `Currency` | |

Adding a second provider = a new package implementing `FlightProvider`; **no service rewrite**.

---

## 9. Cross-cutting concerns

- **Config** (`internal/config`): `HTTP_PORT`, `DB_HOST/PORT/USER/PASSWORD/NAME/SSLMODE`, `LOG_LEVEL`,
  `OFFER_TTL` (default `30m`), `COMMISSION_PERCENT` (default `5`), `SERVICE_FEE_ADULT/CHILD/INFANT` (1500/1000/0 cents).
- **request_id**: middleware reads `X-Request-ID` or generates a UUID; stored in `context`, echoed in response
  header, attached to every `slog` line.
- **Structured logging**: `slog` JSON; access-log middleware logs method, path, status, latency_ms, request_id.
- **Graceful shutdown**: `signal.NotifyContext(SIGINT, SIGTERM)`; `http.Server.Shutdown(ctx)` with a drain timeout.
- **Recovery**: panic middleware → `INTERNAL_ERROR` 500 (never leak stack to client; log it).
- **DI**: all wiring is explicit in `cmd/app/main.go`; services depend on **interfaces** (`FlightProvider`,
  the repository interfaces), enabling the unit tests in §10.

---

## 10. Phase breakdown (maps 1:1 to `CODEX_PROMPTS.md`)

Each phase is independently runnable and has a hard **acceptance gate**. Run them in order.

| # | Phase | Key deliverables | Acceptance gate |
|---|---|---|---|
| 0 | Scaffold & tooling | dirs, `go.mod`, deps, config, slog, Makefile, `.env.example`, `.gitignore`, `AGENTS.md`, stub `main` + `/health`, `git init` | `go build ./...` ok; `go run` → `GET /health` = `{"status":"ok"}` |
| 1 | Money, pricing, errors, domain, DTOs | `money`, `pricing`(+tests incl. spec example), `apperror`, `domain`, `dto`(+validation tests) | `go test ./internal/money/... ./internal/pricing/... ./internal/dto/...` green |
| 2 | Provider layer | `FlightProvider` iface, `ProviderOffer`, `mockavia` + normalization + test | `go test ./internal/providers/...` green |
| 3 | Persistence | `000001` migration (3 tables), GORM models, repo interfaces + impls | `migrate up` applies on a live Postgres; `go build ./...` ok |
| 4 | Services / usecases | `flight_service`, `booking_service`, validation wiring, service tests (mock repo+provider) | `go test ./internal/services/...` green |
| 5 | HTTP layer + wiring | handlers (3), middleware (request_id/logger/recovery), router, full `main.go` DI + graceful shutdown | live curl: search→booking happy path + unified error on bad input |
| 6 | Swagger / OpenAPI | swag annotations, `swag init`, `/swagger/*any` route | `/swagger/index.html` loads; `docs/swagger.json` exists |
| 7 | Docker & Compose | multi-stage `Dockerfile`, `docker-compose.yml` (app+db+migrate), healthchecks | `docker compose up --build` → `/health` ok; full curl flow works |
| 8 | Docs & deliverables | `README` (arch, setup, curl, env), `ANSWERS.md` (§15), "what could be improved" | README complete; repo ready to push |

> Phases 1+2 and 6 are small — they can be merged into adjacent runs if you prefer fewer Codex invocations.
> Keeping them separate gives clean `go test` gates between them.

---

## 11. Deliverables checklist (spec §13)

- [ ] GitHub/GitLab repo (the `flysoft-flight-service/` dir; `git init` in Phase 0, push at the end)
- [ ] README with run instructions (local + Docker)
- [ ] Architecture description (README section, backed by §4 here)
- [ ] curl request examples (README)
- [ ] Docker Compose run instructions (README)
- [ ] "What could be improved with more time" section (README)
- [ ] Answers to the 5 post-task questions (`ANSWERS.md`)

---

## 12. Post-task questions (spec §15) — answer outline (finalize in `ANSWERS.md`)

1. **Add a 2nd provider** → new package implementing `FlightProvider`; register it in a provider registry/slice;
   `flight_service` fans out (parallel `Search`), merges + normalizes results; zero changes to pricing/persistence.
2. **Cache search results** → key by normalized criteria (`from|to|dates|pax|currency`); Redis/in-mem with short TTL
   (e.g. 1–5 min) since fares are volatile; cache the normalized provider offers (pre-pricing) to keep pricing fresh;
   stampede protection via singleflight.
3. **Prevent duplicate bookings** → `Idempotency-Key` header persisted with a unique constraint; same key returns the
   same booking; plus a unique constraint / app check on (offer_id) to stop double-booking one offer.
4. **Logs & monitoring** → `slog` JSON to stdout → shipped (Loki/ELK); `request_id` correlation; Prometheus metrics
   (`/metrics`: RED — rate/errors/duration, provider latency, booking counts); OpenTelemetry traces across
   handler→service→provider→db; alerts on error rate & provider availability.
5. **Future microservice splits** → Search/Provider-Integration service (provider adapters + normalization),
   Pricing service (the money engine), Booking service (offers/bookings persistence + lifecycle), an API gateway.

---

## 13. What could be improved (future work — state in README)

Provider timeout + retry/backoff; idempotency keys; search-result caching; auth & rate limiting; Prometheus
metrics + OTel tracing; offer-expiry sweeper job; richer multi-provider fan-out; contract/integration tests
(testcontainers); CI pipeline (lint+test+build).

---

## 14. Risks & things to watch

- **Money correctness** is the #1 graded item — the pricing unit test must encode the spec's worked example exactly.
- Keep **business logic out of handlers** — reviewers look for this specifically.
- `from`/`to` reserved-word handling — documented as `origin`/`destination` columns.
- GORM is used for queries only; schema is owned by **migrations** (don't `AutoMigrate` in prod path).
- Docker network/daemon access under the Codex sandbox — see the run notes in `CODEX_PROMPTS.md`.
