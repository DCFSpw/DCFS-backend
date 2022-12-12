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
	"dcfs/util/logger"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"golang.org/x/oauth2"
	"strconv"
)

// CreateDisk - handler for Create disk request
//
// Create disk (POST /disks/manage) - creating a new disk. This is the final
// creation request for the credential-based disks.
//
// params:
//   - c *gin.Context: context of the request
//
// return type:
//   - API response with appropriate HTTP code
func CreateDisk(c *gin.Context) {
	var requestBody requests.DiskCreateRequest = requests.DiskCreateRequest{}
	var provider *dbo.Provider = dbo.NewProvider()
	var authCode string = ""
	var userUUID uuid.UUID

	// Retrieve userUUID from context
	userUUID = c.MustGet("UserData").(middleware.UserData).UserUUID

	// Retrieve and validate data from request
	if err := c.ShouldBindJSON(&requestBody); err != nil {
		logger.Logger.Error("api", "Wrong request body.")
		c.JSON(422, responses.NewValidationErrorResponse(err))
		return
	}

	// Get provider info
	dbErr := db.DB.DatabaseHandle.Where("uuid = ?", requestBody.ProviderUUID).First(&provider).Error
	if dbErr != nil {
		logger.Logger.Error("api", "Received a request with a wrong provider UUID.")
		c.JSON(422, responses.NewValidationErrorResponseSingle(constants.VAL_UUID_INVALID, "ProviderUUID", "A provider with provided UUID does not exists"))
		return
	}

	// Parse volumeUUID from request
	volumeUUID, err := uuid.Parse(requestBody.VolumeUUID)
	if err != nil {
		logger.Logger.Error("api", "Received a request with a wrong provider UUID.")
		c.JSON(422, responses.NewValidationErrorResponseSingle(constants.VAL_UUID_INVALID, "VolumeUUID", "Provided VolumeUUID is not a valid UUID"))
		return
	}

	// Retrieve volume from transport
	volume := models.Transport.GetVolume(volumeUUID)
	if volume == nil {
		logger.Logger.Error("api", "Received a request with a wrong volume UUID")
		c.JSON(422, responses.NewValidationErrorResponseSingle(constants.VAL_UUID_INVALID, "VolumeUUID", "A volume of provided UUID does not exist"))
		return
	}

	// Create disk object
	_disk := dbo.Disk{
		AbstractDatabaseObject: dbo.AbstractDatabaseObject{
			UUID: uuid.New(),
		},
		UserUUID:        userUUID,
		VolumeUUID:      volumeUUID,
		ProviderUUID:    provider.UUID,
		Credentials:     requestBody.Credentials.ToString(),
		Provider:        *provider,
		Name:            requestBody.Name,
		UsedSpace:       0,
		TotalSpace:      requestBody.TotalSpace,
		IsVirtual:       false,
		VirtualDiskUUID: uuid.Nil,
	}
	disk := models.CreateDisk(models.CreateDiskMetadata{
		Disk:   &_disk,
		Volume: volume,
	})
	if disk == nil {
		logger.Logger.Error("api", "Could not create the desired disk.")
		c.JSON(422, responses.NewValidationErrorResponseSingle(constants.VAL_PROVIDER_NOT_SUPPORTED, "ProviderUUID", "Provided ProviderUUID is not a supported provider"))
		return
	}

	// Prepare OAuth disk config
	_, ok := disk.(OAuthDisk.OAuthDisk)
	if ok {
		config := disk.(OAuthDisk.OAuthDisk).GetConfig()
		authCode = config.AuthCodeURL("state-token", oauth2.AccessTypeOffline, oauth2.ApprovalForce)
		logger.Logger.Debug("api", "Got the oauth link code for the selected provider.")
	} else {
		// Retrieve disk quota from provider
		// OAuth disks will retrieve quota after authorization
		_, totalSpace, errCode := disk.GetProviderSpace()
		if errCode == constants.SUCCESS && totalSpace < _disk.TotalSpace {
			_disk.TotalSpace = totalSpace
			logger.Logger.Debug("api", "Set the disk total space to: ", strconv.FormatUint(totalSpace, 10))
		}

		// check if the credentials are correct
		if !disk.GetReadiness().IsReadyForce(c) {
			logger.Logger.Error("api", "Credentials for a new disk: ", requestBody.Credentials.ToString(), " were incorrect.")
			c.JSON(500, responses.NewOperationFailureResponse(constants.VAL_CREDENTIALS_INVALID, "Provided credentials were incorrect"))
			return
		}
	}

	// Find virtual disk uuid for new disk
	virtualDiskUUID, err := volume.GenerateVirtualDisk(disk)
	if err != nil {
		logger.Logger.Error("api", "Could not generate virtual disk UUID.")
		c.JSON(500, responses.NewOperationFailureResponse(constants.DATABASE_ERROR, "Could not generate virtual disk UUID (for volumes with backup): "+err.Error()))
		return
	} else if virtualDiskUUID != uuid.Nil {
		_disk.VirtualDiskUUID = virtualDiskUUID
	}

	// Save disk to database
	result := db.DB.DatabaseHandle.Create(&_disk)
	if result.Error != nil {
		logger.Logger.Error("api", "Could not save the newly created disk with the uuid: ", _disk.UUID.String(), " in the db.")
		c.JSON(500, responses.NewOperationFailureResponse(constants.DATABASE_ERROR, "Database operation failed: "+result.Error.Error()))
		return
	}
	logger.Logger.Debug("api", "Saved the newly created disk with the uuid: ", _disk.UUID.String(), " in the db.")

	// Load full database object with a provider and a volume to return
	err = db.DB.DatabaseHandle.Where("uuid = ?", disk.GetUUID().String()).Preload("Provider").Preload("Volume").Find(&_disk).Error
	if err != nil {
		c.JSON(500, responses.NewOperationFailureResponse(constants.DATABASE_DISK_NOT_FOUND, "Could not validate database change"))
		return
	}

	// Refresh partitioner after disk is added to the volume
	go volume.RefreshPartitioner()

	logger.Logger.Debug("api", "CreateDisk endpoint successful exit.")
	c.JSON(200, responses.NewCreateDiskSuccessResponse(_disk, authCode))
}

