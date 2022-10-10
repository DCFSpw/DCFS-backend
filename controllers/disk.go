package controllers

import (
	"dcfs/constants"
	"dcfs/db"
	"dcfs/db/dbo"
	"dcfs/middleware"
	"dcfs/models"
	credentials2 "dcfs/models/credentials"
	disk2 "dcfs/models/disk"
	"dcfs/models/disk/DriveFactory"
	"dcfs/requests"
	"dcfs/responses"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"golang.org/x/oauth2"
	"log"
)

func DiskCreate(c *gin.Context) {
	var requestBody requests.DiskCreateRequest = requests.DiskCreateRequest{}
	var provider *dbo.Provider = dbo.NewProvider()
	var authCode string = ""
	userData, _ := c.Get("UserData")
	_userUUID := userData.(middleware.UserData).UserUUID

	// Retrieve and validate data from request
	if err := c.ShouldBindJSON(&requestBody); err != nil {
		c.JSON(422, responses.NewValidationErrorResponse(err))
		return
	}

	log.Println("requests body: ", requestBody)
	log.Println("userUUID: ", _userUUID)

	// Get provider info
	db.DB.DatabaseHandle.Where("uuid = ?", requestBody.ProviderUUID).First(&provider)

	volumeUUID, err := uuid.Parse(requestBody.VolumeUUID)
	if err != nil {
		c.JSON(422, responses.NewValidationErrorResponseSingle(constants.VAL_UUID_INVALID, "VolumeUUID", "Provided VolumeUUID is not a valid UUID"))
		return
	}

	userUUID := _userUUID

	disk := DriveFactory.NewDisk(provider.Type)
	if disk == nil {
		c.JSON(422, responses.NewValidationErrorResponseSingle(constants.VAL_PROVIDER_NOT_SUPPORTED, "ProviderUUID", "Provided ProviderUUID is not a supported provider"))
		return
	}

	disk.SetUUID(uuid.New())

	// Get volume handle
	volume := models.Transport.GetVolume(userUUID, volumeUUID)
	volume.AddDisk(disk.GetUUID(), disk)
	providerUUID, _ := uuid.Parse(requestBody.ProviderUUID)

	if provider.Type == constants.PROVIDER_TYPE_SFTP {
		disk.SetCredentials(credentials2.NewSFTPCredentials(requestBody.Credentials))
		db.DB.DatabaseHandle.Create(disk.GetDiskDBO(userUUID, providerUUID, volumeUUID))
	}

	if provider.Type == constants.PROVIDER_TYPE_ONEDRIVE || provider.Type == constants.PROVIDER_TYPE_GDRIVE {
		db.DB.DatabaseHandle.Create(disk.GetDiskDBO(userUUID, providerUUID, volumeUUID))
		config := disk.(disk2.OAuthDisk).GetConfig()
		authCode = config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	}

	c.JSON(200, responses.DiskCreateSuccessResponse{
		SuccessResponse: responses.SuccessResponse{Success: true, Message: "Successfully started the procedure of adding a disk"},
		Data: responses.DiskOAuthCodeResponse{
			UUID:         disk.GetUUID().String(),
			Name:         requestBody.Name,
			ProviderUUID: providerUUID.String(),
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

	disk := (volume.GetDisk(diskUUID)).(disk2.OAuthDisk)
	if disk == nil {
		c.JSON(404, responses.NewNotFoundErrorResponse(constants.TRANSPORT_DISK_NOT_FOUND, "Cannot find disk with provided UUID or disk is not associated with provided volume"))
		return
	}

	config := disk.GetConfig()
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
	c.JSON(200, responses.SuccessResponse{Success: true, Message: "Disk Update Endpoint"})
}

func DiskDelete(c *gin.Context) {
	c.JSON(200, responses.SuccessResponse{Success: true, Message: "Disk Delete Endpoint"})
}

func GetDisks(c *gin.Context) {
	c.JSON(200, responses.SuccessResponse{Success: true, Message: "Get Disks Endpoint"})
}

func DiskAssociate(c *gin.Context) {
	c.JSON(200, responses.SuccessResponse{Success: true, Message: "Disk Associate Endpoint"})
}

func DiskDissociate(c *gin.Context) {
	c.JSON(200, responses.SuccessResponse{Success: true, Message: "Disk Dissociate Endpoint"})
}
