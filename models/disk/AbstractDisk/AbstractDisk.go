package AbstractDisk

import (
	"dcfs/apicalls"
	"dcfs/db"
	"dcfs/db/dbo"
	"dcfs/models"
	"dcfs/models/credentials"
	"dcfs/util/logger"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"strconv"
	"time"
)

type AbstractDisk struct {
	Disk        models.Disk
	UUID        uuid.UUID
	Credentials credentials.Credentials
	BlockSize   int
	Volume      *models.Volume
	Name        string

	CreationTime time.Time

	IsVirtual       bool
	VirtualDiskUUID uuid.UUID

	Size      uint64
	UsedSpace uint64
}

/* Mandatory Disk interface implementations */

func (d *AbstractDisk) Upload(bm *apicalls.BlockMetadata) *apicalls.ErrorWrapper {
	panic("Unimplemented abstract method")
}

func (d *AbstractDisk) Download(bm *apicalls.BlockMetadata) *apicalls.ErrorWrapper {
	panic("Unimplemented abstract method")
}

func (d *AbstractDisk) Rename(bm *apicalls.BlockMetadata) *apicalls.ErrorWrapper {
	panic("Unimplemented abstract method")
}

func (d *AbstractDisk) Remove(bm *apicalls.BlockMetadata) *apicalls.ErrorWrapper {
	panic("Unimplemented abstract method")
}

func (d *AbstractDisk) SetUUID(UUID uuid.UUID) {
	logger.Logger.Debug("disk", "Set uuid for a disk to: ", UUID.String())
	d.UUID = UUID
}

func (d *AbstractDisk) GetUUID() uuid.UUID {
	return d.UUID
}

func (d *AbstractDisk) SetVolume(volume *models.Volume) {
	logger.Logger.Debug("disk", "Set the volume of a disk object to: ", volume.UUID.String(), ".")
	d.Volume = volume
}

func (d *AbstractDisk) GetVolume() *models.Volume {
	return d.Volume
}

func (d *AbstractDisk) SetName(name string) {
	logger.Logger.Debug("disk", "Changed the name of a disk object from: ", d.GetName(), " to: ", name, ".")
	d.Name = name
}

func (d *AbstractDisk) GetName() string {
	return d.Name
}

func (d *AbstractDisk) GetCredentials() credentials.Credentials {
	return d.Credentials
}

func (d *AbstractDisk) SetCredentials(c credentials.Credentials) {
	logger.Logger.Debug("disk", "Changed the credentials of a disk object named: ", d.GetName())
	d.Credentials = c
}

func (d *AbstractDisk) CreateCredentials(credentials string) {
	panic("Unimplemented abstract method")
}

func (d *AbstractDisk) GetProviderUUID() uuid.UUID {
	panic("Unimplemented abstract method")
}

func (d *AbstractDisk) GetProviderSpace() (uint64, uint64, string) {
	panic("Unimplemented abstract method")
}

func (d *AbstractDisk) SetCreationTime(creationTime time.Time) {
	logger.Logger.Debug("disk", "Set the creation time of a disk object named: ", d.GetName(), " to: ", creationTime.String(), ".")
	d.CreationTime = creationTime
}

func (d *AbstractDisk) GetCreationTime() time.Time {
	return d.CreationTime
}

func (d *AbstractDisk) SetIsVirtualFlag(isVirtual bool) {
	d.IsVirtual = isVirtual
}

func (d *AbstractDisk) GetIsVirtualFlag() bool {
	return d.IsVirtual
}

func (d *AbstractDisk) SetVirtualDiskUUID(uuid uuid.UUID) {
	d.VirtualDiskUUID = uuid
}

func (d *AbstractDisk) GetVirtualDiskUUID() uuid.UUID {
	return d.VirtualDiskUUID
}

func (d *AbstractDisk) SetTotalSpace(quota uint64) {
	logger.Logger.Debug("disk", "Set total space of a disk object named: ", d.GetName(), " to: ", strconv.FormatUint(quota, 10), ".")
	d.Size = quota
}

func (d *AbstractDisk) GetTotalSpace() uint64 {
	return d.Size
}

func (d *AbstractDisk) SetUsedSpace(usage uint64) {
	logger.Logger.Debug("disk", "Set used space of a disk object named: ", d.GetName(), " to: ", strconv.FormatUint(usage, 10), ".")
	d.UsedSpace = usage
}

func (d *AbstractDisk) GetUsedSpace() uint64 {
	return d.UsedSpace
}

func (d *AbstractDisk) UpdateUsedSpace(change int64) {
	// Update internal disk usage
	if change > 0 {
		d.UsedSpace += uint64(change)
	} else {
		d.UsedSpace -= uint64(-change)
	}

	// Update disk usage in database
	diskDBO := d.GetDiskDBO(uuid.Nil, uuid.Nil, uuid.Nil)
	db.DB.DatabaseHandle.Model(&diskDBO).Update("used_space", d.UsedSpace)
	logger.Logger.Debug("disk", "Updated used space of a disk object named: ", d.GetName(), " to: ", strconv.FormatUint(d.UsedSpace, 10), ".")
}

func (d *AbstractDisk) GetDiskDBO(userUUID uuid.UUID, providerUUID uuid.UUID, volumeUUID uuid.UUID) dbo.Disk {
	credentials := ""
	if d.Credentials != nil {
		credentials = d.Credentials.ToString()
	}

	var provider dbo.Provider
	var user dbo.User
	var volume dbo.Volume

	err := db.DB.DatabaseHandle.Where("uuid = ?", providerUUID).First(&provider).Error
	if err != nil {
		logger.Logger.Warning("disk", "could not fetch the provider object with uuid: ", providerUUID.String(), ".")
	}

	err = db.DB.DatabaseHandle.Where("uuid = ?", userUUID).First(&user).Error
	if err != nil {
		logger.Logger.Warning("disk", "could not fetch the user object with uuid: ", userUUID.String(), ".")
	}

	err = db.DB.DatabaseHandle.Where("uuid = ?", volumeUUID).First(&volume).Error
	if err != nil {
		logger.Logger.Warning("disk", "could not fetch the volume object with uuid: ", volumeUUID.String(), ".")
	}

	return dbo.Disk{
		AbstractDatabaseObject: dbo.AbstractDatabaseObject{UUID: d.UUID},
		UserUUID:               userUUID,
		ProviderUUID:           providerUUID,
		VolumeUUID:             volumeUUID,
		Credentials:            credentials,
		Name:                   d.Name,
		TotalSpace:             d.Size,
		UsedSpace:              d.UsedSpace,
		IsVirtual:              d.IsVirtual,
		VirtualDiskUUID:        d.VirtualDiskUUID,
		User:                   user,
		Volume:                 volume,
		Provider:               provider,
	}
}

func (d *AbstractDisk) AssignDisk(disk models.Disk) {
	panic("Not supported for real disk")
}

func (d *AbstractDisk) IsReady(ctx *gin.Context) bool {
	// check if it is possible to connect to a disk
	if d.GetCredentials().Authenticate(&apicalls.CredentialsAuthenticateMetadata{
		Ctx:      ctx,
		Config:   nil,
		DiskUUID: d.GetUUID(),
	}) == nil {
		return false
	}

	return true
}

func (d *AbstractDisk) GetResponse(_disk *dbo.Disk, ctx *gin.Context) *models.DiskResponse {
	return &models.DiskResponse{
		Disk:    *_disk,
		Array:   nil,
		IsReady: d.IsReady(ctx),
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