// DiskOAuth - handler for Provide OAuth token for disk request
//
// Provide OAuth token for disk (POST /disks/oauth/{diskUUID}) - providing an
// OAuth token for OAuth-based disks.
//
// params:
//   - c *gin.Context: context of the request
//
// return type:
//   - API response with appropriate HTTP code
func DiskOAuth(c *gin.Context) {
	var requestBody requests.OAuthRequest
	var _diskUUID string
	var diskUUID uuid.UUID
	var userUUID uuid.UUID
	var err error
	var _disk dbo.Disk

	// Retrieve and validate data from request
	if err = c.ShouldBindJSON(&requestBody); err != nil {
		logger.Logger.Error("api", "Wrong request body.")
		c.JSON(422, responses.NewValidationErrorResponse(err))
		return
	}

	// Retrieve userUUID from context
	userUUID = c.MustGet("UserData").(middleware.UserData).UserUUID

	// Retrieve nad parse diskUUID from param
	_diskUUID = c.Param("DiskUUID")
	diskUUID, err = uuid.Parse(_diskUUID)

	// Retrieve disk from database
	err = db.DB.DatabaseHandle.Where("uuid = ? AND is_virtual = ?", _diskUUID, false).Preload("Provider").Preload("Volume").Find(&_disk).Error
	if err != nil {
		logger.Logger.Error("api", "Could not find a disk with the given uuid: ", _diskUUID, " in the db.")
		c.JSON(404, responses.NewNotFoundErrorResponse(constants.DATABASE_DISK_NOT_FOUND, "Cannot find a disk with the provided UUID"))
		return
	}

	// Retrieve volume from transport
	volume := models.Transport.GetVolume(_disk.VolumeUUID)
	if volume == nil {
		logger.Logger.Error("api", "Could not find a volume with the given uuid: ", _disk.VolumeUUID.String(), " in the db.")
		c.JSON(404, responses.NewNotFoundErrorResponse(constants.TRANSPORT_VOLUME_NOT_FOUND, "Cannot find a volume with the provided UUID"))
		return
	}

	disk := (volume.GetDisk(diskUUID)).(OAuthDisk.OAuthDisk)
	if disk == nil {
		logger.Logger.Error("api", "The requested disk:", _diskUUID, " is not associated with the requested volume: ", volume.UUID.String(), ".")
		c.JSON(404, responses.NewNotFoundErrorResponse(constants.TRANSPORT_DISK_NOT_FOUND, "The provided disk is not associated with the provided volume"))
		return
	}

	// Exchange OAuth token for refresh token
	config := disk.GetConfig()
	config.Endpoint.AuthStyle = oauth2.AuthStyleInParams
	tok, err := config.Exchange(c, requestBody.Code, oauth2.AccessTypeOffline, oauth2.ApprovalForce)
	if err != nil {
		logger.Logger.Error("api", "Could not retrieve the oauth token, got error: ", err.Error())
		c.JSON(500, responses.NewOperationFailureResponse(constants.OAUTH_BAD_CODE, "Could not retrieve the oauth token"))
		return
	}
	logger.Logger.Debug("api", "Got the oauth token pair for the disk: ", _diskUUID, ".")

	if tok.RefreshToken == "" {
		logger.Logger.Error("api", "The refresh token was null")
		c.JSON(500, responses.NewOperationFailureResponse(constants.OAUTH_BAD_CODE, "The refresh token was null"))
		return
	}

	// Save refresh token to disk credentials
	disk.SetCredentials(&credentials2.OauthCredentials{Token: tok})

	// Retrieve disk quota from provider
	_, totalSpace, errCode := disk.GetProviderSpace()
	if errCode == constants.SUCCESS && totalSpace < disk.GetTotalSpace() {
		disk.SetTotalSpace(totalSpace)
	}
	logger.Logger.Debug("api", "The disk space has been set to: ", strconv.FormatUint(disk.GetTotalSpace(), 10), ".")

	// Save disk credentials to database
	_diskDBO := disk.GetDiskDBO(userUUID, _disk.ProviderUUID, _disk.VolumeUUID)
	result := db.DB.DatabaseHandle.Save(&_diskDBO)
	if result.Error != nil {
		logger.Logger.Error("api", "Could not save the disk: ", _diskUUID, " in the db.")
		c.JSON(500, responses.NewOperationFailureResponse(constants.DATABASE_ERROR, "Database operation failed: "+result.Error.Error()))
		return
	}
	logger.Logger.Debug("api", "Saved the disk: ", _diskUUID, " in the db.")

	// Load full database object with a provider and a volume to return
	err = db.DB.DatabaseHandle.Where("uuid = ?", _diskUUID).Preload("Provider").Preload("Volume").Find(&_disk).Error
	if err != nil {
		logger.Logger.Error("api", "Could not validate that the disk: ", _diskUUID, " has been saved in the db.")
		c.JSON(500, responses.NewOperationFailureResponse(constants.DATABASE_DISK_NOT_FOUND, "Could not validate database change"))
		return
	}

	go volume.RefreshPartitioner()

	logger.Logger.Debug("api", "DiskOAuth endpoint successful exit.")
	c.JSON(200, responses.NewSuccessResponse(_disk))
}

