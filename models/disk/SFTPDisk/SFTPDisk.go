package SFTPDisk

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
	"github.com/pkg/sftp"
	"io"
	"os"
)

type SFTPDisk struct {
	abstractDisk AbstractDisk.AbstractDisk
}

/* Mandatory Disk interface implementations */

func (d *SFTPDisk) Upload(blockMetadata *apicalls.BlockMetadata) *apicalls.ErrorWrapper {
	var _client interface{} = d.GetCredentials().Authenticate(nil)
	if _client == nil {
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
		return apicalls.CreateErrorWrapper(constants.REMOTE_BAD_FILE, "Cannot open remote file:", err.Error())
	}
	err = nil

	// Create remote file
	dstFile, err := client.OpenFile(downloadPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC)
	if err != nil {
		return apicalls.CreateErrorWrapper(constants.REMOTE_BAD_FILE, "Cannot open remote file:", err.Error())
	}
	defer dstFile.Close()

	// Upload file content
	_, err = io.Copy(dstFile, bytes.NewReader(*blockMetadata.Content))
	if err != nil {
		return apicalls.CreateErrorWrapper(constants.REMOTE_FAILED_JOB, "Cannot upload a local file:", err.Error())
	}

	blockMetadata.CompleteCallback(blockMetadata.FileUUID, blockMetadata.Status)
	return nil
}

func (d *SFTPDisk) Download(blockMetadata *apicalls.BlockMetadata) *apicalls.ErrorWrapper {
	var _client interface{} = d.GetCredentials().Authenticate(nil)
	if _client == nil {
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
		return apicalls.CreateErrorWrapper(constants.REMOTE_BAD_FILE, "Cannot open remote file:", err.Error())
	}
	defer remoteFile.Close()

	// Download remote file
	buff, err := io.ReadAll(remoteFile)
	if err != nil {
		return apicalls.CreateErrorWrapper(constants.REMOTE_FAILED_JOB, "Cannot download remote file:", err.Error())
	}
	blockMetadata.Content = &buff
	blockMetadata.Size = int64(len(buff))
	blockMetadata.CompleteCallback(blockMetadata.FileUUID, blockMetadata.Status)

	return nil
}

func (d *SFTPDisk) Rename(blockMetadata *apicalls.BlockMetadata) *apicalls.ErrorWrapper {
	panic("Unimplemented")
}

func (d *SFTPDisk) Remove(blockMetadata *apicalls.BlockMetadata) *apicalls.ErrorWrapper {
	panic("Unimplemented")
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

func (d *SFTPDisk) Delete() (string, error) {
	return d.abstractDisk.Delete()
}

func (d *SFTPDisk) GetProviderFreeSpace() (uint64, string) {
	var stats *sftp.StatVFS
	var err error

	// Authenticate to the remote server
	var _client interface{} = d.GetCredentials().Authenticate(nil)
	if _client == nil {
		return 0, constants.REMOTE_CANNOT_AUTHENTICATE
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
		return 0, constants.OPERATION_NOT_SUPPORTED
	}

	return stats.FreeSpace(), constants.SUCCESS
}

func (d *SFTPDisk) SetTotalSpace(quota uint64) {
	d.abstractDisk.SetTotalSpace(quota)
}

func (d *SFTPDisk) GetTotalSpace() uint64 {
	return d.abstractDisk.GetTotalSpace()
}

func (d *SFTPDisk) GetUsedSpace() uint64 {
	return d.abstractDisk.GetUsedSpace()
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
