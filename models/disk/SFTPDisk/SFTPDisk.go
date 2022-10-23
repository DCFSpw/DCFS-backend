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
	"path/filepath"
)

type SFTPDisk struct {
	abstractDisk AbstractDisk.AbstractDisk
}

/* Mandatory Disk interface implementations */

func (d *SFTPDisk) Upload(blockMetadata *apicalls.BlockMetadata) error {
	var _client interface{} = d.GetCredentials().Authenticate(nil)
	if _client == nil {
		return fmt.Errorf("cannot connect to the remote server")
	}

	var client *sftp.Client = _client.(*sftp.Client)
	defer client.Close()
	var filepath string = filepath.Join(d.GetCredentials().GetPath(), blockMetadata.UUID.String())

	// Check if the file already exists
	remoteFile, err := client.Open(filepath)
	if err == nil {
		remoteFile.Close()
		return fmt.Errorf("Cannot o open remote file: %v", err)
	}
	err = nil

	// Create remote file
	dstFile, err := client.OpenFile(filepath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC)
	if err != nil {
		return fmt.Errorf("Cannot o open remote file: %v", err)
	}
	defer dstFile.Close()

	// Upload file content
	_, err = io.Copy(dstFile, bytes.NewReader(*blockMetadata.Content))
	if err != nil {
		return fmt.Errorf("Cannot upload local file: %v", err)
	}

	blockMetadata.CompleteCallback(blockMetadata.FileUUID, blockMetadata.Status)
	return nil
}

func (d *SFTPDisk) Download(blockMetadata *apicalls.BlockMetadata) error {
	var _client interface{} = d.GetCredentials().Authenticate(nil)
	if _client == nil {
		return fmt.Errorf("cannot connect to the remote server")
	}

	var client *sftp.Client = _client.(*sftp.Client)
	defer client.Close()

	// Open remote file
	remoteFile, err := client.OpenFile(blockMetadata.UUID.String(), os.O_RDONLY)
	if err != nil {
		return fmt.Errorf("Cannot open remote file: %v", err)
	}
	defer remoteFile.Close()

	// Download remote file
	buff, err := io.ReadAll(remoteFile)
	if err != nil {
		return fmt.Errorf("Cannot download remote file: %v", err)
	}
	blockMetadata.Content = &buff
	blockMetadata.Size = int64(len(buff))

	return nil
}

func (d *SFTPDisk) Rename(blockMetadata *apicalls.BlockMetadata) error {
	panic("Unimplemented")
}

func (d *SFTPDisk) Remove(blockMetadata *apicalls.BlockMetadata) error {
	panic("Unimplemented")
}

func (d *SFTPDisk) SetVolume(volume *models.Volume) {
	d.abstractDisk.SetVolume(volume)
}

func (d *SFTPDisk) GetVolume() *models.Volume {
	return d.abstractDisk.GetVolume()
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

/* Factory methods */

func NewSFTPDisk() *SFTPDisk {
	var d *SFTPDisk = new(SFTPDisk)
	d.abstractDisk.Disk = d
	return d
}

func init() {
	models.DiskTypesRegistry[constants.PROVIDER_TYPE_SFTP] = func() models.Disk { return NewSFTPDisk() }
}
