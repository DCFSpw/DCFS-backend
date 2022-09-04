package dbo

import "github.com/google/uuid"

type Disk struct {
	AbstractDatabaseObject
	UserUUID     uuid.UUID
	VolumeUUID   uuid.UUID
	ProviderUUID uuid.UUID
	Credentials  string
}

func NewDisk() *Disk {
	var d *Disk = new(Disk)
	d.AbstractDatabaseObject.DatabaseObject = d
	return d
}
