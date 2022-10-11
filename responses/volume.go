package responses

import "dcfs/db/dbo"

type VolumeDataSuccessResponse struct {
	Success bool       `json:"success"`
	Data    dbo.Volume `json:"data"`
}

func NewVolumeDataSuccessResponse(volumeData *dbo.Volume) *VolumeDataSuccessResponse {
	var r *VolumeDataSuccessResponse = new(VolumeDataSuccessResponse)
	r.Success = true
	r.Data = *volumeData
	return r
}
