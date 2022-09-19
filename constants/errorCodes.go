package constants

// List of error codes returned by the server in the failed responses body
const (
	SUCCESS = "ACC-000"

	VAL_VALIDATOR_ERROR      = "VAL-000"
	VAL_EMAIL_ALREADY_EXISTS = "VAL-001"

	DATABASE_ERROR = "DB-001"

	AUTH_INVALID_EMAIL    = "AUTH-001"
	AUTH_INVALID_PASSWORD = "AUTH-002"
	AUTH_JWT_FAILURE      = "AUTH-011"
	AUTH_JWT_MISSING      = "AUTH-012"
	AUTH_JWT_INVALID      = "AUTH-013"
	AUTH_JWT_EXPIRED      = "AUTH-014"
)
