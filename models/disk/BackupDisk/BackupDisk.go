package BackupDisk

import (
	"dcfs/apicalls"
	"dcfs/constants"
	"dcfs/db/dbo"
	"dcfs/models"
	"dcfs/models/credentials"
	"dcfs/models/disk/AbstractDisk"
	"dcfs/util/checksum"
	"dcfs/util/logger"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"sync"
	"time"
)

type BackupDisk struct {
	abstractDisk AbstractDisk.AbstractDisk

	firstDisk  models.Disk
	secondDisk models.Disk
}

/* Mandatory Disk interface methods */

func (d *BackupDisk) Upload(blockMetadata *apicalls.BlockMetadata) *apicalls.ErrorWrapper {
	// TODO: Verify that onedrive doesn't modify the content

	// Create a copy of the block contents
	contents := &blockMetadata.Content

	// Create a copy of the api call
	blockMetadata1 := *blockMetadata
	blockMetadata2 := *blockMetadata

	blockMetadata2.Content = *contents

	var emptyCallback = func(uuid.UUID, *int) {
	}
	blockMetadata1.CompleteCallback = emptyCallback
	blockMetadata2.CompleteCallback = emptyCallback

	var status1 int
	var status2 int
	blockMetadata1.Status = &status1
	blockMetadata2.Status = &status2

	// Prepare for upload
	var waitGroup sync.WaitGroup
	var err1 *apicalls.ErrorWrapper
	var err2 *apicalls.ErrorWrapper

	waitGroup.Add(2)

	// Upload to the first disk
	go func() {
		defer waitGroup.Done()
		err1 = d.firstDisk.Upload(&blockMetadata1)
	}()

	// Upload to the second disk
	go func() {
		defer waitGroup.Done()
		err2 = d.secondDisk.Upload(&blockMetadata2)
	}()

	// Wait for the upload to finish
	waitGroup.Wait()

	// Check for errors
	if err1 != nil {
		logger.Logger.Error("disk", "Cannot upload to the first disk, got an error: ", err1.Error.Error())
	}

	if err2 != nil {
		logger.Logger.Error("disk", "Cannot upload to the second disk, got an error: ", err2.Error.Error())
	}

	if err1 != nil || err2 != nil {
		return apicalls.CreateErrorWrapper(constants.REMOTE_FAILED_JOB, "Cannot upload to at least one of the backup disks.")
	}

	// Call the original callback
	blockMetadata.CompleteCallback(blockMetadata.FileUUID, blockMetadata.Status)

	logger.Logger.Debug("disk", "Successfully uploaded the block to backup disk: ", blockMetadata.UUID.String(), ".")
	return nil
}

