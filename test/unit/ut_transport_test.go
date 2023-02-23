package unit

import (
	"dcfs/constants"
	"dcfs/db/dbo"
	"dcfs/models"
	"dcfs/test/unit/mock"
	_ "dcfs/util/logger"
	"fmt"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	. "github.com/smartystreets/goconvey/convey"
	"regexp"
	"testing"
	"time"
)

var testInstance models.File = &models.RegularFile{AbstractFile: models.AbstractFile{UUID: uuid.New()}}

func TestMarkAsUsed(t *testing.T) {
	instances := new(models.ConcurrentInstances)
	instances.EnqueueInstance(testInstance.GetUUID(), testInstance)
	old := instances.Instances[testInstance.GetUUID()].Counter
	err := instances.MarkAsUsed(testInstance.GetUUID())

	Convey("Test instance should be correctly marked as used", t, func() {
		Convey("No error should be returned", func() {
			So(err, ShouldEqual, nil)
		})
		Convey("Counter should be increased by one", func() {
			So(instances.Instances[testInstance.GetUUID()].Counter, ShouldEqual, old+1)
		})
	})
	Convey("The method returns error on an unknown key", t, func() {
		So(instances.MarkAsUsed(uuid.New()), ShouldNotEqual, nil)
	})
	Convey("The database call should be correct", t, func() {
		So(mock.DBMock.ExpectationsWereMet(), ShouldEqual, nil)
	})
}

func TestMarkAsCompleted(t *testing.T) {
	models.Transport.WaitTime = 3 * time.Second

	instances := new(models.ConcurrentInstances)
	instances.EnqueueInstance(testInstance.GetUUID(), testInstance)

	// block instance
	err := instances.MarkAsUsed(testInstance.GetUUID())
	Convey("All errors should be nil", t, func() {
		So(err, ShouldEqual, nil)
	})

	// wait for the first timer to go off
	time.Sleep(4 * time.Second)

	Convey("Instance counter should be equal to 1", t, func() {
		So(instances.Instances[testInstance.GetUUID()].Counter, ShouldEqual, 1)
	})

	models.Transport.WaitTime = time.Second
	instances.MarkAsCompleted(testInstance.GetUUID())
	time.Sleep(2 * time.Second)

	Convey("Instance should be deleted", t, func() {
		So(instances.GetEnqueuedInstance(testInstance.GetUUID()), ShouldEqual, nil)
	})
	Convey("The method does not crash on an unknown key", t, func() {
		instances.MarkAsCompleted(uuid.New())

		time.Sleep(2 * time.Second)

		So(true, ShouldEqual, true)
	})
	Convey("The database call should be correct", t, func() {
		So(mock.DBMock.ExpectationsWereMet(), ShouldEqual, nil)
	})

	models.Transport.WaitTime = 6 * time.Minute
}

func TestEnqueueInstance(t *testing.T) {
	models.Transport.WaitTime = 3 * time.Second

	instances := new(models.ConcurrentInstances)
	instances.EnqueueInstance(testInstance.GetUUID(), testInstance)

	Convey("The test instance was properly placed in the collection", t, func() {
		So(instances.Instances[testInstance.GetUUID()].Instance, ShouldEqual, testInstance)
		So(instances.GetEnqueuedInstance(testInstance.GetUUID()), ShouldEqual, testInstance)
	})

	time.Sleep(4 * time.Second)

	Convey("The test instance should be properly deleted from the collection after the allotted time", t, func() {
		So(instances.GetEnqueuedInstance(testInstance.GetUUID()), ShouldEqual, nil)
	})
	Convey("The database call should be correct", t, func() {
		So(mock.DBMock.ExpectationsWereMet(), ShouldEqual, nil)
	})

	models.Transport.WaitTime = 6 * time.Minute
}

func TestGetEnqueuedInstance(t *testing.T) {
	instances := new(models.ConcurrentInstances)
	instances.EnqueueInstance(testInstance.GetUUID(), testInstance)

	Convey("The instance can be properly retrieved from the collection", t, func() {
		So(instances.GetEnqueuedInstance(testInstance.GetUUID()), ShouldEqual, testInstance)
	})

	instances.Instances = nil
	Convey("The method does not crash when the instances are a nil array", t, func() {
		So(instances.GetEnqueuedInstance(uuid.New()), ShouldEqual, nil)
	})

	Convey("The database call should be correct", t, func() {
		So(mock.DBMock.ExpectationsWereMet(), ShouldEqual, nil)
	})
}

