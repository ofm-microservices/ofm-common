# OFM Common

## Purpose

`ofm-common` is the shared Go module for infrastructure-grade building blocks
used across OFM services.

It exists to remove low-value duplication between service repositories while
preserving strict bounded-context ownership. The module should stay small and
focused on shared contracts and generic infrastructure code.

## Current Packages

- `pkg/logging`
  Shared structured logging abstraction and zap-backed implementation.
- `proto/registration/v1`
  Shared registration gRPC contract and generated Go stubs.
- `proto/auth/v1`
  Shared auth query gRPC contract and generated Go stubs.
- `proto/user/v1`
  Shared user query gRPC contract and generated Go stubs.

## Run

This repository is a library module, not a runnable service.

To verify it builds:

```bash
go build ./...
```

To regenerate protobuf code:

```bash
just proto-gen
```

To lint the protobuf module with Buf:

```bash
just buf-lint
```

## Buf

`ofm-common` is configured as a Buf workspace rooted at `proto/`.

- `buf.yaml` defines the module and lint/breaking policy
- `buf.gen.yaml` defines Go and gRPC stub generation

Buf is the source of truth for the shared protobuf contracts. The readable
schema documentation is expected to come from the Buf Schema Registry after
publishing the module, not from a local HTML generator.

## Module Path

```text
github.com/ofm-microservices/ofm-common
```

Example import:

```go
import "github.com/ofm-microservices/ofm-common/pkg/logging"
```

## Technologies

- Go
- Zap for structured JSON logging

Main libraries from `go.mod`:

- `go.uber.org/zap`

## Architecture Notes

- keep the shared surface small
- prefer generic infrastructure and shared contracts only
- optimize for Docker and CI consumption through the real GitHub module path
