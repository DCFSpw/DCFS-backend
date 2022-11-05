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

	VolumeUUID uuid.UUID
	RootUUID   uuid.UUID
	UserUUID   uuid.UUID
	Type       int
	Name       string

	Size     int
	Checksum int

	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt

	Volume Volume `gorm:"foreignKey:VolumeUUID;references:UUID"`
	User   User   `gorm:"foreignKey:UserUUID;references:UUID"`
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
