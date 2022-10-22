package SFTPDisk

import (
	"bytes"
	"dcfs/apicalls"
	"dcfs/db/dbo"
	"dcfs/models/credentials"
	"dcfs/models/disk"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/pkg/sftp"
	"io"
	"os"
	"path/filepath"
)

type SFTPDisk struct {
	abstractDisk disk.AbstractDisk
}

func (d *SFTPDisk) Connect(c *gin.Context) error {
	// Import generic credentials
	/*
		d.credentials = d.abstractDisk.Credentials.(*credentials.SFTPCredentials)

		// Authenticate and connect to SFTP server
		d.credentials.Authenticate(nil)
	*/

	return nil
}

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

func (d *SFTPDisk) Rename(c *gin.Context) error {
	// unpack gin content to get old name and new name
	/*
		var client *sftp.Client = d.credentials.Authenticate(nil).(*sftp.Client)
		defer client.Close()

		err := client.Rename(oldName.String(), newName.String())
		if err != nil {
			return err
		}
	*/
	panic("Unimplemented")
	return nil
}

func (d *SFTPDisk) Remove(c *gin.Context) error {
	// unpack gin context

	/*
		var client *sftp.Client = d.credentials.Authenticate(nil).(*sftp.Client)
		defer client.Close()

		err := client.Remove(fileName.String())
		if err != nil {
			return err
		}
	*/

	panic("Unimplemented")
	return nil
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

func (d *SFTPDisk) GetDiskDBO(userUUID uuid.UUID, providerUUID uuid.UUID, volumeUUID uuid.UUID) dbo.Disk {
	return d.abstractDisk.GetDiskDBO(userUUID, providerUUID, volumeUUID)
}

func NewSFTPDisk() *SFTPDisk {
	var d *SFTPDisk = new(SFTPDisk)
	d.abstractDisk.Disk = d
	return d
}
