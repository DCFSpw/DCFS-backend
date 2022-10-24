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

func IsVolumeEmpty(uuid uuid.UUID) (bool, error) {
	var blockCount int64
	err := DB.DatabaseHandle.Model(&dbo.Block{}).Where("VolumeUUID = ?", uuid).Count(&blockCount).Error
	return blockCount == 0, err
}
