package responses

type Response interface {
}

type SuccessResponse struct {
	Success bool
	Msg     string
}

type DiskOAuthCodeResponse struct {
	SuccessResponse
	DiskUUID string
}
