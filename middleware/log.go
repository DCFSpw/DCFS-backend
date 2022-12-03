package middleware

import (
	"dcfs/util/logger"
	"github.com/gin-gonic/gin"
	"net/http/httputil"
	"strconv"
)

func LogApi() gin.HandlerFunc {
	return func(c *gin.Context) {
		req, err := httputil.DumpRequest(c.Request, true)
		if err != nil {
			logger.Logger.Warning("api", "Could not get the raw http request.")
		}
		logger.Logger.Debug("api", "Received a request: ", string(req[:]))

		c.Next()

		logger.Logger.Debug("api", "Returning the response code ", strconv.Itoa(c.Writer.Status()))
	}
}
