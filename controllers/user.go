package controllers

import (
	"dcfs/constants"
	"dcfs/db"
	"dcfs/db/dbo"
	"dcfs/requests"
	"dcfs/responses"
	"dcfs/validators"
	"github.com/gin-gonic/gin"
)

func RegisterUser(c *gin.Context) {
	var requestBody requests.RegisterUserRequest

	// Get data from request
	if err := c.BindJSON(&requestBody); err != nil {
		c.JSON(422, responses.ValidationErrorResponse{Success: false, Msg: "JSON validation failed: " + err.Error()})
		return
	}

	// Validate data
	errCode := validators.ValidateFirstName(requestBody.FirstName)
	if errCode != constants.SUCCESS {
		c.JSON(422, responses.ValidationErrorResponse{Success: false, Msg: "Validation error", Code: errCode})
		return
	}

	errCode = validators.ValidateLastName(requestBody.LastName)
	if errCode != constants.SUCCESS {
		c.JSON(422, responses.ValidationErrorResponse{Success: false, Msg: "Validation error", Code: errCode})
		return
	}

	errCode = validators.ValidateEmail(requestBody.Email)
	if errCode != constants.SUCCESS {
		c.JSON(422, responses.ValidationErrorResponse{Success: false, Msg: "Validation error", Code: errCode})
		return
	}

	errCode = validators.ValidatePassword(requestBody.Password)
	if errCode != constants.SUCCESS {
		c.JSON(422, responses.ValidationErrorResponse{Success: false, Msg: "Validation error", Code: errCode})
		return
	}

	// Create a new user
	var user *dbo.User = dbo.NewUserFromRequest(requestBody)

	// Save user to database
	result := db.DB.DatabaseHandle.Create(&user)
	if result.Error != nil {
		c.JSON(500, responses.OperationFailureResponse{Success: false, Msg: "Database operation failed: " + result.Error.Error(), Code: constants.DATABASE_ERROR})
		return
	}

	c.JSON(200, responses.NewRegisterUserSuccessResponse(*user))
}

func LoginUser(c *gin.Context) {

}
