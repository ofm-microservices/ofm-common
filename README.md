# OFM Common

## Purpose

ofm-common is a Go library containing shared contracts and generic infrastructure used by OFM services. It is not a runnable service and must not become a business-domain dumping ground. Status: active library.

## Ownership and contents

The repository owns reusable logging, transport-neutral messaging contracts, protobuf definitions, generated Go stubs, and generic observability helpers. Service-specific entities, repositories, configuration types, and business rules remain in their owning repositories.

Important contract areas include registration, auth, user, order flow, payment flow, and transport/realtime messages. proto/ and Buf configuration are the source of truth for protobuf APIs.

## Local development

    go test ./...
    go build ./...
    just proto-gen
    just buf-lint

The module path is github.com/ofm-microservices/ofm-common. Generated code must be regenerated from schemas; do not edit generated files manually.

## Build and limitations

The repository is consumed as a Go module and is not deployed as a workload. It has no database, broker, HTTP listener, or .env runtime configuration. Changes to shared contracts require compatibility review because many services compile against this module.

