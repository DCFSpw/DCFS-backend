package SFTPDisk

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
	"github.com/google/uuid"
	"github.com/pkg/sftp"
	"io"
	"os"
	"time"
)

type SFTPDisk struct {
	abstractDisk AbstractDisk.AbstractDisk
}

/* Mandatory Disk interface implementations */

func (d *SFTPDisk) Upload(blockMetadata *apicalls.BlockMetadata) *apicalls.ErrorWrapper {
	var _client interface{} = d.GetCredentials().Authenticate(nil)
	if _client == nil {
		logger.Logger.Error("disk", "Cannot connect to the remote server.")
		return apicalls.CreateErrorWrapper(constants.REMOTE_CANNOT_AUTHENTICATE, "Cannot connect to the remote server")
	}

	var client *sftp.Client = _client.(*sftp.Client)
	defer client.Close()

	_p := d.abstractDisk.Credentials.GetPath()
	downloadPath := fmt.Sprintf("%s/%s", _p, blockMetadata.UUID.String())
	if _p == "/" {
		downloadPath = fmt.Sprintf("%s%s", _p, blockMetadata.UUID.String())
	} else if _p == "" {
		downloadPath = blockMetadata.UUID.String()
	}

	// Check if the file already exists
	remoteFile, err := client.Open(downloadPath)
	if err == nil {
		remoteFile.Close()
		logger.Logger.Error("disk", "Cannot open the remote file, file already exists.")
		return apicalls.CreateErrorWrapper(constants.REMOTE_BAD_FILE, "Cannot open remote file:", "File already exists")
	}
	err = nil

	// Create remote file
	dstFile, err := client.OpenFile(downloadPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC)
	if err != nil {
		logger.Logger.Error("disk", "Cannot open the remote file: ", err.Error(), ".")
		return apicalls.CreateErrorWrapper(constants.REMOTE_BAD_FILE, "Cannot open remote file:", err.Error())
	}
	defer dstFile.Close()

	// Upload file content
	_, err = io.Copy(dstFile, bytes.NewReader(*blockMetadata.Content))
	if err != nil {
		logger.Logger.Error("disk", "Cannot upload a local file: ", err.Error(), ".")
		return apicalls.CreateErrorWrapper(constants.REMOTE_FAILED_JOB, "Cannot upload a local file:", err.Error())
	}

	blockMetadata.CompleteCallback(blockMetadata.FileUUID, blockMetadata.Status)
	logger.Logger.Debug("disk", "Successfully uploaded the block: ", blockMetadata.UUID.String(), ".")
	return nil
}

func (d *SFTPDisk) Download(blockMetadata *apicalls.BlockMetadata) *apicalls.ErrorWrapper {
	var _client interface{} = d.GetCredentials().Authenticate(nil)
	if _client == nil {
		logger.Logger.Error("disk", "Cannot connect to the remote server.")
		return apicalls.CreateErrorWrapper(constants.REMOTE_CANNOT_AUTHENTICATE, "Cannot connect to the remote server")
	}

	var client *sftp.Client = _client.(*sftp.Client)
	defer client.Close()

	_p := d.abstractDisk.Credentials.GetPath()
	downloadPath := fmt.Sprintf("%s/%s", _p, blockMetadata.UUID.String())
	if _p == "/" {
		downloadPath = fmt.Sprintf("%s%s", _p, blockMetadata.UUID.String())
	} else if _p == "" {
		downloadPath = blockMetadata.UUID.String()
	}

	// Open remote file
	remoteFile, err := client.OpenFile(downloadPath, os.O_RDONLY)
	if err != nil {
		logger.Logger.Error("disk", "Cannot open the remote time: ", err.Error())
		return apicalls.CreateErrorWrapper(constants.REMOTE_BAD_FILE, "Cannot open remote file:", err.Error())
	}
	defer remoteFile.Close()

	// Download remote file
	buff, err := io.ReadAll(remoteFile)
	if err != nil {
		logger.Logger.Error("disk", "Cannot download the remote file: ", err.Error())
		return apicalls.CreateErrorWrapper(constants.REMOTE_FAILED_JOB, "Cannot download remote file:", err.Error())
	}
	blockMetadata.Content = &buff
	blockMetadata.Size = int64(len(buff))
	blockMetadata.CompleteCallback(blockMetadata.FileUUID, blockMetadata.Status)

	logger.Logger.Debug("disk", "Successfully downloaded the block: ", blockMetadata.UUID.String(), ".")
	return nil
}

