# Payments

The Payments service is part of the [Walletera](https://walletera.dev/designing-a-digital-wallet) project.
It exposes two APIs, an external one used by end user to create and consult withdrawals and deposits, and an internal one
used by the payments gateways to create deposits and update the status of withdrawals.

What's on this repo
- the Payments API OpenApi specification
- a generated golang client for the Payments API
- TODO 

## How to generate the golang client

To generate the golang client code from the openapi specification you need to install [ogen](https://ogen.dev/) by executing in a terminal
```bash
go install -v github.com/ogen-go/ogen/cmd/ogen@latest
```

Once you have the generator installed open a terminal an execute the following command
```bash
go generate ./...
```