// GetDisk - handler for Get disk details request
//
// Get disk details (GET /disks/manage/{diskUUID}) - retrieving details of
// the specified disk.
//
// params:
//   - c *gin.Context: context of the request
//
// return type:
//   - API response with appropriate HTTP code
func GetDisk(c *gin.Context) {
	var _diskUUID string
	var _disk dbo.Disk
	var volumeModel *models.Volume
	var diskModel models.Disk
	var userUUID uuid.UUID
	var err error

	// Retrieve disk UUID from request
	_diskUUID = c.Param("DiskUUID")

	// Retrieve disk from database
	err = db.DB.DatabaseHandle.Where("uuid = ? AND is_virtual = ?", _diskUUID, false).Preload("Provider").Preload("Volume").Find(&_disk).Error
	if err != nil {
		logger.Logger.Error("api", "Could not find a disk with the provided uuid: ", _diskUUID, " in the db.")
		c.JSON(404, responses.NewNotFoundErrorResponse(constants.DATABASE_DISK_NOT_FOUND, "Cannot find a disk with the provided UUID"))
		return
	}

	// Retrieve userUUID from context
	userUUID = c.MustGet("UserData").(middleware.UserData).UserUUID

	// Verify that the user is owner of the disk
	if userUUID != _disk.UserUUID {
		c.JSON(404, responses.NewNotFoundErrorResponse(constants.OWNER_MISMATCH, "Disk not found"))
		return
	}

	// Compute free and total disk space
	volumeModel = models.Transport.GetVolume(_disk.VolumeUUID)
	if volumeModel == nil {
		logger.Logger.Error("api", "Could not find a volume with the provided uuid: ", _disk.VolumeUUID.String(), " in the db.")
		c.JSON(404, responses.NewNotFoundErrorResponse(constants.TRANSPORT_VOLUME_NOT_FOUND, "Cannot find a volume with the provided UUID"))
		return
	}

	diskModel = volumeModel.GetDisk(_disk.UUID)
	if diskModel == nil {
		logger.Logger.Error("api", "Could not find a disk with the provided uuid: ", _disk.UUID.String(), " in the provided volume.")
		c.JSON(404, responses.NewNotFoundErrorResponse(constants.TRANSPORT_DISK_NOT_FOUND, "Cannot find a disk with the provided UUID"))
		return
	}

	_disk.FreeSpace = models.ComputeFreeSpace(diskModel)
	_disk.TotalSpace = diskModel.GetTotalSpace()
	logger.Logger.Debug("api", "The disk capacity is: ", strconv.FormatUint(_disk.FreeSpace, 10), "/", strconv.FormatUint(_disk.TotalSpace, 10), ".")

	logger.Logger.Debug("api", "GetDisk endpoint successful exit.")
	c.JSON(200, responses.NewSuccessResponse(diskModel.GetResponse(&_disk, c)))
}

