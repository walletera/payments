# Payments

[![Go](https://github.com/walletera/payments/actions/workflows/release.yml/badge.svg)](https://github.com/walletera/payments/actions/workflows/release.yml)

## Overview

Event-sourced payments microservice for the [Walletera](https://walletera.dev/) platform. Manages payment lifecycle (create, update, retrieve) using EventStoreDB as the event store and RabbitMQ for async event publishing. Exposes a JWT-authenticated public API and an internal private API.

## Getting Started

### Prerequisites

- Go 1.23 or newer
- Docker (required — tests spin up EventStoreDB, RabbitMQ, and Mockserver via testcontainers-go)

### Run tests

```bash
git clone https://github.com/walletera/payments.git
cd payments
```

Run all tests:
```bash
make test_all
```

Or run each suite individually:
```bash
make test_publicapi
make test_privateapi
```

## API Documentation

The service exposes two HTTP APIs:

- **Public API** — JWT-authenticated, for external integrations and client applications.
- **Private API** — Internal, for trusted communication within the Walletera platform.

OpenAPI specifications for both APIs are maintained in the [payments-types repository](https://github.com/walletera/payments-types/tree/main/openapi):

- [Public API](https://github.com/walletera/payments-types/blob/main/openapi/public-api.yaml)
- [Private API](https://github.com/walletera/payments-types/blob/main/openapi/private-api.yaml)

### Endpoints

| Method | Path | API |
|---|---|---|
| `POST` | `/payments` | Public, Private |
| `GET` | `/payments/{paymentId}` | Public, Private |
| `PATCH` | `/payments/{paymentId}` | Private |
