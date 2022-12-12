package controllers

import (
	"dcfs/constants"
	"dcfs/db"
	"dcfs/db/dbo"
	"dcfs/middleware"
	"dcfs/models"
	"dcfs/requests"
	"dcfs/responses"
	"dcfs/util/logger"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"strconv"
)

// CreateVolume - handler for Create volume request
//
// Create volume (POST /volumes/manage) - creating a new volume.
//
// params:
//   - c *gin.Context: context of the request
//
// return type:
//   - API response with appropriate HTTP code
func CreateVolume(c *gin.Context) {
	var requestBody requests.VolumeCreateRequest
	var user *dbo.User
	var volume *dbo.Volume

	// Retrieve user account
	user, dbErr := db.UserFromDatabase(c.MustGet("UserData").(middleware.UserData).UserUUID)
	if dbErr != constants.SUCCESS {
		logger.Logger.Error("api", "Could not find the user in the db.")
		c.JSON(401, responses.NewInvalidCredentialsResponse())
		return
	}

	// Retrieve and validate data from request
	if err := c.ShouldBindJSON(&requestBody); err != nil {
		logger.Logger.Error("api", "Wrong request body.")
		c.JSON(422, responses.NewValidationErrorResponse(err))
		return
	}

	// Create a new volume
	volume = dbo.NewVolumeFromRequest(&requestBody, user.UUID)

	// Save volume to database
	result := db.DB.DatabaseHandle.Create(&volume)
	if result.Error != nil {
		logger.Logger.Error("api", "Could not save the newly created volume in the db. Got err: ", result.Error.Error())
		c.JSON(500, responses.NewOperationFailureResponse(constants.DATABASE_ERROR, "Database operation failed: "+result.Error.Error()))
		return
	}

	// Initiate volume in transport
	_ = models.Transport.GetVolume(volume.UUID)

	logger.Logger.Debug("api", "CreateVolume endpoint successful exit.")
	c.JSON(200, responses.NewVolumeDataSuccessResponse(volume))
}

// GetVolume - handler for Get volume details request
//
// Get volume details (GET /volumes/manage/{volumeUUID}) - retrieving details
// of the specified volume.
//
// params:
//   - c *gin.Context: context of the request
//
// return type:
//   - API response with appropriate HTTP code
func GetVolume(c *gin.Context) {
	var volume *dbo.Volume
	var volumeUUID string
	var userUUID uuid.UUID

	// Retrieve volumeUUID from path parameters
	volumeUUID = c.Param("VolumeUUID")

	// Retrieve userUUID from context
	userUUID = c.MustGet("UserData").(middleware.UserData).UserUUID

	// Retrieve volume from database
	volume, dbErr := db.VolumeFromDatabase(volumeUUID)
	if dbErr != constants.SUCCESS {
		logger.Logger.Error("api", "A volume with the provided uuid: ", volumeUUID, " was not found in the db.")
		c.JSON(404, responses.NewNotFoundErrorResponse(dbErr, "Volume not found"))
		return
	}

	// Verify that the user is owner of the volume
	if userUUID != volume.UserUUID {
		logger.Logger.Error("api", "The user: ", userUUID.String(), " is not the owner of the volume: ", volumeUUID)
		c.JSON(404, responses.NewNotFoundErrorResponse(constants.OWNER_MISMATCH, "Volume not found"))
		return
	}

	v := models.Transport.GetVolume(volume.UUID)
	if v == nil {
		logger.Logger.Error("api", "A volume with the provided uuid: ", volumeUUID, " was not found in the db.")
		c.JSON(404, responses.NewNotFoundErrorResponse(dbErr, "Volume not found"))
	}

	// Return volume data
	logger.Logger.Debug("api", "GetVolume endpoint successful exit.")
	c.JSON(200, responses.NewVolumeListSuccessResponse(&responses.VolumeResponse{
		Volume:  *volume,
		IsReady: v.IsReady(c, false),
	}))
}

