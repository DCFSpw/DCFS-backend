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
)

func CreateDisk(c *gin.Context) {
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

	volume := models.Transport.GetVolume(volumeUUID)
	if volume == nil {
		c.JSON(422, responses.NewValidationErrorResponseSingle(constants.VAL_UUID_INVALID, "VolumeUUID", "A volume of provided UUID does not exist"))
		return
	}

	_disk := dbo.Disk{
		AbstractDatabaseObject: dbo.AbstractDatabaseObject{
			UUID: uuid.New(),
		},
		UserUUID:     userUUID,
		VolumeUUID:   volumeUUID,
		ProviderUUID: provider.UUID,
		Credentials:  requestBody.Credentials.ToString(),
		Provider:     *provider,
		Name:         requestBody.Name,
		UsedSpace:    0,
		TotalSpace:   requestBody.TotalSpace,
	}
	disk := models.CreateDisk(models.CreateDiskMetadata{
		Disk:   &_disk,
		Volume: volume,
	})
	if disk == nil {
		c.JSON(422, responses.NewValidationErrorResponseSingle(constants.VAL_PROVIDER_NOT_SUPPORTED, "ProviderUUID", "Provided ProviderUUID is not a supported provider"))
		return
	}
	db.DB.DatabaseHandle.Create(&_disk)

	_, ok := disk.(OAuthDisk.OAuthDisk)
	if ok {
		config := disk.(OAuthDisk.OAuthDisk).GetConfig()
		authCode = config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	} else {
		// Refresh partitioner for credential based disks
		// OAuth disks will refresh partitioner after authorization
		go volume.RefreshPartitioner()
	}

	// TO DO! Verify that the disk space quota is valid
	/*_, totalSpace, errCode := disk.GetProviderSpace()
	if errCode == constants.SUCCESS && totalSpace < _disk.TotalSpace {
		c.JSON(422, responses.NewValidationErrorResponseSingle(constants.VAL_QUOTA_EXCEEDED, "TotalSpace", "Provided total space exceeds the disk space quota"))
	}*/

	// load full database object with a provider and a volume to return
	err = db.DB.DatabaseHandle.Where("uuid = ?", disk.GetUUID().String()).Preload("Provider").Preload("Volume").Find(&_disk).Error
	if err != nil {
		// this should never happen
		c.JSON(500, responses.NewOperationFailureResponse(constants.DATABASE_DISK_NOT_FOUND, "Could not validate database change"))
		return
	}

	c.JSON(200, responses.CreateDiskSuccessResponse(_disk, authCode))
}

func DiskOAuth(c *gin.Context) {
	var requestBody requests.OAuthRequest
	var _diskUUID string
	var diskUUID uuid.UUID
	var err error
	var _disk dbo.Disk

	// Retrieve and validate data from request
	if err = c.ShouldBindJSON(&requestBody); err != nil {
		c.JSON(422, responses.NewValidationErrorResponse(err))
		return
	}

	userData, _ := c.Get("UserData")
	userUUID := userData.(middleware.UserData).UserUUID

	_diskUUID = c.Param("DiskUUID")
	diskUUID, err = uuid.Parse(_diskUUID)

	err = db.DB.DatabaseHandle.Where("uuid = ?", _diskUUID).Preload("Provider").Preload("Volume").Find(&_disk).Error
	if err != nil {
		c.JSON(404, responses.NewNotFoundErrorResponse(constants.DATABASE_DISK_NOT_FOUND, "Cannot find a disk with the provided UUID"))
		return
	}

	volume := models.Transport.GetVolume(_disk.VolumeUUID)
	if volume == nil {
		c.JSON(404, responses.NewNotFoundErrorResponse(constants.TRANSPORT_VOLUME_NOT_FOUND, "Cannot find a volume with the provided UUID"))
		return
	}

	disk := (volume.GetDisk(diskUUID)).(OAuthDisk.OAuthDisk)
	if disk == nil {
		c.JSON(404, responses.NewNotFoundErrorResponse(constants.TRANSPORT_DISK_NOT_FOUND, "The provided disk is not associated with the provided volume"))
		return
	}

	config := disk.GetConfig()
	config.Endpoint.AuthStyle = oauth2.AuthStyleInParams
	tok, err := config.Exchange(c, requestBody.Code)
	if err != nil {
		c.JSON(500, responses.NewOperationFailureResponse(constants.OAUTH_BAD_CODE, "Could not retrieve the oauth token"))
		return
	}

	disk.SetCredentials(&credentials2.OauthCredentials{Token: tok})
	db.DB.DatabaseHandle.Save(disk.GetDiskDBO(userUUID, _disk.ProviderUUID, _disk.VolumeUUID))

	// load full database object with a provider and a volume to return
	err = db.DB.DatabaseHandle.Where("uuid = ?", _diskUUID).Preload("Provider").Preload("Volume").Find(&_disk).Error
	if err != nil {
		// this should never happen
		c.JSON(500, responses.NewOperationFailureResponse(constants.DATABASE_DISK_NOT_FOUND, "Could not validate database change"))
		return
	}

	go volume.RefreshPartitioner()

	c.JSON(200, responses.CreateEmptySuccessResponse(_disk))
}

