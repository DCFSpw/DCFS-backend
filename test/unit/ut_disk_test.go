package unit

import (
	"dcfs/constants"
	"dcfs/db/dbo"
	"dcfs/models"
	"dcfs/test/unit/mock"
	_ "dcfs/util/logger"
	"fmt"
	. "github.com/smartystreets/goconvey/convey"
	"regexp"
	"testing"
)

func TestCreateDiskAndGetDiskDbo(t *testing.T) {
	disks := mock.GetDiskDBOs(1)
	volume := mock.Volume

	disk := models.CreateDisk(models.CreateDiskMetadata{
		Disk:   &disks[0],
		Volume: volume,
	})
	diskDBO := disk.GetDiskDBO(disks[0].UserUUID, disks[0].ProviderUUID, disks[0].VolumeUUID)

	Convey("The object is initialized correctly", t, func() {
		So(disks[0].UUID, ShouldEqual, diskDBO.UUID)
		So(disks[0].UserUUID, ShouldEqual, diskDBO.UserUUID)
		So(disks[0].VolumeUUID, ShouldEqual, diskDBO.VolumeUUID)
		So(disks[0].ProviderUUID, ShouldEqual, diskDBO.ProviderUUID)
		So(disks[0].Credentials, ShouldEqual, diskDBO.Credentials)
		So(disks[0].Name, ShouldEqual, diskDBO.Name)
		So(disks[0].UsedSpace, ShouldEqual, diskDBO.UsedSpace)
		So(disks[0].TotalSpace, ShouldEqual, diskDBO.TotalSpace)
		So(disks[0].CreatedAt, ShouldEqual, diskDBO.CreatedAt)
	})
	Convey("The database call should be correct", t, func() {
		So(mock.DBMock.ExpectationsWereMet(), ShouldEqual, nil)
	})
}

func TestCreateDiskFromUUID(t *testing.T) {
	disks := mock.GetDiskDBOs(1)
	fileDBO := mock.GetFileDBO(disks[0].UUID, constants.FILE_TYPE_REGULAR, constants.DEFAULT_VOLUME_BLOCK_SIZE)
	provider, _ := mock.GetProviderDBO(disks[0].Provider.Type)
	// volume := MockNewVolume(*mock.VolumeDBO, disks, false)

	mock.DBMock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `disks` WHERE uuid = ? ORDER BY `disks`.`uuid` LIMIT 1")).
		WithArgs(disks[0].UUID).
		WillReturnRows(mock.DiskRow(&disks[0]))

	Convey("CreateDiskFromUUID function works correctly", t, func() {
		Convey("Should return nil if the disk does not exist", func() {
			mock.DBMock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `disks` WHERE uuid = ?")).
				WithArgs(disks[0].UUID).
				WillReturnError(fmt.Errorf(""))
			So(models.CreateDiskFromUUID(disks[0].UUID), ShouldEqual, nil)
		})
		Convey("Should return a disk if it is in the db", func() {
			disk := CreateDummyDisk(&disks[0], provider, false)
			So(disk.GetUUID(), ShouldEqual, disks[0].UUID)
		})
		Convey("All db expectations were met", func() {
			So(mock.DBMock.ExpectationsWereMet(), ShouldEqual, nil)
		})
	})

	models.Transport.FileDownloadQueue.RemoveEnqueuedInstance(fileDBO.UUID)
	models.Transport.ActiveVolumes.RemoveEnqueuedInstance(mock.VolumeUUID)
}

func TestComputeFreeSpace(t *testing.T) {
	gdriveDisks := mock.GetSpecifiedDisksDBO(1, constants.PROVIDER_TYPE_GDRIVE)
	onedriveDisks := mock.GetSpecifiedDisksDBO(1, constants.PROVIDER_TYPE_ONEDRIVE)
	sftpDisks := mock.GetSpecifiedDisksDBO(1, constants.PROVIDER_TYPE_SFTP)

	gdriveProvider, _ := mock.GetProviderDBO(gdriveDisks[0].Provider.Type)
	onedriveProvider, _ := mock.GetProviderDBO(onedriveDisks[0].Provider.Type)
	sftpProvider, _ := mock.GetProviderDBO(sftpDisks[0].Provider.Type)

	gdriveDisk := CreateDummyDisk(&gdriveDisks[0], gdriveProvider, false)
	onedriveDisk := CreateDummyDisk(&onedriveDisks[0], onedriveProvider, false)
	sftpDisk := CreateDummyDisk(&sftpDisks[0], sftpProvider, false)

	Convey("ComputeFreeSpace works for the Google drive provider", t, func() {
		mock.DBMock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `disks` WHERE uuid = ? ORDER BY `disks`.`uuid` LIMIT 1")).
			WithArgs(gdriveDisk.GetUUID().String()).
			WillReturnRows(mock.DiskRow(&gdriveDisks[0]))
		So(models.ComputeFreeSpace(gdriveDisk), ShouldBeGreaterThan, 0)
	})
	Convey("ComputeFreeSpace works for the OneDrive drive provider", t, func() {
		mock.DBMock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `disks` WHERE uuid = ? ORDER BY `disks`.`uuid` LIMIT 1")).
			WithArgs(onedriveDisk.GetUUID().String()).
			WillReturnRows(mock.DiskRow(&onedriveDisks[0]))
		So(models.ComputeFreeSpace(onedriveDisk), ShouldBeGreaterThan, 0)
	})
	Convey("ComputeFreeSpace works for the SFTP drive provider", t, func() {
		So(models.ComputeFreeSpace(sftpDisk), ShouldBeGreaterThan, 0)
	})
	Convey("All db expectations were met", t, func() {
		So(mock.DBMock.ExpectationsWereMet(), ShouldEqual, nil)
	})

	models.Transport.ActiveVolumes.RemoveEnqueuedInstance(mock.VolumeUUID)
}

