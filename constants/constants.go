package constants

// Provider types
const (
	PROVIDER_TYPE_SFTP     int = 1
	PROVIDER_TYPE_GDRIVE   int = 2
	PROVIDER_TYPE_ONEDRIVE int = 3
	PROVIDER_TYPE_FTP      int = 4
)

// File types
const (
	FILE_TYPE_DIRECTORY       int = 1
	FILE_TYPE_REGULAR         int = 2
	FILE_TYPE_SMALLER_WRAPPER int = 3
	FILE_TYPE_WRAPPER         int = 4
)

// Backup types
const (
	BACKUP_TYPE_RAID_1    int = 1
	BACKUP_TYPE_NO_BACKUP int = 2
)

// Encryption types
const (
	ENCRYPTION_TYPE_AES_256       int = 1
	ENCRYPTION_TYPE_NO_ENCRYPTION int = 2
)

// FilePartition types
const (
	PARTITION_TYPE_BALANCED   int = 1
	PARTITION_TYPE_PRIORITY   int = 2
	PARTITION_TYPE_THROUGHPUT int = 3
)

// Block status
const (
	BLOCK_STATUS_QUEUED      int = 0
	BLOCK_STATUS_IN_PROGRESS int = 1
	BLOCK_STATUS_TRANSFERRED int = 2
	BLOCK_STATUS_FAILED      int = 3
)

// Pagination constants
const (
	PAGINATION_RECORDS_PER_PAGE int = 12
)

// OneDrive constants
const (
	ONEDRIVE_SIZE_LIMIT   int = 4 * 1024 * 1024
	ONEDRIVE_UPLOAD_LIMIT int = 192 * 320 * 1024 // 60 MiB
)

// Sizes constants
const (
	DEFAULT_VOLUME_BLOCK_SIZE int = 8 * 1024 * 1024
	FRONT_RAM_CAPACITY        int = 8 * 1024 * 1024
)

// Deletion constants
const (
	DELETION   bool = false
	RELOCATION bool = true
)
