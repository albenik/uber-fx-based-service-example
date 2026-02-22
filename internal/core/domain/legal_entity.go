package domain

import "time"

type LegalEntity struct {
	ID        string
	Name      string
	TaxID     string
	DeletedAt *time.Time
}
