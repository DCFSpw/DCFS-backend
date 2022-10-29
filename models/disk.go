package models

import (
	"dcfs/apicalls"
	"dcfs/db/dbo"
	"dcfs/models/credentials"
	"github.com/google/uuid"
)

var RootUUID uuid.UUID
var DiskTypesRegistry map[int]func() Disk = make(map[int]func() Disk)

type Disk interface {
	Upload(bm *apicalls.BlockMetadata) error
	Download(bm *apicalls.BlockMetadata) error
	Rename(bm *apicalls.BlockMetadata) error
	Remove(bm *apicalls.BlockMetadata) error

	SetUUID(uuid.UUID)
	GetUUID() uuid.UUID

	SetVolume(volume *Volume)
	GetVolume() *Volume

	GetName() string
	SetName(name string)

	GetCredentials() credentials.Credentials
	SetCredentials(credentials.Credentials)
	CreateCredentials(credentials string)
	GetProviderUUID() uuid.UUID

	GetDiskDBO(userUUID uuid.UUID, providerUUID uuid.UUID, volumeUUID uuid.UUID) dbo.Disk

	Delete() (string, error)
}

type CreateDiskMetadata struct {
	Disk   *dbo.Disk
	Volume *Volume
}

func CreateDisk(cdm CreateDiskMetadata) Disk {
	if DiskTypesRegistry[cdm.Disk.Provider.Type] == nil {
		return nil
	}
	var disk Disk = DiskTypesRegistry[cdm.Disk.Provider.Type]()

	disk.SetVolume(cdm.Volume)
	disk.CreateCredentials(cdm.Disk.Credentials)
	disk.SetUUID(cdm.Disk.UUID)
	disk.SetName(cdm.Disk.Name)
	cdm.Volume.AddDisk(disk.GetUUID(), disk)

	return disk
}
