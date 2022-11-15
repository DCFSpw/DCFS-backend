package responses

import "dcfs/db/dbo"

// NewVolumeDataSuccessResponse - create volume data success response
//
// params:
//   - volumeData: dbo.Volume pointer with volume data to return
//
// return type:
//   - response: SuccessResponse with volume data
func NewVolumeDataSuccessResponse(volumeData *dbo.Volume) *SuccessResponse {
	var r *SuccessResponse = new(SuccessResponse)

	r.Success = true
	r.Data = *volumeData

	return r
}
