package models

import (
	"dcfs/apicalls"
	"dcfs/constants"
	"dcfs/db"
	"dcfs/db/dbo"
	"dcfs/models/credentials"
	"dcfs/util/logger"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"net/http/httptest"
	"strconv"
	"time"
)

var RootUUID uuid.UUID
var DiskTypesRegistry map[int]func() Disk = make(map[int]func() Disk)
var DiskReadinessRegistry map[int]func(d Disk) DiskReadiness = make(map[int]func(d Disk) DiskReadiness)
var ProviderTypesRegistry map[int]func() = make(map[int]func())

type Disk interface {
	// Upload - upload a block to the disk
	//
	// params:
	//   - bm *apicalls.BlockMetadata - metadata for the block operation
	//
	// return type:
	//   - *apicalls.ErrorWrapper - nil if operation was successful, error otherwise
	Upload(bm *apicalls.BlockMetadata) *apicalls.ErrorWrapper

	// Download - download a block from the disk
	//
	// params:
	//   - bm *apicalls.BlockMetadata - metadata for the block operation
	//
	// return type:
	//   - *apicalls.ErrorWrapper - nil if operation was successful, error otherwise
	Download(bm *apicalls.BlockMetadata) *apicalls.ErrorWrapper

	// Remove - remove a block from the disk
	//
	// params:
	//   - bm *apicalls.BlockMetadata - metadata for the block operation
	//
	// return type:
	//   - *apicalls.ErrorWrapper - nil if operation was successful, error otherwise
	Remove(bm *apicalls.BlockMetadata) *apicalls.ErrorWrapper

	// SetUUID - set disk's UUID
	//
	// params:
	//   - uuid.UUID: disk's UUID
	SetUUID(uuid.UUID)

	// GetUUID - get disk's UUID
	//
	// return type:
	//   - uuid.UUID: disk's UUID
	GetUUID() uuid.UUID

	// SetVolume - set disk's volume
	//
	// params:
	//   - *Volume - pointer to volume object the disk should belong to
	SetVolume(volume *Volume)

	// GetVolume - get volume the disk is attached to
	//
	// return type:
	//   - *Volume - pointer to volume object the disk belongs to
	GetVolume() *Volume

	// GetName - get disk name
	//
	// return type:
	//   - string - disk name
	GetName() string

	// SetName - set disk name
	//
	// params:
	//   - name string - disk name
	SetName(name string)

	// GetCredentials - get credentials for the disk
	//
	// return type:
	//   - credentials.Credentials - credentials object for the disk
	GetCredentials() credentials.Credentials

	// SetCredentials - set credentials for the disk
	//
	// params:
	//   - credentials.Credentials - credentials object for the disk
	SetCredentials(credentials.Credentials)

	// CreateCredentials - create credentials for the disk based on provided connection string
	//
	// params:
	//   - credentials string - connection string for the disk
	CreateCredentials(credentials string)

	// GetProviderUUID - get disk provider uuid
	//
	// return type:
	//   - uuid.UUID - disk provider uuid
	GetProviderUUID() uuid.UUID

	// SetCreationTime - get disk's creation time
	//
	// params:
	//   - creationTime time.Time - creation time of the disk
	SetCreationTime(creationTime time.Time)

	// GetCreationTime - get disk's creation time
	//
	// params:
	//   - time.Time - creation time of the disk
	GetCreationTime() time.Time

	// SetIsVirtualFlag - set is virtual flag
	//
	// params:
	//   - isVirtual bool - true if disk is virtual, false otherwise
	SetIsVirtualFlag(isVirtual bool)

	// GetIsVirtualFlag - get is virtual flag
	//
	// params:
	//   - bool - true if disk is virtual, false otherwise
	GetIsVirtualFlag() bool

	// SetVirtualDiskUUID - set disk's virtual disk uuid
	//
	// params:
	//   - uuid uuid.UUID - uuid of the virtual disk the disk should be attached to
	SetVirtualDiskUUID(uuid uuid.UUID)

	// GetVirtualDiskUUID - get disk's virtual disk uuid
	//
	// return type:
	//   - uuid.UUID - uuid of the virtual disk the disk is attached to
	GetVirtualDiskUUID() uuid.UUID

	// GetProviderSpace - get disk space information from cloud provider
	//
	// return type:
	//   - uint64 - used space in bytes
	//   - uint64 - total space in bytes
	//   - string - completion code, constants.SUCCESS if operation was successful, error code if operation failed,
	//              constants.OPERATION_NOT_SUPPORTED if provider does not support disk space information retrieval
	GetProviderSpace() (uint64, uint64, string)

	// SetTotalSpace - set disk total space as an internal data
	//
	// params:
	//   - usage uint64 - new total space in bytes
	SetTotalSpace(quota uint64)

	// GetTotalSpace - get disk total space based on the internal data
	//
	// return type:
	//   - uint64 - total space in bytes
	GetTotalSpace() uint64

	// SetUsedSpace - set disk used space as an internal data
	//
	// params:
	//   - usage uint64 - new used space in bytes
	SetUsedSpace(usage uint64)

	// GetUsedSpace - get disk used space based on the internal data
	//
	// return type:
	//   - uint64 - used space in bytes
	GetUsedSpace() uint64

	// UpdateUsedSpace - set disk used space as an internal data
	//
	// params:
	//   - change int64 - change in used space in bytes (positive for new data or negative for removed data)
	UpdateUsedSpace(change int64)

	// GetDiskDBO - get disk dbo object
	//
	// params:
	//   - userUUID uuid.UUID - uuid of the owner of the disk
	//   - provider uuid.UUID - uuid of the provider of the disk
	//   - volumeUUID uuid.UUID - uuid of the volume the disk is attached to
	//
	// return type:
	//   - dbo.Disk - disk dbo object
	GetDiskDBO(userUUID uuid.UUID, providerUUID uuid.UUID, volumeUUID uuid.UUID) dbo.Disk

	// GetReadiness - get disk readiness information for performing operations
	//
	// return type:
	//   - DiskReadiness - disk readiness information
	GetReadiness() DiskReadiness

	// GetResponse - get disk information to be returned to the API
	//
	// params:
	//   - _disk *dbo.Disk - disk information
	//   - ctx *gin.Context - gin API context
	//
	// return type:
	//   - *DiskResponse - disk information
	GetResponse(_disk *dbo.Disk, ctx *gin.Context) *DiskResponse
}

