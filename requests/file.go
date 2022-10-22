package requests

type FileDataRequest struct {
	Name string `json:"name" binding:"required,gte=1,lte=64"`
	Type int    `json:"type" binding:"required,min=0,max=1"`
	Size int    `json:"size" binding:"required,min=1"`
}

type GetFileRequest struct {
	VolumeUUID string          `json:"volumeUUID" binding:"required"`
	File       FileDataRequest `json:"file" binding:"required"`
}

type FileUploadRequest struct {
}

type FileRequestCompleteRequest struct {
	Direction bool `json:"direction"` // true - download, false - upload
}
