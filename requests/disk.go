package requests

type DiskCreateRequest struct {
	Name         string `json:"name" binding:"required,gte=1,lte=64"`
	ProviderUUID string `json:"providerUUID" binding:"required"`
	VolumeUUID   string `json:"volumeUUID" binding:"required"`
	Credentials  string `json:"credentials" binding:"required"`
}

type OAuthRequest struct {
	VolumeUUID   string `json:"volumeUUID" binding:"required"`
	DiskUUID     string `json:"diskUUID" binding:"required"`
	ProviderUUID string `json:"providerUUID" binding:"required"`
	Code         string `json:"code" binding:"required"`
}
