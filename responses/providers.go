package responses

import (
	"dcfs/db/dbo"
)

type GetProvidersSuccessResponse struct {
	Success bool           `json:"success"`
	Data    []dbo.Provider `json:"data"`
}

func NewGetProvidersSuccessResponse(providers []dbo.Provider) *GetProvidersSuccessResponse {
	var r *GetProvidersSuccessResponse = new(GetProvidersSuccessResponse)

	r.Success = true
	r.Data = providers

	return r
}
