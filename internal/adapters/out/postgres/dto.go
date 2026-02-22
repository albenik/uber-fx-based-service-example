package postgres

import (
	"time"

	"github.com/albenik/uber-fx-based-service-example/internal/core/domain"
)

type legalEntityRow struct {
	ID        string     `db:"id"`
	Name      string     `db:"name"`
	TaxID     string     `db:"tax_id"`
	DeletedAt *time.Time `db:"deleted_at"`
}

func (r *legalEntityRow) toDomain() *domain.LegalEntity {
	return &domain.LegalEntity{ID: r.ID, Name: r.Name, TaxID: r.TaxID, DeletedAt: r.DeletedAt}
}

func legalEntityToRow(e *domain.LegalEntity) *legalEntityRow {
	return &legalEntityRow{ID: e.ID, Name: e.Name, TaxID: e.TaxID, DeletedAt: e.DeletedAt}
}

type fleetRow struct {
	ID            string     `db:"id"`
	LegalEntityID string     `db:"legal_entity_id"`
	Name          string     `db:"name"`
	DeletedAt     *time.Time `db:"deleted_at"`
}

func (r *fleetRow) toDomain() *domain.Fleet {
	return &domain.Fleet{ID: r.ID, LegalEntityID: r.LegalEntityID, Name: r.Name, DeletedAt: r.DeletedAt}
}

func fleetToRow(e *domain.Fleet) *fleetRow {
	return &fleetRow{ID: e.ID, LegalEntityID: e.LegalEntityID, Name: e.Name, DeletedAt: e.DeletedAt}
}

type vehicleRow struct {
	ID           string     `db:"id"`
	FleetID      string     `db:"fleet_id"`
	Make         string     `db:"make"`
	Model        string     `db:"model"`
	Year         int        `db:"year"`
	LicensePlate string     `db:"license_plate"`
	DeletedAt    *time.Time `db:"deleted_at"`
}

func (r *vehicleRow) toDomain() *domain.Vehicle {
	return &domain.Vehicle{
		ID:           r.ID,
		FleetID:      r.FleetID,
		Make:         r.Make,
		Model:        r.Model,
		Year:         r.Year,
		LicensePlate: r.LicensePlate,
		DeletedAt:    r.DeletedAt,
	}
}

func vehicleToRow(e *domain.Vehicle) *vehicleRow {
	return &vehicleRow{
		ID: e.ID, FleetID: e.FleetID, Make: e.Make, Model: e.Model,
		Year: e.Year, LicensePlate: e.LicensePlate, DeletedAt: e.DeletedAt,
	}
}

type driverRow struct {
	ID            string     `db:"id"`
	FirstName     string     `db:"first_name"`
	LastName      string     `db:"last_name"`
	LicenseNumber string     `db:"license_number"`
	DeletedAt     *time.Time `db:"deleted_at"`
}

func (r *driverRow) toDomain() *domain.Driver {
	return &domain.Driver{
		ID: r.ID, FirstName: r.FirstName, LastName: r.LastName,
		LicenseNumber: r.LicenseNumber, DeletedAt: r.DeletedAt,
	}
}

func driverToRow(e *domain.Driver) *driverRow {
	return &driverRow{
		ID: e.ID, FirstName: e.FirstName, LastName: e.LastName,
		LicenseNumber: e.LicenseNumber, DeletedAt: e.DeletedAt,
	}
}

type contractRow struct {
	ID            string     `db:"id"`
	DriverID      string     `db:"driver_id"`
	LegalEntityID string     `db:"legal_entity_id"`
	FleetID       string     `db:"fleet_id"`
	StartDate     time.Time  `db:"start_date"`
	EndDate       time.Time  `db:"end_date"`
	TerminatedAt  *time.Time `db:"terminated_at"`
	TerminatedBy  string     `db:"terminated_by"`
	DeletedAt     *time.Time `db:"deleted_at"`
}

func (r *contractRow) toDomain() *domain.Contract {
	return &domain.Contract{
		ID: r.ID, DriverID: r.DriverID, LegalEntityID: r.LegalEntityID, FleetID: r.FleetID,
		StartDate: r.StartDate, EndDate: r.EndDate,
		TerminatedAt: r.TerminatedAt, TerminatedBy: r.TerminatedBy, DeletedAt: r.DeletedAt,
	}
}

func contractToRow(e *domain.Contract) *contractRow {
	return &contractRow{
		ID: e.ID, DriverID: e.DriverID, LegalEntityID: e.LegalEntityID, FleetID: e.FleetID,
		StartDate: e.StartDate, EndDate: e.EndDate,
		TerminatedAt: e.TerminatedAt, TerminatedBy: e.TerminatedBy, DeletedAt: e.DeletedAt,
	}
}

type vehicleAssignmentRow struct {
	ID         string     `db:"id"`
	DriverID   string     `db:"driver_id"`
	VehicleID  string     `db:"vehicle_id"`
	ContractID string     `db:"contract_id"`
	StartTime  time.Time  `db:"start_time"`
	EndTime    *time.Time `db:"end_time"`
	DeletedAt  *time.Time `db:"deleted_at"`
}

func (r *vehicleAssignmentRow) toDomain() *domain.VehicleAssignment {
	return &domain.VehicleAssignment{
		ID: r.ID, DriverID: r.DriverID, VehicleID: r.VehicleID, ContractID: r.ContractID,
		StartTime: r.StartTime, EndTime: r.EndTime, DeletedAt: r.DeletedAt,
	}
}

func vehicleAssignmentToRow(e *domain.VehicleAssignment) *vehicleAssignmentRow {
	return &vehicleAssignmentRow{
		ID: e.ID, DriverID: e.DriverID, VehicleID: e.VehicleID, ContractID: e.ContractID,
		StartTime: e.StartTime, EndTime: e.EndTime, DeletedAt: e.DeletedAt,
	}
}
