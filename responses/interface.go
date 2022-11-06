package responses

type Response interface {
}

type EmptySuccessResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data"`
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

type BlockDownloadResponse struct {
	Success bool    `json:"success"`
	Message string  `json:"message"`
	Block   []uint8 `json:"block"`
}

func NewEmptySuccessResponse() *EmptySuccessResponse {
	var r *EmptySuccessResponse = new(EmptySuccessResponse)

	r.Success = true
	r.Data = nil

	return r
}

func CreateEmptySuccessResponse(data interface{}) *EmptySuccessResponse {
	return &EmptySuccessResponse{
		Success: true,
		Data:    data,
	}
}
