package controllers

import (
	"dcfs/constants"
	"dcfs/db"
	"dcfs/db/dbo"
	"dcfs/middleware"
	"dcfs/models"
	credentials2 "dcfs/models/credentials"
	"dcfs/models/disk/OAuthDisk"
	"dcfs/requests"
	"dcfs/responses"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"golang.org/x/oauth2"
	"log"
	"strconv"
)

func DiskCreate(c *gin.Context) {
	var requestBody requests.DiskCreateRequest = requests.DiskCreateRequest{}
	var provider *dbo.Provider = dbo.NewProvider()
	var authCode string = ""
	userData, _ := c.Get("UserData")
	userUUID := userData.(middleware.UserData).UserUUID

	// Retrieve and validate data from request
	if err := c.ShouldBindJSON(&requestBody); err != nil {
		c.JSON(422, responses.NewValidationErrorResponse(err))
		return
	}

	log.Println("requests body: ", requestBody)
	log.Println("userUUID: ", userUUID)

	// Get provider info
	db.DB.DatabaseHandle.Where("uuid = ?", requestBody.ProviderUUID).First(&provider)

	volumeUUID, err := uuid.Parse(requestBody.VolumeUUID)
	if err != nil {
		c.JSON(422, responses.NewValidationErrorResponseSingle(constants.VAL_UUID_INVALID, "VolumeUUID", "Provided VolumeUUID is not a valid UUID"))
		return
	}

	volume := models.Transport.GetVolume(userUUID, volumeUUID)
	_disk := dbo.Disk{
		AbstractDatabaseObject: dbo.AbstractDatabaseObject{
			UUID: uuid.New(),
		},
		UserUUID:     userUUID,
		VolumeUUID:   volumeUUID,
		ProviderUUID: provider.UUID,
		Credentials:  requestBody.Credentials.ToString(),
		Provider:     *provider,
	}
	disk := models.CreateDisk(models.CreateDiskMetadata{
		Disk:   &_disk,
		Volume: volume,
	})
	if disk == nil {
		c.JSON(422, responses.NewValidationErrorResponseSingle(constants.VAL_PROVIDER_NOT_SUPPORTED, "ProviderUUID", "Provided ProviderUUID is not a supported provider"))
		return
	}
	db.DB.DatabaseHandle.Omit("Provider", "Volume").Create(_disk)

	_, ok := disk.(OAuthDisk.OAuthDisk)
	if ok {
		config := disk.(OAuthDisk.OAuthDisk).GetConfig()
		authCode = config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	}

	c.JSON(200, responses.DiskCreateSuccessResponse{
		SuccessResponse: responses.SuccessResponse{Success: true, Message: "Successfully started the procedure of adding a disk"},
		Data: responses.DiskOAuthCodeResponse{
			UUID:         disk.GetUUID().String(),
			Name:         requestBody.Name,
			ProviderUUID: provider.UUID.String(),
			Link:         authCode,
		},
	})
}

func DiskOAuth(c *gin.Context) {
	var requestBody requests.OAuthRequest

	// Retrieve and validate data from request
	if err := c.ShouldBindJSON(&requestBody); err != nil {
		c.JSON(422, responses.NewValidationErrorResponse(err))
		return
	}

	userData, _ := c.Get("UserData")
	userUUID := userData.(middleware.UserData).UserUUID

	volumeUUID, err := uuid.Parse(requestBody.VolumeUUID)
	if err != nil {
		c.JSON(422, responses.NewValidationErrorResponseSingle(constants.VAL_UUID_INVALID, "VolumeUUID", "Provided VolumeUUID is not a valid UUID"))
		return
	}

	diskUUID, err := uuid.Parse(requestBody.DiskUUID)
	if err != nil {
		c.JSON(422, responses.NewValidationErrorResponseSingle(constants.VAL_UUID_INVALID, "DiskUUID", "Provided DiskUUID is not a valid UUID"))
		return
	}

	providerUUID, err := uuid.Parse(requestBody.ProviderUUID)
	if err != nil {
		c.JSON(422, responses.NewValidationErrorResponseSingle(constants.VAL_UUID_INVALID, "ProviderUUID", "Provided ProviderUUID is not a valid UUID"))
		return
	}

	volume := models.Transport.GetVolume(userUUID, volumeUUID)
	if volume == nil {
		c.JSON(404, responses.NewNotFoundErrorResponse(constants.TRANSPORT_VOLUME_NOT_FOUND, "Cannot find volume with provided UUID"))
		return
	}

	disk := (volume.GetDisk(diskUUID)).(OAuthDisk.OAuthDisk)
	if disk == nil {
		c.JSON(404, responses.NewNotFoundErrorResponse(constants.TRANSPORT_DISK_NOT_FOUND, "Cannot find disk with provided UUID or disk is not associated with provided volume"))
		return
	}

	config := disk.GetConfig()
	config.Endpoint.AuthStyle = oauth2.AuthStyleInParams
	tok, err := config.Exchange(c, requestBody.Code)
	if err != nil {
		log.Fatalf("Unable to retrieve token from web %v", err)
	}

	disk.SetCredentials(&credentials2.OauthCredentials{Token: tok})
	db.DB.DatabaseHandle.Save(disk.GetDiskDBO(userUUID, providerUUID, volumeUUID))

	c.JSON(200, responses.NewEmptySuccessResponse())
}

