package domain

import "time"

type Vehicle struct {
	ID           string
	FleetID      string
	Make         string
	Model        string
	Year         int
	LicensePlate string
	DeletedAt    *time.Time
}
