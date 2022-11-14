package models

import (
	"dcfs/constants"
	"sort"
)

type Partitioner interface {
	AssignDisk(int) Disk
	FetchDisks()
}

func CreatePartitioner(partitionerType int, volume *Volume) Partitioner {
	switch partitionerType {
	case constants.PARTITION_TYPE_BALANCED:
		return NewBalancedPartitioner(volume)
	case constants.PARTITION_TYPE_PRIORITY:
		return NewPriorityPartitioner(volume)
	case constants.PARTITION_TYPE_THROUGHPUT:
		return NewThroughputPartitioner(volume)

	default:
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

func (p *BalancedPartitioner) AssignDisk(size int) Disk {
	// If there are no disks, return nil
	if len(p.Disks) == 0 {
		return nil
	}

	// Choose the next disk
	p.LastPickedDiskIndex = (p.LastPickedDiskIndex + 1) % len(p.Disks)
	return p.Disks[p.LastPickedDiskIndex]
}

func (p *BalancedPartitioner) FetchDisks() {
	// Load disk list again in case something has changed in volume
	p.Disks = make([]Disk, 0)
	for _, disk := range p.AbstractPartitioner.Volume.disks {
		if ComputeFreeSpace(disk) > uint64(p.AbstractPartitioner.Volume.BlockSize) {
			p.Disks = append(p.Disks, disk)
		}
	}

	// Reset last picked disk index
	p.LastPickedDiskIndex = -1
}

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
			return i
		}
	}

	return -1
}

func (p *PriorityPartitioner) AssignDisk(size int) Disk {
	// If there are no disks, return nil
	if len(p.Disks) == 0 {
		return nil
	}

	// Choose the next disk
	index := p.getNextDiskIndex(size)
	if index == -1 {
		// All disks are full
		return nil
	}
	p.CachedFreeSpace[index] -= uint64(size)

	return p.Disks[index]
}

func (p *PriorityPartitioner) FetchDisks() {
	// Load disk list again in case something has changed in volume
	var _disks []Disk = make([]Disk, 0)
	for _, disk := range p.AbstractPartitioner.Volume.disks {
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
}

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

func (p *ThroughputPartitioner) AssignDisk(size int) Disk {
	// If there are no disks, return nil
	if len(p.Disks) == 0 {
		return nil
	}

	// Choose the next disk
	index := p.getNextDiskIndex(size)
	p.Allocations[index] += 1

	return p.Disks[index]
}

func (p *ThroughputPartitioner) FetchDisks() {
	// Load disk list again in case something has changed in volume
	p.Disks = make([]Disk, 0)
	for _, disk := range p.AbstractPartitioner.Volume.disks {
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
}

func NewThroughputPartitioner(volume *Volume) *ThroughputPartitioner {
	var p ThroughputPartitioner

	p.AbstractPartitioner.Volume = volume

	return &p
}
