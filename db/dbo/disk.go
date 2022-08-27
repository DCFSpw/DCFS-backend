package dbo

import "github.com/google/uuid"

type Disk struct {
	AbstractDatabaseObject
	UserUUID     uuid.UUID
	VolumeUUID   uuid.UUID
	providerUUID uuid.UUID
	credentials  string
}

func NewDisk() *Disk {
	var d *Disk = new(Disk)
	d.AbstractDatabaseObject.DatabaseObject = d
	return d
}
