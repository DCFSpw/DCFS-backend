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
	"io"
	"os"
)

type SFTPDisk struct {
	abstractDisk disk.AbstractDisk
	credentials  *credentials.SFTPCredentials
}

func (d *SFTPDisk) Connect(c *gin.Context) error {
	// Import generic credentials
	d.credentials = d.abstractDisk.Credentials.(*credentials.SFTPCredentials)

	// Authenticate and connect to SFTP server
	d.credentials.Authenticate(nil)

	return nil
}

func (d *SFTPDisk) Upload(bm apicalls.BlockMetadata) error {
	var blockMetadata *apicalls.SFTPBlockMetadata = bm.(*apicalls.SFTPBlockMetadata)

	// Create remote file
	remoteFile, err := d.credentials.Client.OpenFile(blockMetadata.UUID.String(), os.O_WRONLY|os.O_CREATE|os.O_TRUNC)
	if err != nil {
		return fmt.Errorf("Cannot o open remote file: %v", err)
	}
	defer remoteFile.Close()

	// Upload file content
	_, err = io.Copy(remoteFile, bytes.NewReader(*blockMetadata.Content))
	if err != nil {
		return fmt.Errorf("Cannot upload local file: %v", err)
	}

	return nil
}

func (d *SFTPDisk) Download(bm apicalls.BlockMetadata) error {
	var blockMetadata *apicalls.SFTPBlockMetadata = bm.(*apicalls.SFTPBlockMetadata)

	// Open remote file
	remoteFile, err := d.credentials.Client.OpenFile(blockMetadata.UUID.String(), os.O_RDONLY)
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
	// unpack gin context
	// d.rename(oldName, newName)
	panic("Unimplemented")
	return nil
}

func (d *SFTPDisk) Remove(c *gin.Context) error {
	// unpack gin context
	// d.remove(fileName)
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
