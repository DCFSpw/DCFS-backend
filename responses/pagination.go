package responses

type data struct {
	Pagination interface{}
	Data       interface{}
}

type PaginationResponse struct {
	Success bool
	Data    data
}

type PaginationData struct {
	Pagination interface{}
	Data       interface{}
}

func NewPaginationResponse(paginationData PaginationData) PaginationResponse {
	return PaginationResponse{Success: true, Data: data{Pagination: paginationData.Pagination, Data: paginationData.Data}}
}
