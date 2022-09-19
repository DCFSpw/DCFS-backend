package validators

import (
	"dcfs/constants"
	"golang.org/x/crypto/bcrypt"
)

func ValidateUserPassword(hashedPassword string, providedPassword string) string {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(providedPassword))
	if err != nil {
		return constants.AUTH_INVALID_PASSWORD
	}
	return constants.SUCCESS
}
