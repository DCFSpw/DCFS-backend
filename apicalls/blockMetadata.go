package apicalls

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type BlockMetadata interface {
}

type AbstractBlockMetadata struct {
	UUID uuid.UUID
	Size int64

	Content *[]uint8
}

type SFTPBlockMetadata struct {
	AbstractBlockMetadata
}

type GDriveBlockMetadata struct {
	AbstractBlockMetadata
	Ctx *gin.Context
}