// UpdateDisk - handler for Update disk details request
//
// Update disk details (PUT /disks/manage/{diskUUID}) - updating the name or
// credentials of specified disk.
//
// params:
//   - c *gin.Context: context of the request
//
// return type:
//   - API response with appropriate HTTP code
func UpdateDisk(c *gin.Context) {
	var body requests.DiskUpdateRequest
	var _diskUUID string
	var diskUUID uuid.UUID
	var userUUID uuid.UUID
	var err error
	var volumes []*models.Volume
	var volume *models.Volume = nil
	var disk models.Disk = nil

	// Retrieve and validate data from request
	if err = c.ShouldBindJSON(&body); err != nil {
		logger.Logger.Error("api", "Wrong request body.")
		c.JSON(422, responses.NewValidationErrorResponse(err))
		return
	}

	// Retrieve userUUID from context
	userUUID = c.MustGet("UserData").(middleware.UserData).UserUUID

	// Parse diskUUID from request
	_diskUUID = c.Param("DiskUUID")
	diskUUID, err = uuid.Parse(_diskUUID)
	if err != nil {
		logger.Logger.Error("api", "Wrong disk UUID: ", _diskUUID)
		c.JSON(422, responses.NewValidationErrorResponseSingle(constants.VAL_UUID_INVALID, "DiskUUID", "Provided DiskUUID is not a valid UUID"))
		return
	}

	// Retrieve volumes from transport
	volumes = models.Transport.GetVolumes(userUUID)
	if volumes == nil {
		logger.Logger.Error("api", "Could not find any volume for the user with the provided uuid: ", userUUID.String())
		c.JSON(404, responses.NewNotFoundErrorResponse(constants.TRANSPORT_VOLUME_NOT_FOUND, "Cannot find volume with provided UUID"))
		return
	}

	// Find volume associated with provided disk
	for _, _volume := range volumes {
		disk = _volume.GetDisk(diskUUID)
		if disk != nil {
			volume = _volume
			break
		}
	}

	if disk == nil {
		logger.Logger.Error("api", "Could not find a disk with the provided uuid: ", _diskUUID)
		c.JSON(404, responses.NewNotFoundErrorResponse(constants.TRANSPORT_DISK_NOT_FOUND, "Cannot find volume with provided UUID"))
		return
	}
	logger.Logger.Debug("api", "Found the disk with the provided uuid: ", _diskUUID, ", it belongs to the volume: ", volume.UUID.String())

	// Check whether disk is enqueued for IO operation
	// Changes cannot be performed on busy disk.
	_disk := models.Transport.FindEnqueuedDisk(diskUUID)
	if _disk != nil {
		logger.Logger.Error("api", "The disk with the uuid: ", _diskUUID, " is being enqueued for upload / download and cannot be updated at the moment.")
		c.JSON(405, responses.NewOperationFailureResponse(constants.TRANSPORT_DISK_IS_BEING_USED, "Requested disk is enqueued for an IO operation, can't update it now"))
		return
	}

	// Convert credentials to JSON string
	cred := body.Credentials.ToString()
	if cred != "" {
		_, ok := disk.(OAuthDisk.OAuthDisk)
		if ok {
			logger.Logger.Error("api", "attempted to change the credentials of an oauth disk, which is not allowed.")
			c.JSON(405, responses.NewOperationFailureResponse(constants.TRANSPORT_DISK_IS_BEING_USED, "It is not allowed to change the credentials of an OAuth disk"))
			return
		}

		disk.CreateCredentials(cred)
		if !disk.GetReadiness().IsReadyForce(c) {
			logger.Logger.Error("api", "The provided credentials: ", body.Credentials.ToString(), " are invalid.")
			c.JSON(405, responses.NewOperationFailureResponse(constants.VAL_CREDENTIALS_INVALID, "Provided credentials are invalid"))
			return
		}
		logger.Logger.Debug("api", "Updated the credentials of the disk with the uuid: ", _diskUUID)
	}

	// Change name of the disk
	disk.SetName(body.Name)
	logger.Logger.Debug("api", "Updated the name of the disk with the uuid: ", _diskUUID, " to: ", body.Name, ".")

	// Verify that the disk space quota is valid
	_, totalSpace, errCode := disk.GetProviderSpace()
	if errCode == constants.SUCCESS {
		if totalSpace < body.TotalSpace {
			logger.Logger.Error("api", "The provided total space: ", strconv.FormatUint(body.TotalSpace, 10), " exeeds the disk space quota: ", strconv.FormatUint(body.TotalSpace, 10), ".")
			c.JSON(422, responses.NewValidationErrorResponseSingle(constants.VAL_QUOTA_EXCEEDED, "TotalSpace", "Provided total space exceeds the disk space quota"))
			return
		} else if disk.GetUsedSpace() > body.TotalSpace {
			logger.Logger.Error("api", "The provided total space: ", strconv.FormatUint(body.TotalSpace, 10), " is lower than the currently used space: ", strconv.FormatUint(disk.GetUsedSpace(), 10), ".")
			c.JSON(422, responses.NewValidationErrorResponseSingle(constants.VAL_QUOTA_EXCEEDED, "TotalSpace", "Provided total space is lower than currently used space"))
			return
		} else {
			logger.Logger.Debug("api", "Set the disk total space to: ", strconv.FormatUint(body.TotalSpace, 10), " (including space retrieved from provider).")
			disk.SetTotalSpace(body.TotalSpace)
		}
	} else if errCode == constants.OPERATION_NOT_SUPPORTED {
		logger.Logger.Debug("api", "Set the disk total space to: ", strconv.FormatUint(body.TotalSpace, 10), " (based on user input, provider doesn't support space information retrieval).")
		disk.SetTotalSpace(body.TotalSpace)
	}

	// Save disk details to database
	diskDBO := disk.GetDiskDBO(userUUID, disk.GetProviderUUID(), volume.UUID)
	result := db.DB.DatabaseHandle.Save(&diskDBO)
	if result.Error != nil {
		logger.Logger.Error("api", "Could not update the disk metadata in the db.")
		c.JSON(500, responses.NewOperationFailureResponse(constants.DATABASE_ERROR, "Database operation failed: "+result.Error.Error()))
		return
	}

	// Load full database object with a provider and a volume to return
	err = db.DB.DatabaseHandle.Where("uuid = ? AND is_virtual = ?", _diskUUID, false).Preload("Provider").Preload("Volume").Find(&diskDBO).Error
	if err != nil {
		logger.Logger.Error("api", "Could not validate the previous db operation.")
		c.JSON(500, responses.NewOperationFailureResponse(constants.DATABASE_DISK_NOT_FOUND, "Could not validate database change"))
		return
	}

	go volume.RefreshPartitioner()

	logger.Logger.Debug("api", "UpdateDisk endpoint successful exit.")
	c.JSON(200, responses.NewSuccessResponse(diskDBO))
}

