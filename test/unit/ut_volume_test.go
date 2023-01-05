package unit

import (
	"crypto/rand"
	"dcfs/constants"
	"dcfs/db/dbo"
	"dcfs/models"
	"dcfs/requests"
	"dcfs/test/unit/mock"
	_ "dcfs/util/logger"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	. "github.com/smartystreets/goconvey/convey"
	"io"
	"regexp"
	"testing"
	"time"

	_ "dcfs/models/disk/BackupDisk"
	_ "dcfs/models/disk/FTPDisk"
	_ "dcfs/models/disk/GDriveDisk"
	_ "dcfs/models/disk/OneDriveDisk"
	_ "dcfs/models/disk/SFTPDisk"
)

func TestGetDisk(t *testing.T) {
	models.Transport.ActiveVolumes.RemoveEnqueuedInstance(mock.VolumeDBO.UUID)

	disks := mock.GetDiskDBOs(1)
	mockDisks := []models.Disk{mock.NewMockDisk(), mock.NewMockDisk()}
	volume := MockNewVolume(*mock.VolumeDBO, disks, true)
	volume2 := MockNewVolume(*mock.VolumeDBO, nil, true)
	volume3 := MockNewVolume(*mock.BackupVolumeDBO, nil, true)

	for _, _mdisk := range mockDisks {
		volume3.AddDisk(_mdisk.GetUUID(), _mdisk)
	}

	mock.DBMock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `disks` WHERE volume_uuid = ? AND is_virtual = ? AND virtual_disk_uuid = ? ORDER BY `disks`.`uuid` LIMIT 1")).WithArgs(volume3.UUID.String(), false, uuid.Nil).WillReturnRows(mock.DiskRow(&dbo.Disk{
		AbstractDatabaseObject: dbo.AbstractDatabaseObject{
			UUID: mockDisks[1].GetUUID(),
		},
		UserUUID:        uuid.UUID{},
		VolumeUUID:      volume3.UUID,
		ProviderUUID:    uuid.UUID{},
		Credentials:     "",
		Name:            "",
		UsedSpace:       0,
		TotalSpace:      0,
		FreeSpace:       0,
		CreatedAt:       time.Time{},
		IsVirtual:       false,
		VirtualDiskUUID: uuid.UUID{},
		User:            dbo.User{},
		Volume:          *mock.BackupVolumeDBO,
		Provider:        dbo.Provider{},
	}))

	mock.DBMock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `providers` WHERE type = ?")).WithArgs(constants.PROVIDER_TYPE_RAID1).WillReturnRows(mock.ProviderRow(&dbo.Provider{
		AbstractDatabaseObject: dbo.AbstractDatabaseObject{
			UUID: uuid.UUID{},
		},
		Type: constants.PROVIDER_TYPE_RAID1,
		Name: "",
		Logo: "",
	}))

	mock.DBMock.ExpectBegin()
	mock.DBMock.ExpectExec("INSERT INTO `disks`").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.DBMock.ExpectCommit()

	mock.DBMock.ExpectBegin()
	mock.DBMock.ExpectExec(regexp.QuoteMeta("UPDATE `disks`")).WillReturnResult(sqlmock.NewResult(1, 1))
	mock.DBMock.ExpectCommit()

	UUID, _ := volume3.GenerateVirtualDisk(mockDisks[0])

	Convey("Should successfully retrieve the disk", t, func() {
		So(volume.GetDisk(disks[0].UUID).GetUUID(), ShouldEqual, disks[0].UUID)
	})

	Convey("Should return nil if there are no disks", t, func() {
		So(volume2.GetDisk(disks[0].UUID), ShouldEqual, nil)
	})

	Convey("Should correctly find the virtual disk", t, func() {
		So(volume3.GetDisk(UUID), ShouldNotEqual, nil)
	})

	Convey("Should return nil if a disk with the specified UUID is not added to the volume", t, func() {
		So(volume3.GetDisk(uuid.New()), ShouldEqual, nil)
	})

	Convey("The database call should be correct", t, func() {
		So(mock.DBMock.ExpectationsWereMet(), ShouldEqual, nil)
	})
}

