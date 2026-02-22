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

func (e *exposableError) Error() string   { return e.msg }
func (e *exposableError) Exposable()      {}

var (
	ErrNotFound                 = &exposableError{msg: "entity not found"}
	ErrInvalidInput             = &exposableError{msg: "invalid input"}
	ErrConflict                 = &exposableError{msg: "conflict"}
	ErrContractNotActive        = &exposableError{msg: "contract is not active"}
	ErrVehicleAlreadyAssigned   = &exposableError{msg: "driver already has an active vehicle assignment for this fleet"}
	ErrDriverHasActiveContracts = &exposableError{msg: "driver has active contracts; terminate them before deletion"}
	ErrDriverHasActiveAssignments = &exposableError{msg: "driver has active vehicle assignments; return vehicles before deletion"}
	ErrAlreadyDeleted           = &exposableError{msg: "entity is already deleted"}
	ErrValidationServiceUnavailable = &exposableError{msg: "driver license validation service not available"}
	ErrLicenseValidationFailed  = &exposableError{msg: "driver license validation failed"}
)

// IsExposable reports whether err (or any error in its chain) implements Exposable.
func IsExposable(err error) bool {
	var e Exposable
	return errors.As(err, &e)
}
