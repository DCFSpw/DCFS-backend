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

// RegisterUser - handler for Register as user request
//
// Register as user (POST /auth/register) - registering new user account.
//
// params:
//   - c *gin.Context: context of the request
//
// return type:
//   - API response with appropriate HTTP code
func RegisterUser(c *gin.Context) {
	var requestBody requests.RegisterUserRequest
	var user *dbo.User

	// Retrieve and validate data from request
	if err := c.ShouldBindJSON(&requestBody); err != nil {
		c.JSON(422, responses.NewValidationErrorResponse(err))
		return
	}

	// Check if e-mail exists
	result := db.DB.DatabaseHandle.Where("email = ?", requestBody.Email).First(&dbo.User{})
	if result.RowsAffected > 0 {
		c.JSON(422, responses.NewValidationErrorResponseSingle(constants.VAL_EMAIL_ALREADY_EXISTS, "email", "Specified e-mail already exists."))
		return
	}

	// Create a new user
	user = dbo.NewUserFromRequest(&requestBody)

	// Save user to database
	result = db.DB.DatabaseHandle.Create(&user)
	if result.Error != nil {
		c.JSON(500, responses.NewOperationFailureResponse(constants.DATABASE_ERROR, "Database operation failed: "+result.Error.Error()))
		return
	}

	c.JSON(200, responses.NewUserDataSuccessResponse(user))
}

// LoginUser - handler for Register as user request
//
// Login as user (POST /auth/register) - logging in using account credentials
// and obtaining a Bearer token required by all authorized requests
//
// params:
//   - c *gin.Context: context of the request
//
// return type:
//   - API response with appropriate HTTP code
func LoginUser(c *gin.Context) {
	var requestBody requests.LoginUserRequest
	var user dbo.User

	// Retrieve and validate data from request
	if err := c.ShouldBindJSON(&requestBody); err != nil {
		c.JSON(422, responses.NewValidationErrorResponse(err))
		return
	}

	// Check if user exists
	result := db.DB.DatabaseHandle.Where("email = ?", requestBody.Email).First(&user)
	if result.Error != nil {
		c.JSON(401, responses.NewOperationFailureResponse(constants.AUTH_INVALID_EMAIL, "Unauthorized"))
		return
	}

	// Check if password is correct
	errCode := validators.ValidateUserPassword(user.Password, requestBody.Password)
	if errCode != constants.SUCCESS {
		c.JSON(401, responses.NewInvalidCredentialsResponse())
		return
	}

	// Generate JWT token
	signedToken, err := middleware.GenerateToken(user.UUID, user.Email)
	if err != nil {
		c.JSON(401, responses.NewOperationFailureResponse(constants.AUTH_JWT_FAILURE, "Unauthorized"))
		return
	}

	c.JSON(200, responses.NewLoginSuccessResponse(&user, signedToken))
}

// GetUserProfile - handler for Get user profile request
//
// Get user profile (GET /user/profile) - retrieving account information
// of a user.
//
// params:
//   - c *gin.Context: context of the request
//
// return type:
//   - API response with appropriate HTTP code
func GetUserProfile(c *gin.Context) {
	var user *dbo.User

	// Retrieve user account
	user, dbErr := db.UserFromDatabase(c.MustGet("UserData").(middleware.UserData).UserUUID)
	if dbErr != constants.SUCCESS {
		c.JSON(401, responses.NewInvalidCredentialsResponse())
		return
	}

	// Return user profile
	c.JSON(200, responses.NewUserDataSuccessResponse(user))
}

// UpdateUserProfile - handler for Update user profile request
//
// Update user profile (PUT /user/profile) - updating account information
// of a user.
//
// params:
//   - c *gin.Context: context of the request
//
// return type:
//   - API response with appropriate HTTP code
func UpdateUserProfile(c *gin.Context) {
	var requestBody requests.UpdateUserProfileRequest
	var user *dbo.User

	// Retrieve and validate data from request
	if err := c.ShouldBindJSON(&requestBody); err != nil {
		c.JSON(422, responses.NewValidationErrorResponse(err))
		return
	}

	// Retrieve user account
	user, dbErr := db.UserFromDatabase(c.MustGet("UserData").(middleware.UserData).UserUUID)
	if dbErr != constants.SUCCESS {
		c.JSON(401, responses.NewInvalidCredentialsResponse())
		return
	}

	// Update user profile
	user.FirstName = requestBody.FirstName
	user.LastName = requestBody.LastName

	result := db.DB.DatabaseHandle.Save(&user)
	if result.Error != nil {
		c.JSON(500, responses.NewOperationFailureResponse(constants.DATABASE_ERROR, "Database operation failed: "+result.Error.Error()))
		return
	}

	// Return updated user profile
	c.JSON(200, responses.NewUserDataSuccessResponse(user))
}

// ChangeUserPassword - handler for Change user password request
//
// Change user password (PUT /user/password) - changing account password
// of a user.
//
// params:
//   - c *gin.Context: context of the request
//
// return type:
//   - API response with appropriate HTTP code
func ChangeUserPassword(c *gin.Context) {
	var requestBody requests.ChangeUserPasswordRequest
	var user *dbo.User

	// Retrieve and validate data from request
	if err := c.ShouldBindJSON(&requestBody); err != nil {
		c.JSON(422, responses.NewValidationErrorResponse(err))
		return
	}

	// Retrieve user account
	user, dbErr := db.UserFromDatabase(c.MustGet("UserData").(middleware.UserData).UserUUID)
	if dbErr != constants.SUCCESS {
		c.JSON(401, responses.NewInvalidCredentialsResponse())
		return
	}

	// Check if password is correct
	errCode := validators.ValidateUserPassword(user.Password, requestBody.OldPassword)
	if errCode != constants.SUCCESS {
		c.JSON(422, responses.NewValidationErrorResponseSingle(constants.AUTH_INVALID_OLD_PASSWORD, "OldPassword", "Old password is incorrect"))
		return
	}

	// Change password
	user.Password = dbo.HashPassword(requestBody.NewPassword)

	result := db.DB.DatabaseHandle.Save(&user)
	if result.Error != nil {
		c.JSON(500, responses.NewOperationFailureResponse(constants.DATABASE_ERROR, "Database operation failed: "+result.Error.Error()))
		return
	}

	c.JSON(200, responses.NewEmptySuccessResponse())
}
