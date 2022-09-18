package validators

import (
	"dcfs/constants"
	"dcfs/db"
	"dcfs/db/dbo"
	"net/mail"
)

func ValidateEmail(email string) string {
	// Check whether field is empty
	if email == "" {
		return constants.VAL_MISSING_EMAIL
	}

	// Check whether field is too long
	if len(email) > constants.EMAIL_MAX_LENGTH {
		return constants.VAL_INVALID_EMAIL
	}

	// Check whether email is valid
	_, err := mail.ParseAddress(email)
	if err != nil {
		return constants.VAL_INVALID_EMAIL
	}

	// Check if email already exists in database
	result := db.DB.DatabaseHandle.Where("email = ?", email).First(&dbo.User{})
	if result.RowsAffected > 0 {
		return constants.VAL_EMAIL_ALREADY_EXISTS
	}

	// Return success
	return constants.SUCCESS
}

func ValidateFirstName(name string) string {
	// Check whether field is empty
	if name == "" {
		return constants.VAL_MISSING_FIRST_NAME
	}

	// Check whether field is too long
	if len(name) > constants.NAME_MAX_LENGTH {
		return constants.VAL_INVALID_FIRST_NAME
	}

	// Return success
	return constants.SUCCESS
}

func ValidateLastName(name string) string {
	// Check whether field is empty
	if name == "" {
		return constants.VAL_MISSING_LAST_NAME
	}

	// Check whether field is too long
	if len(name) > constants.NAME_MAX_LENGTH {
		return constants.VAL_INVALID_LAST_NAME
	}

	// Return success
	return constants.SUCCESS
}

func ValidatePassword(password string) string {
	// Check whether field is empty
	if password == "" {
		return constants.VAL_MISSING_PASSWORD
	}

	// Check whether field is too short or too long
	if len(password) < constants.PASSWORD_MIN_LENGTH || len(password) > constants.PASSWORD_MAX_LENGTH {
		return constants.VAL_INVALID_PASSWORD
	}

	// Return success
	return constants.SUCCESS
}
