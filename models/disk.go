package models

import (
	"dcfs/apicalls"
	"dcfs/constants"
	"dcfs/db"
	"dcfs/db/dbo"
	"dcfs/models/credentials"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"log"
	"net/http/httptest"
	"time"
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

	GetProviderSpace() (uint64, uint64, string)
	SetTotalSpace(quota uint64)
	GetTotalSpace() uint64
	GetUsedSpace() uint64
	UpdateUsedSpace(change int64)

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
	disk.SetTotalSpace(cdm.Disk.TotalSpace)
	cdm.Volume.AddDisk(disk.GetUUID(), disk)

	return disk
}

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

	// TO DO: Remove test block

	// Calculate throughput
	throughput = int((uploadTime.Milliseconds()+downloadTime.Milliseconds())/2 + 1)

	log.Println("Disk ", d.GetName(), " has throughput of ", throughput, "(upload: ", uploadTime.Milliseconds(), " ms, download: ", downloadTime.Milliseconds(), " ms).")
	return throughput
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

	log.Println("Free space on disk", d.GetName(), "is", freeSpace, "bytes", " (user defined:", userDefinedSpace, "bytes, provider defined:", providerDefinedSpace, "bytes, provider total:", providerTotalSpace, "bytes)")

	return freeSpace
}
