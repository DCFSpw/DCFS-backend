package apicalls

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type BlockMetadata struct {
	Ctx      *gin.Context
	FileUUID uuid.UUID
	UUID     uuid.UUID
	Size     int64
	Status   *int

	Content *[]uint8

	CompleteCallback func(uuid.UUID, *int)
}