func (d *BackupDisk) Download(blockMetadata *apicalls.BlockMetadata) *apicalls.ErrorWrapper {
	// Create a copy of the api call
	blockMetadata1 := *blockMetadata
	blockMetadata2 := *blockMetadata

	var contents1 = make([]byte, len(*blockMetadata.Content))
	var contents2 = make([]byte, len(*blockMetadata.Content))
	blockMetadata2.Content = &contents1
	blockMetadata2.Content = &contents2

	var emptyCallback = func(uuid.UUID, *int) {
	}
	blockMetadata1.CompleteCallback = emptyCallback
	blockMetadata2.CompleteCallback = emptyCallback

	var status1 int
	var status2 int
	blockMetadata1.Status = &status1
	blockMetadata2.Status = &status2

	// Prepare for download
	var waitGroup sync.WaitGroup
	var err1 *apicalls.ErrorWrapper
	var err2 *apicalls.ErrorWrapper
	var checksum1 string
	var checksum2 string

	waitGroup.Add(2)

	// Download from the first disk
	go func() {
		defer waitGroup.Done()
		err1 = d.firstDisk.Download(&blockMetadata1)
		if err1 == nil {
			checksum1 = checksum.CalculateChecksum(*blockMetadata1.Content)
		}
	}()

	// Download from the second disk
	go func() {
		defer waitGroup.Done()
		err2 = d.secondDisk.Download(&blockMetadata2)
		if err2 == nil {
			checksum2 = checksum.CalculateChecksum(*blockMetadata2.Content)
		}
	}()

	// Wait for the download to finish
	waitGroup.Wait()

	// Check for errors
	if err1 != nil {
		logger.Logger.Error("disk", "Cannot download from the first disk, got an error: ", err1.Error.Error())
	}

	if err2 != nil {
		logger.Logger.Error("disk", "Cannot download from the second disk, got an error: ", err2.Error.Error())
	}

	// If both disks worked and one of the blocks is corrupted, make attempt to recover
	if err1 == nil && err2 == nil {
		d.fixBlock(blockMetadata, *blockMetadata1.Content, *blockMetadata2.Content, checksum1, checksum2)
	}

	// Return block
	if checksum1 == blockMetadata.Checksum || checksum2 == blockMetadata.Checksum {
		// Return block with the correct checksum if possible
		if err1 == nil && checksum1 == blockMetadata.Checksum {
			logger.Logger.Debug("disk", "Successfully downloaded the block from the first disk: ", blockMetadata.UUID.String(), ".")
			blockMetadata.Content = blockMetadata1.Content
			blockMetadata.CompleteCallback(blockMetadata.FileUUID, blockMetadata.Status)
			return nil
		}

		if err2 == nil && checksum2 == blockMetadata.Checksum {
			logger.Logger.Debug("disk", "Successfully downloaded the block from the second disk: ", blockMetadata.UUID.String(), ".")
			blockMetadata.Content = blockMetadata2.Content
			blockMetadata.CompleteCallback(blockMetadata.FileUUID, blockMetadata.Status)
			return nil
		}
	} else {
		// Return block with the wrong checksum if one of the disks is available
		if err1 == nil {
			logger.Logger.Debug("disk", "Downloaded corrupted block from the first disk: ", blockMetadata.UUID.String(), ".")
			blockMetadata.Content = blockMetadata1.Content
			blockMetadata.CompleteCallback(blockMetadata.FileUUID, blockMetadata.Status)
			return nil
		}

		if err2 == nil {
			logger.Logger.Debug("disk", "Downloaded corrupted block from the second disk: ", blockMetadata.UUID.String(), ".")
			blockMetadata.Content = blockMetadata2.Content
			blockMetadata.CompleteCallback(blockMetadata.FileUUID, blockMetadata.Status)
			return nil
		}

		// Return error if both disks failed
		if err1 != nil && err2 != nil {
			return apicalls.CreateErrorWrapper(constants.REMOTE_FAILED_JOB, "Cannot download from both of the backup disks.")
		}
	}

	return nil
}

func (d *BackupDisk) Remove(blockMetadata *apicalls.BlockMetadata) *apicalls.ErrorWrapper {
	// Create a copy of the api call
	blockMetadata1 := *blockMetadata
	blockMetadata2 := *blockMetadata

	var emptyCallback = func(uuid.UUID, *int) {
	}
	blockMetadata1.CompleteCallback = emptyCallback
	blockMetadata2.CompleteCallback = emptyCallback

	var status1 int
	var status2 int
	blockMetadata1.Status = &status1
	blockMetadata2.Status = &status2

	// Prepare for download
	var waitGroup sync.WaitGroup
	var err1 *apicalls.ErrorWrapper
	var err2 *apicalls.ErrorWrapper

	waitGroup.Add(2)

	// Remove from the first disk
	go func() {
		defer waitGroup.Done()
		err1 = d.firstDisk.Remove(&blockMetadata1)
	}()

	// Remove from the second disk
	go func() {
		defer waitGroup.Done()
		err2 = d.secondDisk.Remove(&blockMetadata2)
	}()

	// Wait for the removal to finish
	waitGroup.Wait()

	// Check for errors
	if err1 != nil || err2 != nil {
		if err1 != nil {
			logger.Logger.Error("disk", "Cannot remove from the first disk, got an error: ", err1.Error.Error())
		}

		if err2 != nil {
			logger.Logger.Error("disk", "Cannot remove from the second disk, got an error: ", err2.Error.Error())
		}

		return apicalls.CreateErrorWrapper(constants.REMOTE_FAILED_JOB, "Cannot remove from at least one of the backup disks.")
	}

	// Call the original callback
	blockMetadata.CompleteCallback(blockMetadata.FileUUID, blockMetadata.Status)

	return nil
}

