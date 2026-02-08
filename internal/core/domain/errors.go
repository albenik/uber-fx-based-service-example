package domain

import "errors"

var (
	ErrEntityNotFound      = errors.New("entity not found")
	ErrEntityAlreadyExists = errors.New("entity already exists")
)
