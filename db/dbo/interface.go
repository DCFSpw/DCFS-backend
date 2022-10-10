package dbo

import (
	"github.com/google/uuid"
)

type DatabaseObject interface {
}

type AbstractDatabaseObject struct {
	DatabaseObject `gorm:"-" json:"-"`
	UUID           uuid.UUID `gorm:"primaryKey;type:varchar(36)" json:"uuid"`
}
