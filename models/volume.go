package models

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"dcfs/constants"
	"dcfs/db"
	"dcfs/db/dbo"
	"dcfs/requests"
	"dcfs/util/logger"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"io"
	"log"
	"math"
	"os"
	"strconv"
)

type Volume struct {
	UUID      uuid.UUID
	BlockSize int

	Name           string
	UserUUID       uuid.UUID
	VolumeSettings dbo.VolumeSettings

	disks        map[uuid.UUID]Disk
	virtualDisks map[uuid.UUID]Disk
	partitioner  Partitioner
}

// GetDisk - retrieve disk model from the volume
//
// params:
//   - diskUUID uuid.UUID: UUID of the disk to be retrieved
func (v *Volume) GetDisk(diskUUID uuid.UUID) Disk {
	if v.disks == nil {
		logger.Logger.Warning("volume", "Could not find the disk: ", diskUUID.String(), " (volume's disk map is not initialized).")
		return nil
	} else {
		disk, exists := v.disks[diskUUID]
		if exists {
			logger.Logger.Debug("volume", "Found a disk with the uuid: ", diskUUID.String(), ".")
			return disk
		}
	}

	if v.virtualDisks == nil {
		logger.Logger.Warning("volume", "Could not find the disk: ", diskUUID.String(), " (volume's virtual disk map is not initialized).")
		return nil
	} else {
		disk, exists := v.virtualDisks[diskUUID]
		if exists {
			logger.Logger.Debug("volume", "Found a virtual disk with the uuid: ", diskUUID.String(), ".")
			return disk
		}
	}

	logger.Logger.Warning("volume", "Could not find the disk: ", diskUUID.String(), ".")
	return nil
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
	logger.Logger.Debug("volume", "Added the disk: ", diskUUID.String(), " to the volume: ", v.UUID.String(), ".")
}

// AddVirtualDisk - add virtual disk to the volume
//
// params:
//   - diskUUID uuid.UUID: UUID of the virtual disk to be added to the volume
//   - _disk Disk: data of the disk
func (v *Volume) AddVirtualDisk(diskUUID uuid.UUID, _disk Disk) {
	if v.virtualDisks == nil {
		v.virtualDisks = make(map[uuid.UUID]Disk)
	}

	v.virtualDisks[diskUUID] = _disk
	logger.Logger.Debug("volume", "Added the virtual disk: ", diskUUID.String(), " to the volume: ", v.UUID.String(), ".")
}

// CreateVirtualDiskAddToVolume - create and add virtual disk with provided UUID to the volume
//
// params:
//   - diskUUID uuid.UUID: UUID of the virtual disk to be added to the volume
func (v *Volume) CreateVirtualDiskAddToVolume(_virtualDisk dbo.Disk) {
	var virtualDisk Disk

	// Initialize virtual disk map
	if v.virtualDisks == nil {
		v.virtualDisks = make(map[uuid.UUID]Disk)
	}

	// Create virtual disk based on the target backup type
	switch v.VolumeSettings.Backup {
	// No backup is used
	case constants.BACKUP_TYPE_NO_BACKUP:
		return

	// RAID1+0 backup
	case constants.BACKUP_TYPE_RAID_1:
		// Initialize RAID1 virtual disks
		var firstDisk Disk = nil
		var secondDisk Disk = nil

		// Locate real disks assigned to the virtual disk
		for _, disk := range v.disks {
			if disk.GetVirtualDiskUUID() == _virtualDisk.UUID {
				if firstDisk == nil {
					firstDisk = disk
				} else if secondDisk == nil {
					secondDisk = disk
				} else {
					logger.Logger.Error("volume", "RAID1+0 error. More than 2 disks were assigned to single RAID1 drive.")
					return
				}
			}
		}

		// Create RAID1 virtual disk
		if firstDisk != nil && secondDisk != nil {
			virtualDisk = DiskTypesRegistry[constants.PROVIDER_TYPE_RAID1]()
			virtualDisk.SetUUID(_virtualDisk.UUID)
			virtualDisk.AssignDisk(firstDisk)
			virtualDisk.AssignDisk(secondDisk)
			virtualDisk.SetVolume(v)
		} else {
			logger.Logger.Error("volume", "RAID1+0 error. Cannot load disks assigned to virtual RAID1 drive.")
			return
		}
	default:
		logger.Logger.Warning("volume", "Cannot initialize backup drives. Unknown backup type.")
	}

	// Add virtual disk to virtual disk map
	if virtualDisk != nil {
		v.virtualDisks[virtualDisk.GetUUID()] = virtualDisk
		logger.Logger.Debug("volume", "Added the virtual disk: ", virtualDisk.GetUUID().String(), " to the volume: ", v.UUID.String(), ".")
	}
}

