package domain

import "time"

type Fleet struct {
	ID           string
	LegalEntityID string
	Name         string
	DeletedAt    *time.Time
}
