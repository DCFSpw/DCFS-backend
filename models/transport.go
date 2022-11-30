package models

import (
	"context"
	"dcfs/apicalls"
	"dcfs/constants"
	"dcfs/db"
	"dcfs/db/dbo"
	"dcfs/util/logger"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"strconv"
	"golang.org/x/sync/errgroup"
	"log"
	"net/http/httptest"
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
	logger.Logger.Debug("transport", "The counter of the instance", key.String(), " has been updated, it is now at: ", strconv.FormatInt(instances.Instances[key].Counter, 10), ".")

	time.AfterFunc(Transport.WaitTime, func() {
		instances.InstanceMutex.Lock()
		defer instances.InstanceMutex.Unlock()

		if instances.Instances == nil || instances.Instances[key] == nil {
			logger.Logger.Warning("transport", "The instance: ", key.String(), " was not present in the collection.")
			return
		}

		instances.Instances[key].Counter--
		logger.Logger.Debug("transport", "Decreased the counter of the instance: ", key.String(), " to: ", strconv.FormatInt(instances.Instances[key].Counter, 10), ".")

		if instances.Instances[key].Counter <= 0 {
			delete(instances.Instances, key)
			logger.Logger.Debug("transport", "Deleted the instance: ", key.String(), " from Transport.")
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
		logger.Logger.Error("transport", "An instance with the UUID: ", key.String(), " is not enqueued.")
		return errors.New(fmt.Sprintf("instance with UUID: %s is not enqueued", key.String()))
	}

	container.Counter++
	logger.Logger.Debug("transport", "The instance: ", key.String(), " has been marked used, the counter is: ", strconv.FormatInt(container.Counter, 10), ".")
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
			logger.Logger.Warning("transport", "The instance: ", key.String(), " has been deleted previously.")
			return
		}

		container.Counter--
		logger.Logger.Debug("transport", "Decreased the counter of the instance: ", key.String(), " it now is at: ", strconv.FormatInt(container.Counter, 10), ".")

		if container.Counter <= 0 {
			delete(instances.Instances, key)
			logger.Logger.Debug("transport", "Deleted the instance: ", key.String(), ".")
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

	logger.Logger.Debug("transport", "Successfully enqueued the instance: ", key.String(), ".")
}

// GetEnqueuedInstance - gets the enqueued instance.
//
// fields:
//   - key
func (instances *ConcurrentInstances) GetEnqueuedInstance(key uuid.UUID) interface{} {
	instances.InstanceMutex.Lock()
	defer instances.InstanceMutex.Unlock()

	if instances.Instances == nil {
		logger.Logger.Warning("transport", "There is no instance: ", key.String(), " enqueued.")
		return nil
	}

	if instances.Instances[key] == nil {
		logger.Logger.Warning("transport", "There is no instance: ", key.String(), " enqueued.")
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
		logger.Logger.Warning("transport", "There is no instance: ", key.String(), " enqueued.")
		return
	}

	delete(instances.Instances, key)
	logger.Logger.Debug("transport", "Successfully deleted the instance: ", key.String(), ".")
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
	logger.Logger.Debug("transport", "Volume: ", volumeUUID.String(), " kept alive.")
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
		logger.Logger.Warning("transport", "Could not find a Volume with the uuid: ", volumeUUID.String(), " enqueued.")
		return nil
	}

	logger.Logger.Debug("transport", "Found the volume: ", volumeUUID.String(), ".")
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

	logger.Logger.Debug("transport", "Found ", strconv.Itoa(len(rsp)), " volumes for the user: ", userUUID.String(), ".")
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
				logger.Logger.Debug("transport", "Found the disk: ", block.Disk.GetUUID().String(), " enqueued.")
				return block.Disk
			}
		}
	}

	for _, instance := range transport.FileDownloadQueue.Instances {
		for _, block := range instance.Instance.(File).GetBlocks() {
			if block.Disk != nil && block.Disk.GetUUID() == diskUUID {
				logger.Logger.Debug("transport", "Found the disk: ", block.Disk.GetUUID().String(), " enqueued.")
				return block.Disk
			}
		}
	}

	logger.Logger.Warning("transport", "Could not find a disk with the UUID: ", diskUUID.String(), " enqueued.")
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
				logger.Logger.Debug("transport", "Found a volume: ", volumeUUID.String(), " enqueued.")
				return volume
			}
		}
	}

	for _, instance := range transport.FileDownloadQueue.Instances {
		for _, block := range instance.Instance.(File).GetBlocks() {
			volume := block.Disk.GetVolume()
			if volume.UUID == volumeUUID {
				logger.Logger.Debug("transport", "Found a volume: ", volumeUUID.String(), " enqueued.")
				return volume
			}
		}
	}

	logger.Logger.Warning("transport", "Did not find any volumes with uuid: ", volumeUUID.String(), " enqueued.")
	return nil
}

