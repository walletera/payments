# Payments

[![Go](https://github.com/walletera/payments/actions/workflows/release.yml/badge.svg)](https://github.com/walletera/payments/actions/workflows/release.yml)

## Overview

This repository contains the **Payments** service, a core component of **Walletera**, an open-source Payment as a Service (PaaS) platform. Walletera aims to simplify the integration of payment systems, making it easy for developers and businesses to build, scale, and maintain robust digital wallet and payment solutions.

The Payments service is responsible for managing payment processing, transaction handling, and related logic within the broader Walletera ecosystem.

Learn more about the platform, design decisions, and updates by visiting the [Walletera Blog](https://walletera.dev/).

## Key Features

- Modular design, easy to extend or adapt to specific use-cases
- Built with Go for maximum performance and reliability
- Integrates seamlessly with other Walletera components
- Includes robust testing and continuous integration setup

## Getting Started

### Prerequisites

- Go 1.23 or newer
- Docker (optional, for containerized deployment)

### Run tests

Clone the repository:
```
bash git clone [https://github.com/walletera/payments.git](https://github.com/walletera/payments.git) cd payments
``` 

Run:
```
make test-all
```
## API Documentation

The Payments service exposes two HTTP APIs:

- **Public API:** For general operations intended for external integrations and client applications.
- **Private API:** For administrative, internal, and trusted communications within the Walletera platform.

### OpenAPI Specifications

You can find up-to-date OpenAPI (Swagger) specifications for both APIs in the [payments-types repository](https://github.com/walletera/payments-types/tree/main/openapi):

- [Public API OpenAPI definition](https://github.com/walletera/payments-types/blob/main/openapi/publicapi.yml)
- [Private API OpenAPI definition](https://github.com/walletera/payments-types/blob/main/openapi/privateapi.yml)

These specification files fully describe available endpoints, request and response schemas, authentication, and error formats.

#### Examples

- **Create Payment**: `POST /payments` (private API)
- **Get Payment**: `GET /payments/{paymentId}` (both APIs)
- **Update Payment**: `PATCH /payments/{paymentId}` (private API)

Full details of each operation, including expected parameters and example payloads, are maintained in the referenced OpenAPI files.

#### Browsing the API

You can visualize the OpenAPI files using tools such as:

- [Swagger Editor](https://editor.swagger.io/)
- [Redoc](https://redocly.github.io/redoc/)

This allows easy exploration and testing of the API definitions outside the codebase.



