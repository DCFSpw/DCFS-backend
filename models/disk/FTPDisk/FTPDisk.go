package FTPDisk

import (
	"bytes"
	"context"
	"dcfs/apicalls"
	"dcfs/constants"
	"dcfs/db"
	"dcfs/db/dbo"
	"dcfs/models"
	"dcfs/models/credentials"
	"dcfs/models/disk/AbstractDisk"
	"dcfs/util/logger"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jlaffaye/ftp"
	"io"
	"time"
)

type FTPDisk struct {
	abstractDisk AbstractDisk.AbstractDisk
}

/* Mandatory Disk interface methods */

func (d *FTPDisk) Upload(blockMetadata *apicalls.BlockMetadata) *apicalls.ErrorWrapper {
	// Authenticate
	var _client interface{} = d.GetCredentials().Authenticate(nil)
	if _client == nil {
		logger.Logger.Error("disk", "Could not connect to the remote server.")
		return apicalls.CreateErrorWrapper(constants.REMOTE_CANNOT_AUTHENTICATE, "could not connect to the remote server")
	}

	var client *ftp.ServerConn = _client.(*ftp.ServerConn)

	// Generate remote path
	_p := d.abstractDisk.Credentials.GetPath()
	downloadPath := fmt.Sprintf("%s/%s", _p, blockMetadata.UUID.String())
	if _p == "/" {
		downloadPath = fmt.Sprintf("%s%s", _p, blockMetadata.UUID.String())
	} else if _p == "" {
		downloadPath = blockMetadata.UUID.String()
	}

	logger.Logger.Debug("disk", "Established an upload path: ", downloadPath, " for the block: ", blockMetadata.UUID.String(), ".")

	// Upload file to server
	err := client.Stor(downloadPath, bytes.NewReader(*blockMetadata.Content))
	if err != nil {
		logger.Logger.Error("disk", "Cannot open the remote file with error: ", err.Error())
		return apicalls.CreateErrorWrapper(constants.REMOTE_BAD_FILE, "cannot open remote file:", err.Error())
	}

	blockMetadata.CompleteCallback(blockMetadata.FileUUID, blockMetadata.Status)

	logger.Logger.Debug("disk", "Successfully uploaded the block: ", blockMetadata.UUID.String(), ".")
	return nil
}

func (d *FTPDisk) Download(blockMetadata *apicalls.BlockMetadata) *apicalls.ErrorWrapper {
	// Authenticate
	var _client interface{} = d.GetCredentials().Authenticate(nil)
	if _client == nil {
		logger.Logger.Error("disk", "Could not connect to the remote server.")
		return apicalls.CreateErrorWrapper(constants.REMOTE_CANNOT_AUTHENTICATE, "could not connect to the remote server")
	}

	var client *ftp.ServerConn = _client.(*ftp.ServerConn)

	// Generate remote path
	_p := d.abstractDisk.Credentials.GetPath()
	downloadPath := fmt.Sprintf("%s/%s", _p, blockMetadata.UUID.String())
	if _p == "/" {
		downloadPath = fmt.Sprintf("%s%s", _p, blockMetadata.UUID.String())
	} else if _p == "" {
		downloadPath = blockMetadata.UUID.String()
	}

	logger.Logger.Debug("disk", "Established a download path: ", downloadPath, " for the block: ", blockMetadata.UUID.String(), ".")

	// Download file from server
	reader, err := client.Retr(downloadPath)
	if err != nil {
		logger.Logger.Error("disk", "Cannot open the remote file, got an error: ", err.Error())
		return apicalls.CreateErrorWrapper(constants.REMOTE_BAD_FILE, "cannot open remote file:", err.Error())
	}

	// Load file content
	buff, err := io.ReadAll(reader)
	if err != nil {
		logger.Logger.Error("disk", "Cannot open the remote file, got an error: ", err.Error())
		return apicalls.CreateErrorWrapper(constants.REMOTE_BAD_FILE, "cannot open remote file:", err.Error())
	}
	blockMetadata.Content = &buff
	blockMetadata.Size = int64(len(buff))
	blockMetadata.CompleteCallback(blockMetadata.FileUUID, blockMetadata.Status)

	logger.Logger.Debug("disk", "Successfully downloaded the block: ", blockMetadata.UUID.String(), ".")
	return nil
}

func (d *FTPDisk) Remove(blockMetadata *apicalls.BlockMetadata) *apicalls.ErrorWrapper {
	// Authenticate
	var _client interface{} = d.GetCredentials().Authenticate(nil)
	if _client == nil {
		logger.Logger.Error("disk", "Could not connect to the remote server.")
		return apicalls.CreateErrorWrapper(constants.REMOTE_CANNOT_AUTHENTICATE, "could not connect to the remote server")
	}

	var client *ftp.ServerConn = _client.(*ftp.ServerConn)

	// Generate remote path
	_p := d.abstractDisk.Credentials.GetPath()
	downloadPath := fmt.Sprintf("%s/%s", _p, blockMetadata.UUID.String())
	if _p == "/" {
		downloadPath = fmt.Sprintf("%s%s", _p, blockMetadata.UUID.String())
	} else if _p == "" {
		downloadPath = blockMetadata.UUID.String()
	}

	// Delete file from server
	err := client.Delete(downloadPath)
	if err != nil {
		logger.Logger.Error("disk", "Cannot remove the remote file: ", err.Error())
		return apicalls.CreateErrorWrapper(constants.REMOTE_FAILED_JOB, "Cannot remove remote file:", err.Error())
	}

	blockMetadata.CompleteCallback(blockMetadata.FileUUID, blockMetadata.Status)

	logger.Logger.Debug("disk", "Successfully removed the block: ", blockMetadata.UUID.String(), ".")
	return nil
}

