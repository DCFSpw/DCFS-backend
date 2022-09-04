package middleware

import (
	"dcfs/models/disk"
	"github.com/gin-gonic/gin"
)

type UserData struct {
	UserUUID string
}

func Authenticate() gin.HandlerFunc {
	return func(c *gin.Context) {
		// not implemented yet
		c.Set("UserData", UserData{UserUUID: disk.RootUUID.String()})
	}
}