func DiskGet(c *gin.Context) {
	c.JSON(200, responses.SuccessResponse{Success: true, Message: "Disk Get Endpoint"})
}

func DiskUpdate(c *gin.Context) {
	var body requests.DiskUpdateRequest
	var _diskUUID string
	var diskUUID uuid.UUID
	var err error
	var volumes []*models.Volume
	var volume *models.Volume = nil
	var disk models.Disk = nil

	err = c.ShouldBindJSON(&body)
	if err != nil {
		c.JSON(401, responses.NewValidationErrorResponse(err))
	}

	userData, _ := c.Get("UserData")
	userUUID := userData.(middleware.UserData).UserUUID

	_diskUUID = c.Param("DiskUUID")
	diskUUID, err = uuid.Parse(_diskUUID)
	if err != nil {
		c.JSON(422, responses.NewValidationErrorResponseSingle(constants.VAL_UUID_INVALID, "DiskUUID", "Provided DiskUUID is not a valid UUID"))
		return
	}

	volumes = models.Transport.GetVolumes(userUUID)
	if volumes == nil {
		c.JSON(404, responses.NewNotFoundErrorResponse(constants.TRANSPORT_VOLUME_NOT_FOUND, "Cannot find volume with provided UUID"))
		return
	}

	for _, _volume := range volumes {
		disk = _volume.GetDisk(diskUUID)
		if disk != nil {
			volume = _volume
			break
		}
	}

	if disk == nil {
		c.JSON(404, responses.NewNotFoundErrorResponse(constants.TRANSPORT_DISK_NOT_FOUND, "Cannot find volume with provided UUID"))
		return
	}

	_disk := models.Transport.FindEnqueuedDisk(diskUUID)
	if _disk != nil {
		c.JSON(405, responses.NewOperationFailureResponse(constants.TRANSPORT_DISK_IS_BEING_USED, "Requested disk is enqueued for an IO operation, can't update it now"))
		return
	}

	// TODO: change disk's name: in db and in the backend
	cred := body.Credentials.ToString()
	if cred != "" {
		_, ok := disk.(OAuthDisk.OAuthDisk)
		if ok {
			c.JSON(405, responses.NewOperationFailureResponse(constants.TRANSPORT_DISK_IS_BEING_USED, "It is not allowed to change the credentials of an OAuth disk"))
			return
		}

		disk.CreateCredentials(cred)
	}

	diskDBO := disk.GetDiskDBO(userUUID, disk.GetProviderUUID(), volume.UUID)
	err = db.DB.DatabaseHandle.Save(&diskDBO).Error
	if err != nil {
		c.JSON(500, responses.NewOperationFailureResponse(constants.DATABASE_ERROR, "Could not update database"))
		return
	}

	// load full database object with a provider and a volume to return
	err = db.DB.DatabaseHandle.Where("uuid = ?", _diskUUID).Preload("Provider").Preload("Volume").Find(&diskDBO).Error
	if err != nil {
		// this should never happen
		c.JSON(500, responses.NewOperationFailureResponse(constants.DATABASE_DISK_NOT_FOUND, "Could not validate database change"))
		return
	}

	c.JSON(200, responses.CreateEmptySuccessResponse(diskDBO))
}

func DiskDelete(c *gin.Context) {
	c.JSON(200, responses.SuccessResponse{Success: true, Message: "Disk Delete Endpoint"})
}

func GetDisks(c *gin.Context) {
	var userUUID uuid.UUID
	var _disks []dbo.Disk
	var disks []interface{}
	var page int
	var err error

	page, err = strconv.Atoi(c.Request.URL.Query().Get("page"))
	if err != nil {
		page = 1
	}

	userData, _ := c.Get("UserData")
	userUUID = userData.(middleware.UserData).UserUUID

	db.DB.DatabaseHandle.Where("user_uuid = ?", userUUID.String()).Preload("Provider").Preload("Volume").Find(&_disks)
	for _, disk := range _disks {
		disks = append(disks, disk)
	}

	pagination := models.Paginate(disks, page, constants.PAGINATION_RECORDS_PER_PAGE)
	c.JSON(200, responses.NewPaginationResponse(responses.PaginationData{Pagination: pagination.Pagination, Data: pagination.Data}))
}

func DiskAssociate(c *gin.Context) {
	c.JSON(200, responses.SuccessResponse{Success: true, Message: "Disk Associate Endpoint"})
}

func DiskDissociate(c *gin.Context) {
	c.JSON(200, responses.SuccessResponse{Success: true, Message: "Disk Dissociate Endpoint"})
}