func (d *FTPDisk) SetUUID(uuid uuid.UUID) {
	d.abstractDisk.SetUUID(uuid)
}

func (d *FTPDisk) GetUUID() uuid.UUID {
	return d.abstractDisk.GetUUID()
}

func (d *FTPDisk) SetVolume(volume *models.Volume) {
	d.abstractDisk.SetVolume(volume)
}

func (d *FTPDisk) GetVolume() *models.Volume {
	return d.abstractDisk.GetVolume()
}

func (d *FTPDisk) SetName(name string) {
	d.abstractDisk.SetName(name)
}

func (d *FTPDisk) GetName() string {
	return d.abstractDisk.GetName()
}

func (d *FTPDisk) GetCredentials() credentials.Credentials {
	return d.abstractDisk.GetCredentials()
}

func (d *FTPDisk) SetCredentials(credentials credentials.Credentials) {
	d.abstractDisk.SetCredentials(credentials)
}

func (d *FTPDisk) CreateCredentials(c string) {
	d.abstractDisk.Credentials = credentials.NewFTPCredentials(c)
}

func (d *FTPDisk) SetCreationTime(creationTime time.Time) {
	d.abstractDisk.SetCreationTime(creationTime)
}

func (d *FTPDisk) GetCreationTime() time.Time {
	return d.abstractDisk.GetCreationTime()
}

func (d *FTPDisk) GetProviderUUID() uuid.UUID {
	return d.abstractDisk.GetProvider(constants.PROVIDER_TYPE_FTP)
}

func (d *FTPDisk) SetIsVirtualFlag(isVirtual bool) {
	d.abstractDisk.SetIsVirtualFlag(isVirtual)
}

func (d *FTPDisk) GetIsVirtualFlag() bool {
	return d.abstractDisk.GetIsVirtualFlag()
}

func (d *FTPDisk) SetVirtualDiskUUID(uuid uuid.UUID) {
	d.abstractDisk.SetVirtualDiskUUID(uuid)
}

func (d *FTPDisk) GetVirtualDiskUUID() uuid.UUID {
	return d.abstractDisk.GetVirtualDiskUUID()
}

func (d *FTPDisk) GetProviderSpace() (uint64, uint64, string) {
	return 0, 0, constants.OPERATION_NOT_SUPPORTED
}

func (d *FTPDisk) SetTotalSpace(quota uint64) {
	d.abstractDisk.SetTotalSpace(quota)
}

func (d *FTPDisk) GetTotalSpace() uint64 {
	return d.abstractDisk.GetTotalSpace()
}

func (d *FTPDisk) SetUsedSpace(usage uint64) {
	d.abstractDisk.SetUsedSpace(usage)
}

func (d *FTPDisk) GetUsedSpace() uint64 {
	return d.abstractDisk.GetUsedSpace()
}

func (d *FTPDisk) UpdateUsedSpace(change int64) {
	d.abstractDisk.UpdateUsedSpace(change)
}

func (d *FTPDisk) GetDiskDBO(userUUID uuid.UUID, providerUUID uuid.UUID, volumeUUID uuid.UUID) dbo.Disk {
	return d.abstractDisk.GetDiskDBO(userUUID, providerUUID, volumeUUID)
}

func (d *FTPDisk) AssignDisk(disk models.Disk) {
	d.abstractDisk.AssignDisk(disk)
}

func (d *FTPDisk) GetReadiness() models.DiskReadiness {
	return d.abstractDisk.DiskReadiness
}

func (d *FTPDisk) GetResponse(_disk *dbo.Disk, ctx *gin.Context) *models.DiskResponse {
	return d.abstractDisk.GetResponse(_disk, ctx)
}

/* Factory methods */

func NewFTPDisk() *FTPDisk {
	var d *FTPDisk = new(FTPDisk)
	d.abstractDisk.Disk = d
	d.abstractDisk.DiskReadiness = models.DiskReadinessRegistry[constants.PROVIDER_TYPE_FTP](d)
	return d
}

func init() {
	models.DiskTypesRegistry[constants.PROVIDER_TYPE_FTP] = func() models.Disk { return NewFTPDisk() }
	models.DiskReadinessRegistry[constants.PROVIDER_TYPE_FTP] = func(d models.Disk) models.DiskReadiness {
		return models.NewRealDiskReadiness(func(ctx context.Context) bool {
			logger.Logger.Debug("drive", "Checking readiness for FTP drive: ", d.GetUUID().String(), ".")
			if d.GetCredentials().Authenticate(&apicalls.CredentialsAuthenticateMetadata{
				Ctx:      ctx,
				Config:   nil,
				DiskUUID: d.GetUUID(),
			}) == nil {
				return false
			}

			return true
		}, func() bool { return models.Transport.ActiveVolumes.GetEnqueuedInstance(d.GetVolume().UUID) != nil })
	}
	models.ProviderTypesRegistry[constants.PROVIDER_TYPE_FTP] = func() {
		provider := dbo.Provider{}
		db.DB.DatabaseHandle.Where("type = ?", constants.PROVIDER_TYPE_FTP).First(&provider)
		if provider.Type != constants.PROVIDER_TYPE_FTP {
			provider.UUID = uuid.New()
			provider.Type = constants.PROVIDER_TYPE_FTP
			provider.Name = "FTP drive"
			provider.Logo = "https://upload.wikimedia.org/wikipedia/commons/thumb/8/80/Antu_gFTP.svg/640px-Antu_gFTP.svg.png"

			db.DB.DatabaseHandle.Create(&provider)
		}
	}
}
