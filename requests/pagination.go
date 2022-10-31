package requests

import (
	"github.com/gin-gonic/gin"
	"strconv"
)

func GetPageFromQuery(c *gin.Context) int {
	// Retrieve page from query
	page, err := strconv.Atoi(c.DefaultQuery("page", "1"))

	// If page is not provided or invalid, set it to 1
	if err != nil || page < 1 {
		page = 1
	}

	return page
}
