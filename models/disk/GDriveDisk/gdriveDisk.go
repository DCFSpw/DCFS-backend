package GDriveDisk

import (
	"bytes"
	"dcfs/apicalls"
	"dcfs/constants"
	"dcfs/db/dbo"
	"dcfs/models"
	"dcfs/models/credentials"
	"dcfs/models/disk/AbstractDisk"
	"dcfs/util/logger"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/drive/v3"
	"google.golang.org/api/option"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"time"
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
		logger.Logger.Error("disk", "Unable to retrieve the Google Drive client, got an error: ", err.Error())
		return apicalls.CreateErrorWrapper(constants.REMOTE_CLIENT_UNAVAILABLE, "Unable to retrieve Drive client:", err.Error())
	}

	fileCreate = srv.Files.
		Create(&(drive.File{Name: blockMetadata.UUID.String()})).
		Media(bytes.NewReader(*blockMetadata.Content)).
		ProgressUpdater(func(now, size int64) {
			logger.Logger.Debug("disk", "block upload: ", blockMetadata.UUID.String(), " progress: ", strconv.FormatInt(now, 10), "/", strconv.FormatInt(size, 10))
		})
	_, err = fileCreate.Do()
	if err != nil {
		logger.Logger.Error("disk", "Failed to upload block: ", blockMetadata.UUID.String(), " with err: ", err.Error(), ".")
		return apicalls.CreateErrorWrapper(constants.REMOTE_FAILED_JOB, "Failed to upload block:", blockMetadata.UUID.String(), "with err:", err.Error())
	}

	blockMetadata.CompleteCallback(blockMetadata.FileUUID, blockMetadata.Status)
	logger.Logger.Debug("disk", "Successfully uploaded the block: ", blockMetadata.UUID.String())
	return nil
}

func (d *GDriveDisk) Download(blockMetadata *apicalls.BlockMetadata) *apicalls.ErrorWrapper {
	var cred *credentials.OauthCredentials = d.GetCredentials().(*credentials.OauthCredentials)
	var client *http.Client = cred.Authenticate(&apicalls.CredentialsAuthenticateMetadata{Ctx: blockMetadata.Ctx, Config: d.GetConfig(), DiskUUID: d.GetUUID()}).(*http.Client)
	var err error

	srv, err := drive.NewService(blockMetadata.Ctx, option.WithHTTPClient(client))
	if err != nil {
		logger.Logger.Error("disk", "Unable to retrieve the Google Drive client, got an error: ", err.Error())
		return apicalls.CreateErrorWrapper(constants.REMOTE_CLIENT_UNAVAILABLE, "Unable to retrieve Drive client:", err.Error())
	}

	// we assume there is only one file named as the block UUID
	files, err := srv.Files.
		List().
		Q(fmt.Sprintf("name = '%s'", blockMetadata.UUID.String())).
		Do()
	if err != nil {
		logger.Logger.Error("disk", "Unable to retrieve files from Google Drive, got an error: ", err.Error())
		return apicalls.CreateErrorWrapper(constants.REMOTE_FAILED_JOB, "Unable to retrieve files from Google Drive:", err.Error())
	}

	if len(files.Files) == 0 {
		logger.Logger.Debug("disk", "file with the given block uuid", blockMetadata.UUID.String(), " not found on the Google Drive.")
		return apicalls.CreateErrorWrapper(constants.REMOTE_BAD_FILE, "can't find the file with the given blockUUID: %s", blockMetadata.UUID.String())
	}

	rsp, err := srv.Files.Get(files.Files[0].Id).Download()
	if err != nil {
		logger.Logger.Error("disk", "Download failed: ", err.Error())
		return apicalls.CreateErrorWrapper(constants.REMOTE_FAILED_JOB, "download failed:", err.Error())
	}
	defer func() { _ = rsp.Body.Close() }()
	buf := bytes.NewBuffer(nil)

	n, err := io.Copy(buf, rsp.Body)
	if err != nil {
		logger.Logger.Error("disk", "Download failed: ", err.Error())
		return apicalls.CreateErrorWrapper(constants.REMOTE_FAILED_JOB, "download failed:", err.Error())
	}

	if n < blockMetadata.Size {
		logger.Logger.Error("disk", "Downloaded not enough bytes: ", strconv.FormatInt(n, 10))
		return apicalls.CreateErrorWrapper(constants.REMOTE_FAILED_JOB, "downloaded not enough bytes:", fmt.Sprint(n), "out of:", strconv.FormatInt(blockMetadata.Size, 10))
	}

	block := buf.Bytes()[0:blockMetadata.Size]
	blockMetadata.Content = &block
	blockMetadata.CompleteCallback(blockMetadata.FileUUID, blockMetadata.Status)

	logger.Logger.Debug("disk", "Successfully downloaded the block: ", blockMetadata.UUID.String(), ".")
	return nil
}

