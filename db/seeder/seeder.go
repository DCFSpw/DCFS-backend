package db

// TODO: move seeding to another package to avoid cyclic imports
/*func Seed() {
	rootUUID := disk.RootUUID
	volume := dbo.Volume{}
	provider := dbo.Provider{}

	// add some volumes
	DB.DatabaseHandle.Where("user_uuid = ?", rootUUID).First(&volume)
	if volume.UserUUID != rootUUID {
		// create a new volume
		volume.UUID = uuid.New()
		volume.UserUUID = rootUUID
		volume.VolumeSettings = dbo.VolumeSettings{Backup: dbo.NO_BACKUP, Encryption: dbo.NO_ENCRYPTION, FilePartition: dbo.BALANCED}

		DB.DatabaseHandle.Create(&volume)
	}

	// add some providers
	DB.DatabaseHandle.Where("provider_type = ?", constants.PROVIDER_TYPE_GDRIVE).First(&provider)
	if provider.ProviderType != constants.PROVIDER_TYPE_GDRIVE {
		provider.UUID = uuid.New()
		provider.ProviderType = constants.PROVIDER_TYPE_GDRIVE

		DB.DatabaseHandle.Create(&provider)
	}
}*/