// GenerateVirtualDisk - generate virtual disk for new disk (used by backup disks)
//
// params:
//   - newDisk Disk: new disk to be connected to the new virtual disk
//
// return type:
//   - uuid.UUID: virtual disk if matching is possible, uuid.Nil otherwise
//   - error: database operation error
func (v *Volume) GenerateVirtualDisk(newDisk Disk) (uuid.UUID, error) {
	var virtualDisk *dbo.Disk = dbo.NewVirtualDisk()
	virtualDisk.UUID = uuid.Nil

	switch v.VolumeSettings.Backup {
	case constants.BACKUP_TYPE_NO_BACKUP:
		return uuid.Nil, nil

	case constants.BACKUP_TYPE_RAID_1:
		var disk dbo.Disk

		// Find unassigned disk to pair with
		result := db.DB.DatabaseHandle.Where("volume_uuid = ? AND is_virtual = ? AND virtual_disk_uuid = ?", v.UUID, false, uuid.Nil).First(&disk)
		if result.Error != nil {
			logger.Logger.Debug("disk", "Could not find an unassigned disk to pair with.")
			if result.Error == gorm.ErrRecordNotFound {
				return uuid.Nil, nil
			} else {
				return uuid.Nil, result.Error
			}
		}

		// Retrieve RAID1 virtual provider from database
		var provider dbo.Provider
		result = db.DB.DatabaseHandle.Where("type = ?", constants.PROVIDER_TYPE_RAID1).First(&provider)
		if result.Error != nil {
			logger.Logger.Error("disk", "Could not find the provider with the type: ", string(constants.PROVIDER_TYPE_RAID1), " from the db.")
			return uuid.Nil, result.Error
		}

		// Generate virtual disk
		virtualDisk.UUID = uuid.New()
		virtualDisk.UserUUID = disk.UserUUID
		virtualDisk.VolumeUUID = v.UUID
		virtualDisk.ProviderUUID = provider.UUID

		result = db.DB.DatabaseHandle.Create(&virtualDisk)
		if result.Error != nil {
			return uuid.Nil, result.Error
		}

		// Save virtual disk uuid to selected disk
		result = db.DB.DatabaseHandle.Model(disk).Update("virtual_disk_uuid", virtualDisk.UUID)
		if result.Error != nil {
			return uuid.Nil, result.Error
		}
		v.disks[disk.UUID].SetVirtualDiskUUID(virtualDisk.UUID)
		v.disks[newDisk.GetUUID()].SetVirtualDiskUUID(virtualDisk.UUID)

		// Save virtual disk to the volume
		v.CreateVirtualDiskAddToVolume(*virtualDisk)

		return virtualDisk.UUID, nil
	default:
		return uuid.Nil, nil
	}
}

// DeleteDisk - remove disk from the volume
//
// params:
//   - diskUUID uuid.UUID: UUID of the disk to be deleted from the volume
func (v *Volume) DeleteDisk(diskUUID uuid.UUID) {
	if v.disks == nil {
		logger.Logger.Warning("volume", "There are no disks in the volume: ", v.UUID.String(), ".")
		return
	}

	delete(v.disks, diskUUID)
	logger.Logger.Debug("volume", "Successfully deleted the disk: ", diskUUID.String(), " from the volume: ", v.UUID.String(), ".")
}

