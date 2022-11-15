package requests

import (
	"github.com/gin-gonic/gin"
	"strconv"
)

// GetPageFromQuery - retrieve page from query params.
//
// This function retrieves page number from query params (for pagination requests).
// If page number is not provided or invalid, default value is 1.
//
// params:
//   - c *gin.Context: pointer to current request context
//
// return type:
//   - int: pagination page
func GetPageFromQuery(c *gin.Context) int {
	// Retrieve page from query
	page, err := strconv.Atoi(c.DefaultQuery("page", "1"))

	// If page is not provided or invalid, set it to 1
	if err != nil || page < 1 {
		page = 1
	}

	return page
}
