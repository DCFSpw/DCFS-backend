package responses

import "dcfs/db/dbo"

func NewGetFilesSuccessResponse(files []dbo.File) *EmptySuccessResponse {
	var r *EmptySuccessResponse = new(EmptySuccessResponse)

	r.Success = true
	r.Data = files

	return r
}
