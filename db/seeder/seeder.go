package seeder

import (
	"dcfs/constants"
	"dcfs/db"
	"dcfs/db/dbo"
	"dcfs/models/disk"
	"github.com/google/uuid"
)

// TODO: move seeding to another package to avoid cyclic imports
func Seed() {
	rootUUID := disk.RootUUID
	volume := dbo.Volume{}
	provider := dbo.Provider{}
	user := dbo.User{}

	// add a root user
	db.DB.DatabaseHandle.Where("uuid = ?", disk.RootUUID).First(&user)
	if user.UUID != disk.RootUUID {
		user.UUID = disk.RootUUID
		user.FirstName = "Root"
		user.LastName = "Root"
		user.Email = "root@root.com"
		user.Password = dbo.HashPassword("password")
		db.DB.DatabaseHandle.Create(&user)
	}

	// add some volumes
	db.DB.DatabaseHandle.Where("user_uuid = ?", rootUUID).First(&volume)
	if volume.UserUUID != rootUUID {
		// create a new volume
		volume.UUID = uuid.New()
		volume.UserUUID = rootUUID
		volume.VolumeSettings = dbo.VolumeSettings{Backup: constants.BACKUP_TYPE_NO_BACKUP, Encryption: constants.ENCRYPTION_TYPE_NO_ENCRYPTION, FilePartition: constants.PARTITION_TYPE_BALANCED}

		db.DB.DatabaseHandle.Create(&volume)
	}

	// add some providers
	db.DB.DatabaseHandle.Where("provider_type = ?", constants.PROVIDER_TYPE_GDRIVE).First(&provider)
	if provider.ProviderType != constants.PROVIDER_TYPE_GDRIVE {
		provider.UUID = uuid.New()
		provider.ProviderType = constants.PROVIDER_TYPE_GDRIVE

		db.DB.DatabaseHandle.Create(&provider)
	}
}
