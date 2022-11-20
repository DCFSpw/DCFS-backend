package mock

import (
	"dcfs/db/dbo"
	"github.com/google/uuid"
)

var UserUUID uuid.UUID = uuid.New()

var UserDBO *dbo.User = &dbo.User{
	AbstractDatabaseObject: dbo.AbstractDatabaseObject{
		UUID: UserUUID,
	},
	FirstName: "Mock",
	LastName:  "User",
	Email:     "mock.user@fake.com",
	Password:  "password",
}
