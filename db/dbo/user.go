package dbo

type User struct {
	AbstractDatabaseObject
	FirstName string `gorm:"type:varchar(64)"`
	LastName  string `gorm:"type:varchar(64)"`
	Email     string `gorm:"type:varchar(128)"`
	Password  string `gorm:"type:varchar(32)"`
}

func NewUser() *User {
	var u *User = new(User)
	u.AbstractDatabaseObject.DatabaseObject = u
	return u
}
