package controllers

import (
	"dcfs/responses"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func CreateVolume(c *gin.Context) {
	var requestBody requests.VolumeCreateRequest
	var user *dbo.User
	var volume *dbo.Volume

	// Retrieve user account
	user, dbErr := db.UserFromDatabase(c.MustGet("UserData").(middleware.UserData).UserUUID)
	if dbErr != constants.SUCCESS {
		c.JSON(401, responses.InvalidCredentialsResponse{Success: false, Message: "Unauthorized", Code: constants.AUTH_UNAUTHORIZED})
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
		c.JSON(500, responses.OperationFailureResponse{Success: false, Message: "Database operation failed: " + result.Error.Error(), Code: constants.DATABASE_ERROR})
		return
	}

	// Return volume data
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
	// TO DO: Discuss if we should allow users to change the volume settings
	//volume.VolumeSettings.Backup = requestBody.Settings.Backup
	//volume.VolumeSettings.Encryption = requestBody.Settings.Encryption
	//volume.VolumeSettings.FilePartition = requestBody.Settings.FilePartition

	result := db.DB.DatabaseHandle.Save(&volume)
	if result.Error != nil {
		c.JSON(500, responses.OperationFailureResponse{Success: false, Message: "Database operation failed: " + result.Error.Error(), Code: constants.DATABASE_ERROR})
		return
	}

	// Return volume data
	c.JSON(200, responses.NewVolumeDataSuccessResponse(volume))
}

func DeleteVolume(c *gin.Context) {
	c.JSON(200, responses.SuccessResponse{Success: true, Message: "Delete Volume Endpoint"})
}

func GetVolumes(c *gin.Context) {
	c.JSON(200, responses.SuccessResponse{Success: true, Message: "Get Volumes Endpoint"})
}
