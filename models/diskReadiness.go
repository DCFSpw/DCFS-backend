package models

import (
	"context"
	"sync"
	"time"
)

type DiskReadiness interface {
	IsReady(ctx context.Context) bool
	IsReadyForce(ctx context.Context) bool
	IsReadyForceNonBlocking(ctx context.Context) bool
}

type RealDiskReadiness struct {
	isReady    bool
	isReadyMtx sync.Mutex

	isReadyCheckQueued    bool
	isReadyCheckQueuedMtx sync.Mutex

	readinessChecker func(ctx context.Context) bool
	alivenessChecker func() bool
}

// IsReady - verifies if the disk is ready to be used and sets up its periodic check
//
// params:
//   - ctx context.Context: context of the request
//
// return type:
//   - bool: true if the disk is ready, false otherwise
func (dr *RealDiskReadiness) IsReady(ctx context.Context) bool {
	dr.isReadyMtx.Lock()
	defer dr.isReadyMtx.Unlock()

	if dr.isReady {
		// check readiness in 6 minutes
		dr.isReadyPeriodicCheck(3*time.Minute, ctx)
		return true
	}

	// check readiness now
	dr.isReadyPeriodicCheck(0, ctx)
	return false
}

// isReadyPeriodicCheck - verifies periodically if the disk is ready to be used and sets up its readiness flag
//
// params:
//   - timeout time.Duration: duration to wait before checking the disk readiness
//   - ctx context.Context: context of the request
func (dr *RealDiskReadiness) isReadyPeriodicCheck(timeout time.Duration, ctx context.Context) {
	dr.isReadyCheckQueuedMtx.Lock()
	defer dr.isReadyCheckQueuedMtx.Unlock()

	// check if another call to check readiness is queued
	if dr.isReadyCheckQueued {
		return
	}

	dr.isReadyCheckQueued = true
	go func() {
		time.Sleep(timeout)

		dr.isReadyMtx.Lock()
		defer dr.isReadyMtx.Unlock()

		dr.isReady = dr.readinessChecker(ctx)

		dr.isReadyCheckQueuedMtx.Lock()
		defer dr.isReadyCheckQueuedMtx.Unlock()

		dr.isReadyCheckQueued = false

		if !dr.alivenessChecker() {
			return
		}

		// run the periodic check again in the background
		go dr.isReadyPeriodicCheck(IsReadyPeriodicCheckInterval, ctx)
	}()
}

// IsReadyForce - verifies if the disk is ready right now (regardless of the periodic checks) and sets up its readiness flag
//
// params:
//   - ctx context.Context: context of the request
//
// return type:
//   - bool: true if the disk is ready, false otherwise
func (dr *RealDiskReadiness) IsReadyForce(ctx context.Context) bool {
	dr.isReadyMtx.Lock()
	defer dr.isReadyMtx.Unlock()

	dr.isReady = dr.readinessChecker(ctx)
	return dr.isReady
}

// IsReadyForceNonBlocking - verifies in background if the disk is ready to be used and sets up its readiness flag
//
// params:
//   - ctx context.Context: context of the request
//
// return type:
//   - bool: true if the disk is ready, false otherwise
func (dr *RealDiskReadiness) IsReadyForceNonBlocking(ctx context.Context) bool {
	dr.isReadyMtx.Lock()
	defer dr.isReadyMtx.Unlock()

	go func(ctx context.Context) {
		time.Sleep(time.Second)

		dr.isReadyMtx.Lock()
		defer dr.isReadyMtx.Unlock()

		dr.isReady = dr.readinessChecker(ctx)
	}(ctx)

	return dr.isReady
}

func NewRealDiskReadiness(readinessChecker func(ctx context.Context) bool, alivenessChecker func() bool) DiskReadiness {
	return &RealDiskReadiness{
		isReady:               true,
		isReadyMtx:            sync.Mutex{},
		isReadyCheckQueued:    false,
		isReadyCheckQueuedMtx: sync.Mutex{},
		readinessChecker:      readinessChecker,
		alivenessChecker:      alivenessChecker,
	}
}

type VirtualDiskReadiness struct {
	RealDiskReadinessObjects []DiskReadiness
}

func NewVirtualDiskReadiness(objects ...DiskReadiness) *VirtualDiskReadiness {
	vdr := &VirtualDiskReadiness{}

	if objects == nil || len(objects) == 0 {
		return vdr
	}

	arr := make([]DiskReadiness, 0)
	for _, obj := range objects {
		if obj == nil {
			continue
		}
		arr = append(arr, obj)
	}

	vdr.RealDiskReadinessObjects = arr
	return vdr
}

// forAll - perform operation on all real disks associated with the virtual disk
//
// params:
//   - op func(dr DiskReadiness) bool: operation to perform on each real disk
//
// return type:
//   - bool: true if the operation was successful on all real disks, false otherwise
func (vdr *VirtualDiskReadiness) forAll(op func(dr DiskReadiness) bool) bool {
	for _, obj := range vdr.RealDiskReadinessObjects {
		if op(obj) == false {
			return false
		}
	}

	return true
}

// IsReady - verifies if the virtual disk is ready to be used (all its real disks are ready)
//
// params:
//   - ctx context.Context: context of the request
//
// return type:
//   - bool: true if the virtual disk is ready, false otherwise
func (vdr *VirtualDiskReadiness) IsReady(ctx context.Context) bool {
	return vdr.forAll(func(dr DiskReadiness) bool { return dr.IsReady(ctx) })
}

// IsReadyForce - verifies if the virtual disk is ready right now (regardless of the periodic checks) and sets up its readiness flag
//
// params:
//   - ctx context.Context: context of the request
//
// return type:
//   - bool: true if the virtual disk is ready, false otherwise
func (vdr *VirtualDiskReadiness) IsReadyForce(ctx context.Context) bool {
	return vdr.forAll(func(dr DiskReadiness) bool { return dr.IsReadyForce(ctx) })
}

// IsReadyForceNonBlocking - verifies in background if the virtual disk is ready to be used and sets up its readiness flag
//
// params:
//   - ctx context.Context: context of the request
//
// return type:
//   - bool: true if the virtual disk is ready, false otherwise
func (vdr *VirtualDiskReadiness) IsReadyForceNonBlocking(ctx context.Context) bool {
	return vdr.forAll(func(dr DiskReadiness) bool { return dr.IsReadyForceNonBlocking(ctx) })
}

var IsReadyPeriodicCheckInterval time.Duration = 3 * time.Minute
