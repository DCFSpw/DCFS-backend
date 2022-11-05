package responses

type NotFoundErrorResponse struct {
	// Error code: 404
	FailureResponse
}

func NewNotFoundErrorResponse(code string, message string) *NotFoundErrorResponse {
	var r *NotFoundErrorResponse = new(NotFoundErrorResponse)

	r.Success = false
	r.Message = message
	r.Code = code

	return r
}

type OperationFailureResponse struct {
	// Error code: 500 or 405
	Success bool   `json:"success"`
	Message string `json:"message"`
	Code    string `json:"code"`
}

func NewOperationFailureResponse(code string, message string) *OperationFailureResponse {
	var r *OperationFailureResponse = new(OperationFailureResponse)

	r.Success = false
	r.Message = message
	r.Code = code
	
	return r
}