func TestRemoveEnqueuedInstance(t *testing.T) {
	instances := new(models.ConcurrentInstances)
	instances.EnqueueInstance(testInstance.GetUUID(), testInstance)
	instances.RemoveEnqueuedInstance(testInstance.GetUUID())

	Convey("The test item should be successfully deleted", t, func() {
		So(instances.GetEnqueuedInstance(testInstance.GetUUID()), ShouldEqual, nil)
	})

	instances.Instances = nil
	Convey("The method does not crash when the instances are a nil array", t, func() {
		instances.RemoveEnqueuedInstance(uuid.New())
		So(true, ShouldEqual, true)
	})

	Convey("The database call should be correct", t, func() {
		So(mock.DBMock.ExpectationsWereMet(), ShouldEqual, nil)
	})
}

func TestVolumeKeepAlive(t *testing.T) {
	models.Transport.ActiveVolumes.RemoveEnqueuedInstance(mock.VolumeDBO.UUID)

	volume := MockNewVolume(*mock.VolumeDBO, nil, true)
	models.Transport.ActiveVolumes.EnqueueInstance(volume.UUID, volume)
	old := models.Transport.ActiveVolumes.Instances[volume.UUID].Counter

	models.Transport.VolumeKeepAlive(volume.UUID)

	Convey("The counter of an existing volume got updated", t, func() {
		So(models.Transport.ActiveVolumes.Instances[volume.UUID].Counter, ShouldEqual, old+1)
	})

	//mock.DBMock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `disks` WHERE volume_uuid = ?")).
	//	WithArgs(volume.UUID).
	//	WillReturnRows(mock.DiskRow(nil))
	mock.DBMock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `volumes` WHERE uuid = ? AND `volumes`.`deleted_at` IS NULL ORDER BY `volumes`.`uuid` LIMIT 1")).
		WithArgs(volume.UUID.String()).
		WillReturnRows(mock.VolumeRow(mock.VolumeDBO))

	// test db querying
	models.Transport.ActiveVolumes.RemoveEnqueuedInstance(volume.UUID)
	models.Transport.VolumeKeepAlive(volume.UUID)

	Convey("The volume from db should appear in the transport queue", t, func() {
		v := models.Transport.ActiveVolumes.GetEnqueuedInstance(volume.UUID).(*models.Volume)
		So(*v, ShouldResemble, *volume)
	})
	Convey("The database call should be correct", t, func() {
		So(mock.DBMock.ExpectationsWereMet(), ShouldEqual, nil)
	})

	models.Transport.ActiveVolumes.RemoveEnqueuedInstance(volume.UUID)
}

func TestGetVolume(t *testing.T) {
	volume := MockNewVolume(*mock.VolumeDBO, nil, true)
	models.Transport.ActiveVolumes.EnqueueInstance(volume.UUID, volume)

	Convey("The object should be properly queued from the transport", t, func() {
		So(models.Transport.ActiveVolumes.GetEnqueuedInstance(volume.UUID), ShouldEqual, volume)
	})
	Convey("The database call should be correct", t, func() {
		So(mock.DBMock.ExpectationsWereMet(), ShouldEqual, nil)
	})

	models.Transport.ActiveVolumes.RemoveEnqueuedInstance(volume.UUID)
}

