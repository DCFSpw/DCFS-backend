package models

import (
	"dcfs/models/disk/GDriveDisk"
	"github.com/google/uuid"
	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

func TestDummyPartitioner(t *testing.T) {
	var volume Volume = Volume{}
	var partitioner *DummyPartitioner = NewDummyPartitioner(&volume)
	var firstDisk *GDriveDisk.GDriveDisk = GDriveDisk.NewGDriveDisk()
	var secondDisk *GDriveDisk.GDriveDisk = GDriveDisk.NewGDriveDisk()

	firstDisk.SetUUID(uuid.New())
	secondDisk.SetUUID(uuid.New())

	volume.AddDisk(firstDisk.GetUUID(), firstDisk)
	volume.AddDisk(secondDisk.GetUUID(), secondDisk)

	Convey("Test if partitioner assigns disks correctly", t, func() {
		Convey("Second disk for the first time", func() {
			So(partitioner.AssignDisk(0), ShouldEqual, secondDisk)
		})
		Convey("First disk for the second time", func() {
			So(partitioner.AssignDisk(0), ShouldEqual, firstDisk)
		})
		Convey("Second disk for the third time", func() {
			So(partitioner.AssignDisk(0), ShouldEqual, secondDisk)
		})
	})
}
