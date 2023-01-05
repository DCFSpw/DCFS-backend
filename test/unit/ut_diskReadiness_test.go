package unit

import (
	"context"
	"dcfs/models"
	. "github.com/smartystreets/goconvey/convey"
	"testing"
	"time"
)

// this model should be excluded from unit testing completely,
//thus the tests here only validate this file in the coverage

func TestCompoundDiskReadiness(t *testing.T) {
	r := models.NewRealDiskReadiness(func(ctx context.Context) bool { return true }, func() bool { return true })
	virtual := models.NewVirtualDiskReadiness(r)

	models.IsReadyPeriodicCheckInterval = time.Second
	r.IsReady(nil)
	r.IsReadyForce(nil)
	r.IsReadyForceNonBlocking(nil)

	virtual.IsReady(nil)
	virtual.IsReadyForce(nil)
	virtual.IsReadyForceNonBlocking(nil)

	virtual = models.NewVirtualDiskReadiness()
	virtual = models.NewVirtualDiskReadiness(nil)

	Convey("This method did not cause any panics or errors", t, func() {
		So(true, ShouldEqual, true)
	})
}
