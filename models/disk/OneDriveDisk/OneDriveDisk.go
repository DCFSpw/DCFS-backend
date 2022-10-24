package OneDriveDisk

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

type OneDriveDisk struct {
	abstractDisk disk.AbstractDisk
}

func (d *OneDriveDisk) Connect(c *gin.Context) error {
	return nil
}

func (d *OneDriveDisk) Upload(bm *apicalls.BlockMetadata) error {
	return nil
}

func (d *OneDriveDisk) Download(bm *apicalls.BlockMetadata) error {
	return nil
}

func (d *OneDriveDisk) Rename(c *gin.Context) error {
	return nil
}

func (d *OneDriveDisk) Remove(c *gin.Context) error {
	return nil
}

func (d *OneDriveDisk) SetUUID(uuid uuid.UUID) {
	d.abstractDisk.SetUUID(uuid)
}

func (d *OneDriveDisk) GetUUID() uuid.UUID {
	return d.abstractDisk.GetUUID()
}

func (d *OneDriveDisk) GetCredentials() credentials.Credentials {
	return d.abstractDisk.GetCredentials()
}

func (d *OneDriveDisk) SetCredentials(credentials credentials.Credentials) {
	d.abstractDisk.SetCredentials(credentials)
}

func (d *OneDriveDisk) CreateCredentials(c string) {
	d.abstractDisk.Credentials = credentials.NewOauthCredentials(c)
}

func (d *OneDriveDisk) GetDiskDBO(userUUID uuid.UUID, providerUUID uuid.UUID, volumeUUID uuid.UUID) dbo.Disk {
	return d.abstractDisk.GetDiskDBO(userUUID, providerUUID, volumeUUID)
}

func NewOneDriveDisk() *OneDriveDisk {
	var d *OneDriveDisk = new(OneDriveDisk)
	d.abstractDisk.Disk = d
	return d
}

func (d *OneDriveDisk) GetConfig() *oauth2.Config {
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
