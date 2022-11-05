package dbo

import (
	"github.com/google/uuid"
)

type Block struct {
	AbstractDatabaseObject
	UserUUID   uuid.UUID
	VolumeUUID uuid.UUID
	DiskUUID   uuid.UUID
	
	Size     int
	Checksum string

	//User   User   `gorm:"foreignKey:UserUUID;references:UUID"`
	Volume Volume `gorm:"foreignKey:VolumeUUID;references:UUID"`
	Disk   Disk   `gorm:"foreignKey:DiskUUID;references:UUID"`
}

func NewBlock() *Block {
	var f *Block = new(Block)
	f.AbstractDatabaseObject.DatabaseObject = f
	return f
}
