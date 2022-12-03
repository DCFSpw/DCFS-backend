package controllers

import (
	"dcfs/constants"
	"dcfs/db"
	"dcfs/db/dbo"
	"dcfs/responses"
	"dcfs/util/logger"
	"github.com/gin-gonic/gin"
)

// GetProviders - handler for Create disk request
//
// Get list of providers (GET /providers - retrieving a list of
// available disk providers.
//
// params:
//   - c *gin.Context: context of the request
//
// return type:
//   - API response with appropriate HTTP code
func GetProviders(c *gin.Context) {
	var providers []dbo.Provider

	// Retrieve list of providers from the database
	err := db.DB.DatabaseHandle.Find(&providers).Error
	if err != nil {
		logger.Logger.Error("api", "Could not retrieve a list of providers from the db.")
		c.JSON(500, responses.NewOperationFailureResponse(constants.DATABASE_ERROR, "Database operation failed: "+err.Error()))
		return
	}

	// Return list of providers
	logger.Logger.Debug("api", "GetProviders endpoint successful exit.")
	c.JSON(200, responses.NewGetProvidersSuccessResponse(providers))
}
