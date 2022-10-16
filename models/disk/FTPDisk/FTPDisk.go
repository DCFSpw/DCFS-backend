package FTPDisk

import (
	"bytes"
	"dcfs/apicalls"
	"dcfs/db/dbo"
	"dcfs/models/credentials"
	"dcfs/models/disk"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jlaffaye/ftp"
	"io"
)

type FTPDisk struct {
	abstractDisk disk.AbstractDisk
}

func (d *FTPDisk) Connect(c *gin.Context) error {
	// Import generic credentials
	/*
		d.Credentials = d.abstractDisk.Credentials.(*credentials.FTPCredentials)

		// Authenticate and connect to SFTP server
		err := d.credentials.Authenticate(nil)
		if err != nil {
			return fmt.Errorf("Cannot connect to FTP server: %v", err)
		}
	*/
	return nil
}

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

func (d *FTPDisk) Rename(c *gin.Context) error {
	// unpack gin context
	// d.rename(oldName, newName)
	panic("Unimplemented")
	return nil
}

func (d *FTPDisk) Remove(c *gin.Context) error {
	// unpack gin context
	// d.remove(fileName)
	panic("Unimplemented")
	return nil
}

func (d *FTPDisk) SetUUID(uuid uuid.UUID) {
	d.abstractDisk.SetUUID(uuid)
}

func (d *FTPDisk) GetUUID() uuid.UUID {
	return d.abstractDisk.GetUUID()
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

func (d *FTPDisk) GetDiskDBO(userUUID uuid.UUID, providerUUID uuid.UUID, volumeUUID uuid.UUID) dbo.Disk {
	return d.abstractDisk.GetDiskDBO(userUUID, providerUUID, volumeUUID)
}

func NewFTPDisk() *FTPDisk {
	var d *FTPDisk = new(FTPDisk)
	d.abstractDisk.Disk = d
	return d
}
