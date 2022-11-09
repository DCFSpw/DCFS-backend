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

func (d *FTPDisk) Upload(blockMetadata *apicalls.BlockMetadata) *apicalls.ErrorWrapper {
	// Create and upload remote file
	var _client interface{} = d.GetCredentials().Authenticate(nil)
	if _client == nil {
		return apicalls.CreateErrorWrapper(constants.REMOTE_CANNOT_AUTHENTICATE, "could not connect to the remote server")
	}

	var client *ftp.ServerConn = _client.(*ftp.ServerConn)

	_p := d.abstractDisk.Credentials.GetPath()
	downloadPath := fmt.Sprintf("%s/%s", _p, blockMetadata.UUID.String())
	if _p == "/" {
		downloadPath = fmt.Sprintf("%s%s", _p, blockMetadata.UUID.String())
	} else if _p == "" {
		downloadPath = blockMetadata.UUID.String()
	}

	err := client.Stor(downloadPath, bytes.NewReader(*blockMetadata.Content))
	if err != nil {
		return apicalls.CreateErrorWrapper(constants.REMOTE_BAD_FILE, "cannot open remote file:", err.Error())
	}

	blockMetadata.CompleteCallback(blockMetadata.FileUUID, blockMetadata.Status)
	return nil
}

func (d *FTPDisk) Download(blockMetadata *apicalls.BlockMetadata) *apicalls.ErrorWrapper {
	// Download remote file
	var _client interface{} = d.GetCredentials().Authenticate(nil)
	if _client == nil {
		return apicalls.CreateErrorWrapper(constants.REMOTE_CANNOT_AUTHENTICATE, "could not connect to the remote server")
	}

	var client *ftp.ServerConn = _client.(*ftp.ServerConn)

	_p := d.abstractDisk.Credentials.GetPath()
	downloadPath := fmt.Sprintf("%s/%s", _p, blockMetadata.UUID.String())
	if _p == "/" {
		downloadPath = fmt.Sprintf("%s%s", _p, blockMetadata.UUID.String())
	} else if _p == "" {
		downloadPath = blockMetadata.UUID.String()
	}

	reader, err := client.Retr(downloadPath)
	if err != nil {
		return apicalls.CreateErrorWrapper(constants.REMOTE_BAD_FILE, "cannot open remote file:", err.Error())
	}
	//defer reader.Close()

	buff, err := io.ReadAll(reader)
	if err != nil {
		return apicalls.CreateErrorWrapper(constants.REMOTE_BAD_FILE, "cannot open remote file:", err.Error())
	}
	blockMetadata.Content = &buff
	blockMetadata.Size = int64(len(buff))
	blockMetadata.CompleteCallback(blockMetadata.FileUUID, blockMetadata.Status)

	return nil
}

func (d *FTPDisk) Rename(blockMetadata *apicalls.BlockMetadata) *apicalls.ErrorWrapper {
	panic("Unimplemented")
}

func (d *FTPDisk) Remove(blockMetadata *apicalls.BlockMetadata) *apicalls.ErrorWrapper {
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