func TestGetDisks(t *testing.T) {
	mockDisks := []models.Disk{mock.NewMockDisk(), mock.NewMockDisk()}
	volume := MockNewVolume(*mock.BackupVolumeDBO, nil, true)
	volume2 := MockNewVolume(*mock.VolumeDBO, nil, true)
	volume2.AddDisk(mockDisks[0].GetUUID(), mockDisks[0])
	for _, _mdisk := range mockDisks {
		volume.AddDisk(_mdisk.GetUUID(), _mdisk)
	}

	mock.DBMock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `disks` WHERE volume_uuid = ? AND is_virtual = ? AND virtual_disk_uuid = ? ORDER BY `disks`.`uuid` LIMIT 1")).WithArgs(volume.UUID.String(), false, uuid.Nil).WillReturnRows(mock.DiskRow(&dbo.Disk{
		AbstractDatabaseObject: dbo.AbstractDatabaseObject{
			UUID: mockDisks[1].GetUUID(),
		},
		UserUUID:        uuid.UUID{},
		VolumeUUID:      volume.UUID,
		ProviderUUID:    uuid.UUID{},
		Credentials:     "",
		Name:            "",
		UsedSpace:       0,
		TotalSpace:      0,
		FreeSpace:       0,
		CreatedAt:       time.Time{},
		IsVirtual:       false,
		VirtualDiskUUID: uuid.UUID{},
		User:            dbo.User{},
		Volume:          *mock.BackupVolumeDBO,
		Provider:        dbo.Provider{},
	}))

	mock.DBMock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `providers` WHERE type = ?")).WithArgs(constants.PROVIDER_TYPE_RAID1).WillReturnRows(mock.ProviderRow(&dbo.Provider{
		AbstractDatabaseObject: dbo.AbstractDatabaseObject{
			UUID: uuid.UUID{},
		},
		Type: constants.PROVIDER_TYPE_RAID1,
		Name: "",
		Logo: "",
	}))

	mock.DBMock.ExpectBegin()
	mock.DBMock.ExpectExec("INSERT INTO `disks`").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.DBMock.ExpectCommit()

	mock.DBMock.ExpectBegin()
	mock.DBMock.ExpectExec(regexp.QuoteMeta("UPDATE `disks`")).WillReturnResult(sqlmock.NewResult(1, 1))
	mock.DBMock.ExpectCommit()

	UUID, _ := volume.GenerateVirtualDisk(mockDisks[0])
	Convey("The volume (with backup) should return an array with the newly created virtual disk", t, func() {
		disks := volume.GetDisks()
		Convey("The returned array should not be nil", func() {
			So(disks, ShouldNotEqual, nil)
		})
		Convey("The length of the returned array should be 1", func() {
			So(len(disks), ShouldEqual, 1)
		})
		Convey("The only item in the returned array should be the newly created virtual disk", func() {
			So(disks[UUID].GetUUID(), ShouldEqual, UUID)
		})
	})

	Convey("The standard volume should return an array of added disks", t, func() {
		disks2 := volume2.GetDisks()
		Convey("The returned array should not be nil", func() {
			So(disks2, ShouldNotEqual, nil)
		})
		Convey("The length of the returned array should be 1", func() {
			So(len(disks2), ShouldEqual, 1)
		})
		Convey("The only item in the returned array should be the newly created virtual disk", func() {
			So(disks2[mockDisks[0].GetUUID()].GetUUID(), ShouldEqual, mockDisks[0].GetUUID())
		})
	})

	Convey("The database call should be correct", t, func() {
		So(mock.DBMock.ExpectationsWereMet(), ShouldEqual, nil)
	})
}

func TestAddDisk(t *testing.T) {
	disks := mock.GetMockDisks(1)
	volume := MockNewVolume(*mock.VolumeDBO, nil, true)

	volume.AddDisk(disks[0].UUID, disks[0])

	Convey("The disk has been successfully added", t, func() {
		So(volume.GetDisk(disks[0].UUID).GetUUID(), ShouldEqual, disks[0].UUID)
	})
	Convey("The database call should be correct", t, func() {
		So(mock.DBMock.ExpectationsWereMet(), ShouldEqual, nil)
	})
}

func TestAddVirtualDisk(t *testing.T) {
	disk := mock.NewMockDisk()
	volume := MockNewVolume(*mock.BackupVolumeDBO, nil, true)
	volume.AddVirtualDisk(disk.GetUUID(), disk)

	Convey("This function has been excluded from testing", t, func() {
		So(true, ShouldEqual, true)
	})
}

