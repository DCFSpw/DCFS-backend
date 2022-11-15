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
	var user *dbo.User

	// Retrieve and validate data from request
	if err := c.ShouldBindJSON(&requestBody); err != nil {
		c.JSON(422, responses.NewValidationErrorResponse(err))
		return
	}

	// Check if e-mail exists
	result := db.DB.DatabaseHandle.Where("email = ?", requestBody.Email).First(&dbo.User{})
	if result.RowsAffected > 0 {
		c.JSON(422, responses.FailureResponse{Success: false, Message: "Specified e-mail already exists.", Code: constants.VAL_EMAIL_ALREADY_EXISTS})
		return
	}

	// Create a new user
	user = dbo.NewUserFromRequest(&requestBody)

	// Save user to database
	result = db.DB.DatabaseHandle.Create(&user)
	if result.Error != nil {
		c.JSON(500, responses.OperationFailureResponse{Success: false, Message: "Database operation failed: " + result.Error.Error(), Code: constants.DATABASE_ERROR})
		return
	}

	c.JSON(200, responses.NewUserDataSuccessResponse(user))
}

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
		c.JSON(500, responses.OperationFailureResponse{Success: false, Message: "Database operation failed: " + result.Error.Error(), Code: constants.DATABASE_ERROR})
		return
	}

	// Return updated user profile
	c.JSON(200, responses.NewUserDataSuccessResponse(user))
}

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
		c.JSON(500, responses.OperationFailureResponse{Success: false, Message: "Database operation failed: " + result.Error.Error(), Code: constants.DATABASE_ERROR})
		return
	}

	c.JSON(200, responses.NewEmptySuccessResponse())
}
