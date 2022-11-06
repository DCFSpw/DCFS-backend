package requests

type FileDataRequest struct {
	Name string `json:"name" binding:"required,gte=1,lte=64"`
	Type int    `json:"type" binding:"required,min=2,max=2"` // 1 is not allowed (directory)
	Size int    `json:"size" binding:"required,min=1"`
}

type DirectoryCreateRequest struct {
	Name       string `json:"name" binding:"required,gte=1,lte=64"`
	VolumeUUID string `json:"volumeUUID" binding:"required"`
	RootUUID   string `json:"rootUUID"`
}

type InitFileUploadRequest struct {
	VolumeUUID string          `json:"volumeUUID" binding:"required"`
	RootUUID   string          `json:"rootUUID"`
	File       FileDataRequest `json:"file" binding:"required"`
}

type UpdateFileRequest struct {
	Name     string `json:"name" binding:"required,gte=1,lte=64"`
	RootUUID string `json:"rootUUID"`
}
