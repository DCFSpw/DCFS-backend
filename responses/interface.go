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
	UUID         string `json:"uuid"`
	Name         string `json:"name"`
	ProviderUUID string `json:"provider_uuid"`
	Link         string `json:"link"`
}

type DiskCreateResponse struct {
	SuccessResponse
	Response DiskOAuthCodeResponse `json:"response"`
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
