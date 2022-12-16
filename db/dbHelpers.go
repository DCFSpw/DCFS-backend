package db

import (
	"dcfs/constants"
	"dcfs/db/dbo"
	"dcfs/util/logger"
	"github.com/google/uuid"
)

// UserFromDatabase - retrieve user from database
//
// params:
//   - uuid uuid.UUID: UUID of the requested user
//
// return type:
//   - *dbo.User: user DBO data retrieved from database
//   - string: completion code
func UserFromDatabase(uuid uuid.UUID) (*dbo.User, string) {
	var user *dbo.User = dbo.NewUser()

	result := DB.DatabaseHandle.Where("uuid = ?", uuid).First(&user)
	if result.Error != nil {
		logger.Logger.Warning("db", "Could not find a user with the provided uuid: ", uuid.String(), " in the db.")
		return nil, constants.DATABASE_USER_NOT_FOUND
	}

	logger.Logger.Debug("db", "Found a user with the uuid: ", uuid.String(), " in the db.")
	return user, constants.SUCCESS
}

// VolumeFromDatabase - retrieve volume from database
//
// params:
//   - uuid string: UUID of the requested volume
//
// return type:
//   - *dbo.Volume: volume DBO data retrieved from database
//   - string: completion code
func VolumeFromDatabase(uuid string) (*dbo.Volume, string) {
	var volume *dbo.Volume = dbo.NewVolume()

	result := DB.DatabaseHandle.Where("uuid = ?", uuid).First(&volume)
	if result.Error != nil {
		logger.Logger.Warning("db", "Could not find a volume with the provided uuid: ", uuid, " in the db.")
		return nil, constants.DATABASE_VOLUME_NOT_FOUND
	}

	logger.Logger.Debug("db", "Found a volume with the uuid: ", uuid, " in the db.")
	return volume, constants.SUCCESS
}

// FileFromDatabase - retrieve file from database
//
// params:
//   - uuid string: UUID of the requested file
//
// return type:
//   - *dbo.File: file DBO data retrieved from database
//   - string: completion code
func FileFromDatabase(uuid string) (*dbo.File, string) {
	var file *dbo.File = dbo.NewFile()

	result := DB.DatabaseHandle.Where("uuid = ?", uuid).First(&file)
	if result.Error != nil {
		logger.Logger.Warning("db", "Could not find a file with the uuid: ", uuid, " in the db.")
		return nil, constants.DATABASE_FILE_NOT_FOUND
	}

	logger.Logger.Debug("db", "Found a file with the provided uuid: ", uuid, " in the db.")
	return file, constants.SUCCESS
}

// BlocksFromDatabase - retrieve blocks of the file from database
//
// params:
//   - fileUUID string: UUID of the file
//
// return type:
//   - []*dbo.Block: blocks DBO data retrieved from database
//   - string: completion code
func BlocksFromDatabase(fileUUID string) ([]*dbo.Block, string) {
	var blocks []*dbo.Block

	err := DB.DatabaseHandle.Where("file_uuid = ?", fileUUID).Find(&blocks).Error
	if err != nil {
		logger.Logger.Warning("db", "Could not find a block with the provided uuid: ", fileUUID, " in the db.")
		return nil, constants.DATABASE_ERROR
	}

	logger.Logger.Debug("db", "Found a block with the uuid: ", fileUUID, " in the db.")
	return blocks, constants.SUCCESS
}

// IsVolumeEmpty - verify whether volume is empty
//
// params:
//   - uuid string: UUID of the volume to check
//
// return type:
//   - bool: true if volume is empty, false otherwise
//   - error: database operation error
func IsVolumeEmpty(uuid uuid.UUID) (bool, error) {
	var blockCount int64
	err := DB.DatabaseHandle.Model(&dbo.Block{}).Where("volume_uuid = ?", uuid).Count(&blockCount).Error

	if err != nil {
		logger.Logger.Warning("db", "Could not find a volume with the provided uuid: ", uuid.String(), " in the db.")
	}

	logger.Logger.Debug("db", "Found a volume with the uuid: ", uuid.String(), " in the db.")
	return blockCount == 0, err
}

// FindUnassignedDisk - find disk from provided volume that is not assigned to any virtual disk
//
// params:
//   - volumeUUID uuid.UUID: UUID of the volume to search in
//
// return type:
//   - *dbo.Disk: unassigned disk, nil if none found
//   - error: database operation error
func FindUnassignedDisk(volumeUUID uuid.UUID) (*dbo.Disk, error) {
	var disk dbo.Disk

	result := DB.DatabaseHandle.Where("volume_uuid = ? AND is_virtual = ? AND virtual_disk_uuid = ?", volumeUUID, false, uuid.Nil).First(&disk)
	if result.Error != nil {
		return nil, result.Error
	}

	return &disk, nil
}

