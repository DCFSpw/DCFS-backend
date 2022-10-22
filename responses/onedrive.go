package responses

type CreateUploadSessionResponse struct {
	UploadUrl          string `json:"uploadUrl"`
	ExpirationDateTime string `json:"expirationDateTime"`
}

type UploadSessionResponse struct {
	ExpirationDateTime string   `json:"expirationDateTime"`
	NextExpectedRanges []string `json:"nextExpectedRanges"`
}

type UploadSessionFinalResponse struct {
	ID   string      `json:"id"`
	Name string      `json:"name"`
	Size int         `json:"size"`
	File interface{} `json:"file"`
}
