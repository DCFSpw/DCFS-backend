package SFTPDisk

import "dcfs/models/disk"

type SFTPDisk struct {
	abstractDisk disk.AbstractDisk
}

func (d *SFTPDisk) Connect() error {
	return nil
}

func (d *SFTPDisk) Upload() error {
	return nil
}

func (d *SFTPDisk) Download() error {
	return nil
}

func (d *SFTPDisk) Rename() error {
	return nil
}

func (d *SFTPDisk) Remove() error {
	return nil
}

func NewSFTPDisk() *SFTPDisk {
	var d *SFTPDisk = new(SFTPDisk)
	d.abstractDisk.Disk = d
	return d
}
