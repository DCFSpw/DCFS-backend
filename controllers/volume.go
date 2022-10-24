package controllers

import (
	"dcfs/responses"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func CreateVolume(c *gin.Context) {
	c.JSON(200, responses.SuccessResponse{Success: true, Message: "Create Volume Endpoint"})
}

func UpdateVolume(c *gin.Context) {
	c.JSON(200, responses.SuccessResponse{Success: true, Message: "Update Volume Endpoint"})
}

func DeleteVolume(c *gin.Context) {
	c.JSON(200, responses.SuccessResponse{Success: true, Message: "Delete Volume Endpoint"})
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

	// Return volume
	c.JSON(200, responses.NewVolumeDataSuccessResponse(volume))
}

func ShareVolume(c *gin.Context) {
	c.JSON(200, responses.SuccessResponse{Success: true, Message: "Share Volume Endpoint"})
}

func GetVolumes(c *gin.Context) {
	c.JSON(200, responses.SuccessResponse{Success: true, Message: "Get Volumes Endpoint"})
}
