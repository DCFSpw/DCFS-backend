package models

import (
	"dcfs/apicalls"
	"dcfs/constants"
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

func MeasureDiskThroughput(d Disk) int {
	var uploadTime time.Duration
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
	start := time.Now()

	err := d.Upload(blockMetadata)

	end := time.Now()
	uploadTime = end.Sub(start)

	// TO DO: Measure download time

	// TO DO: Remove test block

	// Calculate throughput
	throughput = int(uploadTime.Milliseconds() + 1)

	log.Println("Disk ", d.GetName(), " has throughput of ", uploadTime.Milliseconds(), blockMetadata.UUID, err, size)
	return throughput
}
