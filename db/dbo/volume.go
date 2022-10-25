package dbo

import (
	"github.com/google/uuid"
)

type VolumeSettings struct {
	Backup        int `json:"backup"`
	Encryption    int `json:"encryption"`
	FilePartition int `json:"filePartition"`
}

type Volume struct {
	AbstractDatabaseObject
	UserUUID       uuid.UUID      `json:"-"`
	VolumeSettings VolumeSettings `gorm:"embedded" json:"settings"`
}

func NewVolume() *Volume {
	var v *Volume = new(Volume)
	v.AbstractDatabaseObject.DatabaseObject = v
	return v
}