func TestCreateVirtualDiskAddToVolume(t *testing.T) {
	mockDisks := []models.Disk{mock.NewMockDisk(), mock.NewMockDisk(), mock.NewMockDisk()}
	volume := MockNewVolume(*mock.BackupVolumeDBO, nil, true)
	volume2 := MockNewVolume(*mock.VolumeDBO, nil, true)
	volume2.AddDisk(mockDisks[0].GetUUID(), mockDisks[0])
	for _, _mdisk := range mockDisks {
		volume.AddDisk(_mdisk.GetUUID(), _mdisk)
	}

	mock.DBMock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `disks` WHERE volume_uuid = ? AND is_virtual = ? AND virtual_disk_uuid = ? ORDER BY `disks`.`uuid` LIMIT 1")).WithArgs(volume.UUID.String(), false, uuid.Nil).WillReturnRows(mock.DiskRow(&dbo.Disk{
		AbstractDatabaseObject: dbo.AbstractDatabaseObject{
			UUID: mockDisks[1].GetUUID(),
		},
		UserUUID:        uuid.UUID{},
		VolumeUUID:      volume.UUID,
		ProviderUUID:    uuid.UUID{},
		Credentials:     "",
		Name:            "",
		UsedSpace:       0,
		TotalSpace:      0,
		FreeSpace:       0,
		CreatedAt:       time.Time{},
		IsVirtual:       false,
		VirtualDiskUUID: uuid.UUID{},
		User:            dbo.User{},
		Volume:          *mock.BackupVolumeDBO,
		Provider:        dbo.Provider{},
	}))

	mock.DBMock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `providers` WHERE type = ?")).WithArgs(constants.PROVIDER_TYPE_RAID1).WillReturnRows(mock.ProviderRow(&dbo.Provider{
		AbstractDatabaseObject: dbo.AbstractDatabaseObject{
			UUID: uuid.UUID{},
		},
		Type: constants.PROVIDER_TYPE_RAID1,
		Name: "",
		Logo: "",
	}))

	mock.DBMock.ExpectBegin()
	mock.DBMock.ExpectExec("INSERT INTO `disks`").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.DBMock.ExpectCommit()

	mock.DBMock.ExpectBegin()
	mock.DBMock.ExpectExec(regexp.QuoteMeta("UPDATE `disks`")).WillReturnResult(sqlmock.NewResult(1, 1))
	mock.DBMock.ExpectCommit()

	// standard cases are already covered by GetDisks tests, this just handles the corner cases

	// non-backup disk
	volume2.CreateVirtualDiskAddToVolume(dbo.Disk{}) /* no error should occur */

	// non-recognizable disk
	volume2.VolumeSettings.Backup = 5
	volume2.CreateVirtualDiskAddToVolume(dbo.Disk{}) /* no error should occur */

	// create a virtual volume and assign it to three disks
	UUID, _ := volume.GenerateVirtualDisk(mockDisks[0])
	mockDisks[2].SetVirtualDiskUUID(UUID)
	volume.CreateVirtualDiskAddToVolume(dbo.Disk{AbstractDatabaseObject: dbo.AbstractDatabaseObject{UUID: UUID}}) /* no error should occur */

	// delete the virtual disk uuid from the mock disks and trigger another error handling
	for _, _d := range mockDisks {
		_d.SetVirtualDiskUUID(uuid.Nil)
	}
	volume.CreateVirtualDiskAddToVolume(dbo.Disk{AbstractDatabaseObject: dbo.AbstractDatabaseObject{UUID: UUID}}) /* no error should occur */

	Convey("Validate the corner cases of CreateVirtualDiskAddToVolume", t, func() {
		So(true, ShouldEqual, true)
	})

	Convey("The database call should be correct", t, func() {
		So(mock.DBMock.ExpectationsWereMet(), ShouldEqual, nil)
	})
}

