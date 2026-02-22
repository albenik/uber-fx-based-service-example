package domain

type LicenseValidationResult string

const (
	LicenseValid        LicenseValidationResult = "ok"
	LicenseNotFound     LicenseValidationResult = "not_found"
	LicenseDataMismatch LicenseValidationResult = "data_mismatch"
	LicenseValidationUnknown LicenseValidationResult = "unknown"
)
