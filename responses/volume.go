package responses

import "dcfs/db/dbo"

type VolumeResponse struct {
	dbo.Volume
	IsReady bool `json:"isReady"`
}

// NewVolumeDataSuccessResponse - create volume data success response
//
// params:
//   - volumeData *dbo.Volume: volume data to return
//
// return type:
//   - *SuccessResponse: response with volume data
func NewVolumeDataSuccessResponse(volumeData *dbo.Volume) *SuccessResponse {
	var r *SuccessResponse = new(SuccessResponse)

	r.Success = true
	r.Data = *volumeData

	return r
}

// NewVolumeListSuccessResponse - create volume data success response for get requests
//
// params:
//   - volumeData *VolumeResponse: volume data to return
//
// return type:
//   - *SuccessResponse: response with volume data
func NewVolumeListSuccessResponse(volumeData *VolumeResponse) *SuccessResponse {
	var r *SuccessResponse = new(SuccessResponse)

	r.Success = true
	r.Data = *volumeData

	return r
}
