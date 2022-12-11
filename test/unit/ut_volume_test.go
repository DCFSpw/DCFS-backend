package unit

import (
	"crypto/rand"
	"dcfs/constants"
	"dcfs/db/dbo"
	"dcfs/models"
	"dcfs/requests"
	"dcfs/test/unit/mock"
	_ "dcfs/util/logger"
	"github.com/google/uuid"
	. "github.com/smartystreets/goconvey/convey"
	"io"
	"regexp"
	"testing"

	_ "dcfs/models/disk/FTPDisk"
	_ "dcfs/models/disk/GDriveDisk"
	_ "dcfs/models/disk/OneDriveDisk"
	_ "dcfs/models/disk/SFTPDisk"
)

func TestGetDisk(t *testing.T) {
	disks := mock.GetDiskDBOs(1)
	volume := MockNewVolume(*mock.VolumeDBO, disks)
	volume2 := MockNewVolume(*mock.VolumeDBO, nil)

	Convey("Should successfully retrieve the disk", t, func() {
		So(volume.GetDisk(disks[0].UUID).GetUUID(), ShouldEqual, disks[0].UUID)
	})

	Convey("Should return nil if there are no disks", t, func() {
		So(volume2.GetDisk(disks[0].UUID), ShouldEqual, nil)
	})

	Convey("The database call should be correct", t, func() {
		So(mock.DBMock.ExpectationsWereMet(), ShouldEqual, nil)
	})
}

func TestAddDisk(t *testing.T) {
	disks := mock.GetMockDisks(1)
	volume := MockNewVolume(*mock.VolumeDBO, nil)

	volume.AddDisk(disks[0].UUID, disks[0])

	Convey("The disk has been successfully added", t, func() {
		So(volume.GetDisk(disks[0].UUID).GetUUID(), ShouldEqual, disks[0].UUID)
	})
}

func TestDeleteDisk(t *testing.T) {
	disks := mock.GetDiskDBOs(1)
	volume := MockNewVolume(*mock.VolumeDBO, disks)
	volume2 := MockNewVolume(*mock.VolumeDBO, nil)

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
}

func TestFileUploadRequest(t *testing.T) {
	disks := mock.GetDiskDBOs(10)
	volume := MockNewVolume(*mock.VolumeDBO, disks)
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
}

func TestGetVolumeDBO(t *testing.T) {
	volume := models.NewVolume(mock.VolumeDBO, nil)
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
}

func TestRefreshPartitioner(t *testing.T) {
	volume := models.NewVolume(mock.VolumeDBO, nil)
	volume.RefreshPartitioner()

	Convey("This method is excluded from the Unit Tests", t, func() {
		So(true, ShouldEqual, true)
	})
}

func TestNewVolume(t *testing.T) {
	disks := mock.GetDiskDBOs(10)
	volume := MockNewVolume(*mock.VolumeDBO, disks)

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

func TestEncrypt(t *testing.T) {
	volume := MockNewVolume(*mock.VolumeDBO, nil)
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
}

func TestDecrypt(t *testing.T) {
	volume := MockNewVolume(*mock.VolumeDBO, nil)
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
}

func MockNewVolume(_volumeDBO dbo.Volume, _disks []dbo.Disk) *models.Volume {
	for _, _disk := range _disks {
		if _disk.Provider.Type != constants.PROVIDER_TYPE_GDRIVE && _disk.Provider.Type != constants.PROVIDER_TYPE_ONEDRIVE {
			continue
		}

		mock.DBMock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `disks` WHERE uuid = ? ORDER BY `disks`.`uuid` LIMIT 1")).WithArgs(_disk.UUID.String()).WillReturnRows(mock.DiskRow(&_disk))
	}

	return models.NewVolume(&_volumeDBO, _disks)
}
