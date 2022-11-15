package responses

type PaginationResponseData struct {
	Pagination interface{} `json:"pagination"`
	Data       interface{} `json:"data"`
}

type PaginationData struct {
	Pagination interface{} `json:"pagination"`
	Data       interface{} `json:"data"`
}

// NewPaginationResponse - create init file upload success response
//
// params:
//   - paginationData PaginationData: pagination and data for target page
//
// return type:
//   - *SuccessResponse: response with pagination data and target page data
func NewPaginationResponse(paginationData PaginationData) *SuccessResponse {
	var r *SuccessResponse = new(SuccessResponse)
	var _data *PaginationResponseData = new(PaginationResponseData)

	_data.Pagination = paginationData.Pagination
	_data.Data = paginationData.Data

	r.Success = true
	r.Data = _data

	return r
}
