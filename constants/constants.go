package constants

// provider types
const (
	PROVIDER_TYPE_SFTP     int = 1
	PROVIDER_TYPE_GDRIVE   int = 2
	PROVIDER_TYPE_ONEDRIVE int = 3
	PROVIDER_TYPE_FTP      int = 4
)

// file types
const (
	FILE_TYPE_REGULAR   int = 1
	FILE_TYPE_DIRECTORY int = 0
)

// Backup types
const (
	BACKUP_TYPE_NO_BACKUP int = 0
	BACKUP_TYPE_RAID_1    int = 1
)

// Encryption types
const (
	ENCRYPTION_TYPE_NO_ENCRYPTION int = 0
	ENCRYPTION_TYPE_AES_256       int = 1
)

// FilePartition types
const (
	PARTITION_TYPE_BALANCED int = 0
	PARTITION_TYPE_PRIORITY int = 1
)

// Block status
const (
	BLOCK_STATUS_QUEUED      int = 0
	BLOCK_STATUS_IN_PROGRESS int = 1
	BLOCK_STATUS_TRANSFERRED int = 2
	BLOCK_STATUS_FAILED      int = 3
)

// Pagination
const (
	PAGINATION_RECORDS_PER_PAGE int = 10
)

// OneDrive
const (
	ONEDRIVE_SIZE_LIMIT   int = 4 * 1024 * 1024
	ONEDRIVE_UPLOAD_LIMIT int = 192 * 320 * 1024 // 60 MiB
)