func TestGetVolumes(t *testing.T) {
	models.Transport.ActiveVolumes.RemoveEnqueuedInstance(mock.VolumeDBO.UUID)

	_vol1DBO := *mock.VolumeDBO
	_vol2DBO := *mock.VolumeDBO
	_vol2DBO.UUID = uuid.New()

	vol1 := MockNewVolume(_vol1DBO, nil, true)
	vol2 := MockNewVolume(_vol2DBO, nil, true)

	models.Transport.ActiveVolumes.EnqueueInstance(vol1.UUID, vol1)
	models.Transport.ActiveVolumes.EnqueueInstance(vol2.UUID, vol2)

	mock.DBMock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `volumes` WHERE user_uuid = ?")).
		WithArgs(_vol1DBO.UserUUID).
		WillReturnRows(mock.VolumeRow(&_vol1DBO, &_vol2DBO))

	Convey("Should return correct two volumes for a given user", t, func() {
		arr := models.Transport.GetVolumes(vol1.UserUUID)
		So(arr, ShouldContain, vol1)
		So(arr, ShouldContain, vol2)
	})
	Convey("The database call should be correct", t, func() {
		So(mock.DBMock.ExpectationsWereMet(), ShouldEqual, nil)
	})

	models.Transport.ActiveVolumes.RemoveEnqueuedInstance(vol1.UUID)
	models.Transport.ActiveVolumes.RemoveEnqueuedInstance(vol2.UUID)
}

func TestFindEnqueuedDisk(t *testing.T) {
	disk := mock.GetMockDisks(1)[0]
	file := models.RegularFile{
		AbstractFile: models.AbstractFile{UUID: uuid.New()},
	}
	block := models.Block{
		UUID: uuid.New(),
		File: &file,
		Disk: disk,
	}
	file.Blocks = make(map[uuid.UUID]*models.Block)
	file.Blocks[block.UUID] = &block

	Convey("Return nil when not enqueued", t, func() {
		So(models.Transport.FindEnqueuedDisk(disk.UUID), ShouldEqual, nil)
	})

	models.Transport.FileDownloadQueue.EnqueueInstance(file.UUID, &file)
	Convey("Return the object from download queue", t, func() {
		So(models.Transport.FindEnqueuedDisk(disk.UUID), ShouldEqual, disk)
	})
	models.Transport.FileDownloadQueue.RemoveEnqueuedInstance(file.UUID)

	models.Transport.FileUploadQueue.EnqueueInstance(file.UUID, &file)
	Convey("Return the object from upload queue", t, func() {
		So(models.Transport.FindEnqueuedDisk(disk.UUID), ShouldEqual, disk)
	})
	Convey("The database call should be correct", t, func() {
		So(mock.DBMock.ExpectationsWereMet(), ShouldEqual, nil)
	})

	models.Transport.FileUploadQueue.RemoveEnqueuedInstance(file.UUID)
	models.Transport.FileDownloadQueue.RemoveEnqueuedInstance(file.UUID)
	models.Transport.ActiveVolumes.RemoveEnqueuedInstance(mock.VolumeUUID)
}

func TestFindEnqueuedVolume(t *testing.T) {
	volume := MockNewVolume(*mock.VolumeDBO, nil, true)
	disk := mock.GetMockDisks(1)[0]
	disk.Volume = volume
	file := models.RegularFile{
		AbstractFile: models.AbstractFile{UUID: uuid.New(), Volume: volume},
	}
	block := models.Block{
		UUID: uuid.New(),
		File: &file,
		Disk: disk,
	}
	file.Blocks = make(map[uuid.UUID]*models.Block)
	file.Blocks[block.UUID] = &block

	Convey("Return nil when not enqueued", t, func() {
		So(models.Transport.FindEnqueuedVolume(volume.UUID), ShouldEqual, nil)
	})

	models.Transport.FileDownloadQueue.EnqueueInstance(file.UUID, &file)
	Convey("Return the object from download queue", t, func() {
		So(models.Transport.FindEnqueuedVolume(volume.UUID), ShouldEqual, volume)
	})
	models.Transport.FileDownloadQueue.RemoveEnqueuedInstance(file.UUID)

	models.Transport.FileUploadQueue.EnqueueInstance(file.UUID, &file)
	Convey("Return the object from upload queue", t, func() {
		So(models.Transport.FindEnqueuedVolume(volume.UUID), ShouldEqual, volume)
	})
	Convey("The database call should be correct", t, func() {
		So(mock.DBMock.ExpectationsWereMet(), ShouldEqual, nil)
	})
	models.Transport.FileUploadQueue.RemoveEnqueuedInstance(file.UUID)
}

