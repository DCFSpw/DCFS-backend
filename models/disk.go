package models

import (
	"dcfs/apicalls"
	"dcfs/db"
	"dcfs/db/dbo"
	"dcfs/models/credentials"
	"github.com/google/uuid"
)

var RootUUID uuid.UUID
var DiskTypesRegistry map[int]func() Disk = make(map[int]func() Disk)

type Disk interface {
	Upload(bm *apicalls.BlockMetadata) *apicalls.ErrorWrapper
	Download(bm *apicalls.BlockMetadata) *apicalls.ErrorWrapper
	Rename(bm *apicalls.BlockMetadata) *apicalls.ErrorWrapper
	Remove(bm *apicalls.BlockMetadata) *apicalls.ErrorWrapper

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

func CreateDiskFromUUID(UUID uuid.UUID) Disk {
	var disk dbo.Disk
	var volume *Volume

	d := Transport.FindEnqueuedDisk(UUID)
	if d != nil {
		return d
	}

	err := db.DB.DatabaseHandle.Where("uuid = ?", UUID).Preload("Provider").Preload("User").Preload("Volume").Find(&disk).Error
	if err != nil {
		return nil
	}

	volume = Transport.GetVolume(disk.VolumeUUID)
	if volume == nil {
		return nil
	}

	return CreateDisk(CreateDiskMetadata{
		Disk:   &disk,
		Volume: volume,
	})
}
