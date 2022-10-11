package dbo

import "github.com/google/uuid"

type Disk struct {
	AbstractDatabaseObject
	UserUUID     uuid.UUID `json:"-"`
	VolumeUUID   uuid.UUID `json:"-"`
	ProviderUUID uuid.UUID `json:"-"`
	Credentials  string    `json:"credentials"`

	User     User     `gorm:"foreignKey:UserUUID;references:UUID"`
	Volume   Volume   `gorm:"foreignKey:VolumeUUID;references:UUID"`
	Provider Provider `gorm:"foreignKey:ProviderUUID;references:UUID"`
}

func NewDisk() *Disk {
	var d *Disk = new(Disk)
	d.AbstractDatabaseObject.DatabaseObject = d
	return d
}
