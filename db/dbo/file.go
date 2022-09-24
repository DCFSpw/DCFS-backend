package dbo

import (
	"github.com/google/uuid"
	"time"
)

// FileType types
const (
	FILE_TYPE_REGULAR   int = 1
	FILE_TYPE_DIRECTORY int = 0
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
}

func NewFile() *File {
	var f *File = new(File)
	f.AbstractDatabaseObject.DatabaseObject = f
	return f
}
