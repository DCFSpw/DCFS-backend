package models

import (
	"dcfs/constants"
	"dcfs/db/dbo"
	"dcfs/requests"
	"github.com/google/uuid"
	"log"
	"math"
)

type Volume struct {
	UUID      uuid.UUID
	BlockSize int

	Name           string
	UserUUID       uuid.UUID
	VolumeSettings dbo.VolumeSettings

	disks       map[uuid.UUID]Disk
	partitioner Partitioner
}

// GetDisk - retrieve disk model from the volume
//
// params:
//   - diskUUID uuid.UUID: UUID of the disk to be retrieved
func (v *Volume) GetDisk(diskUUID uuid.UUID) Disk {
	if v.disks == nil {
		return nil
	}

	return v.disks[diskUUID]
}

// AddDisk - add disk to the volume
//
// params:
//   - diskUUID uuid.UUID: UUID of the disk to be added to the volume
//   - _disk Disk: data of the disk
func (v *Volume) AddDisk(diskUUID uuid.UUID, _disk Disk) {
	if v.disks == nil {
		v.disks = make(map[uuid.UUID]Disk)
	}

	v.disks[diskUUID] = _disk
}

// DeleteDisk - remove disk from the volume
//
// params:
//   - diskUUID uuid.UUID: UUID of the disk to be deleted from the volume
func (v *Volume) DeleteDisk(diskUUID uuid.UUID) {
	if v.disks == nil {
		return
	}

	delete(v.disks, diskUUID)
}

// FileUploadRequest - handle initial request for uploading file to the volume
//
// This function prepares file for upload to the volume. It receives data from the init
// upload file request, partitions file into blocks, and prepares list of blocks to be
// uploaded to the volume (along with assignment of each block to target disk).
//
// params:
//   - request *requests.InitFileUploadRequest: init file upload request data from API request
//   - userUUID uuid.UUID: UUID of the user who is uploading the file
//   - rootUUID uuid.UUID: UUID of the root directory where the file is uploaded
//
// return type:
//   - RegularFile: created volume model
func (v *Volume) FileUploadRequest(request *requests.InitFileUploadRequest, userUUID uuid.UUID, rootUUID uuid.UUID) RegularFile {
	var f File = NewFileFromRequest(request, rootUUID)
	f.SetVolume(v)

	// Prepare partition of the file
	var _f *RegularFile = f.(*RegularFile)
	var blockCount int = int(math.Max(math.Ceil(float64(request.File.Size)/float64(v.BlockSize)), 1))
	var cumulativeSize int = 0

	// Partition the file into blocks
	_f.Blocks = make(map[uuid.UUID]*Block)
	for i := 0; i < blockCount; i++ {
		// Compute the size of the block
		var currentSize int = v.BlockSize
		cumulativeSize += v.BlockSize
		if cumulativeSize > f.GetSize() {
			currentSize = v.BlockSize - (cumulativeSize - f.GetSize())
		}

		// Create a new block
		var block *Block = NewBlock(uuid.New(), userUUID, f, v.partitioner.AssignDisk(currentSize), currentSize, 0, constants.BLOCK_STATUS_QUEUED, i)
		_f.Blocks[block.UUID] = block

		log.Println("Block ", i, " assigned to", block.Disk.GetName())
	}

	return *_f
}

// GetVolumeDBO - generate volume DBO object based on the volume model
//
// return_type:
//   - *dbo.Volume: volume DBO data generated from the volume
func (v *Volume) GetVolumeDBO() dbo.Volume {
	return dbo.Volume{
		AbstractDatabaseObject: dbo.AbstractDatabaseObject{UUID: v.UUID},
		Name:                   v.Name,
		UserUUID:               v.UserUUID,
		VolumeSettings:         v.VolumeSettings,
	}
}

// RefreshPartitioner - refresh partitioner data of the volume
//
// This function refreshes partitioner data of the volume. It is used
// to update partitioner data after some changes in the volume (for example
// adding or removing disks) or to refresh data used to assign disks (for
// example disk usage or throughput).
func (v *Volume) RefreshPartitioner() {
	v.partitioner.FetchDisks()
}

// NewVolume - create new volume model based on volume and disks DBO
//
// This function creates volume model used internally by backend based on
// volume and disks data obtained from database. It also initialized the
// volume by assigning disks to the volume and creating appropriate function
// handlers, for example partitioner.
//
// params:
//   - _volume *dbo.Volume: volume DBO data (from database)
//   - _disks []dbo.Disk: disks DBO data (from database)
//
// return type:
//   - *Volume: created volume model
func NewVolume(_volume *dbo.Volume, _disks []dbo.Disk) *Volume {
	var v *Volume = new(Volume)
	v.UUID = _volume.UUID
	v.BlockSize = constants.DEFAULT_VOLUME_BLOCK_SIZE

	v.Name = _volume.Name
	v.UserUUID = _volume.UserUUID
	v.VolumeSettings = _volume.VolumeSettings

	v.partitioner = CreatePartitioner(v.VolumeSettings.FilePartition, v)

	for _, _d := range _disks {
		_ = CreateDisk(CreateDiskMetadata{
			Disk:   &_d,
			Volume: v,
		})
	}

	v.RefreshPartitioner()

	log.Println("Created a new Volume: ", v)
	return v
}
