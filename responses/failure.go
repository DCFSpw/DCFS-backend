package responses

type FailureResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Code    string `json:"code"`
}

type OperationFailureResponse struct {
	// Error code: 500 or 405
	Success bool   `json:"success"`
	Message string `json:"message"`
	Code    string `json:"code"`
}

// NewOperationFailureResponse - create generic operation failure response
//
// params:
//   - code string: error code
//   - message string: with error message
//
// return type:
//   - *OperationFailureResponse: response with error data
func NewOperationFailureResponse(code string, message string) *OperationFailureResponse {
	var r *OperationFailureResponse = new(OperationFailureResponse)

	r.Success = false
	r.Message = message
	r.Code = code

	return r
}

type NotFoundErrorResponse struct {
	// Error code: 404
	FailureResponse
}

// NewNotFoundErrorResponse - create not found error response
//
// params:
//   - code string - error code
//   - message string - error message
//
// return type:
//   - *NotFoundErrorResponse: response with error data
func NewNotFoundErrorResponse(code string, message string) *NotFoundErrorResponse {
	var r *NotFoundErrorResponse = new(NotFoundErrorResponse)

	r.Success = false
	r.Message = message
	r.Code = code

	return r
}