func (d *BackupDisk) SetUUID(uuid uuid.UUID) {
	d.abstractDisk.SetUUID(uuid)
}

func (d *BackupDisk) GetUUID() uuid.UUID {
	return d.abstractDisk.GetUUID()
}

func (d *BackupDisk) SetVolume(volume *models.Volume) {
	d.abstractDisk.SetVolume(volume)
}

func (d *BackupDisk) GetVolume() *models.Volume {
	return d.abstractDisk.GetVolume()
}

func (d *BackupDisk) SetName(name string) {
	panic("Not supported for backup disk")
}

func (d *BackupDisk) GetName() string {
	return "Virtual RAID1 backup disk"
}

func (d *BackupDisk) GetCredentials() credentials.Credentials {
	panic("Not supported for backup disk")
}

func (d *BackupDisk) SetCredentials(credentials credentials.Credentials) {
	panic("Not supported for backup disk")
}

func (d *BackupDisk) CreateCredentials(c string) {
	panic("Not supported for backup disk")
}

func (d *BackupDisk) SetCreationTime(creationTime time.Time) {
	d.abstractDisk.SetCreationTime(creationTime)
}

func (d *BackupDisk) GetCreationTime() time.Time {
	return d.abstractDisk.GetCreationTime()
}

func (d *BackupDisk) GetProviderUUID() uuid.UUID {
	panic("Not supported for backup disk")
}

func (d *BackupDisk) SetIsVirtualFlag(isVirtual bool) {
	d.abstractDisk.SetIsVirtualFlag(isVirtual)
}

func (d *BackupDisk) GetIsVirtualFlag() bool {
	return d.abstractDisk.GetIsVirtualFlag()
}

func (d *BackupDisk) SetVirtualDiskUUID(uuid uuid.UUID) {
	d.abstractDisk.SetVirtualDiskUUID(uuid)
}

func (d *BackupDisk) GetVirtualDiskUUID() uuid.UUID {
	return d.abstractDisk.GetVirtualDiskUUID()
}

func (d *BackupDisk) GetProviderSpace() (uint64, uint64, string) {
	// Retrieve provider space from both disks
	used1, total1, result1 := d.firstDisk.GetProviderSpace()
	used2, total2, result2 := d.secondDisk.GetProviderSpace()

	// Return not supported if one of the disks does not support it
	if result1 == constants.OPERATION_NOT_SUPPORTED || result2 == constants.OPERATION_NOT_SUPPORTED {
		return 0, 0, constants.OPERATION_NOT_SUPPORTED
	}

	// Return error if one of the disks returned an error
	if result1 != constants.SUCCESS || result2 != constants.SUCCESS {
		return 0, 0, result1
	}

	// Return available space in both disks
	var used uint64
	var total uint64

	if used1 > used2 {
		used = used1
	} else {
		used = used2
	}

	if total1 < total2 {
		total = total1
	} else {
		total = total2
	}

	return used, total, constants.SUCCESS
}

func (d *BackupDisk) SetTotalSpace(quota uint64) {
	d.firstDisk.SetTotalSpace(quota)
	d.secondDisk.SetTotalSpace(quota)
}

func (d *BackupDisk) GetTotalSpace() uint64 {
	var space1 = d.firstDisk.GetTotalSpace()
	var space2 = d.secondDisk.GetTotalSpace()

	if space1 < space2 {
		return space1
	} else {
		return space2
	}
}

func (d *BackupDisk) SetUsedSpace(usage uint64) {
	d.firstDisk.SetUsedSpace(usage)
	d.secondDisk.SetUsedSpace(usage)
}

func (d *BackupDisk) GetUsedSpace() uint64 {
	return d.firstDisk.GetUsedSpace()
}

func (d *BackupDisk) UpdateUsedSpace(change int64) {
	d.firstDisk.UpdateUsedSpace(change)
	d.secondDisk.UpdateUsedSpace(change)
}

func (d *BackupDisk) GetDiskDBO(userUUID uuid.UUID, providerUUID uuid.UUID, volumeUUID uuid.UUID) dbo.Disk {
	panic("Not supported for backup disk")
}

