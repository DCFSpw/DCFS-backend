package dbo

import "github.com/google/uuid"

type Disk struct {
	AbstractDatabaseObject
	UserUUID     uuid.UUID
	VolumeUUID   uuid.UUID
	ProviderUUID uuid.UUID
	Credentials  string

	User     User     `gorm:"foreignKey:UserUUID;references:UUID"`
	Volume   Volume   `gorm:"foreignKey:VolumeUUID;references:UUID"`
	Provider Provider `gorm:"foreignKey:ProviderUUID;references:UUID"`
}

func NewDisk() *Disk {
	var d *Disk = new(Disk)
	d.AbstractDatabaseObject.DatabaseObject = d
	return d
}
