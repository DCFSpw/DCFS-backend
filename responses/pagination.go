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
	var r *PaginationResponse = new(PaginationResponse)

	r.Success = true
	r.Data.Pagination = paginationData.Pagination
	r.Data.Data = paginationData.Data
	
	return *r
}
