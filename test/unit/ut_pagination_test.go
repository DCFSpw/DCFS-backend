package unit

import (
	"dcfs/models"
	"dcfs/test/unit/mock"
	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

func TestPaginate(t *testing.T) {
	objects := []interface{}{
		mock.NewMockDisk(),
		mock.NewMockDisk(),
		mock.NewMockDisk(),
		mock.NewMockDisk(),
		mock.NewMockDisk(),
		mock.NewMockDisk(),
		mock.NewMockDisk(),
		mock.NewMockDisk(),
		mock.NewMockDisk(),
		mock.NewMockDisk()}
	data := models.Paginate(objects, 1, 2)

	Convey("The pagination data has been split properly", t, func() {
		So(data.Pagination.PerPage, ShouldEqual, 2)
		So(data.Pagination.TotalPages, ShouldEqual, 5)
		So(data.Pagination.CurrentPage, ShouldEqual, 1)
		So(data.Pagination.RecordsOnPage, ShouldEqual, 2)
		So(data.Pagination.TotalRecords, ShouldEqual, 10)
	})

	data = models.Paginate(objects, 0, 2)

	Convey("The pagination data has been split properly", t, func() {
		So(data.Pagination.PerPage, ShouldEqual, 2)
		So(data.Pagination.TotalPages, ShouldEqual, 5)
		So(data.Pagination.CurrentPage, ShouldEqual, 0)
		So(data.Pagination.RecordsOnPage, ShouldEqual, 0)
		So(data.Pagination.TotalRecords, ShouldEqual, 10)
	})
}