func DiskGet(c *gin.Context) {
	var _diskUUID string
	var _disk dbo.Disk
	var volumeModel *models.Volume
	var diskModel models.Disk
	var err error

	// Retrieve disk UUID from request
	_diskUUID = c.Param("DiskUUID")

	// Retrieve disk from database
	err = db.DB.DatabaseHandle.Where("uuid = ?", _diskUUID).Preload("Provider").Preload("Volume").Find(&_disk).Error
	if err != nil {
		c.JSON(404, responses.NewNotFoundErrorResponse(constants.DATABASE_DISK_NOT_FOUND, "Cannot find a disk with the provided UUID"))
		return
	}

	// Compute free and total disk space
	volumeModel = models.Transport.GetVolume(_disk.VolumeUUID)
	if volumeModel == nil {
		c.JSON(404, responses.NewNotFoundErrorResponse(constants.TRANSPORT_VOLUME_NOT_FOUND, "Cannot find a volume with the provided UUID"))
		return
	}

	diskModel = volumeModel.GetDisk(_disk.UUID)
	if diskModel == nil {
		c.JSON(404, responses.NewNotFoundErrorResponse(constants.TRANSPORT_DISK_NOT_FOUND, "Cannot find a disk with the provided UUID"))
		return
	}

	_disk.FreeSpace = models.ComputeFreeSpace(diskModel)
	_disk.TotalSpace = diskModel.GetTotalSpace()

	c.JSON(200, responses.CreateEmptySuccessResponse(_disk))
}

func DiskUpdate(c *gin.Context) {
	var body requests.DiskUpdateRequest
	var _diskUUID string
	var diskUUID uuid.UUID
	var err error
	var volumes []*models.Volume
	var volume *models.Volume = nil
	var disk models.Disk = nil

	// Retrieve and validate data from request
	if err = c.ShouldBindJSON(&body); err != nil {
		c.JSON(422, responses.NewValidationErrorResponse(err))
		return
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

	cred := body.Credentials.ToString()
	if cred != "" {
		_, ok := disk.(OAuthDisk.OAuthDisk)
		if ok {
			c.JSON(405, responses.NewOperationFailureResponse(constants.TRANSPORT_DISK_IS_BEING_USED, "It is not allowed to change the credentials of an OAuth disk"))
			return
		}

		disk.CreateCredentials(cred)
	}

	disk.SetName(body.Name)

	// Verify that the disk space quota is valid
	_, totalSpace, errCode := disk.GetProviderSpace()
	if errCode == constants.SUCCESS {
		if totalSpace < body.TotalSpace {
			c.JSON(422, responses.NewValidationErrorResponseSingle(constants.VAL_QUOTA_EXCEEDED, "TotalSpace", "Provided total space exceeds the disk space quota"))
			return
		} else if disk.GetUsedSpace() > body.TotalSpace {
			c.JSON(422, responses.NewValidationErrorResponseSingle(constants.VAL_QUOTA_EXCEEDED, "TotalSpace", "Provided total space is lower than currently used space"))
			return
		} else {
			disk.SetTotalSpace(body.TotalSpace)
		}
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

	go volume.RefreshPartitioner()

	c.JSON(200, responses.CreateEmptySuccessResponse(diskDBO))
}

func DiskDelete(c *gin.Context) {
	var _diskUUID string
	var _disk dbo.Disk
	var _blocks []dbo.Block
	var volume *models.Volume
	var err error

	_diskUUID = c.Param("DiskUUID")
	err = db.DB.DatabaseHandle.Where("uuid = ?", _diskUUID).Preload("Provider").Preload("Volume").Find(&_disk).Error
	if err != nil {
		c.JSON(404, responses.NewNotFoundErrorResponse(constants.DATABASE_DISK_NOT_FOUND, "Could not find the disk with the provided UUID"))
		return
	}

	_d := models.Transport.FindEnqueuedDisk(_disk.UUID)
	if _d != nil {
		c.JSON(405, responses.NewOperationFailureResponse(constants.TRANSPORT_DISK_IS_BEING_USED, "Requested disk is enqueued for an IO operation, can't delete it now"))
		return
	}

	err = db.DB.DatabaseHandle.Where("disk_uuid = ?", _diskUUID).Find(&_blocks).Error
	if err != nil {
		c.JSON(500, responses.NewOperationFailureResponse(constants.DATABASE_ERROR, "Could not update database"))
		return
	}

	if len(_blocks) > 0 {
		c.JSON(405, responses.NewOperationFailureResponse(constants.TRANSPORT_DISK_IS_BEING_USED, "The provided disk is not empty and thus cannot be deleted"))
		return
	}

	volume = models.Transport.GetVolume(_disk.Volume.UUID)
	volume.DeleteDisk(_disk.UUID)
	db.DB.DatabaseHandle.Where("uuid = ?", _diskUUID).Delete(&_disk)

	go volume.RefreshPartitioner()

	c.JSON(200, responses.CreateEmptySuccessResponse(_disk))
}

func GetDisks(c *gin.Context) {
	var userUUID uuid.UUID
	var _disks []dbo.Disk
	var disks []interface{}
	var page int

	page = requests.GetPageFromQuery(c)

	userData, _ := c.Get("UserData")
	userUUID = userData.(middleware.UserData).UserUUID

	db.DB.DatabaseHandle.Where("user_uuid = ?", userUUID.String()).Preload("Provider").Preload("Volume").Find(&_disks)
	for _, disk := range _disks {
		// Update disk spaced based on local data (for performance reasons)
		disk.FreeSpace = disk.TotalSpace - disk.UsedSpace
		disk.TotalSpace = disk.TotalSpace

		// Append disk to the list
		disks = append(disks, disk)
	}

	pagination := models.Paginate(disks, page, constants.PAGINATION_RECORDS_PER_PAGE)
	c.JSON(200, responses.NewPaginationResponse(responses.PaginationData{Pagination: pagination.Pagination, Data: pagination.Data}))
}