func (d *SFTPDisk) Remove(blockMetadata *apicalls.BlockMetadata) *apicalls.ErrorWrapper {
	var _client interface{} = d.GetCredentials().Authenticate(nil)
	if _client == nil {
		logger.Logger.Error("disk", "Cannot connect to the remote server.")
		return apicalls.CreateErrorWrapper(constants.REMOTE_CANNOT_AUTHENTICATE, "Cannot connect to the remote server")
	}

	var client *sftp.Client = _client.(*sftp.Client)
	defer client.Close()

	_p := d.abstractDisk.Credentials.GetPath()
	downloadPath := fmt.Sprintf("%s/%s", _p, blockMetadata.UUID.String())
	if _p == "/" {
		downloadPath = fmt.Sprintf("%s%s", _p, blockMetadata.UUID.String())
	} else if _p == "" {
		downloadPath = blockMetadata.UUID.String()
	}

	// Remove remote file
	err := client.Remove(downloadPath)
	if err != nil {
		logger.Logger.Error("disk", "Cannot remove the remote file: ", err.Error())
		return apicalls.CreateErrorWrapper(constants.REMOTE_FAILED_JOB, "Cannot remove remote file:", err.Error())
	}

	blockMetadata.CompleteCallback(blockMetadata.FileUUID, blockMetadata.Status)

	logger.Logger.Debug("disk", "Successfully removed the block: ", blockMetadata.UUID.String(), ".")
	return nil
}

func (d *SFTPDisk) SetVolume(volume *models.Volume) {
	d.abstractDisk.SetVolume(volume)
}

func (d *SFTPDisk) GetVolume() *models.Volume {
	return d.abstractDisk.GetVolume()
}

func (d *SFTPDisk) SetName(name string) {
	d.abstractDisk.SetName(name)
}

func (d *SFTPDisk) GetName() string {
	return d.abstractDisk.GetName()
}

func (d *SFTPDisk) SetUUID(uuid uuid.UUID) {
	d.abstractDisk.SetUUID(uuid)
}

func (d *SFTPDisk) GetUUID() uuid.UUID {
	return d.abstractDisk.GetUUID()
}

func (d *SFTPDisk) GetCredentials() credentials.Credentials {
	return d.abstractDisk.GetCredentials()
}

func (d *SFTPDisk) SetCreationTime(creationTime time.Time) {
	d.abstractDisk.SetCreationTime(creationTime)
}

func (d *SFTPDisk) GetCreationTime() time.Time {
	return d.abstractDisk.GetCreationTime()
}

func (d *SFTPDisk) SetCredentials(credentials credentials.Credentials) {
	d.abstractDisk.SetCredentials(credentials)
}

func (d *SFTPDisk) CreateCredentials(c string) {
	d.abstractDisk.Credentials = credentials.NewSFTPCredentials(c)
}

func (d *SFTPDisk) GetProviderUUID() uuid.UUID {
	return d.abstractDisk.GetProvider(constants.PROVIDER_TYPE_SFTP)
}

func (d *SFTPDisk) GetDiskDBO(userUUID uuid.UUID, providerUUID uuid.UUID, volumeUUID uuid.UUID) dbo.Disk {
	return d.abstractDisk.GetDiskDBO(userUUID, providerUUID, volumeUUID)
}

func (d *SFTPDisk) GetProviderSpace() (uint64, uint64, string) {
	var stats *sftp.StatVFS
	var err error

	// Authenticate to the remote server
	var _client interface{} = d.GetCredentials().Authenticate(nil)
	if _client == nil {
		logger.Logger.Error("disk", "Could not authenticate to get the remote provider space.")
		return 0, 0, constants.REMOTE_CANNOT_AUTHENTICATE
	}

	// Connect to the remote server
	var client *sftp.Client = _client.(*sftp.Client)
	defer client.Close()

	path := d.abstractDisk.Credentials.GetPath()
	if path == "" {
		path = "/"
	}

	// Get the disk stats from the remote server
	stats, err = client.StatVFS(path)

	if err != nil {
		logger.Logger.Error("disk", "Could not get the remote provider stats.")
		return 0, 0, constants.OPERATION_NOT_SUPPORTED
	}

	return stats.TotalSpace() - stats.FreeSpace(), stats.TotalSpace(), constants.SUCCESS
}

func (d *SFTPDisk) SetTotalSpace(quota uint64) {
	d.abstractDisk.SetTotalSpace(quota)
}

func (d *SFTPDisk) GetTotalSpace() uint64 {
	return d.abstractDisk.GetTotalSpace()
}

func (d *SFTPDisk) SetUsedSpace(usage uint64) {
	d.abstractDisk.SetUsedSpace(usage)
}

func (d *SFTPDisk) GetUsedSpace() uint64 {
	return d.abstractDisk.GetUsedSpace()
}

func (d *SFTPDisk) UpdateUsedSpace(change int64) {
	d.abstractDisk.UpdateUsedSpace(change)
}

func (d *SFTPDisk) IsReady() bool {
	return d.abstractDisk.IsReady()
}

/* Factory methods */
func NewSFTPDisk() *SFTPDisk {
	var d *SFTPDisk = new(SFTPDisk)
	d.abstractDisk.Disk = d
	return d
}

func init() {
	models.DiskTypesRegistry[constants.PROVIDER_TYPE_SFTP] = func() models.Disk { return NewSFTPDisk() }
}
