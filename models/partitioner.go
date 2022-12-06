package models

import (
	"dcfs/constants"
	"dcfs/util/logger"
	"sort"
	"strconv"
)

type Partitioner interface {
	AssignDisk(int) Disk
	FetchDisks(disks []Disk)
}

// CreatePartitioner - create a partitioner based on the partitioner type
//
// params:
//   - partitionerType int: partitioner type (from constants)
//   - volume *models.Volume: volume to create partitioner for
//
// return type:
//   - models.Partitioner: created partitioner of appropriate type or nil if type is invalid
func CreatePartitioner(partitionerType int, volume *Volume) Partitioner {
	switch partitionerType {
	case constants.PARTITION_TYPE_BALANCED:
		logger.Logger.Debug("partitioner", "Created a new balanced partitioner.")
		return NewBalancedPartitioner(volume)
	case constants.PARTITION_TYPE_PRIORITY:
		logger.Logger.Debug("partitioner", "Created a new priority partitioner.")
		return NewPriorityPartitioner(volume)
	case constants.PARTITION_TYPE_THROUGHPUT:
		logger.Logger.Debug("partitioner", "Created a new throughput partitioner.")
		return NewThroughputPartitioner(volume)

	default:
		logger.Logger.Warning("partitioner", "Could not create a partitioner.")
		return nil
	}
}

type AbstractPartitioner struct {
	Volume *Volume
}

func (p *AbstractPartitioner) AssignDisk(size int) *Disk {
	panic("Unimplemented abstract method!")
}

func (p *AbstractPartitioner) FetchDisks() {
	panic("Unimplemented abstract method!")
}

type BalancedPartitioner struct {
	AbstractPartitioner
	Disks               []Disk
	LastPickedDiskIndex int
}

// AssignDisk - assign a disk to write a block of given size to
//
// Balanced partitioner will assign a next disk to write a block of given
// size to in a round-robin fashion.
//
// params:
//   - size int: size of the block to write
//
// return type:
//   - *models.Disk: disk to write to or nil if no disk is available
func (p *BalancedPartitioner) AssignDisk(size int) Disk {
	// If there are no disks, return nil
	if len(p.Disks) == 0 {
		logger.Logger.Warning("partitioner", "There are no disk available to this partitioner.")
		return nil
	}

	// Choose the next disk
	p.LastPickedDiskIndex = (p.LastPickedDiskIndex + 1) % len(p.Disks)
	logger.Logger.Debug("partitioner", "Chosen disk is: ", p.Disks[p.LastPickedDiskIndex].GetName(), ".")
	return p.Disks[p.LastPickedDiskIndex]
}

// FetchDisks - fetch disks from volume and reset last picked disk index
func (p *BalancedPartitioner) FetchDisks(disks []Disk) {
	// Load disk list again in case something has changed in volume
	p.Disks = make([]Disk, 0)
	for _, disk := range disks {
		if ComputeFreeSpace(disk) > uint64(p.AbstractPartitioner.Volume.BlockSize) {
			p.Disks = append(p.Disks, disk)
		}
	}

	// Reset last picked disk index
	p.LastPickedDiskIndex = -1

	logger.Logger.Debug("partitioner", "Fetched disks.")
}

// NewBalancedPartitioner - create new balanced partitioner object
//
// return type:
//   - *models.BalancedPartitioner: created partitioner object
func NewBalancedPartitioner(volume *Volume) *BalancedPartitioner {
	var p BalancedPartitioner

	p.AbstractPartitioner.Volume = volume

	return &p
}

type PriorityPartitioner struct {
	AbstractPartitioner
	Disks           []Disk
	CachedFreeSpace []uint64
}

func (p *PriorityPartitioner) getNextDiskIndex(size int) int {
	// Find first disk which has enough free space
	for i, _ := range p.Disks {
		if p.CachedFreeSpace[i] >= uint64(size) {
			logger.Logger.Debug("partitioner", "Selected disk no. #", strconv.Itoa(i), ".")
			return i
		}
	}

	logger.Logger.Warning("partitioner", "Could not find a suitable disk.")
	return -1
}

