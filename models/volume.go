package models

import (
	"dcfs/models/disk"
	"github.com/google/uuid"
)

type Volume struct {
	disks map[uuid.UUID]disk.Disk
}

func (v *Volume) GetDisk(diskUUID uuid.UUID) disk.Disk {
	if v.disks == nil {
		return nil
	}

	return v.disks[diskUUID]
}

func (v *Volume) AddDisk(diskUUID uuid.UUID, _disk disk.Disk) {
	if v.disks == nil {
		v.disks = make(map[uuid.UUID]disk.Disk)
	}

	v.disks[diskUUID] = _disk
}
