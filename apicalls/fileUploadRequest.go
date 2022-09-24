package apicalls

import "github.com/google/uuid"

type FileUploadRequest struct {
	Name string
	Size int
	Type int

	UserUUID uuid.UUID
}
