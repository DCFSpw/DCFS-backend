package apicalls

import (
	"github.com/gin-gonic/gin"
	"golang.org/x/oauth2"
)

type CredentialsAuthenticateMetadata struct {
	Ctx    *gin.Context
	Config *oauth2.Config
}
