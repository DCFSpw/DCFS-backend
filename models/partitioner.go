package models

type Partitioner interface {
	AssignDisk(int) Disk
	FetchDisks()
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
		p.Disks = append(p.Disks, disk)
	}

	// Reset last picked disk index
	p.LastPickedDiskIndex = -1
}

func NewBalancedPartitioner(volume *Volume) *BalancedPartitioner {
	var p BalancedPartitioner

	p.AbstractPartitioner.Volume = volume
	p.FetchDisks()

	return &p
}
