# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Commands

```bash
# Run all tests (private API, then public API)
make test_all

# Run only private API tests
make test_privateapi

# Run only public API tests
make test_publicapi

# Run a single test file or scenario (example)
go test --tags=privateapi -count=1 -v ./internal/tests/privateapi/...
go test --tags=publicapi -count=1 -v ./internal/tests/publicapi/...
```

Tests require Docker to spin up EventStoreDB, RabbitMQ, and Mockserver containers via testcontainers-go.

## Architecture

This is a **payments microservice** using Event Sourcing and dual HTTP APIs (public + private).

### Layers

- **`cmd/`** — Entry point; loads env vars, wires dependencies, starts the app
- **`internal/app/`** — Bootstrap: starts HTTP servers, connects ESDB and RabbitMQ, manages graceful shutdown
- **`internal/adapters/input/http/`** — HTTP handlers implementing generated interfaces from `payments-types`
  - `public/` — JWT-authenticated public API
  - `private/` — Internal API (no auth)
- **`internal/domain/payment/`** — Business logic and aggregate
  - `service.go` — CreatePayment, UpdatePayment, GetPayment (reconstructs aggregate from event stream)
  - `aggregate.go` — Payment aggregate with status transition validation (Pending → Delivered/Confirmed/Failed)
  - `event/handlers/` — Publishes domain events to RabbitMQ
- **`pkg/`** — Shared utilities: JWT auth, UUID generation, log attribute constants

### Data Flow

HTTP request → Handler → `payment.Service` → append/read events from EventStoreDB → publish events to RabbitMQ

State is never stored directly; it's always reconstructed by replaying events from the ESDB stream.

### External Dependencies

| Dependency | Purpose |
|---|---|
| EventStoreDB | Event store (persistent subscription group: `payments-service`, projection: `$ce-payments.service.payment`) |
| RabbitMQ | Async event publishing (exchange: `payments.events`, routing keys: `payment.created`, `payment.updated`) |
| `payments-types` package | Generated OpenAPI types, HTTP server interfaces, and event definitions |

### Required Environment Variables

```
EVENTSTOREDB_URL
RABBITMQ_HOST, RABBITMQ_PORT, RABBITMQ_USER, RABBITMQ_PASSWORD
PUBLIC_API_HTTP_SERVER_PORT, PRIVATE_API_HTTP_SERVER_PORT
BASE64_AUTH_PUB_KEY   # Base64-encoded RSA public key for JWT validation
```

## Testing

Tests are BDD-style using [godog](https://github.com/cucumber/godog) with Gherkin feature files located in `internal/tests/{privateapi,publicapi}/features/`.

- Build tags `privateapi` / `publicapi` select which suite to run
- `TestMain` in each suite starts real Docker containers (EventStoreDB, RabbitMQ, Mockserver)
- Test state is passed via `context.WithValue`; log watching via `slogwatcher` verifies service behavior
- JSON assertions use `${json-unit.regex}` / `${json-unit.any-string}` placeholders

Mocks are generated with [mockery](https://github.com/vektra/mockery) using `.mockery.yaml`; regenerate with `mockery`.

- **Shared ESDB state across test suites** — `TestCreatePayment` and `TestGetPayment` run against the same EventStoreDB container within a `make test_publicapi` run. Payment IDs in testdata files must be unique across all suites to avoid 409 conflicts. The naming convention for testdata files encodes the UUID prefix (e.g. `_bdf4`, `_af2e`) to make ownership visible at a glance.

- **`payments-types` is vendored** — generated types, validators, and HTTP interfaces live under `vendor/github.com/walletera/payments-types/`. Changes to the OpenAPI specs require regenerating and re-vendoring. Domain validation that goes beyond what the generated validators enforce (e.g. `amount > 0`) belongs in `internal/domain/payment/service.go:validatePayment`.
