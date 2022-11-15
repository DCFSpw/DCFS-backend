package responses

import "dcfs/db/dbo"

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
