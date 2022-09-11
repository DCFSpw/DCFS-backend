package responses

type Response interface {
}

type SuccessResponse struct {
	Success bool
	Msg     string
}

type ValidationErrorResponse struct {
	Success bool
	Msg     string
}

type OperationFailureResponse struct {
	Success bool
	Msg     string
}

type BlockDownloadResponse struct {
	Success bool
	Msg     string
	Block   []uint8
}

type DiskOAuthCodeResponse struct {
	SuccessResponse
	DiskUUID string
}
