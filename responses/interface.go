package responses

type Response interface {
}

type SuccessResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

type FailureResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

type ValidationErrorResponse struct {
	Success bool   `json:"success"`
	Code    string `json:"code"`
	Message string `json:"message"`
}

type OperationFailureResponse struct {
	Success bool   `json:"success"`
	Code    string `json:"code"`
	Message string `json:"message"`
}

type BlockDownloadResponse struct {
	Success bool    `json:"success"`
	Message string  `json:"message"`
	Block   []uint8 `json:"block"`
}

type DiskOAuthCodeResponse struct {
	SuccessResponse
	DiskUUID string `json:"diskUUID"`
}
