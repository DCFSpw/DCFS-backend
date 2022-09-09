package dbo

import (
	"github.com/google/uuid"
)

type DatabaseObject interface {
}

type AbstractDatabaseObject struct {
	DatabaseObject `gorm:"-"`
	UUID           uuid.UUID `gorm:"primaryKey"`
}
