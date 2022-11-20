package mock

import (
	"dcfs/constants"
	"dcfs/db/dbo"
	"dcfs/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"time"
)

var VolumeUUID uuid.UUID = uuid.New()

var VolumeSettings dbo.VolumeSettings = dbo.VolumeSettings{
	Backup:        constants.BACKUP_TYPE_NO_BACKUP,
	Encryption:    constants.ENCRYPTION_TYPE_NO_ENCRYPTION,
	FilePartition: constants.PARTITION_TYPE_BALANCED,
}

var VolumeDBO *dbo.Volume = &dbo.Volume{
	AbstractDatabaseObject: dbo.AbstractDatabaseObject{
		UUID: VolumeUUID,
	},
	Name:           "MockVolume",
	UserUUID:       UserUUID,
	VolumeSettings: VolumeSettings,
	CreatedAt:      time.Time{},
	DeletedAt:      gorm.DeletedAt{},
	User:           *UserDBO,
}

var Volume *models.Volume = &models.Volume{
	UUID:           VolumeUUID,
	BlockSize:      constants.DEFAULT_VOLUME_BLOCK_SIZE,
	Name:           "Mock Volume",
	UserUUID:       UserUUID,
	VolumeSettings: VolumeSettings,
}
