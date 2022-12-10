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

type FileResponse struct {
	dbo.File
	FileReady bool `json:"fileReady"`
}

// NewFileDataSuccessResponse - create file data success response
//
// params:
//   - fileData dbo.File: file data to return
//
// return type:
//   - *SuccessResponse: response with file data
func NewFileDataSuccessResponse(fileData *dbo.File) *SuccessResponse {
	var r *SuccessResponse = new(SuccessResponse)

	r.Success = true
	r.Data = *fileData

	return r
}

// NewFileDataWithPathSuccessResponse - create file data with DCFS filesystem path success response
//
// params:
//   - fileData dbo.File: file data to return
//   - path []dbo.PathEntry- array of elements of DCFS filesystem path of the file (from file level up to root level)
//
// return type:
//   - *SuccessResponse: response with file and path data
func NewFileDataWithPathSuccessResponse(fileData *FileResponse, filePath []dbo.PathEntry) *SuccessResponse {
	var r *SuccessResponse = new(SuccessResponse)
	var data = new(FileDetailsWithPathResponse)

	data.File = fileData.File
	data.Path = filePath

	r.Success = true
	r.Data = data

	return r
}

// NewGetFilesSuccessResponse - create get files success response
//
// params:
//   - filesData []dbo.File: array of files data from single directory to return
//
// return type:
//   - *SuccessResponse: response with files data
func NewGetFilesSuccessResponse(filesData []FileResponse) *SuccessResponse {
	var r *SuccessResponse = new(SuccessResponse)

	r.Success = true
	r.Data = filesData

	return r
}

// NewInitFileUploadRequestResponse - create init file upload success response
//
// params:
//   - userUUID uuid.UUID: UUID of owner of the file
//   - filesData models.File: file and block data to return
//
// return type:
//   - *SuccessResponse: response with file data
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
//   - blocks []FileRequestBlockResponse: array of data of blocks that failed to transfer
//
// return type:
//   - *SuccessResponse: response with list of blocks that failed to transfer
func NewBlockTransferFailureResponse(blocks []FileRequestBlockResponse) *SuccessResponse {
	var r *SuccessResponse = new(SuccessResponse)

	r.Success = false
	r.Data = blocks

	return r
}
