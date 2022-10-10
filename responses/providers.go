package responses

import (
	"dcfs/db/dbo"
	"github.com/google/uuid"
)

type ProviderDataResponse struct {
	UUID uuid.UUID `json:"uuid"`
	Type int       `json:"type"`
	Name string    `json:"name"`
	Logo string    `json:"logo"`
}

type GetProvidersSuccessResponse struct {
	Success bool                   `json:"success"`
	Data    []ProviderDataResponse `json:"data"`
}

func NewGetProvidersSuccessResponse(providers []dbo.Provider) *GetProvidersSuccessResponse {
	var r *GetProvidersSuccessResponse = new(GetProvidersSuccessResponse)
	r.Data = make([]ProviderDataResponse, len(providers))

	// Prepare response data
	r.Success = true
	for i, provider := range providers {
		r.Data[i].UUID = provider.UUID
		r.Data[i].Type = provider.Type
		r.Data[i].Name = provider.Name
		r.Data[i].Logo = provider.Logo
	}

	return r
}