// IsDirectoryEmpty - verify whether directory is empty
//
// params:
//   - uuid string: UUID of the directory to check
//
// return type:
//   - bool: true if directory is empty, false otherwise
//   - error: database operation error
func IsDirectoryEmpty(uuid uuid.UUID) (bool, error) {
	var fileCount int64
	err := DB.DatabaseHandle.Model(&dbo.File{}).Where("root_uuid = ?", uuid).Count(&fileCount).Error
	return fileCount == 0, err
}

// ValidateRootDirectory - verify whether provided root is valid
//
// This function verifies whether provided root entity is a valid root.
// It verifies whether provided root exists in the filesystem and
// checks its type (only directory can be a root for another entity).
// If provided root is uuid.Nil (meaning: root of the volume), it is accepted as
// a valid root directory.
//
// params:
//   - rootUUID uuid.UUID: UUID of the target root
//   - volumeUUID uuid.UUID: UUID of the volume
//
// return type:
//   - string: completion code (constants.SUCCESS if root is valid)
func ValidateRootDirectory(rootUUID uuid.UUID, volumeUUID uuid.UUID) string {
	var rootDirectory *dbo.File = dbo.NewFile()

	// Check if the root directory refers to volume's root directory
	if rootUUID == uuid.Nil {
		logger.Logger.Debug("db", "The provided directory is a root directory.")
		return constants.SUCCESS
	}

	// Check if the root directory exists
	rootDirectory, errCode := FileFromDatabase(rootUUID.String())
	if errCode != constants.SUCCESS {
		logger.Logger.Warning("db", "Could not retrieve the root directory from the db with the error: ", errCode)
		return errCode
	} else if rootDirectory == nil {
		logger.Logger.Warning("db", "Could not find the root directory in the db.")
		return constants.DATABASE_FILE_NOT_FOUND
	}

	// Check if root directory belongs to the provided volume
	if rootDirectory.VolumeUUID != volumeUUID {
		logger.Logger.Warning("db", "The provided directory is not the root directory of this volume.")
		return constants.FS_VOLUME_MISMATCH
	}

	// Check if root directory is a directory
	if rootDirectory.Type != constants.FILE_TYPE_DIRECTORY {
		logger.Logger.Warning("db", "The root directory is not a directory - possible db failure.")
		return constants.FS_FILE_TYPE_MISMATCH
	}

	logger.Logger.Debug("db", "The provided directory is a root directory.")
	return constants.SUCCESS
}

// GenerateFileFullPath - generate full path for provided file/directory
//
// This function generates full path for provided file or directory.
// It traverses through subsequent filesystem roots until it reaches the root
// of the volume. Returned list is ordered in the order of traversal - first
// entry is provided file/directory, last entry is root of volume.
//
// params:
//   - rootUUID uuid.UUID: UUID of the file/directory to generate full path
//
// return type:
//   - []dbo.PathEntry: list of directories from the full path
//   - string: completion code
func GenerateFileFullPath(rootUUID uuid.UUID) ([]dbo.PathEntry, string) {
	var pathMap map[uuid.UUID]bool = make(map[uuid.UUID]bool)
	var path []dbo.PathEntry = make([]dbo.PathEntry, 0)
	var _path string = ""

	// Iterate through file's path to root directory of the volume
	for rootUUID != uuid.Nil {
		var parent dbo.File

		// Retrieve parent directory from database
		result := DB.DatabaseHandle.Where("uuid = ?", rootUUID).First(&parent)
		if result.Error != nil {
			logger.Logger.Error("db", "Could not find the root file with the uuid: ", rootUUID.String(), " from the db.")
			return nil, constants.DATABASE_ERROR
		}

		// Add parent directory to path
		path = append(path, dbo.PathEntry{
			UUID: parent.UUID,
			Name: parent.Name,
		})
		_path = "/" + _path + parent.Name
		pathMap[parent.UUID] = true

		// Move to parent directory
		rootUUID = parent.RootUUID
		if pathMap[rootUUID] {
			logger.Logger.Error("api", "Found a cycle in the file system.")
			return nil, constants.FS_PATH_CYCLE
		}
	}

	logger.Logger.Debug("db", "Successfully generated a path: ", _path, " for a rootUUID: ", rootUUID.String())
	return path, constants.SUCCESS
}
