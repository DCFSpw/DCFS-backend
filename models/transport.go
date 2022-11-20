package models

import (
	"dcfs/constants"
	"dcfs/db"
	"dcfs/db/dbo"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"sync"
	"time"
)

type InstanceContainer struct {
	Instance interface{}
	Counter  int64
}

type ConcurrentInstances struct {
	InstanceMutex sync.Mutex
	Instances     map[uuid.UUID]*InstanceContainer
}

func NewConcurrentInstances() *ConcurrentInstances {
	return &ConcurrentInstances{
		InstanceMutex: sync.Mutex{},
		Instances:     make(map[uuid.UUID]*InstanceContainer),
	}
}

func (instances *ConcurrentInstances) updateCounter(key uuid.UUID) {
	instances.Instances[key].Counter++

	time.AfterFunc(Transport.WaitTime, func() {
		instances.InstanceMutex.Lock()
		defer instances.InstanceMutex.Unlock()

		if instances.Instances == nil || instances.Instances[key] == nil {
			return
		}

		instances.Instances[key].Counter--

		if instances.Instances[key].Counter <= 0 {
			delete(instances.Instances, key)
		}
	})
}

// MarkAsUsed - add '1' to the instance from instances under the provided key.
//
//	It is a blocking call, there is no callback that would reduce
//	the instance counter later - the instance will only be
//	deleted manually by the call to MarkAsCompleted
//
// fields:
//   - key - UUID of the instance to be marked
func (instances *ConcurrentInstances) MarkAsUsed(key uuid.UUID) error {
	instances.InstanceMutex.Lock()
	defer instances.InstanceMutex.Unlock()

	var container *InstanceContainer = instances.Instances[key]
	if container == nil {
		return errors.New(fmt.Sprintf("instance with UUID: %s is not enqueued", key.String()))
	}

	container.Counter++
	return nil
}

// MarkAsCompleted - removes '1' from the instance from instances under the provided key.
//
//	It manually checks whether the instance should be deleted because
//	it has been blocked by a call to MarkAsUsed
//
// fields:
//   - key - UUID of the instance to be marked
func (instances *ConcurrentInstances) MarkAsCompleted(key uuid.UUID) {
	time.AfterFunc(Transport.WaitTime, func() {
		instances.InstanceMutex.Lock()
		defer instances.InstanceMutex.Unlock()

		var container *InstanceContainer = instances.Instances[key]
		if container == nil {
			return
		}

		container.Counter--
		if container.Counter <= 0 {
			delete(instances.Instances, key)
		}
	})
}

// EnqueueInstance - enqueues instance in the instance queues and triggers its automatic deletion.
//
// fields:
//   - key - UUID of the newly enqueued instance
//   - instance
func (instances *ConcurrentInstances) EnqueueInstance(key uuid.UUID, instance interface{}) {
	instances.InstanceMutex.Lock()
	defer instances.InstanceMutex.Unlock()

	if instances.Instances == nil {
		instances.Instances = make(map[uuid.UUID]*InstanceContainer)
	}

	var container *InstanceContainer = &InstanceContainer{
		Instance: instance,
		Counter:  0,
	}

	instances.Instances[key] = container
	instances.updateCounter(key)
}

// GetEnqueuedInstance - gets the enqueued instance.
//
// fields:
//   - key
func (instances *ConcurrentInstances) GetEnqueuedInstance(key uuid.UUID) interface{} {
	instances.InstanceMutex.Lock()
	defer instances.InstanceMutex.Unlock()

	if instances.Instances == nil {
		return nil
	}

	if instances.Instances[key] == nil {
		return nil
	}

	return instances.Instances[key].Instance
}

// RemoveEnqueuedInstance - removes the enqueued instance.
//
// fields:
//   - key
func (instances *ConcurrentInstances) RemoveEnqueuedInstance(key uuid.UUID) {
	instances.InstanceMutex.Lock()
	defer instances.InstanceMutex.Unlock()

	if instances.Instances == nil {
		return
	}

	delete(instances.Instances, key)
}

type transport struct {
	ActiveVolumes     *ConcurrentInstances
	FileDownloadQueue *ConcurrentInstances
	FileUploadQueue   *ConcurrentInstances

	WaitTime time.Duration
}

/* public methods */

// VolumeKeepAlive - prolongs the volume life in the ActiveVolumes instance array
//
// fields:
//   - volumeUUID
func (transport *transport) VolumeKeepAlive(volumeUUID uuid.UUID) {
	transport.ActiveVolumes.InstanceMutex.Lock()
	defer transport.ActiveVolumes.InstanceMutex.Unlock()

	_ = transport.getVolumeContainer(volumeUUID)
}

