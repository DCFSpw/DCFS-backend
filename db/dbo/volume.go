package dbo

import (
	"github.com/google/uuid"
)

type VolumeSettings struct {
	Backup, Encryption, FilePartition int
}

type Volume struct {
	AbstractDatabaseObject
	UserUUID       uuid.UUID
	VolumeSettings VolumeSettings `gorm:"embedded"`
}

func NewVolume() *Volume {
	var v *Volume = new(Volume)
	v.AbstractDatabaseObject.DatabaseObject = v
	return v
}
