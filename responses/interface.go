package responses

type Response interface {
}

type EmptySuccessResponse struct {
	Success bool    `json:"success"`
	Data    []uint8 `json:"data"`
}

type SuccessResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

type FailureResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Code    string `json:"code"`
}

type OperationFailureResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Code    string `json:"code"`
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

func NewEmptySuccessResponse() *EmptySuccessResponse {
	var r *EmptySuccessResponse = new(EmptySuccessResponse)
	r.Success = true
	r.Data = nil
	return r
}
