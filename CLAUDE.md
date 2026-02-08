# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Go reference implementation demonstrating Uber FX dependency injection with Hexagonal Architecture (Ports & Adapters). Uses Go 1.24, Uber FX v1.24, and Uber Zap for structured logging. In-memory storage only (no database).

## Build & Run Commands

```bash
go build ./cmd/server        # Build the server binary
go run ./cmd/server          # Run the server (listens on 0.0.0.0:8080)
go test ./...                # Run all tests
go test ./internal/core/...  # Run tests for a specific package subtree
go vet ./...                 # Static analysis
go mod tidy                  # Clean up dependencies
```

## Architecture

The project follows **Hexagonal Architecture** with strict layer separation:

**Domain** (`internal/core/domain/`) — Pure domain models (`FooEntity`) and domain errors. No external dependencies.

**Ports** (`internal/core/ports/`) — Interfaces defining contracts between layers:
- `FooEntityRepository` (output port for persistence)
- `FooEntityService` (input port for business operations)

**Services** (`internal/core/services/`) — Business logic implementations of input ports. Depend only on port interfaces, never on concrete adapters.

**Input Adapters** (`internal/adapters/in/http/`) — HTTP handlers translating REST requests into service calls. Uses Go 1.22+ route syntax (`GET /foos/{id}`).

**Output Adapters** (`internal/adapters/out/repository/`) — Concrete implementations of output ports. Currently only in-memory (`MemoryFooEntityRepository` with `sync.RWMutex`).

**Telemetry** (`internal/telemetry/`) — Zap logger initialization, injected across all components.

### Uber FX Module Composition

Each architectural layer exposes an FX module via `fx.go` files. The application is composed in `cmd/server/main.go`:

```
fx.New(telemetry.Module(), repository.Module(), services.Module(), httpAdapter.Module())
```

Interface binding uses `fx.Annotate` with `fx.As` (see `repository/fx.go`, `services/fx.go`). The HTTP adapter registers FX lifecycle hooks for graceful server start/stop.

### Dependency Flow

```
HTTP Handler → ports.FooEntityService → ports.FooEntityRepository → in-memory map
```

Dependencies always point inward: adapters depend on ports, services depend on ports, domain depends on nothing.

## API Endpoints

- `GET /health` — Health check
- `GET /foos` — List all entities
- `POST /foos` — Create entity (JSON: `{"name", "email"}`)
- `GET /foos/{id}` — Get entity by ID
- `DELETE /foos/{id}` — Delete entity

## Key Conventions

- Each FX module lives in an `fx.go` file alongside its implementation
- Constructor injection via struct fields (see `userservice.Service`, `FooEntityHandler`)
- Domain errors defined in `core/domain/errors.go` using `errors.New`
- HTTP handlers map domain errors to appropriate HTTP status codes
