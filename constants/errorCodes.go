package constants

// List of error codes returned by the server in the failed responses body
const (
	SUCCESS = "ACC-000"

	VAL_VALIDATOR_ERROR        = "VAL-000"
	VAL_EMAIL_ALREADY_EXISTS   = "VAL-001"
	VAL_UUID_INVALID           = "VAL-010"
	VAL_PROVIDER_NOT_SUPPORTED = "VAL-020"

	DATABASE_ERROR          = "DB-001"
	DATABASE_USER_NOT_FOUND = "DB-002"

	TRANSPORT_VOLUME_NOT_FOUND   = "TRN-001"
	TRANSPORT_DISK_NOT_FOUND     = "TRN-002"
	TRANSPORT_DISK_IS_BEING_USED = "TRN-003"
	TRANSPORT_LOCK_FAILED        = "TRN-010"

	AUTH_UNAUTHORIZED         = "AUTH-000"
	AUTH_INVALID_EMAIL        = "AUTH-001"
	AUTH_INVALID_PASSWORD     = "AUTH-002"
	AUTH_INVALID_OLD_PASSWORD = "AUTH-003"
	AUTH_JWT_FAILURE          = "AUTH-011"
	AUTH_JWT_MISSING          = "AUTH-012"
	AUTH_JWT_INVALID          = "AUTH-013"
	AUTH_JWT_EXPIRED          = "AUTH-014"
	AUTH_JWT_NOT_BEARER       = "AUTH-015"

	FS_ERROR               = "FS-000"
	FS_CANNOT_OPEN_BLOCK   = "FS-001"
	FS_CANNOT_LOAD_BLOCK   = "FS-002"
	FS_BLOCK_MISMATCH      = "FS-010"
	FS_BLOCK_UPLOAD_FAILED = "FS-020"
)
