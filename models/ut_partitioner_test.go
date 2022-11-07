package models

import (
	"dcfs/constants"
	"github.com/google/uuid"
	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

func TestBalancedPartitioner(t *testing.T) {
	var volume Volume = Volume{UUID: uuid.New()}
	var partitioner Partitioner = CreatePartitioner(constants.PARTITION_TYPE_BALANCED, &volume)
	volume.partitioner = partitioner

	var firstDisk Disk = &dummyDisk{}
	var secondDisk Disk = &dummyDisk{}

	firstDisk.SetUUID(uuid.New())
	secondDisk.SetUUID(uuid.New())

	volume.AddDisk(firstDisk.GetUUID(), firstDisk)
	volume.AddDisk(secondDisk.GetUUID(), secondDisk)

	Convey("Test if partitioner assigns disks correctly", t, func() {
		Convey("First disk for the first time", func() {
			So(partitioner.AssignDisk(0), ShouldEqual, firstDisk)
		})
		Convey("Second disk for the first time", func() {
			So(partitioner.AssignDisk(0), ShouldEqual, secondDisk)
		})
		Convey("First disk for the second time", func() {
			So(partitioner.AssignDisk(0), ShouldEqual, firstDisk)
		})
		Convey("Second disk for the second time", func() {
			So(partitioner.AssignDisk(0), ShouldEqual, secondDisk)
		})
	})
}
