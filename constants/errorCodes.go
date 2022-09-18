package constants

// List of error codes returned by the server in the failed responses body
const (
	VAL_MISSING_FIRST_NAME = "VAL-001"
	VAL_MISSING_LAST_NAME  = "VAL-002"
	VAL_MISSING_EMAIL      = "VAL-003"
	VAL_MISSING_PASSWORD   = "VAL-004"

	VAL_INVALID_FIRST_NAME = "VAL-011"
	VAL_INVALID_LAST_NAME  = "VAL-012"
	VAL_INVALID_EMAIL      = "VAL-013"
	VAL_INVALID_PASSWORD   = "VAL-014"

	VAL_EMAIL_ALREADY_EXISTS = "VAL-023"
)