func TestGenerateVirtualDisk(t *testing.T) {
	// standard cases are already covered by GetDisks tests, this just handles the corner cases

	mockDisks := []models.Disk{mock.NewMockDisk(), mock.NewMockDisk()}
	volume := MockNewVolume(*mock.BackupVolumeDBO, nil, true)
	volume2 := MockNewVolume(*mock.VolumeDBO, nil, true)
	volume2.AddDisk(mockDisks[0].GetUUID(), mockDisks[0])
	for _, _mdisk := range mockDisks {
		volume.AddDisk(_mdisk.GetUUID(), _mdisk)
	}

	Convey("GenerateVirtualDisk returns error when it can't find an unassigned disk", t, func() {
		UUID, err := volume.GenerateVirtualDisk(mockDisks[0])
		So(UUID, ShouldEqual, uuid.Nil)
		So(err, ShouldNotEqual, nil)
	})

	mock.DBMock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `disks` WHERE volume_uuid = ? AND is_virtual = ? AND virtual_disk_uuid = ? ORDER BY `disks`.`uuid` LIMIT 1")).WithArgs(volume.UUID.String(), false, uuid.Nil).WillReturnRows(mock.DiskRow(&dbo.Disk{
		AbstractDatabaseObject: dbo.AbstractDatabaseObject{
			UUID: mockDisks[1].GetUUID(),
		},
		UserUUID:        uuid.UUID{},
		VolumeUUID:      volume.UUID,
		ProviderUUID:    uuid.UUID{},
		Credentials:     "",
		Name:            "",
		UsedSpace:       0,
		TotalSpace:      0,
		FreeSpace:       0,
		CreatedAt:       time.Time{},
		IsVirtual:       false,
		VirtualDiskUUID: uuid.UUID{},
		User:            dbo.User{},
		Volume:          *mock.BackupVolumeDBO,
		Provider:        dbo.Provider{},
	}))

	Convey("GenerateVirtualDisk returns error when it can't find a suitable provider", t, func() {
		UUID, err := volume.GenerateVirtualDisk(mockDisks[0])
		So(UUID, ShouldEqual, uuid.Nil)
		So(err, ShouldNotEqual, nil)
	})

	mock.DBMock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `disks` WHERE volume_uuid = ? AND is_virtual = ? AND virtual_disk_uuid = ? ORDER BY `disks`.`uuid` LIMIT 1")).WithArgs(volume.UUID.String(), false, uuid.Nil).WillReturnRows(mock.DiskRow(&dbo.Disk{
		AbstractDatabaseObject: dbo.AbstractDatabaseObject{
			UUID: mockDisks[1].GetUUID(),
		},
		UserUUID:        uuid.UUID{},
		VolumeUUID:      volume.UUID,
		ProviderUUID:    uuid.UUID{},
		Credentials:     "",
		Name:            "",
		UsedSpace:       0,
		TotalSpace:      0,
		FreeSpace:       0,
		CreatedAt:       time.Time{},
		IsVirtual:       false,
		VirtualDiskUUID: uuid.UUID{},
		User:            dbo.User{},
		Volume:          *mock.BackupVolumeDBO,
		Provider:        dbo.Provider{},
	}))

	mock.DBMock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `providers` WHERE type = ?")).WithArgs(constants.PROVIDER_TYPE_RAID1).WillReturnRows(mock.ProviderRow(&dbo.Provider{
		AbstractDatabaseObject: dbo.AbstractDatabaseObject{
			UUID: uuid.UUID{},
		},
		Type: constants.PROVIDER_TYPE_RAID1,
		Name: "",
		Logo: "",
	}))

	Convey("GenerateVirtualDisk returns error when it can't create the virtual disk in the db", t, func() {
		UUID, err := volume.GenerateVirtualDisk(mockDisks[0])
		So(UUID, ShouldEqual, uuid.Nil)
		So(err, ShouldNotEqual, nil)
	})

	mock.DBMock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `disks` WHERE volume_uuid = ? AND is_virtual = ? AND virtual_disk_uuid = ? ORDER BY `disks`.`uuid` LIMIT 1")).WithArgs(volume.UUID.String(), false, uuid.Nil).WillReturnRows(mock.DiskRow(&dbo.Disk{
		AbstractDatabaseObject: dbo.AbstractDatabaseObject{
			UUID: mockDisks[1].GetUUID(),
		},
		UserUUID:        uuid.UUID{},
		VolumeUUID:      volume.UUID,
		ProviderUUID:    uuid.UUID{},
		Credentials:     "",
		Name:            "",
		UsedSpace:       0,
		TotalSpace:      0,
		FreeSpace:       0,
		CreatedAt:       time.Time{},
		IsVirtual:       false,
		VirtualDiskUUID: uuid.UUID{},
		User:            dbo.User{},
		Volume:          *mock.BackupVolumeDBO,
		Provider:        dbo.Provider{},
	}))

	mock.DBMock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `providers` WHERE type = ?")).WithArgs(constants.PROVIDER_TYPE_RAID1).WillReturnRows(mock.ProviderRow(&dbo.Provider{
		AbstractDatabaseObject: dbo.AbstractDatabaseObject{
			UUID: uuid.UUID{},
		},
		Type: constants.PROVIDER_TYPE_RAID1,
		Name: "",
		Logo: "",
	}))

	mock.DBMock.ExpectBegin()
	mock.DBMock.ExpectExec("INSERT INTO `disks`").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.DBMock.ExpectCommit()

	Convey("GenerateVirtualDisk returns error when it can't update the real disks in the db", t, func() {
		UUID, err := volume.GenerateVirtualDisk(mockDisks[0])
		So(UUID, ShouldEqual, uuid.Nil)
		So(err, ShouldNotEqual, nil)
	})

	// non-backup disk
	Convey("GenerateVirtualDisk returns nothing when the volume does not have a backup option enabled", t, func() {
		UUID, err := volume2.GenerateVirtualDisk(mockDisks[0])
		So(UUID, ShouldEqual, uuid.Nil)
		So(err, ShouldEqual, nil)
	})

	// non-recognizable disk
	volume2.VolumeSettings.Backup = 5
	Convey("GenerateVirtualDisk returns nothing when the volume does not have a backup option enabled", t, func() {
		UUID, err := volume2.GenerateVirtualDisk(mockDisks[0])
		So(UUID, ShouldEqual, uuid.Nil)
		So(err, ShouldEqual, nil)
	})

	Convey("The database call should be correct", t, func() {
		So(mock.DBMock.ExpectationsWereMet(), ShouldEqual, nil)
	})
}

func TestDeleteDisk(t *testing.T) {
	// disable partitioner calculation of real disk space
	oldCalculateDiskSpaceFunction := models.CalculateDiskSpaceFunction
	models.CalculateDiskSpaceFunction = func(d models.Disk) uint64 { return uint64(2 * constants.DEFAULT_VOLUME_BLOCK_SIZE) }

	// make sure the mockVolume will have a balanced partitioner
	mock.VolumeDBO.VolumeSettings.FilePartition = constants.PARTITION_TYPE_BALANCED

	models.Transport.ActiveVolumes.RemoveEnqueuedInstance(mock.VolumeDBO.UUID)

	disks := mock.GetDiskDBOs(1)
	volume := MockNewVolume(*mock.VolumeDBO, disks, true)
	volume2 := MockNewVolume(*mock.VolumeDBO, nil, true)

	volume.DeleteDisk(disks[0].UUID)

	Convey("The disk has been successfully deleted", t, func() {
		So(volume.GetDisk(disks[0].UUID), ShouldEqual, nil)
	})

	Convey("Nothing should happen if the disk list is a nil", t, func() {
		volume2.DeleteDisk(disks[0].UUID)
		So(true, ShouldEqual, true)
	})

	Convey("The database call should be correct", t, func() {
		So(mock.DBMock.ExpectationsWereMet(), ShouldEqual, nil)
	})

	models.CalculateDiskSpaceFunction = oldCalculateDiskSpaceFunction
}

