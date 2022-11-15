package dbo

import (
	"dcfs/requests"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"time"
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

	CreatedAt time.Time      `gorm:"<-:create" json:"-"`
	DeletedAt gorm.DeletedAt `json:"-"`

	User User `gorm:"foreignKey:UserUUID;references:UUID" json:"-"`
}

// NewVolume - create new volume object
//
// return type:
//   - *dbo.Volume: created volume DBO
func NewVolume() *Volume {
	var v *Volume = new(Volume)
	v.AbstractDatabaseObject.DatabaseObject = v
	return v
}

// NewVolumeFromRequest - create volume DBO from volume create request
//
// params:
//   - request *requests.VolumeCreateRequest: volume create request data from API request
//   - userUUID uuid.UUID: UUID of the user who is creating the volume
//
// return type:
//   - *dbo.Volume: created volume DBO
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

// GetCreationTime - get creation time of the volume
//
// return type:
//   - time.Time: creation time of the volume
func (v Volume) GetCreationTime() time.Time {
	return v.CreatedAt
}