// DeleteDisk - handler for Delete disk request
//
// Delete disk (DELETE /disks/manage/{diskUUID}) - deleting the specified disk.
//
// params:
//   - c *gin.Context: context of the request
//
// return type:
//   - API response with appropriate HTTP code
func DeleteDisk(c *gin.Context) {
	var _diskUUID string
	var _disk dbo.Disk
	var userUUID uuid.UUID
	var errCode string
	var err error

	// Retrieve disk from database
	_diskUUID = c.Param("DiskUUID")
	err = db.DB.DatabaseHandle.Where("uuid = ? AND is_virtual = ?", _diskUUID, false).Preload("Provider").Preload("Volume").Find(&_disk).Error
	if err != nil {
		logger.Logger.Error("api", "Could not find a disk with the provided uuid: ", _diskUUID, " in the db.")
		c.JSON(404, responses.NewNotFoundErrorResponse(constants.DATABASE_DISK_NOT_FOUND, "Could not find the disk with the provided UUID"))
		return
	}

	// Retrieve userUUID from context
	userUUID = c.MustGet("UserData").(middleware.UserData).UserUUID

	// Verify that the user is owner of the disk
	if userUUID != _disk.UserUUID {
		c.JSON(404, responses.NewNotFoundErrorResponse(constants.OWNER_MISMATCH, "Disk not found"))
		return
	}

	// Check whether disk is not enqueued for IO operation
	_d := models.Transport.FindEnqueuedDisk(_disk.UUID)
	if _d != nil {
		logger.Logger.Error("api", "The disk with the uuid: ", _diskUUID, " is being enqueued for upload / download and cannot be deleted at the moment.")
		c.JSON(405, responses.NewOperationFailureResponse(constants.TRANSPORT_DISK_IS_BEING_USED, "Requested disk is enqueued for an IO operation, can't delete it now"))
		return
	}

	// Retrieve volume from transport
	volume := models.Transport.GetVolume(_disk.VolumeUUID)
	if volume == nil {
		c.JSON(404, responses.NewNotFoundErrorResponse(constants.TRANSPORT_VOLUME_NOT_FOUND, "Volume not found"))
		return
	}

	// Trigger delete process
	if _disk.VirtualDiskUUID == uuid.Nil {
		// Delete actual disk since it's not connected to virtual disk
		newDisk := volume.FindAnotherDisk(_disk.UUID)
		errCode, err = models.Transport.DeleteDisk(volume.GetDisk(_disk.UUID), volume, constants.RELOCATION, newDisk)
	} else {
		// Delete virtual disk to which the actual disk is connected
		newDisk := volume.FindAnotherDisk(_disk.VirtualDiskUUID)
		errCode, err = models.Transport.DeleteDisk(volume.GetDisk(_disk.VirtualDiskUUID), volume, constants.RELOCATION, newDisk)
	}

	if errCode != constants.SUCCESS {
		c.JSON(500, responses.NewOperationFailureResponse(errCode, "Deletion of the disk failed: "+err.Error()))
		return
	}

	// Refresh volume partitioner after disk list change
	go volume.RefreshPartitioner()

	c.JSON(200, responses.NewSuccessResponse(_disk))
}

