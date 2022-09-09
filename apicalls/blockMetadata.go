package apicalls

import "github.com/google/uuid"

type BlockMetadata struct {
	UUID uuid.UUID
	Size int64

	Content *[]uint8
}
