package responses

import (
	"dcfs/db/dbo"
)

// NewGetProvidersSuccessResponse - create get providers success response
//
// params:
//   - paginationData PaginationData: pagination information and data for target page
//
// return type:
//   - *SuccessResponse: response with pagination data and target page data
func NewGetProvidersSuccessResponse(providers []dbo.Provider) *SuccessResponse {
	var r *SuccessResponse = new(SuccessResponse)

	r.Success = true
	r.Data = providers

	return r
}
