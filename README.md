# Uber FX Based Service Example

[![CI](https://github.com/albenik/uber-fx-based-service-example/actions/workflows/go.yml/badge.svg)](https://github.com/albenik/uber-fx-based-service-example/actions/workflows/go.yml)

A reference Go implementation demonstrating **Hexagonal Architecture** (Ports & Adapters) with
[Uber FX](https://github.com/uber-go/fx) dependency injection. The domain models a fleet management system for managing
legal entities, fleets, vehicles, drivers, contracts, and vehicle assignments.

> **For AI assistants and contributors:** detailed domain model, business rules, architecture conventions, environment
> variables, build commands, and API endpoint reference are maintained in [CLAUDE.md](CLAUDE.md).

## Table of Contents

- [Uber FX Based Service Example](#uber-fx-based-service-example)
  - [Table of Contents](#table-of-contents)
  - [Technology Stack](#technology-stack)
  - [Architecture Overview](#architecture-overview)
  - [Prerequisites](#prerequisites)
  - [Quick Start](#quick-start)
  - [Project Structure](#project-structure)
  - [Key Design Decisions](#key-design-decisions)
    - [Why Hexagonal Architecture?](#why-hexagonal-architecture)
    - [Why Uber FX?](#why-uber-fx)
    - [Master/replica splitting](#masterreplica-splitting)
    - [Migrations on startup](#migrations-on-startup)
    - [Centralised error mapping](#centralised-error-mapping)
  - [License](#license)

---

## Technology Stack

| Concern              | Library / Tool                                                          |
| -------------------- | ----------------------------------------------------------------------- |
| Dependency injection | [Uber FX v1.24](https://github.com/uber-go/fx)                          |
| HTTP routing         | [go-chi/chi v5](https://github.com/go-chi/chi)                          |
| Structured logging   | [Uber Zap v1.27](https://github.com/uber-go/zap)                        |
| PostgreSQL driver    | [pgx/v5](https://github.com/jackc/pgx) with master/replica splitting    |
| Database migrations  | [goose v3](https://github.com/pressly/goose) (embedded, run on startup) |
| Mocks                | [uber-go/mock](https://github.com/uber-go/mock)                         |
| External validation  | gRPC (protobuf-defined `DriverLicenseValidationService`)                |
| Protobuf toolchain   | [buf](https://buf.build)                                                |

---

## Architecture Overview

The project follows **Hexagonal Architecture** with strict separation between the domain core and infrastructure.

```plaintext
┌───────────────────────────────────────────────────────────┐
│                        Driving Side                       │
│                   (Input Adapters / Primary)              │
│                 internal/adapters/in/http/                │
│           HTTP handlers using go-chi/chi/v5               │
└───────────────────────────┬───────────────────────────────┘
                            │ calls via Port interfaces
┌───────────────────────────▼───────────────────────────────┐
│                          Core                             │
│  ┌─────────────────────────────────────────────────────┐  │
│  │          Domain  (internal/core/domain/)            │  │
│  │  Pure models + sentinel errors. No dependencies.    │  │
│  └─────────────────────────────────────────────────────┘  │
│  ┌─────────────────────────────────────────────────────┐  │
│  │          Ports  (internal/core/ports/)              │  │
│  │  Go interfaces for repos + service + validators.    │  │
│  └─────────────────────────────────────────────────────┘  │
│  ┌─────────────────────────────────────────────────────┐  │
│  │         Services (internal/core/services/)          │  │
│  │  Business logic only; depends on port interfaces.   │  │
│  └─────────────────────────────────────────────────────┘  │
└───────────────────────────┬───────────────────────────────┘
                            │ implemented by
┌───────────────────────────▼───────────────────────────────┐
│                     Driven Side                           │
│              (Output Adapters / Secondary)                │
│  internal/adapters/out/postgres/   — PostgreSQL (pgx/v5)  │
│  internal/adapters/out/grpc/       — gRPC client          │
└───────────────────────────────────────────────────────────┘
```

For a full breakdown of each layer, FX module composition, domain model, business rules, and API endpoints, see
[CLAUDE.md](CLAUDE.md).

---

## Prerequisites

- **Go 1.26+**
- **PostgreSQL** (any recent version)
- **buf** CLI — only needed to regenerate protobuf stubs (`make proto-generate`)
- **golangci-lint** — only needed for `make lint`

---

## Quick Start

```bash
export DATABASE_MASTER_URL="postgres://postgres:secret@localhost:5432/fleet?sslmode=disable"
export DRIVER_LICENSE_GRPC_ADDR="localhost:50051"  # optional; omit to disable license validation

go run ./cmd/server
# or: make build && ./bin/server
```

The server automatically runs pending database migrations on startup and listens on `:8080` by default.

For all environment variables, `make` targets, and database migration commands, see [CLAUDE.md](CLAUDE.md).

---

## Project Structure

```plaintext
.
├── cmd/server/          # Main entry point; wires FX modules
├── internal/
│   ├── adapters/
│   │   ├── in/http/     # HTTP handlers (chi router), one file per resource
│   │   └── out/
│   │       ├── grpc/    # gRPC output adapters (driverlicense client)
│   │       └── postgres/# PostgreSQL repositories, master/replica pools
│   ├── config/          # Env-based config structs + FX providers
│   ├── core/
│   │   ├── domain/      # Pure domain models and sentinel errors
│   │   ├── ports/       # Repository, service and validator interfaces
│   │   └── services/    # Business logic (legalentity, fleet, vehicle,
│   │                    #   driver, contract, assignment)
│   ├── gen/             # Protobuf-generated code (do not edit)
│   └── telemetry/       # Zap logger setup + FX provider
├── migrations/          # goose SQL migrations (embedded in binary)
├── proto/               # Protobuf source definitions
├── buf.yaml             # buf configuration
├── buf.gen.yaml         # buf code generation config
├── Makefile             # Developer task runner (`make help` lists all targets)
└── go.mod
```

Each FX module exposes its wiring in an `fx.go` file alongside its implementation.

---

## Key Design Decisions

### Why Hexagonal Architecture?

It makes the domain and business rules independently testable. Services depend only on Go interfaces (`ports`), so the
entire core can be unit-tested with mocks — no database required.

### Why Uber FX?

FX provides reflection-based constructor wiring with compile-time dependency checking. Each layer declares what it
provides and what it needs; the container assembles the application automatically, detects missing dependencies, and
manages lifecycle hooks (`OnStart`/`OnStop`).

### Master/replica splitting

Write operations target `DATABASE_MASTER_URL`; read operations use `DATABASE_REPLICA_URL` when set, falling back to
master. This allows horizontal read scaling with zero application-level changes.

### Migrations on startup

Goose migrations are embedded with `//go:embed` and run automatically at startup. Deployment stays atomic — no separate
migration job needed.

### Centralised error mapping

Domain sentinel errors (`ErrNotFound`, `ErrConflict`, etc.) are defined once in `internal/core/domain/errors.go` and
mapped to HTTP status codes in a single place in `internal/adapters/in/http/common.go`. Business logic never mentions
HTTP.

---

## License

MIT © 2026 Veniamin Albaev. See [LICENSE](LICENSE) for the full text.
