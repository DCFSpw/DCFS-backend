package mock

import (
	"dcfs/constants"
	"dcfs/db/dbo"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"time"
)

func GetBlockDBOs(number int, diskUUID uuid.UUID, fileUUID uuid.UUID) []dbo.Block {
	blocks := make([]dbo.Block, 0)

	for i := 0; i < number; i++ {
		blocks = append(blocks, dbo.Block{
			AbstractDatabaseObject: dbo.AbstractDatabaseObject{
				UUID: uuid.New(),
			},
			UserUUID:   UserUUID,
			VolumeUUID: VolumeUUID,
			DiskUUID:   diskUUID,
			FileUUID:   fileUUID,
			Size:       constants.DEFAULT_VOLUME_BLOCK_SIZE,
			Order:      i,
			Checksum:   "",
			Volume:     dbo.Volume{},
			Disk:       dbo.Disk{},
			File:       dbo.File{},
		})
	}

	return blocks
}

func GetFileDBO(diskUUID uuid.UUID, fileType int, fileSize int) dbo.File {
	return dbo.File{
		AbstractDatabaseObject: dbo.AbstractDatabaseObject{
			UUID: uuid.New(),
		},
		VolumeUUID: VolumeUUID,
		RootUUID:   uuid.Nil,
		UserUUID:   UserUUID,
		Type:       fileType,
		Name:       "dummy-file",
		Size:       fileSize,
		Checksum:   "",
		CreatedAt:  time.Time{},
		UpdatedAt:  time.Time{},
		DeletedAt:  gorm.DeletedAt{},
		Volume:     *VolumeDBO,
		User:       *UserDBO,
	}
}
