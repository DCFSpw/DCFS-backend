package AbstractDisk

import (
	"dcfs/apicalls"
	"dcfs/constants"
	"dcfs/db"
	"dcfs/db/dbo"
	"dcfs/models"
	"dcfs/models/credentials"
	"github.com/google/uuid"
	"log"
)

type AbstractDisk struct {
	Disk        models.Disk
	UUID        uuid.UUID
	Credentials credentials.Credentials
	BlockSize   int
	Volume      *models.Volume
	Name        string
}

/* Mandatory Disk interface implementations */

func (d *AbstractDisk) Upload(bm *apicalls.BlockMetadata) error {
	panic("Unimplemented abstract method")
}

func (d *AbstractDisk) Download(bm *apicalls.BlockMetadata) error {
	panic("Unimplemented abstract method")
}

func (d *AbstractDisk) Rename(bm *apicalls.BlockMetadata) error {
	panic("Unimplemented abstract method")
}

func (d *AbstractDisk) Remove(bm *apicalls.BlockMetadata) error {
	panic("Unimplemented abstract method")
}

func (d *AbstractDisk) SetUUID(UUID uuid.UUID) {
	d.UUID = UUID
}

func (d *AbstractDisk) GetUUID() uuid.UUID {
	return d.UUID
}

func (d *AbstractDisk) SetVolume(volume *models.Volume) {
	d.Volume = volume
}

func (d *AbstractDisk) GetVolume() *models.Volume {
	return d.Volume
}

func (d *AbstractDisk) SetName(name string) {
	d.Name = name
}

func (d *AbstractDisk) GetName() string {
	return d.Name
}

func (d *AbstractDisk) GetCredentials() credentials.Credentials {
	return d.Credentials
}

func (d *AbstractDisk) SetCredentials(c credentials.Credentials) {
	d.Credentials = c
}

func (d *AbstractDisk) CreateCredentials(credentials string) {
	panic("Unimplemented abstract method")
}

func (d *AbstractDisk) GetProviderUUID() uuid.UUID {
	panic("Unimplemented abstract method")
}

func (d *AbstractDisk) GetDiskDBO(userUUID uuid.UUID, providerUUID uuid.UUID, volumeUUID uuid.UUID) dbo.Disk {
	credentials := ""
	if d.Credentials != nil {
		credentials = d.Credentials.ToString()
	}

	return dbo.Disk{
		AbstractDatabaseObject: dbo.AbstractDatabaseObject{UUID: d.UUID},
		UserUUID:               userUUID,
		ProviderUUID:           providerUUID,
		VolumeUUID:             volumeUUID,
		Credentials:            credentials,
		Name:                   d.Name,
	}
}

/* Additional abstract functions */

func (d *AbstractDisk) GetProvider(providerType int) uuid.UUID {
	var provider dbo.Provider
	db.DB.DatabaseHandle.Where("type = ?", providerType).First(&provider)

	if provider.Type != providerType {
		return uuid.Nil
	}

	return provider.UUID
}

func (d *AbstractDisk) Delete() (string, error) {
	// TO DO: deletion process worker
	log.Println("Deleting disk" + d.UUID.String())
	return constants.SUCCESS, nil
}
