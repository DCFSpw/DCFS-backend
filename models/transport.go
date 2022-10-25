package models

import (
	"dcfs/db"
	"dcfs/db/dbo"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"sync"
	"time"
)

// TODO: transform to an abstract object
type VolumeContainer struct {
	Volume  *Volume
	Counter int64
}

type FileContainer struct {
	File    File
	Counter int64
}

func (transport *transport) updateCounter(vc *VolumeContainer, userUUID uuid.UUID, volumeUUID uuid.UUID) {
	vc.Counter++

	time.AfterFunc(6*time.Minute, func() {
		vc.Counter--

		if vc.Counter <= 0 {
			transport.ActiveVolumesMutex.Lock()
			defer transport.ActiveVolumesMutex.Unlock()

			delete(transport.ActiveVolumes[userUUID], volumeUUID)
		}
	})
}

type transport struct {
	ActiveVolumesMutex sync.Mutex
	ActiveVolumes      map[uuid.UUID]map[uuid.UUID]*VolumeContainer

	FileUploadQueueMutex sync.Mutex
	FileUploadQueue      map[uuid.UUID]*FileContainer
}

func (transport *transport) KeepAlive(userUUID uuid.UUID, volumeUUID uuid.UUID) {
	transport.ActiveVolumesMutex.Lock()
	defer transport.ActiveVolumesMutex.Unlock()

	_ = transport.getVolumeContainer(userUUID, volumeUUID)
}

func (transport *transport) GetVolume(userUUID uuid.UUID, volumeUUID uuid.UUID) *Volume {
	transport.ActiveVolumesMutex.Lock()
	defer transport.ActiveVolumesMutex.Unlock()

	return transport.getVolumeContainer(userUUID, volumeUUID).Volume
}

func (transport *transport) GetVolumes(userUUID uuid.UUID) []*Volume {
	transport.ActiveVolumesMutex.Lock()
	defer transport.ActiveVolumesMutex.Unlock()

	var rsp []*Volume
	var _volumes []dbo.Volume
	db.DB.DatabaseHandle.Where("user_uuid = ?", userUUID.String()).Find(&_volumes)

	for _, volume := range _volumes {
		rsp = append(rsp, transport.getVolumeContainer(userUUID, volume.UUID).Volume)
	}

	return rsp
}

func (transport *transport) getVolumeContainer(userUUID uuid.UUID, volumeUUID uuid.UUID) *VolumeContainer {
	if transport.ActiveVolumes == nil {
		transport.ActiveVolumes = make(map[uuid.UUID]map[uuid.UUID]*VolumeContainer)
	}

	if transport.ActiveVolumes[userUUID] == nil {
		transport.ActiveVolumes[userUUID] = make(map[uuid.UUID]*VolumeContainer)
	}

	var container *VolumeContainer
	if vc, ok := transport.ActiveVolumes[userUUID][volumeUUID]; ok {
		container = vc
	} else {
		var _volume dbo.Volume = dbo.Volume{}
		var _disks []dbo.Disk

		db.DB.DatabaseHandle.Where("volume_uuid = ?", volumeUUID).Preload("Volume").Preload("Provider").Find(&_disks)
		db.DB.DatabaseHandle.Where("uuid = ?", volumeUUID).First(&_volume)
		if _volume.UUID != volumeUUID {
			return &VolumeContainer{Volume: nil}
		}

		container = new(VolumeContainer)
		container.Volume = NewVolume(&_volume, _disks)
	}

	transport.ActiveVolumes[userUUID][volumeUUID] = container
	transport.updateCounter(container, userUUID, volumeUUID)
	return container
}

func (transport *transport) _updateCounter(UUID uuid.UUID) {
	var fc *FileContainer = transport.FileUploadQueue[UUID]
	fc.Counter++

	time.AfterFunc(6*time.Minute, func() {
		transport.FileUploadQueueMutex.Lock()
		defer transport.FileUploadQueueMutex.Unlock()

		var _fc *FileContainer = transport.FileUploadQueue[UUID]
		if _fc == nil {
			return
		}

		_fc.Counter--
		if _fc.Counter <= 0 {
			delete(transport.FileUploadQueue, UUID)
		}
	})
}

func (transport *transport) MarkAsUsed(UUID uuid.UUID) error {
	transport.FileUploadQueueMutex.Lock()
	defer transport.FileUploadQueueMutex.Unlock()

	var fc *FileContainer = transport.FileUploadQueue[UUID]
	if fc == nil {
		return errors.New(fmt.Sprintf("file with UUID: %s is not enqueued", UUID.String()))
	}

	fc.Counter++
	return nil
}

func (transport *transport) MarkAsCompleted(UUID uuid.UUID) {
	time.AfterFunc(6*time.Minute, func() {
		transport.FileUploadQueueMutex.Lock()
		defer transport.FileUploadQueueMutex.Unlock()

		var _fc *FileContainer = transport.FileUploadQueue[UUID]
		if _fc == nil {
			return
		}

		_fc.Counter--
		if _fc.Counter <= 0 {
			delete(transport.FileUploadQueue, UUID)
		}
	})
}

func (transport *transport) EnqueueFileUpload(UUID uuid.UUID, file File) {
	transport.FileUploadQueueMutex.Lock()
	defer transport.FileUploadQueueMutex.Unlock()

	if transport.FileUploadQueue == nil {
		transport.FileUploadQueue = make(map[uuid.UUID]*FileContainer)
	}

	var fc *FileContainer = new(FileContainer)
	fc.File = file
	transport.FileUploadQueue[UUID] = fc
	transport._updateCounter(UUID)
}

func (transport *transport) GetEnqueuedFileUpload(UUID uuid.UUID) File {
	transport.FileUploadQueueMutex.Lock()
	defer transport.FileUploadQueueMutex.Unlock()

	if transport.FileUploadQueue == nil {
		return nil
	}

	return transport.FileUploadQueue[UUID].File
}

func (transport *transport) RemoveEnqueuedFileUpload(UUID uuid.UUID) {
	transport.FileUploadQueueMutex.Lock()
	defer transport.FileUploadQueueMutex.Unlock()

	if transport.FileUploadQueue == nil {
		return
	}
	delete(transport.FileUploadQueue, UUID)
}

func (transport *transport) FindEnqueuedDisk(diskUUID uuid.UUID) Disk {
	for _, fc := range transport.FileUploadQueue {
		for _, block := range *fc.File.GetBlocks() {
			if block.Disk.GetUUID() == diskUUID {
				return block.Disk
			}
		}
	}

	return nil
}

// Transport - global variable
var Transport *transport = new(transport)
