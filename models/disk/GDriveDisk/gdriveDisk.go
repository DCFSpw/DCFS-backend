package GDriveDisk

import (
	"bytes"
	"dcfs/apicalls"
	"dcfs/constants"
	"dcfs/db/dbo"
	"dcfs/models"
	"dcfs/models/credentials"
	"dcfs/models/disk/AbstractDisk"
	"fmt"
	"github.com/google/uuid"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/drive/v3"
	"google.golang.org/api/option"
	"log"
	"net/http"
	"os"
)

type GDriveDisk struct {
	abstractDisk AbstractDisk.AbstractDisk
}

func (d *GDriveDisk) Upload(blockMetadata *apicalls.BlockMetadata) error {
	var cred *credentials.OauthCredentials = d.GetCredentials().(*credentials.OauthCredentials)
	var client *http.Client = cred.Authenticate(&apicalls.CredentialsAuthenticateMetadata{Ctx: blockMetadata.Ctx, Config: d.GetConfig(), DiskUUID: d.GetUUID()}).(*http.Client)
	var fileCreate *drive.FilesCreateCall
	var err error

	srv, err := drive.NewService(blockMetadata.Ctx, option.WithHTTPClient(client))
	if err != nil {
		log.Fatalf("Unable to retrieve Drive client: %v", err)
	}

	//cred.Token.AccessToken

	fileCreate = srv.Files.
		Create(&(drive.File{Name: blockMetadata.UUID.String()})).
		Media(bytes.NewReader(*blockMetadata.Content)).
		ProgressUpdater(func(now, size int64) { fmt.Printf("%d, %d\r", now, size) })
	_, err = fileCreate.Do()
	if err != nil {
		// TODO: error handling
	}

	blockMetadata.CompleteCallback(blockMetadata.FileUUID, blockMetadata.Status)
	return nil
}

func (d *GDriveDisk) Download(bm *apicalls.BlockMetadata) error {
	panic("unimplemented")
}

func (d *GDriveDisk) Rename(bm *apicalls.BlockMetadata) error {
	panic("unimplemented")
}

func (d *GDriveDisk) Remove(bm *apicalls.BlockMetadata) error {
	panic("unimplemented")
}

func (d *GDriveDisk) SetUUID(uuid uuid.UUID) {
	d.abstractDisk.SetUUID(uuid)
}

func (d *GDriveDisk) GetUUID() uuid.UUID {
	return d.abstractDisk.GetUUID()
}

func (d *GDriveDisk) SetVolume(volume *models.Volume) {
	d.abstractDisk.SetVolume(volume)
}

func (d *GDriveDisk) GetVolume() *models.Volume {
	return d.abstractDisk.GetVolume()
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

func (d *GDriveDisk) GetProviderUUID() uuid.UUID {
	return d.abstractDisk.GetProvider(constants.PROVIDER_TYPE_GDRIVE)
}

func (d *GDriveDisk) GetDiskDBO(userUUID uuid.UUID, providerUUID uuid.UUID, volumeUUID uuid.UUID) dbo.Disk {
	return d.abstractDisk.GetDiskDBO(userUUID, providerUUID, volumeUUID)
}

func (d *GDriveDisk) Delete() (string, error) {
	return d.abstractDisk.Delete()
}

/* Mandatory OAuthDisk interface methods */
func (d *GDriveDisk) GetConfig() *oauth2.Config {
	b, err := os.ReadFile("./models/disk/GDriveDisk/credentials.json")
	if err != nil {
		log.Fatalf("Unable to read client secret file: %v", err)
	}

	// If modifying these scopes, delete your previously saved token.json.
	config, err := google.ConfigFromJSON(b, drive.DriveScope)
	if err != nil {
		log.Fatalf("Unable to parse client secret file to config: %v", err)
	}

	return config
}

/* Factory methods */

func NewGDriveDisk() *GDriveDisk {
	var d *GDriveDisk = new(GDriveDisk)
	d.abstractDisk.Disk = d
	return d
}

func init() {
	models.DiskTypesRegistry[constants.PROVIDER_TYPE_GDRIVE] = func() models.Disk { return NewGDriveDisk() }
}
