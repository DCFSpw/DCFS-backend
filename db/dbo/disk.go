package dbo

import (
	"github.com/google/uuid"
	"time"
)

type Disk struct {
	AbstractDatabaseObject
	UserUUID     uuid.UUID `json:"-"`
	VolumeUUID   uuid.UUID `json:"-"`
	ProviderUUID uuid.UUID `json:"-"`
	Credentials  string    `json:"credentials"`
	Name         string    `json:"name"`

	UsedSpace  uint64 `json:"-"`
	TotalSpace uint64 `json:"totalSpace"`
	FreeSpace  uint64 `gorm:"-" json:"freeSpace"`

	CreatedAt time.Time `gorm:"<-:create" json:"-"`

	IsVirtual       bool      `json:"-"`
	VirtualDiskUUID uuid.UUID `json:"-"`

	User     User     `gorm:"foreignKey:UserUUID;references:UUID" json:"-"`
	Volume   Volume   `gorm:"foreignKey:VolumeUUID;references:UUID" json:"volume"`
	Provider Provider `gorm:"foreignKey:ProviderUUID;references:UUID" json:"provider"`
}

// NewDisk - create new disk object
//
// return type:
//   - *dbo.Disk: created disk DBO
func NewDisk() *Disk {
	var d *Disk = new(Disk)
	d.AbstractDatabaseObject.DatabaseObject = d
	d.IsVirtual = false
	d.VirtualDiskUUID = uuid.Nil
	return d
}

// NewVirtualDisk - create new virtual disk object
func NewVirtualDisk() *Disk {
	var d *Disk = new(Disk)
	d.AbstractDatabaseObject.DatabaseObject = d
	d.IsVirtual = true
	d.VirtualDiskUUID = uuid.Nil
	return d
}

// GetCreationTime - get creation time of the disk
//
// return type:
//   - time.Time: creation time of the disk
func (d Disk) GetCreationTime() time.Time {
	return d.CreatedAt
}
