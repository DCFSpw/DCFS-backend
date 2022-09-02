package GDriveDisk

import "dcfs/models/disk"

type GDriveDisk struct {
	abstractDisk disk.AbstractDisk
}

func (d *GDriveDisk) Connect() error {
	return nil
}

func (d *GDriveDisk) Upload() error {
	return nil
}

func (d *GDriveDisk) Download() error {
	return nil
}

func (d *GDriveDisk) Rename() error {
	return nil
}

func (d *GDriveDisk) Remove() error {
	return nil
}

func NewGDriveDisk() *GDriveDisk {
	var d *GDriveDisk = new(GDriveDisk)
	d.abstractDisk.Disk = d
	return d
}
