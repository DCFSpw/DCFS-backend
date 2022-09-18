package dbo

import (
	"dcfs/requests"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type User struct {
	AbstractDatabaseObject
	FirstName string `gorm:"type:varchar(64)"`
	LastName  string `gorm:"type:varchar(64)"`
	Email     string `gorm:"type:varchar(128)"`
	Password  string `gorm:"type:varchar(64)"`
}

func NewUser() *User {
	var u *User = new(User)
	u.AbstractDatabaseObject.DatabaseObject = u
	u.UUID, _ = uuid.NewUUID()
	return u
}

func HashPassword(password string) string {
	bytes, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes)
}

func NewUserFromRequest(request *requests.RegisterUserRequest) *User {
	var u *User = new(User)
	u.AbstractDatabaseObject.DatabaseObject = u
	u.UUID, _ = uuid.NewUUID()
	u.FirstName = request.FirstName
	u.LastName = request.LastName
	u.Email = request.Email
	u.Password = HashPassword(request.Password)
	return u
}
