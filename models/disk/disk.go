package disk

import (
	"dcfs/apicalls"
	"dcfs/db/dbo"
	"dcfs/models/credentials"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"golang.org/x/net/context"
	"golang.org/x/oauth2"
)

// for debug purposes
var RootUUID uuid.UUID

type Disk interface {
	Connect(c *gin.Context) error
	Upload(bm apicalls.BlockMetadata) error
	Download(bm apicalls.BlockMetadata) error
	Rename(c *gin.Context) error
	Remove(c *gin.Context) error

	SetUUID(uuid.UUID)
	GetUUID() uuid.UUID

	GetCredentials() credentials.Credentials
	SetCredentials(credentials.Credentials)
	CreateCredentials(credentials string)

	GetDiskDBO(userUUID uuid.UUID, providerUUID uuid.UUID, volumeUUID uuid.UUID) dbo.Disk
}

type OAuthDisk interface {
	Disk
	GetConfig() *oauth2.Config
}

type AbstractDisk struct {
	Disk
	UUID        uuid.UUID
	Credentials credentials.Credentials
}

func (d *AbstractDisk) Connect(ctx context.Context) error {
	panic("Unimplemented")
	return nil
}

func (d *AbstractDisk) SetUUID(UUID uuid.UUID) {
	d.UUID = UUID
}

func (d *AbstractDisk) GetUUID() uuid.UUID {
	return d.UUID
}

func (d *AbstractDisk) GetCredentials() credentials.Credentials {
	return d.Credentials
}

func (d *AbstractDisk) SetCredentials(c credentials.Credentials) {
	d.Credentials = c
}

func (d *AbstractDisk) GetDiskDBO(userUUID uuid.UUID, providerUUID uuid.UUID, volumeUUID uuid.UUID) dbo.Disk {
	credentials := ""
	if d.Credentials != nil {
		credentials = d.Credentials.ToString()
	}

	return dbo.Disk{
		AbstractDatabaseObject: dbo.AbstractDatabaseObject{UUID: d.UUID},
		UserUUID:               userUUID,
		ProviderUUID:           providerUUID,
		VolumeUUID:             volumeUUID,
		Credentials:            credentials,
	}
}
