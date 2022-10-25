package responses

type DiskCreateSuccessResponse struct {
	// do not import the 'modules' package here
	Disk interface{} `json:"disk"`
	Link string      `json:"link"`
}

func CreateDiskSuccessResponse(disk interface{}, link string) *EmptySuccessResponse {
	_data := DiskCreateSuccessResponse{
		Disk: disk,
		Link: link,
	}
	return CreateEmptySuccessResponse(_data)
}