func TestDeleteVolume(t *testing.T) {
	/* standard case */
	volume := MockNewVolume(*mock.VolumeDBO, nil, false)

	models.Transport.ActiveVolumes.EnqueueInstance(volume.UUID, volume)
	_, err := models.Transport.DeleteVolume(volume.UUID)

	Convey("No error is returned", t, func() {
		So(err, ShouldEqual, nil)
	})

	mock.DBMock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `volumes` WHERE uuid = ? AND `volumes`.`deleted_at` IS NULL ORDER BY `volumes`.`uuid` LIMIT 1")).
		WithArgs(volume.UUID.String()).
		WillReturnRows(mock.VolumeRow(nil))

	Convey("Item successfully deleted from Transport", t, func() {
		So(models.Transport.GetVolume(volume.UUID), ShouldEqual, nil)
	})

	emptyVolume := MockNewVolume(*mock.VolumeDBO, nil, false)
	mock.DBMock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `volumes` WHERE uuid = ? AND `volumes`.`deleted_at` IS NULL ORDER BY `volumes`.`uuid` LIMIT 1")).
		WithArgs(emptyVolume.UUID.String()).
		WillReturnRows(mock.VolumeRow(nil))
	_, err = models.Transport.DeleteVolume(emptyVolume.UUID)

	Convey("The method returns an error when the volume cannot be found", t, func() {
		So(err, ShouldNotEqual, nil)
	})

	/* standard case with one disk */

	disk := mock.NewMockDisk()
	volume.AddDisk(disk.GetUUID(), disk)
	models.Transport.ActiveVolumes.EnqueueInstance(volume.UUID, volume)

	mock.DBMock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `blocks` WHERE disk_uuid = ?")).WithArgs(disk.GetUUID()).WillReturnError(fmt.Errorf("test_error"))
	_, err = models.Transport.DeleteVolume(volume.UUID)

	Convey("When the application fails to delete all the disks, the method should return an error", t, func() {
		So(err, ShouldNotEqual, nil)
	})

	mock.DBMock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `blocks` WHERE disk_uuid = ?")).WithArgs(disk.GetUUID()).WillReturnRows(mock.BlockRow(nil))
	mock.DBMock.ExpectBegin()
	mock.DBMock.ExpectExec(regexp.QuoteMeta("DELETE FROM `disks` WHERE `disks`.`uuid` = ?")).WithArgs(disk.GetUUID().String()).WillReturnResult(sqlmock.NewResult(1, 1))
	mock.DBMock.ExpectCommit()

	_, err = models.Transport.DeleteVolume(volume.UUID)

	Convey("The volume should be successfully deleted", t, func() {
		So(err, ShouldEqual, nil)
	})

	/* case with the backup volume */

	mockDisks := []models.Disk{mock.NewMockDisk(), mock.NewMockDisk()}
	volume = MockNewVolume(*mock.BackupVolumeDBO, nil, true)

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

	mock.DBMock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `blocks` WHERE disk_uuid = ?")).WithArgs(UUID).WillReturnRows(mock.BlockRow(nil))
	mock.DBMock.ExpectBegin()
	mock.DBMock.ExpectExec(regexp.QuoteMeta("DELETE FROM `disks` WHERE `disks`.`uuid` = ?")).WithArgs(UUID).WillReturnResult(sqlmock.NewResult(1, 1))
	mock.DBMock.ExpectCommit()

	mock.DBMock.ExpectBegin()
	mock.DBMock.ExpectExec(regexp.QuoteMeta("DELETE FROM `disks`")).WillReturnResult(sqlmock.NewResult(1, 1))
	mock.DBMock.ExpectCommit()

	models.Transport.ActiveVolumes.EnqueueInstance(volume.UUID, volume)
	_, err = models.Transport.DeleteVolume(volume.UUID)

	/* case when the ClearFilesystemFunc fails */
	oldClearFileSystemFunc := models.ClearFilesystemFunc
	models.ClearFilesystemFunc = func(v *models.Volume) error { return fmt.Errorf("test_error") }

	volume = MockNewVolume(*mock.VolumeDBO, nil, true)

	models.Transport.ActiveVolumes.EnqueueInstance(volume.UUID, volume)
	_, err = models.Transport.DeleteVolume(volume.UUID)

	Convey("No error is returned", t, func() {
		So(err, ShouldNotEqual, nil)
	})

	models.ClearFilesystemFunc = oldClearFileSystemFunc

	Convey("The database call should be correct", t, func() {
		So(mock.DBMock.ExpectationsWereMet(), ShouldEqual, nil)
	})
}

