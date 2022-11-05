package dbo

import (
	"github.com/google/uuid"
	"time"
)

type File struct {
	AbstractDatabaseObject

	RootUUID uuid.UUID
	UserUUID uuid.UUID
	Type     int
	Name     string

	Size     int
	Checksum int

	CreationDate     time.Time
	ModificationDate time.Time

	User User `gorm:"foreignKey:UserUUID;references:UUID"`
}

func NewFile() *File {
	var f *File = new(File)
	f.AbstractDatabaseObject.DatabaseObject = f
	return f
}
