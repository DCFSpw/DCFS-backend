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

type FileRequestBlockResponse struct {
	UUID  string
	Order int
	Size  int
}

type FileRequestResponse struct {
	SuccessResponse
	UUID   string
	Name   string
	Type   int
	Size   int
	Blocks []FileRequestBlockResponse
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
