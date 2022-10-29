package FTPDisk

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
	"github.com/jlaffaye/ftp"
	"io"
)

type FTPDisk struct {
	abstractDisk AbstractDisk.AbstractDisk
}

/* Mandatory Disk interface methods */

func (d *FTPDisk) Upload(blockMetadata *apicalls.BlockMetadata) error {
	// Create and upload remote file
	var _client interface{} = d.GetCredentials().Authenticate(nil)
	if _client == nil {
		return fmt.Errorf("could not connect to the remote server")
	}

	var client *ftp.ServerConn = _client.(*ftp.ServerConn)

	err := client.Stor(blockMetadata.UUID.String(), bytes.NewReader(*blockMetadata.Content))
	if err != nil {
		return fmt.Errorf("cannot open remote file: %v", err)
	}

	return nil
}

func (d *FTPDisk) Download(blockMetadata *apicalls.BlockMetadata) error {
	// Download remote file
	var _client interface{} = d.GetCredentials().Authenticate(nil)
	if _client == nil {
		return fmt.Errorf("could not connect to the remote server")
	}

	var client *ftp.ServerConn = _client.(*ftp.ServerConn)

	reader, err := client.Retr(blockMetadata.UUID.String())
	if err != nil {
		return fmt.Errorf("Cannot open remote file: %v", err)
	}
	//defer reader.Close()

	buff, err := io.ReadAll(reader)
	if err != nil {
		return fmt.Errorf("Cannot read remote file: %v", err)
	}
	blockMetadata.Content = &buff
	blockMetadata.Size = int64(len(buff))

	return nil
}

func (d *FTPDisk) Rename(blockMetadata *apicalls.BlockMetadata) error {
	panic("Unimplemented")
}

func (d *FTPDisk) Remove(blockMetadata *apicalls.BlockMetadata) error {
	panic("Unimplemented")
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

func (d *FTPDisk) GetProviderUUID() uuid.UUID {
	return d.abstractDisk.GetProvider(constants.PROVIDER_TYPE_FTP)
}

func (d *FTPDisk) Delete() (string, error) {
	return d.abstractDisk.Delete()
}

func (d *FTPDisk) GetDiskDBO(userUUID uuid.UUID, providerUUID uuid.UUID, volumeUUID uuid.UUID) dbo.Disk {
	return d.abstractDisk.GetDiskDBO(userUUID, providerUUID, volumeUUID)
}

/* Factory methods */

func NewFTPDisk() *FTPDisk {
	var d *FTPDisk = new(FTPDisk)
	d.abstractDisk.Disk = d
	return d
}

func init() {
	models.DiskTypesRegistry[constants.PROVIDER_TYPE_FTP] = func() models.Disk { return NewFTPDisk() }
}
