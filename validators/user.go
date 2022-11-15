package validators

import (
	"dcfs/constants"
	"golang.org/x/crypto/bcrypt"
)

// ValidateUserPassword - validate provided password against the hashed password
//
// params:
//   - hashedPassword: string with user's password hashed by backend
//   - providedPassword: string password to verify
//
// return type:
//   - errorCode: (constant.SUCCESS if password match, constant.AUTH_INVALID_PASSWORD otherwise)
func ValidateUserPassword(hashedPassword string, providedPassword string) string {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(providedPassword))
	if err != nil {
		return constants.AUTH_INVALID_PASSWORD
	}
	return constants.SUCCESS
}
