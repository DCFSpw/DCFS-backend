package SFTPDisk

import (
	"dcfs/db/dbo"
	"dcfs/models/credentials"
	"dcfs/models/disk"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type SFTPDisk struct {
	abstractDisk disk.AbstractDisk
	credentials  *credentials.SFTPCredentials
}

func (d *SFTPDisk) Connect(c *gin.Context) error {
	// unpack gin context
	// d.connect(credentials)
	panic("Unimplemented")
	return nil
}

func (d *SFTPDisk) Upload(c *gin.Context) error {
	// unpack gin context
	// d.upload(fileName, fileContents)
	panic("Unimplemented")
	return nil
}

func (d *SFTPDisk) Download(c *gin.Context) error {
	// unpack gin context
	// d.download(fileName, fileContents)
	panic("Unimplemented")
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
