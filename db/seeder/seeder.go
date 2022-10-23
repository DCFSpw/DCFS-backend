package seeder

import (
	"dcfs/constants"
	"dcfs/db"
	"dcfs/db/dbo"
	"dcfs/models"
	"github.com/google/uuid"
)

// TODO: move seeding to another package to avoid cyclic imports
func Seed() {
	rootUUID := models.RootUUID
	volume := dbo.Volume{}
	provider1 := dbo.Provider{}
	provider2 := dbo.Provider{}
	provider3 := dbo.Provider{}
	provider4 := dbo.Provider{}
	user := dbo.User{}

	// Add a root user
	db.DB.DatabaseHandle.Where("uuid = ?", models.RootUUID).First(&user)
	if user.UUID != models.RootUUID {
		user.UUID = models.RootUUID
		user.FirstName = "Root"
		user.LastName = "Root"
		user.Email = "root@root.com"
		user.Password = dbo.HashPassword("password")
		db.DB.DatabaseHandle.Create(&user)
	}

	// Add some volumes
	db.DB.DatabaseHandle.Where("user_uuid = ?", rootUUID).First(&volume)
	if volume.UserUUID != rootUUID {
		// Create a new volume
		volume.UUID = uuid.New()
		volume.UserUUID = rootUUID
		volume.VolumeSettings = dbo.VolumeSettings{Backup: constants.BACKUP_TYPE_NO_BACKUP, Encryption: constants.ENCRYPTION_TYPE_NO_ENCRYPTION, FilePartition: constants.PARTITION_TYPE_BALANCED}

		db.DB.DatabaseHandle.Create(&volume)
	}

	// Add providers
	db.DB.DatabaseHandle.Where("type = ?", constants.PROVIDER_TYPE_SFTP).First(&provider1)
	if provider1.Type != constants.PROVIDER_TYPE_SFTP {
		provider1.UUID = uuid.New()
		provider1.Type = constants.PROVIDER_TYPE_SFTP
		provider1.Name = "SFTP drive"
		provider1.Logo = "https://freesvg.org/img/1538300664.png"

		db.DB.DatabaseHandle.Create(&provider1)
	}

	db.DB.DatabaseHandle.Where("type = ?", constants.PROVIDER_TYPE_GDRIVE).First(&provider2)
	if provider2.Type != constants.PROVIDER_TYPE_GDRIVE {
		provider2.UUID = uuid.New()
		provider2.Type = constants.PROVIDER_TYPE_GDRIVE
		provider2.Name = "GoogleDrive"
		provider2.Logo = "https://upload.wikimedia.org/wikipedia/commons/d/da/Google_Drive_logo.png"

		db.DB.DatabaseHandle.Create(&provider2)
	}

	db.DB.DatabaseHandle.Where("type = ?", constants.PROVIDER_TYPE_ONEDRIVE).First(&provider3)
	if provider3.Type != constants.PROVIDER_TYPE_ONEDRIVE {
		provider3.UUID = uuid.New()
		provider3.Type = constants.PROVIDER_TYPE_ONEDRIVE
		provider3.Name = "OneDrive"
		provider3.Logo = "https://upload.wikimedia.org/wikipedia/commons/3/3c/Microsoft_Office_OneDrive_%282019%E2%80%93present%29.svg"

		db.DB.DatabaseHandle.Create(&provider3)
	}

	db.DB.DatabaseHandle.Where("type = ?", constants.PROVIDER_TYPE_FTP).First(&provider4)
	if provider4.Type != constants.PROVIDER_TYPE_FTP {
		provider4.UUID = uuid.New()
		provider4.Type = constants.PROVIDER_TYPE_FTP
		provider4.Name = "FTP drive"
		provider4.Logo = "https://upload.wikimedia.org/wikipedia/commons/thumb/8/80/Antu_gFTP.svg/640px-Antu_gFTP.svg.png"

		db.DB.DatabaseHandle.Create(&provider4)
	}
}
