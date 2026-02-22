package domain

import "time"

type VehicleAssignment struct {
	ID         string
	DriverID   string
	VehicleID  string
	ContractID string
	StartTime  time.Time
	EndTime    *time.Time
	DeletedAt  *time.Time
}
