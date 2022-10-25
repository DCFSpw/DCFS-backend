package responses

import "dcfs/db/dbo"

func NewVolumeDataSuccessResponse(volumeData *dbo.Volume) *EmptySuccessResponse {
	var r *EmptySuccessResponse = new(EmptySuccessResponse)
	r.Success = true
	r.Data = *volumeData
	return r
}
