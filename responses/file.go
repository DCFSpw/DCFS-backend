package responses

import (
	"dcfs/db/dbo"
	"dcfs/models"
	"github.com/google/uuid"
)

type FileRequestBlockResponse struct {
	UUID  uuid.UUID `json:"UUID"`
	Order int       `json:"order"`
	Size  int       `json:"size"`
}

type FileRequestResponse struct {
	File   dbo.File                   `json:"file"`
	Blocks []FileRequestBlockResponse `json:"blocks"`
}

func NewFileDataSuccessResponse(fileData *dbo.File) *EmptySuccessResponse {
	var r *EmptySuccessResponse = new(EmptySuccessResponse)

	r.Success = true
	r.Data = *fileData

	return r
}

func NewGetFilesSuccessResponse(files []dbo.File) *EmptySuccessResponse {
	var r *EmptySuccessResponse = new(EmptySuccessResponse)

	r.Success = true
	r.Data = files

	return r
}

func NewInitFileUploadRequestResponse(userUUID uuid.UUID, file *models.RegularFile) *EmptySuccessResponse {
	var r *EmptySuccessResponse = new(EmptySuccessResponse)
	var fr *FileRequestResponse = new(FileRequestResponse)

	// Prepare file for response
	fr.File = file.GetFileDBO(userUUID)

	// Prepare blocks for response
	var blocks []FileRequestBlockResponse
	for _, block := range file.Blocks {
		blocks = append(blocks, FileRequestBlockResponse{
			UUID:  block.UUID,
			Order: block.Order,
			Size:  block.Size,
		})
	}
	fr.Blocks = blocks

	// Prepare final response
	r.Success = true
	r.Data = fr

	return r
}
