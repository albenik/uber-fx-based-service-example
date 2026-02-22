package domain

import "errors"

// Exposable marks errors that may be safely exposed to API clients.
// Only errors implementing this interface have their message returned in HTTP responses.
type Exposable interface {
	Exposable()
}

type exposableError struct {
	msg string
}

func (e *exposableError) Error() string { return e.msg }

func (e *exposableError) Exposable() {}

// IsExposable reports whether err (or any error in its chain) implements Exposable.
func IsExposable(err error) bool {
	var e Exposable
	return errors.As(err, &e)
}

func exposable(msg string) error {
	return &exposableError{msg: msg}
}

var (
	ErrNotFound                     = exposable("entity not found")
	ErrInvalidInput                 = exposable("invalid input")
	ErrConflict                     = exposable("conflict")
	ErrContractNotActive            = exposable("contract is not active")
	ErrDriverAlreadyAssignedInFleet = exposable("driver already has an active vehicle assignment for this fleet")
	ErrDriverHasActiveContracts     = exposable("driver has active contracts; terminate them before deletion")
	ErrDriverHasActiveAssignments   = exposable("driver has active vehicle assignments; return vehicles before deletion")
	ErrAlreadyDeleted               = exposable("entity is already deleted")
	ErrValidationServiceUnavailable = exposable("driver license validation service not available")
	ErrLicenseValidationFailed      = exposable("driver license validation failed")
)
