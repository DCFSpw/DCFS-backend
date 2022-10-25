package responses

type data struct {
	Pagination interface{} `json:"pagination"`
	Data       interface{} `json:"data"`
}

type PaginationResponse struct {
	Success bool `json:"success"`
	Data    data `json:"data"`
}

type PaginationData struct {
	Pagination interface{} `json:"pagination"`
	Data       interface{} `json:"data"`
}

func NewPaginationResponse(paginationData PaginationData) PaginationResponse {
	return PaginationResponse{Success: true, Data: data{Pagination: paginationData.Pagination, Data: paginationData.Data}}
}
