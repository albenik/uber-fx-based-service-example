package domain

import "time"

type Contract struct {
	ID             string
	DriverID       string
	LegalEntityID  string
	FleetID        string
	StartDate      time.Time
	EndDate        time.Time
	TerminatedAt   *time.Time
	TerminatedBy   string
	DeletedAt      *time.Time
}
