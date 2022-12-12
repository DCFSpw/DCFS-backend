package constants

// List of completion codes used internally and returned by the server
// in the responses body. If requested operation is performed successfully
// constants.SUCCESS code will be returned, otherwise appropriate error
// code will be returned.
const (
	// Success response
	SUCCESS = "ACC-000"

	// Validation errors
	VAL_VALIDATOR_ERROR        = "VAL-000"
	VAL_EMAIL_ALREADY_EXISTS   = "VAL-001"
	VAL_UUID_INVALID           = "VAL-010"
	VAL_PROVIDER_NOT_SUPPORTED = "VAL-020"
	VAL_SIZE_INVALID           = "VAL-030"
	VAL_QUOTA_EXCEEDED         = "VAL-031"
	VAL_CREDENTIALS_INVALID    = "VAL=040"

	// Database errors
	DATABASE_ERROR            = "DB-001"
	DATABASE_USER_NOT_FOUND   = "DB-002"
	DATABASE_DISK_NOT_FOUND   = "DB-003"
	DATABASE_VOLUME_NOT_FOUND = "DB-004"
	DATABASE_FILE_NOT_FOUND   = "DB-005"

	// Encryption errors
	ENCRYPTION_JOB_FAILED = "ENC-001"

	// Transport errors
	TRANSPORT_VOLUME_NOT_FOUND     = "TRN-001"
	TRANSPORT_DISK_NOT_FOUND       = "TRN-002"
	TRANSPORT_DISK_IS_BEING_USED   = "TRN-003"
	TRANSPORT_VOLUME_IS_BEING_USED = "TRN-004"
	TRANSPORT_VOLUME_NOT_READY     = "TRN-005"
	TRANSPORT_LOCK_FAILED          = "TRN-010"

	// Remote filesystem errors
	REMOTE_CANNOT_AUTHENTICATE = "RMT-000"
	REMOTE_CLIENT_UNAVAILABLE  = "RMT-001"
	REMOTE_BAD_REQUEST         = "RMT-002"
	REMOTE_BAD_FILE            = "RMT-010"
	REMOTE_CORRUPTED_FILES     = "RMT-011"
	REMOTE_FAILED_JOB          = "RMT-020"
	REMOTE_CORRUPTED_BLOCKS    = "RMT-021"
	REMOTE_CANNOT_GET_STATS    = "RMT-030"

	// Authorization errors
	AUTH_UNAUTHORIZED         = "AUTH-000"
	AUTH_INVALID_EMAIL        = "AUTH-001"
	AUTH_INVALID_PASSWORD     = "AUTH-002"
	AUTH_INVALID_OLD_PASSWORD = "AUTH-003"
	AUTH_JWT_FAILURE          = "AUTH-011"
	AUTH_JWT_MISSING          = "AUTH-012"
	AUTH_JWT_INVALID          = "AUTH-013"
	AUTH_JWT_EXPIRED          = "AUTH-014"
	AUTH_JWT_NOT_BEARER       = "AUTH-015"

	// OAuth errors
	OAUTH_BAD_CODE = "AUTH-001"

	// DCFS filesystem errors
	FS_ERROR               = "FS-000"
	FS_CANNOT_OPEN_BLOCK   = "FS-001"
	FS_CANNOT_LOAD_BLOCK   = "FS-002"
	FS_BLOCK_MISMATCH      = "FS-010"
	FS_DISK_MISMATCH       = "FS-011"
	FS_VOLUME_MISMATCH     = "FS-012"
	FS_FILE_TYPE_MISMATCH  = "FS-013"
	FS_BLOCK_UPLOAD_FAILED = "FS-020"
	FS_BAD_FILE            = "FS-030"
	FS_DIRECTORY_NOT_EMPTY = "FS-040"

	// Ownership errors
	OWNER_MISMATCH = "OWN-001"

	// Pagination errors
	INT_PAGINATION_ERROR = "INT-001"

	// Local filesystem errors
	REAL_FS_CREATE_DIR_ERROR  = "RFS-000"
	REAL_FS_CLOSE_DIR_ERROR   = "RFS-001"
	REAL_FS_CREATE_FILE_ERROR = "RFS-010"
	REAL_FS_CLOSE_FILE_ERROR  = "RFS-011"

	// Operation errors
	OPERATION_NOT_SUPPORTED = "OP-000"
	OPERATION_FAILED        = "OP-001"
)
