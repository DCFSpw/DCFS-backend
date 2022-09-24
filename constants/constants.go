package constants

// provider types
const (
	PROVIDER_TYPE_SFTP     int = 0
	PROVIDER_TYPE_GDRIVE   int = 1
	PROVIDER_TYPE_ONEDRIVE int = 2
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