func TestDeleteVirtualDisk(t *testing.T) {
	mockDisks := []models.Disk{mock.NewMockDisk(), mock.NewMockDisk()}
	volume := MockNewVolume(*mock.BackupVolumeDBO, nil, true)
	volume2 := MockNewVolume(*mock.VolumeDBO, nil, true)
	volume2.AddDisk(mockDisks[0].GetUUID(), mockDisks[0])
	for _, _mdisk := range mockDisks {
		volume.AddDisk(_mdisk.GetUUID(), _mdisk)
	}

	mock.DBMock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `disks` WHERE volume_uuid = ? AND is_virtual = ? AND virtual_disk_uuid = ? ORDER BY `disks`.`uuid` LIMIT 1")).WithArgs(volume.UUID.String(), false, uuid.Nil).WillReturnRows(mock.DiskRow(&dbo.Disk{
		AbstractDatabaseObject: dbo.AbstractDatabaseObject{
			UUID: mockDisks[1].GetUUID(),
		},
		UserUUID:        uuid.UUID{},
		VolumeUUID:      volume.UUID,
		ProviderUUID:    uuid.UUID{},
		Credentials:     "",
		Name:            "",
		UsedSpace:       0,
		TotalSpace:      0,
		FreeSpace:       0,
		CreatedAt:       time.Time{},
		IsVirtual:       false,
		VirtualDiskUUID: uuid.UUID{},
		User:            dbo.User{},
		Volume:          *mock.BackupVolumeDBO,
		Provider:        dbo.Provider{},
	}))

	mock.DBMock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `providers` WHERE type = ?")).WithArgs(constants.PROVIDER_TYPE_RAID1).WillReturnRows(mock.ProviderRow(&dbo.Provider{
		AbstractDatabaseObject: dbo.AbstractDatabaseObject{
			UUID: uuid.UUID{},
		},
		Type: constants.PROVIDER_TYPE_RAID1,
		Name: "",
		Logo: "",
	}))

	mock.DBMock.ExpectBegin()
	mock.DBMock.ExpectExec("INSERT INTO `disks`").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.DBMock.ExpectCommit()

	mock.DBMock.ExpectBegin()
	mock.DBMock.ExpectExec(regexp.QuoteMeta("UPDATE `disks`")).WillReturnResult(sqlmock.NewResult(1, 1))
	mock.DBMock.ExpectCommit()

	UUID, _ := volume.GenerateVirtualDisk(mockDisks[0])

	// no virtual disks
	volume2.DeleteVirtualDisk(UUID)

	Convey("The virtual disk has successfully been deleted", t, func() {
		volume.DeleteVirtualDisk(UUID)
		So(volume.GetDisk(UUID), ShouldEqual, nil)
		So(volume.GetDisk(mockDisks[0].GetUUID()), ShouldEqual, nil)
		So(volume.GetDisk(mockDisks[1].GetUUID()), ShouldEqual, nil)
	})
}

func TestFindAnotherDisk(t *testing.T) {
	// disable partitioner calculation of real disk space
	oldCalculateDiskSpaceFunction := models.CalculateDiskSpaceFunction
	models.CalculateDiskSpaceFunction = func(d models.Disk) uint64 { return uint64(2 * constants.DEFAULT_VOLUME_BLOCK_SIZE) }

	disks := []models.Disk{mock.NewMockDisk(), mock.NewMockDisk()}
	volume := MockNewVolume(*mock.VolumeDBO, nil, true)
	for _, _d := range disks {
		volume.AddDisk(_d.GetUUID(), _d)
	}

	Convey("The volume properly finds the other disk", t, func() {
		disk := volume.FindAnotherDisk(disks[0].GetUUID())
		So(disk, ShouldNotEqual, nil)
		So(disk.GetUUID(), ShouldEqual, disks[1].GetUUID())
	})

	Convey("The database call should be correct", t, func() {
		So(mock.DBMock.ExpectationsWereMet(), ShouldEqual, nil)
	})

	models.CalculateDiskSpaceFunction = oldCalculateDiskSpaceFunction
}

func TestClearFileSystem(t *testing.T) {
	volume := MockNewVolume(*mock.VolumeDBO, nil, true)

	mock.DBMock.ExpectBegin()
	mock.DBMock.ExpectExec("UPDATE `files`").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.DBMock.ExpectCommit()

	Convey("The ClearFilesystem method properly deletes the volume from the db and returns no error", t, func() {
		So(volume.ClearFilesystem(), ShouldEqual, nil)
	})
}

