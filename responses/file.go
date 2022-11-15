package responses

import (
	"dcfs/db/dbo"
	"dcfs/models"
	"github.com/google/uuid"
)

type FileDetailsWithPathResponse struct {
	File dbo.File        `json:"file"`
	Path []dbo.PathEntry `json:"path"`
}

type FileRequestBlockResponse struct {
	UUID  uuid.UUID `json:"uuid"`
	Order int       `json:"order"`
	Size  int       `json:"size"`
}

type FileRequestResponse struct {
	File   dbo.File                   `json:"file"`
	Blocks []FileRequestBlockResponse `json:"blocks"`
}

// NewFileDataSuccessResponse - create file data success response
//
// params:
//   - fileData: dbo.File pointer with file data to return
//
// return type:
//   - response: SuccessResponse with file data
func NewFileDataSuccessResponse(fileData *dbo.File) *SuccessResponse {
	var r *SuccessResponse = new(SuccessResponse)

	r.Success = true
	r.Data = *fileData

	return r
}

// NewFileDataWithPathSuccessResponse - create file data with DCFS filesystem path success response
//
// params:
//   - fileData: dbo.File pointer with file data to return
//   - path - array of dbo.PathEntry objects containing elements of DCFS filesystem path of the file (from file level up to root level)
//
// return type:
//   - response: SuccessResponse with file and path data
func NewFileDataWithPathSuccessResponse(fileData *dbo.File, filePath []dbo.PathEntry) *SuccessResponse {
	var r *SuccessResponse = new(SuccessResponse)
	var data = new(FileDetailsWithPathResponse)

	data.File = *fileData
	data.Path = filePath

	r.Success = true
	r.Data = data

	return r
}

// NewGetFilesSuccessResponse - create get files success response
//
// params:
//   - filesData: array of dbo.File with files data from one directory to return
//
// return type:
//   - response: SuccessResponse with files data
func NewGetFilesSuccessResponse(filesData []dbo.File) *SuccessResponse {
	var r *SuccessResponse = new(SuccessResponse)

	r.Success = true
	r.Data = filesData

	return r
}

// NewInitFileUploadRequestResponse - create init file upload success response
//
// params:
//   - userUUID - uuid.UUID with UUID of owner of the file
//   - filesData: models.File object with file and block data to return
//
// return type:
//   - response: SuccessResponse with file data
func NewInitFileUploadRequestResponse(userUUID uuid.UUID, file models.File) *SuccessResponse {
	var r *SuccessResponse = new(SuccessResponse)
	var fr *FileRequestResponse = new(FileRequestResponse)

	// Prepare file for response
	fr.File = file.GetFileDBO(userUUID)

	// Prepare blocks for response
	var blocks []FileRequestBlockResponse
	for _, block := range file.GetBlocks() {
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

// NewBlockTransferFailureResponse - create block transfer failure response
//
// params:
//   - blocks - array of FileRequestBlockResponse with data of blocks that failed to transfer
//
// return type:
//   - response: SuccessResponse with list of blocks that failed to transfer
func NewBlockTransferFailureResponse(blocks []FileRequestBlockResponse) *SuccessResponse {
	var r *SuccessResponse = new(SuccessResponse)

	r.Success = false
	r.Data = blocks

	return r
}
