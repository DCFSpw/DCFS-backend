package responses

type DiskCreateSuccessResponse struct {
	// Do not import the 'modules' package here
	// to avoid circular dependencies
	Disk interface{} `json:"disk"`
	Link string      `json:"link"`
}

// NewCreateDiskSuccessResponse - create disk creation success response
//
// params:
//   - diskData: dbo.Disk object with disk data to return
//   - link: string with authorization link for OAuth disks
//
// return type:
//   - response: SuccessResponse with disk data
func NewCreateDiskSuccessResponse(diskData interface{}, link string) *SuccessResponse {
	_data := DiskCreateSuccessResponse{
		Disk: diskData,
		Link: link,
	}
	return NewSuccessResponse(_data)
}