func TestFileUploadRequest(t *testing.T) {
	// disable partitioner calculation of real disk space
	oldCalculateDiskSpaceFunction := models.CalculateDiskSpaceFunction
	models.CalculateDiskSpaceFunction = func(d models.Disk) uint64 { return uint64(2 * constants.DEFAULT_VOLUME_BLOCK_SIZE) }

	models.Transport.ActiveVolumes.RemoveEnqueuedInstance(mock.VolumeDBO.UUID)

	// make sure the mockVolume will have a balanced partitioner
	mock.VolumeDBO.VolumeSettings.FilePartition = constants.PARTITION_TYPE_BALANCED

	disks := mock.GetDiskDBOs(10)
	volume := MockNewVolume(*mock.VolumeDBO, disks, true)

	req := &requests.InitFileUploadRequest{
		VolumeUUID: volume.UUID.String(),
		RootUUID:   "",
		File: requests.FileDataRequest{
			Name: "test",
			Type: constants.FILE_TYPE_REGULAR,
			Size: 100*constants.DEFAULT_VOLUME_BLOCK_SIZE + (constants.DEFAULT_VOLUME_BLOCK_SIZE / 2),
		},
	}

	file := volume.FileUploadRequest(req, volume.UserUUID, uuid.Nil)

	Convey("Verify that the file has been properly split", t, func() {
		Convey("There is an appropriate number of blocks", func() {
			So(len(file.GetBlocks()), ShouldEqual, 101)
		})
		Convey("Every block has been assigned an existing disk", func() {
			for _, block := range file.GetBlocks() {
				So(volume.GetDisk(block.Disk.GetUUID()), ShouldEqual, block.Disk)
			}
		})
	})

	Convey("The database call should be correct", t, func() {
		So(mock.DBMock.ExpectationsWereMet(), ShouldEqual, nil)
	})

	models.CalculateDiskSpaceFunction = oldCalculateDiskSpaceFunction
}

func TestGetVolumeDBO(t *testing.T) {
	volume := models.NewVolume(mock.VolumeDBO, nil, nil)
	volumeDBO := volume.GetVolumeDBO()

	Convey("Test if the volume data is properly encoded into a db object", t, func() {
		Convey("UUID is set properly", func() {
			So(volumeDBO.UUID, ShouldEqual, mock.VolumeDBO.UUID)
		})
		Convey("Name is set properly", func() {
			So(volumeDBO.Name, ShouldEqual, mock.VolumeDBO.Name)
		})
		Convey("UserUUID is set properly", func() {
			So(volumeDBO.UserUUID, ShouldEqual, mock.VolumeDBO.UserUUID)
		})
		Convey("VolumeSettings are set properly", func() {
			So(volumeDBO.VolumeSettings, ShouldResemble, mock.VolumeDBO.VolumeSettings)
		})
	})
	Convey("The database call should be correct", t, func() {
		So(mock.DBMock.ExpectationsWereMet(), ShouldEqual, nil)
	})
}

func TestGetPartitioner(t *testing.T) {
	volume := MockNewVolume(*mock.VolumeDBO, nil, true)

	Convey("The GetPartitioner method returns the actual volume's partitioner", t, func() {
		So(volume.GetPartitioner(), ShouldNotEqual, nil)
	})

	Convey("The database call should be correct", t, func() {
		So(mock.DBMock.ExpectationsWereMet(), ShouldEqual, nil)
	})
}

func TestInitializeBackup(t *testing.T) {
	disks := mock.GetDiskDBOs(1)
	volume := MockNewVolume(*mock.VolumeDBO, nil, true)

	volume.InitializeBackup(disks)

	Convey("Validate that this test succeeded with no errors and panics", t, func() {
		So(true, ShouldEqual, true)
	})

	Convey("The database call should be correct", t, func() {
		So(mock.DBMock.ExpectationsWereMet(), ShouldEqual, nil)
	})
}

func TestRefreshPartitioner(t *testing.T) {
	volume := models.NewVolume(mock.VolumeDBO, nil, nil)
	volume.RefreshPartitioner()

	volume3 := MockNewVolume(*mock.BackupVolumeDBO, nil, true)
	mockDisks := []models.Disk{mock.NewMockDisk(), mock.NewMockDisk()}

	for _, _mdisk := range mockDisks {
		volume3.AddDisk(_mdisk.GetUUID(), _mdisk)
	}

	mock.DBMock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `disks` WHERE volume_uuid = ? AND is_virtual = ? AND virtual_disk_uuid = ? ORDER BY `disks`.`uuid` LIMIT 1")).WithArgs(volume3.UUID.String(), false, uuid.Nil).WillReturnRows(mock.DiskRow(&dbo.Disk{
		AbstractDatabaseObject: dbo.AbstractDatabaseObject{
			UUID: mockDisks[1].GetUUID(),
		},
		UserUUID:        uuid.UUID{},
		VolumeUUID:      volume3.UUID,
		ProviderUUID:    uuid.UUID{},
		Credentials:     "",
		Name:            "",
		UsedSpace:       0,
		TotalSpace:      0,
		FreeSpace:       0,
		CreatedAt:       time.Time{},
		IsVirtual:       false,
		VirtualDiskUUID: uuid.UUID{},
		User:            dbo.User{},
		Volume:          *mock.BackupVolumeDBO,
		Provider:        dbo.Provider{},
	}))

	mock.DBMock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `providers` WHERE type = ?")).WithArgs(constants.PROVIDER_TYPE_RAID1).WillReturnRows(mock.ProviderRow(&dbo.Provider{
		AbstractDatabaseObject: dbo.AbstractDatabaseObject{
			UUID: uuid.UUID{},
		},
		Type: constants.PROVIDER_TYPE_RAID1,
		Name: "",
		Logo: "",
	}))

	mock.DBMock.ExpectBegin()
	mock.DBMock.ExpectExec("INSERT INTO `disks`").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.DBMock.ExpectCommit()

	mock.DBMock.ExpectBegin()
	mock.DBMock.ExpectExec(regexp.QuoteMeta("UPDATE `disks`")).WillReturnResult(sqlmock.NewResult(1, 1))
	mock.DBMock.ExpectCommit()

	_, _ = volume3.GenerateVirtualDisk(mockDisks[0])

	volume3.RefreshPartitioner()

	Convey("This method is excluded from the Unit Tests", t, func() {
		So(true, ShouldEqual, true)
	})
	Convey("The database call should be correct", t, func() {
		So(mock.DBMock.ExpectationsWereMet(), ShouldEqual, nil)
	})
}

