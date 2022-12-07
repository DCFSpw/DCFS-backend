package unit

import (
	"dcfs/constants"
	"dcfs/models"
	"dcfs/requests"
	"dcfs/test/unit/mock"
	_ "dcfs/util/logger"
	"github.com/google/uuid"
	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

func TestBalancedPartitioner_EmptyVolume(t *testing.T) {
	disks := mock.GetDiskDBOs(0)

	mock.VolumeDBO.VolumeSettings.FilePartition = constants.PARTITION_TYPE_BALANCED
	volume := MockNewVolume(*mock.VolumeDBO, disks)
	partitioner := volume.GetPartitioner()

	Convey("Test if balanced partitioner returns nil when volume is empty", t, func() {
		disk := partitioner.AssignDisk(constants.DEFAULT_VOLUME_BLOCK_SIZE)

		So(disk, ShouldBeNil)
	})
}

func TestBalancedPartitioner_FullDisks(t *testing.T) {
	disks := mock.GetDiskDBOs(2)
	size := uint64(1024 * constants.DEFAULT_VOLUME_BLOCK_SIZE)
	for i, _ := range disks {
		disks[i].TotalSpace = size
		disks[i].UsedSpace = size
		disks[i].FreeSpace = 0
	}

	mock.VolumeDBO.VolumeSettings.FilePartition = constants.PARTITION_TYPE_BALANCED
	volume := MockNewVolume(*mock.VolumeDBO, disks)
	partitioner := volume.GetPartitioner()

	Convey("Test if balanced partitioner returns nil when all disks are full", t, func() {
		disk := partitioner.AssignDisk(constants.DEFAULT_VOLUME_BLOCK_SIZE)

		So(disk, ShouldBeNil)
	})
}

func TestBalancedPartitioner_AssignEvenBlocks(t *testing.T) {
	disks := mock.GetDiskDBOs(2)

	mock.VolumeDBO.VolumeSettings.FilePartition = constants.PARTITION_TYPE_BALANCED
	volume := MockNewVolume(*mock.VolumeDBO, disks)
	partitioner := volume.GetPartitioner()

	Convey("Test if balanced partitioner assigns disks correctly for even number of blocks", t, func() {
		numberOfBlocks := 10

		firstDisk := 0
		secondDisk := 0

		for i := 0; i < numberOfBlocks; i++ {
			disk := partitioner.AssignDisk(constants.DEFAULT_VOLUME_BLOCK_SIZE)
			So(disk, ShouldNotBeNil)

			if disk.GetUUID() == disks[0].UUID {
				firstDisk++
			} else if disk.GetUUID() == disks[1].UUID {
				secondDisk++
			}
		}

		So(firstDisk, ShouldEqual, numberOfBlocks/2)
		So(secondDisk, ShouldEqual, numberOfBlocks/2)
	})
}

func TestBalancedPartitioner_AssignOddBlocks(t *testing.T) {
	disks := mock.GetDiskDBOs(2)

	mock.VolumeDBO.VolumeSettings.FilePartition = constants.PARTITION_TYPE_BALANCED
	volume := MockNewVolume(*mock.VolumeDBO, disks)
	partitioner := volume.GetPartitioner()

	Convey("Test if balanced partitioner assigns disks correctly for odd number of blocks", t, func() {
		numberOfBlocks := 9

		firstDisk := 0
		secondDisk := 0

		for i := 0; i < numberOfBlocks; i++ {
			disk := partitioner.AssignDisk(constants.DEFAULT_VOLUME_BLOCK_SIZE)
			So(disk, ShouldNotBeNil)

			if disk.GetUUID() == disks[0].UUID {
				firstDisk++
			} else if disk.GetUUID() == disks[1].UUID {
				secondDisk++
			}
		}

		So(firstDisk, ShouldEqual, secondDisk+1)
	})
}

func TestPriorityPartitioner_EmptyVolume(t *testing.T) {
	disks := mock.GetDiskDBOs(0)

	mock.VolumeDBO.VolumeSettings.FilePartition = constants.PARTITION_TYPE_PRIORITY
	volume := MockNewVolume(*mock.VolumeDBO, disks)
	partitioner := volume.GetPartitioner()

	Convey("Test if priority partitioner returns nil when volume is empty", t, func() {
		disk := partitioner.AssignDisk(constants.DEFAULT_VOLUME_BLOCK_SIZE)

		So(disk, ShouldBeNil)
	})
}

func TestPriorityPartitioner_FullDisks(t *testing.T) {
	disks := mock.GetDiskDBOs(2)
	size := uint64(1024 * constants.DEFAULT_VOLUME_BLOCK_SIZE)
	for i, _ := range disks {
		disks[i].TotalSpace = size
		disks[i].UsedSpace = size
	}

	mock.VolumeDBO.VolumeSettings.FilePartition = constants.PARTITION_TYPE_PRIORITY
	volume := MockNewVolume(*mock.VolumeDBO, disks)
	partitioner := volume.GetPartitioner()

	Convey("Test if priority partitioner returns nil when all disks are full", t, func() {
		disk := partitioner.AssignDisk(constants.DEFAULT_VOLUME_BLOCK_SIZE)

		So(disk, ShouldBeNil)
	})
}

func TestPriorityPartitioner_NotEnoughSpaceOnAllDisks(t *testing.T) {
	disks := mock.GetDiskDBOs(2)
	size := uint64(1024 * constants.DEFAULT_VOLUME_BLOCK_SIZE)
	for i, _ := range disks {
		disks[i].TotalSpace = size
		disks[i].UsedSpace = size - 16*uint64(constants.DEFAULT_VOLUME_BLOCK_SIZE)
	}

	mock.VolumeDBO.VolumeSettings.FilePartition = constants.PARTITION_TYPE_PRIORITY
	volume := MockNewVolume(*mock.VolumeDBO, disks)
	partitioner := volume.GetPartitioner()

	Convey("Test if priority partitioner returns nil when all disks are full", t, func() {
		disk := partitioner.AssignDisk(128 * constants.DEFAULT_VOLUME_BLOCK_SIZE)

		So(disk, ShouldBeNil)
	})
}

func TestPriorityPartitioner_AssignAllBlocksToFirstDisk(t *testing.T) {
	disks := mock.GetDiskDBOs(2)

	mock.VolumeDBO.VolumeSettings.FilePartition = constants.PARTITION_TYPE_PRIORITY
	volume := MockNewVolume(*mock.VolumeDBO, disks)
	partitioner := volume.GetPartitioner()

	Convey("Test if priority partitioner assigns all blocks to first disk", t, func() {
		numberOfBlocks := 10

		firstDisk := 0
		secondDisk := 0

		for i := 0; i < numberOfBlocks; i++ {
			disk := partitioner.AssignDisk(constants.DEFAULT_VOLUME_BLOCK_SIZE)
			So(disk, ShouldNotBeNil)

			if disk.GetUUID() == disks[0].UUID {
				firstDisk++
			} else if disk.GetUUID() == disks[1].UUID {
				secondDisk++
			}
		}

		So(firstDisk, ShouldEqual, numberOfBlocks)
		So(secondDisk, ShouldEqual, 0)
	})
}

func TestPriorityPartitioner_AssignBlocksToNextAvailableDisk(t *testing.T) {
	disks := mock.GetDiskDBOs(2)

	size := uint64(1024 * constants.DEFAULT_VOLUME_BLOCK_SIZE)
	disks[0].TotalSpace = size
	disks[0].UsedSpace = size - 2*16*uint64(constants.DEFAULT_VOLUME_BLOCK_SIZE)

	mock.VolumeDBO.VolumeSettings.FilePartition = constants.PARTITION_TYPE_PRIORITY
	volume := MockNewVolume(*mock.VolumeDBO, disks)
	partitioner := volume.GetPartitioner()

	Convey("Test if priority partitioner assigns disks correctly when current disk becomes full", t, func() {
		numberOfBlocks := 10

		firstDisk := 0
		secondDisk := 0

		for i := 0; i < numberOfBlocks; i++ {
			disk := partitioner.AssignDisk(16 * constants.DEFAULT_VOLUME_BLOCK_SIZE)
			So(disk, ShouldNotBeNil)

			if disk.GetUUID() == disks[0].UUID {
				firstDisk++
			} else if disk.GetUUID() == disks[1].UUID {
				secondDisk++
			}
		}

		So(firstDisk, ShouldEqual, 2)
		So(secondDisk, ShouldEqual, numberOfBlocks-2)
	})
}

func TestThroughputPartitioner_EmptyVolume(t *testing.T) {
	disks := mock.GetDiskDBOs(0)

	mock.VolumeDBO.VolumeSettings.FilePartition = constants.PARTITION_TYPE_THROUGHPUT
	volume := MockNewVolume(*mock.VolumeDBO, disks)
	partitioner := volume.GetPartitioner()

	Convey("Test if throughput partitioner returns nil when volume is empty", t, func() {
		disk := partitioner.AssignDisk(constants.DEFAULT_VOLUME_BLOCK_SIZE)

		So(disk, ShouldBeNil)
	})
}

func TestThroughputPartitioner_FullDisks(t *testing.T) {
	disks := mock.GetDiskDBOs(2)
	size := uint64(1024 * constants.DEFAULT_VOLUME_BLOCK_SIZE)
	for i, _ := range disks {
		disks[i].TotalSpace = size
		disks[i].UsedSpace = size
		disks[i].FreeSpace = 0
	}

	mock.VolumeDBO.VolumeSettings.FilePartition = constants.PARTITION_TYPE_THROUGHPUT
	volume := MockNewVolume(*mock.VolumeDBO, disks)
	partitioner := volume.GetPartitioner()

	Convey("Test if throughput partitioner returns nil when all disks are full", t, func() {
		disk := partitioner.AssignDisk(constants.DEFAULT_VOLUME_BLOCK_SIZE)

		So(disk, ShouldBeNil)
	})
}

func TestThroughputPartitioner_AssignBlocks(t *testing.T) {
	disks := mock.GetDiskDBOs(2)

	mock.VolumeDBO.VolumeSettings.FilePartition = constants.PARTITION_TYPE_THROUGHPUT
	volume := MockNewVolume(*mock.VolumeDBO, disks)
	partitioner := volume.GetPartitioner()

	Convey("Test if throughput partitioner assigns more blocks to faster disk", t, func() {
		numberOfBlocks := 10

		firstDisk := 0
		secondDisk := 0

		for i := 0; i < numberOfBlocks; i++ {
			disk := partitioner.AssignDisk(constants.DEFAULT_VOLUME_BLOCK_SIZE)
			So(disk, ShouldNotBeNil)

			if disk.GetUUID() == disks[0].UUID {
				firstDisk++
			} else if disk.GetUUID() == disks[1].UUID {
				secondDisk++
			}
		}

		if partitioner.(*models.ThroughputPartitioner).Weights[0] < partitioner.(*models.ThroughputPartitioner).Weights[1] {
			So(firstDisk, ShouldBeGreaterThanOrEqualTo, secondDisk)
		} else {
			So(firstDisk, ShouldBeLessThanOrEqualTo, secondDisk)
		}
	})
}

func TestPartitionerFactory(t *testing.T) {
	var balancedPartitioner models.BalancedPartitioner
	var priorityPartitioner models.PriorityPartitioner
	var throughputPartitioner models.ThroughputPartitioner

	disks := mock.GetDiskDBOs(0)
	volume := MockNewVolume(*mock.VolumeDBO, disks)

	Convey("Test if partitioner factory creates appropriate partitioner object", t, func() {
		p1 := models.CreatePartitioner(constants.PARTITION_TYPE_BALANCED, volume)
		p2 := models.CreatePartitioner(constants.PARTITION_TYPE_PRIORITY, volume)
		p3 := models.CreatePartitioner(constants.PARTITION_TYPE_THROUGHPUT, volume)
		p4 := models.CreatePartitioner(-1, volume)

		So(p1, ShouldHaveSameTypeAs, &balancedPartitioner)
		So(p2, ShouldHaveSameTypeAs, &priorityPartitioner)
		So(p3, ShouldHaveSameTypeAs, &throughputPartitioner)
		So(p4, ShouldBeNil)
	})
}

func TestAbstractPartitioner(t *testing.T) {
	Convey("Test if abstract partitioner panics on actual usage", t, func() {
		var p models.AbstractPartitioner

		So(p.FetchDisks, ShouldPanic)
		So(func() { p.AssignDisk(constants.DEFAULT_VOLUME_BLOCK_SIZE) }, ShouldPanic)
	})
}

func TestBalancedPartitioner_Integration(t *testing.T) {
	disks := mock.GetDiskDBOs(2)

	mock.VolumeDBO.VolumeSettings.FilePartition = constants.PARTITION_TYPE_BALANCED
	volume := MockNewVolume(*mock.VolumeDBO, disks)

	req := &requests.InitFileUploadRequest{
		VolumeUUID: volume.UUID.String(),
		RootUUID:   "",
		File: requests.FileDataRequest{
			Name: "test",
			Type: constants.FILE_TYPE_REGULAR,
			Size: 3*volume.BlockSize + (volume.BlockSize / 2),
		},
	}

	file := volume.FileUploadRequest(req, volume.UserUUID, uuid.Nil)

	Convey("Test if balanced partitioner assigns disks correctly in file upload request", t, func() {
		firstDisk := 0
		secondDisk := 0
		for _, b := range file.Blocks {
			if b.Disk.GetUUID() == disks[0].UUID {
				firstDisk++
			} else if b.Disk.GetUUID() == disks[1].UUID {
				secondDisk++
			}
		}

		So(firstDisk, ShouldEqual, secondDisk)
	})

	Convey("The database call should be correct", t, func() {
		So(mock.DBMock.ExpectationsWereMet(), ShouldEqual, nil)
	})
}

func TestPriorityPartitioner_Integration(t *testing.T) {
	disks := mock.GetDiskDBOs(2)

	mock.VolumeDBO.VolumeSettings.FilePartition = constants.PARTITION_TYPE_PRIORITY
	volume := MockNewVolume(*mock.VolumeDBO, disks)

	req := &requests.InitFileUploadRequest{
		VolumeUUID: volume.UUID.String(),
		RootUUID:   "",
		File: requests.FileDataRequest{
			Name: "test",
			Type: constants.FILE_TYPE_REGULAR,
			Size: 3*volume.BlockSize + (volume.BlockSize / 2),
		},
	}

	file := volume.FileUploadRequest(req, volume.UserUUID, uuid.Nil)

	Convey("Test if priority partitioner assigns disks correctly in file upload request", t, func() {
		firstDisk := 0
		secondDisk := 0
		for _, b := range file.Blocks {
			if b.Disk.GetUUID() == disks[0].UUID {
				firstDisk++
			} else if b.Disk.GetUUID() == disks[1].UUID {
				secondDisk++
			}
		}

		So(firstDisk, ShouldEqual, 4)
		So(secondDisk, ShouldEqual, 0)
	})

	Convey("The database call should be correct", t, func() {
		So(mock.DBMock.ExpectationsWereMet(), ShouldEqual, nil)
	})
}

func TestThroughputPartitioner_Integration(t *testing.T) {
	disks := mock.GetDiskDBOs(2)

	mock.VolumeDBO.VolumeSettings.FilePartition = constants.PARTITION_TYPE_THROUGHPUT
	volume := MockNewVolume(*mock.VolumeDBO, disks)

	req := &requests.InitFileUploadRequest{
		VolumeUUID: volume.UUID.String(),
		RootUUID:   "",
		File: requests.FileDataRequest{
			Name: "test",
			Type: constants.FILE_TYPE_REGULAR,
			Size: 3*volume.BlockSize + (volume.BlockSize / 2),
		},
	}

	file := volume.FileUploadRequest(req, volume.UserUUID, uuid.Nil)

	Convey("Test if throughput partitioner assigns disks correctly in file upload request", t, func() {
		firstDisk := 0
		secondDisk := 0
		for _, b := range file.Blocks {
			if b.Disk.GetUUID() == disks[0].UUID {
				firstDisk++
			} else if b.Disk.GetUUID() == disks[1].UUID {
				secondDisk++
			}
		}

		partitioner := volume.GetPartitioner()
		if partitioner.(*models.ThroughputPartitioner).Weights[0] < partitioner.(*models.ThroughputPartitioner).Weights[1] {
			So(firstDisk, ShouldBeGreaterThanOrEqualTo, secondDisk)
		} else {
			So(firstDisk, ShouldBeLessThanOrEqualTo, secondDisk)
		}
	})

	Convey("The database call should be correct", t, func() {
		So(mock.DBMock.ExpectationsWereMet(), ShouldEqual, nil)
	})

	models.Transport.ActiveVolumes.RemoveEnqueuedInstance(mock.VolumeUUID)
}