// GetDisks - handler for Get list of disks request
//
// Get list of disks (GET /disks/manage) - retrieving a paginated list of
// disks owned by a user.
//
// params:
//   - c *gin.Context: context of the request
//
// return type:
//   - API response with appropriate HTTP code
func GetDisks(c *gin.Context) {
	var userUUID uuid.UUID
	var _disks []dbo.Disk
	var disks []interface{}
	var page int

	// Retrieve page from query
	page = requests.GetPageFromQuery(c)

	// Retrieve userUUID from context
	userUUID = c.MustGet("UserData").(middleware.UserData).UserUUID

	// Load list of disks from database
	db.DB.DatabaseHandle.Where("user_uuid = ? AND virtual_disk_uuid = ?", userUUID.String(), uuid.Nil).Preload("Provider").Preload("Volume").Find(&_disks)
	for _, _disk := range _disks {
		// Update disk spaced based on local data (for performance reasons)
		_disk.FreeSpace = _disk.TotalSpace - _disk.UsedSpace
		volume := models.Transport.GetVolume(_disk.VolumeUUID)
		disk := volume.GetDisk(_disk.UUID)

		// Append disk to the list
		disks = append(disks, disk.GetResponse(&_disk, c))
	}

	// Prepare pagination
	pagination := models.Paginate(disks, page, constants.PAGINATION_RECORDS_PER_PAGE)

	logger.Logger.Debug("api", "GetDisks endpoint successful exit.")
	c.JSON(200, responses.NewPaginationResponse(responses.PaginationData{Pagination: pagination.Pagination, Data: pagination.Data}))
}
