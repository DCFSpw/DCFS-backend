package models

import (
	"dcfs/db"
	"dcfs/db/dbo"
	"dcfs/models/disk/DriveFactory"
	"github.com/google/uuid"
	"sync"
	"time"
)

type VolumeContainer struct {
	Volume  *Volume
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

		db.DB.DatabaseHandle.Where("volume_uuid = ?", volumeUUID).Find(&_disks)
		db.DB.DatabaseHandle.Where("uuid = ?", volumeUUID).First(&_volume)

		container = new(VolumeContainer)
		container.Volume = new(Volume)

		if _volume.UUID == volumeUUID {
			// TODO: add volume fields

			for _, _d := range _disks {
				provider := dbo.Provider{}
				db.DB.DatabaseHandle.Where("uuid = ?", _d.ProviderUUID).First(&provider)

				d := DriveFactory.NewDisk(provider.ProviderType)
				d.SetUUID(_d.UUID)
				d.CreateCredentials(_d.Credentials)
				container.Volume.AddDisk(d.GetUUID(), d)
			}
		}
	}

	transport.ActiveVolumes[userUUID][volumeUUID] = container
	transport.updateCounter(container, userUUID, volumeUUID)
	return container
}

// Transport - global variable
var Transport *transport = new(transport)
