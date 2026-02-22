---
name: Driver License Pre-Create Validation
overview: Add mandatory driver license validation via DriverLicenseValidator before every driver creation. If validation fails or the service is unavailable, return an error and do not persist the driver.
todos:
  - id: domain-error
    content: Add `ErrLicenseValidationFailed` to `internal/core/domain/errors.go`
    status: completed
  - id: service-create
    content: Add `ValidateLicense` call in `Service.Create()` before ID generation and save
    status: completed
  - id: http-mapping
    content: Map `ErrLicenseValidationFailed` to HTTP 422 in `common.go` and `handler_driver.go`
    status: completed
  - id: unit-tests
    content: Add unit tests for validation-on-create in `service_test.go`
    status: completed
isProject: false
---

# Driver License Pre-Create Validation

## Current State

- `Service.Create()` in `[internal/core/services/driver/service.go](internal/core/services/driver/service.go)` validates inputs, generates an ID, and saves directly to the repository -- no license validation.
- `DriverLicenseValidator` is already injected into the driver service but only used by the `ValidateLicense()` method (called from `POST /drivers/{id}/validate`).
- The noop validator (no gRPC configured) returns `ErrValidationServiceUnavailable`, which maps to HTTP 503.

## Changes

### 1. Add `ErrLicenseValidationFailed` domain error

**File:** `[internal/core/domain/errors.go](internal/core/domain/errors.go)`

Add a new sentinel error:

```go
ErrLicenseValidationFailed = errors.New("driver license validation failed")
```

### 2. Call validator in `Service.Create()`

**File:** `[internal/core/services/driver/service.go](internal/core/services/driver/service.go)`

After input validation (the three `TrimSpace` + empty checks) and before ID generation, call `s.validator.ValidateLicense(ctx, firstName, lastName, licenseNumber)`. If the call returns an error, propagate it. If the result is not `domain.LicenseValid`, return a wrapped `ErrLicenseValidationFailed` with the result detail (e.g. `"driver license validation failed: not_found"`).

### 3. Map new error to HTTP 422

**File:** `[internal/adapters/in/http/common.go](internal/adapters/in/http/common.go)`

Add a case for `ErrLicenseValidationFailed` returning `http.StatusUnprocessableEntity` (422).

**File:** `[internal/adapters/in/http/handler_driver.go](internal/adapters/in/http/handler_driver.go)`

Add `domain.ErrLicenseValidationFailed` to the known-error check in `handleError` so it returns the domain message instead of "internal server error".

### 4. Add unit tests

**File:** `[internal/core/services/driver/service_test.go](internal/core/services/driver/service_test.go)`

Add three test cases:

- **Create succeeds when validation returns `ok`** -- validator returns `LicenseValid`, repo `Save` is called, driver returned.
- **Create fails when validation returns `not_found`** -- validator returns `LicenseNotFound`, no repo interaction, error wraps `ErrLicenseValidationFailed`.
- **Create fails when validator returns error** -- validator returns an error (e.g. service unavailable), error propagated, no repo interaction.

