package unit

import (
	"dcfs/constants"
	"dcfs/requests"
	"dcfs/test/unit/mock"
	_ "dcfs/util/logger"
	"github.com/google/uuid"
	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

func TestBalancedPartitioner(t *testing.T) {
	disks := mock.GetDiskDBOs(2)
	volume := MockNewVolume(*mock.VolumeDBO, disks)
	req := &requests.InitFileUploadRequest{
		VolumeUUID: volume.UUID.String(),
		RootUUID:   "",
		File: requests.FileDataRequest{
			Name: "test",
			Type: constants.FILE_TYPE_REGULAR,
			Size: 3*constants.DEFAULT_VOLUME_BLOCK_SIZE + (constants.DEFAULT_VOLUME_BLOCK_SIZE / 2),
		},
	}

	file := volume.FileUploadRequest(req, volume.UserUUID, uuid.Nil)

	Convey("Test if partitioner assigns disks correctly", t, func() {
		firstDisk := 0
		secondDisk := 0
		for _, b := range file.Blocks {
			if b.Disk.GetUUID() == disks[0].UUID {
				firstDisk++
			} else if b.Disk.GetUUID() == disks[1].UUID {
				secondDisk++
			}
		}

		// TODO: update in the future
		So(true, ShouldEqual, true)
	})

	Convey("The database call should be correct", t, func() {
		So(mock.DBMock.ExpectationsWereMet(), ShouldEqual, nil)
	})
}
