# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Go reference implementation demonstrating Uber FX dependency injection with Hexagonal Architecture (Ports & Adapters). Fleet management domain with Legal Entities, Fleets, Vehicles, Drivers, Contracts, and Vehicle Assignments. Uses Go 1.24, Uber FX v1.24, and Uber Zap for structured logging. PostgreSQL storage with master/replica read splitting (pgx/v5), goose migrations, and plain SQL. All entities support soft-delete and undelete.

## Environment Variables

| Variable                   | Description                                                                                                               |
| -------------------------- | ------------------------------------------------------------------------------------------------------------------------- |
| `DATABASE_MASTER_URL`      | PostgreSQL connection string for writes (required)                                                                        |
| `DATABASE_REPLICA_URL`     | PostgreSQL connection string for reads (optional; uses master if unset)                                                   |
| `DRIVER_LICENSE_GRPC_ADDR` | Address of the driver license validation gRPC service (required for driver creation; if unset, POST /drivers returns 503) |
| `HTTP_ADDR`                | Server listen address (default `:8080`)                                                                                   |
| `LOG_DEV`                  | Enable development logging (`true`/`false`)                                                                               |

## Build & Run Commands

```bash
go build ./cmd/server        # Build the server binary
go run ./cmd/server          # Run the server (listens on :8080; requires DATABASE_MASTER_URL)
go test ./...                # Run all tests
go test ./internal/core/...  # Run tests for a specific package subtree
go vet ./...                 # Static analysis
go mod tidy                  # Clean up dependencies
go generate ./internal/core/ports/...  # Regenerate gomock mocks
make proto-generate                    # Generate Go code from protobuf to internal/gen/ (requires buf, protoc-gen-go, protoc-gen-go-grpc)

# Database migrations (goose; optional, app runs them on startup)
go get -tool github.com/pressly/goose/v3/cmd/goose   # Add goose as Go 1.24 tool
go tool goose -dir migrations postgres "$DATABASE_MASTER_URL" status
go tool goose -dir migrations postgres "$DATABASE_MASTER_URL" up
```

## Domain Model

**LegalEntity** — `ID`, `Name`, `TaxID`, `DeletedAt *time.Time`

**Fleet** — `ID`, `LegalEntityID`, `Name`, `DeletedAt *time.Time`

**Vehicle** — `ID`, `FleetID`, `Make`, `Model`, `Year`, `LicensePlate`, `DeletedAt *time.Time`

**Driver** — `ID`, `FirstName`, `LastName`, `LicenseNumber`, `DeletedAt *time.Time`

**Contract** — `ID`, `DriverID`, `LegalEntityID`, `FleetID`, `StartDate`, `EndDate`, `TerminatedAt`, `TerminatedBy`, `DeletedAt *time.Time`. Driver concludes contract with legal entity for a specific fleet.

**VehicleAssignment** — `ID`, `DriverID`, `VehicleID`, `ContractID`, `StartTime`, `EndTime *time.Time`, `DeletedAt *time.Time`. Links driver to vehicle for a limited time under an active contract.

### Business Rules

1. **Soft-delete**: All entities use `DeletedAt` for soft-delete. List/find queries exclude soft-deleted records. `Undelete` restores by clearing `DeletedAt`.
2. **Contract overlap**: For the same `(driverID, legalEntityID, fleetID)`, contract date ranges must not intersect.
3. **Vehicle assignment** requires an active (non-terminated, within date range) contract for the fleet.
4. **One vehicle per driver per fleet**: Only one active assignment (`EndTime == nil`) per driver per fleet at a time.
5. **Driver deletion preconditions**: Driver cannot be soft-deleted while having active contracts or active vehicle assignments.
6. **Driver creation validation**: Before creating a driver, the license is validated via the external gRPC service. Creation fails with 422 if validation returns `not_found` or `data_mismatch`; 503 if the validation service is unavailable.

## Architecture

The project follows **Hexagonal Architecture** with strict layer separation:

**Domain** (`internal/core/domain/`) — Pure domain models and errors. No external dependencies.

**Ports** (`internal/core/ports/`) — Interfaces: `LegalEntityRepository`, `FleetRepository`, `VehicleRepository`, `DriverRepository`, `ContractRepository`, `VehicleAssignmentRepository`; `DriverLicenseValidator` (output port for external validation); and corresponding service interfaces.

**Services** (`internal/core/services/`) — Business logic: `legalentity/`, `fleet/`, `vehicle/`, `driver/`, `contract/`, `assignment/`.

**Input Adapters** (`internal/adapters/in/http/`) — HTTP handlers per resource. Uses `go-chi/chi/v5`. Multiple handlers collected via `fx.Group("routes")`.

**Output Adapters** (`internal/adapters/out/`) — `postgres/`: PostgreSQL implementations with master/replica connection pools, goose migrations (embedded), plain SQL (pgx/v5). `grpc/`: gRPC clients for external services (e.g. `driverlicense/` for driver license validation).