type DiskResponse struct {
	dbo.Disk
	Array   []DiskResponse `json:"array"`
	IsReady bool           `json:"isReady"`
}

type VirtualDisk interface {
	// AssignDisk - assign a real disk to the virtual disk. If all disks are assigned, request is ignored by the virtual disk.
	//
	// params:
	//   - disk models.Disk - disk to be attached to the virtual disk
	AssignDisk(disk Disk)

	// ReplaceDisk - replace provided disk in the virtual disk with a new disk
	//
	// params:
	//   - disk models.Disk - disk to be disattached from the virtual disk
	//   - newDisk models.Disk - disk to be attached to the virtual disk in place of the old disk
	//   - block []dbo.Block - list of blocks located on the disk (to be transferred to the new disk)
	//
	// return type:
	//   - string - completion code, constants.SUCCESS if operation was successful, otherwise an error code
	ReplaceDisk(oldDisk Disk, newDisk Disk, blocks []dbo.Block) string
}

type CreateDiskMetadata struct {
	Disk   *dbo.Disk
	Volume *Volume
}

// CreateDisk - create new disk model based on provided metadata
//
// This function creates disk model used internally by backend based on
// provided metadata.
//
// params:
//   - cdm CreateDiskMetadata: disk data
//
// return type:
//   - models.Disk: created disk model, nil if provider is invalid
func CreateDisk(cdm CreateDiskMetadata) Disk {
	if DiskTypesRegistry[cdm.Disk.Provider.Type] == nil || cdm.Disk.Provider.Type < 0 {
		return nil
	}
	var disk Disk = DiskTypesRegistry[cdm.Disk.Provider.Type]()

	disk.SetVolume(cdm.Volume)
	disk.CreateCredentials(cdm.Disk.Credentials)
	disk.SetUUID(cdm.Disk.UUID)
	disk.SetName(cdm.Disk.Name)
	disk.SetUsedSpace(cdm.Disk.UsedSpace)
	disk.SetTotalSpace(cdm.Disk.TotalSpace)
	disk.SetCreationTime(cdm.Disk.CreatedAt)
	disk.SetIsVirtualFlag(cdm.Disk.IsVirtual)
	disk.SetVirtualDiskUUID(cdm.Disk.VirtualDiskUUID)
	cdm.Volume.AddDisk(disk.GetUUID(), disk)

	logger.Logger.Debug("disk", "Successfully created a new disk.")

	return disk
}

// CreateDiskFromUUID - retrieve disk from database and create disk model
//
// params:
//   - uuid uuid.UUID: disk data
//
// return type:
//   - models.Disk: created disk model, nil if database operations failed
func CreateDiskFromUUID(uuid uuid.UUID) Disk {
	var disk dbo.Disk
	var volume *Volume

	// Retrieve disk data from database
	err := db.DB.DatabaseHandle.Where("uuid = ?", uuid).Find(&disk).Error
	if err != nil {
		return nil
	}

	// Load volume from database to transport
	volume = Transport.GetVolume(disk.VolumeUUID)
	if volume == nil {
		return nil
	}

	return volume.GetDisk(disk.UUID)
}

