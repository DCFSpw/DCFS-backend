package controllers

import (
	"dcfs/constants"
	"dcfs/db"
	"dcfs/db/dbo"
	"dcfs/middleware"
	"dcfs/models"
	"dcfs/requests"
	"dcfs/responses"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"strconv"
)

func CreateVolume(c *gin.Context) {
	var requestBody requests.VolumeCreateRequest
	var user *dbo.User
	var volume *dbo.Volume

	// Retrieve user account
	user, dbErr := db.UserFromDatabase(c.MustGet("UserData").(middleware.UserData).UserUUID)
	if dbErr != constants.SUCCESS {
		c.JSON(401, responses.NewInvalidCredentialsResponse())
		return
	}

	// Retrieve and validate data from request
	if err := c.ShouldBindJSON(&requestBody); err != nil {
		c.JSON(422, responses.NewValidationErrorResponse(err))
		return
	}

	// Create a new volume
	volume = dbo.NewVolumeFromRequest(&requestBody, user.UUID)

	// Save user to database
	result := db.DB.DatabaseHandle.Create(&volume)
	if result.Error != nil {
		c.JSON(500, responses.NewOperationFailureResponse(constants.DATABASE_ERROR, "Database operation failed: "+result.Error.Error()))
		return
	}

	c.JSON(200, responses.NewVolumeDataSuccessResponse(volume))
}

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
		c.JSON(404, responses.NewNotFoundErrorResponse(dbErr, "Volume not found"))
		return
	}

	// Verify that the user is owner of the volume
	if userUUID != volume.UserUUID {
		c.JSON(404, responses.NewNotFoundErrorResponse(constants.OWNER_MISMATCH, "Volume not found"))
		return
	}

	// Return volume data
	c.JSON(200, responses.NewVolumeDataSuccessResponse(volume))
}

func UpdateVolume(c *gin.Context) {
	var requestBody requests.VolumeCreateRequest
	var volume *dbo.Volume
	var volumeUUID string
	var userUUID uuid.UUID

	// Retrieve and validate data from request
	if err := c.ShouldBindJSON(&requestBody); err != nil {
		c.JSON(422, responses.NewValidationErrorResponse(err))
		return
	}

	// Retrieve volumeUUID from path parameters
	volumeUUID = c.Param("VolumeUUID")

	// Retrieve userUUID from context
	userUUID = c.MustGet("UserData").(middleware.UserData).UserUUID

	// Retrieve volume from database
	volume, dbErr := db.VolumeFromDatabase(volumeUUID)
	if dbErr != constants.SUCCESS {
		c.JSON(404, responses.NewNotFoundErrorResponse(dbErr, "Volume not found"))
		return
	}

	// Verify that the user is owner of the volume
	if userUUID != volume.UserUUID {
		c.JSON(404, responses.NewNotFoundErrorResponse(constants.OWNER_MISMATCH, "Volume not found"))
		return
	}

	// Update volume data
	volume.Name = requestBody.Name
	volume.VolumeSettings.FilePartition = requestBody.Settings.FilePartition

	// Update options for empty volume
	empty, err := db.IsVolumeEmpty(volume.UUID)
	if empty && err == nil {
		volume.VolumeSettings.Backup = requestBody.Settings.Backup
		volume.VolumeSettings.Encryption = requestBody.Settings.Encryption
	}

	// Save volume to database
	result := db.DB.DatabaseHandle.Save(&volume)
	if result.Error != nil {
		c.JSON(500, responses.NewOperationFailureResponse(constants.DATABASE_ERROR, "Database operation failed: "+result.Error.Error()))
		return
	}

	// Return volume data
	c.JSON(200, responses.NewVolumeDataSuccessResponse(volume))
}

func DeleteVolume(c *gin.Context) {
	var volume *dbo.Volume
	var volumeModel *models.Volume
	var volumeUUID string
	var userUUID uuid.UUID
	var errCode string
	var err error

	// Retrieve volumeUUID from path parameters
	volumeUUID = c.Param("VolumeUUID")

	// Retrieve userUUID from context
	userUUID = c.MustGet("UserData").(middleware.UserData).UserUUID

	// Retrieve volume from database
	volume, dbErr := db.VolumeFromDatabase(volumeUUID)
	if dbErr != constants.SUCCESS {
		c.JSON(404, responses.NewNotFoundErrorResponse(dbErr, "Volume not found"))
		return
	}

	// Verify that the user is owner of the volume
	if userUUID != volume.UserUUID {
		c.JSON(404, responses.NewNotFoundErrorResponse(constants.OWNER_MISMATCH, "Volume not found"))
		return
	}

	// Trigger delete process
	volumeModel = models.Transport.GetVolume(userUUID, uuid.MustParse(volumeUUID))
	if volumeModel == nil {
		c.JSON(404, responses.NewNotFoundErrorResponse(constants.TRANSPORT_VOLUME_NOT_FOUND, "Cannot find volume with provided UUID"))
		return
	}

	errCode, err = volumeModel.Delete()
	if err != nil {
		c.JSON(500, responses.NewOperationFailureResponse(errCode, "Deletion request failed: "+err.Error()))
		return
	}

	// Delete volume from database
	result := db.DB.DatabaseHandle.Delete(&volume)
	if result.Error != nil {
		c.JSON(500, responses.NewOperationFailureResponse(constants.DATABASE_ERROR, "Database operation failed: "+result.Error.Error()))
		return
	}

	// Return volume data
	c.JSON(200, responses.NewEmptySuccessResponse())
}

func GetVolumes(c *gin.Context) {
	var volumes []dbo.Volume
	var volumesPagination []interface{}
	var userUUID uuid.UUID
	var page int
	var err error

	// Retrieve page from query
	page, err = strconv.Atoi(c.DefaultQuery("page", "1"))
	if err != nil {
		page = 1
	}

	// Retrieve userUUID from context
	userUUID = c.MustGet("UserData").(middleware.UserData).UserUUID

	// Retrieve list of volumes of current user from the database
	err = db.DB.DatabaseHandle.Where("user_uuid = ?", userUUID).Find(&volumes).Error
	if err != nil {
		c.JSON(500, responses.NewOperationFailureResponse(constants.DATABASE_ERROR, "Database operation failed: "+err.Error()))
		return
	}

	// Prepare pagination list
	for _, volume := range volumes {
		volumesPagination = append(volumesPagination, volume)
	}
	pagination := models.Paginate(volumesPagination, page, constants.PAGINATION_RECORDS_PER_PAGE)
	if pagination == nil {
		c.JSON(500, responses.NewOperationFailureResponse(constants.INT_PAGINATION_ERROR, "Pagination process failed."))
		return
	}

	// Return list of volumes
	c.JSON(200, responses.NewPaginationResponse(responses.PaginationData{Pagination: pagination.Pagination, Data: pagination.Data}))
}
