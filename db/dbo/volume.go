package dbo

import (
	"dcfs/requests"
	"github.com/google/uuid"
)

type VolumeSettings struct {
	Backup        int `json:"backup"`
	Encryption    int `json:"encryption"`
	FilePartition int `json:"filePartition"`
}

type Volume struct {
	AbstractDatabaseObject
	Name           string         `json:"name"`
	UserUUID       uuid.UUID      `json:"-"`
	VolumeSettings VolumeSettings `gorm:"embedded" json:"settings"`

	User User `gorm:"foreignKey:UserUUID;references:UUID" json:"-"`
}

func NewVolume() *Volume {
	var v *Volume = new(Volume)
	v.AbstractDatabaseObject.DatabaseObject = v
	return v
}

func NewVolumeFromRequest(request *requests.VolumeCreateRequest, userUUID uuid.UUID) *Volume {
	var v *Volume = NewVolume()

	v.AbstractDatabaseObject.DatabaseObject = v
	v.UUID, _ = uuid.NewUUID()
	v.UserUUID = userUUID
	v.Name = request.Name
	v.VolumeSettings.Backup = request.Settings.Backup
	v.VolumeSettings.Encryption = request.Settings.Encryption
	v.VolumeSettings.FilePartition = request.Settings.FilePartition

	return v
}
