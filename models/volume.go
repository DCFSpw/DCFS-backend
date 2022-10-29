package models

import (
	"dcfs/apicalls"
	"dcfs/constants"
	"dcfs/db/dbo"
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

func (v *Volume) GetDisk(diskUUID uuid.UUID) Disk {
	if v.disks == nil {
		return nil
	}

	return v.disks[diskUUID]
}

func (v *Volume) AddDisk(diskUUID uuid.UUID, _disk Disk) {
	if v.disks == nil {
		v.disks = make(map[uuid.UUID]Disk)
	}

	v.disks[diskUUID] = _disk
}

func (v *Volume) DeleteDisk(diskUUID uuid.UUID) {
	if v.disks == nil {
		return
	}

	delete(v.disks, diskUUID)
}

func (v *Volume) FileUploadRequest(req *apicalls.FileUploadRequest) File {
	var f File = NewFileFromReq(req)
	f.SetVolume(v)

	if req.Type == constants.FILE_TYPE_REGULAR {
		var _f *RegularFile = f.(*RegularFile)
		var blockCount int = int(math.Max(math.Ceil(float64(req.Size/v.BlockSize)), 1))
		var cumulativeSize int = 0

		_f.Blocks = make(map[uuid.UUID]*Block)
		for i := 0; i < blockCount; i++ {
			var currentSize int = v.BlockSize
			cumulativeSize += v.BlockSize
			if cumulativeSize > f.GetSize() {
				currentSize = v.BlockSize - (cumulativeSize - f.GetSize())
			}

			var block *Block = NewBlock(uuid.New(), req.UserUUID, &f, v.partitioner.AssignDisk(currentSize), currentSize, 0, constants.BLOCK_STATUS_QUEUED, i)
			_f.Blocks[block.UUID] = block
		}
	} else {
		panic("unimplemented")
	}

	return f
}

func (v *Volume) GetVolumeDBO() dbo.Volume {
	return dbo.Volume{
		AbstractDatabaseObject: dbo.AbstractDatabaseObject{UUID: v.UUID},
		Name:                   v.Name,
		UserUUID:               v.UserUUID,
		VolumeSettings:         v.VolumeSettings,
	}
}

func NewVolume(_volume *dbo.Volume, _disks []dbo.Disk) *Volume {
	var v *Volume = new(Volume)
	v.partitioner = NewDummyPartitioner(v)
	v.UUID = _volume.UUID
	v.BlockSize = 8 * 1024 * 1024

	v.Name = _volume.Name
	v.UserUUID = _volume.UserUUID
	v.VolumeSettings = _volume.VolumeSettings

	for _, _d := range _disks {
		_ = CreateDisk(CreateDiskMetadata{
			Disk:   &_d,
			Volume: v,
		})
	}

	log.Println("Created a new Volume: ", v)
	return v
}