// AssignDisk - assign a disk to write a block of given size to
//
// Balanced partitioner will assign a next disk based on the creation
// order of the disks. First disk with enough free space will be returned.
//
// params:
//   - size int: size of the block to write
//
// return type:
//   - *models.Disk: disk to write to or nil if no disk is available
func (p *PriorityPartitioner) AssignDisk(size int) Disk {
	// If there are no disks, return nil
	if len(p.Disks) == 0 {
		return nil
	}

	// Choose the next disk
	index := p.getNextDiskIndex(size)
	if index == -1 {
		// All disks are full
		logger.Logger.Warning("partitioner", "Could not find a suitable disk.")
		return nil
	}
	p.CachedFreeSpace[index] -= uint64(size)

	logger.Logger.Debug("partitioner", "Selected the disk: ", p.Disks[index].GetName(), ".")
	return p.Disks[index]
}

// FetchDisks - fetch disks from volume and retrieve free space
func (p *PriorityPartitioner) FetchDisks(disks []Disk) {
	// Load disk list again in case something has changed in volume
	var _disks []Disk = make([]Disk, 0)
	for _, disk := range disks {
		_disks = append(_disks, disk)
	}

	// Sort disks by creation order
	sort.Slice(_disks, func(i, j int) bool {
		return _disks[i].GetCreationTime().Before(_disks[j].GetCreationTime())
	})

	// Compute free space for each disk
	p.Disks = make([]Disk, 0)
	p.CachedFreeSpace = make([]uint64, 0)
	for _, disk := range _disks {
		freeSpace := ComputeFreeSpace(disk)
		if freeSpace > uint64(p.AbstractPartitioner.Volume.BlockSize) {
			p.Disks = append(p.Disks, disk)
			p.CachedFreeSpace = append(p.CachedFreeSpace, freeSpace)
		}
	}

	logger.Logger.Debug("partitioner", "Fetched disks.")
}

// NewPriorityPartitioner - create new priority partitioner object
//
// return type:
//   - *models.PriorityPartitioner: created partitioner object
func NewPriorityPartitioner(volume *Volume) *PriorityPartitioner {
	var p PriorityPartitioner

	p.AbstractPartitioner.Volume = volume

	return &p
}

type ThroughputPartitioner struct {
	AbstractPartitioner
	Disks               []Disk
	Weights             []int // Weights based on disk throughput
	Allocations         []int // Number of blocks allocations per disk
	LastPickedDiskIndex int
}

func (p *ThroughputPartitioner) getNextDiskIndex(size int) int {
	var minValue int = p.Weights[0] * p.Allocations[0]
	var minValueIdx int = 0

	// Find the disk with the lowest throughput utilization value
	for i := 1; i < len(p.Disks); i++ {
		value := p.Weights[i] * p.Allocations[i]
		if value < minValue {
			minValue = value
			minValueIdx = i
		}
	}

	return minValueIdx
}

// AssignDisk - assign a disk to write a block of given size to
//
// Throughput partitioner will assign a next disk based on the disk
// throughput weights and number of allocations. Disk with the lowest
// coefficient will be returned.
//
// params:
//   - size int: size of the block to write
//
// return type:
//   - *models.Disk: disk to write to or nil if no disk is available
func (p *ThroughputPartitioner) AssignDisk(size int) Disk {
	// If there are no disks, return nil
	if len(p.Disks) == 0 {
		return nil
	}

	// Choose the next disk
	index := p.getNextDiskIndex(size)
	p.Allocations[index] += 1

	logger.Logger.Debug("partitioner", "Selected the disk: ", p.Disks[index].GetName(), ".")
	return p.Disks[index]
}

// FetchDisks - fetch disks from volume and compute weights based on throughput
func (p *ThroughputPartitioner) FetchDisks(disks []Disk) {
	// Load disk list again in case something has changed in volume
	p.Disks = make([]Disk, 0)
	for _, disk := range disks {
		if ComputeFreeSpace(disk) > uint64(p.AbstractPartitioner.Volume.BlockSize) {
			p.Disks = append(p.Disks, disk)
		}
	}

	// Reset weights and allocations
	p.Weights = make([]int, len(p.Disks))
	p.Allocations = make([]int, len(p.Disks))

	// Compute throughput weights and reset allocations
	for i, disk := range p.Disks {
		p.Weights[i] = MeasureDiskThroughput(disk)
		p.Allocations[i] = 0
	}

	logger.Logger.Debug("partitioner", "Fetched disks.")
}

// NewThroughputPartitioner - create new throughput partitioner object
//
// return type:
//   - *models.ThroughputPartitioner: created partitioner object
func NewThroughputPartitioner(volume *Volume) *ThroughputPartitioner {
	var p ThroughputPartitioner

	p.AbstractPartitioner.Volume = volume

	return &p
}
