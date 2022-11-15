package responses

type Response interface {
}

type SuccessResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data"`
}

// NewEmptySuccessResponse - create success response with no data
//
// return type:
//   - SuccessResponse: response with no data
func NewEmptySuccessResponse() *SuccessResponse {
	var r *SuccessResponse = new(SuccessResponse)

	r.Success = true
	r.Data = nil

	return r
}

// NewSuccessResponse - create success response with provided data
//
// params:
//   - data interface{} - object to return inside response
//
// return type:
//   - *SuccessResponse: response with provided data
func NewSuccessResponse(data interface{}) *SuccessResponse {
	var r *SuccessResponse = new(SuccessResponse)

	r.Success = true
	r.Data = data

	return r
}
