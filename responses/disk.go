package responses

type DiskCreateSuccessResponse struct {
	EmptySuccessResponse
	Link string `json:"link"`
}

func CreateDiskSuccessResponse(data interface{}, link string) *DiskCreateSuccessResponse {
	return &DiskCreateSuccessResponse{
		EmptySuccessResponse: *CreateEmptySuccessResponse(data),
		Link:                 link,
	}
}
