package dbo

import (
	"dcfs/constants"
	"dcfs/requests"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"time"
)

type File struct {
	AbstractDatabaseObject

	VolumeUUID uuid.UUID `json:"-"`
	RootUUID   uuid.UUID `json:"-"`
	UserUUID   uuid.UUID `json:"-"`
	Type       int       `json:"type"`
	Name       string    `json:"name"`

	Size     int    `json:"size"`
	Checksum string `json:"checksum"`

	CreatedAt time.Time      `json:"creationDate"`
	UpdatedAt time.Time      `json:"modificationDate"`
	DeletedAt gorm.DeletedAt `json:"-"`

	Volume Volume `gorm:"foreignKey:VolumeUUID;references:UUID" json:"-"`
	User   User   `gorm:"foreignKey:UserUUID;references:UUID" json:"-"`
}

func NewFile() *File {
	var f *File = new(File)
	f.AbstractDatabaseObject.DatabaseObject = f
	return f
}

func NewDirectoryFromRequest(request *requests.DirectoryCreateRequest, userUUID uuid.UUID, rootUUID uuid.UUID) *File {
	var d *File = NewFile()

	d.AbstractDatabaseObject.DatabaseObject = d
	d.UUID, _ = uuid.NewUUID()

	d.VolumeUUID = uuid.MustParse(request.VolumeUUID)
	d.RootUUID = rootUUID
	d.UserUUID = userUUID

	d.Type = constants.FILE_TYPE_DIRECTORY
	d.Name = request.Name

	return d
}