// ComputeFreeSpace - compute free space on disk
//
// This function calculates free space on disk based on two data sources:
// 1. Disk quota provided by the user and space used by DCFS (stored in database),
// 2. Disk quota provided by the provider and used space reported by provider.
// If disk quota provided by user is smaller than real disk quota, then free space
// will be limited to disk quota provided by user. If real usage of cloud drive
// exceeds theoretical free space (calculated based on user-provided quota and local
// data sum), then free space will be limited to real available space.
// In case of lack support of obtaining provider-based data (indicated by
// constants.OPERATION_NOT_SUPPORTED), only user-provided (local) data will be used.
// This is the case in FTP drive and SFTP drive, if server doesn't support VSTATS
// SFTP extension.
//
// params:
//   - d models.Disk: disk to compute free space for
//
// return type:
//   - uint64: free space in bytes
func ComputeFreeSpace(d Disk) uint64 {
	var userDefinedSpace uint64
	var providerDefinedSpace uint64
	var freeSpace uint64

	// Get needed values
	userTotalSpace := d.GetTotalSpace()
	userUsedSpace := d.GetUsedSpace()
	providerUsedSpace, providerTotalSpace, errCode := d.GetProviderSpace()

	// Compute free space based on disk quota provided by the user
	userDefinedSpace = userTotalSpace - userUsedSpace
	if userTotalSpace < userUsedSpace {
		userDefinedSpace = 0
	}

	// Compute free space based on the disk quota provided by the provider
	if errCode == constants.SUCCESS {
		providerDefinedSpace = providerTotalSpace - providerUsedSpace
		if providerTotalSpace < providerUsedSpace {
			providerDefinedSpace = 0
		}
	} else if errCode == constants.OPERATION_NOT_SUPPORTED {
		// In case the provider does not support this operation,
		// we assume that the real disk space is equal to user defined space
		providerDefinedSpace = userDefinedSpace
	} else {
		providerDefinedSpace = 0
	}

	// Return the minimum of user defined space and provider defined space
	freeSpace = userDefinedSpace
	if providerDefinedSpace < freeSpace {
		freeSpace = providerDefinedSpace
	}

	logger.Logger.Debug("disk", "Free space on disk", d.GetName(), "is", strconv.FormatUint(freeSpace, 10), "bytes", " (user defined:", strconv.FormatUint(userDefinedSpace, 10), "bytes, provider defined:", strconv.FormatUint(providerDefinedSpace, 10), "bytes, provider total:", strconv.FormatUint(providerTotalSpace, 10), "bytes)")

	return freeSpace
}

// MeasureDiskThroughput - measure disk throughput and calculate throughput weight
//
// This function measures disk throughput and calculates throughput weight for
// throughput partitioner. Throughput weight is calculated as an average of
// upload and download time for sample block (size of volume block size).
// Throughput time is measured in miliseconds.
//
// params:
//   - d models.Disk: disk to measure throughput
//
// return type:
//   - int: throughput weight of the disk
func MeasureDiskThroughput(d Disk) int {
	var uploadTime time.Duration
	var downloadTime time.Duration
	var throughput int

	// Prepare test context
	writer := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(writer)

	// Prepare test block
	var status int
	var size = constants.DEFAULT_VOLUME_BLOCK_SIZE

	var content []uint8
	content = make([]uint8, size)
	for i := 0; i < size; i++ {
		content[i] = 1
	}

	var blockMetadata *apicalls.BlockMetadata = new(apicalls.BlockMetadata)
	blockMetadata.Ctx = ctx
	blockMetadata.FileUUID = uuid.Nil
	blockMetadata.Content = &content
	blockMetadata.UUID = uuid.New()
	blockMetadata.Size = int64(size)
	blockMetadata.Status = &status
	blockMetadata.CompleteCallback = func(UUID uuid.UUID, status *int) {
	}

	// Measure upload time
	uploadStart := time.Now()
	d.Upload(blockMetadata)
	uploadEnd := time.Now()
	uploadTime = uploadEnd.Sub(uploadStart)

	// Measure download time
	downloadStart := time.Now()
	d.Download(blockMetadata)
	downloadEnd := time.Now()
	downloadTime = downloadEnd.Sub(downloadStart)

	// Remove test block
	d.Remove(blockMetadata)

	// Calculate throughput
	throughput = int((uploadTime.Milliseconds()+downloadTime.Milliseconds())/2 + 1)

	logger.Logger.Debug("disk", "Disk ", d.GetName(), " has throughput of ", strconv.Itoa(throughput), "(upload: ", strconv.FormatInt(uploadTime.Milliseconds(), 10), " ms, download: ", strconv.FormatInt(downloadTime.Milliseconds(), 10), " ms).")
	return throughput
}