// DeleteVirtualDisk - remove virtual disk from the volume
//
// params:
//   - diskUUID uuid.UUID: UUID of the virtual disk to be deleted from the volume
func (v *Volume) DeleteVirtualDisk(diskUUID uuid.UUID) {
	if v.virtualDisks == nil {
		logger.Logger.Warning("volume", "There are no virtual disks in the volume: ", v.UUID.String(), ".")
		return
	}

	if v.disks == nil {
		logger.Logger.Warning("volume", "There are no disks in the volume: ", v.UUID.String(), ".")
		return
	}

	delete(v.virtualDisks, diskUUID)

	for _, disk := range v.disks {
		if disk.GetVirtualDiskUUID() == diskUUID {
			delete(v.disks, disk.GetUUID())
		}
	}

	logger.Logger.Debug("volume", "Successfully deleted the virtual disk: ", diskUUID.String(), " from the volume: ", v.UUID.String(), ".")
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
		var block *Block = NewBlock(uuid.New(), userUUID, f, v.partitioner.AssignDisk(currentSize), currentSize, "", constants.BLOCK_STATUS_QUEUED, i)
		_f.Blocks[block.UUID] = block

		logger.Logger.Debug("disk", "Block ", strconv.Itoa(i), " assigned to", block.Disk.GetName())
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

// GetPartitioner - refresh partitioner of the volume
//
// return type:
//   - Partitioner: partitioner of the volume
func (v *Volume) GetPartitioner() Partitioner {
	return v.partitioner
}

// RefreshPartitioner - refresh partitioner data of the volume
//
// This function refreshes partitioner data of the volume. It is used
// to update partitioner data after some changes in the volume (for example
// adding or removing disks) or to refresh data used to assign disks (for
// example disk usage or throughput).
func (v *Volume) RefreshPartitioner() {
	var disks []Disk

	if v.VolumeSettings.Backup == constants.BACKUP_TYPE_NO_BACKUP {
		for _, disk := range v.disks {
			disks = append(disks, disk)
		}
	} else {
		for _, disk := range v.virtualDisks {
			disks = append(disks, disk)
		}
	}

	v.partitioner.FetchDisks(disks)
}

// InitializeBackup - initialize virtual disks if backup is enabled
//
// This function creates virtual disks for the target backup solution
// enabled for the volume. It also assigns real disks according to their
// assigned virtual disk UUID.
//
// params:
//   - virtualDisks []dbo.VirtualDisk: list of virtual disks to be created
func (v *Volume) InitializeBackup(virtualDisks []dbo.Disk) {
	for _, disk := range virtualDisks {
		v.CreateVirtualDiskAddToVolume(disk)
	}
}

// Encrypt - encrypt a []byte using a predefined 256 byte key
//
// params: block - []byte to be encrypted
//
// return: error
func (v *Volume) Encrypt(block *[]uint8) error {
	if v.VolumeSettings.Encryption == constants.ENCRYPTION_TYPE_NO_ENCRYPTION {
		return nil
	}

	key, err := os.ReadFile("./encryption.key")
	if err != nil {
		logger.Logger.Error("volume", "Could not read the encryption key, files will not be encrypted: ", err.Error())
		return err
	}

	cb, err := aes.NewCipher(key)
	if err != nil {
		logger.Logger.Error("volume", "Could not generate a block cipher object: ", err.Error())
		return err
	}

	gcm, err := cipher.NewGCM(cb)
	if err != nil {
		logger.Logger.Error("volume", "Could not generate a gcm object: ", err.Error())
		return err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		logger.Logger.Error("volume", "Could not populate the cipher nonce with a random seed: ", err.Error())
		return err
	}

	*block = gcm.Seal(nonce, nonce, *block, nil)

	return nil
}

// Decrypt - decrypt a []byte using a predefined 256 byte key
//
// params: block - []byte to be decrypted
//
// return: error
func (v *Volume) Decrypt(block *[]uint8) error {
	if v.VolumeSettings.Encryption == constants.ENCRYPTION_TYPE_NO_ENCRYPTION {
		return nil
	}

	key, err := os.ReadFile("./encryption.key")
	if err != nil {
		logger.Logger.Error("volume", "Could not read the encryption key, files will not be encrypted: ", err.Error(), ".")
		return err
	}

	cb, err := aes.NewCipher(key)
	if err != nil {
		logger.Logger.Error("volume", "Could not generate a block cipher object: ", err.Error(), ".")
		return err
	}

	gcm, err := cipher.NewGCM(cb)
	if err != nil {
		logger.Logger.Error("volume", "Could not generate a gcm object: ", err.Error(), ".")
		return err
	}

	nonce := (*block)[:gcm.NonceSize()]
	ciphertext := (*block)[gcm.NonceSize():]
	*block = make([]uint8, len(*block)-gcm.NonceSize())
	*block, err = gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		logger.Logger.Error("volume", "Could not decode the file: ", err.Error(), ".")
		return err
	}

	return nil
}

// IsReady - check if the volume is ready to begin operations on files
//
// return type: bool
func (v *Volume) IsReady() bool {
	if len(v.disks) == 0 {
		return false
	}

	for _, d := range v.disks {
		if !d.IsReady() {
			return false
		}
	}

	return true
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
func NewVolume(_volume *dbo.Volume, _disks []dbo.Disk, _virtualDisks []dbo.Disk) *Volume {
	var v *Volume = new(Volume)
	v.UUID = _volume.UUID
	v.BlockSize = constants.DEFAULT_VOLUME_BLOCK_SIZE

	v.Name = _volume.Name
	v.UserUUID = _volume.UserUUID
	v.VolumeSettings = _volume.VolumeSettings

	v.partitioner = CreatePartitioner(v.VolumeSettings.FilePartition, v)

	for _, _d := range _disks {
		d := CreateDisk(CreateDiskMetadata{
			Disk:   &_d,
			Volume: v,
		})

		if d != nil {
			v.AddDisk(d.GetUUID(), d)
		}
	}

	if v.VolumeSettings.Backup != constants.BACKUP_TYPE_NO_BACKUP {
		v.InitializeBackup(_virtualDisks)
	}

	v.RefreshPartitioner()

	log.Println("Created a new Volume: ", v)
	return v
}
