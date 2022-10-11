package dbo

import (
	"github.com/google/uuid"
	"time"
)

type File struct {
	AbstractDatabaseObject
	UserUUID         uuid.UUID
	Type             int
	Name             string
	CreationDate     time.Time
	ModificationDate time.Time
	size             int
	checksum         int

	User User `gorm:"foreignKey:UserUUID;references:UUID"`
}

func NewFile() *File {
	var f *File = new(File)
	f.AbstractDatabaseObject.DatabaseObject = f
	return f
}
