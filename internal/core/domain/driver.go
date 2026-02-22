package domain

import "time"

type Driver struct {
	ID            string
	FirstName     string
	LastName      string
	LicenseNumber string
	DeletedAt     *time.Time
}
