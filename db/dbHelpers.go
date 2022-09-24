package db

import (
	"dcfs/constants"
	"dcfs/db/dbo"
	"github.com/google/uuid"
)

func UserFromDatabase(uuid uuid.UUID) (*dbo.User, string) {
	var user *dbo.User = dbo.NewUser()

	result := DB.DatabaseHandle.Where("uuid = ?", uuid).First(&user)
	if result.Error != nil {
		return nil, constants.DATABASE_USER_NOT_FOUND
	}

	return user, constants.SUCCESS
}