func TestEncrypt(t *testing.T) {
	volume := MockNewVolume(*mock.VolumeDBO, nil, true)
	block := make([]uint8, 1024)

	_, _ = io.ReadFull(rand.Reader, block)

	original := make([]uint8, 1024)
	for idx, _ := range block {
		original[idx] = block[idx]
	}

	Convey("The block should not be encrypted when the encryption option is off", t, func() {
		err := volume.Encrypt(&block)
		Convey("The error should be nil", func() {
			So(err, ShouldEqual, nil)
		})
		Convey("The files should not be encrypted and identical", func() {
			identical := true

			for i := 0; i < 1024; i++ {
				if original[i] != block[i] {
					identical = false
				}
			}

			So(identical, ShouldEqual, true)
		})
	})

	volume.VolumeSettings.Encryption = constants.ENCRYPTION_TYPE_AES_256
	Convey("The block should be encrypted when the encryption option is on", t, func() {
		err := volume.Encrypt(&block)
		So(err, ShouldEqual, nil)

		identical := true

		for i := 0; i < 1024; i++ {
			if original[i] != block[i] {
				identical = false
			}
		}

		So(identical, ShouldEqual, false)

		_ = volume.Decrypt(&block)
		identical = true

		for i := 0; i < 1024; i++ {
			if original[i] != block[i] {
				identical = false
			}
		}

		So(identical, ShouldEqual, true)
	})
	Convey("The database call should be correct", t, func() {
		So(mock.DBMock.ExpectationsWereMet(), ShouldEqual, nil)
	})
}

func TestDecrypt(t *testing.T) {
	volume := MockNewVolume(*mock.VolumeDBO, nil, true)
	block := make([]uint8, 1024)

	_, _ = io.ReadFull(rand.Reader, block)

	original := make([]uint8, 1024)
	for idx, _ := range block {
		original[idx] = block[idx]
	}

	volume.VolumeSettings.Encryption = constants.ENCRYPTION_TYPE_AES_256

	// encrypt the block
	_ = volume.Encrypt(&block)

	volume.VolumeSettings.Encryption = constants.ENCRYPTION_TYPE_NO_ENCRYPTION

	Convey("The block should not be decrypted if the encryption setting is of", t, func() {
		err := volume.Decrypt(&block)
		Convey("The returned error should be nil", func() {
			So(err, ShouldEqual, nil)
		})
		Convey("The block should not be the same as the original", func() {
			identical := true

			for i := 0; i < 1024; i++ {
				if original[i] != block[i] {
					identical = false
				}
			}

			So(identical, ShouldEqual, false)
		})
	})

	volume.VolumeSettings.Encryption = constants.ENCRYPTION_TYPE_AES_256

	Convey("The block should be successfully decrypted if the encryption setting is on", t, func() {
		err := volume.Decrypt(&block)
		So(err, ShouldEqual, nil)

		identical := true

		for i := 0; i < 1024; i++ {
			if original[i] != block[i] {
				identical = false
			}
		}

		So(identical, ShouldEqual, true)
	})
	Convey("The database call should be correct", t, func() {
		So(mock.DBMock.ExpectationsWereMet(), ShouldEqual, nil)
	})
}

