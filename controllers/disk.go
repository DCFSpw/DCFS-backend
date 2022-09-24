package controllers

import (
	"dcfs/db"
	"dcfs/db/dbo"
	"dcfs/middleware"
	"dcfs/models"
	credentials2 "dcfs/models/credentials"
	disk2 "dcfs/models/disk"
	"dcfs/models/disk/DriveFactory"
	"dcfs/responses"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"golang.org/x/oauth2"
	"log"
)

type diskCreateBody struct {
	Name         string `json:"name"`
	ProviderUUID string `json:"providerUUID"`
	VolumeUUID   string `json:"volumeUUID"`
	Credentials  string `json:"credentials"`
}

func DiskCreate(c *gin.Context) {
	var body diskCreateBody = diskCreateBody{}
	var provider *dbo.Provider = dbo.NewProvider()
	userData, _ := c.Get("UserData")
	_userUUID := userData.(middleware.UserData).UserUUID

	// unpack request
	if err := c.Bind(&body); err != nil {
		// TODO: error handling
		panic("Unimplemented")
	}
	log.Println("requests body: ", body)
	log.Println("userUUID: ", _userUUID)

	// get provider info
	db.DB.DatabaseHandle.Where("uuid = ?", body.ProviderUUID).First(&provider)

	volumeUUID, err := uuid.Parse(body.VolumeUUID)
	if err != nil {
		// TODO: error handling
		panic(err)
	}

	userUUID := _userUUID

	disk := DriveFactory.NewDisk(provider.ProviderType)
	if disk == nil {
		c.JSON(401, responses.SuccessResponse{Success: true, Message: "Provider not supported"})
		return
	}

	disk.SetUUID(uuid.New())

	// get volume handle
	volume := models.Transport.GetVolume(userUUID, volumeUUID)
	volume.AddDisk(disk.GetUUID(), disk)
	providerUUID, _ := uuid.Parse(body.ProviderUUID)

	if provider.ProviderType == dbo.SFTP {
		disk.SetCredentials(credentials2.NewSFTPCredentials(body.Credentials))
		db.DB.DatabaseHandle.Create(disk.GetDiskDBO(userUUID, providerUUID, volumeUUID))

		// TODO: update return value
		c.JSON(200, responses.SuccessResponse{Success: true, Message: "Success"})
		return
	}

	if provider.ProviderType == dbo.ONEDRIVE || provider.ProviderType == dbo.GDRIVE {
		db.DB.DatabaseHandle.Create(disk.GetDiskDBO(userUUID, providerUUID, volumeUUID))

		config := disk.(disk2.OAuthDisk).GetConfig()
		c.JSON(200, responses.DiskOAuthCodeResponse{SuccessResponse: responses.SuccessResponse{Success: true, Message: config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)}, DiskUUID: disk.GetUUID().String()})
		return
	}
}

type oauthBody struct {
	VolumeUUID   string `json:"volumeUUID"`
	DiskUUID     string `json:"diskUUID"`
	ProviderUUID string `json:"providerUUID"`
	Code         string `json:"code"`
}

func DiskOAuth(c *gin.Context) {
	var body oauthBody

	err := c.Bind(&body)
	if err != nil {
		// TODO: error handling
		panic("Unimplemented")
	}

	userData, _ := c.Get("UserData")
	userUUID := userData.(middleware.UserData).UserUUID

	volumeUUID, err := uuid.Parse(body.VolumeUUID)
	if err != nil {
		// TODO
		panic("Unimplemented")
	}

	diskUUID, err := uuid.Parse(body.DiskUUID)
	if err != nil {
		// TODO
		panic("Unimplemented")
	}

	providerUUID, err := uuid.Parse(body.ProviderUUID)
	if err != nil {
		// TODO
		panic("Unimplemented")
	}

	volume := models.Transport.GetVolume(userUUID, volumeUUID)
	if volume == nil {
		// TODO
		panic("Unimplemented")
	}

	disk := (volume.GetDisk(diskUUID)).(disk2.OAuthDisk)
	if disk == nil {
		// TODO
		panic("Unimplemented")
	}

	config := disk.GetConfig()
	tok, err := config.Exchange(c, body.Code)
	if err != nil {
		log.Fatalf("Unable to retrieve token from web %v", err)
	}

	disk.SetCredentials(&credentials2.OauthCredentials{Token: tok})
	db.DB.DatabaseHandle.Save(disk.GetDiskDBO(userUUID, providerUUID, volumeUUID))
	// TODO: update return value
	c.JSON(200, responses.SuccessResponse{Success: true, Message: "Success"})
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