// UpdateVolume - handler for Update volume details request
//
// Update volume details (PUT /volumes/manage/{volumeUUID}) - updating the name
// or settings (such as backup, partition and encryption modes) of the specified
// volume.
//
// params:
//   - c *gin.Context: context of the request
//
// return type:
//   - API response with appropriate HTTP code
func UpdateVolume(c *gin.Context) {
	var requestBody requests.VolumeCreateRequest
	var volume *models.Volume
	var volumeDBO dbo.Volume
	var volumeUUID uuid.UUID
	var userUUID uuid.UUID
	var err error

	// Retrieve and validate data from request
	if err := c.ShouldBindJSON(&requestBody); err != nil {
		logger.Logger.Error("api", "Wrong request body.")
		c.JSON(422, responses.NewValidationErrorResponse(err))
		return
	}

	// Retrieve volumeUUID from path parameters
	volumeUUID, err = uuid.Parse(c.Param("VolumeUUID"))
	if err != nil {
		logger.Logger.Error("api", "Wrong volume uuid.")
		c.JSON(404, responses.NewNotFoundErrorResponse(constants.VAL_UUID_INVALID, "Volume not found (invalid UUID)"))
		return
	}

	// Retrieve userUUID from context
	userUUID = c.MustGet("UserData").(middleware.UserData).UserUUID

	// Retrieve volume from transport
	volume = models.Transport.GetVolume(volumeUUID)
	if volume == nil {
		logger.Logger.Error("api", "A volume with the provided uuid: ", volumeUUID.String(), " was not found.")
		c.JSON(404, responses.NewNotFoundErrorResponse(constants.TRANSPORT_VOLUME_NOT_FOUND, "Volume not found"))
		return
	}

	// Verify that the user is owner of the volume
	if userUUID != volume.UserUUID {
		logger.Logger.Error("The user: ", userUUID.String(), " is not the owner of the volume: ", volumeUUID.String())
		c.JSON(404, responses.NewNotFoundErrorResponse(constants.OWNER_MISMATCH, "Volume not found"))
		return
	}

	// Update volume data
	volume.Name = requestBody.Name
	volume.VolumeSettings.FilePartition = requestBody.Settings.FilePartition
	logger.Logger.Debug("api", "Updated name to: ", requestBody.Name, ",  partitioning settings to: ", strconv.Itoa(requestBody.Settings.FilePartition), " of the volume: ", volumeUUID.String(), ".")

	// Update options for empty volume
	empty, err := db.IsVolumeEmpty(volume.UUID)
	if empty && err == nil {
		volume.VolumeSettings.Encryption = requestBody.Settings.Encryption

		logger.Logger.Debug("api", "Updated encryption to: ", strconv.Itoa(requestBody.Settings.Encryption), " of the volume: ", volumeUUID.String(), ".")
	}

	// Save volume to database
	volumeDBO = volume.GetVolumeDBO()

	result := db.DB.DatabaseHandle.Save(&volumeDBO)
	if result.Error != nil {
		logger.Logger.Error("api", "Could not update the volume data in the db.")
		c.JSON(500, responses.NewOperationFailureResponse(constants.DATABASE_ERROR, "Database operation failed: "+result.Error.Error()))
		return
	}

	// Invalidate volume in transport
	models.Transport.ActiveVolumes.RemoveEnqueuedInstance(volumeUUID)

	// Return volume data
	logger.Logger.Debug("api", "UpdateVolume endpoint successful exit.")
	c.JSON(200, responses.NewVolumeDataSuccessResponse(&volumeDBO))
}