func (d *GDriveDisk) Remove(bm *apicalls.BlockMetadata) *apicalls.ErrorWrapper {
	var cred *credentials.OauthCredentials = d.GetCredentials().(*credentials.OauthCredentials)
	var client *http.Client = cred.Authenticate(&apicalls.CredentialsAuthenticateMetadata{Ctx: bm.Ctx, Config: d.GetConfig(), DiskUUID: d.GetUUID()}).(*http.Client)
	var fileDelete *drive.FilesDeleteCall
	var err error

	srv, err := drive.NewService(bm.Ctx, option.WithHTTPClient(client))
	if err != nil {
		logger.Logger.Error("disk", "Unable to retrieve the Google Drive client, got an error: ", err.Error())
		return apicalls.CreateErrorWrapper(constants.REMOTE_CLIENT_UNAVAILABLE, "Unable to retrieve Drive client:", err.Error())
	}

	files, err := srv.Files.
		List().
		Q(fmt.Sprintf("name = '%s'", bm.UUID.String())).
		Do()
	if err != nil {
		logger.Logger.Error("disk", "Unable to retrieve files from Google Drive, got an error: ", err.Error())
		return apicalls.CreateErrorWrapper(constants.REMOTE_FAILED_JOB, "Unable to retrieve files from Google Drive:", err.Error())
	}

	if len(files.Files) == 0 {
		logger.Logger.Debug("disk", "file with the given block uuid", bm.UUID.String(), " not found on the Google Drive.")
		return apicalls.CreateErrorWrapper(constants.REMOTE_BAD_FILE, "can't find the file with the given blockUUID: %s", bm.UUID.String())
	}

	fileDelete = srv.Files.Delete(files.Files[0].Id)
	err = fileDelete.Do()
	if err != nil {
		logger.Logger.Error("disk", "Failed to remove block: ", bm.UUID.String(), " with err: ", err.Error(), ".")
		return apicalls.CreateErrorWrapper(constants.REMOTE_FAILED_JOB, "Failed to remove block:", bm.UUID.String(), "with err:", err.Error())
	}

	bm.CompleteCallback(bm.FileUUID, bm.Status)

	logger.Logger.Debug("disk", "Successfully removed the block: ", bm.UUID.String(), ".")
	return nil
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

func (d *GDriveDisk) SetCreationTime(creationTime time.Time) {
	d.abstractDisk.SetCreationTime(creationTime)
}

func (d *GDriveDisk) GetCreationTime() time.Time {
	return d.abstractDisk.GetCreationTime()
}

func (d *GDriveDisk) GetProviderUUID() uuid.UUID {
	return d.abstractDisk.GetProvider(constants.PROVIDER_TYPE_GDRIVE)
}

func (d *GDriveDisk) SetIsVirtualFlag(isVirtual bool) {
	d.abstractDisk.SetIsVirtualFlag(isVirtual)
}

func (d *GDriveDisk) GetIsVirtualFlag() bool {
	return d.abstractDisk.GetIsVirtualFlag()
}

func (d *GDriveDisk) SetVirtualDiskUUID(uuid uuid.UUID) {
	d.abstractDisk.SetVirtualDiskUUID(uuid)
}

func (d *GDriveDisk) GetVirtualDiskUUID() uuid.UUID {
	return d.abstractDisk.GetVirtualDiskUUID()
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
		logger.Logger.Warning("disk", "Cannot authenticate to get the provider space.")
		return 0, 0, constants.REMOTE_CANNOT_AUTHENTICATE
	}

	// Connect to the remote server
	srv, err := drive.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		logger.Logger.Warning("disk", "Remote client unavailable")
		return 0, 0, constants.REMOTE_CLIENT_UNAVAILABLE
	}

	// Get the disk stats from the remote server
	about, err := srv.About.Get().Fields("storageQuota").Do()
	if err != nil {
		logger.Logger.Warning("disk", "Could not get the remote provider free space.")
		return 0, 0, constants.REMOTE_CANNOT_GET_STATS
	}

	logger.Logger.Debug("disk", "Successfully obtained the Google Drive free space.")
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

/* Mandatory OAuthDisk interface methods */
func (d *GDriveDisk) GetConfig() *oauth2.Config {
	b, err := os.ReadFile("./models/disk/GDriveDisk/credentials.json")
	if err != nil {
		logger.Logger.Error("disk", "Unable to read client secret file: ", err.Error())
		return nil
	}

	// If modifying these scopes, delete your previously saved token.json.
	config, err := google.ConfigFromJSON(b, drive.DriveScope)
	if err != nil {
		logger.Logger.Error("disk", "Unable to parse client secret file to config: ", err.Error())
		return nil
	}

	return config
}

func (d *GDriveDisk) AssignDisk(disk models.Disk) {
	d.abstractDisk.AssignDisk(disk)
}

func (d *GDriveDisk) IsReady(ctx *gin.Context) bool {
	// check if it is possible to connect to a disk
	client := d.GetCredentials().Authenticate(&apicalls.CredentialsAuthenticateMetadata{
		Ctx:      ctx,
		Config:   d.GetConfig(),
		DiskUUID: uuid.UUID{},
	}).(*http.Client)

	srv, err := drive.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		return false
	}

	_, err = srv.Files.
		List().
		Q(fmt.Sprintf("name = ''")).
		Do()
	if err != nil {
		return false
	}

	return true
}

func (d *GDriveDisk) GetResponse(_disk *dbo.Disk, ctx *gin.Context) *models.DiskResponse {
	return &models.DiskResponse{
		Disk:    *_disk,
		Array:   nil,
		IsReady: d.IsReady(ctx),
	}
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