func TestTransportDeleteDisk(t *testing.T) {
	Convey("Without providing a disk for relocation, the method should return an eror", t, func() {
		_, err := models.Transport.DeleteDisk(nil, nil, constants.RELOCATION, nil)
		So(err, ShouldNotEqual, nil)
	})

	disks := []*mock.MockDisk{mock.NewMockDisk().(*mock.MockDisk), mock.NewMockDisk().(*mock.MockDisk)}
	mock.DBMock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `blocks` WHERE disk_uuid = ?")).WithArgs(disks[0].GetUUID()).WillReturnRows(mock.BlockRow(&dbo.Block{
		AbstractDatabaseObject: dbo.AbstractDatabaseObject{
			UUID: uuid.New(),
		},
		DiskUUID: disks[0].GetUUID(),
	}))

	/* return error on download failure */
	disks[0].DownloadSuccess = false

	Convey("The method should return an error when the download fails", t, func() {
		_, err := models.Transport.DeleteDisk(disks[0], nil, constants.RELOCATION, disks[1])
		So(err, ShouldNotEqual, nil)
	})

	mock.DBMock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `blocks` WHERE disk_uuid = ?")).WithArgs(disks[0].GetUUID()).WillReturnRows(mock.BlockRow(&dbo.Block{
		AbstractDatabaseObject: dbo.AbstractDatabaseObject{
			UUID: uuid.New(),
		},
		DiskUUID: disks[0].GetUUID(),
	}))

	/* return error on upload failure */
	disks[0].DownloadSuccess = true
	disks[1].UploadSuccess = false

	Convey("The method should return an error when the upload fails", t, func() {
		_, err := models.Transport.DeleteDisk(disks[0], nil, constants.RELOCATION, disks[1])
		So(err, ShouldNotEqual, nil)
	})

	mock.DBMock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `blocks` WHERE disk_uuid = ?")).WithArgs(disks[0].GetUUID()).WillReturnRows(mock.BlockRow(&dbo.Block{
		AbstractDatabaseObject: dbo.AbstractDatabaseObject{
			UUID: uuid.New(),
		},
		DiskUUID: disks[0].GetUUID(),
	}))

	/* return error on db failure */
	disks[1].UploadSuccess = true
	mock.DBMock.ExpectBegin()
	mock.DBMock.ExpectExec("UPDATE `blocks`").WillReturnError(fmt.Errorf("test_error"))
	mock.DBMock.ExpectRollback()

	Convey("The method should return an error when a call to update info in the db fails", t, func() {
		_, err := models.Transport.DeleteDisk(disks[0], nil, constants.RELOCATION, disks[1])
		time.Sleep(time.Second)
		So(err, ShouldNotEqual, nil)
	})

	mock.DBMock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `blocks` WHERE disk_uuid = ?")).WithArgs(disks[0].GetUUID()).WillReturnRows(mock.BlockRow(&dbo.Block{
		AbstractDatabaseObject: dbo.AbstractDatabaseObject{
			UUID: uuid.New(),
		},
		DiskUUID: disks[0].GetUUID(),
	}))

	/* return error on remove failure */
	disks[0].RemoveSuccess = false
	mock.DBMock.ExpectBegin()
	mock.DBMock.ExpectExec("UPDATE `blocks`").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.DBMock.ExpectCommit()

	Convey("The method should return an error when a call to update info in the db fails", t, func() {
		_, err := models.Transport.DeleteDisk(disks[0], nil, constants.RELOCATION, disks[1])
		So(err, ShouldNotEqual, nil)
	})

	mock.DBMock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `blocks` WHERE disk_uuid = ?")).WithArgs(disks[0].GetUUID()).WillReturnRows(mock.BlockRow(&dbo.Block{
		AbstractDatabaseObject: dbo.AbstractDatabaseObject{
			UUID: uuid.New(),
		},
		DiskUUID: disks[0].GetUUID(),
	}))

	/* return error on db failure */
	disks[0].RemoveSuccess = true
	mock.DBMock.ExpectBegin()
	mock.DBMock.ExpectExec("DELETE FROM `blocks`").WillReturnError(fmt.Errorf("test_error"))
	mock.DBMock.ExpectRollback()

	Convey("The method should return an error when a call to update info in the db fails", t, func() {
		_, err := models.Transport.DeleteDisk(disks[0], nil, constants.DELETION, disks[1])
		So(err, ShouldNotEqual, nil)
	})

	mock.DBMock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `blocks` WHERE disk_uuid = ?")).WithArgs(disks[0].GetUUID()).WillReturnRows(mock.BlockRow(&dbo.Block{
		AbstractDatabaseObject: dbo.AbstractDatabaseObject{
			UUID: uuid.New(),
		},
		DiskUUID: disks[0].GetUUID(),
	}))

	/* return error on db failure */
	disks[0].RemoveSuccess = true
	mock.DBMock.ExpectBegin()
	mock.DBMock.ExpectExec("DELETE FROM `blocks`").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.DBMock.ExpectCommit()

	mock.DBMock.ExpectBegin()
	mock.DBMock.ExpectExec("DELETE FROM `disks`").WillReturnError(fmt.Errorf("test_error"))
	mock.DBMock.ExpectRollback()

	Convey("The method should return an error when a call to update info in the db fails", t, func() {
		_, err := models.Transport.DeleteDisk(disks[0], nil, constants.DELETION, disks[1])
		So(err, ShouldNotEqual, nil)
	})

	volume := MockNewVolume(*mock.VolumeDBO, nil, true)
	volume.AddDisk(disks[0].GetUUID(), disks[0])

	mock.DBMock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `blocks` WHERE disk_uuid = ?")).WithArgs(disks[0].GetUUID()).WillReturnRows(mock.BlockRow(&dbo.Block{
		AbstractDatabaseObject: dbo.AbstractDatabaseObject{
			UUID: uuid.New(),
		},
		DiskUUID: disks[0].GetUUID(),
	}))

	/* standard case */
	mock.DBMock.ExpectBegin()
	mock.DBMock.ExpectExec("DELETE FROM `blocks`").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.DBMock.ExpectCommit()

	mock.DBMock.ExpectBegin()
	mock.DBMock.ExpectExec("DELETE FROM `disks`").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.DBMock.ExpectCommit()

	Convey("The method should not return any errors if it succeeded", t, func() {
		_, err := models.Transport.DeleteDisk(disks[0], volume, constants.DELETION, disks[1])
		So(err, ShouldEqual, nil)
	})

	/* backed up disk */
	volume = MockNewVolume(*mock.BackupVolumeDBO, nil, true)
	for _, _mdisk := range disks {
		volume.AddDisk(_mdisk.GetUUID(), _mdisk)
	}

	mock.DBMock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `disks` WHERE volume_uuid = ? AND is_virtual = ? AND virtual_disk_uuid = ? ORDER BY `disks`.`uuid` LIMIT 1")).WithArgs(volume.UUID.String(), false, uuid.Nil).WillReturnRows(mock.DiskRow(&dbo.Disk{
		AbstractDatabaseObject: dbo.AbstractDatabaseObject{
			UUID: disks[1].GetUUID(),
		},
		VolumeUUID: volume.UUID,
		Volume:     *mock.BackupVolumeDBO,
	}))

	mock.DBMock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `providers` WHERE type = ?")).WithArgs(constants.PROVIDER_TYPE_RAID1).WillReturnRows(mock.ProviderRow(&dbo.Provider{
		AbstractDatabaseObject: dbo.AbstractDatabaseObject{
			UUID: uuid.UUID{},
		},
		Type: constants.PROVIDER_TYPE_RAID1,
	}))

	mock.DBMock.ExpectBegin()
	mock.DBMock.ExpectExec("INSERT INTO `disks`").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.DBMock.ExpectCommit()

	mock.DBMock.ExpectBegin()
	mock.DBMock.ExpectExec(regexp.QuoteMeta("UPDATE `disks`")).WillReturnResult(sqlmock.NewResult(1, 1))
	mock.DBMock.ExpectCommit()

	UUID, _ := volume.GenerateVirtualDisk(disks[0])
	disk := volume.GetDisk(UUID)

	mock.DBMock.ExpectBegin()
	mock.DBMock.ExpectExec("DELETE FROM `blocks`").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.DBMock.ExpectCommit()

	mock.DBMock.ExpectBegin()
	mock.DBMock.ExpectExec("DELETE FROM `disks`").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.DBMock.ExpectCommit()

	mock.DBMock.ExpectBegin()
	mock.DBMock.ExpectExec("DELETE FROM `disks`").WillReturnError(fmt.Errorf("Test_error"))
	mock.DBMock.ExpectRollback()

	mock.DBMock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `blocks` WHERE disk_uuid = ?")).WithArgs(disk.GetUUID()).WillReturnRows(mock.BlockRow(&dbo.Block{
		AbstractDatabaseObject: dbo.AbstractDatabaseObject{
			UUID: uuid.New(),
		},
		DiskUUID: disk.GetUUID(),
	}))

	Convey("The method should return an error when a call to delete the real disks from the db fails", t, func() {
		_, err := models.Transport.DeleteDisk(disk, volume, constants.DELETION, disks[1])
		So(err, ShouldNotEqual, nil)
	})

	Convey("The database call should be correct", t, func() {
		So(mock.DBMock.ExpectationsWereMet(), ShouldEqual, nil)
	})
}

