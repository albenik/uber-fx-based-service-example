package domain

import "errors"

var (
	ErrNotFound               = errors.New("entity not found")
	ErrEntityNotFound          = errors.New("entity not found")
	ErrInvalidInput            = errors.New("invalid input")
	ErrConflict                = errors.New("conflict")
	ErrContractNotActive       = errors.New("contract is not active")
	ErrVehicleAlreadyAssigned  = errors.New("driver already has an active vehicle assignment for this fleet")
	ErrDriverHasActiveContracts   = errors.New("driver has active contracts; terminate them before deletion")
	ErrDriverHasActiveAssignments = errors.New("driver has active vehicle assignments; return vehicles before deletion")
	ErrAlreadyDeleted            = errors.New("entity is already deleted")
	ErrValidationServiceUnavailable = errors.New("driver license validation service not available")
)
