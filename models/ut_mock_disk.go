package models

import (
	"dcfs/apicalls"
	"dcfs/db/dbo"
	"dcfs/models/credentials"
	"github.com/google/uuid"
	"time"
)

type dummyDisk struct {
	UUID   uuid.UUID
	Volume *Volume
	Name   string
}

/* Mandatory Disk interface implementations */

func (d *dummyDisk) Upload(blockMetadata *apicalls.BlockMetadata) *apicalls.ErrorWrapper {
	panic("Unimplemented")
}

func (d *dummyDisk) Download(blockMetadata *apicalls.BlockMetadata) *apicalls.ErrorWrapper {
	panic("Unimplemented")
}

func (d *dummyDisk) Rename(blockMetadata *apicalls.BlockMetadata) *apicalls.ErrorWrapper {
	panic("Unimplemented")
}

func (d *dummyDisk) Remove(blockMetadata *apicalls.BlockMetadata) *apicalls.ErrorWrapper {
	panic("Unimplemented")
}

func (d *dummyDisk) SetVolume(volume *Volume) {
	d.Volume = volume
}

func (d *dummyDisk) GetVolume() *Volume {
	return d.Volume
}

func (d *dummyDisk) SetUUID(uuid uuid.UUID) {
	d.UUID = uuid
}

func (d *dummyDisk) GetUUID() uuid.UUID {
	return d.UUID
}

func (d *dummyDisk) SetName(name string) {
	d.Name = name
}

func (d *dummyDisk) GetName() string {
	return d.Name
}

func (d *dummyDisk) GetThroughput() int {
	panic("Unimplemented")
}

func (d *dummyDisk) GetCredentials() credentials.Credentials {
	panic("Unimplemented")
}

func (d *dummyDisk) SetCredentials(credentials credentials.Credentials) {
	panic("Unimplemented")
}

func (d *dummyDisk) CreateCredentials(c string) {
	panic("Unimplemented")
}

func (d *dummyDisk) GetProviderUUID() uuid.UUID {
	panic("Unimplemented")
}

func (d *dummyDisk) GetProviderFreeSpace() (uint64, string) {
	panic("Unimplemented")
}

func (d *dummyDisk) GetTotalSize() uint64 {
	panic("Unimplemented")
}

func (d *dummyDisk) SetUsedSpace(usage uint64) {
	panic("Unimplemented")
}

func (d *dummyDisk) GetUsedSpace() uint64 {
	panic("Unimplemented")
}

func (d *dummyDisk) GetProviderSpace() (uint64, uint64, string) {
	panic("Unimplemented")
}

func (d *dummyDisk) SetTotalSpace(quota uint64) {
	panic("Unimplemented")
}

func (d *dummyDisk) GetTotalSpace() uint64 {
	panic("Unimplemented")
}

func (d *dummyDisk) UpdateUsedSpace(change int64) {
	panic("Unimplemented")
}

func (d *dummyDisk) SetCreationTime(creationTime time.Time) {
	panic("Unimplemented")
}

func (d *dummyDisk) GetCreationTime() time.Time {
	panic("Unimplemented")
}

func (d *dummyDisk) GetDiskDBO(userUUID uuid.UUID, providerUUID uuid.UUID, volumeUUID uuid.UUID) dbo.Disk {
	panic("Unimplemented")
}

func (d *dummyDisk) Delete() (string, error) {
	panic("Unimplemented")
}
