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
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/drive/v3"
	"google.golang.org/api/option"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
)

type GDriveDisk struct {
	abstractDisk AbstractDisk.AbstractDisk
}

func (d *GDriveDisk) Upload(blockMetadata *apicalls.BlockMetadata) *apicalls.ErrorWrapper {
	var cred *credentials.OauthCredentials = d.GetCredentials().(*credentials.OauthCredentials)
	var client *http.Client = cred.Authenticate(&apicalls.CredentialsAuthenticateMetadata{Ctx: blockMetadata.Ctx, Config: d.GetConfig(), DiskUUID: d.GetUUID()}).(*http.Client)
	var fileCreate *drive.FilesCreateCall
	var err error

	srv, err := drive.NewService(blockMetadata.Ctx, option.WithHTTPClient(client))
	if err != nil {
		log.Printf("Unable to retrieve Drive client: %v", err)
		return apicalls.CreateErrorWrapper(constants.REMOTE_CLIENT_UNAVAILABLE, "Unable to retrieve Drive client:", err.Error())
	}

	fileCreate = srv.Files.
		Create(&(drive.File{Name: blockMetadata.UUID.String()})).
		Media(bytes.NewReader(*blockMetadata.Content)).
		ProgressUpdater(func(now, size int64) { log.Printf("[%s] %d, %d\r", blockMetadata.UUID.String(), now, size) })
	_, err = fileCreate.Do()
	if err != nil {
		log.Printf("Failed to upload block: %s, with err: %s", blockMetadata.UUID.String(), err.Error())
		return apicalls.CreateErrorWrapper(constants.REMOTE_FAILED_JOB, "Failed to upload block:", blockMetadata.UUID.String(), "with err:", err.Error())
	}

	blockMetadata.CompleteCallback(blockMetadata.FileUUID, blockMetadata.Status)
	return nil
}

func (d *GDriveDisk) Download(blockMetadata *apicalls.BlockMetadata) *apicalls.ErrorWrapper {
	var cred *credentials.OauthCredentials = d.GetCredentials().(*credentials.OauthCredentials)
	var client *http.Client = cred.Authenticate(&apicalls.CredentialsAuthenticateMetadata{Ctx: blockMetadata.Ctx, Config: d.GetConfig(), DiskUUID: d.GetUUID()}).(*http.Client)
	var err error

	srv, err := drive.NewService(blockMetadata.Ctx, option.WithHTTPClient(client))
	if err != nil {
		log.Printf("Unable to retrieve Drive client: %v", err)
		return apicalls.CreateErrorWrapper(constants.REMOTE_CLIENT_UNAVAILABLE, "Unable to retrieve Drive client:", err.Error())
	}

	// we assume there is only one file named as the block UUID
	files, err := srv.Files.
		List().
		Q(fmt.Sprintf("name = '%s'", blockMetadata.UUID.String())).
		Do()
	if err != nil {
		log.Printf("unable to retreve files from gdrive: %s", err.Error())
		return apicalls.CreateErrorWrapper(constants.REMOTE_FAILED_JOB, "Unable to retrieve Drive client:", err.Error())
	}

	if len(files.Files) == 0 {
		log.Printf("file with the given blockUUID: %s not found on gdrive", blockMetadata.UUID.String())
		return apicalls.CreateErrorWrapper(constants.REMOTE_BAD_FILE, "can't find the file with the given blockUUID: %s", blockMetadata.UUID.String())
	}

	rsp, err := srv.Files.Get(files.Files[0].Id).Download()
	if err != nil {
		log.Printf("download failed: %s", err.Error())
		return apicalls.CreateErrorWrapper(constants.REMOTE_FAILED_JOB, "download failed:", err.Error())
	}
	defer func() { _ = rsp.Body.Close() }()
	buf := bytes.NewBuffer(nil)

	n, err := io.Copy(buf, rsp.Body)
	if err != nil {
		log.Printf("download failed: %s", err.Error())
		return apicalls.CreateErrorWrapper(constants.REMOTE_FAILED_JOB, "download failed:", err.Error())
	}

	if n < blockMetadata.Size {
		log.Printf("downloaded not enough bytes: %d", n)
		return apicalls.CreateErrorWrapper(constants.REMOTE_FAILED_JOB, "downloaded not enough bytes:", fmt.Sprint(n), "out of:", strconv.FormatInt(blockMetadata.Size, 10))
	}

	block := buf.Bytes()[0:blockMetadata.Size]
	blockMetadata.Content = &block
	blockMetadata.CompleteCallback(blockMetadata.FileUUID, blockMetadata.Status)
	return nil
}

func (d *GDriveDisk) Rename(bm *apicalls.BlockMetadata) *apicalls.ErrorWrapper {
	panic("unimplemented")
}

func (d *GDriveDisk) Remove(bm *apicalls.BlockMetadata) *apicalls.ErrorWrapper {
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

func (d *GDriveDisk) SetName(name string) {
	d.abstractDisk.SetName(name)
}

func (d *GDriveDisk) GetName() string {
	return d.abstractDisk.GetName()
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

func (d *GDriveDisk) GetProviderSpace() (uint64, uint64, string) {
	var err error

	// Prepare test context
	writer := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(writer)

	// Authenticate to the remote server
	var cred *credentials.OauthCredentials = d.GetCredentials().(*credentials.OauthCredentials)
	var client *http.Client = cred.Authenticate(&apicalls.CredentialsAuthenticateMetadata{Ctx: ctx, Config: d.GetConfig(), DiskUUID: d.GetUUID()}).(*http.Client)
	if client == nil {
		return 0, 0, constants.REMOTE_CANNOT_AUTHENTICATE
	}

	// Connect to the remote server
	srv, err := drive.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		return 0, 0, constants.REMOTE_CLIENT_UNAVAILABLE
	}

	// Get the disk stats from the remote server
	about, err := srv.About.Get().Fields("storageQuota").Do()
	if err != nil {
		return 0, 0, constants.REMOTE_CANNOT_GET_STATS
	}

	return uint64(about.StorageQuota.Usage), uint64(about.StorageQuota.Limit), constants.SUCCESS
}

func (d *GDriveDisk) SetTotalSpace(quota uint64) {
	d.abstractDisk.SetTotalSpace(quota)
}

func (d *GDriveDisk) GetTotalSpace() uint64 {
	return d.abstractDisk.GetTotalSpace()
}

func (d *GDriveDisk) SetUsedSpace(usage uint64) {
	d.abstractDisk.SetUsedSpace(usage)
}

func (d *GDriveDisk) GetUsedSpace() uint64 {
	return d.abstractDisk.GetUsedSpace()
}

func (d *GDriveDisk) UpdateUsedSpace(change int64) {
	d.abstractDisk.UpdateUsedSpace(change)
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
