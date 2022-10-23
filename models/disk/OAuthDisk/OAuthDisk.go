package OAuthDisk

import (
	"dcfs/models"
	"golang.org/x/oauth2"
)

type OAuthDisk interface {
	models.Disk
	GetConfig() *oauth2.Config
}