// GetVolume - gets the volume handle from the ActiveVolumes instance array.
//
// fields:
//   - volumeUUID
func (transport *transport) GetVolume(volumeUUID uuid.UUID) *Volume {
	transport.ActiveVolumes.InstanceMutex.Lock()
	defer transport.ActiveVolumes.InstanceMutex.Unlock()

	c := transport.getVolumeContainer(volumeUUID).Instance
	if c == nil {
		return nil
	}

	return c.(*Volume)
}

// GetVolumes - gets an array of volume handles belonging to the given user
//
// fields:
//   - userUUID
func (transport *transport) GetVolumes(userUUID uuid.UUID) []*Volume {
	transport.ActiveVolumes.InstanceMutex.Lock()
	defer transport.ActiveVolumes.InstanceMutex.Unlock()

	var rsp []*Volume
	var _volumes []dbo.Volume
	db.DB.DatabaseHandle.Where("user_uuid = ?", userUUID.String()).Find(&_volumes)

	for _, volume := range _volumes {
		rsp = append(rsp, transport.getVolumeContainer(volume.UUID).Instance.(*Volume))
	}

	return rsp
}

// FindEnqueuedDisk - checks whether any block belonging to the given disk has been enqueued and returns it.
//
// fields:
//   - diskUUID
func (transport *transport) FindEnqueuedDisk(diskUUID uuid.UUID) Disk {
	for _, instance := range transport.FileUploadQueue.Instances {
		for _, block := range instance.Instance.(File).GetBlocks() {
			if block.Disk != nil && block.Disk.GetUUID() == diskUUID {
				return block.Disk
			}
		}
	}

	for _, instance := range transport.FileDownloadQueue.Instances {
		for _, block := range instance.Instance.(File).GetBlocks() {
			if block.Disk != nil && block.Disk.GetUUID() == diskUUID {
				return block.Disk
			}
		}
	}

	return nil
}

// FindEnqueuedVolume - checks whether any block belonging to the given volume has been enqueued and returns it.
//
// fields:
//   - volumeUUID
func (transport *transport) FindEnqueuedVolume(volumeUUID uuid.UUID) *Volume {
	for _, instance := range transport.FileUploadQueue.Instances {
		for _, block := range instance.Instance.(File).GetBlocks() {
			volume := block.Disk.GetVolume()
			if volume.UUID == volumeUUID {
				return volume
			}
		}
	}

	for _, instance := range transport.FileDownloadQueue.Instances {
		for _, block := range instance.Instance.(File).GetBlocks() {
			volume := block.Disk.GetVolume()
			if volume.UUID == volumeUUID {
				return volume
			}
		}
	}

	return nil
}

// DeleteVolume - deletes the given volume from the ActiveVolumes array.
//
// fields:
//   - volumeUUID
func (transport *transport) DeleteVolume(volumeUUID uuid.UUID) (string, error) {
	// TODO: Implement deletion process worker
	var errCode string
	var err error
	var volume *Volume

	// Retrieve volume from transport
	volume = Transport.GetVolume(volumeUUID)
	if volume == nil {
		return constants.TRANSPORT_VOLUME_NOT_FOUND, errors.New("Volume not found in transport layer")
	}

	// Trigger delete process in all disks assigned to this volume
	for _, disk := range volume.disks {
		errCode, err = disk.Delete()
		if err != nil {
			return errCode, err
		}
	}

	// Remove volume from transport
	transport.ActiveVolumes.RemoveEnqueuedInstance(volumeUUID)

	return constants.SUCCESS, nil
}

/* private methods */

func (transport *transport) getVolumeContainer(volumeUUID uuid.UUID) *InstanceContainer {
	if transport.ActiveVolumes == nil {
		transport.ActiveVolumes = &ConcurrentInstances{
			InstanceMutex: sync.Mutex{},
			Instances:     make(map[uuid.UUID]*InstanceContainer),
		}
	}

	var container *InstanceContainer
	if vc, ok := transport.ActiveVolumes.Instances[volumeUUID]; ok {
		container = vc
	} else {
		var _volume dbo.Volume = dbo.Volume{}
		var _disks []dbo.Disk

		db.DB.DatabaseHandle.Where("volume_uuid = ?", volumeUUID).Preload("Volume").Preload("Provider").Find(&_disks)
		db.DB.DatabaseHandle.Where("uuid = ?", volumeUUID).First(&_volume)
		if _volume.UUID != volumeUUID {
			return &InstanceContainer{Instance: nil}
		}

		container = new(InstanceContainer)
		container.Instance = NewVolume(&_volume, _disks)
	}

	transport.ActiveVolumes.Instances[volumeUUID] = container
	transport.ActiveVolumes.updateCounter(volumeUUID)

	return container
}

func NewTransport() *transport {
	return &transport{
		ActiveVolumes:     NewConcurrentInstances(),
		FileDownloadQueue: NewConcurrentInstances(),
		FileUploadQueue:   NewConcurrentInstances(),
		WaitTime:          6 * time.Minute,
	}
}

// Transport - global variable
var Transport *transport = NewTransport()
