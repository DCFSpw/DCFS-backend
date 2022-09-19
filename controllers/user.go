package controllers

import (
	"dcfs/constants"
	"dcfs/db"
	"dcfs/db/dbo"
	"dcfs/middleware"
	"dcfs/requests"
	"dcfs/responses"
	"dcfs/validators"
	"github.com/gin-gonic/gin"
)

func RegisterUser(c *gin.Context) {
	var requestBody requests.RegisterUserRequest

	// Get data from request
	if err := c.BindJSON(&requestBody); err != nil {
		c.JSON(422, responses.ValidationErrorResponse{Success: false, Message: "JSON validation failed: " + err.Error(), Code: constants.REQ_JSON_BIND})
		return
	}

	// Validate data
	errCode := validators.ValidateFirstName(requestBody.FirstName)
	if errCode != constants.SUCCESS {
		c.JSON(422, responses.ValidationErrorResponse{Success: false, Message: "Validation error", Code: errCode})
		return
	}

	errCode = validators.ValidateLastName(requestBody.LastName)
	if errCode != constants.SUCCESS {
		c.JSON(422, responses.ValidationErrorResponse{Success: false, Message: "Validation error", Code: errCode})
		return
	}

	errCode = validators.ValidateEmail(requestBody.Email)
	if errCode != constants.SUCCESS {
		c.JSON(422, responses.ValidationErrorResponse{Success: false, Message: "Validation error", Code: errCode})
		return
	}

	errCode = validators.ValidatePassword(requestBody.Password)
	if errCode != constants.SUCCESS {
		c.JSON(422, responses.ValidationErrorResponse{Success: false, Message: "Validation error", Code: errCode})
		return
	}

	// Create a new user
	var user *dbo.User = dbo.NewUserFromRequest(&requestBody)

	// Save user to database
	result := db.DB.DatabaseHandle.Create(&user)
	if result.Error != nil {
		c.JSON(500, responses.OperationFailureResponse{Success: false, Message: "Database operation failed: " + result.Error.Error(), Code: constants.DATABASE_ERROR})
		return
	}

	c.JSON(200, responses.NewRegisterUserSuccessResponse(user))
}

func LoginUser(c *gin.Context) {
	var requestBody requests.RegisterUserRequest
	var user dbo.User

	// Get data from request
	if err := c.BindJSON(&requestBody); err != nil {
		c.JSON(422, responses.ValidationErrorResponse{Success: false, Message: "JSON validation failed: " + err.Error(), Code: constants.REQ_JSON_BIND})
		return
	}

	// Check if user exists
	result := db.DB.DatabaseHandle.Where("email = ?", requestBody.Email).First(&user)
	if result.Error != nil {
		c.JSON(401, responses.InvalidCredentialsResponse{Success: false, Message: "Invalid credentials", Code: constants.AUTH_INVALID_EMAIL})
		return
	}

	// Check if password is correct
	errCode := validators.ValidateUserPassword(user.Password, requestBody.Password)
	if errCode != constants.SUCCESS {
		c.JSON(401, responses.InvalidCredentialsResponse{Success: false, Message: "Invalid credentials", Code: constants.AUTH_INVALID_PASSWORD})
		return
	}

	// Generate JWT token
	signedToken, err := middleware.GenerateToken(user.UUID, user.Email)
	if err != nil {
		c.JSON(401, responses.InvalidCredentialsResponse{Success: false, Message: "Invalid credentials", Code: constants.AUTH_JWT_FAILURE})
		return
	}

	c.JSON(200, responses.NewLoginSuccessResponse(&user, signedToken))
}