func TestMeasureDiskThroughput(t *testing.T) {
	gdriveDisks := mock.GetSpecifiedDisksDBO(1, constants.PROVIDER_TYPE_GDRIVE)
	onedriveDisks := mock.GetSpecifiedDisksDBO(1, constants.PROVIDER_TYPE_ONEDRIVE)
	sftpDisks := mock.GetSpecifiedDisksDBO(1, constants.PROVIDER_TYPE_SFTP)

	gdriveProvider, _ := mock.GetProviderDBO(gdriveDisks[0].Provider.Type)
	onedriveProvider, _ := mock.GetProviderDBO(onedriveDisks[0].Provider.Type)
	sftpProvider, _ := mock.GetProviderDBO(sftpDisks[0].Provider.Type)

	gdriveDisk := CreateDummyDisk(&gdriveDisks[0], gdriveProvider, false)
	onedriveDisk := CreateDummyDisk(&onedriveDisks[0], onedriveProvider, false)
	sftpDisk := CreateDummyDisk(&sftpDisks[0], sftpProvider, false)

	Convey("MeasureDiskThroughput works for the Google drive provider", t, func() {
		mock.DBMock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `disks` WHERE uuid = ? ORDER BY `disks`.`uuid` LIMIT 1")).
			WithArgs(gdriveDisk.GetUUID().String()).
			WillReturnRows(mock.DiskRow(&gdriveDisks[0]))
		mock.DBMock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `disks` WHERE uuid = ? ORDER BY `disks`.`uuid` LIMIT 1")).
			WithArgs(gdriveDisk.GetUUID().String()).
			WillReturnRows(mock.DiskRow(&gdriveDisks[0]))
		So(models.MeasureDiskThroughput(gdriveDisk), ShouldBeGreaterThan, 0)
	})
	Convey("MeasureDiskThroughput works for the OneDrive drive provider", t, func() {
		mock.DBMock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `disks` WHERE uuid = ? ORDER BY `disks`.`uuid` LIMIT 1")).
			WithArgs(onedriveDisk.GetUUID().String()).
			WillReturnRows(mock.DiskRow(&onedriveDisks[0]))
		mock.DBMock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `disks` WHERE uuid = ? ORDER BY `disks`.`uuid` LIMIT 1")).
			WithArgs(onedriveDisk.GetUUID().String()).
			WillReturnRows(mock.DiskRow(&onedriveDisks[0]))
		So(models.MeasureDiskThroughput(onedriveDisk), ShouldBeGreaterThan, 0)
	})
	Convey("MeasureDiskThroughput works for the SFTP drive provider", t, func() {
		So(models.MeasureDiskThroughput(sftpDisk), ShouldBeGreaterThan, 0)
	})
	Convey("All db expectations were met", t, func() {
		So(mock.DBMock.ExpectationsWereMet(), ShouldEqual, nil)
	})

	models.Transport.ActiveVolumes.RemoveEnqueuedInstance(mock.VolumeUUID)
}

func CreateDummyDisk(disk *dbo.Disk, provider *dbo.Provider, dry_run bool) models.Disk {
	mock.DBMock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `disks` WHERE uuid = ?")).
		WithArgs(disk.UUID).
		WillReturnRows(mock.DiskRow(disk))
	mock.DBMock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `providers` WHERE `providers`.`uuid` = ?")).
		WithArgs(disk.ProviderUUID).
		WillReturnRows(mock.ProviderRow(provider))
	mock.DBMock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `volumes` WHERE `volumes`.`uuid` = ? AND `volumes`.`deleted_at` IS NULL")).
		WithArgs(disk.VolumeUUID).
		WillReturnRows(mock.VolumeRow(nil))
	mock.DBMock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `volumes` WHERE uuid = ? AND `volumes`.`deleted_at` IS NULL ORDER BY `volumes`.`uuid` LIMIT 1")).
		WithArgs(disk.VolumeUUID).
		WillReturnRows(mock.VolumeRow(mock.VolumeDBO))
	mock.DBMock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `disks` WHERE volume_uuid = ? AND is_virtual = ?")).
		WithArgs(disk.VolumeUUID, false).
		WillReturnRows(mock.DiskRow(disk))
	mock.DBMock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `disks` WHERE volume_uuid = ? AND is_virtual = ?")).
		WithArgs(disk.VolumeUUID, true).
		WillReturnRows(mock.VolumeRow(nil))
	//mock.DBMock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `disks` WHERE volume_uuid = ?")).
	//	WithArgs(disk.VolumeUUID).
	//	WillReturnRows(mock.DiskRow(disk))
	//mock.DBMock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `users` WHERE `users`.`uuid` = ?")).
	//	WithArgs(disk.UserUUID).
	//	WillReturnRows(mock.UserRow(nil))

	if dry_run {
		return nil
	}

	_disk := models.CreateDiskFromUUID(disk.UUID)
	models.Transport.ActiveVolumes.RemoveEnqueuedInstance(disk.VolumeUUID)

	return _disk
}
