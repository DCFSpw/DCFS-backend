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
