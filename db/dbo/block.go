package dbo

import (
	"github.com/google/uuid"
)

type Block struct {
	AbstractDatabaseObject
	UserUUID   uuid.UUID `json:"-"`
	VolumeUUID uuid.UUID `json:"-"`
	DiskUUID   uuid.UUID `json:"-"`
	FileUUID   uuid.UUID `json:"-"`

	Size     int    `json:"size"`
	Order    int    `json:"order"`
	Checksum string `json:"-"`

	//User   User   `gorm:"foreignKey:UserUUID;references:UUID"`
	Volume Volume `gorm:"foreignKey:VolumeUUID;references:UUID" json:"-"`
	Disk   Disk   `gorm:"foreignKey:DiskUUID;references:UUID" json:"-"`
	File   File   `gorm:"foreignKey:FileUUID;references:UUID" json:"-"`
}

// NewBlock - create new block object
//
// return type:
//   - *dbo.Block: created block DBO
func NewBlock() *Block {
	var f *Block = new(Block)
	f.AbstractDatabaseObject.DatabaseObject = f
	return f
}
