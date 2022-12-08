package unit

import (
	"dcfs/models"
	"dcfs/test/unit/mock"
	_ "dcfs/util/logger"
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

	models.Transport.WaitTime = 6 * time.Minute
}

func TestGetEnqueuedInstance(t *testing.T) {
	instances := new(models.ConcurrentInstances)
	instances.EnqueueInstance(testInstance.GetUUID(), testInstance)

	Convey("The instance can be properly retrieved from the collection", t, func() {
		So(instances.GetEnqueuedInstance(testInstance.GetUUID()), ShouldEqual, testInstance)
	})
}

func TestRemoveEnqueuedInstance(t *testing.T) {
	instances := new(models.ConcurrentInstances)
	instances.EnqueueInstance(testInstance.GetUUID(), testInstance)
	instances.RemoveEnqueuedInstance(testInstance.GetUUID())

	Convey("The test item should be successfully deleted", t, func() {
		So(instances.GetEnqueuedInstance(testInstance.GetUUID()), ShouldEqual, nil)
	})
}

func TestVolumeKeepAlive(t *testing.T) {
	volume := MockNewVolume(*mock.VolumeDBO, nil)
	models.Transport.ActiveVolumes.EnqueueInstance(volume.UUID, volume)
	old := models.Transport.ActiveVolumes.Instances[volume.UUID].Counter

	models.Transport.VolumeKeepAlive(volume.UUID)

	Convey("The counter of an existing volume got updated", t, func() {
		So(models.Transport.ActiveVolumes.Instances[volume.UUID].Counter, ShouldEqual, old+1)
	})

	mock.DBMock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `disks` WHERE volume_uuid = ?")).
		WithArgs(volume.UUID).
		WillReturnRows(mock.DiskRow(nil))
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
	volume := MockNewVolume(*mock.VolumeDBO, nil)
	models.Transport.ActiveVolumes.EnqueueInstance(volume.UUID, volume)

	Convey("The object should be properly queued from the transport", t, func() {
		So(models.Transport.ActiveVolumes.GetEnqueuedInstance(volume.UUID), ShouldEqual, volume)
	})

	models.Transport.ActiveVolumes.RemoveEnqueuedInstance(volume.UUID)
}

func TestGetVolumes(t *testing.T) {
	_vol1DBO := *mock.VolumeDBO
	_vol2DBO := *mock.VolumeDBO
	_vol2DBO.UUID = uuid.New()

	vol1 := MockNewVolume(_vol1DBO, nil)
	vol2 := MockNewVolume(_vol2DBO, nil)

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

	models.Transport.FileUploadQueue.RemoveEnqueuedInstance(file.UUID)
	models.Transport.FileDownloadQueue.RemoveEnqueuedInstance(file.UUID)
	models.Transport.ActiveVolumes.RemoveEnqueuedInstance(mock.VolumeUUID)
}

func TestFindEnqueuedVolume(t *testing.T) {
	volume := MockNewVolume(*mock.VolumeDBO, nil)
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
	models.Transport.FileUploadQueue.RemoveEnqueuedInstance(file.UUID)
}

func TestDeleteVolume(t *testing.T) {
	volume := MockNewVolume(*mock.VolumeDBO, nil)
	models.Transport.ActiveVolumes.EnqueueInstance(volume.UUID, volume)
	_, err := models.Transport.DeleteVolume(volume.UUID)

	Convey("No error is returned", t, func() {
		So(err, ShouldEqual, nil)
	})

	mock.DBMock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `disks` WHERE volume_uuid = ?")).
		WithArgs(volume.UUID).
		WillReturnRows(mock.DiskRow(nil))
	mock.DBMock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `volumes` WHERE uuid = ? AND `volumes`.`deleted_at` IS NULL ORDER BY `volumes`.`uuid` LIMIT 1")).
		WithArgs(volume.UUID.String()).
		WillReturnRows(mock.VolumeRow(nil))

	Convey("Item successfully deleted from Transport", t, func() {
		So(models.Transport.GetVolume(volume.UUID), ShouldEqual, nil)
	})
}
