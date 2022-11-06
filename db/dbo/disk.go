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

	CreatedAt time.Time `gorm:"<-:create" json:"-"`

	User     User     `gorm:"foreignKey:UserUUID;references:UUID" json:"user"`
	Volume   Volume   `gorm:"foreignKey:VolumeUUID;references:UUID" json:"volume"`
	Provider Provider `gorm:"foreignKey:ProviderUUID;references:UUID" json:"provider"`
}

func NewDisk() *Disk {
	var d *Disk = new(Disk)
	d.AbstractDatabaseObject.DatabaseObject = d
	return d
}

func (d Disk) GetCreationTime() time.Time {
	return d.CreatedAt
}
