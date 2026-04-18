# OFM Common

## Purpose

`ofm-common` is the shared Go module for infrastructure-grade building blocks
used across OFM services.

It exists to remove low-value duplication between service repositories while
preserving strict bounded-context ownership. Only code that is genuinely generic
and service-agnostic belongs here.

Good candidates for this repository:

- logging
- generic messaging bootstrap helpers
- generic database bootstrap helpers
- observability helpers
- test utilities that are not tied to one domain

Code that does **not** belong here:

- service-specific DTOs
- NATS subjects for one service
- domain entities from bounded contexts
- HTTP handlers
- repository implementations owned by a single service
- auth, user, mail, or saga business rules

## Current Packages

- `pkg/logging`
  Shared structured logging abstraction and zap-backed implementation.

## Run

This repository is a library module, not a runnable service.

To verify it builds:

```bash
go build ./...
```

## Module Path

```text
github.com/ofm-microseervices/ofm-common
```

Example import:

```go
import "github.com/ofm-microseervices/ofm-common/pkg/logging"
```

## Technologies

- Go
- Zap for structured JSON logging

Main libraries from `go.mod`:

- `go.uber.org/zap`

## Architecture Notes

- keep packages generic
- avoid service-bound semantics
- optimize for Docker and CI consumption through the real GitHub module path
- prefer a small shared surface over a large “common” dump