**Generated Code** (`internal/gen/`) — Protobuf-generated Go stubs (from `proto/` via `make proto-generate`).

### Uber FX Module Composition

```go
fx.New(
    telemetry.Module(),
    config.Module(),
    postgres.Module(),
    grpcAdapter.Module(),
    services.Module(),
    httpAdapter.Module(),
).Run()
```

HTTP handlers are provided with `fx.ResultTags(\`group:"routes"\`)`and the server receives`[]RouteRegistrar`via`fx.ParamTags(\`\`, \`group:"routes"\`)`.

## API Endpoints

| Resource          | Method | Path                                     | Description                                                                                                     |
| ----------------- | ------ | ---------------------------------------- | --------------------------------------------------------------------------------------------------------------- |
| LegalEntity       | POST   | `/legal-entities`                        | Create                                                                                                          |
|                   | GET    | `/legal-entities`                        | List all                                                                                                        |
|                   | GET    | `/legal-entities/{id}`                   | Get by ID                                                                                                       |
|                   | DELETE | `/legal-entities/{id}`                   | Soft-delete                                                                                                     |
|                   | POST   | `/legal-entities/{id}/undelete`          | Restore                                                                                                         |
| Fleet             | POST   | `/legal-entities/{legalEntityId}/fleets` | Create fleet                                                                                                    |
|                   | GET    | `/legal-entities/{legalEntityId}/fleets` | List fleets                                                                                                     |
|                   | GET    | `/fleets/{id}`                           | Get by ID                                                                                                       |
|                   | DELETE | `/fleets/{id}`                           | Soft-delete                                                                                                     |
|                   | POST   | `/fleets/{id}/undelete`                  | Restore                                                                                                         |
| Vehicle           | POST   | `/fleets/{fleetId}/vehicles`             | Create vehicle                                                                                                  |
|                   | GET    | `/fleets/{fleetId}/vehicles`             | List vehicles                                                                                                   |
|                   | GET    | `/vehicles/{id}`                         | Get by ID                                                                                                       |
|                   | DELETE | `/vehicles/{id}`                         | Soft-delete                                                                                                     |
|                   | POST   | `/vehicles/{id}/undelete`                | Restore                                                                                                         |
| Driver            | POST   | `/drivers`                               | Create (validates license via gRPC; 422 on validation failure, 503 if service unavailable)                      |
|                   | GET    | `/drivers`                               | List all                                                                                                        |
|                   | GET    | `/drivers/{id}`                          | Get by ID                                                                                                       |
|                   | DELETE | `/drivers/{id}`                          | Soft-delete (requires no active contracts/assignments)                                                          |
|                   | POST   | `/drivers/{id}/undelete`                 | Restore                                                                                                         |
|                   | POST   | `/drivers/{id}/validate`                 | Validate driver license via external gRPC service (response: `driver_id`, `result`: ok/not_found/data_mismatch) |
| Contract          | POST   | `/drivers/{driverId}/contracts`          | Create (JSON: `legal_entity_id`, `fleet_id`, `start_date`, `end_date` as YYYY-MM-DD)                            |
|                   | GET    | `/drivers/{driverId}/contracts`          | List driver's contracts                                                                                         |
|                   | GET    | `/contracts/{id}`                        | Get by ID                                                                                                       |
|                   | POST   | `/contracts/{id}/terminate`              | Terminate (JSON: `terminated_by`)                                                                               |
|                   | DELETE | `/contracts/{id}`                        | Soft-delete                                                                                                     |
|                   | POST   | `/contracts/{id}/undelete`               | Restore                                                                                                         |
| VehicleAssignment | POST   | `/contracts/{contractId}/assignments`    | Assign vehicle (JSON: `vehicle_id`)                                                                             |
|                   | GET    | `/contracts/{contractId}/assignments`    | List assignments                                                                                                |
|                   | GET    | `/assignments/{id}`                      | Get by ID                                                                                                       |
|                   | POST   | `/assignments/{id}/return`               | Return vehicle                                                                                                  |
|                   | DELETE | `/assignments/{id}`                      | Soft-delete                                                                                                     |
|                   | POST   | `/assignments/{id}/undelete`             | Restore                                                                                                         |

- `GET /health` — Health check

## Key Conventions

- Each FX module lives in an `fx.go` file alongside its implementation
- Constructor functions with explicit parameters
- Domain errors in `core/domain/errors.go`: `ErrNotFound`, `ErrInvalidInput`, `ErrConflict`, `ErrContractNotActive`, `ErrVehicleAlreadyAssigned`, `ErrDriverHasActiveContracts`, `ErrDriverHasActiveAssignments`, `ErrAlreadyDeleted`, `ErrValidationServiceUnavailable`, `ErrLicenseValidationFailed`
- HTTP handlers map domain errors to HTTP status codes via `mapDomainErrorToStatus`