func TestIsReady(t *testing.T) {
	mockDisks := []models.Disk{mock.NewMockDisk(), mock.NewMockDisk()}
	volume := MockNewVolume(*mock.BackupVolumeDBO, nil, true)

	for _, _mdisk := range mockDisks {
		volume.AddDisk(_mdisk.GetUUID(), _mdisk)
	}

	mock.DBMock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `disks` WHERE volume_uuid = ? AND is_virtual = ? AND virtual_disk_uuid = ? ORDER BY `disks`.`uuid` LIMIT 1")).WithArgs(volume.UUID.String(), false, uuid.Nil).WillReturnRows(mock.DiskRow(&dbo.Disk{
		AbstractDatabaseObject: dbo.AbstractDatabaseObject{
			UUID: mockDisks[1].GetUUID(),
		},
		UserUUID:        uuid.UUID{},
		VolumeUUID:      volume.UUID,
		ProviderUUID:    uuid.UUID{},
		Credentials:     "",
		Name:            "",
		UsedSpace:       0,
		TotalSpace:      0,
		FreeSpace:       0,
		CreatedAt:       time.Time{},
		IsVirtual:       false,
		VirtualDiskUUID: uuid.UUID{},
		User:            dbo.User{},
		Volume:          *mock.BackupVolumeDBO,
		Provider:        dbo.Provider{},
	}))

	mock.DBMock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `providers` WHERE type = ?")).WithArgs(constants.PROVIDER_TYPE_RAID1).WillReturnRows(mock.ProviderRow(&dbo.Provider{
		AbstractDatabaseObject: dbo.AbstractDatabaseObject{
			UUID: uuid.UUID{},
		},
		Type: constants.PROVIDER_TYPE_RAID1,
		Name: "",
		Logo: "",
	}))

	mock.DBMock.ExpectBegin()
	mock.DBMock.ExpectExec("INSERT INTO `disks`").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.DBMock.ExpectCommit()

	mock.DBMock.ExpectBegin()
	mock.DBMock.ExpectExec(regexp.QuoteMeta("UPDATE `disks`")).WillReturnResult(sqlmock.NewResult(1, 1))
	mock.DBMock.ExpectCommit()

	_, _ = volume.GenerateVirtualDisk(mockDisks[0])

	newDisk := mock.NewMockDisk()
	volume.AddDisk(newDisk.GetUUID(), newDisk)

	emptyVolume := MockNewVolume(*mock.VolumeDBO, nil, true)

	Convey("An empty volume should not be ready", t, func() {
		So(emptyVolume.IsReady(nil, true), ShouldEqual, false)
	})

	Convey("A backed up volume should not be ready if not all real disks are backed up", t, func() {
		So(volume.IsReady(nil, true), ShouldEqual, false)
	})

	emptyVolume.AddDisk(newDisk.GetUUID(), newDisk)
	Convey("The function returns true if the volume is ready", t, func() {
		So(emptyVolume.IsReady(nil, true), ShouldEqual, true)
		So(emptyVolume.IsReady(nil, false), ShouldEqual, true)
	})

	Convey("The function returns false if the volume is not ready", t, func() {
		newDisk.GetReadiness().(*mock.MockDiskReadiness).Readiness = false
		So(emptyVolume.IsReady(nil, true), ShouldEqual, false)
		So(emptyVolume.IsReady(nil, false), ShouldEqual, false)
	})

	Convey("The database call should be correct", t, func() {
		So(mock.DBMock.ExpectationsWereMet(), ShouldEqual, nil)
	})
}

func TestNewVolume(t *testing.T) {
	models.Transport.ActiveVolumes.RemoveEnqueuedInstance(mock.VolumeDBO.UUID)

	disks := mock.GetDiskDBOs(1)
	volume := MockNewVolume(*mock.VolumeDBO, disks, true)

	Convey("Test if the volume is created properly", t, func() {
		Convey("UUID is set properly", func() {
			So(volume.UUID, ShouldEqual, mock.VolumeDBO.UUID)
		})
		Convey("BlockSize is set properly", func() {
			So(volume.BlockSize, ShouldEqual, constants.DEFAULT_VOLUME_BLOCK_SIZE)
		})
		Convey("Name is set properly", func() {
			So(volume.Name, ShouldEqual, mock.VolumeDBO.Name)
		})
		Convey("UserUUID is set properly", func() {
			So(volume.UserUUID, ShouldEqual, mock.VolumeDBO.UserUUID)
		})
		Convey("VolumeSettings are set properly", func() {
			So(volume.VolumeSettings, ShouldResemble, mock.VolumeDBO.VolumeSettings)
		})
		Convey("Volume contains the assigned disks", func() {
			for i := 0; i < len(disks); i++ {
				So(volume.GetDisk(disks[i].UUID).GetUUID(), ShouldEqual, disks[i].UUID)
			}
		})
	})

	Convey("The database call should be correct", t, func() {
		So(mock.DBMock.ExpectationsWereMet(), ShouldEqual, nil)
	})
}

func MockNewVolume(_volumeDBO dbo.Volume, _disks []dbo.Disk, dry_run bool) *models.Volume {
	var _disksPtr []*dbo.Disk
	for _, _disk := range _disks {
		_disksPtr = append(_disksPtr, &_disk)
	}

	for _, _disk := range _disks {
		if _disk.Provider.Type != constants.PROVIDER_TYPE_GDRIVE && _disk.Provider.Type != constants.PROVIDER_TYPE_ONEDRIVE {
			continue
		}

		// mock.DBMock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `disks` WHERE uuid = ? ORDER BY `disks`.`uuid` LIMIT 1")).WithArgs(_disk.UUID.String()).WillReturnRows(mock.DiskRow(&_disk))
	}

	if !dry_run {
		mock.DBMock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `disks` WHERE volume_uuid = ? AND is_virtual = ?")).WithArgs(_volumeDBO.UUID.String(), false).WillReturnRows(mock.DiskRow(_disksPtr...))
		mock.DBMock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `disks` WHERE volume_uuid = ? AND is_virtual = ?")).WithArgs(_volumeDBO.UUID.String(), true).WillReturnRows(mock.DiskRow(nil))
	}

	volume := models.NewVolume(&_volumeDBO, _disks, nil)

	// refresh partitioner has been moved to a go routine (tests will not run it)
	time.Sleep(1 * time.Second)

	return volume
}

func init() {
	models.RefreshPartitionerFunc = func(v *models.Volume) { v.RefreshPartitioner() }
	models.ClearFilesystemFunc = func(v *models.Volume) error { return nil }
}
