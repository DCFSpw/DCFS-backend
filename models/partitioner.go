package models

type Partitioner interface {
	AssignDisk(int) Disk
}

type AbstractPartitioner struct {
	Volume *Volume
}

func (p *AbstractPartitioner) AssignDisk(size int) *Disk {
	panic("Unimplemented abstract method!")
}

type DummyPartitioner struct {
	AbstractPartitioner
	LastPickedDiskIndex int
}

func (p *DummyPartitioner) AssignDisk(size int) Disk {
	// load disk list again in case something has changed in volume
	var disks []Disk
	for _, _d := range p.Volume.disks {
		disks = append(disks, _d)
	}

	// choose the next disk
	p.LastPickedDiskIndex = (p.LastPickedDiskIndex + 1) % len(p.Volume.disks)
	return disks[p.LastPickedDiskIndex]
}

func NewDummyPartitioner(volume *Volume) *DummyPartitioner {
	return &DummyPartitioner{AbstractPartitioner: AbstractPartitioner{Volume: volume}}
}