func TestTransportDeleteFile(t *testing.T) {
	volume := MockNewVolume(*mock.VolumeDBO, nil, true)
	disk := mock.NewMockDisk()
	volume.AddDisk(disk.GetUUID(), disk)
	file := models.RegularFile{
		AbstractFile: models.AbstractFile{
			UUID:   uuid.New(),
			Type:   constants.FILE_TYPE_REGULAR,
			Volume: volume,
		},
		Blocks: make(map[uuid.UUID]*models.Block),
	}
	block := models.Block{
		UUID: uuid.New(),
		File: &file,
		Disk: disk,
	}
	file.Blocks[block.UUID] = &block

	/* standard case */
	mock.DBMock.ExpectBegin()
	mock.DBMock.ExpectExec("DELETE FROM `blocks`").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.DBMock.ExpectCommit()

	mock.DBMock.ExpectBegin()
	mock.DBMock.ExpectExec("UPDATE `files`").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.DBMock.ExpectCommit()

	Convey("The method should not return any error on successful completion", t, func() {
		_, err := models.Transport.DeleteFile(&file, volume)
		So(err, ShouldEqual, nil)
	})

	/* the method should fail when the disk remove fails */
	disk.(*mock.MockDisk).RemoveSuccess = false

	Convey("The method should return an error if the disk fails to delete the block", t, func() {
		_, err := models.Transport.DeleteFile(&file, volume)
		So(err, ShouldNotEqual, nil)
	})

	/* the method should fail when it fails to remove the block from the db */
	mock.DBMock.ExpectBegin()
	mock.DBMock.ExpectExec("DELETE FROM `blocks`").WillReturnError(fmt.Errorf("test_error"))
	mock.DBMock.ExpectRollback()

	disk.(*mock.MockDisk).RemoveSuccess = true

	Convey("The method should return an error if it fails to remove the block from the db", t, func() {
		_, err := models.Transport.DeleteFile(&file, volume)
		So(err, ShouldNotEqual, nil)
	})

	/* the method should fail when it fails to remove the file from the db */
	mock.DBMock.ExpectBegin()
	mock.DBMock.ExpectExec("DELETE FROM `blocks`").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.DBMock.ExpectCommit()

	mock.DBMock.ExpectBegin()
	mock.DBMock.ExpectExec("UPDATE `files`").WillReturnError(fmt.Errorf("test_error"))
	mock.DBMock.ExpectRollback()

	Convey("The method should return an error if it fails to remove the file from the db", t, func() {
		_, err := models.Transport.DeleteFile(&file, volume)
		So(err, ShouldNotEqual, nil)
	})

	Convey("The database call should be correct", t, func() {
		So(mock.DBMock.ExpectationsWereMet(), ShouldEqual, nil)
	})
}