// DeleteVolume - deletes the given volume, its disks and removes it from the ActiveVolumes array.
//
// fields:
//   - volumeUUID
func (transport *transport) DeleteVolume(volumeUUID uuid.UUID) (string, error) {
	var volume *Volume

	// Retrieve volume from transport
	volume = Transport.GetVolume(volumeUUID)
	if volume == nil {
		logger.Logger.Error("transport", "Volume: ", volumeUUID.String(), " not found in the transport layer.")
		return constants.TRANSPORT_VOLUME_NOT_FOUND, errors.New("Volume not found in transport layer")
	}

	// Trigger delete process in all disks assigned to this volume
	waitGroup, _ := errgroup.WithContext(context.Background())

	for _, disk := range volume.disks {
		waitGroup.Go(func() error {
			errCode, err := transport.DeleteDisk(disk, volume, constants.DELETION)
			if errCode != constants.SUCCESS {
				logger.Logger.Error("transport", "Could not delete the disk: ", disk.GetUUID().String(), ".")
				return err
			}

			return nil
		})
	}

	err := waitGroup.Wait()
	if err != nil {
		return constants.OPERATION_FAILED, err
	}

	// Remove volume from transport
	transport.ActiveVolumes.RemoveEnqueuedInstance(volumeUUID)

	logger.Logger.Debug("transport", "Successfully removed the volume: ", volumeUUID.String(), " from Transport.")
	return constants.SUCCESS, nil
}

func (transport *transport) DeleteDisk(disk Disk, volume *Volume, relocate bool) (string, error) {
	var blocks []dbo.Block

	// Retrieve list of blocks on disk
	dBErr := db.DB.DatabaseHandle.Where("disk_uuid = ?", disk.GetUUID()).Find(&blocks).Error
	if dBErr != nil {
		return constants.DATABASE_ERROR, dBErr
	}

	// Delete blocks from disk
	var waitGroup sync.WaitGroup
	var taskCompleted bool = true

	waitGroup.Add(len(blocks))

	for _, block := range blocks {
		go func(block dbo.Block) {
			log.Println("Deleting block", block.UUID)
			defer waitGroup.Done()

			// Prepare test context
			writer := httptest.NewRecorder()
			_ctx, _ := gin.CreateTestContext(writer)

			// Prepare apicall metadata
			var status int
			var blockMetadata *apicalls.BlockMetadata = new(apicalls.BlockMetadata)
			blockMetadata.Ctx = _ctx
			blockMetadata.FileUUID = block.FileUUID
			blockMetadata.Content = nil
			blockMetadata.UUID = block.UUID
			blockMetadata.Size = int64(block.Size)
			blockMetadata.Status = &status
			blockMetadata.CompleteCallback = func(UUID uuid.UUID, status *int) {
			}

			// Delete block from current disk
			var result *apicalls.ErrorWrapper

			result = disk.Remove(blockMetadata)
			if result != nil {
				taskCompleted = false
				return
			}

			// Remove block from database
			dBErr := db.DB.DatabaseHandle.Delete(&block).Error
			if dBErr != nil {
				taskCompleted = false
				return
			}

			return
		}(block)
	}
	waitGroup.Wait()
	if taskCompleted != true {
		return constants.OPERATION_FAILED, errors.New("Failed to delete blocks from disk")
	}

	// Remove disk from database
	dbErr := db.DB.DatabaseHandle.Delete(&dbo.Disk{}, disk.GetUUID()).Error
	if dbErr != nil {
		return constants.DATABASE_ERROR, dbErr
	}

	// Disattach disk from volume
	volume.DeleteDisk(disk.GetUUID())

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
			logger.Logger.Warning("transport", "Could not find a volume: ", volumeUUID.String(), " in the db.")
			return &InstanceContainer{Instance: nil}
		}

		container = new(InstanceContainer)
		container.Instance = NewVolume(&_volume, _disks)
	}

	transport.ActiveVolumes.Instances[volumeUUID] = container
	transport.ActiveVolumes.updateCounter(volumeUUID)

	logger.Logger.Debug("transport", "Successfully enqueued the volume: ", volumeUUID.String(), " in Transport.")
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
