package responses

import "dcfs/db/dbo"

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
