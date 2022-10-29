package requests

type VolumeSettingsRequest struct {
	Backup        int `json:"backup" binding:"required,min=1,max=2"`
	Encryption    int `json:"encryption" binding:"required,min=1,max=2"`
	FilePartition int `json:"filePartition" binding:"required,min=1,max=2"`
}

type VolumeCreateRequest struct {
	Name     string                `json:"name" binding:"required,gte=1,lte=64"`
	Settings VolumeSettingsRequest `json:"settings" binding:"required"`
}