func (d *BackupDisk) AssignDisk(disk models.Disk) {
	if d.firstDisk == nil {
		d.firstDisk = disk
	} else if d.secondDisk == nil {
		d.secondDisk = disk
	} else {
		// If both disks are already assigned, ignore the new disk
		logger.Logger.Error("disk", "Cannot assign disk to backup disk, both disks are already assigned.")
		return
	}
}

func (d *BackupDisk) IsReady(ctx *gin.Context) bool {
	return d.firstDisk.IsReady(ctx) && d.secondDisk.IsReady(ctx)
}

func (d *BackupDisk) GetResponse(_disk *dbo.Disk, ctx *gin.Context) *models.DiskResponse {
	arr := make([]models.DiskResponse, 0)

	firstDiskDBO := d.firstDisk.GetDiskDBO(_disk.UserUUID, d.firstDisk.GetProviderUUID(), _disk.VolumeUUID)
	secondDiskDBO := d.secondDisk.GetDiskDBO(_disk.UserUUID, d.secondDisk.GetProviderUUID(), _disk.VolumeUUID)

	arr = append(arr, *d.firstDisk.GetResponse(&firstDiskDBO, ctx))
	arr = append(arr, *d.secondDisk.GetResponse(&secondDiskDBO, ctx))

	return &models.DiskResponse{
		Disk:    *_disk,
		Array:   arr,
		IsReady: d.IsReady(ctx),
	}
}

func (d *BackupDisk) fixBlock(blockMetadata *apicalls.BlockMetadata, firstContents []uint8, secondContents []uint8, firstChecksum string, secondChecksum string) {
	var err *apicalls.ErrorWrapper
	var targetDisk models.Disk

	// Verify if the action should be performed
	if firstChecksum == blockMetadata.Checksum && secondChecksum == blockMetadata.Checksum {
		// Both disks have the correct block
		return
	}
	if firstChecksum != blockMetadata.Checksum && secondChecksum != blockMetadata.Checksum {
		// Unrecoverable, both disks have the wrong block
		logger.Logger.Error("disk", "RAID10 recovery failed: both disks have corrupted block ", blockMetadata.UUID.String(), ".")
		return
	}

	// Create copy of the api call
	_blockMetadata := *blockMetadata
	_blockMetadata.Content = nil

	_blockMetadata.CompleteCallback = func(uuid.UUID, *int) {
	}

	var status int
	_blockMetadata.Status = &status

	var contents []uint8

	// Set the target disk to disk with the wrong version of the block
	if firstChecksum == blockMetadata.Checksum {
		targetDisk = d.secondDisk
		contents = make([]uint8, len(firstContents))
		copy(contents, firstContents)
	} else {
		targetDisk = d.firstDisk
		contents = make([]uint8, len(secondContents))
		copy(contents, secondContents)
	}

	// Remove the invalid block from the target disk
	err = targetDisk.Remove(&_blockMetadata)
	if err != nil {
		logger.Logger.Error("disk", "RAID10 recovery failed: cannot remove invalid block ", blockMetadata.UUID.String(), "from disk ", targetDisk.GetUUID().String(), ".")
		return
	}

	// Upload the correct block to the target disk
	_blockMetadata.Content = &contents
	err = targetDisk.Upload(&_blockMetadata)
	if err != nil {
		logger.Logger.Error("disk", "RAID10 recovery failed: cannot upload valid block ", blockMetadata.UUID.String(), "to disk ", targetDisk.GetUUID().String(), ".")
		return
	}

	logger.Logger.Warning("disk", "RAID10 recovery completed: disk ", targetDisk.GetUUID().String(), " now has the correct block ", blockMetadata.UUID.String(), ".")
	return
}

/* Factory methods */

func NewBackupDisk() *BackupDisk {
	var d *BackupDisk = new(BackupDisk)
	d.abstractDisk.Disk = d
	d.abstractDisk.UUID = uuid.New()

	d.abstractDisk.IsVirtual = true
	d.abstractDisk.VirtualDiskUUID = uuid.Nil

	return d
}

func init() {
	models.DiskTypesRegistry[constants.PROVIDER_TYPE_RAID1] = func() models.Disk { return NewBackupDisk() }
}
