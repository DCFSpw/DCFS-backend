package dbo

import (
	"github.com/google/uuid"
)

// Backup types
const (
	NO_BACKUP int = 0
	RAID_1
)

// Encryption types
const (
	NO_ENCRYPTION int = 0
)

// FilePartition types
const (
	BALANCED int = 0
	PRIORITY
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
