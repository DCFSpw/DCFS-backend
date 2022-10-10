package controllers

import (
	"dcfs/constants"
	"dcfs/db"
	"dcfs/db/dbo"
	"dcfs/responses"
	"github.com/gin-gonic/gin"
)

func GetProviders(c *gin.Context) {
	var providers []dbo.Provider

	// Retrieve list of providers from the database
	err := db.DB.DatabaseHandle.Find(&providers).Error
	if err != nil {
		c.JSON(500, responses.OperationFailureResponse{Success: false, Message: "Database operation failed: " + err.Error(), Code: constants.DATABASE_ERROR})
		return
	}

	// Return list of providers
	c.JSON(200, responses.NewGetProvidersSuccessResponse(providers))
}