// DeleteVolume - handler for Delete volume request
//
// Delete volume (DELETE /volumes/manage/{volumeUUID}) - deleting the specified
// volume (and all associated disks and files).
//
// params:
//   - c *gin.Context: context of the request
//
// return type:
//   - API response with appropriate HTTP code
func DeleteVolume(c *gin.Context) {
	var volume *models.Volume
	var volumeDBO dbo.Volume
	var volumeUUID uuid.UUID
	var userUUID uuid.UUID
	var errCode string
	var err error

	// Retrieve volumeUUID from path parameters
	volumeUUID, err = uuid.Parse(c.Param("VolumeUUID"))
	if err != nil {
		logger.Logger.Error("api", "Wrong volume uuid.")
		c.JSON(404, responses.NewNotFoundErrorResponse(constants.VAL_UUID_INVALID, "Volume not found (invalid UUID)"))
		return
	}

	// Retrieve userUUID from context
	userUUID = c.MustGet("UserData").(middleware.UserData).UserUUID

	// Retrieve volume from transport
	volume = models.Transport.GetVolume(volumeUUID)
	if volume == nil {
		logger.Logger.Error("api", "A volume with the provided uuid: ", volumeUUID.String(), " was not found.")
		c.JSON(404, responses.NewNotFoundErrorResponse(constants.TRANSPORT_VOLUME_NOT_FOUND, "Volume not found"))
		return
	}

	// Verify that the user is owner of the volume
	if userUUID != volume.UserUUID {
		logger.Logger.Error("api", "The user: ", userUUID.String(), " is not the owner of the volume: ", volume.UUID.String(), ".")
		c.JSON(404, responses.NewNotFoundErrorResponse(constants.OWNER_MISMATCH, "Volume not found"))
		return
	}

	// Check if volume is enqueued for upload or download
	// We do not allow deleting volumes that are in use
	volumeFromQueue := models.Transport.FindEnqueuedVolume(volumeUUID)
	if volumeFromQueue != nil {
		logger.Logger.Error("api", "The volume: ", volumeUUID.String(), " is currently queued for upload / download and cannot be deleted.")
		c.JSON(409, responses.NewOperationFailureResponse(constants.TRANSPORT_VOLUME_IS_BEING_USED, "Volume is currently enqueued for upload or download"))
		return
	}

	// Trigger delete process
	errCode, err = models.Transport.DeleteVolume(volumeUUID)
	if err != nil {
		logger.Logger.Error("api", "Could not delete the volume: ", volumeUUID.String())
		c.JSON(500, responses.NewOperationFailureResponse(errCode, "Volume deletion request failed: "+err.Error()))
		return
	}

	// Delete volume from database
	volumeDBO = volume.GetVolumeDBO()

	result := db.DB.DatabaseHandle.Delete(&volumeDBO)
	if result.Error != nil {
		logger.Logger.Error("Could not delete the volume: ", volumeUUID.String(), " from the db.")
		c.JSON(500, responses.NewOperationFailureResponse(constants.DATABASE_ERROR, "Database operation failed: "+result.Error.Error()))
		return
	}

	// Return volume data
	logger.Logger.Debug("api", "DeleteVolume endpoint successful exit.")
	c.JSON(200, responses.NewEmptySuccessResponse())
}

// GetVolumes - handler for Get list of volumes request
//
// Get list of volumes (GET /volumes/manage) - retrieving a paginated list of
// volumes owned by a user.
//
// params:
//   - c *gin.Context: context of the request
//
// return type:
//   - API response with appropriate HTTP code
func GetVolumes(c *gin.Context) {
	var _volumes []dbo.Volume
	var volumesPagination []interface{}
	var userUUID uuid.UUID
	var page int
	var err error

	// Retrieve page from query
	page = requests.GetPageFromQuery(c)

	// Retrieve userUUID from context
	userUUID = c.MustGet("UserData").(middleware.UserData).UserUUID

	// Retrieve list of volumes of current user from the database
	err = db.DB.DatabaseHandle.Where("user_uuid = ?", userUUID).Find(&_volumes).Error
	if err != nil {
		logger.Logger.Error("api", "Could not retrieve a list of volumes from the db.")
		c.JSON(500, responses.NewOperationFailureResponse(constants.DATABASE_ERROR, "Database operation failed: "+err.Error()))
		return
	}

	// Prepare pagination list
	for _, _v := range _volumes {
		v := models.Transport.GetVolume(_v.UUID)
		if v == nil {
			logger.Logger.Error("api", "A volume with the provided uuid: ", v.UUID.String(), " was not found in the db.")
			continue
		}

		volumesPagination = append(volumesPagination, responses.VolumeResponse{
			Volume:  _v,
			IsReady: v.IsReady(c, false),
		})
	}

	pagination := models.Paginate(volumesPagination, page, constants.PAGINATION_RECORDS_PER_PAGE)
	if pagination == nil {
		logger.Logger.Error("api", "Could not paginate the provided list of volumes.")
		c.JSON(500, responses.NewOperationFailureResponse(constants.INT_PAGINATION_ERROR, "Pagination process failed."))
		return
	}

	// Return list of volumes
	logger.Logger.Debug("api", "GetVolumes endpoint successful exit.")
	c.JSON(200, responses.NewPaginationResponse(responses.PaginationData{Pagination: pagination.Pagination, Data: pagination.Data}))
}
