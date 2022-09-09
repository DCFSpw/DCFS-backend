package GDriveDisk

import (
	"dcfs/apicalls"
	"dcfs/db/dbo"
	"dcfs/models/credentials"
	"dcfs/models/disk"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/drive/v3"
	"log"
	"os"
)

type GDriveDisk struct {
	abstractDisk disk.AbstractDisk
}

func (d *GDriveDisk) Connect(c *gin.Context) error {
	return nil
}

func (d *GDriveDisk) Upload(bm *apicalls.BlockMetadata) error {
	return nil
}

func (d *GDriveDisk) Download(c *gin.Context) error {
	return nil
}

func (d *GDriveDisk) Rename(c *gin.Context) error {
	return nil
}

func (d *GDriveDisk) Remove(c *gin.Context) error {
	return nil
}

func (d *GDriveDisk) SetUUID(uuid uuid.UUID) {
	d.abstractDisk.SetUUID(uuid)
}

func (d *GDriveDisk) GetUUID() uuid.UUID {
	return d.abstractDisk.GetUUID()
}

func (d *GDriveDisk) GetCredentials() credentials.Credentials {
	return d.abstractDisk.GetCredentials()
}

func (d *GDriveDisk) SetCredentials(credentials credentials.Credentials) {
	d.abstractDisk.SetCredentials(credentials)
}

func (d *GDriveDisk) CreateCredentials(c string) {
	d.abstractDisk.Credentials = credentials.NewOauthCredentials(c)
}

func (d *GDriveDisk) GetDiskDBO(userUUID uuid.UUID, providerUUID uuid.UUID, volumeUUID uuid.UUID) dbo.Disk {
	return d.abstractDisk.GetDiskDBO(userUUID, providerUUID, volumeUUID)
}

func NewGDriveDisk() *GDriveDisk {
	var d *GDriveDisk = new(GDriveDisk)
	d.abstractDisk.Disk = d
	return d
}

func (d *GDriveDisk) GetConfig() *oauth2.Config {
	b, err := os.ReadFile("./models/disk/GDriveDisk/credentials.json")
	if err != nil {
		log.Fatalf("Unable to read client secret file: %v", err)
	}

	// If modifying these scopes, delete your previously saved token.json.
	config, err := google.ConfigFromJSON(b, drive.DriveMetadataReadonlyScope)
	if err != nil {
		log.Fatalf("Unable to parse client secret file to config: %v", err)
	}

	return config
}
