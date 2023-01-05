package unit

import (
	"dcfs/models"
	"dcfs/test/unit/mock"
	"github.com/google/uuid"
	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

func TestNewBlockFromDBO(t *testing.T) {
	blockDBO := mock.GetBlockDBOs(1, uuid.New(), uuid.New())[0]
	block := models.NewBlockFromDBO(&blockDBO)

	Convey("The block fields are reflected properly in the final object", t, func() {
		So(block.UUID, ShouldEqual, blockDBO.UUID)
		So(block.UserUUID, ShouldEqual, blockDBO.UserUUID)
		So(block.File, ShouldEqual, nil)
		So(block.Disk, ShouldEqual, nil)
		So(block.Size, ShouldEqual, blockDBO.Size)
		So(block.Checksum, ShouldEqual, blockDBO.Checksum)
		So(block.Status, ShouldEqual, 0)
		So(block.Order, ShouldEqual, blockDBO.Order)
	})
}
