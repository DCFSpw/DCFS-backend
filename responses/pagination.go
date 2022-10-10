package responses

import "dcfs/models"

type data struct {
	Pagination models.Pagination
	Data       interface{}
}

type PaginationResponse struct {
	Success bool
	Data    data
}

func NewPaginationResponse(paginationData models.PaginationData) PaginationResponse {
	return PaginationResponse{Success: true, Data: data{Pagination: paginationData.Pagination, Data: paginationData.Data}}
}
