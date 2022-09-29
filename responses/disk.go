package responses

type DiskCreateSuccessResponse struct {
	SuccessResponse
	Data DiskOAuthCodeResponse `json:"Data"`
}

type DiskOAuthCodeResponse struct {
	UUID         string `json:"uuid"`
	Name         string `json:"name"`
	ProviderUUID string `json:"provider_uuid"`
	Link         string `json:"link"`
}
