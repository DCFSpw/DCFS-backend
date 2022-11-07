package db

import (
	"dcfs/constants"
	"dcfs/db/dbo"
	"github.com/google/uuid"
)

func UserFromDatabase(uuid uuid.UUID) (*dbo.User, string) {
	var user *dbo.User = dbo.NewUser()

	result := DB.DatabaseHandle.Where("uuid = ?", uuid).First(&user)
	if result.Error != nil {
		return nil, constants.DATABASE_USER_NOT_FOUND
	}

	return user, constants.SUCCESS
}

func VolumeFromDatabase(uuid string) (*dbo.Volume, string) {
	var volume *dbo.Volume = dbo.NewVolume()

	result := DB.DatabaseHandle.Where("uuid = ?", uuid).First(&volume)
	if result.Error != nil {
		return nil, constants.DATABASE_VOLUME_NOT_FOUND
	}

	return volume, constants.SUCCESS
}

func FileFromDatabase(uuid string) (*dbo.File, string) {
	var file *dbo.File = dbo.NewFile()

	result := DB.DatabaseHandle.Where("uuid = ?", uuid).First(&file)
	if result.Error != nil {
		return nil, constants.DATABASE_FILE_NOT_FOUND
	}

	return file, constants.SUCCESS
}

func IsVolumeEmpty(uuid uuid.UUID) (bool, error) {
	var blockCount int64
	err := DB.DatabaseHandle.Model(&dbo.Block{}).Where("volume_uuid = ?", uuid).Count(&blockCount).Error
	return blockCount == 0, err
}

func ValidateRootDirectory(rootUUID uuid.UUID, volumeUUID uuid.UUID) string {
	var rootDirectory *dbo.File = dbo.NewFile()

	// Check if the root directory refers to volume's root directory
	if rootUUID == uuid.Nil {
		return constants.SUCCESS
	}

	// Check if the root directory exists
	rootDirectory, errCode := FileFromDatabase(rootUUID.String())
	if errCode != constants.SUCCESS {
		return errCode
	} else if rootDirectory == nil {
		return constants.DATABASE_FILE_NOT_FOUND
	}

	// Check if root directory belongs to the provided volume
	if rootDirectory.VolumeUUID != volumeUUID {
		return constants.FS_VOLUME_MISMATCH
	}

	// Check if root directory is a directory
	if rootDirectory.Type != constants.FILE_TYPE_DIRECTORY {
		return constants.FS_FILE_TYPE_MISMATCH
	}

	return constants.SUCCESS
}

func GenerateFileFullPath(rootUUID uuid.UUID) ([]dbo.PathEntry, string) {
	var path []dbo.PathEntry = make([]dbo.PathEntry, 0)

	// Iterate through file's path to root directory of the volume
	for rootUUID != uuid.Nil {
		var parent dbo.File

		// Retrieve parent directory from database
		result := DB.DatabaseHandle.Where("uuid = ?", rootUUID).First(&parent)
		if result.Error != nil {
			return nil, constants.DATABASE_ERROR
		}

		// Add parent directory to path
		path = append(path, dbo.PathEntry{
			UUID: parent.UUID,
			Name: parent.Name,
		})

		// Move to parent directory
		rootUUID = parent.RootUUID
	}

	return path, constants.SUCCESS
}
