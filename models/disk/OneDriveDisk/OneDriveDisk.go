package OneDriveDisk

import "dcfs/models/disk"

type OneDriveDisk struct {
	abstractDisk disk.AbstractDisk
}

func (d *OneDriveDisk) Connect() error {
	return nil
}

func (d *OneDriveDisk) Upload() error {
	return nil
}

func (d *OneDriveDisk) Download() error {
	return nil
}

func (d *OneDriveDisk) Rename() error {
	return nil
}

func (d *OneDriveDisk) Remove() error {
	return nil
}

func NewOneDriveDisk() *OneDriveDisk {
	var d *OneDriveDisk = new(OneDriveDisk)
	d.abstractDisk.Disk = d
	return d
}
